import { Observable, of } from "rxjs";
import { catchError, map } from "rxjs/operators";

import { EntradaActividad } from "../../core/models/human-promotion.model";
import { EducationService } from "../../services/education.service";

/** Un elemento de la línea de tiempo: un hito del equipo o actividad en el centro. */
export type ItemLineaTiempo<H> =
  | { tipo: "hito"; key: string; fecha: string; hito: H }
  | {
      tipo: "actividad";
      key: string;
      fecha: string;
      actividad: EntradaActividad;
      /** Nombre del curso, si se conoce (vive en Educación). */
      curso?: string;
    };

/**
 * Mezcla los hitos con la actividad en el centro en una sola línea, por
 * fecha (HU-33, HU-34). En la misma fecha va primero la actividad: una
 * inscripción suele ser lo que después comenta el equipo.
 */
export function mezclarLineaDeTiempo<H extends { id: string; fecha: string }>(
  hitos: H[],
  actividad: EntradaActividad[],
  cursos: ReadonlyMap<string, string> = new Map(),
): ItemLineaTiempo<H>[] {
  const items: ItemLineaTiempo<H>[] = [
    ...actividad.map((a) => ({
      tipo: "actividad" as const,
      key: `a-${a.id}`,
      fecha: a.fecha,
      actividad: a,
      curso: a.curso_id ? cursos.get(a.curso_id) : undefined,
    })),
    ...hitos.map((h) => ({
      tipo: "hito" as const,
      key: `h-${h.id}`,
      fecha: h.fecha,
      hito: h,
    })),
  ];
  // sort es estable: en la misma fecha se respeta el orden de arriba.
  return items.sort((a, b) => a.fecha.localeCompare(b.fecha));
}

/**
 * id -> título de los cursos publicados, para mostrar el nombre en la
 * actividad. Es un dato secundario: si Educación falla, sale vacío y la
 * línea de tiempo se muestra igual, sin nombres de curso.
 */
export function nombresDeCursos(
  education: EducationService,
): Observable<Map<string, string>> {
  return education.listCourses().pipe(
    map((cursos) => new Map(cursos.map((c) => [c.id, c.title]))),
    catchError(() => of(new Map<string, string>())),
  );
}
