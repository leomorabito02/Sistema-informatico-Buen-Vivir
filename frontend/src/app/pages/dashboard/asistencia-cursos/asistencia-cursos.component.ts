import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { RouterLink } from "@angular/router";
import { ButtonModule } from "primeng/button";
import { SelectModule } from "primeng/select";
import { forkJoin, of } from "rxjs";
import { map, switchMap } from "rxjs/operators";

import { EducationService } from "../../../services/education.service";
import { Activity, Course } from "../../../core/models/education.model";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import {
  AttendanceMatrixComponent,
  AttendanceMatrixRow,
  buildAttendanceMatrixRows,
} from "../../../shared/components/attendance-matrix.component";
import { etiquetaParticipante } from "../../../shared/components/participant-picker.component";

interface Option {
  label: string;
  value: string;
}

/**
 * Historial de asistencia (HU-19) para admin y voluntario, solo consulta: la asistencia
 * la genera cada entrega. Muestra la matriz participante × actividad del curso elegido.
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
    AttendanceMatrixComponent,
  ],
  templateUrl: "./asistencia-cursos.component.html",
  styleUrls: ["./asistencia-cursos.component.css"],
})
export class AsistenciaCursosComponent implements OnInit {
  private education = inject(EducationService);

  loadingCourses = true;
  error = "";

  courseOptions: Option[] = [];
  selectedCourseId: string | null = null;

  loadingMatrix = false;
  activities: Activity[] = [];
  rows: AttendanceMatrixRow[] = [];

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
