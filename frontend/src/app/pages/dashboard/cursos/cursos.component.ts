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
  ParticipanteResumen,
  materialHref,
  richTextToPlain,
} from "../../../core/models/education.model";
import { RichTextComponent } from "../../../shared/components/rich-text.component";
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

interface CursoVM {
  id: string;
  title: string;
  area: string;
  descripcion: string;
  mode: string;
  duration: string;
  teacherId: string;
  /** Nombre del docente (Educación lo resuelve); vacío si no se pudo. */
  teacherName: string;
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
    ParticipantPickerComponent,
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
  /** Buscador y área elegida; los dos los aplica el backend (?q=&area=). */
  private busqueda$ = new Subject<{ q: string; area: string }>();

  searchTerm = "";
  /** Filtro por área formativa (HU-28), combinable con el buscador. */
  areaOptions: Option[] = [];
  selectedArea = "";
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
  /** true cuando ya se sabe el rol (para no mostrar un título y cambiarlo). */
  rolCargado = false;
  private roleId = "";
  private userId = "";
  enrollVisible = false;
  enrollCurso: CursoVM | null = null;
  selectedParticipant: ParticipanteResumen | null = null;
  inscribiendo = false;

  ngOnInit() {
    // Espera a que se deje de escribir y cancela la consulta anterior si
    // llega una nueva (switchMap).
    this.busqueda$
      .pipe(
        map((b) => ({ q: b.q.trim(), area: b.area })),
        debounceTime(300),
        distinctUntilChanged((a, b) => a.q === b.q && a.area === b.area),
        switchMap(({ q, area }) =>
          this.cursos$(q, area).pipe(
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
        this.rolCargado = true;
        this.canEnrollParticipants =
          this.roleId === "role_admin" || this.roleId === "role_volunteer";
        this.loadAreas();
        this.loadCursos();
      },
      error: () => {
        this.rolCargado = true;
        this.loadCursos();
      },
    });
  }

  private cursos$(q = "", area = "") {
    return this.esParticipante
      ? this.education.listMyEnrolledCourses(q, area)
      : this.education.listCourses(q, area);
  }

  /** Opciones del filtro por área; si fallan, el filtro no se muestra. */
  private loadAreas() {
    this.education.listTrainingAreas().subscribe({
      next: (areas) => {
        this.areaOptions = [
          { label: "Todas las áreas", value: "" },
          ...areas
            .map((a) => ({ label: a.name, value: a.id }))
            .sort((a, b) => a.label.localeCompare(b.label)),
        ];
      },
      error: () => (this.areaOptions = []),
    });
  }

  private loadCursos() {
    this.cursos$(this.searchTerm.trim(), this.selectedArea).subscribe({
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

  openEnrollDialog(curso: CursoVM) {
    this.enrollCurso = curso;
    this.selectedParticipant = null;
    this.enrollVisible = true;
  }

  inscribir() {
    const participante = this.selectedParticipant;
    if (!this.enrollCurso || !participante) return;
    this.inscribiendo = true;
    const participantLabel = etiquetaParticipante(participante);
    this.education
      .enrollParticipant(this.enrollCurso.id, {
        participanteId: participante.participanteId,
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
      teacherName: c.responsibleTeacherName?.trim() ?? "",
      materialsCount: c.material?.length ?? 0,
      raw: c,
    };
  }

  /** Lo llaman el buscador (en cada tecla) y el filtro de área: filtra el backend. */
  filterCursos() {
    this.busqueda$.next({ q: this.searchTerm, area: this.selectedArea });
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
