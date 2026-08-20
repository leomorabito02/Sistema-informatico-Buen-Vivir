import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class ProfileService {
  private http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl;

  constructor() {}

  getMyProfile(): Observable<any> {
    return this.http.get<any>(`${this.apiUrl}/api/v1/profile/me`, { withCredentials: true }).pipe(
      catchError(this.handleError)
    );
  }

  updateMyPassword(oldPassword: string, newPassword: string): Observable<any> {
    return this.http.put<any>(
      `${this.apiUrl}/api/v1/profile/me/password`,
      { old_password: oldPassword, new_password: newPassword },
      { withCredentials: true }
    ).pipe(
      catchError(this.handleError)
    );
  }

  linkGoogleAccount(idToken: string, googleId: string): Observable<any> {
    return this.http.post<any>(
      `${this.apiUrl}/api/v1/profile/me/link-google`,
      { id_token: idToken, google_id: googleId },
      { withCredentials: true }
    ).pipe(
      catchError(this.handleError)
    );
  }

  unlinkGoogleAccount(): Observable<any> {
    return this.http.post<any>(
      `${this.apiUrl}/api/v1/profile/me/unlink-google`,
      {},
      { withCredentials: true }
    ).pipe(
      catchError(this.handleError)
    );
  }
  private handleError = (error: any) => {
    let errorMessage = 'Ocurrió un error inesperado. Por favor, intente nuevamente.';
    if (error.error && error.error.message) {
      errorMessage = error.error.message;
    } else if (error.status === 401) {
      errorMessage = 'No autorizado. Verifique sus credenciales.';
    } else if (error.status === 403) {
      errorMessage = 'No tiene permisos para realizar esta acción.';
    }
    return throwError(() => new Error(errorMessage));
  }
}
