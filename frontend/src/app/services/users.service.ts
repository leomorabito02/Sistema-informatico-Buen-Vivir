import { Injectable, inject } from "@angular/core";
import {
  HttpClient,
  HttpErrorResponse,
  HttpParams,
} from "@angular/common/http";
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
 * Lo pueden pedir admin y voluntario (el voluntario recibe solo
 * participantes); con otro rol devuelve 403.
 *
 * El texto (`q`) y el rol (`role`) los filtra auth-service: así el
 * buscador de Usuarios y el selector de "docente responsable" (HU-14) no
 * bajan la lista entera.
 */
@Injectable({ providedIn: "root" })
export class UsersService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl ?? "";

  listUsers(filtros: { q?: string; role?: string } = {}): Observable<User[]> {
    let params = new HttpParams();
    const q = filtros.q?.trim();
    if (q) params = params.set("q", q);
    if (filtros.role) params = params.set("role", filtros.role);
    return this.http
      .get<UsersEnvelope>(`${this.apiUrl}/api/v1/users`, {
        params,
        withCredentials: true,
      })
      .pipe(
        map((res) => res.data ?? []),
        catchError(this.handleError),
      );
  }

  listByRole(roleId: string): Observable<User[]> {
    return this.listUsers({ role: roleId }).pipe(
      map((users) => users.filter((u) => u.role_id === roleId)),
    );
  }

  getUserProfile(id: string): Observable<{ user: any; role_data?: any }> {
    return this.http
      .get<UserProfileEnvelope>(`${this.apiUrl}/api/v1/users/${id}/profile`, {
        withCredentials: true,
      })
      .pipe(
        map((res) => res.data),
        catchError((err) => {
          return throwError(
            () => new Error(err.error?.error || "Error al obtener perfil"),
          );
        }),
      );
  }

  updateUser(id: string, data: any): Observable<any> {
    return this.http
      .put(`${this.apiUrl}/api/v1/users/${id}`, data, {
        withCredentials: true,
      })
      .pipe(
        catchError((err) => {
          return throwError(
            () => new Error(err.error?.error || "Error al actualizar usuario"),
          );
        }),
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
            return throwError(
              () =>
                new Error(
                  "Tu sesión expiró o no es válida. Volvé a iniciar sesión.",
                ),
            );
          }

          let errMsg = "Error al crear usuario";
          if (err.error) {
            if (typeof err.error === "string") {
              errMsg = err.error;
            } else if (err.error.error) {
              if (typeof err.error.error === "string") {
                errMsg = err.error.error;
              } else {
                // It's likely a validation map
                const keys = Object.keys(err.error.error);
                if (keys.length > 0) {
                  const fieldsMsg = keys
                    .map((k) => k + " (" + err.error.error[k] + ")")
                    .join(", ");
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
        }),
      );
  }

  /**
   * DELETE /api/v1/users/{id}. auth-service permite borrar cualquier cuenta
   * a role_admin / role_dev y solo la propia al resto.
   */
  deleteUser(id: string): Observable<void> {
    return this.http
      .delete<unknown>(`${this.apiUrl}/api/v1/users/${id}`, {
        withCredentials: true,
      })
      .pipe(
        map(() => undefined),
        catchError((err) => {
          const message =
            err.status === 401
              ? "Tu sesión expiró o no es válida. Volvé a iniciar sesión."
              : err.error?.error || "Error al eliminar el usuario";
          return throwError(() => new Error(message));
        }),
      );
  }

  /** Aprueba o rechaza una cuenta pendiente; Auth avisa por email (HU-12, HU-38). */
  setApproval(userId: string, status: "approved" | "rejected"): Observable<any> {
    return this.http
      .patch(
        `${this.apiUrl}/api/v1/users/${userId}/approval`,
        { approval_status: status },
        { withCredentials: true },
      )
      .pipe(
        catchError((err) =>
          throwError(
            () =>
              new Error(
                err.error?.error ||
                  err.error?.message ||
                  "No se pudo cambiar la aprobación",
              ),
          ),
        ),
      );
  }

  updateRoleData(tableName: string, id: string, data: any): Observable<any> {
    return this.http
      .put(`${this.apiUrl}/api/v1/${tableName}/${id}`, data, {
        withCredentials: true,
      })
      .pipe(
        catchError((err) => {
          return throwError(
            () => new Error(err.error?.error || "Error al actualizar rol"),
          );
        }),
      );
  }

  // Document Management Methods for Participants
  getParticipantDocuments(participantId: string): Observable<any[]> {
    return this.http
      .get<{ success: boolean; data: any[] }>(
        `${this.apiUrl}/api/v1/participants/${participantId}/documents`,
        {
          withCredentials: true,
        },
      )
      .pipe(map((res) => res.data));
  }

  uploadParticipantDocument(
    participantId: string,
    file: File,
    name: string,
  ): Observable<any> {
    const formData = new FormData();
    formData.append("file", file);
    if (name) {
      formData.append("name", name);
    }
    return this.http
      .post<{ success: boolean; data: any }>(
        `${this.apiUrl}/api/v1/participants/${participantId}/documents`,
        formData,
        {
          withCredentials: true,
        },
      )
      .pipe(map((res) => res?.data || res));
  }

  updateParticipantDocument(
    participantId: string,
    docId: string,
    file?: File,
    name?: string,
  ): Observable<any> {
    const formData = new FormData();
    if (file) formData.append("file", file);
    if (name) formData.append("name", name);

    return this.http
      .put<{ success: boolean; data: any }>(
        `${this.apiUrl}/api/v1/participants/${participantId}/documents/${docId}`,
        formData,
        {
          withCredentials: true,
        },
      )
      .pipe(map((res) => res?.data || res));
  }

  deleteParticipantDocument(
    participantId: string,
    docId: string,
  ): Observable<any> {
    return this.http.delete(
      `${this.apiUrl}/api/v1/participants/${participantId}/documents/${docId}`,
      {
        withCredentials: true,
      },
    );
  }

  downloadParticipantDocument(
    participantId: string,
    docId: string,
  ): Observable<Blob> {
    return this.http.get(
      `${this.apiUrl}/api/v1/participants/${participantId}/documents/${docId}/download`,
      {
        withCredentials: true,
        responseType: "blob",
      },
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
