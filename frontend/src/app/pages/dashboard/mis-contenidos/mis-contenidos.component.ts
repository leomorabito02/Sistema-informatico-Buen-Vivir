import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { DialogModule } from "primeng/dialog";
import { InputTextModule } from "primeng/inputtext";
import { TextareaModule } from "primeng/textarea";
import { ToastModule } from "primeng/toast";
import { ConfirmDialogModule } from "primeng/confirmdialog";
import { ConfirmationService, MessageService } from "primeng/api";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import {
  ContenidoDocente,
  ContenidoPublico,
  ContenidosService,
  EstadoModeracion,
  guardarArchivo,
  tamanoLegible,
} from "../../../services/contenidos.service";

const ACEPTA =
  ".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.jpg,.jpeg,.png,.webp";
const MAX_BYTES = 10 * 1024 * 1024;

/**
 * Contenido del docente (HU-23): sube material, que queda pendiente de
 * moderación; lo edita o elimina mientras siga pendiente, y ve si se
 * aprobó o se rechazó. Abajo, el material aprobado del centro (HU-22).
 * Es la única pantalla del docente externo.
 */
@Component({
  selector: "app-mis-contenidos",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ButtonModule,
    TagModule,
    DialogModule,
    InputTextModule,
    TextareaModule,
    ToastModule,
    ConfirmDialogModule,
    EmptyStateComponent,
  ],
  providers: [MessageService, ConfirmationService],
  templateUrl: "./mis-contenidos.component.html",
  styleUrls: ["./mis-contenidos.component.css"],
})
export class MisContenidosComponent implements OnInit {
  private readonly service = inject(ContenidosService);
  private readonly toast = inject(MessageService);
  private readonly confirm = inject(ConfirmationService);

  readonly acepta = ACEPTA;
  readonly tamano = tamanoLegible;

  contenidos: ContenidoDocente[] = [];
  cargando = true;
  error = "";

  biblioteca: ContenidoPublico[] = [];
  errorBiblioteca = "";

  dialogo = false;
  editandoId: string | null = null;
  titulo = "";
  descripcion = "";
  archivo: File | null = null;
  errorForm = "";
  guardando = false;

  ngOnInit(): void {
    this.cargar();
    this.cargarBiblioteca();
  }

  cargar(): void {
    this.cargando = true;
    this.error = "";
    this.service.misContenidos().subscribe({
      next: (lista) => {
        this.contenidos = lista ?? [];
        this.cargando = false;
      },
      error: (e: Error) => {
        this.error = e.message;
        this.cargando = false;
      },
    });
  }

  cargarBiblioteca(): void {
    this.service.biblioteca().subscribe({
      next: (lista) => (this.biblioteca = lista ?? []),
      error: (e: Error) => (this.errorBiblioteca = e.message),
    });
  }

  etiqueta(estado: EstadoModeracion): string {
    return {
      pendiente: "Pendiente de moderación",
      aprobado: "Aprobado",
      rechazado: "Rechazado",
    }[estado];
  }

  severidad(estado: EstadoModeracion): "warn" | "success" | "danger" {
    return (
      { pendiente: "warn", aprobado: "success", rechazado: "danger" } as const
    )[estado];
  }

  fecha(iso: string): string {
    const d = new Date(iso);
    return isNaN(d.getTime())
      ? ""
      : d.toLocaleDateString("es-AR", {
          day: "2-digit",
          month: "2-digit",
          year: "numeric",
        });
  }

  nuevo(): void {
    this.editandoId = null;
    this.titulo = "";
    this.descripcion = "";
    this.archivo = null;
    this.errorForm = "";
    this.dialogo = true;
  }

  editar(c: ContenidoDocente): void {
    this.editandoId = c.id;
    this.titulo = c.titulo;
    this.descripcion = c.descripcion;
    this.archivo = null;
    this.errorForm = "";
    this.dialogo = true;
  }

  elegirArchivo(ev: Event): void {
    const input = ev.target as HTMLInputElement;
    const f = input.files?.[0] ?? null;
    this.errorForm = "";
    if (f && f.size > MAX_BYTES) {
      this.errorForm = "El archivo puede pesar hasta 10 MB";
      input.value = "";
      this.archivo = null;
      return;
    }
    this.archivo = f;
  }

  guardar(): void {
    if (this.titulo.trim().length < 3) {
      this.errorForm = "El título tiene que tener al menos 3 caracteres";
      return;
    }
    if (!this.editandoId && !this.archivo) {
      this.errorForm = "Adjuntá el archivo del material";
      return;
    }
    this.errorForm = "";
    this.guardando = true;
    const datos = {
      titulo: this.titulo,
      descripcion: this.descripcion,
      archivo: this.archivo,
    };
    const pedido = this.editandoId
      ? this.service.editar(this.editandoId, datos)
      : this.service.subir(datos);
    pedido.subscribe({
      next: () => {
        this.guardando = false;
        this.dialogo = false;
        this.toast.add({
          severity: "success",
          summary: this.editandoId
            ? "Material actualizado"
            : "Material enviado",
          detail: this.editandoId
            ? "Los cambios se guardaron."
            : "Queda pendiente hasta que la administración lo revise.",
        });
        this.cargar();
      },
      error: (e: Error) => {
        this.guardando = false;
        this.errorForm = e.message;
      },
    });
  }

  eliminar(c: ContenidoDocente): void {
    this.confirm.confirm({
      header: "Eliminar material",
      message: `¿Eliminar «${c.titulo}»? No se puede deshacer.`,
      acceptLabel: "Eliminar",
      rejectLabel: "Cancelar",
      accept: () =>
        this.service.eliminar(c.id).subscribe({
          next: () => {
            this.contenidos = this.contenidos.filter((x) => x.id !== c.id);
            this.toast.add({
              severity: "success",
              summary: "Material eliminado",
            });
          },
          error: (e: Error) =>
            this.toast.add({
              severity: "error",
              summary: "No se pudo eliminar",
              detail: e.message,
            }),
        }),
    });
  }

  descargar(c: { id: string; archivoNombre: string }): void {
    this.service.archivo(c.id).subscribe({
      next: (blob) => guardarArchivo(blob, c.archivoNombre),
      error: (e: Error) =>
        this.toast.add({
          severity: "error",
          summary: "No se pudo descargar",
          detail: e.message,
        }),
    });
  }
}
