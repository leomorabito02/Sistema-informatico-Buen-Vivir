import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Auth, signInWithPopup, GoogleAuthProvider } from '@angular/fire/auth';
import { environment } from '../../environments/environment';
import { Observable, from, switchMap } from 'rxjs';
import { Router } from '@angular/router';
import { LoginResponse, RegisterRequest, User } from '../core/models/auth.model';

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
   * Login using email and password via API Gateway.
   *
   * The backend is responsible for setting the authentication
   * and refresh cookies using Set-Cookie.
   */
  loginWithEmail(
    email: string,
    password: string,
    role: string
  ): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(
      `${this.apiUrl}/api/v1/auth/login`,
      {
        email,
        password,
        role
      },
      {
        withCredentials: true
      }
    );
  }

  /**
   * Register a new user via API Gateway.
   */
  registerWithEmail(userData: RegisterRequest): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(
      `${this.apiUrl}/api/v1/auth/register`,
      userData,
      {
        withCredentials: true
      }
    );
  }

  /**
   * Login using Google via Firebase Auth and send the Firebase ID token
   * to the backend.
   *
   * The backend creates the application session and sets the
   * authentication cookies.
   */
  loginWithGoogle(role: string): Observable<LoginResponse> {
    const provider = new GoogleAuthProvider();

    return from(signInWithPopup(this.auth, provider)).pipe(
      switchMap(credential =>
        from(credential.user.getIdToken())
      ),
      switchMap(token =>
        this.http.post<LoginResponse>(
          `${this.apiUrl}/api/v1/auth/login`,
          {
            googleToken: token,
            role
          },
          {
            withCredentials: true
          }
        )
      )
    );
  }

  /**
   * Checks whether the current session is valid.
   *
   * The browser automatically sends the HttpOnly authentication cookie.
   */
  isAuthenticated(): Observable<{user: User}> {
    return this.http.get<{user: User}>(
      `${this.apiUrl}/api/v1/auth/me`,
      {
        withCredentials: true
      }
    );
  }

  updateProfile(data: any): Observable<any> {
    return this.http.put<any>(
      `${this.apiUrl}/api/v1/profile/me`,
      data,
      {
        withCredentials: true
      }
    );
  }

  /**
   * Logs out the current session.
   *
   * The backend is responsible for invalidating the session
   * and clearing the authentication cookies.
   */
  logout(): void {
    this.http.post(
      `${this.apiUrl}/api/v1/auth/logout`,
      {},
      {
        withCredentials: true
      }
    ).subscribe({
      next: () => {
        this.auth.signOut();
        this.router.navigate(['/login']);
      },
      error: () => {
        // Even if the backend request fails, clear the Firebase session
        // and redirect the user.
        this.auth.signOut();
        this.router.navigate(['/login']);
      }
    });
  }
}