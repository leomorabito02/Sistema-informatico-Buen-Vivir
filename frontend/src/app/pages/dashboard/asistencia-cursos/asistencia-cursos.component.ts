import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { RouterLink } from "@angular/router";
import { ButtonModule } from "primeng/button";
import { SelectModule } from "primeng/select";
import { SelectButtonModule } from "primeng/selectbutton";
import { TagModule } from "primeng/tag";
import { forkJoin, of } from "rxjs";
import { map, switchMap } from "rxjs/operators";

import { EducationService } from "../../../services/education.service";
import {
  Activity,
  Course,
  HistorialCurso,
  ParticipanteResumen,
} from "../../../core/models/education.model";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import {
  AttendanceMatrixComponent,
  AttendanceMatrixRow,
  buildAttendanceMatrixRows,
} from "../../../shared/components/attendance-matrix.component";
import {
  ParticipantPickerComponent,
  etiquetaParticipante,
} from "../../../shared/components/participant-picker.component";

interface Option {
  label: string;
  value: string;
}

type Vista = "curso" | "participante";

/**
 * Historial de asistencia (HU-19/RF-19) para role_admin y
 * role_volunteer, de solo consulta: la asistencia no se carga a mano, la
 * genera cada entrega de actividad del participante (ver
 * `/dashboard/cursos/:id`). Dos vistas:
 *  - por curso: matriz participante × actividad con el progreso de cada uno;
 *  - por participante: sus cursos, y en cada uno qué actividades asistió
 *    (`GET /api/participants/{id}/asistencia`).
 */
@Component({
  selector: "app-asistencia-cursos",
  standalone: true,
  imports: [
    LoadingStateComponent,
    ErrorStateComponent,
    EmptyStateComponent,
    CommonModule,
    FormsModule,
    RouterLink,
    ButtonModule,
    SelectModule,
    SelectButtonModule,
    TagModule,
    AttendanceMatrixComponent,
    ParticipantPickerComponent,
  ],
  templateUrl: "./asistencia-cursos.component.html",
  styleUrls: ["./asistencia-cursos.component.css"],
})
export class AsistenciaCursosComponent implements OnInit {
  private education = inject(EducationService);

  vista: Vista = "curso";
  readonly vistas: { label: string; value: Vista }[] = [
    { label: "Por curso", value: "curso" },
    { label: "Por participante", value: "participante" },
  ];

  loadingCourses = true;
  error = "";

  courseOptions: Option[] = [];
  selectedCourseId: string | null = null;

  loadingMatrix = false;
  activities: Activity[] = [];
  rows: AttendanceMatrixRow[] = [];

  participante: ParticipanteResumen | null = null;
  loadingHistorial = false;
  historial: HistorialCurso[] = [];

  ngOnInit() {
    this.education.listCourses().subscribe({
      next: (courses) => {
        this.courseOptions = courses.map((c: Course) => ({
          label: c.title || c.id,
          value: c.id,
        }));
        this.loadingCourses = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loadingCourses = false;
      },
    });
  }

  get nombreParticipante(): string {
    return this.participante ? etiquetaParticipante(this.participante) : "";
  }

  /** Historial del participante elegido en el buscador. */
  onParticipanteChange(p: ParticipanteResumen | null) {
    this.participante = p;
    this.historial = [];
    if (!p) return;
    this.loadingHistorial = true;
    this.error = "";
    this.education.getParticipantAttendanceHistory(p.participanteId).subscribe({
      next: (h) => {
        this.historial = h;
        this.loadingHistorial = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loadingHistorial = false;
      },
    });
  }

  estadoSeverity(estado: string): "success" | "info" | "danger" | "secondary" {
    switch (estado) {
      case "activa":
        return "success";
      case "finalizada":
        return "info";
      case "baja":
        return "danger";
      default:
        return "secondary";
    }
  }

  onCourseChange() {
    this.activities = [];
    this.rows = [];
    const cursoId = this.selectedCourseId;
    if (!cursoId) return;
    this.loadingMatrix = true;
    this.error = "";

    forkJoin({
      enrollments: this.education.listCourseEnrollments(cursoId),
      activities: this.education.listCourseActivities(cursoId),
    })
      .pipe(
        switchMap(({ enrollments, activities }) =>
          (activities.length
            ? forkJoin(
                activities.map((a) =>
                  this.education.listActivityAttendance(a.id),
                ),
              )
            : of([])
          ).pipe(
            map((porActividad) => ({
              enrollments,
              activities,
              attendance: porActividad.flat(),
            })),
          ),
        ),
      )
      .subscribe({
        next: ({ enrollments, activities, attendance }) => {
          this.activities = activities;
          // Los nombres los resuelve el backend en las inscripciones.
          const nombres = new Map(
            enrollments.map((e) => [e.participanteId, etiquetaParticipante(e)]),
          );
          this.rows = buildAttendanceMatrixRows(
            enrollments,
            attendance,
            nombres,
          );
          this.loadingMatrix = false;
        },
        error: (err: Error) => {
          this.error = err.message;
          this.loadingMatrix = false;
        },
      });
  }
}
