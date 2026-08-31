import { Injectable, inject } from "@angular/core";
import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { Observable, map, throwError } from "rxjs";
import { catchError } from "rxjs/operators";
import { environment } from "../../environments/environment";
import { User } from "../core/models/auth.model";

interface UsersEnvelope {
  success: boolean;
  data: User[];
}

/**
 * Acceso al listado de usuarios de auth-service (`GET /api/v1/users`).
 * Solo lo permite el backend a `role_admin`; con otro rol devuelve 403/500.
 *
 * Se usa, por ahora, para el selector de "docente responsable" del alta
 * de cursos (HU-14). A futuro reemplazar por un endpoint de auth-service
 * que filtre por rol del lado del servidor.
 */
@Injectable({ providedIn: "root" })
export class UsersService {
  private http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl ?? "";

  listUsers(): Observable<User[]> {
    return this.http
      .get<UsersEnvelope>(`${this.apiUrl}/api/v1/users`, {
        withCredentials: true,
      })
      .pipe(
        map((res) => res.data ?? []),
        catchError(this.handleError),
      );
  }

  listByRole(roleId: string): Observable<User[]> {
    return this.listUsers().pipe(
      map((users) => users.filter((u) => u.role_id === roleId)),
    );
  }

  private handleError = (error: HttpErrorResponse) => {
    let message = "No se pudo obtener el listado de usuarios.";
    const body = error.error as { errors?: string; error?: string } | null;
    if (body?.errors) message = body.errors;
    else if (body?.error) message = body.error;
    else if (error.status === 401)
      message = "Sesión inválida. Volvé a iniciar sesión.";
    else if (error.status === 403 || error.status === 500)
      message = "Solo un administrador puede listar usuarios.";
    return throwError(() => new Error(message));
  };
}
