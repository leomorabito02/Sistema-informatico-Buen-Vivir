import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { TagModule } from "primeng/tag";
import { CardModule } from "primeng/card";
import { SelectModule } from "primeng/select";
import { ProgressBarModule } from "primeng/progressbar";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { MessageModule } from "primeng/message";
import { forkJoin } from "rxjs";
import { EducationService } from "../../../services/education.service";
import { Course } from "../../../core/models/education.model";

interface Option {
  label: string;
  value: string;
}
interface SeguimientoRow {
  participanteId: string;
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
    CommonModule,
    FormsModule,
    TableModule,
    TagModule,
    CardModule,
    SelectModule,
    ProgressBarModule,
    ProgressSpinnerModule,
    MessageModule,
  ],
  templateUrl: "./seguimiento.component.html",
  styleUrls: ["./seguimiento.component.css"],
})
export class SeguimientoComponent implements OnInit {
  private education = inject(EducationService);

  loadingCourses = true;
  loadingData = false;
  error = "";
  courseOptions: Option[] = [];
  selectedCourseId: string | null = null;
  rows: SeguimientoRow[] = [];

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
    if (!this.selectedCourseId) return;
    this.loadingData = true;
    this.error = "";
    this.rows = [];
    const id = this.selectedCourseId;
    forkJoin({
      enrollments: this.education.listEnrollments(id),
      progress: this.education.getCourseProgress(id),
    }).subscribe({
      next: ({ enrollments, progress }) => {
        const byId = new Map(progress.map((p) => [p.participanteId, p]));
        this.rows = enrollments.map((e) => {
          const p = byId.get(e.participanteId);
          const pct = p?.porcentajeAsistencia ?? 0;
          return {
            participanteId: e.participanteId,
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
