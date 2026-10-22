import { Component, DestroyRef, OnInit, inject } from "@angular/core";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { Router } from "@angular/router";
import { ButtonModule } from "primeng/button";
import { CardModule } from "primeng/card";
import { TagModule } from "primeng/tag";
import { InputTextModule } from "primeng/inputtext";
import { TooltipModule } from "primeng/tooltip";
import { DialogModule } from "primeng/dialog";
import { SelectModule } from "primeng/select";
import { ToastModule } from "primeng/toast";
import { MessageService } from "primeng/api";
import { Subject, of } from "rxjs";
import {
  catchError,
  debounceTime,
  distinctUntilChanged,
  map,
  switchMap,
} from "rxjs/operators";
import { AuthService } from "../../../services/auth.service";
import { EducationService } from "../../../services/education.service";
import {
  Course,
  materialHref,
  richTextToPlain,
} from "../../../core/models/education.model";
import { RichTextComponent } from "../../../shared/components/rich-text.component";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
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
    LoadingStateComponent,
    ErrorStateComponent,
    EmptyStateComponent,
    CommonModule,
    FormsModule,
    ButtonModule,
    CardModule,
    TagModule,
    InputTextModule,
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
  private router = inject(Router);
  private destroyRef = inject(DestroyRef);
  /** Texto del buscador; la búsqueda la hace el backend (?q=). */
  private busqueda$ = new Subject<string>();

  searchTerm = "";
  loading = true;
  error = "";
  cursos: CursoVM[] = [];
  filteredCursos: CursoVM[] = [];

  detailVisible = false;
  detail: CursoVM | null = null;

  /** true para role_admin/role_volunteer: pueden inscribir participantes
   * desde el catálogo (HU-18/RF-18). */
  canEnrollParticipants = false;
  /** El participante ve solo sus cursos inscriptos ("Mis cursos"). */
  esParticipante = false;
  private roleId = "";
  private userId = "";
  participantOptions: Option[] = [];

  enrollVisible = false;
  enrollCurso: CursoVM | null = null;
  selectedParticipantId: string | null = null;
  inscribiendo = false;

  ngOnInit() {
    // Espera a que se deje de escribir y cancela la consulta anterior si
    // llega una nueva (switchMap).
    this.busqueda$
      .pipe(
        map((t) => t.trim()),
        debounceTime(300),
        distinctUntilChanged(),
        switchMap((q) =>
          this.cursos$(q).pipe(
            catchError((err: Error) => {
              this.error = err.message;
              return of(null);
            }),
          ),
        ),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe((courses) => {
        if (courses) {
          this.error = "";
          this.mostrar(courses);
        }
      });

    this.authService.isAuthenticated().subscribe({
      next: (res) => {
        this.roleId = res.role_id ?? res.user?.role_id ?? "";
        this.userId = res.user?.user_id ?? "";
        this.esParticipante = this.roleId === "role_participant";
        this.canEnrollParticipants =
          this.roleId === "role_admin" || this.roleId === "role_volunteer";
        if (this.canEnrollParticipants) {
          this.loadParticipantOptions();
        }
        this.loadCursos();
      },
      error: () => this.loadCursos(),
    });
  }

  private cursos$(q = "") {
    return this.esParticipante
      ? this.education.listMyEnrolledCourses(q)
      : this.education.listCourses(q);
  }

  private loadCursos() {
    this.cursos$(this.searchTerm.trim()).subscribe({
      next: (courses) => {
        this.mostrar(courses);
        this.loading = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loading = false;
      },
    });
  }

  private mostrar(courses: Course[]) {
    this.cursos = courses.map((c) => this.toVM(c));
    this.filteredCursos = this.cursos;
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
      this.participantOptions.find(
        (o) => o.value === this.selectedParticipantId,
      )?.label ?? this.selectedParticipantId;
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

  /** Lo llama el buscador en cada tecla: el backend filtra (?q=). */
  filterCursos() {
    this.busqueda$.next(this.searchTerm);
  }

  /**
   * Abre la pantalla de detalle (`/dashboard/cursos/:id`) para quien
   * puede verla: admin, voluntario, el participante inscripto y el
   * docente responsable de ese curso. Para el resto (otro docente,
   * empresa) se muestra el resumen público en un diálogo.
   */
  openDetail(c: CursoVM) {
    if (this.tieneDetalleCompleto(c)) {
      this.router.navigate(["/dashboard/cursos", c.id]);
      return;
    }
    this.detail = c;
    this.detailVisible = true;
  }

  private tieneDetalleCompleto(c: CursoVM): boolean {
    switch (this.roleId) {
      case "role_admin":
      case "role_dev":
      case "role_volunteer":
      case "role_participant":
        return true;
      case "role_teacher":
        return !!this.userId && c.teacherId === this.userId;
      default:
        return false;
    }
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
