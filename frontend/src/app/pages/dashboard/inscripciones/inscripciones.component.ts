import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { CardModule } from "primeng/card";
import { SelectModule } from "primeng/select";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { MessageModule } from "primeng/message";
import { ToastModule } from "primeng/toast";
import { MessageService } from "primeng/api";
import { forkJoin } from "rxjs";

import { EducationService } from "../../../services/education.service";
import { Course, Enrollment } from "../../../core/models/education.model";
import { todayYMD } from "../../../shared/utils/date-ymd";

interface Option {
  label: string;
  value: string;
}
type EnrollmentRow = Enrollment & { nombreCompleto: string };

/**
 * Inscripción de participantes a cursos (HU-18/RF-18), para role_admin y
 * role_volunteer. A diferencia de "Mis cursos" (docente), acá se puede
 * inscribir en cualquier curso publicado del catálogo.
 *
 * El selector de participantes reutiliza `GET /api/asistencia-general`
 * (fecha de hoy) porque es, hoy, el único endpoint del backend que
 * devuelve el directorio de participantes aprobados con nombre y
 * apellido ya resueltos contra auth-service — evita tener que pedir el
 * `participant_id` a mano (ver advertencia de
 * education/GUIA_PRUEBAS_INSCRIPCION_ASISTENCIA.md sobre confundirlo con
 * `user_id`).
 */
@Component({
  selector: "app-inscripciones",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    TableModule,
    ButtonModule,
    TagModule,
    CardModule,
    SelectModule,
    ProgressSpinnerModule,
    MessageModule,
    ToastModule,
  ],
  providers: [MessageService],
  templateUrl: "./inscripciones.component.html",
  styleUrls: ["./inscripciones.component.css"],
})
export class InscripcionesComponent implements OnInit {
  private education = inject(EducationService);
  private toast = inject(MessageService);

  loading = true;
  error = "";

  courseOptions: Option[] = [];
  selectedCourseId: string | null = null;

  participantOptions: Option[] = [];
  private participantNames = new Map<string, string>();
  participantsError = "";

  loadingEnrollments = false;
  enrollments: EnrollmentRow[] = [];

  selectedParticipantId: string | null = null;
  inscribiendo = false;

  ngOnInit() {
    forkJoin({
      courses: this.education.listCourses(),
      directory: this.education.getGeneralAttendance(todayYMD()),
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
        this.participantOptions = directory
          .map((p) => ({
            label: this.participantNames.get(p.participanteId)!,
            value: p.participanteId,
          }))
          .sort((a, b) => a.label.localeCompare(b.label));
        this.loading = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loading = false;
      },
    });
  }

  onCourseChange() {
    this.selectedParticipantId = null;
    this.enrollments = [];
    if (!this.selectedCourseId) return;
    this.loadingEnrollments = true;
    this.error = "";
    this.education.listCourseEnrollments(this.selectedCourseId).subscribe({
      next: (list) => {
        this.enrollments = list.map((e) => ({
          ...e,
          nombreCompleto:
            this.participantNames.get(e.participanteId) ?? e.participanteId,
        }));
        this.loadingEnrollments = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loadingEnrollments = false;
      },
    });
  }

  /** Participantes que todavía no tienen una inscripción activa en el curso elegido. */
  get availableParticipantOptions(): Option[] {
    const enrolled = new Set(
      this.enrollments
        .filter((e) => e.estado === "activa")
        .map((e) => e.participanteId),
    );
    return this.participantOptions.filter((o) => !enrolled.has(o.value));
  }

  inscribir() {
    if (!this.selectedCourseId || !this.selectedParticipantId) return;
    this.inscribiendo = true;
    this.error = "";
    this.education
      .enrollParticipant(this.selectedCourseId, {
        participanteId: this.selectedParticipantId,
      })
      .subscribe({
        next: () => {
          this.inscribiendo = false;
          this.toast.add({
            severity: "success",
            summary: "Inscripción registrada",
            detail:
              this.participantNames.get(this.selectedParticipantId!) ??
              this.selectedParticipantId!,
          });
          this.selectedParticipantId = null;
          this.onCourseChange();
        },
        error: (err: Error) => {
          this.inscribiendo = false;
          this.toast.add({
            severity: "error",
            summary: "No se pudo inscribir",
            detail: err.message,
          });
        },
      });
  }

  estadoSeverity(estado: string): string {
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
}
