import { Injectable, inject } from "@angular/core";
import {
  HttpClient,
  HttpErrorResponse,
  HttpParams,
} from "@angular/common/http";
import { Observable, throwError } from "rxjs";
import { catchError, map } from "rxjs/operators";
import { environment } from "../../environments/environment";
import {
  CreateFichaInput,
  Ficha,
  FichaDetalle,
  Hito,
  HitoDetalle,
  HitoInput,
  HitoParticipante,
  ResultadoBusqueda,
  UpdateFichaInput,
  UpdateFichaResponse,
} from "../core/models/human-promotion.model";

/** Error del servicio con el status HTTP, para distinguir "sin ficha" (404). */
export class HumanPromotionError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
  }
}

/**
 * Cliente del microservicio de Promoción Humana (human-promotion).
 *
 * Mismo patrón que EducationService: `withCredentials: true` (la cookie
 * de sesión de auth-service viaja sola y el servicio la valida) y
 * `catchError` que normaliza el `{ "error": "..." }` del backend.
 *
 * Base URL: vacía en dev local (el proxy manda `/api/v1/human-promotion`
 * a :8085) y `apiUrl` (el gateway) en Docker.
 */
@Injectable({ providedIn: "root" })
export class HumanPromotionService {
  private readonly http = inject(HttpClient);
  private readonly base = `${environment.apiUrl ?? ""}/api/v1/human-promotion`;

  // ---- Búsqueda (HU-36) --------------------------------------------------
  buscarParticipantes(query: string): Observable<ResultadoBusqueda[]> {
    const params = new HttpParams().set("q", query);
    return this.http
      .get<ResultadoBusqueda[]>(`${this.base}/participantes`, {
        params,
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  // ---- Fichas (HU-29, HU-30, HU-31) --------------------------------------
  /** Falla con status 404 si el participante todavía no tiene ficha. */
  getFichaPorParticipante(participanteId: string): Observable<FichaDetalle> {
    const params = new HttpParams().set("participante_id", participanteId);
    return this.http
      .get<FichaDetalle>(`${this.base}/fichas`, {
        params,
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  crearFicha(input: CreateFichaInput): Observable<Ficha> {
    return this.http
      .post<Ficha>(`${this.base}/fichas`, input, { withCredentials: true })
      .pipe(catchError(this.handleError));
  }

  editarFicha(
    fichaId: string,
    input: UpdateFichaInput,
  ): Observable<UpdateFichaResponse> {
    return this.http
      .put<UpdateFichaResponse>(
        `${this.base}/fichas/${encodeURIComponent(fichaId)}`,
        input,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  /** PDF de la ficha y su línea de hitos, generado en el backend (HU-37). */
  descargarPdf(fichaId: string): Observable<Blob> {
    return this.http
      .get(`${this.base}/fichas/${encodeURIComponent(fichaId)}/pdf`, {
        responseType: "blob",
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  // ---- Hitos (HU-32, HU-33) ----------------------------------------------
  listarHitos(fichaId: string): Observable<HitoDetalle[]> {
    return this.http
      .get<HitoDetalle[]>(
        `${this.base}/fichas/${encodeURIComponent(fichaId)}/hitos`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  crearHito(fichaId: string, input: HitoInput): Observable<Hito> {
    return this.http
      .post<Hito>(
        `${this.base}/fichas/${encodeURIComponent(fichaId)}/hitos`,
        input,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  editarHito(
    fichaId: string,
    hitoId: string,
    input: HitoInput,
  ): Observable<Hito> {
    return this.http
      .put<Hito>(
        `${this.base}/fichas/${encodeURIComponent(fichaId)}/hitos/${encodeURIComponent(hitoId)}`,
        input,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  // ---- Participante (HU-34) ----------------------------------------------
  miLineaDeTiempo(): Observable<HitoParticipante[]> {
    return this.http
      .get<{
        hitos: HitoParticipante[];
      }>(`${this.base}/mi-linea-de-tiempo`, { withCredentials: true })
      .pipe(
        map((res) => res.hitos ?? []),
        catchError(this.handleError),
      );
  }

  // ------------------------------------------------------------------------
  private readonly handleError = (error: HttpErrorResponse) => {
    let message =
      "Ocurrió un error al comunicarse con el módulo de Promoción Humana.";
    const body = error.error as { error?: string } | string | null;
    if (body && typeof body === "object" && body.error) {
      message = body.error;
    } else if (error.status === 0) {
      message =
        "No se pudo conectar con el módulo de Promoción Humana. Revisá tu conexión.";
    } else if (error.status === 401) {
      message = "Tu sesión expiró o no es válida. Volvé a iniciar sesión.";
    } else if (error.status === 403) {
      message = "No tenés permisos para acceder a esta información.";
    } else if (error.status === 404) {
      message = "No se encontró la información solicitada.";
    }
    return throwError(() => new HumanPromotionError(message, error.status));
  };
}
