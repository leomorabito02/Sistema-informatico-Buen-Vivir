/**
 * Modelos del microservicio de Promoción Humana (human-promotion).
 * Las fechas de hitos viajan como "YYYY-MM-DD" y se muestran DD/MM/AAAA.
 * Ver human-promotion/openapi.yaml.
 */

/** Participante encontrado por nombre (HU-36). `ficha_id` null = sin ficha. */
export interface ResultadoBusqueda {
  participante_id: string;
  nombre: string;
  apellido: string;
  ficha_id: string | null;
}

/** Datos de solo lectura que vienen de auth-service (HU-31). */
export interface DatosParticipante {
  nombre: string;
  apellido: string;
  fecha_nacimiento?: string;
  edad?: number;
}

export interface Ficha {
  id: string;
  participante_id: string;
  situacion_actual: string;
  avances: string;
  creado_por: string;
  updated_at: string;
}

/** Ficha completa: `participante` es null si auth-service no respondió. */
export interface FichaDetalle extends Ficha {
  participante: DatosParticipante | null;
  creado_por_nombre?: string;
  habilidades: string[];
}

export interface CreateFichaInput {
  participante_id: string;
  situacion_actual: string;
  avances: string;
}

export interface UpdateFichaInput {
  situacion_actual: string;
  avances: string;
}

export interface UpdateFichaResponse {
  message: string;
  ficha: Ficha;
}

export interface Hito {
  id: string;
  ficha_id: string;
  fecha: string;
  descripcion: string;
  responsable_id: string;
  visible_participante: boolean;
}

/** Hito de la línea de tiempo profesional (HU-33). */
export interface HitoDetalle extends Hito {
  responsable_nombre?: string;
}

export interface HitoInput {
  fecha: string;
  descripcion: string;
  responsable_id?: string;
  visible_participante?: boolean;
}

/** Lo único que ve el participante de un hito (HU-34). */
export interface HitoParticipante {
  id: string;
  fecha: string;
  descripcion: string;
}

/** Mensaje exacto de HU-29 cuando falta un campo obligatorio. */
export const MSG_CAMPOS_OBLIGATORIOS =
  "Completar todos los campos obligatorios para continuar";

/** Límites de texto, iguales a los del backend. */
export const MAX_TEXTO_FICHA = 5000;
export const MAX_DESCRIPCION_HITO = 2000;
