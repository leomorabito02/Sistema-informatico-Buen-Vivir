import { Injectable, inject } from "@angular/core";
import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { Auth, signInWithPopup, GoogleAuthProvider } from "@angular/fire/auth";
import { environment } from "../../environments/environment";
import { Observable, from, throwError } from "rxjs";
import { catchError, switchMap, tap } from "rxjs/operators";
import { Router } from "@angular/router";
import {
  CheckCredentialsResponse,
  LoginResponse,
  MeResponse,
  RegisterRequest,
  UpdateCompanyProfile,
  UpdateParticipantProfile,
  UpdateTeacherProfile,
  User,
} from "../core/models/auth.model";

import { PushNotificationService } from "./push-notification.service";

@Injectable({
  providedIn: "root",
})
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly auth = inject(Auth);
  private readonly router = inject(Router);
  private readonly pushService = inject(PushNotificationService);

  private readonly apiUrl = environment.apiUrl;

  constructor() {}

  async syncPushNotifications(): Promise<void> {
    try {
      const token = await this.pushService.requestPermissionAndGetToken();
      if (token) {
        this.pushService.registerDeviceToken(token).subscribe();
        this.pushService.listenForegroundMessages();
      }
    } catch {
      // Non-blocking: push token retrieval does not interrupt auth flow
    }
  }

  isLocallyAuthenticated(): boolean {
    return localStorage.getItem("isAuthenticated") === "true";
  }

  setLocallyAuthenticated(): void {
    localStorage.setItem("isAuthenticated", "true");
  }

  clearLocallyAuthenticated(): void {
    localStorage.removeItem("isAuthenticated");
  }

  /**
   * Step 1: Validates email and password. Returns a list of available accounts/roles.
   * No token is issued at this point.
   */
  checkCredentials(
    email: string,
    password: string,
  ): Observable<CheckCredentialsResponse> {
    return this.http
      .post<CheckCredentialsResponse>(
        `${this.apiUrl}/api/v1/auth/check-credentials`,
        { email, password },
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  /**
   * Step 1 (Google): Validates Firebase ID token. Returns a list of available accounts/roles.
   * No token is issued at this point.
   */
  checkGoogleCredentials(): Observable<CheckCredentialsResponse> {
    const provider = new GoogleAuthProvider();

    return from(signInWithPopup(this.auth, provider)).pipe(
      switchMap((credential) =>
        from(
          credential.user
            .getIdToken()
            .then((idToken) => ({ idToken, googleId: credential.user.uid })),
        ),
      ),
      switchMap(({ idToken, googleId }) =>
        this.http
          .post<CheckCredentialsResponse>(
            `${this.apiUrl}/api/v1/auth/google-check-credentials`,
            { id_token: idToken, google_id: googleId },
            { withCredentials: true },
          )
          .pipe(catchError(this.handleError)),
      ),
      catchError(this.handleError),
    );
  }

  /**
   * Step 2: Given a user_id, role_id and a single-use login nonce, issues the JWT token and sets auth cookies.
   * The nonce is obtained from the Step 1 (checkCredentials) response and is valid for 2 minutes.
   */
  loginWithAccount(
    userId: string,
    roleId: string,
    loginNonce: string,
  ): Observable<LoginResponse> {
    return this.http
      .post<LoginResponse>(
        `${this.apiUrl}/api/v1/auth/login`,
        { user_id: userId, role_id: roleId, login_nonce: loginNonce },
        { withCredentials: true },
      )
      .pipe(
        tap(() => {
          this.setLocallyAuthenticated();
          void this.syncPushNotifications();
        }),
        catchError(this.handleError),
      );
  }

  refreshToken(): Observable<LoginResponse> {
    return this.http
      .post<LoginResponse>(
        `${this.apiUrl}/api/v1/auth/refresh`,
        {},
        { withCredentials: true },
      )
      .pipe(
        tap(() => this.setLocallyAuthenticated()),
        catchError(this.handleError),
      );
  }

  /**
   * Register a new user via API Gateway.
   */
  registerWithEmail(userData: RegisterRequest): Observable<LoginResponse> {
    return this.http
      .post<LoginResponse>(`${this.apiUrl}/api/v1/auth/register`, userData, {
        withCredentials: true,
      })
      .pipe(
        tap(() => {
          this.setLocallyAuthenticated();
          void this.syncPushNotifications();
        }),
        catchError(this.handleError),
      );
  }

  /**
   * Checks whether the current session is valid.
   * The browser automatically sends the HttpOnly authentication cookie.
   */
  isAuthenticated(): Observable<MeResponse> {
    return this.http
      .get<MeResponse>(`${this.apiUrl}/api/v1/auth/me`, {
        withCredentials: true,
      })
      .pipe(
        tap(() => {
          void this.syncPushNotifications();
        }),
        catchError(this.handleError),
      );
  }

  updateProfile(
    data:
      UpdateParticipantProfile | UpdateCompanyProfile | UpdateTeacherProfile,
  ): Observable<User> {
    return this.http
      .put<User>(`${this.apiUrl}/api/v1/profile/me`, data, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  /**
   * Logs out the current session.
   * The backend is responsible for invalidating the session and clearing the cookies.
   */
  logout(): void {
    this.clearLocallyAuthenticated();
    this.http
      .post(`${this.apiUrl}/api/v1/auth/logout`, {}, { withCredentials: true })
      .subscribe({
        next: () => {
          void this.auth.signOut();
          void this.router.navigate(["/login"]);
        },
        error: () => {
          // Even if the backend request fails, clear the Firebase session and redirect
          void this.auth.signOut();
          void this.router.navigate(["/login"]);
        },
      });
  }

  private readonly handleError = (error: HttpErrorResponse) => {
    let errorMessage =
      "Ocurrió un error inesperado. Por favor, intente nuevamente.";
    if (error.status === 401) {
      this.clearLocallyAuthenticated();
    }

    if (error.error && (error.error as { message?: string }).message) {
      errorMessage = (error.error as { message: string }).message;
    } else if (error.status === 401) {
      errorMessage = "No autorizado. Verifique sus credenciales.";
    } else if (error.status === 403) {
      errorMessage = "No tiene permisos para realizar esta acción.";
    }
    return throwError(() => new Error(errorMessage));
  };
}
