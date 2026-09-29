import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { RouterLink } from "@angular/router";
import { ButtonModule } from "primeng/button";
import { SelectModule } from "primeng/select";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { MessageModule } from "primeng/message";
import { forkJoin, of } from "rxjs";
import { catchError, map, switchMap } from "rxjs/operators";

import { EducationService } from "../../../services/education.service";
import { Activity, Course } from "../../../core/models/education.model";
import { todayYMD } from "../../../shared/utils/date-ymd";
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
 * Historial de asistencia por curso (HU-19/RF-19) para role_admin y
 * role_volunteer, sobre cualquier curso del centro. Es de solo
 * consulta: la asistencia no se carga a mano, la genera cada entrega de
 * actividad del participante (ver `/dashboard/cursos/:id`). Muestra una
 * matriz participante × actividad con el progreso de cada uno.
 */
@Component({
  selector: "app-asistencia-cursos",
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
  templateUrl: "./asistencia-cursos.component.html",
  styleUrls: ["./asistencia-cursos.component.css"],
})
export class AsistenciaCursosComponent implements OnInit {
  private education = inject(EducationService);

  loadingCourses = true;
  error = "";

  courseOptions: Option[] = [];
  selectedCourseId: string | null = null;
  private participantNames = new Map<string, string>();

  loadingMatrix = false;
  activities: Activity[] = [];
  rows: AttendanceMatrixRow[] = [];

  ngOnInit() {
    forkJoin({
      courses: this.education.listCourses(),
      // Directorio de nombres: reutiliza el listado ya resuelto de la
      // asistencia general (participantes aprobados con nombre).
      directory: this.education
        .getGeneralAttendance(todayYMD())
        .pipe(catchError(() => of([]))),
    }).subscribe({
      next: ({ courses, directory }) => {
        this.courseOptions = courses.map((c: Course) => ({
          label: c.title || c.id,
          value: c.id,
        }));
        this.participantNames = new Map(
          directory.map((p) => [
            p.participanteId,
            `${p.nombre ?? ""} ${p.apellido ?? ""}`.trim() || p.participanteId,
          ]),
        );
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
          this.rows = buildAttendanceMatrixRows(
            enrollments,
            attendance,
            this.participantNames,
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
