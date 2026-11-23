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

interface Option {
  label: string;
  value: string;
}

/**
 * Historial de asistencia de los cursos del docente responsable
 * (HU-19). Solo consulta: la asistencia la generan las entregas de los
 * participantes; las actividades se gestionan desde el detalle del
 * curso (`/dashboard/cursos/:id`).
 */
@Component({
  selector: "app-asistencia",
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
  templateUrl: "./asistencia.component.html",
  styleUrls: ["./asistencia.component.css"],
})
export class AsistenciaComponent implements OnInit {
  private education = inject(EducationService);

  loadingCourses = true;
  error = "";

  courseOptions: Option[] = [];
  selectedCourseId: string | null = null;

  loadingMatrix = false;
  activities: Activity[] = [];
  rows: AttendanceMatrixRow[] = [];

  ngOnInit() {
    this.education.listMyCourses().subscribe({
      next: (courses: Course[]) => {
        this.courseOptions = courses.map((c) => ({
          label: c.title || c.id,
          value: c.id,
        }));
        this.loadingCourses = false;
        if (this.courseOptions.length === 1) {
          this.selectedCourseId = this.courseOptions[0].value;
          this.onCourseChange();
        }
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
      enrollments: this.education.listEnrollments(cursoId),
      activities: this.education.listActivities(cursoId),
    })
      .pipe(
        switchMap(({ enrollments, activities }) =>
          (activities.length
            ? forkJoin(
                activities.map((a) => this.education.listAttendance(a.id)),
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
          // El endpoint del docente ya trae nombre y apellido; si falta
          // alguno, la tabla muestra el id.
          const nombres = new Map<string, string>();
          for (const e of enrollments) {
            const nombre = `${e.nombre ?? ""} ${e.apellido ?? ""}`.trim();
            if (nombre) nombres.set(e.participanteId, nombre);
          }
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
