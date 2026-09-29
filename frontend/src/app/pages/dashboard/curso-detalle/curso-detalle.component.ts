import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { ActivatedRoute, Router } from "@angular/router";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { DialogModule } from "primeng/dialog";
import { ConfirmDialogModule } from "primeng/confirmdialog";
import { DatePickerModule } from "primeng/datepicker";
import { InputTextModule } from "primeng/inputtext";
import { TextareaModule } from "primeng/textarea";
import { TableModule } from "primeng/table";
import { ProgressBarModule } from "primeng/progressbar";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { MessageModule } from "primeng/message";
import { TooltipModule } from "primeng/tooltip";
import { ToastModule } from "primeng/toast";
import { ConfirmationService, MessageService } from "primeng/api";
import { Observable, catchError, map, of, switchMap } from "rxjs";

import { AuthService } from "../../../services/auth.service";
import { EducationService } from "../../../services/education.service";
import { UsersService } from "../../../services/users.service";
import {
  Activity,
  ActivityInput,
  Course,
  ParticipantActivity,
  ParticipantProgress,
  Submission,
  SubmissionFile,
  SubmissionStatusRow,
} from "../../../core/models/education.model";
import { RichTextComponent } from "../../../shared/components/rich-text.component";
import {
  activityFechaToDate,
  toActivityFecha,
} from "../../../shared/utils/date-ymd";

/**
 * - gestion: role_admin / role_dev / role_volunteer — cualquier curso,
 *   CRUD de actividades vía /api/cursos y /api/actividades.
 * - docente: role_teacher — solo sus cursos publicados, mismo CRUD vía
 *   /api/docente (el backend valida que sea el docente responsable).
 * - participante: role_participant — solo cursos donde está inscripto;
 *   ve las actividades y las entrega.
 */
type Modo = "gestion" | "docente" | "participante" | "sin-acceso";

interface ActividadForm {
  titulo: string;
  consigna: string;
  fecha: Date | null;
}

const MAX_ARCHIVOS = 5;
const MAX_BYTES_ARCHIVO = 10 * 1024 * 1024;
const EXTENSIONES_PERMITIDAS = [
  ".pdf",
  ".doc",
  ".docx",
  ".xls",
  ".xlsx",
  ".ppt",
  ".pptx",
  ".txt",
  ".jpg",
  ".jpeg",
  ".png",
  ".webp",
];

/**
 * Detalle de un curso (`/dashboard/cursos/:id`): datos del curso arriba
 * y sus actividades abajo. Admin, voluntario y docente responsable
 * gestionan las actividades (crear, editar, eliminar) y ven las
 * entregas de cada una; el participante ve las actividades de un curso
 * en el que está inscripto y las entrega. Cada entrega marca la
 * actividad como cumplida y genera la asistencia del participante (ver
 * education/CLAUDE.md, sección Entregas).
 */
@Component({
  selector: "app-curso-detalle",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ButtonModule,
    TagModule,
    DialogModule,
    ConfirmDialogModule,
    DatePickerModule,
    InputTextModule,
    TextareaModule,
    TableModule,
    ProgressBarModule,
    ProgressSpinnerModule,
    MessageModule,
    TooltipModule,
    ToastModule,
    RichTextComponent,
  ],
  providers: [MessageService, ConfirmationService],
  templateUrl: "./curso-detalle.component.html",
  styleUrls: ["./curso-detalle.component.css"],
})
export class CursoDetalleComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private authService = inject(AuthService);
  private education = inject(EducationService);
  private users = inject(UsersService);
  private toast = inject(MessageService);
  private confirm = inject(ConfirmationService);

  readonly extensionesPermitidas = EXTENSIONES_PERMITIDAS.join(",");
  readonly maxArchivos = MAX_ARCHIVOS;

  cursoId = "";
  modo: Modo = "sin-acceso";
  loading = true;
  error = "";

  curso: Course | null = null;
  docenteNombre = "";
  actividades: ParticipantActivity[] = [];
  progreso: ParticipantProgress | null = null;

  // --- gestión: alta / edición de actividad ---
  formVisible = false;
  editando: Activity | null = null;
  form: ActividadForm = { titulo: "", consigna: "", fecha: null };
  guardando = false;

  // --- gestión: entregas de una actividad ---
  entregasVisible = false;
  entregasActividad: Activity | null = null;
  entregas: SubmissionStatusRow[] = [];
  loadingEntregas = false;
  expandedRows: Record<string, boolean> = {};

  // --- participante: realizar / ver entrega ---
  entregaVisible = false;
  entregaActividad: ParticipantActivity | null = null;
  entregaTexto = "";
  entregaArchivos: File[] = [];
  entregando = false;

  descargando: string | null = null;

  ngOnInit() {
    this.cursoId = this.route.snapshot.paramMap.get("id") ?? "";
    this.authService.isAuthenticated().subscribe({
      next: (res) => {
        const roleId = res.role_id ?? res.user?.role_id ?? "";
        this.modo = modoPorRol(roleId);
        this.cargar();
      },
      error: () => {
        this.loading = false;
        this.error = "No se pudo verificar tu sesión. Volvé a iniciar sesión.";
      },
    });
  }

  get puedeGestionar(): boolean {
    return this.modo === "gestion" || this.modo === "docente";
  }

  get esParticipante(): boolean {
    return this.modo === "participante";
  }

  get entregadasCount(): number {
    return this.entregas.filter((e) => e.estado === "entregado").length;
  }

  cargar() {
    this.loading = true;
    this.error = "";
    if (this.modo === "sin-acceso") {
      this.loading = false;
      this.error = "Tu rol no tiene acceso al detalle de los cursos.";
      return;
    }

    const carga$: Observable<void> =
      this.modo === "participante"
        ? this.education.getMyCourse(this.cursoId).pipe(
            map((detalle) => {
              this.curso = detalle.curso;
              this.actividades = detalle.actividades;
              this.progreso = detalle.progreso;
            }),
          )
        : this.cargarCurso().pipe(
            switchMap((curso) => {
              this.curso = curso;
              return this.modo === "docente"
                ? this.education.listActivities(this.cursoId)
                : this.education.listCourseActivities(this.cursoId);
            }),
            map((actividades) => {
              this.actividades = actividades;
            }),
          );

    carga$.subscribe({
      next: () => {
        this.loading = false;
        this.resolverDocente();
      },
      error: (err: Error) => {
        this.loading = false;
        this.error = this.mensajeDeAcceso(err.message);
      },
    });
  }

  /**
   * El curso publicado sale de la CDA (`GET /api/cursos/{id}`). Un
   * administrador también puede abrir un borrador desde "Gestión de
   * cursos": si no está publicado, se busca entre los borradores.
   */
  private cargarCurso(): Observable<Course> {
    return this.education.getCourse(this.cursoId).pipe(
      catchError((err: Error) => {
        if (this.modo !== "gestion") throw err;
        return this.education.listDraftCourses().pipe(
          map((borradores) => {
            const curso = borradores.find((c) => c.id === this.cursoId);
            if (!curso) throw err;
            return curso;
          }),
        );
      }),
    );
  }

  private mensajeDeAcceso(message: string): string {
    if (this.modo === "docente" && /permis|responsable/i.test(message)) {
      return "Solo podés ver el detalle de los cursos donde figurás como docente responsable.";
    }
    if (
      this.modo === "docente" &&
      /no existe|no está publicado/i.test(message)
    ) {
      return "Este curso todavía no está publicado. Vas a poder gestionar sus actividades cuando un administrador lo publique.";
    }
    return message;
  }

  /** Nombre del docente responsable; si auth-service no lo resuelve para este rol, se omite. */
  private resolverDocente() {
    const id = this.curso?.responsibleTeacher;
    if (!id) return;
    this.users
      .listUsers()
      .pipe(catchError(() => of([])))
      .subscribe((lista) => {
        const u = lista.find((x) => x.user_id === id);
        if (u) {
          this.docenteNombre =
            `${u.first_name ?? ""} ${u.last_name ?? ""}`.trim();
        }
      });
  }

  volver() {
    const previa = this.router.lastSuccessfulNavigation?.previousNavigation;
    if (previa) {
      history.back();
      return;
    }
    const destino =
      this.modo === "docente"
        ? "/dashboard/mis-cursos"
        : this.modo === "gestion"
          ? "/dashboard/gestion-cursos"
          : "/dashboard/cursos";
    this.router.navigateByUrl(destino);
  }

  modeSeverity(
    mode: string | undefined,
  ): "success" | "info" | "warn" | "secondary" {
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

  // ---------------------------------------------------------------------
  // Gestión de actividades (admin / voluntario / docente)

  nuevaActividad() {
    this.editando = null;
    this.form = { titulo: "", consigna: "", fecha: null };
    this.formVisible = true;
  }

  editarActividad(a: Activity) {
    this.editando = a;
    this.form = {
      titulo: a.titulo,
      consigna: a.consigna ?? "",
      fecha: activityFechaToDate(a.fecha),
    };
    this.formVisible = true;
  }

  get formValido(): boolean {
    return !!(
      this.form.titulo.trim() &&
      this.form.consigna.trim() &&
      this.form.fecha
    );
  }

  guardarActividad() {
    if (!this.formValido || this.guardando) return;
    const input: ActivityInput = {
      titulo: this.form.titulo.trim(),
      consigna: this.form.consigna.trim(),
      fecha: toActivityFecha(this.form.fecha!),
    };
    const docente = this.modo === "docente";
    const editando = this.editando;
    const req$ = editando
      ? docente
        ? this.education.updateActivity(editando.id, input)
        : this.education.updateCourseActivity(editando.id, input)
      : docente
        ? this.education.createActivity(this.cursoId, input)
        : this.education.createCourseActivity(this.cursoId, input);

    this.guardando = true;
    req$.subscribe({
      next: (actividad) => {
        this.guardando = false;
        this.formVisible = false;
        this.actividades = ordenarPorFecha(
          editando
            ? this.actividades.map((a) =>
                a.id === actividad.id ? actividad : a,
              )
            : [...this.actividades, actividad],
        );
        this.toast.add({
          severity: "success",
          summary: editando ? "Actividad actualizada" : "Actividad creada",
          detail: actividad.titulo,
        });
      },
      error: (err: Error) => {
        this.guardando = false;
        this.toast.add({
          severity: "error",
          summary: "No se pudo guardar la actividad",
          detail: err.message,
        });
      },
    });
  }

  eliminarActividad(a: Activity) {
    this.confirm.confirm({
      header: "Eliminar actividad",
      icon: "pi pi-exclamation-triangle",
      message: `¿Eliminar la actividad "${a.titulo}"? También se eliminan todas sus entregas y los registros de asistencia que generaron, y el progreso de los participantes se recalcula. Esta acción no se puede deshacer.`,
      acceptLabel: "Eliminar",
      rejectLabel: "Cancelar",
      acceptButtonStyleClass: "p-button-danger",
      rejectButtonStyleClass: "p-button-text",
      accept: () => {
        const req$ =
          this.modo === "docente"
            ? this.education.deleteActivity(a.id)
            : this.education.deleteCourseActivity(a.id);
        req$.subscribe({
          next: () => {
            this.actividades = this.actividades.filter((x) => x.id !== a.id);
            this.toast.add({
              severity: "success",
              summary: "Actividad eliminada",
              detail: a.titulo,
            });
          },
          error: (err: Error) => {
            this.toast.add({
              severity: "error",
              summary: "No se pudo eliminar la actividad",
              detail: err.message,
            });
          },
        });
      },
    });
  }

  verEntregas(a: Activity) {
    this.entregasActividad = a;
    this.entregas = [];
    this.expandedRows = {};
    this.entregasVisible = true;
    this.loadingEntregas = true;
    const req$ =
      this.modo === "docente"
        ? this.education.listActivitySubmissions(a.id)
        : this.education.listCourseActivitySubmissions(a.id);
    req$.subscribe({
      next: (rows) => {
        this.entregas = rows;
        this.loadingEntregas = false;
      },
      error: (err: Error) => {
        this.loadingEntregas = false;
        this.toast.add({
          severity: "error",
          summary: "No se pudieron cargar las entregas",
          detail: err.message,
        });
      },
    });
  }

  nombreParticipante(r: SubmissionStatusRow): string {
    return `${r.nombre ?? ""} ${r.apellido ?? ""}`.trim() || r.participanteId;
  }

  // ---------------------------------------------------------------------
  // Participante: realizar y ver entregas

  abrirEntrega(a: ParticipantActivity) {
    this.entregaActividad = a;
    this.entregaTexto = "";
    this.entregaArchivos = [];
    this.entregaVisible = true;
  }

  onArchivosSeleccionados(event: Event) {
    const input = event.target as HTMLInputElement;
    const nuevos = Array.from(input.files ?? []);
    input.value = "";
    for (const f of nuevos) {
      const ext = f.name.slice(f.name.lastIndexOf(".")).toLowerCase();
      if (!EXTENSIONES_PERMITIDAS.includes(ext)) {
        this.avisoArchivo(`"${f.name}" tiene un formato no permitido.`);
        continue;
      }
      if (f.size > MAX_BYTES_ARCHIVO) {
        this.avisoArchivo(`"${f.name}" supera los 10 MB.`);
        continue;
      }
      if (this.entregaArchivos.length >= MAX_ARCHIVOS) {
        this.avisoArchivo(`Podés adjuntar hasta ${MAX_ARCHIVOS} archivos.`);
        break;
      }
      this.entregaArchivos = [...this.entregaArchivos, f];
    }
  }

  quitarArchivo(i: number) {
    this.entregaArchivos = this.entregaArchivos.filter((_, idx) => idx !== i);
  }

  private avisoArchivo(detail: string) {
    this.toast.add({
      severity: "warn",
      summary: "Archivo no agregado",
      detail,
    });
  }

  confirmarEntrega() {
    const a = this.entregaActividad;
    if (!a || this.entregando) return;
    const vacia =
      !this.entregaTexto.trim() && this.entregaArchivos.length === 0;
    this.confirm.confirm({
      header: "Entregar actividad",
      icon: "pi pi-send",
      message: vacia
        ? "Vas a marcar la actividad como entregada sin texto ni archivos. Una vez entregada no vas a poder modificarla ni anular la entrega. ¿Continuar?"
        : "Una vez entregada no vas a poder modificarla ni anular la entrega. ¿Entregar ahora?",
      acceptLabel: "Entregar",
      rejectLabel: "Cancelar",
      rejectButtonStyleClass: "p-button-text",
      accept: () => this.entregar(a),
    });
  }

  private entregar(a: ParticipantActivity) {
    this.entregando = true;
    this.education
      .submitActivity(a.id, this.entregaTexto.trim(), this.entregaArchivos)
      .subscribe({
        next: (entrega) => {
          this.entregando = false;
          const actualizada = { ...a, entrega };
          this.actividades = this.actividades.map((x) =>
            x.id === a.id ? actualizada : x,
          );
          this.entregaActividad = actualizada;
          this.actualizarProgreso();
          this.toast.add({
            severity: "success",
            summary: "Actividad entregada",
            detail: a.titulo,
          });
        },
        error: (err: Error) => {
          this.entregando = false;
          this.toast.add({
            severity: "error",
            summary: "No se pudo entregar",
            detail: err.message,
          });
        },
      });
  }

  private actualizarProgreso() {
    const total = this.actividades.length;
    const entregadas = this.actividades.filter((a) => !!a.entrega).length;
    this.progreso = {
      participanteId: this.progreso?.participanteId ?? "",
      actividadesTotal: total,
      actividadesAsistidas: entregadas,
      porcentajeAsistencia: total > 0 ? (entregadas / total) * 100 : 0,
    };
  }

  // ---------------------------------------------------------------------
  // Adjuntos (todos los roles)

  descargar(entrega: Submission, indice: number, archivo: SubmissionFile) {
    const clave = `${entrega.id}-${indice}`;
    this.descargando = clave;
    this.education.downloadSubmissionFile(entrega.id, indice).subscribe({
      next: (blob) => {
        this.descargando = null;
        const url = URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        link.download = archivo.nombre;
        link.click();
        setTimeout(() => URL.revokeObjectURL(url), 1000);
      },
      error: (err: Error) => {
        this.descargando = null;
        this.toast.add({
          severity: "error",
          summary: "No se pudo descargar el archivo",
          detail: err.message,
        });
      },
    });
  }

  tamanoLegible(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }

  iconoArchivo(tipo: string): string {
    if (tipo === "application/pdf") return "pi-file-pdf";
    if (tipo.startsWith("image/")) return "pi-image";
    if (tipo.includes("spreadsheet") || tipo.includes("excel"))
      return "pi-file-excel";
    if (tipo.includes("word")) return "pi-file-word";
    return "pi-file";
  }
}

function modoPorRol(roleId: string): Modo {
  switch (roleId) {
    case "role_admin":
    case "role_dev":
    case "role_volunteer":
      return "gestion";
    case "role_teacher":
      return "docente";
    case "role_participant":
      return "participante";
    default:
      return "sin-acceso";
  }
}

function ordenarPorFecha<T extends Activity>(actividades: T[]): T[] {
  return [...actividades].sort(
    (a, b) =>
      a.fecha.localeCompare(b.fecha) || a.createdAt.localeCompare(b.createdAt),
  );
}
