import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { TagModule } from "primeng/tag";
import { CardModule } from "primeng/card";
import { SelectModule } from "primeng/select";
import { ProgressBarModule } from "primeng/progressbar";
import { Observable, forkJoin } from "rxjs";
import { EducationService } from "../../../services/education.service";
import { AuthService } from "../../../services/auth.service";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import {
  Course,
  Enrollment,
  ParticipantProgress,
} from "../../../core/models/education.model";

interface Option {
  label: string;
  value: string;
}
interface SeguimientoRow {
  participanteId: string;
  /** Nombre y apellido (solo en la vista de admin/voluntario). */
  nombre?: string;
  estado: string;
  porcentajeAsistencia: number;
  actividadesAsistidas: number;
  actividadesTotal: number;
  riesgo: "Sin datos" | "En riesgo" | "Regular" | "Buen avance";
}

@Component({
  selector: "app-seguimiento",
  standalone: true,
  imports: [
    LoadingStateComponent,
    ErrorStateComponent,
    EmptyStateComponent,
    CommonModule,
    FormsModule,
    TableModule,
    TagModule,
    CardModule,
    SelectModule,
    ProgressBarModule,
  ],
  templateUrl: "./seguimiento.component.html",
  styleUrls: ["./seguimiento.component.css"],
})
export class SeguimientoComponent implements OnInit {
  private education = inject(EducationService);
  private auth = inject(AuthService);

  /**
   * Admin y voluntario ven cualquier curso del catálogo (HU-20, HU-06);
   * el docente, solo los suyos (/api/docente). Antes la pantalla usaba
   * siempre /api/docente: al voluntario le daba 403 y al admin una lista
   * vacía (auditoría 2026-11-11, Problema 4).
   */
  gestion = false;

  loadingCourses = true;
  loadingData = false;
  error = "";
  courseOptions: Option[] = [];
  selectedCourseId: string | null = null;
  rows: SeguimientoRow[] = [];

  ngOnInit() {
    this.auth.isAuthenticated().subscribe({
      next: (res) => {
        const rol = res.role_id ?? res.user?.role_id ?? "";
        this.gestion = ["role_admin", "role_dev", "role_volunteer"].includes(
          rol,
        );
        this.cargarCursos();
      },
      error: () => {
        this.error = "No se pudo verificar tu sesión. Volvé a iniciar sesión.";
        this.loadingCourses = false;
      },
    });
  }

  private cargarCursos() {
    const cursos$: Observable<Course[]> = this.gestion
      ? this.education.listCourses()
      : this.education.listMyCourses();
    cursos$.subscribe({
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
    if (!this.selectedCourseId) return;
    this.loadingData = true;
    this.error = "";
    this.rows = [];
    const id = this.selectedCourseId;
    const enrollments$: Observable<Enrollment[]> = this.gestion
      ? this.education.listCourseEnrollments(id)
      : this.education.listEnrollments(id);
    const progress$: Observable<ParticipantProgress[]> = this.gestion
      ? this.education.getCourseProgressGestion(id)
      : this.education.getCourseProgress(id);
    forkJoin({
      enrollments: enrollments$,
      progress: progress$,
    }).subscribe({
      next: ({ enrollments, progress }) => {
        const byId = new Map(progress.map((p) => [p.participanteId, p]));
        this.rows = enrollments.map((e) => {
          const p = byId.get(e.participanteId);
          const pct = p?.porcentajeAsistencia ?? 0;
          const nombre = `${e.nombre ?? ""} ${e.apellido ?? ""}`.trim();
          return {
            participanteId: e.participanteId,
            nombre: nombre || undefined,
            estado: e.estado,
            porcentajeAsistencia: pct,
            actividadesAsistidas: p?.actividadesAsistidas ?? 0,
            actividadesTotal: p?.actividadesTotal ?? 0,
            riesgo: !p
              ? "Sin datos"
              : pct < 50
                ? "En riesgo"
                : pct < 75
                  ? "Regular"
                  : "Buen avance",
          };
        });
        this.loadingData = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loadingData = false;
      },
    });
  }

  riesgoSeverity(r: string): string {
    switch (r) {
      case "Buen avance":
        return "success";
      case "Regular":
        return "info";
      case "En riesgo":
        return "danger";
      default:
        return "secondary";
    }
  }
}
