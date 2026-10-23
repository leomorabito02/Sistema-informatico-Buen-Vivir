import { Injectable, inject } from "@angular/core";
import { HttpClient, HttpErrorResponse, HttpParams } from "@angular/common/http";
import { Observable, catchError, throwError } from "rxjs";
import { environment } from "../../environments/environment";

export type EntidadComentable = "curso" | "campana";

export interface Comentario {
  id: string;
  entidadTipo: EntidadComentable;
  entidadId: string;
  nombre: string;
  apellido: string;
  texto: string;
  fecha: string;
}

export interface NuevoComentario {
  entidadTipo: EntidadComentable;
  entidadId: string;
  nombre: string;
  apellido: string;
  texto: string;
}

/**
 * Comentarios de cursos y campañas (HU-24), en education-service. Leer y
 * comentar no piden sesión: el contenido comentado es público. Borrar es
 * la moderación del administrador.
 */
@Injectable({ providedIn: "root" })
export class ComentariosService {
  private readonly http = inject(HttpClient);
  private readonly base = `${environment.educationApiUrl || environment.apiUrl || ""}/api/comentarios`;

  listar(entidadTipo: EntidadComentable, entidadId: string): Observable<Comentario[]> {
    const params = new HttpParams()
      .set("entidadTipo", entidadTipo)
      .set("entidadId", entidadId);
    return this.http
      .get<Comentario[]>(this.base, { params })
      .pipe(catchError(this.error));
  }

  crear(c: NuevoComentario): Observable<Comentario> {
    return this.http
      .post<Comentario>(this.base, c)
      .pipe(catchError(this.error));
  }

  eliminar(id: string): Observable<void> {
    return this.http
      .delete<void>(`${this.base}/${encodeURIComponent(id)}`, {
        withCredentials: true,
      })
      .pipe(catchError(this.error));
  }

  private readonly error = (e: HttpErrorResponse) => {
    const mensaje =
      e.status === 429
        ? "Mandaste muchos comentarios seguidos. Probá de nuevo en un minuto."
        : e.error?.error || "No se pudo completar la acción con los comentarios.";
    return throwError(() => new Error(mensaje));
  };
}
