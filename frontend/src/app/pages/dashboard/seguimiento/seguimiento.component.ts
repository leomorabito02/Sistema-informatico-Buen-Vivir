import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { TagModule } from "primeng/tag";
import { CardModule } from "primeng/card";
import { SelectModule } from "primeng/select";
import { ProgressBarModule } from "primeng/progressbar";
import { InputTextModule } from "primeng/inputtext";
import { IconFieldModule } from "primeng/iconfield";
import { InputIconModule } from "primeng/inputicon";
import { Observable, forkJoin, from, of } from "rxjs";
import { catchError, map, mergeMap, toArray } from "rxjs/operators";
import { EducationService } from "../../../services/education.service";
import { AuthService } from "../../../services/auth.service";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import { coincideBusqueda } from "../../../shared/utils/busqueda";
import {
  Course,
  Enrollment,
  ParticipantProgress,
} from "../../../core/models/education.model";

const SILENT = { silent: true };
const CURSOS_EN_PARALELO = 3;

interface Option {
  label: string;
  value: string;
}

export type Riesgo = "Sin datos" | "En riesgo" | "Regular" | "Buen avance";

export interface SeguimientoRow {
  /** participante + curso: una fila por inscripción. */
  key: string;
  participanteId: string;
  /** Nombre y apellido, si el backend los resolvió. */
  nombre?: string;
  cursoId: string;
  curso: string;
  estado: string;
  porcentajeAsistencia: number;
  actividadesAsistidas: number;
  actividadesTotal: number;
  riesgo: Riesgo;
}

/** Valor del selector para ver todos los cursos. */
export const TODOS_LOS_CURSOS = "";

/** Riesgo según el porcentaje de asistencia (mismo criterio de siempre). */
export function riesgoDe(p: ParticipantProgress | undefined): Riesgo {
  if (!p) return "Sin datos";
  const pct = p.porcentajeAsistencia ?? 0;
  if (pct < 50) return "En riesgo";
  if (pct < 75) return "Regular";
  return "Buen avance";
}

/**
 * Seguimiento (HU-20): asistencia y avance de cada participante en sus
 * cursos. Al entrar se ven todos los participantes inscriptos (una fila
 * por inscripción); el curso es un filtro opcional y el buscador filtra
 * por nombre sin cambiar de pantalla (pedido del 17/11/2026).
 *
 * Admin y voluntario ven cualquier curso del catálogo; el docente, solo
 * los suyos (/api/docente). El backend valida el acceso a cada curso.
 */
@Component({
  selector: "app-seguimiento",
  standalone: true,
  imports: [
    LoadingStateComponent,
    ErrorStateComponent,
    EmptyStateComponent,
    CommonModule,
    FormsModule,
    TableModule,
    TagModule,
    CardModule,
    SelectModule,
    ProgressBarModule,
    InputTextModule,
    IconFieldModule,
    InputIconModule,
  ],
  templateUrl: "./seguimiento.component.html",
  styleUrls: ["./seguimiento.component.css"],
})
export class SeguimientoComponent implements OnInit {
  private education = inject(EducationService);
  private auth = inject(AuthService);

  gestion = false;

  loading = true;
  error = "";
  /** Cursos cuya información no se pudo leer (el resto se muestra igual). */
  cursosConError: string[] = [];
  cursos: Course[] = [];
  courseOptions: Option[] = [];
  selectedCourseId: string = TODOS_LOS_CURSOS;
  busqueda = "";
  rows: SeguimientoRow[] = [];

  ngOnInit() {
    this.auth.isAuthenticated().subscribe({
      next: (res) => {
        const rol = res.role_id ?? res.user?.role_id ?? "";
        this.gestion = ["role_admin", "role_dev", "role_volunteer"].includes(
          rol,
        );
        this.cargar();
      },
      error: () => {
        this.error = "No se pudo verificar tu sesión. Volvé a iniciar sesión.";
        this.loading = false;
      },
    });
  }

  /** Filas visibles según el curso elegido y el texto del buscador. */
  get filasFiltradas(): SeguimientoRow[] {
    return this.rows.filter(
      (r) =>
        (this.selectedCourseId === TODOS_LOS_CURSOS ||
          r.cursoId === this.selectedCourseId) &&
        coincideBusqueda(this.busqueda, r.nombre, r.participanteId),
    );
  }

  get hayFiltros(): boolean {
    return this.selectedCourseId !== TODOS_LOS_CURSOS || !!this.busqueda.trim();
  }

  limpiarFiltros() {
    this.selectedCourseId = TODOS_LOS_CURSOS;
    this.busqueda = "";
  }

  cargar() {
    this.loading = true;
    this.error = "";
    this.cursosConError = [];
    // silent: si falla, se muestra el error en esta pantalla (con
    // Reintentar) en vez de saltar a la página de error del servidor, y
    // las fallas pasajeras se reintentan solas (22/11/2026).
    const cursos$: Observable<Course[]> = this.gestion
      ? this.education.listCourses("", "", SILENT)
      : this.education.listMyCourses(SILENT);
    cursos$.subscribe({
      next: (cursos) => {
        this.cursos = cursos;
        this.courseOptions = [
          { label: "Todos los cursos", value: TODOS_LOS_CURSOS },
          ...cursos.map((c) => ({ label: c.title || c.id, value: c.id })),
        ];
        if (cursos.length === 0) {
          this.rows = [];
          this.loading = false;
          return;
        }
        // Antes salían 2 pedidos por curso todos juntos; con muchos cursos
        // Cloud Run respondía 5xx hasta "calentarse". Ahora van de a
        // CURSOS_EN_PARALELO.
        from(cursos)
          .pipe(
            mergeMap((c) => this.filasDelCurso(c), CURSOS_EN_PARALELO),
            toArray(),
          )
          .subscribe((porCurso) => {
            this.rows = porCurso
              .flat()
              .sort(
                (a, b) =>
                  (a.nombre ?? a.participanteId).localeCompare(
                    b.nombre ?? b.participanteId,
                    "es",
                  ) || a.curso.localeCompare(b.curso, "es"),
              );
            this.loading = false;
          });
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loading = false;
      },
    });
  }

  /**
   * Inscripciones y progreso de un curso. Si ese curso falla se anota y
   * se siguen mostrando los demás.
   */
  private filasDelCurso(c: Course): Observable<SeguimientoRow[]> {
    const enrollments$: Observable<Enrollment[]> = this.gestion
      ? this.education.listCourseEnrollments(c.id, SILENT)
      : this.education.listEnrollments(c.id, SILENT);
    const progress$: Observable<ParticipantProgress[]> = this.gestion
      ? this.education.getCourseProgressGestion(c.id, SILENT)
      : this.education.getCourseProgress(c.id, SILENT);
    const curso = c.title || c.id;
    return forkJoin({ enrollments: enrollments$, progress: progress$ }).pipe(
      map(({ enrollments, progress }) => {
        const byId = new Map(progress.map((p) => [p.participanteId, p]));
        return enrollments.map((e) => {
          const p = byId.get(e.participanteId);
          const nombre = `${e.nombre ?? ""} ${e.apellido ?? ""}`.trim();
          return {
            key: `${e.participanteId}_${c.id}`,
            participanteId: e.participanteId,
            nombre: nombre || undefined,
            cursoId: c.id,
            curso,
            estado: e.estado,
            // Entero: la barra mostraba "66.666…%".
            porcentajeAsistencia: Math.round(p?.porcentajeAsistencia ?? 0),
            actividadesAsistidas: p?.actividadesAsistidas ?? 0,
            actividadesTotal: p?.actividadesTotal ?? 0,
            riesgo: riesgoDe(p),
          };
        });
      }),
      catchError(() => {
        this.cursosConError = [...this.cursosConError, curso];
        return of([] as SeguimientoRow[]);
      }),
    );
  }

  riesgoSeverity(r: string): string {
    switch (r) {
      case "Buen avance":
        return "success";
      case "Regular":
        return "info";
      case "En riesgo":
        return "danger";
      default:
        return "secondary";
    }
  }
}
