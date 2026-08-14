import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Auth, signInWithPopup, GoogleAuthProvider } from '@angular/fire/auth';
import { environment } from '../../environments/environment';
import { Observable, from, switchMap } from 'rxjs';
import { Router } from '@angular/router';
import { CheckCredentialsResponse, LoginResponse, RegisterRequest, User } from '../core/models/auth.model';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private http = inject(HttpClient);
  private auth = inject(Auth);
  private router = inject(Router);

  private readonly apiUrl = environment.apiUrl;

  constructor() {}

  /**
   * Step 1: Validates email and password. Returns a list of available accounts/roles.
   * No token is issued at this point.
   */
  checkCredentials(email: string, password: string): Observable<CheckCredentialsResponse> {
    return this.http.post<CheckCredentialsResponse>(
      `${this.apiUrl}/api/v1/auth/check-credentials`,
      { email, password },
      { withCredentials: true }
    );
  }

  /**
   * Step 1 (Google): Validates Firebase ID token. Returns a list of available accounts/roles.
   * No token is issued at this point.
   */
  checkGoogleCredentials(): Observable<CheckCredentialsResponse> {
    const provider = new GoogleAuthProvider();

    return from(signInWithPopup(this.auth, provider)).pipe(
      switchMap(credential => from(credential.user.getIdToken().then(idToken => ({ idToken, googleId: credential.user.uid })))),
      switchMap(({ idToken, googleId }) =>
        this.http.post<CheckCredentialsResponse>(
          `${this.apiUrl}/api/v1/auth/google-check-credentials`,
          { id_token: idToken, google_id: googleId },
          { withCredentials: true }
        )
      )
    );
  }

  /**
   * Step 2: Given a user_id, role_id and a single-use login nonce, issues the JWT token and sets auth cookies.
   * The nonce is obtained from the Step 1 (checkCredentials) response and is valid for 2 minutes.
   */
  loginWithAccount(userId: string, roleId: string, loginNonce: string): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(
      `${this.apiUrl}/api/v1/auth/login`,
      { user_id: userId, role_id: roleId, login_nonce: loginNonce },
      { withCredentials: true }
    );
  }

  /**
   * Register a new user via API Gateway.
   */
  registerWithEmail(userData: RegisterRequest): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(
      `${this.apiUrl}/api/v1/auth/register`,
      userData,
      { withCredentials: true }
    );
  }

  /**
   * Checks whether the current session is valid.
   * The browser automatically sends the HttpOnly authentication cookie.
   */
  isAuthenticated(): Observable<{user: User}> {
    return this.http.get<{user: User}>(
      `${this.apiUrl}/api/v1/auth/me`,
      { withCredentials: true }
    );
  }

  updateProfile(data: any): Observable<any> {
    return this.http.put<any>(
      `${this.apiUrl}/api/v1/profile/me`,
      data,
      { withCredentials: true }
    );
  }

  /**
   * Logs out the current session.
   * The backend is responsible for invalidating the session and clearing the cookies.
   */
  logout(): void {
    this.http.post(
      `${this.apiUrl}/api/v1/auth/logout`,
      {},
      { withCredentials: true }
    ).subscribe({
      next: () => {
        this.auth.signOut();
        this.router.navigate(['/login']);
      },
      error: () => {
        // Even if the backend request fails, clear the Firebase session and redirect
        this.auth.signOut();
        this.router.navigate(['/login']);
      }
    });
  }
}