import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { ButtonModule } from "primeng/button";
import { CardModule } from "primeng/card";
import { SelectModule } from "primeng/select";
import { InputTextModule } from "primeng/inputtext";
import { CheckboxModule } from "primeng/checkbox";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { MessageModule } from "primeng/message";
import { ToastModule } from "primeng/toast";
import { MessageService } from "primeng/api";
import { forkJoin, of } from "rxjs";
import { catchError } from "rxjs/operators";

import { EducationService } from "../../../services/education.service";
import { Activity, Course } from "../../../core/models/education.model";
import { todayYMD } from "../../../shared/utils/date-ymd";

interface Option {
  label: string;
  value: string;
}
interface AttendanceRow {
  participanteId: string;
  nombreCompleto: string;
  presente: boolean;
}

/**
 * Asistencia por actividad (HU-19/RF-19) para role_admin y
 * role_volunteer, sin el scope "mis cursos" de la pantalla del docente
 * (`/dashboard/asistencia`). Usa `POST/GET /api/actividades/{id}/asistencia`.
 *
 * A diferencia de la pantalla del docente (que carga participantes a
 * mano por id), acá se listan los inscriptos reales del curso
 * (`GET /api/cursos/{id}/inscripciones`) con nombre resuelto, así se
 * tilda presente/ausente en vez de tipear ids.
 *
 * Limitación conocida: el backend no tiene un endpoint para listar las
 * actividades de un curso abierto a role_volunteer (solo existe
 * `GET /api/docente/cursos/{cursoId}/actividades`, exclusivo del
 * docente responsable — role_admin lo alcanza igual por el bypass
 * global de rol, pero role_volunteer no). Para no bloquear a los
 * voluntarios, si ese listado falla por permisos esta pantalla cae a
 * carga manual del id de actividad (mismo criterio que ya usaba la
 * pantalla del docente antes de existir el módulo de inscripciones).
 */
@Component({
  selector: "app-asistencia-cursos",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    TableModule,
    ButtonModule,
    CardModule,
    SelectModule,
    InputTextModule,
    CheckboxModule,
    ProgressSpinnerModule,
    MessageModule,
    ToastModule,
  ],
  providers: [MessageService],
  templateUrl: "./asistencia-cursos.component.html",
  styleUrls: ["./asistencia-cursos.component.css"],
})
export class AsistenciaCursosComponent implements OnInit {
  private education = inject(EducationService);
  private toast = inject(MessageService);

  loadingCourses = true;
  error = "";

  courseOptions: Option[] = [];
  selectedCourseId: string | null = null;
  private participantNames = new Map<string, string>();

  loadingActivities = false;
  activityOptions: Option[] = [];
  /** true = se pudo listar actividades del curso; false = fallback manual. */
  activityListAvailable = true;
  manualActividadId = "";

  selectedActivityId: string | null = null;
  loadingAttendance = false;
  rows: AttendanceRow[] = [];
  guardando = false;

  ngOnInit() {
    forkJoin({
      courses: this.education.listCourses(),
      // directorio de nombres, ver el comentario de la clase — reutiliza
      // el mismo endpoint resuelto que usa la pantalla de inscripciones.
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
            `${p.nombre ?? ""} ${p.apellido ?? ""}`.trim() ||
              p.participanteId,
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
    this.selectedActivityId = null;
    this.manualActividadId = "";
    this.rows = [];
    if (!this.selectedCourseId) return;
    this.loadingActivities = true;
    this.error = "";
    const cursoId = this.selectedCourseId;

    this.education
      .listActivities(cursoId)
      .pipe(catchError(() => of(null)))
      .subscribe((activities: Activity[] | null) => {
        if (activities) {
          this.activityListAvailable = true;
          this.activityOptions = activities.map((a) => ({
            label: `${a.titulo} · ${new Date(a.fecha).toLocaleString()}`,
            value: a.id,
          }));
        } else {
          this.activityListAvailable = false;
          this.activityOptions = [];
        }
        this.loadingActivities = false;
      });
  }

  useManualActivity() {
    const id = this.manualActividadId.trim();
    if (!id) return;
    this.selectActivity(id);
  }

  selectActivity(id: string) {
    if (!this.selectedCourseId) return;
    this.selectedActivityId = id;
    this.loadingAttendance = true;
    this.error = "";
    const cursoId = this.selectedCourseId;

    forkJoin({
      enrollments: this.education.listCourseEnrollments(cursoId),
      attendance: this.education.listActivityAttendance(id),
    }).subscribe({
      next: ({ enrollments, attendance }) => {
        const presentById = new Map(
          attendance.map((a) => [a.participanteId, a.presente]),
        );
        this.rows = enrollments
          .filter((e) => e.estado === "activa")
          .map((e) => ({
            participanteId: e.participanteId,
            nombreCompleto:
              this.participantNames.get(e.participanteId) ??
              e.participanteId,
            presente: presentById.get(e.participanteId) ?? true,
          }));
        this.loadingAttendance = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loadingAttendance = false;
      },
    });
  }

  guardar() {
    if (!this.selectedActivityId || this.rows.length === 0) return;
    this.guardando = true;
    this.error = "";
    const payload = this.rows.map((r) => ({
      participanteId: r.participanteId,
      presente: r.presente,
    }));
    this.education
      .registerActivityAttendance(this.selectedActivityId, payload)
      .subscribe({
        next: () => {
          this.guardando = false;
          this.toast.add({
            severity: "success",
            summary: "Asistencia registrada",
            detail: `${payload.length} participante(s)`,
          });
        },
        error: (err: Error) => {
          this.guardando = false;
          this.toast.add({
            severity: "error",
            summary: "No se pudo guardar",
            detail: err.message,
          });
        },
      });
  }
}
