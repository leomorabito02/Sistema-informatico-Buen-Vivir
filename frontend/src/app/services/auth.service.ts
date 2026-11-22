import { Injectable, inject } from "@angular/core";
import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { Auth, signInWithPopup, GoogleAuthProvider } from "@angular/fire/auth";
import { environment } from "../../environments/environment";
import { Observable, firstValueFrom, from, throwError } from "rxjs";
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
import { SKIP_GLOBAL_ERROR_HEADER } from "../core/interceptors/error.interceptor";

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

  /** El token se registra una vez por carga de la página, no en cada navegación. */
  private pushSynced = false;

  constructor() {}

  /**
   * Registra el dispositivo para las notificaciones push.
   * @param askPermission `true` solo al iniciar sesión o registrarse: ahí se
   *   muestra el pedido de permiso. Al validar la sesión (en cada guard) va
   *   `false`, para no volver a preguntar en cada pantalla.
   */
  async syncPushNotifications(askPermission = false): Promise<void> {
    if (this.pushSynced) {
      return;
    }
    try {
      const token =
        await this.pushService.requestPermissionAndGetToken(askPermission);
      if (token) {
        this.pushSynced = true;
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
        {
          withCredentials: true,
          headers: { [SKIP_GLOBAL_ERROR_HEADER]: "true" },
        },
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
            {
              withCredentials: true,
              headers: { [SKIP_GLOBAL_ERROR_HEADER]: "true" },
            },
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
          void this.syncPushNotifications(true);
        }),
        catchError(this.handleError),
      );
  }

  /**
   * Completes 2FA verification during login challenge.
   */
  verifyTwoFactor(
    twofaNonce: string,
    code: string,
    trustDevice = false,
  ): Observable<CheckCredentialsResponse> {
    return this.http
      .post<CheckCredentialsResponse>(
        `${this.apiUrl}/api/v1/auth/verify-2fa`,
        { twofa_nonce: twofaNonce, code, trust_device: trustDevice },
        {
          withCredentials: true,
          headers: { [SKIP_GLOBAL_ERROR_HEADER]: "true" },
        },
      )
      .pipe(catchError(this.handleError));
  }

  /**
   * Resends email OTP code for an active 2FA challenge.
   */
  resendTwoFactorCode(twofaNonce: string): Observable<{ message: string }> {
    return this.http
      .post<{ message: string }>(
        `${this.apiUrl}/api/v1/auth/resend-2fa`,
        { twofa_nonce: twofaNonce },
        {
          withCredentials: true,
          headers: { [SKIP_GLOBAL_ERROR_HEADER]: "true" },
        },
      )
      .pipe(catchError(this.handleError));
  }

  /**
   * Renueva la sesión. Con varias pestañas abiertas, solo una pide el
   * refresh a la vez (Web Locks) y las demás reutilizan esa renovación:
   * las cookies son compartidas, así que no hace falta pedir otra. Antes,
   * si dos pestañas renovaban juntas, la segunda mandaba un refresh token
   * recién rotado, auth lo tomaba como reutilización y cerraba todas las
   * sesiones del usuario (D10, 22/11/2026).
   */
  refreshToken(): Observable<LoginResponse> {
    return from(
      conCandadoDeRefresh(() =>
        firstValueFrom(
          this.http.post<LoginResponse>(
            `${this.apiUrl}/api/v1/auth/refresh`,
            {},
            {
              withCredentials: true,
              headers: { [SKIP_GLOBAL_ERROR_HEADER]: "true" },
            },
          ),
        ),
      ),
    ).pipe(
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
          void this.syncPushNotifications(true);
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
    try {
      localStorage.removeItem(REFRESH_AT_KEY);
    } catch {
      /* sin almacenamiento */
    }
    this.http
      .post(
        `${this.apiUrl}/api/v1/auth/logout`,
        {},
        {
          withCredentials: true,
          headers: { [SKIP_GLOBAL_ERROR_HEADER]: "true" },
        },
      )
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

  /**
   * Dispara el reenvío del correo de confirmación con enlace temporal.
   */
  resendVerificationEmail(email: string): Observable<{ message: string }> {
    return this.http
      .post<{ message: string }>(
        `${this.apiUrl}/api/v1/auth/resend-verification`,
        { email },
        {
          withCredentials: true,
          headers: { [SKIP_GLOBAL_ERROR_HEADER]: "true" },
        },
      )
      .pipe(catchError(this.handleError));
  }

  /**
   * Ejecuta el PATCH para validar el token y pasar is_email_verified = true.
   */
  verifyEmail(
    token: string,
  ): Observable<{ is_email_verified: boolean; message: string }> {
    return this.http
      .patch<{ is_email_verified: boolean; message: string }>(
        `${this.apiUrl}/api/v1/auth/verify-email`,
        { token },
        {
          withCredentials: true,
          headers: { [SKIP_GLOBAL_ERROR_HEADER]: "true" },
        },
      )
      .pipe(catchError(this.handleError));
  }

  /**
   * Solicita el restablecimiento de contraseña enviando un enlace al email provisto.
   */
  forgotPassword(email: string): Observable<{ message: string }> {
    return this.http
      .post<{ message: string }>(
        `${this.apiUrl}/api/v1/auth/forgot-password`,
        { email },
        {
          withCredentials: true,
          headers: { [SKIP_GLOBAL_ERROR_HEADER]: "true" },
        },
      )
      .pipe(catchError(this.handleError));
  }

  /**
   * Restablece la contraseña utilizando el token temporal recibido por email.
   */
  resetPassword(
    token: string,
    newPassword: string,
  ): Observable<{ message: string }> {
    return this.http
      .post<{ message: string }>(
        `${this.apiUrl}/api/v1/auth/reset-password`,
        { token, new_password: newPassword },
        {
          withCredentials: true,
          headers: { [SKIP_GLOBAL_ERROR_HEADER]: "true" },
        },
      )
      .pipe(catchError(this.handleError));
  }

  private readonly handleError = (error: HttpErrorResponse) => {
    let errorMessage =
      "Ocurrió un error inesperado. Por favor, intente nuevamente.";
    if (error.status === 401) {
      this.clearLocallyAuthenticated();
    }

    const body = error.error as { message?: string; error?: string } | null;
    if (body?.message) {
      errorMessage = body.message;
    } else if (typeof body?.error === "string" && body.error) {
      // Auth responde {"error": "..."}: antes se perdía este mensaje.
      errorMessage = body.error;
    } else if (error.status === 429) {
      errorMessage =
        "Hiciste demasiados intentos seguidos. Esperá un minuto y probá de nuevo.";
    } else if (error.status === 401) {
      errorMessage = "No autorizado. Verifique sus credenciales.";
    } else if (error.status === 403) {
      errorMessage = "No tiene permisos para realizar esta acción.";
    }
    // Se conservan status y cuerpo para que el login distinga el bloqueo
    // (423), la cuenta suspendida y el resto.
    return throwError(() =>
      Object.assign(new Error(errorMessage), {
        status: error.status,
        error: error.error,
      }),
    );
  };
}

/** Clave de localStorage con la hora de la última renovación exitosa. */
export const REFRESH_AT_KEY = "bv_refresh_at";
/** Si otra pestaña renovó hace menos que esto, se reutiliza su renovación. */
export const REFRESH_REUSE_MS = 15_000;

function leerUltimoRefresh(): number {
  try {
    return Number(localStorage.getItem(REFRESH_AT_KEY) ?? 0) || 0;
  } catch {
    return 0;
  }
}

function anotarRefresh(): void {
  try {
    localStorage.setItem(REFRESH_AT_KEY, String(Date.now()));
  } catch {
    /* sin almacenamiento: cada pestaña renueva por su cuenta */
  }
}

/**
 * Ejecuta `pedir` con un candado compartido entre pestañas del mismo
 * origen. Si mientras esperaba el candado otra pestaña ya renovó, no vuelve
 * a pedir y devuelve una respuesta vacía (la sesión ya está renovada en las
 * cookies). Sin Web Locks (navegadores viejos) se comporta como antes.
 */
export async function conCandadoDeRefresh<T>(
  pedir: () => Promise<T>,
): Promise<T> {
  const ejecutar = async (): Promise<T> => {
    if (Date.now() - leerUltimoRefresh() < REFRESH_REUSE_MS) {
      return {} as T;
    }
    const res = await pedir();
    anotarRefresh();
    return res;
  };
  const locks = (
    globalThis.navigator as Navigator & {
      locks?: { request<R>(name: string, cb: () => Promise<R>): Promise<R> };
    }
  )?.locks;
  return locks?.request
    ? locks.request("bv-auth-refresh", ejecutar)
    : ejecutar();
}
