import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Auth, signInWithPopup, GoogleAuthProvider, idToken } from '@angular/fire/auth';
import { environment } from '../../environments/environment';
import { Observable, from, switchMap } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private http = inject(HttpClient);
  private auth = inject(Auth);

  constructor() {}

  /**
   * Login using email and password via API Gateway
   */
  loginWithEmail(email: string, password: string, role: string): Observable<any> {
    return this.http.post(`${environment.apiUrl}/api/v1/auth/login`, {
      email,
      password,
      role
    });
  }

  /**
   * Login using Google via Firebase Auth and send token to backend
   */
  loginWithGoogle(role: string): Observable<any> {
    const provider = new GoogleAuthProvider();
    return from(signInWithPopup(this.auth, provider)).pipe(
      switchMap(credential => {
        return from(credential.user.getIdToken());
      }),
      switchMap(token => {
        return this.http.post(`${environment.apiUrl}/api/v1/auth/login`, {
          googleToken: token,
          role
        });
      })
    );
  }
}
