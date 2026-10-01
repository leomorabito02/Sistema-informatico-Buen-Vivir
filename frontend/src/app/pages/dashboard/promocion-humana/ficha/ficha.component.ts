import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { ActivatedRoute, Router, RouterLink } from "@angular/router";
import { ButtonModule } from "primeng/button";
import { CardModule } from "primeng/card";
import { CheckboxModule } from "primeng/checkbox";
import { DatePickerModule } from "primeng/datepicker";
import { DialogModule } from "primeng/dialog";
import { MessageModule } from "primeng/message";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { SelectModule } from "primeng/select";
import { TagModule } from "primeng/tag";
import { TextareaModule } from "primeng/textarea";
import { ToastModule } from "primeng/toast";
import { MessageService } from "primeng/api";
import { forkJoin, of } from "rxjs";
import { catchError } from "rxjs/operators";

import {
  HumanPromotionError,
  HumanPromotionService,
} from "../../../../services/human-promotion.service";
import { AuthService } from "../../../../services/auth.service";
import { UsersService } from "../../../../services/users.service";
import {
  FichaDetalle,
  HitoDetalle,
  HitoInput,
  MAX_DESCRIPCION_HITO,
  MAX_TEXTO_FICHA,
  MSG_CAMPOS_OBLIGATORIOS,
} from "../../../../core/models/human-promotion.model";
import {
  formatDMY,
  formatYMD,
  parseYMD,
  todayYMD,
} from "../../../../shared/utils/date-ymd";

interface Opcion {
  label: string;
  value: string;
}

interface HitoForm {
  fecha: Date | null;
  descripcion: string;
  responsableId: string;
  visibleParticipante: boolean;
}

/**
 * Ficha de proceso personal de un participante (HU-29 a HU-33, HU-37).
 *
 *  - Si el participante no tiene ficha (404), muestra el alta.
 *  - Nombre, apellido, edad y habilidades son de solo lectura: vienen de
 *    Auth y se editan en Gestión de usuarios (HU-31).
 *  - Situación actual y avances son obligatorios y editables.
 *  - Línea de tiempo con todos los hitos en orden cronológico, alta y
 *    edición en un diálogo, y exportación a PDF generada en el backend.
 */
@Component({
  selector: "app-ficha",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    ButtonModule,
    CardModule,
    CheckboxModule,
    DatePickerModule,
    DialogModule,
    MessageModule,
    ProgressSpinnerModule,
    SelectModule,
    TagModule,
    TextareaModule,
    ToastModule,
  ],
  providers: [MessageService],
  templateUrl: "./ficha.component.html",
  styleUrls: ["./ficha.component.css"],
})
export class FichaComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly service = inject(HumanPromotionService);
  private readonly auth = inject(AuthService);
  private readonly users = inject(UsersService);
  private readonly toast = inject(MessageService);

  readonly maxTexto = MAX_TEXTO_FICHA;
  readonly maxDescripcion = MAX_DESCRIPCION_HITO;
  readonly formatDMY = formatDMY;

  participanteId = "";
  /** Nombre que llega desde la búsqueda; sirve para el alta, cuando todavía no hay ficha. */
  nombreDesdeBusqueda = "";

  loading = true;
  error = "";
  ficha: FichaDetalle | null = null;
  sinFicha = false;

  // ---- formulario de ficha (alta y edición) ----
  editando = false;
  guardando = false;
  intentoGuardar = false;
  situacionActual = "";
  avances = "";

  // ---- hitos ----
  hitos: HitoDetalle[] = [];
  cargandoHitos = false;
  errorHitos = "";
  dialogoHito = false;
  hitoEditandoId: string | null = null;
  guardandoHito = false;
  intentoGuardarHito = false;
  hitoForm: HitoForm = this.hitoFormVacio();
  responsables: Opcion[] = [];

  descargandoPdf = false;

  private usuarioActualId = "";

  constructor() {
    const state = this.router.getCurrentNavigation()?.extras.state as
      { nombre?: string } | undefined;
    this.nombreDesdeBusqueda = state?.nombre ?? "";
  }

  ngOnInit() {
    this.participanteId =
      this.route.snapshot.paramMap.get("participanteId") ?? "";
    this.cargarUsuarioYResponsables();
    this.cargarFicha();
  }

  // ---------------------------------------------------------------------
  // Ficha
  // ---------------------------------------------------------------------
  get nombreCompleto(): string {
    const p = this.ficha?.participante;
    if (p) return `${p.nombre} ${p.apellido}`.trim();
    return this.nombreDesdeBusqueda || "Participante";
  }

  cargarFicha() {
    this.loading = true;
    this.error = "";
    this.sinFicha = false;
    this.service.getFichaPorParticipante(this.participanteId).subscribe({
      next: (ficha) => {
        this.ficha = ficha;
        this.loading = false;
        this.cargarHitos();
      },
      error: (err: HumanPromotionError) => {
        this.loading = false;
        if (err.status === 404) {
          this.sinFicha = true;
          this.editando = true;
          return;
        }
        this.error = err.message;
      },
    });
  }

  empezarEdicion() {
    if (!this.ficha) return;
    this.situacionActual = this.ficha.situacion_actual;
    this.avances = this.ficha.avances;
    this.intentoGuardar = false;
    this.editando = true;
  }

  cancelarEdicion() {
    this.editando = false;
    this.intentoGuardar = false;
  }

  get faltanCamposFicha(): boolean {
    return !this.situacionActual.trim() || !this.avances.trim();
  }

  guardarFicha() {
    this.intentoGuardar = true;
    if (this.faltanCamposFicha) {
      this.toast.add({
        severity: "warn",
        summary: "Faltan datos",
        detail: MSG_CAMPOS_OBLIGATORIOS,
      });
      return;
    }
    const situacion_actual = this.situacionActual.trim();
    const avances = this.avances.trim();
    this.guardando = true;

    if (this.sinFicha) {
      this.service
        .crearFicha({
          participante_id: this.participanteId,
          situacion_actual,
          avances,
        })
        .subscribe({
          next: () => {
            this.guardando = false;
            this.editando = false;
            this.toast.add({
              severity: "success",
              summary: "Ficha creada",
              detail: "La ficha de proceso personal quedó registrada.",
            });
            this.cargarFicha();
          },
          error: (err: Error) => this.errorAlGuardar(err),
        });
      return;
    }

    this.service
      .editarFicha(this.ficha!.id, { situacion_actual, avances })
      .subscribe({
        next: (res) => {
          this.guardando = false;
          this.editando = false;
          this.ficha = { ...this.ficha!, ...res.ficha };
          this.toast.add({
            severity: "success",
            summary: "Cambios guardados",
            detail: res.message,
          });
        },
        error: (err: Error) => this.errorAlGuardar(err),
      });
  }

  private errorAlGuardar(err: Error) {
    this.guardando = false;
    this.toast.add({
      severity: "error",
      summary: "No se pudo guardar",
      detail: err.message,
    });
  }

  // ---------------------------------------------------------------------
  // PDF (HU-37)
  // ---------------------------------------------------------------------
  exportarPdf() {
    if (!this.ficha) return;
    this.descargandoPdf = true;
    this.service.descargarPdf(this.ficha.id).subscribe({
      next: (blob) => {
        this.descargandoPdf = false;
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `ficha-${this.nombreArchivo()}.pdf`;
        a.click();
        URL.revokeObjectURL(url);
      },
      error: (err: Error) => {
        this.descargandoPdf = false;
        this.toast.add({
          severity: "error",
          summary: "No se pudo exportar",
          detail: err.message,
        });
      },
    });
  }

  private nombreArchivo(): string {
    const base = this.nombreCompleto
      .normalize("NFD")
      .replace(/[̀-ͯ]/g, "")
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-|-$/g, "");
    return base || this.ficha!.id;
  }

  // ---------------------------------------------------------------------
  // Hitos (HU-32, HU-33)
  // ---------------------------------------------------------------------
  cargarHitos() {
    if (!this.ficha) return;
    this.cargandoHitos = true;
    this.errorHitos = "";
    this.service.listarHitos(this.ficha.id).subscribe({
      next: (hitos) => {
        this.hitos = hitos;
        this.cargandoHitos = false;
      },
      error: (err: Error) => {
        this.errorHitos = err.message;
        this.cargandoHitos = false;
      },
    });
  }

  nuevoHito() {
    this.hitoEditandoId = null;
    this.hitoForm = this.hitoFormVacio();
    this.intentoGuardarHito = false;
    this.dialogoHito = true;
  }

  editarHito(h: HitoDetalle) {
    this.hitoEditandoId = h.id;
    this.asegurarResponsable(h.responsable_id, h.responsable_nombre);
    this.hitoForm = {
      fecha: parseYMD(h.fecha),
      descripcion: h.descripcion,
      responsableId: h.responsable_id,
      visibleParticipante: h.visible_participante,
    };
    this.intentoGuardarHito = false;
    this.dialogoHito = true;
  }

  get faltanCamposHito(): boolean {
    return !this.hitoForm.fecha || !this.hitoForm.descripcion.trim();
  }

  guardarHito() {
    this.intentoGuardarHito = true;
    if (!this.ficha || this.faltanCamposHito) return;

    const input: HitoInput = {
      fecha: formatYMD(this.hitoForm.fecha!),
      descripcion: this.hitoForm.descripcion.trim(),
      responsable_id: this.hitoForm.responsableId || undefined,
      visible_participante: this.hitoForm.visibleParticipante,
    };
    this.guardandoHito = true;
    const req$ = this.hitoEditandoId
      ? this.service.editarHito(this.ficha.id, this.hitoEditandoId, input)
      : this.service.crearHito(this.ficha.id, input);

    req$.subscribe({
      next: () => {
        this.guardandoHito = false;
        this.dialogoHito = false;
        this.toast.add({
          severity: "success",
          summary: this.hitoEditandoId ? "Hito modificado" : "Hito registrado",
          detail: "La línea de tiempo quedó actualizada.",
        });
        this.cargarHitos();
      },
      error: (err: Error) => {
        this.guardandoHito = false;
        this.toast.add({
          severity: "error",
          summary: "No se pudo guardar el hito",
          detail: err.message,
        });
      },
    });
  }

  private hitoFormVacio(): HitoForm {
    return {
      fecha: parseYMD(todayYMD()),
      descripcion: "",
      responsableId: this.usuarioActualId,
      visibleParticipante: false,
    };
  }

  /**
   * Responsable por defecto = usuario logueado (HU-32). Si además el
   * usuario puede listar usuarios (hoy, solo admin en auth-service), se
   * ofrecen como responsables los administradores y voluntarios.
   */
  private cargarUsuarioYResponsables() {
    forkJoin({
      me: this.auth.isAuthenticated(),
      usuarios: this.users.listUsers().pipe(catchError(() => of([]))),
    }).subscribe(({ me, usuarios }) => {
      const yo = me.user;
      this.usuarioActualId = yo?.user_id ?? "";
      const opciones = new Map<string, string>();
      if (yo) {
        opciones.set(
          yo.user_id,
          `${yo.first_name} ${yo.last_name} (yo)`.trim(),
        );
      }
      usuarios
        .filter(
          (u) => u.role_id === "role_admin" || u.role_id === "role_volunteer",
        )
        .forEach((u) => {
          if (!opciones.has(u.user_id)) {
            opciones.set(u.user_id, `${u.first_name} ${u.last_name}`.trim());
          }
        });
      this.responsables = [...opciones].map(([value, label]) => ({
        value,
        label,
      }));
      if (!this.hitoForm.responsableId) {
        this.hitoForm.responsableId = this.usuarioActualId;
      }
    });
  }

  /** Al editar, el responsable actual del hito tiene que estar entre las opciones. */
  private asegurarResponsable(id: string, nombre?: string) {
    if (!this.responsables.some((r) => r.value === id)) {
      this.responsables = [
        ...this.responsables,
        { value: id, label: nombre || "Responsable actual" },
      ];
    }
  }
}
