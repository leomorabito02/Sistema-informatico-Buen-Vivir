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

interface UserProfileEnvelope {
	success: boolean;
	data: {
		user: any;
		role_data?: any;
	};
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
  private readonly http = inject(HttpClient);
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

  getUserProfile(id: string): Observable<{user: any, role_data?: any}> {
    return this.http
      .get<UserProfileEnvelope>(`${this.apiUrl}/api/v1/users/${id}/profile`, {
        withCredentials: true,
      })
      .pipe(
        map((res) => res.data),
        catchError((err) => {
          return throwError(() => new Error(err.error?.error || "Error al obtener perfil"));
        })
      );
  }

  updateUser(id: string, data: any): Observable<any> {
    return this.http
      .put(`${this.apiUrl}/api/v1/users/${id}`, data, {
        withCredentials: true,
      })
      .pipe(
        catchError((err) => {
          return throwError(() => new Error(err.error?.error || "Error al actualizar usuario"));
        })
      );
  }

  createUser(data: any): Observable<any> {
    return this.http
      .post(`${this.apiUrl}/api/v1/users`, data, {
        withCredentials: true,
      })
      .pipe(
        catchError((err) => {
          if (err.status === 401) {
            return throwError(() => new Error("Tu sesión expiró o no es válida. Volvé a iniciar sesión."));
          }
          
          let errMsg = "Error al crear usuario";
          if (err.error) {
              if (typeof err.error === 'string') {
                  errMsg = err.error;
              } else if (err.error.error) {
                  if (typeof err.error.error === 'string') {
                      errMsg = err.error.error;
                  } else {
                      // It's likely a validation map
                      const keys = Object.keys(err.error.error);
                      if (keys.length > 0) {
                          const fieldsMsg = keys.map(k => k + " (" + err.error.error[k] + ")").join(', ');
                          errMsg = "Campos inválidos: " + fieldsMsg;
                      } else {
                          errMsg = JSON.stringify(err.error.error);
                      }
                  }
              } else {
                  errMsg = JSON.stringify(err.error);
              }
          }
          return throwError(() => new Error(errMsg));
        })
      );
  }

  updateRoleData(tableName: string, id: string, data: any): Observable<any> {
    return this.http
      .put(`${this.apiUrl}/api/v1/${tableName}/${id}`, data, {
        withCredentials: true,
      })
      .pipe(
        catchError((err) => {
          return throwError(() => new Error(err.error?.error || "Error al actualizar rol"));
        })
      );
  }

  private readonly handleError = (error: HttpErrorResponse) => {
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
