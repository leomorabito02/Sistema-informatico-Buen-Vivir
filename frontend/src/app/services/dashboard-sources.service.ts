import { Injectable, inject } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { Observable } from "rxjs";
import { environment } from "../../environments/environment";
import {
  AreaFormativa,
  CursoOrigen,
  ParticipanteHabilidadOrigen,
  ParticipanteOrigen,
  PostulacionOrigen,
  PromocionOrigen,
} from "../core/models/dashboard.model";

/**
 * Fuentes del dashboard: una llamada al microservicio de Dashboard por
 * cada servicio de origen (vía API Gateway, solo administración).
 *
 * El Dashboard es un agregador: consulta en vivo a Autenticación,
 * Educación e Inserción Laboral con la sesión del admin y devuelve los
 * datos ya traducidos a estos tipos. Si un servicio de origen falla, solo
 * falla su llamada y DashboardService marca ese indicador como
 * "No disponible".
 */
@Injectable({ providedIn: "root" })
export class DashboardSourcesService {
  private http = inject(HttpClient);
  private readonly base = `${environment.apiUrl ?? ""}/api/v1/dashboard`;
  /**
   * Mismo patrón que AuthService y EducationService: en Docker la API está
   * en otro puerto (:8080) y el interceptor no agrega la cookie de sesión a
   * pedidos de otro origen, así que se pide explícitamente.
   */
  private readonly opts = { withCredentials: true };

  /** Servicio de Autenticación: participantes con altas y bajas. */
  participantes(): Observable<ParticipanteOrigen[]> {
    return this.http.get<ParticipanteOrigen[]>(
      `${this.base}/participantes`,
      this.opts,
    );
  }

  /** Servicio de Educación: cursos con sus actividades. */
  cursos(): Observable<CursoOrigen[]> {
    return this.http.get<CursoOrigen[]>(`${this.base}/cursos`, this.opts);
  }

  /** Servicio de Educación: habilidades ganadas, con el área del curso. */
  habilidades(): Observable<ParticipanteHabilidadOrigen[]> {
    return this.http.get<ParticipanteHabilidadOrigen[]>(
      `${this.base}/habilidades`,
      this.opts,
    );
  }

  /** Servicio de Educación: áreas formativas. */
  areas(): Observable<AreaFormativa[]> {
    return this.http.get<AreaFormativa[]>(`${this.base}/areas`, this.opts);
  }

  /** Servicio de Promoción Humana: fichas e hitos (solo ids y fechas). */
  promocion(): Observable<PromocionOrigen> {
    return this.http.get<PromocionOrigen>(`${this.base}/promocion`, this.opts);
  }

  /** Servicio de Inserción Laboral: postulaciones de todas las ofertas. */
  postulaciones(): Observable<PostulacionOrigen[]> {
    return this.http.get<PostulacionOrigen[]>(
      `${this.base}/postulaciones`,
      this.opts,
    );
  }
}
