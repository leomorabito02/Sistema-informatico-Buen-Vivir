/**
 * Modelos del Dashboard (HU-52 a HU-55). El microservicio calcula los indicadores y las
 * series por mes y devuelve el resultado listo (GET /api/v1/dashboard/resumen).
 */

// ── Filtros ────────────────────────────────────────────────────────────

export type ModuloDashboard =
  "todos" | "usuarios" | "educacion" | "promocion" | "insercion";

export interface FiltrosDashboard {
  desde: Date;
  hasta: Date;
  modulo: ModuloDashboard;
  /** null = todas las áreas formativas. */
  areaId: string | null;
}

export interface AreaFormativa {
  id: string;
  nombre: string;
}

// ── Indicadores y series ──────────────────────────────────────────────

export type IndicadorId =
  | "participantesActivos"
  | "empresasActivas"
  | "docentesActivos"
  | "voluntariosActivos"
  | "cursosEnDesarrollo"
  | "cursosFinalizados"
  | "habilidadesDesarrolladas"
  | "insercionesLaborales"
  | "fichasActivas"
  | "participantesConSeguimiento";

export type FuenteDashboard =
  "autenticacion" | "educacion" | "insercion" | "promocion";

export interface Indicador {
  id: IndicadorId;
  titulo: string;
  descripcion: string;
  modulo: Exclude<ModuloDashboard, "todos">;
  fuente: FuenteDashboard;
  /** null cuando el servicio de origen no respondió. */
  valor: number | null;
  disponible: boolean;
}

/** Un punto por mes, "YYYY-MM". */
export interface PuntoSerie {
  mes: string;
  valor: number;
}

export type GraficaId =
  "participantes" | "cursos" | "habilidades" | "inserciones" | "promocion";

export interface SerieGrafica {
  nombre: string;
  puntos: PuntoSerie[];
}

export interface Grafica {
  id: GraficaId;
  titulo: string;
  tipo: "line" | "bar";
  modulo: Exclude<ModuloDashboard, "todos">;
  indicadores: IndicadorId[];
  /** null cuando el servicio de origen no respondió. */
  series: SerieGrafica[] | null;
}

export interface DashboardData {
  indicadores: Indicador[];
  graficas: Grafica[];
}

// ── Reportes exportados (HU-55) ───────────────────────────────────────

export interface FiltrosAplicados {
  modulo: ModuloDashboard;
  areaId: string | null;
  areaNombre: string | null;
}

export interface ReporteExportado {
  id: string;
  tipo_reporte: "dashboard";
  periodo_desde: string; // ISO (fecha)
  periodo_hasta: string; // ISO (fecha)
  filtros_aplicados: FiltrosAplicados;
  /** Id del usuario que lo exportó (FK lógica a Autenticación, DER). */
  generado_por: string;
  /** Nombre para mostrar en el historial. */
  generado_por_nombre: string;
  fecha_generacion: string; // ISO
  /** Ubicación del PDF en Cloud Storage (gs://...). */
  url_archivo: string;
}
