import { Injectable, inject } from "@angular/core";
import { HttpClient, HttpParams } from "@angular/common/http";
import { Observable } from "rxjs";
import { environment } from "../../environments/environment";
import {
  AreaFormativa,
  DashboardData,
  FiltrosDashboard,
} from "../core/models/dashboard.model";
import { fechaLocal } from "./dashboard.fechas";

/**
 * Indicadores y gráficas del dashboard (HU-52 a HU-54). El microservicio calcula todo y
 * devuelve el resultado listo; si una fuente falla, sus indicadores vienen con
 * `disponible: false`.
 */
@Injectable({ providedIn: "root" })
export class DashboardService {
  private http = inject(HttpClient);
  private readonly base = `${environment.apiUrl ?? ""}/api/v1/dashboard`;
  /**
   * Mismo patrón que AuthService y EducationService: en Docker la API está
   * en otro puerto (:8080) y el interceptor no agrega la cookie de sesión a
   * pedidos de otro origen, así que se pide explícitamente.
   */
  private readonly opts = { withCredentials: true };

  /** Áreas formativas, para el filtro (vienen de Educación). */
  areas(): Observable<AreaFormativa[]> {
    return this.http.get<AreaFormativa[]>(`${this.base}/areas`, this.opts);
  }

  /** Indicadores y gráficas ya calculados para los filtros dados. */
  cargar(filtros: FiltrosDashboard): Observable<DashboardData> {
    let params = new HttpParams()
      .set("desde", fechaLocal(filtros.desde))
      .set("hasta", fechaLocal(filtros.hasta))
      .set("modulo", filtros.modulo);
    if (filtros.areaId) params = params.set("area", filtros.areaId);
    return this.http.get<DashboardData>(`${this.base}/resumen`, {
      ...this.opts,
      params,
    });
  }
}
