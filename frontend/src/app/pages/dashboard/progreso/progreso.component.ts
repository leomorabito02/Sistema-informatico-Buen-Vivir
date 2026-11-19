import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { CardModule } from "primeng/card";
import { ProgressBarModule } from "primeng/progressbar";
import { SelectModule } from "primeng/select";
import { forkJoin, of } from "rxjs";
import { catchError } from "rxjs/operators";
import { EducationService } from "../../../services/education.service";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import {
  Course,
  ParticipantProgress,
} from "../../../core/models/education.model";

/** Fila de progreso con el nombre del participante (antes se veía el id). */
interface ProgressRow extends ParticipantProgress {
  nombre: string;
}

interface CourseOption {
  label: string;
  value: string;
}

@Component({
  selector: "app-progreso",
  standalone: true,
  imports: [
    LoadingStateComponent,
    ErrorStateComponent,
    EmptyStateComponent,
    CommonModule,
    FormsModule,
    TableModule,
    CardModule,
    ProgressBarModule,
    SelectModule,
  ],
  templateUrl: "./progreso.component.html",
  styleUrls: ["./progreso.component.css"],
})
export class ProgresoComponent implements OnInit {
  private education = inject(EducationService);

  loadingCourses = true;
  loadingProgress = false;
  error = "";
  courseOptions: CourseOption[] = [];
  selectedCourseId: string | null = null;
  rows: ProgressRow[] = [];

  get overallProgress(): number {
    if (!this.rows.length) return 0;
    const sum = this.rows.reduce((a, r) => a + r.porcentajeAsistencia, 0);
    return Math.round(sum / this.rows.length);
  }

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
    this.loadingProgress = true;
    this.error = "";
    this.rows = [];
    forkJoin({
      progress: this.education.getCourseProgress(this.selectedCourseId),
      // Los nombres vienen con las inscripciones; si fallan, queda el id.
      enrollments: this.education
        .listEnrollments(this.selectedCourseId)
        .pipe(catchError(() => of([]))),
    }).subscribe({
      next: ({ progress, enrollments }) => {
        const nombres = new Map(
          enrollments.map((e) => [
            e.participanteId,
            `${e.nombre ?? ""} ${e.apellido ?? ""}`.trim(),
          ]),
        );
        this.rows = progress
          .map((p) => ({
            ...p,
            nombre: nombres.get(p.participanteId) || p.participanteId,
          }))
          .sort((a, b) => a.nombre.localeCompare(b.nombre));
        this.loadingProgress = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loadingProgress = false;
      },
    });
  }
}
