import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { Injectable, inject } from "@angular/core";
import { Observable, catchError, map, throwError } from "rxjs";

import { environment } from "../../environments/environment";
import {
  FiltroAuditoria,
  RegistroAuditoria,
} from "../core/models/education.model";
import { filtrosAuditoria } from "./education.service";

/** Fila tal como la devuelven Autenticación e Inserción laboral. */
interface RegistroAuditoriaApi {
  id: string | number;
  usuario_id: string;
  usuario_nombre?: string;
  rol?: string;
  accion: string;
  recurso: string;
  recurso_id: string;
  campos?: string;
  resultado?: string;
  fecha: string;
  ip?: string;
}

/**
 * Auditoría de Autenticación (HU-10, más rechazos de HU-03 y HU-06) y de
 * Inserción laboral (HU-42). Mismo formato de pantalla que Educación y
 * Promoción Humana.
 */
@Injectable({ providedIn: "root" })
export class AuditoriaService {
  private readonly http = inject(HttpClient);
  private readonly base = environment.apiUrl;

  listarUsuarios(
    filtro: FiltroAuditoria = {},
  ): Observable<RegistroAuditoria[]> {
    return this.listar(`${this.base}/api/v1/auth/auditoria`, filtro);
  }

  listarLaboral(filtro: FiltroAuditoria = {}): Observable<RegistroAuditoria[]> {
    return this.listar(
      `${this.base}/api/v1/labor-integration/auditoria`,
      filtro,
    );
  }

  /** Auditoría de Campañas (D8, 22/11/2026): alta, edición, publicación y baja. */
  listarCampanas(
    filtro: FiltroAuditoria = {},
  ): Observable<RegistroAuditoria[]> {
    return this.listar(`${this.base}/api/v1/campaigns/auditoria`, filtro);
  }

  /** Acciones de gestión sobre CV de participantes (D8): ver, editar, cambiar foto. */
  listarCv(filtro: FiltroAuditoria = {}): Observable<RegistroAuditoria[]> {
    return this.listar(`${this.base}/api/v1/rxresume/auditoria`, filtro);
  }

  private listar(
    url: string,
    filtro: FiltroAuditoria,
  ): Observable<RegistroAuditoria[]> {
    return this.http
      .get<RegistroAuditoriaApi[]>(url, {
        params: filtrosAuditoria(filtro),
        withCredentials: true,
      })
      .pipe(
        map((rows) => (rows ?? []).map(aRegistro)),
        catchError((err: HttpErrorResponse) =>
          throwError(
            () =>
              new Error(
                err.status === 403
                  ? "Solo administración puede ver la auditoría."
                  : "No se pudo cargar la auditoría.",
              ),
          ),
        ),
      );
  }
}

export function aRegistro(r: RegistroAuditoriaApi): RegistroAuditoria {
  const detalle =
    r.resultado === "denegado"
      ? `Denegado: ${r.campos ?? ""}`.trim()
      : r.campos
        ? `Campos: ${r.campos.split(",").join(", ")}`
        : undefined;
  return {
    id: String(r.id),
    usuarioId: r.usuario_id,
    usuarioNombre: r.usuario_nombre,
    accion: r.accion,
    recurso: r.recurso,
    recursoId: r.recurso_id,
    fecha: r.fecha,
    ip: r.ip,
    detalle,
  };
}
