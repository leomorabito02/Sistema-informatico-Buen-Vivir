import {
  CursoOrigen,
  FichaPromocionOrigen,
  HitoPromocionOrigen,
  ParticipanteHabilidadOrigen,
  ParticipanteOrigen,
  PostulacionOrigen,
  PuntoSerie,
} from "../core/models/dashboard.model";

/**
 * Cálculo de los indicadores, en funciones puras (sin Angular) para poder
 * probarlas y ajustar las reglas de negocio en un solo lugar.
 */

export type EstadoCurso = "en_desarrollo" | "finalizado" | "fuera";

/**
 * REGLA DE NEGOCIO (aislada a propósito, para ajustarla sin tocar el resto):
 *
 * A la fecha de referencia `ref`, un curso
 * - no cuenta ("fuera") si es un borrador o todavía no existía;
 * - está "finalizado" si tiene actividades y todas ya ocurrieron;
 * - está "en_desarrollo" si tiene alguna actividad pendiente, o si todavía
 *   no tiene actividades cargadas (está activo y en preparación).
 */
export function estadoCurso(curso: CursoOrigen, ref: Date): EstadoCurso {
  if (curso.estado !== "activo") return "fuera";
  if (new Date(curso.fechaCreacion) > ref) return "fuera";
  if (curso.actividades.length === 0) return "en_desarrollo";
  const pendiente = curso.actividades.some((a) => new Date(a.fecha) > ref);
  return pendiente ? "en_desarrollo" : "finalizado";
}

/** Participante activo a la fecha `ref` (excluye bajas lógicas). */
export function participanteActivoEn(
  p: ParticipanteOrigen,
  ref: Date,
): boolean {
  if (p.rol !== "participante") return false;
  if (new Date(p.fechaAlta) > ref) return false;
  if (p.estado === "baja") {
    // Sin fecha de baja no se puede saber desde cuándo: no cuenta.
    return !!p.fechaBaja && new Date(p.fechaBaja) > ref;
  }
  return true;
}

/** Solo las habilidades aceptadas por el administrador (IA o manual). */
export function habilidadAceptada(h: ParticipanteHabilidadOrigen): boolean {
  return h.estado === "aceptada";
}

export function insercionLaboral(p: PostulacionOrigen): boolean {
  return p.estado === "seleccionado";
}

/**
 * Ficha de proceso personal abierta a la fecha `ref`: ya existía y no
 * estaba dada de baja (mismo criterio que participanteActivoEn).
 */
export function fichaActivaEn(f: FichaPromocionOrigen, ref: Date): boolean {
  if (new Date(f.fechaAlta) > ref) return false;
  if (f.estado === "baja") {
    return !!f.fechaBaja && new Date(f.fechaBaja) > ref;
  }
  return true;
}

/**
 * Participantes distintos con al menos un hito de avance entre `desde` y
 * `hasta` (inclusive). Las fechas de los hitos son AAAA-MM-DD sin hora: se
 * comparan como texto para que la zona horaria no corra el día.
 */
export function participantesConSeguimiento(
  hitos: HitoPromocionOrigen[],
  desde: Date,
  hasta: Date,
): number {
  const d = fechaLocal(desde);
  const h = fechaLocal(hasta);
  return new Set(
    hitos.filter((x) => x.fecha >= d && x.fecha <= h).map((x) => x.participanteId),
  ).size;
}

/** "AAAA-MM-DD" de una fecha local. */
export function fechaLocal(d: Date): string {
  return `${claveMes(d)}-${String(d.getDate()).padStart(2, "0")}`;
}

// ── Fechas ─────────────────────────────────────────────────────────────

export function inicioDelDia(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate());
}

export function finDelDia(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate(), 23, 59, 59, 999);
}

export function enRango(iso: string, desde: Date, hasta: Date): boolean {
  const t = new Date(iso).getTime();
  return t >= inicioDelDia(desde).getTime() && t <= finDelDia(hasta).getTime();
}

/** "YYYY-MM" de una fecha local. */
export function claveMes(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`;
}

/**
 * Meses del período, cada uno con la fecha que lo representa: el último
 * día del mes, o `hasta` en el último mes (no se mira más allá del filtro).
 */
export function mesesDelPeriodo(
  desde: Date,
  hasta: Date,
): { mes: string; inicio: Date; corte: Date }[] {
  const meses: { mes: string; inicio: Date; corte: Date }[] = [];
  if (desde > hasta) return meses;
  const fin = finDelDia(hasta);
  let cursor = new Date(desde.getFullYear(), desde.getMonth(), 1);
  while (cursor <= fin) {
    const ultimoDia = new Date(
      cursor.getFullYear(),
      cursor.getMonth() + 1,
      0,
      23,
      59,
      59,
      999,
    );
    const inicio = cursor < desde ? inicioDelDia(desde) : cursor;
    meses.push({
      mes: claveMes(cursor),
      inicio,
      corte: ultimoDia < fin ? ultimoDia : fin,
    });
    cursor = new Date(cursor.getFullYear(), cursor.getMonth() + 1, 1);
  }
  return meses;
}

/** Cuenta por mes los elementos cuya fecha cae dentro de cada mes. */
export function contarPorMes<T>(
  items: T[],
  fecha: (item: T) => string,
  desde: Date,
  hasta: Date,
): PuntoSerie[] {
  return mesesDelPeriodo(desde, hasta).map(({ mes, inicio, corte }) => ({
    mes,
    valor: items.filter((i) => {
      const t = new Date(fecha(i)).getTime();
      return t >= inicio.getTime() && t <= corte.getTime();
    }).length,
  }));
}

/** "2026-03" → "mar 2026" */
export function etiquetaMes(mes: string): string {
  const [anio, m] = mes.split("-").map(Number);
  const nombre = new Date(anio, m - 1, 1).toLocaleDateString("es-AR", {
    month: "short",
  });
  return `${nombre.replace(".", "")} ${anio}`;
}

/** DD/MM/AAAA */
export function formatoFecha(d: Date): string {
  const dd = String(d.getDate()).padStart(2, "0");
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  return `${dd}/${mm}/${d.getFullYear()}`;
}
