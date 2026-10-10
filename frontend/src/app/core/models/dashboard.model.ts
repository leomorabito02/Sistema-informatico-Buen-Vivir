/**
 * Modelos del módulo "Dashboard y Reportes" (HU-52 a HU-55).
 *
 * El microservicio de Dashboard (dashboard/, Go) es un agregador: no tiene
 * datos propios y consulta en vivo a Autenticación, Educación e Inserción
 * Laboral. Las interfaces `*Origen` son lo que devuelve cada uno de sus
 * endpoints (GET /api/v1/dashboard/...), ya traducido a castellano.
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

// ── Datos de origen (lo que devuelve cada servicio) ───────────────────

/** Servicio de Autenticación: usuarios con rol participante. */
export interface ParticipanteOrigen {
  id: string;
  rol: "participante";
  estado: "activo" | "baja";
  fechaAlta: string; // ISO
  /** Fecha de la baja lógica, si la hubo. */
  fechaBaja?: string | null;
}

/** Servicio de Educación: curso con sus actividades. */
export interface CursoOrigen {
  id: string;
  titulo: string;
  areaId: string;
  /** "activo" = publicado; los borradores no cuentan como cursos. */
  estado: "activo" | "borrador";
  fechaCreacion: string; // ISO
  actividades: { id: string; fecha: string }[];
}

/** Servicio de Educación: registro Participante_Habilidad. */
export interface ParticipanteHabilidadOrigen {
  id: string;
  participanteId: string;
  cursoId: string;
  areaId: string;
  /** "actividad": la ganó al entregar; "manual": la cargó alguien a mano. */
  origen: "actividad" | "manual";
  estado: "aceptada" | "pendiente" | "rechazada";
  fecha: string; // ISO, cuando se registró
}

/** Servicio de Inserción Laboral: postulación a una oferta. */
export interface PostulacionOrigen {
  id: string;
  participanteId: string;
  estado: "postulado" | "en_revision" | "seleccionado" | "rechazado";
  /** Fecha del último cambio de estado. */
  fechaEstado: string; // ISO
}

// ── Indicadores y series ──────────────────────────────────────────────

export type IndicadorId =
  | "participantesActivos"
  | "cursosEnDesarrollo"
  | "cursosFinalizados"
  | "habilidadesDesarrolladas"
  | "insercionesLaborales";

export type FuenteDashboard = "autenticacion" | "educacion" | "insercion";

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
  "participantes" | "cursos" | "habilidades" | "inserciones";

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
  generado_por: string;
  fecha_generacion: string; // ISO
  url_archivo: string;
}
