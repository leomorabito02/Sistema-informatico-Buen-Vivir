import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { CardModule } from "primeng/card";
import { SelectModule } from "primeng/select";
import { ToastModule } from "primeng/toast";
import { MessageService } from "primeng/api";

import { EducationService } from "../../../services/education.service";
import {
  Course,
  Enrollment,
  ParticipanteResumen,
} from "../../../core/models/education.model";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import {
  ParticipantPickerComponent,
  etiquetaParticipante,
} from "../../../shared/components/participant-picker.component";

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
 * El selector de participantes busca en el backend
 * (`GET /api/participantes?q=&cursoId=`, ver ParticipantPickerComponent):
 * ya no se baja el directorio entero para filtrarlo en el navegador. Los
 * nombres de los inscriptos también los resuelve el backend.
 */
@Component({
  selector: "app-inscripciones",
  standalone: true,
  imports: [
    LoadingStateComponent,
    ErrorStateComponent,
    EmptyStateComponent,
    CommonModule,
    FormsModule,
    TableModule,
    ButtonModule,
    TagModule,
    CardModule,
    SelectModule,
    ToastModule,
    ParticipantPickerComponent,
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

  loadingEnrollments = false;
  enrollments: EnrollmentRow[] = [];

  selectedParticipant: ParticipanteResumen | null = null;
  inscribiendo = false;

  ngOnInit() {
    this.education.listCourses().subscribe({
      next: (courses) => {
        this.courseOptions = courses.map((c: Course) => ({
          label: c.title || c.id,
          value: c.id,
        }));
        this.loading = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loading = false;
      },
    });
  }

  onCourseChange() {
    this.selectedParticipant = null;
    this.enrollments = [];
    if (!this.selectedCourseId) return;
    this.loadEnrollments();
  }

  private loadEnrollments() {
    if (!this.selectedCourseId) return;
    this.loadingEnrollments = true;
    this.error = "";
    this.education.listCourseEnrollments(this.selectedCourseId).subscribe({
      next: (list) => {
        this.enrollments = list.map((e) => ({
          ...e,
          nombreCompleto: etiquetaParticipante(e),
        }));
        this.loadingEnrollments = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loadingEnrollments = false;
      },
    });
  }

  inscribir(picker?: ParticipantPickerComponent) {
    const participante = this.selectedParticipant;
    if (!this.selectedCourseId || !participante) return;
    this.inscribiendo = true;
    this.error = "";
    this.education
      .enrollParticipant(this.selectedCourseId, {
        participanteId: participante.participanteId,
      })
      .subscribe({
        next: () => {
          this.inscribiendo = false;
          this.toast.add({
            severity: "success",
            summary: "Inscripción registrada",
            detail: etiquetaParticipante(participante),
          });
          picker?.limpiar();
          this.selectedParticipant = null;
          this.loadEnrollments();
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
