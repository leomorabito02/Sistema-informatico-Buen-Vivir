import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { RouterLink } from "@angular/router";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { DialogModule } from "primeng/dialog";
import { SelectModule } from "primeng/select";
import { InputTextModule } from "primeng/inputtext";
import { ToastModule } from "primeng/toast";
import { MessageService } from "primeng/api";
import { forkJoin } from "rxjs";
import { EducationService } from "../../../services/education.service";
import {
  Course,
  CourseInput,
  Material,
  MaterialType,
  materialHref,
  plainToRichText,
  richTextToPlain,
} from "../../../core/models/education.model";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";

interface CourseVM {
  id: string;
  title: string;
  area: string;
  mode: string;
  duration: string;
  materialsCount: number;
  publicado: boolean;
  raw: Course;
}

@Component({
  selector: "app-mis-cursos",
  standalone: true,
  imports: [
    LoadingStateComponent,
    ErrorStateComponent,
    CommonModule,
    FormsModule,
    RouterLink,
    ButtonModule,
    TagModule,
    DialogModule,
    SelectModule,
    InputTextModule,
    ToastModule,
  ],
  providers: [MessageService],
  templateUrl: "./mis-cursos.component.html",
  styleUrls: ["./mis-cursos.component.css"],
})
export class MisCursosComponent implements OnInit {
  private education = inject(EducationService);
  private toast = inject(MessageService);

  loading = true;
  error = "";
  courses: CourseVM[] = [];

  matHref = materialHref;

  // ---- diálogo de materiales del curso ----
  matDialog = false;
  matCourse: CourseVM | null = null;
  savingMat = false;
  matFile: File | null = null;
  matFileError = "";
  matForm: { title: string; type: MaterialType; url: string } = {
    title: "",
    type: "PDF",
    url: "",
  };

  private readonly uploadLimits: Record<string, number> = {
    "image/jpeg": 5,
    "image/png": 5,
    "image/webp": 5,
    "application/pdf": 10,
    "application/msword": 10,
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document": 10,
  };

  // ---- crear / editar / eliminar curso propio ----
  courseDialog = false;
  editingCourse: CourseVM | null = null;
  savingCourse = false;
  courseForm = {
    title: "",
    description: "",
    trainingAreaId: "",
    mode: "Presencial",
    duration: "",
  };
  areaOptions: { label: string; value: string }[] = [];
  readonly modeOptions = [
    { label: "Presencial", value: "Presencial" },
    { label: "Virtual", value: "Virtual" },
    { label: "Mixta", value: "Mixta" },
  ];
  deleteDialog = false;
  deletingCourse: CourseVM | null = null;
  deletingBusy = false;

  ngOnInit() {
    this.cargar();
    this.education.listTrainingAreas().subscribe({
      next: (areas) =>
        (this.areaOptions = areas.map((a) => ({ label: a.name, value: a.id }))),
      error: () => (this.areaOptions = []),
    });
  }

  cargar() {
    this.loading = true;
    this.error = "";
    forkJoin({
      publicados: this.education.listMyCourses(),
      borradores: this.education.listMyDraftCourses(),
    }).subscribe({
      next: ({ publicados, borradores }) => {
        this.courses = [
          ...publicados.map((c) => this.toVM(c, true)),
          ...borradores.map((c) => this.toVM(c, false)),
        ];
        this.loading = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loading = false;
      },
    });
  }

  private toVM(c: Course, publicado: boolean): CourseVM {
    return {
      id: c.id,
      title: c.title || "(sin título)",
      area: c.trainingArea?.name ?? "—",
      mode: c.mode || "—",
      duration: c.duration || "—",
      materialsCount: c.material?.length ?? 0,
      publicado,
      raw: c,
    };
  }

  preview(c: Course): string {
    return richTextToPlain(c.description);
  }

  // ---------------- materiales ----------------
  openMaterials(c: CourseVM) {
    this.matCourse = c;
    this.matForm = { title: "", type: "PDF", url: "" };
    this.matFile = null;
    this.matFileError = "";
    this.matDialog = true;
  }

  get matUrlInvalid(): boolean {
    const u = this.matForm.url.trim();
    return !!u && !/^https?:\/\//i.test(u);
  }

  onMatTypeChange() {
    // el campo que no corresponde al tipo elegido no se manda al backend
    this.matForm.url = "";
    this.matFile = null;
    this.matFileError = "";
  }

  onMatFile(event: Event) {
    const input = event.target as HTMLInputElement;
    const f = input.files?.[0] ?? null;
    this.matFileError = "";
    if (!f) {
      this.matFile = null;
      return;
    }
    const limitMb = this.uploadLimits[f.type];
    if (!limitMb) {
      this.matFileError =
        "Tipo no permitido (JPG/PNG/WEBP/PDF/DOC). Video: usá URL.";
      this.matFile = null;
      return;
    }
    if (f.size > limitMb * 1024 * 1024) {
      this.matFileError = `Máximo ${limitMb} MB para este tipo.`;
      this.matFile = null;
      return;
    }
    this.matFile = f;
    if (!this.matForm.title.trim()) {
      this.matForm.title = f.name.replace(/\.[^.]+$/, "");
    }
  }

  addMaterial() {
    if (!this.matCourse || !this.matForm.title.trim() || this.matUrlInvalid)
      return;
    this.savingMat = true;

    const create$ = this.matFile
      ? this.education.uploadMaterial(
          this.matFile,
          this.matForm.title.trim(),
          this.matForm.type,
        )
      : this.education.createMaterial({
          title: this.matForm.title.trim(),
          type: this.matForm.type,
          url: this.matForm.url.trim() || undefined,
        });

    create$.subscribe({
      next: (material: Material) => this.attachToCourse(material),
      error: (err: Error) => {
        this.savingMat = false;
        this.toast.add({
          severity: "error",
          summary: "No se pudo crear el material",
          detail: err.message,
        });
      },
    });
  }

  private attachToCourse(material: Material) {
    const c = this.matCourse!.raw;
    const input: CourseInput = {
      title: c.title,
      description: c.description ?? plainToRichText(""),
      mode: c.mode,
      duration: c.duration,
      trainingAreaId: c.trainingArea?.id ?? "",
      materialIds: [...(c.material ?? []).map((m) => m.id), material.id],
    };
    this.education.updateMyCourse(c.id, input).subscribe({
      next: (updated) => {
        this.savingMat = false;
        this.matDialog = false;
        this.toast.add({
          severity: "success",
          summary: "Material agregado al curso",
          detail: material.title,
        });
        // refrescar la fila
        const vm = this.courses.find((x) => x.id === c.id);
        if (vm) {
          vm.raw = updated;
          vm.materialsCount = updated.material?.length ?? vm.materialsCount + 1;
        }
      },
      error: (err: Error) => {
        this.savingMat = false;
        this.toast.add({
          severity: "warn",
          summary: "Material creado, pero no se pudo adjuntar",
          detail: err.message,
        });
      },
    });
  }

  // ---------------- curso propio ----------------
  openCourseForm(c?: CourseVM) {
    this.editingCourse = c ?? null;
    const raw = c?.raw;
    this.courseForm = {
      title: raw?.title ?? "",
      description: raw ? richTextToPlain(raw.description) : "",
      trainingAreaId: raw?.trainingArea?.id ?? "",
      mode: raw?.mode || "Presencial",
      duration: raw?.duration ?? "",
    };
    this.courseDialog = true;
  }

  get courseFormValid(): boolean {
    const f = this.courseForm;
    return (
      !!f.title.trim() &&
      !!f.description.trim() &&
      !!f.trainingAreaId &&
      !!f.mode &&
      !!f.duration.trim()
    );
  }

  saveCourse() {
    if (!this.courseFormValid || this.savingCourse) return;
    const f = this.courseForm;
    const input: CourseInput = {
      title: f.title.trim(),
      description: plainToRichText(f.description.trim()),
      mode: f.mode,
      duration: f.duration.trim(),
      trainingAreaId: f.trainingAreaId,
      // Al editar se conservan los materiales que ya tenía.
      materialIds: (this.editingCourse?.raw.material ?? []).map((m) => m.id),
    };
    const editing = this.editingCourse;
    this.savingCourse = true;
    const req$ = editing
      ? this.education.updateMyCourse(editing.id, input)
      : this.education.createMyCourse(input);
    req$.subscribe({
      next: () => {
        this.savingCourse = false;
        this.courseDialog = false;
        this.toast.add({
          severity: "success",
          summary: editing ? "Curso actualizado" : "Curso creado",
          detail: editing
            ? "Se guardaron los cambios."
            : "Quedó pendiente de aprobación del administrador.",
        });
        this.cargar();
      },
      error: (err: Error) => {
        this.savingCourse = false;
        this.toast.add({
          severity: "error",
          summary: "No se pudo guardar el curso",
          detail: err.message,
        });
      },
    });
  }

  askDelete(c: CourseVM) {
    this.deletingCourse = c;
    this.deleteDialog = true;
  }

  confirmDelete() {
    const c = this.deletingCourse;
    if (!c || this.deletingBusy) return;
    this.deletingBusy = true;
    this.education.deleteMyCourse(c.id).subscribe({
      next: () => {
        this.deletingBusy = false;
        this.deleteDialog = false;
        this.courses = this.courses.filter((x) => x.id !== c.id);
        this.toast.add({
          severity: "success",
          summary: "Curso eliminado",
          detail: c.title,
        });
      },
      error: (err: Error) => {
        this.deletingBusy = false;
        this.toast.add({
          severity: "error",
          summary: "No se pudo eliminar el curso",
          detail: err.message,
        });
      },
    });
  }
}
