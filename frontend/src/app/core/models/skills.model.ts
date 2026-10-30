/**
 * Habilidades (HU-21 / HU-58). Espejo de education/internal/models/skill.go
 * y skill_suggestion.go.
 */

/** Lista cerrada de categorías (decisión del equipo, 2026-10-06). */
export const SKILL_CATEGORIES = [
  "Tecnología",
  "Cuidado de la naturaleza",
  "Oficio",
  "Socioemocional",
] as const;

export type SkillCategory = (typeof SKILL_CATEGORIES)[number];

/** Habilidad del catálogo. `id` es el nombre normalizado. */
export interface Skill {
  id: string;
  nombre: string;
  categoria: SkillCategory;
}

/**
 * Habilidad tal como se manda al guardar las de una actividad: puede ser
 * una del catálogo (con id) o una nueva (solo nombre y categoría). El
 * backend decide si ya existe por el nombre, no por el id.
 */
export interface SkillInput {
  id?: string;
  nombre: string;
  categoria: SkillCategory;
}

/** Respuesta de GET /api/cursos/{id}/habilidades. */
export interface CourseSkills {
  habilidades: Skill[];
  porActividad: Record<string, Skill[]>;
}

/** Body de POST /api/actividades/sugerir-habilidades. */
export interface SkillSuggestionRequest {
  cursoId: string;
  actividadId?: string;
  titulo: string;
  consigna: string;
  /** Nombres que ya están en el formulario: la IA no los vuelve a ofrecer. */
  excluir?: string[];
}

/** Una habilidad propuesta por la IA. */
export interface SkillSuggestion {
  /** Viene cuando coincide con una habilidad que ya existe en el catálogo. */
  id?: string;
  nombre: string;
  categoria: SkillCategory;
  justificacion: string;
  existente: boolean;
}

export interface SkillSuggestionResponse {
  /** Qué IA respondió: "gemini" (principal) u "openai" (respaldo). */
  proveedor: string;
  sugerencias: SkillSuggestion[];
}

/** Habilidad ganada por un participante. */
export interface ParticipantSkill extends Skill {
  origen: "actividad" | "manual";
  fecha: string;
}

/**
 * Clave normalizada de un nombre: misma regla que SkillKey del backend
 * (minúsculas, sin tildes ni signos, palabras unidas por guiones). Sirve
 * para no mostrar dos veces la misma habilidad escrita distinto.
 */
export function skillKey(nombre: string): string {
  return nombre
    .toLowerCase()
    .normalize("NFD")
    .replace(/[̀-ͯ]/g, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}
