import { Injectable, inject } from "@angular/core";
import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { Observable, throwError } from "rxjs";
import { catchError } from "rxjs/operators";
import { environment } from "../../environments/environment";
import {
  CourseSkills,
  ParticipantSkill,
  Skill,
  SkillInput,
  SkillSuggestionRequest,
  SkillSuggestionResponse,
} from "../core/models/skills.model";

/** Texto que se muestra cuando ninguna IA pudo sugerir (503 del backend). */
export const SUGGESTIONS_UNAVAILABLE =
  "No pudimos generar sugerencias en este momento. Podés cargarlas a mano o probar de nuevo más tarde.";

/**
 * Cliente de las habilidades de education-service (HU-21 / HU-58).
 * Mismo patrón que EducationService: cookie de sesión con
 * `withCredentials` y errores normalizados al `{ "error": "..." }` del
 * backend.
 */
@Injectable({ providedIn: "root" })
export class SkillsService {
  private readonly http = inject(HttpClient);
  private readonly base =
    environment.educationApiUrl || environment.apiUrl || "";

  /** Autocompletado del catálogo. */
  searchCatalog(query: string): Observable<Skill[]> {
    return this.http
      .get<Skill[]>(`${this.base}/api/habilidades`, {
        params: { q: query },
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  /** Habilidades del curso y de cada una de sus actividades. */
  getCourseSkills(cursoId: string): Observable<CourseSkills> {
    return this.http
      .get<CourseSkills>(`${this.base}/api/cursos/${cursoId}/habilidades`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  /** Reemplaza las habilidades de una actividad (lista completa). */
  setActivitySkills(
    actividadId: string,
    habilidades: SkillInput[],
  ): Observable<Skill[]> {
    return this.http
      .put<Skill[]>(
        `${this.base}/api/actividades/${actividadId}/habilidades`,
        { habilidades },
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  /**
   * Pide sugerencias a la IA. Se llama solo cuando alguien toca el botón
   * "Sugerir habilidades con IA", nunca mientras escribe: cada pedido
   * tiene un costo.
   */
  suggest(req: SkillSuggestionRequest): Observable<SkillSuggestionResponse> {
    return this.http
      .post<SkillSuggestionResponse>(
        `${this.base}/api/actividades/sugerir-habilidades`,
        req,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  /** Habilidades del participante con sesión iniciada. */
  mySkills(): Observable<ParticipantSkill[]> {
    return this.http
      .get<ParticipantSkill[]>(`${this.base}/api/participante/habilidades`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  private readonly handleError = (error: HttpErrorResponse) => {
    const body = error.error as { error?: string } | string | null;
    let message = "No se pudieron cargar las habilidades.";
    if (error.status === 503) {
      message = SUGGESTIONS_UNAVAILABLE;
    } else if (error.status === 429) {
      message =
        "Se hicieron demasiados pedidos seguidos. Esperá un minuto y volvé a tocar Sugerir habilidades.";
    } else if (body && typeof body === "object" && body.error) {
      message = body.error;
    } else if (error.status === 401) {
      message = "Tu sesión expiró o no es válida. Volvé a iniciar sesión.";
    } else if (error.status === 403) {
      message =
        "No tenés permisos para gestionar las habilidades de este curso.";
    }
    return throwError(() => new Error(message));
  };
}
