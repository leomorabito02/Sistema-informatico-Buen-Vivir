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
import { EducationService } from "../../../services/education.service";
import {
  Course,
  materialHref,
  richTextToPlain,
} from "../../../core/models/education.model";
import { RichTextComponent } from "../../../shared/components/rich-text.component";

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
    RichTextComponent,
  ],
  templateUrl: "./cursos.component.html",
  styleUrls: ["./cursos.component.css"],
})
export class CursosComponent implements OnInit {
  private education = inject(EducationService);

  searchTerm = "";
  loading = true;
  error = "";
  cursos: CursoVM[] = [];
  filteredCursos: CursoVM[] = [];

  detailVisible = false;
  detail: CursoVM | null = null;

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
