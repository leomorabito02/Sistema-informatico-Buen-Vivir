import { Injectable, inject } from "@angular/core";
import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { environment } from "../../environments/environment";
import { Observable, throwError } from "rxjs";
import { catchError } from "rxjs/operators";
import {
  User,
  ProfileResponse,
} from "../core/models/auth.model";

interface PasswordUpdateRequest {
  old_password: string;
  new_password: string;
}

interface GoogleLinkRequest {
  id_token: string;
  google_id: string;
}

@Injectable({
  providedIn: "root",
})
export class ProfileService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl;

  constructor() {}

  getMyProfile(): Observable<ProfileResponse> {
    return this.http
      .get<ProfileResponse>(`${this.apiUrl}/api/v1/profile/me`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  updateMyPassword(oldPassword: string, newPassword: string): Observable<void> {
    const body: PasswordUpdateRequest = {
      old_password: oldPassword,
      new_password: newPassword,
    };
    return this.http
      .put<void>(`${this.apiUrl}/api/v1/profile/me/password`, body, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  linkGoogleAccount(idToken: string, googleId: string): Observable<User> {
    const body: GoogleLinkRequest = { id_token: idToken, google_id: googleId };
    return this.http
      .post<User>(`${this.apiUrl}/api/v1/profile/me/link-google`, body, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  unlinkGoogleAccount(): Observable<User> {
    return this.http
      .post<User>(
        `${this.apiUrl}/api/v1/profile/me/unlink-google`,
        {},
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  updateProfile(data: Record<string, any>): Observable<User> {
    return this.http
      .put<User>(`${this.apiUrl}/api/v1/profile/me`, data, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  private readonly handleError = (error: HttpErrorResponse) => {
    let errorMessage =
      "Ocurrió un error inesperado. Por favor, intente nuevamente.";
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
