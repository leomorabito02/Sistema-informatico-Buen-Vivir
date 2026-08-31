import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { TagModule } from "primeng/tag";
import { CardModule } from "primeng/card";
import { SelectModule } from "primeng/select";
import { InputTextModule } from "primeng/inputtext";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { MessageModule } from "primeng/message";
import { EducationService } from "../../../services/education.service";
import { Course, Enrollment } from "../../../core/models/education.model";

interface Option {
  label: string;
  value: string;
}

@Component({
  selector: "app-participantes",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    TableModule,
    TagModule,
    CardModule,
    SelectModule,
    InputTextModule,
    ProgressSpinnerModule,
    MessageModule,
  ],
  templateUrl: "./participantes.component.html",
  styleUrls: ["./participantes.component.css"],
})
export class ParticipantesComponent implements OnInit {
  private education = inject(EducationService);

  loadingCourses = true;
  loadingEnrollments = false;
  error = "";
  courseOptions: Option[] = [];
  selectedCourseId: string | null = null;

  searchTerm = "";
  enrollments: Enrollment[] = [];
  filtered: Enrollment[] = [];

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
    this.loadingEnrollments = true;
    this.error = "";
    this.enrollments = [];
    this.filtered = [];
    this.education.listEnrollments(this.selectedCourseId).subscribe({
      next: (list) => {
        this.enrollments = list;
        this.filtered = list;
        this.loadingEnrollments = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loadingEnrollments = false;
      },
    });
  }

  filter() {
    const term = this.searchTerm.toLowerCase().trim();
    this.filtered = this.enrollments.filter(
      (e) =>
        e.participanteId.toLowerCase().includes(term) ||
        e.estado.toLowerCase().includes(term),
    );
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
