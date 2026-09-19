import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { ButtonModule } from "primeng/button";
import { CardModule } from "primeng/card";
import { TagModule } from "primeng/tag";
import { InputTextModule } from "primeng/inputtext";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { MessageModule } from "primeng/message";
import { TooltipModule } from "primeng/tooltip";
import { DialogModule } from "primeng/dialog";
import { SelectModule } from "primeng/select";
import { ToastModule } from "primeng/toast";
import { MessageService } from "primeng/api";
import { AuthService } from "../../../services/auth.service";
import { EducationService } from "../../../services/education.service";
import {
  Course,
  materialHref,
  richTextToPlain,
} from "../../../core/models/education.model";
import { RichTextComponent } from "../../../shared/components/rich-text.component";
import { todayYMD } from "../../../shared/utils/date-ymd";

interface Option {
  label: string;
  value: string;
}

interface CursoVM {
  id: string;
  title: string;
  area: string;
  descripcion: string;
  mode: string;
  duration: string;
  teacherId: string;
  materialsCount: number;
  raw: Course;
}

@Component({
  selector: "app-cursos",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ButtonModule,
    CardModule,
    TagModule,
    InputTextModule,
    ProgressSpinnerModule,
    MessageModule,
    TooltipModule,
    DialogModule,
    SelectModule,
    ToastModule,
    RichTextComponent,
  ],
  providers: [MessageService],
  templateUrl: "./cursos.component.html",
  styleUrls: ["./cursos.component.css"],
})
export class CursosComponent implements OnInit {
  private education = inject(EducationService);
  private authService = inject(AuthService);
  private toast = inject(MessageService);

  searchTerm = "";
  loading = true;
  error = "";
  cursos: CursoVM[] = [];
  filteredCursos: CursoVM[] = [];

  detailVisible = false;
  detail: CursoVM | null = null;

  /** true para role_admin/role_volunteer: "Inscribirme" pasa a ser
   * "Inscribir participante" (HU-18/RF-18), sin auto-inscripción. */
  canEnrollParticipants = false;
  participantOptions: Option[] = [];

  enrollVisible = false;
  enrollCurso: CursoVM | null = null;
  selectedParticipantId: string | null = null;
  inscribiendo = false;

  ngOnInit() {
    this.education.listCourses().subscribe({
      next: (courses) => {
        this.cursos = courses.map((c) => this.toVM(c));
        this.filteredCursos = this.cursos;
        this.loading = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loading = false;
      },
    });

    this.authService.isAuthenticated().subscribe({
      next: (res) => {
        const roleId = res.role_id ?? res.user?.role_id ?? "";
        this.canEnrollParticipants =
          roleId === "role_admin" || roleId === "role_volunteer";
        if (this.canEnrollParticipants) {
          this.loadParticipantOptions();
        }
      },
      error: () => {},
    });
  }

  private loadParticipantOptions() {
    this.education.getGeneralAttendance(todayYMD()).subscribe({
      next: (directory) => {
        this.participantOptions = directory
          .map((p) => ({
            label:
              `${p.nombre ?? ""} ${p.apellido ?? ""}`.trim() ||
              p.participanteId,
            value: p.participanteId,
          }))
          .sort((a, b) => a.label.localeCompare(b.label));
      },
      error: () => {
        this.participantOptions = [];
      },
    });
  }

  openEnrollDialog(curso: CursoVM) {
    this.enrollCurso = curso;
    this.selectedParticipantId = null;
    this.enrollVisible = true;
  }

  inscribir() {
    if (!this.enrollCurso || !this.selectedParticipantId) return;
    this.inscribiendo = true;
    const participantLabel =
      this.participantOptions.find((o) => o.value === this.selectedParticipantId)
        ?.label ?? this.selectedParticipantId;
    this.education
      .enrollParticipant(this.enrollCurso.id, {
        participanteId: this.selectedParticipantId,
      })
      .subscribe({
        next: () => {
          this.inscribiendo = false;
          this.enrollVisible = false;
          this.toast.add({
            severity: "success",
            summary: "Inscripción registrada",
            detail: participantLabel,
          });
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

  private toVM(c: Course): CursoVM {
    return {
      id: c.id,
      title: c.title,
      area: c.trainingArea?.name ?? "Sin área",
      descripcion: richTextToPlain(c.description),
      mode: c.mode || "—",
      duration: c.duration || "Duración a confirmar",
      teacherId: c.responsibleTeacher ?? "",
      materialsCount: c.material?.length ?? 0,
      raw: c,
    };
  }

  filterCursos() {
    const term = this.searchTerm.toLowerCase().trim();
    this.filteredCursos = this.cursos.filter(
      (c) =>
        c.title.toLowerCase().includes(term) ||
        c.area.toLowerCase().includes(term) ||
        c.descripcion.toLowerCase().includes(term),
    );
  }

  openDetail(c: CursoVM) {
    this.detail = c;
    this.detailVisible = true;
  }

  matHref = materialHref;

  modeSeverity(mode: string): string {
    switch (mode) {
      case "Presencial":
        return "success";
      case "Virtual":
        return "info";
      case "Mixta":
        return "warn";
      default:
        return "secondary";
    }
  }
}
