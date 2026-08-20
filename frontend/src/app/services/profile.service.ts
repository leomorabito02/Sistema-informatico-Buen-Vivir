import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class ProfileService {
  private http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl;

  constructor() {}

  getMyProfile(): Observable<any> {
    return this.http.get<any>(`${this.apiUrl}/api/v1/profile/me`, { withCredentials: true });
  }

  updateMyPassword(oldPassword: string, newPassword: string): Observable<any> {
    return this.http.put<any>(
      `${this.apiUrl}/api/v1/profile/me/password`,
      { old_password: oldPassword, new_password: newPassword },
      { withCredentials: true }
    );
  }

  linkGoogleAccount(idToken: string, googleId: string): Observable<any> {
    return this.http.post<any>(
      `${this.apiUrl}/api/v1/profile/me/link-google`,
      { id_token: idToken, google_id: googleId },
      { withCredentials: true }
    );
  }

  unlinkGoogleAccount(): Observable<any> {
    return this.http.post<any>(
      `${this.apiUrl}/api/v1/profile/me/unlink-google`,
      {},
      { withCredentials: true }
    );
  }
}
