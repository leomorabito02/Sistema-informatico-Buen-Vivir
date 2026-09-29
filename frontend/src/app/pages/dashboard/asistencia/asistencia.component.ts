import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { RouterLink } from "@angular/router";
import { ButtonModule } from "primeng/button";
import { SelectModule } from "primeng/select";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { MessageModule } from "primeng/message";
import { forkJoin, of } from "rxjs";
import { map, switchMap } from "rxjs/operators";

import { EducationService } from "../../../services/education.service";
import { Activity, Course } from "../../../core/models/education.model";
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
    CommonModule,
    FormsModule,
    RouterLink,
    ButtonModule,
    SelectModule,
    ProgressSpinnerModule,
    MessageModule,
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
          // auth-service no le resuelve nombres de participantes al rol
          // docente: se muestran por id.
          this.rows = buildAttendanceMatrixRows(
            enrollments,
            attendance,
            new Map(),
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
