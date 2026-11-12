import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { Router, RouterLink } from "@angular/router";
import { forkJoin, of } from "rxjs";
import { catchError } from "rxjs/operators";

import { ButtonModule } from "primeng/button";
import { CardModule } from "primeng/card";
import { TableModule } from "primeng/table";
import { TagModule } from "primeng/tag";
import { DialogModule } from "primeng/dialog";
import { SelectModule } from "primeng/select";
import { MultiSelectModule } from "primeng/multiselect";
import { InputTextModule } from "primeng/inputtext";
import { TextareaModule } from "primeng/textarea";
import { CheckboxModule } from "primeng/checkbox";
import { ToastModule } from "primeng/toast";
import { ConfirmDialogModule } from "primeng/confirmdialog";
import { MessageModule } from "primeng/message";
import { TooltipModule } from "primeng/tooltip";
import { MessageService, ConfirmationService } from "primeng/api";

import { EducationService } from "../../../services/education.service";
import { UsersService } from "../../../services/users.service";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import {
  Course,
  Material,
  MaterialType,
  TrainingArea,
  plainToRichText,
  richTextToPlain,
} from "../../../core/models/education.model";

type Section = "cursos" | "areas" | "materiales";
interface Option {
  label: string;
  value: string;
}
type CourseRow = Course & { publicado: boolean };

@Component({
  selector: "app-admin-cursos",
  standalone: true,
  imports: [
    LoadingStateComponent,
    CommonModule,
    FormsModule,
    ButtonModule,
    CardModule,
    TableModule,
    TagModule,
    DialogModule,
    SelectModule,
    MultiSelectModule,
    InputTextModule,
    TextareaModule,
    CheckboxModule,
    ToastModule,
    ConfirmDialogModule,
    MessageModule,
    TooltipModule,
    RouterLink,
  ],
  providers: [MessageService, ConfirmationService],
  templateUrl: "./admin-cursos.component.html",
  styleUrls: ["./admin-cursos.component.css"],
})
export class AdminCursosComponent implements OnInit {
  private readonly education = inject(EducationService);
  private readonly users = inject(UsersService);
  private readonly toast = inject(MessageService);
  private readonly confirm = inject(ConfirmationService);
  private readonly router = inject(Router);

  section: Section = "cursos";
  loading = true;

  courses: CourseRow[] = [];
  areas: TrainingArea[] = [];
  materials: Material[] = [];
  teacherOptions: Option[] = [];
  teachersError = "";

  readonly modeOptions: Option[] = [
    { label: "Presencial", value: "Presencial" },
    { label: "Virtual", value: "Virtual" },
    { label: "Mixta", value: "Mixta" },
  ];
  readonly materialTypeOptions: Option[] = [
    { label: "PDF", value: "PDF" },
    { label: "Video", value: "Video" },
    { label: "Link", value: "Link" },
  ];

  // ---- diálogo curso ----
  courseDialog = false;
  savingCourse = false;
  editingCourseId: string | null = null;
  courseForm = this.emptyCourseForm();

  // ---- diálogo área ----
  areaDialog = false;
  savingArea = false;
  editingAreaId: string | null = null;
  areaForm = { name: "", description: "" };

  // ---- diálogo material ----
  materialDialog = false;
  savingMaterial = false;
  editingMaterialId: string | null = null;
  materialFile: File | null = null;
  materialFileError = "";
  materialForm: {
    title: string;
    type: MaterialType;
    fileAssetId: string;
    url: string;
  } = {
    title: "",
    type: "PDF",
    fileAssetId: "",
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

  ngOnInit() {
    this.reloadAll();
    this.users.listByRole("role_teacher").subscribe({
      next: (list) => {
        this.teacherOptions = list.map((u) => ({
          label: `${u.first_name ?? ""} ${u.last_name ?? ""}`.trim()
            ? `${u.first_name} ${u.last_name} · ${u.email}`
            : u.email,
          value: u.user_id,
        }));
      },
      error: (err: Error) => {
        this.teachersError = err.message;
      },
    });
  }

  reloadAll() {
    this.loading = true;
    forkJoin({
      courses: this.education.listCourses(),
      // el endpoint de borradores puede no estar desplegado todavía
      // (requiere rebuild de education-service) -> degradar a []
      drafts: this.education.listDraftCourses().pipe(catchError(() => of([]))),
      areas: this.education.listTrainingAreas(),
      materials: this.education.listMaterials(),
    }).subscribe({
      next: ({ courses, drafts, areas, materials }) => {
        this.courses = [
          ...courses.map((c) => ({ ...c, publicado: true })),
          ...drafts.map((c) => ({ ...c, publicado: false })),
        ];
        this.areas = areas;
        this.materials = materials;
        this.loading = false;
      },
      error: (err: Error) => {
        this.loading = false;
        this.toast.add({
          severity: "error",
          summary: "Error al cargar",
          detail: err.message,
        });
      },
    });
  }

  get areaOptions(): Option[] {
    return this.areas.map((a) => ({ label: a.name, value: a.id }));
  }
  get materialOptions(): Option[] {
    return this.materials.map((m) => ({
      label: `${m.title}${m.type ? " (" + m.type + ")" : ""}`,
      value: m.id,
    }));
  }

  descPreview(c: Course): string {
    const t = richTextToPlain(c.description);
    return t.length > 90 ? t.slice(0, 90) + "…" : t;
  }
  teacherName(id: string): string {
    return this.teacherOptions.find((o) => o.value === id)?.label ?? id;
  }

  // ==================== CURSOS ====================
  private emptyCourseForm() {
    return {
      title: "",
      description: "",
      mode: "Presencial",
      duration: "",
      trainingAreaId: null as string | null,
      materialIds: [] as string[],
      responsibleTeacher: null as string | null,
      publicar: true,
    };
  }

  newCourse() {
    this.editingCourseId = null;
    this.courseForm = this.emptyCourseForm();
    this.courseDialog = true;
  }

  /** Detalle del curso con el CRUD de actividades y sus entregas. */
  openDetail(c: Course) {
    this.router.navigate(["/dashboard/cursos", c.id]);
  }

  editCourse(c: Course) {
    this.editingCourseId = c.id;
    this.courseForm = {
      title: c.title,
      description: richTextToPlain(c.description),
      mode: c.mode || "Presencial",
      duration: c.duration ?? "",
      trainingAreaId: c.trainingArea?.id ?? null,
      materialIds: (c.material ?? []).map((m) => m.id),
      responsibleTeacher: c.responsibleTeacher ?? null,
      publicar: true,
    };
    this.courseDialog = true;
  }

  get courseFormValid(): boolean {
    const f = this.courseForm;
    return !!(
      f.title.trim() &&
      f.mode &&
      f.duration.trim() &&
      f.trainingAreaId &&
      f.responsibleTeacher
    );
  }

  saveCourse() {
    if (!this.courseFormValid) return;
    const f = this.courseForm;
    const payload = {
      title: f.title.trim(),
      description: plainToRichText(f.description),
      mode: f.mode,
      duration: f.duration.trim() || undefined,
      trainingAreaId: f.trainingAreaId!,
      materialIds: f.materialIds,
      responsibleTeacher: f.responsibleTeacher!,
    };
    this.savingCourse = true;

    if (this.editingCourseId) {
      this.education.updateCourse(this.editingCourseId, payload).subscribe({
        next: () => {
          this.savingCourse = false;
          this.courseDialog = false;
          this.toast.add({
            severity: "success",
            summary: "Curso actualizado",
            detail: "Se modificó el curso correctamente",
          });
          this.reloadAll();
        },
        error: (err: Error) => this.saveError(err),
      });
      return;
    }

    this.education.createCourse(payload).subscribe({
      next: (created) => {
        if (f.publicar) {
          this.education.publishCourse(created.id).subscribe({
            next: () => this.afterCourseCreate(true),
            // el curso YA se creó (queda en borrador): no es un error de
            // guardado, es que Contentful no lo dejó publicar todavía
            error: (err: Error) => this.afterCourseCreate(false, err.message),
          });
        } else {
          this.afterCourseCreate(false);
        }
      },
      error: (err: Error) => this.saveError(err),
    });
  }

  private afterCourseCreate(published: boolean, publishError?: string) {
    this.savingCourse = false;
    this.courseDialog = false;
    this.toast.add(
      publishError
        ? {
            severity: "warn",
            summary: "Curso creado en borrador",
            detail: `El curso se guardó, pero no se pudo publicar: ${publishError}`,
            life: 8000,
          }
        : {
            severity: "success",
            summary: "Curso creado",
            detail: published
              ? "Se creó el curso correctamente y quedó visible en el catálogo"
              : "Se creó el curso en borrador. Publicalo para que aparezca en el catálogo",
          },
    );
    this.reloadAll();
  }

  togglePublish(c: Course, publish: boolean) {
    const call = publish
      ? this.education.publishCourse(c.id)
      : this.education.unpublishCourse(c.id);
    call.subscribe({
      next: () => {
        this.toast.add({
          severity: "success",
          summary: publish ? "Curso publicado" : "Curso despublicado",
          detail: c.title,
        });
        this.reloadAll();
      },
      error: (err: Error) =>
        this.toast.add({
          severity: "error",
          summary: "Error",
          detail: err.message,
        }),
    });
  }

  /**
   * Baja lógica (HU-16): antes de confirmar, el backend dice cuántos
   * inscriptos activos tiene el curso para avisarlo. Si ese conteo falla,
   * se pregunta igual, sin el número.
   */
  deleteCourse(c: Course) {
    this.education.countActiveEnrollments(c.id).subscribe({
      next: (n) => this.confirmarBaja(c, n),
      error: () => this.confirmarBaja(c, null),
    });
  }

  private confirmarBaja(c: Course, inscriptos: number | null) {
    const aviso =
      inscriptos === null
        ? "No pudimos contar los inscriptos."
        : inscriptos === 0
          ? "No tiene participantes inscriptos."
          : `Tiene ${inscriptos} ${inscriptos === 1 ? "participante inscripto" : "participantes inscriptos"}: lo van a seguir viendo en sus cursos.`;
    this.confirm.confirm({
      header: "Eliminar curso",
      message: `¿Eliminar "${c.title}"? ${aviso} El curso deja de verse en el catálogo y no acepta inscripciones nuevas, pero no se borra: queda guardado y la baja se registra en la auditoría.`,
      acceptLabel: "Eliminar",
      rejectLabel: "Cancelar",
      acceptButtonStyleClass: "p-button-danger",
      accept: () => {
        this.education.deleteCourse(c.id).subscribe({
          next: () => {
            this.toast.add({
              severity: "success",
              summary: "Curso eliminado",
              detail: "Se eliminó el curso correctamente",
            });
            this.reloadAll();
          },
          error: (err: Error) =>
            this.toast.add({
              severity: "error",
              summary: "Error",
              detail: err.message,
            }),
        });
      },
    });
  }

  // ==================== ÁREAS ====================
  newArea() {
    this.editingAreaId = null;
    this.areaForm = { name: "", description: "" };
    this.areaDialog = true;
  }
  editArea(a: TrainingArea) {
    this.editingAreaId = a.id;
    this.areaForm = { name: a.name, description: a.description ?? "" };
    this.areaDialog = true;
  }
  saveArea() {
    if (!this.areaForm.name.trim()) return;
    const input = {
      name: this.areaForm.name.trim(),
      description: this.areaForm.description.trim() || undefined,
    };
    this.savingArea = true;
    const call = this.editingAreaId
      ? this.education.updateTrainingArea(this.editingAreaId, input)
      : this.education.createTrainingArea(input);
    call.subscribe({
      next: () => {
        this.savingArea = false;
        this.areaDialog = false;
        this.toast.add({
          severity: "success",
          summary: this.editingAreaId ? "Área actualizada" : "Área creada",
          detail: input.name,
        });
        this.reloadAll();
      },
      error: (err: Error) => {
        this.savingArea = false;
        this.toast.add({
          severity: "error",
          summary: "Error",
          detail: err.message,
        });
      },
    });
  }
  deleteArea(a: TrainingArea) {
    this.confirm.confirm({
      header: "Eliminar área formativa",
      message: `¿Eliminar el área "${a.name}"? Los cursos que la usan quedarán sin área.`,
      acceptLabel: "Eliminar",
      rejectLabel: "Cancelar",
      acceptButtonStyleClass: "p-button-danger",
      accept: () => {
        this.education.deleteTrainingArea(a.id).subscribe({
          next: () => {
            this.toast.add({
              severity: "success",
              summary: "Área eliminada",
              detail: a.name,
            });
            this.reloadAll();
          },
          error: (err: Error) =>
            this.toast.add({
              severity: "error",
              summary: "Error",
              detail: err.message,
            }),
        });
      },
    });
  }

  // ==================== MATERIALES ====================
  /**
   * true cuando el diálogo de material se abrió desde el formulario de
   * curso: el material nuevo queda adjuntado a ese curso (Problema 1 de
   * la auditoría: subir materiales mientras se crea el curso).
   */
  materialForCourse = false;

  /** Abre el alta de material desde el diálogo de curso. */
  newMaterialForCourse() {
    this.newMaterial();
    this.materialForCourse = true;
  }

  /**
   * Después de crear un material: lo suma a la lista y, si se abrió
   * desde el curso, lo deja seleccionado en el formulario. No recarga
   * todo para no perder lo que se está escribiendo en el curso.
   */
  private materialCreado(m: Material) {
    if (!this.materials.some((x) => x.id === m.id)) {
      this.materials = [...this.materials, m];
    }
    if (this.materialForCourse && !this.courseForm.materialIds.includes(m.id)) {
      this.courseForm.materialIds = [...this.courseForm.materialIds, m.id];
    }
    this.materialForCourse = false;
  }

  newMaterial() {
    this.materialForCourse = false;
    this.editingMaterialId = null;
    this.materialForm = { title: "", type: "PDF", fileAssetId: "", url: "" };
    this.materialFile = null;
    this.materialFileError = "";
    this.materialDialog = true;
  }
  editMaterial(m: Material) {
    this.materialForCourse = false;
    this.editingMaterialId = m.id;
    this.materialForm = {
      title: m.title,
      type: (m.type as MaterialType) ?? "PDF",
      fileAssetId: "",
      url: m.url ?? "",
    };
    this.materialFile = null;
    this.materialFileError = "";
    this.materialDialog = true;
  }

  get materialUrlInvalid(): boolean {
    const u = this.materialForm.url.trim();
    return !!u && !/^https?:\/\//i.test(u);
  }

  onMaterialTypeChange() {
    // el campo que no corresponde al tipo elegido no se manda al backend
    this.materialForm.url = "";
    this.materialForm.fileAssetId = "";
    this.materialFile = null;
    this.materialFileError = "";
  }

  onMaterialFile(event: Event) {
    const input = event.target as HTMLInputElement;
    const f = input.files?.[0] ?? null;
    this.materialFileError = "";
    if (!f) {
      this.materialFile = null;
      return;
    }
    const limitMb = this.uploadLimits[f.type];
    if (!limitMb) {
      this.materialFileError =
        "Tipo no permitido. Se aceptan JPG, PNG, WEBP, PDF, DOC/DOCX. Para video usá el campo URL.";
      this.materialFile = null;
      return;
    }
    if (f.size > limitMb * 1024 * 1024) {
      this.materialFileError = `El archivo pesa ${(f.size / 1024 / 1024).toFixed(1)} MB; el máximo para este tipo es ${limitMb} MB.`;
      this.materialFile = null;
      return;
    }
    this.materialFile = f;
    if (!this.materialForm.title.trim()) {
      this.materialForm.title = f.name.replace(/\.[^.]+$/, "");
    }
  }

  saveMaterial() {
    if (!this.materialForm.title.trim() || this.materialUrlInvalid) return;

    // camino subida de archivo (solo alta)
    if (this.materialFile && !this.editingMaterialId) {
      this.savingMaterial = true;
      this.education
        .uploadMaterial(
          this.materialFile,
          this.materialForm.title.trim(),
          this.materialForm.type,
        )
        .subscribe({
          next: (m) => {
            this.savingMaterial = false;
            this.materialDialog = false;
            this.toast.add({
              severity: "success",
              summary: "Material subido",
              detail: this.materialForm.title,
            });
            this.materialCreado(m);
          },
          error: (err: Error) => {
            this.savingMaterial = false;
            this.toast.add({
              severity: "error",
              summary: "Error al subir",
              detail: err.message,
            });
          },
        });
      return;
    }

    const input = {
      title: this.materialForm.title.trim(),
      type: this.materialForm.type,
      fileAssetId: this.materialForm.fileAssetId.trim() || undefined,
      url: this.materialForm.url.trim() || undefined,
    };
    this.savingMaterial = true;
    const call = this.editingMaterialId
      ? this.education.updateMaterial(this.editingMaterialId, input)
      : this.education.createMaterial(input);
    const editing = !!this.editingMaterialId;
    call.subscribe({
      next: (m) => {
        this.savingMaterial = false;
        this.materialDialog = false;
        this.toast.add({
          severity: "success",
          summary: editing ? "Material actualizado" : "Material creado",
          detail: input.title,
        });
        if (editing) {
          this.reloadAll();
        } else {
          this.materialCreado(m);
        }
      },
      error: (err: Error) => {
        this.savingMaterial = false;
        this.toast.add({
          severity: "error",
          summary: "Error",
          detail: err.message,
        });
      },
    });
  }
  deleteMaterial(m: Material) {
    this.confirm.confirm({
      header: "Eliminar material",
      message: `¿Eliminar el material "${m.title}"?`,
      acceptLabel: "Eliminar",
      rejectLabel: "Cancelar",
      acceptButtonStyleClass: "p-button-danger",
      accept: () => {
        this.education.deleteMaterial(m.id).subscribe({
          next: () => {
            this.toast.add({
              severity: "success",
              summary: "Material eliminado",
              detail: m.title,
            });
            this.reloadAll();
          },
          error: (err: Error) =>
            this.toast.add({
              severity: "error",
              summary: "Error",
              detail: err.message,
            }),
        });
      },
    });
  }

  private saveError(err: Error) {
    this.savingCourse = false;
    this.toast.add({
      severity: "error",
      summary: "No se pudo guardar",
      detail: err.message,
    });
  }
}
