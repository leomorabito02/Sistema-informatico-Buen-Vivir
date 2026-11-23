import { Injectable, inject } from "@angular/core";
import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { Observable, throwError } from "rxjs";
import { catchError } from "rxjs/operators";
import { environment } from "../../environments/environment";
import {
  Campana,
  CampanaEscritura,
  CampanaInput,
  EstadoCampana,
} from "../core/models/campaign.model";

/**
 * Cliente de campaigns-service (HU-47 a HU-51), vía API Gateway.
 * Mismo patrón que el resto: cookie de sesión con `withCredentials` y
 * errores normalizados al `{ "error": "..." }` del backend.
 */
/** Tamaño de página del listado público de campañas (HU-50). */
export const PAGINA_CAMPANAS = 12;

@Injectable({ providedIn: "root" })
export class CampaignsService {
  private readonly http = inject(HttpClient);
  private readonly base = `${environment.apiUrl ?? ""}/api/v1/campaigns`;
  private readonly opts = { withCredentials: true };

  /** Campañas publicadas, sin sesión (HU-50), con búsqueda por título (HU-51). */
  publicas(busqueda = ""): Observable<Campana[]> {
    const params = busqueda.trim() ? { q: busqueda.trim() } : undefined;
    return this.http
      .get<Campana[]>(`${this.base}/publicas`, { params })
      .pipe(catchError(this.handleError));
  }

  /** Página del listado público (HU-50): {items, total}. */
  publicasPagina(
    busqueda = "",
    offset = 0,
    limit = PAGINA_CAMPANAS,
  ): Observable<{ items: Campana[]; total: number }> {
    const params: Record<string, string> = {
      limit: String(limit),
      offset: String(offset),
    };
    if (busqueda.trim()) params["q"] = busqueda.trim();
    return this.http
      .get<{ items: Campana[]; total: number }>(`${this.base}/publicas`, {
        params,
      })
      .pipe(catchError(this.handleError));
  }

  /** Todas las campañas, para la gestión (admin y voluntariado). */
  listar(
    busqueda = "",
    estado: EstadoCampana | "" = "",
  ): Observable<Campana[]> {
    const params: Record<string, string> = {};
    if (busqueda.trim()) params["q"] = busqueda.trim();
    if (estado) params["estado"] = estado;
    return this.http
      .get<Campana[]>(this.base, { ...this.opts, params })
      .pipe(catchError(this.handleError));
  }

  crear(input: CampanaInput): Observable<CampanaEscritura> {
    return this.http
      .post<CampanaEscritura>(this.base, input, this.opts)
      .pipe(catchError(this.handleError));
  }

  editar(id: string, input: CampanaInput): Observable<CampanaEscritura> {
    return this.http
      .put<CampanaEscritura>(
        `${this.base}/${encodeURIComponent(id)}`,
        input,
        this.opts,
      )
      .pipe(catchError(this.handleError));
  }

  publicar(id: string): Observable<CampanaEscritura> {
    return this.http
      .post<CampanaEscritura>(
        `${this.base}/${encodeURIComponent(id)}/publicar`,
        {},
        this.opts,
      )
      .pipe(catchError(this.handleError));
  }

  /**
   * Baja lógica: la campaña deja de verse en la gestión y en la landing. Solo admin y
   * voluntariado.
   */
  eliminar(id: string): Observable<{ mensaje: string }> {
    return this.http
      .delete<{ mensaje: string }>(
        `${this.base}/${encodeURIComponent(id)}`,
        this.opts,
      )
      .pipe(catchError(this.handleError));
  }

  private readonly handleError = (error: HttpErrorResponse) => {
    const body = error.error as { error?: string } | string | null;
    let message =
      "No se pudieron cargar las campañas. Probá de nuevo en unos minutos.";
    if (body && typeof body === "object" && body.error) {
      message = body.error;
    } else if (error.status === 401) {
      message = "Tu sesión expiró o no es válida. Volvé a iniciar sesión.";
    } else if (error.status === 403) {
      message = "No tenés permisos para gestionar campañas.";
    }
    return throwError(() => new Error(message));
  };
}
