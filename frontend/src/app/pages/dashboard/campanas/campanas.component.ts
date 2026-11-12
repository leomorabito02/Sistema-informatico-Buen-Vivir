import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { DialogModule } from "primeng/dialog";
import { SelectModule } from "primeng/select";
import { InputTextModule } from "primeng/inputtext";
import { TextareaModule } from "primeng/textarea";
import { ToastModule } from "primeng/toast";
import { ConfirmDialogModule } from "primeng/confirmdialog";
import { ConfirmationService, MessageService } from "primeng/api";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import { CampaignsService } from "../../../services/campaigns.service";
import {
  Campana,
  CampanaInput,
  EstadoCampana,
  DEFAULT_CAMPAIGN_PLACEHOLDER,
} from "../../../core/models/campaign.model";

/**
 * Gestión de campañas (HU-47 a HU-49), para administración y voluntariado:
 * crear (quedan en borrador), editar en cualquier estado, publicar y dar
 * de baja (baja lógica, con confirmación).
 */
@Component({
  selector: "app-campanas",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ButtonModule,
    TagModule,
    DialogModule,
    SelectModule,
    InputTextModule,
    TextareaModule,
    ToastModule,
    ConfirmDialogModule,
    EmptyStateComponent,
  ],
  providers: [MessageService, ConfirmationService],
  templateUrl: "./campanas.component.html",
  styleUrls: ["./campanas.component.css"],
})
export class CampanasComponent implements OnInit {
  private readonly service = inject(CampaignsService);
  private readonly toast = inject(MessageService);
  private readonly confirm = inject(ConfirmationService);
  /** id de la campaña que se está eliminando (evita doble clic). */
  eliminandoId: string | null = null;

  campaigns: Campana[] = [];
  cargando = true;
  error = "";

  busqueda = "";
  estado: EstadoCampana | "" = "";
  readonly estados = [
    { label: "Todas", value: "" },
    { label: "Borradores", value: "borrador" },
    { label: "Publicadas", value: "publicada" },
  ];

  dialogo = false;
  editandoId: string | null = null;
  guardando = false;
  form: CampanaInput = vacio();

  ngOnInit(): void {
    this.cargar();
  }

  cargar(): void {
    this.cargando = true;
    this.error = "";
    this.service.listar(this.busqueda, this.estado).subscribe({
      next: (c) => {
        this.campaigns = c;
        this.cargando = false;
      },
      error: (e: Error) => {
        this.error = e.message;
        this.campaigns = [];
        this.cargando = false;
      },
    });
  }

  nueva(): void {
    this.editandoId = null;
    this.form = vacio();
    this.dialogo = true;
  }

  editar(c: Campana): void {
    this.editandoId = c.id;
    this.form = {
      titulo: c.titulo,
      texto: c.texto,
      imagen_url: c.imagen_url,
      fecha: c.fecha,
    };
    this.dialogo = true;
  }

  /** Los cuatro datos que pide HU-47; el backend vuelve a validar todo. */
  get formCompleto(): boolean {
    const f = this.form;
    return !!(
      f.titulo.trim() &&
      f.texto.trim() &&
      f.imagen_url.trim() &&
      f.fecha
    );
  }

  guardar(): void {
    if (!this.formCompleto || this.guardando) return;
    this.guardando = true;
    const pedido = this.editandoId
      ? this.service.editar(this.editandoId, this.form)
      : this.service.crear(this.form);
    pedido.subscribe({
      next: (r) => {
        this.guardando = false;
        this.dialogo = false;
        this.toast.add({
          severity: "success",
          summary: "Listo",
          detail: r.mensaje,
        });
        this.cargar();
      },
      error: (e: Error) => {
        this.guardando = false;
        this.toast.add({
          severity: "error",
          summary: "No se pudo guardar",
          detail: e.message,
        });
      },
    });
  }

  publicar(c: Campana): void {
    this.service.publicar(c.id).subscribe({
      next: (r) => {
        this.toast.add({
          severity: "success",
          summary: "Publicada",
          detail: r.mensaje,
        });
        this.cargar();
      },
      error: (e: Error) =>
        this.toast.add({
          severity: "error",
          summary: "No se pudo publicar",
          detail: e.message,
        }),
    });
  }

  /**
   * Baja lógica con confirmación (Problema 5 de la auditoría): pide
   * confirmar, llama al backend y saca la campaña de la lista sin recargar
   * la página.
   */
  eliminar(c: Campana): void {
    if (this.eliminandoId) return;
    this.confirm.confirm({
      header: "Eliminar campaña",
      message: `¿Eliminar "${c.titulo}"? Deja de verse en la gestión y en la página pública. Los datos quedan guardados como baja.`,
      acceptLabel: "Eliminar",
      rejectLabel: "Cancelar",
      acceptButtonStyleClass: "p-button-danger",
      accept: () => {
        this.eliminandoId = c.id;
        this.service.eliminar(c.id).subscribe({
          next: (r) => {
            this.eliminandoId = null;
            this.campaigns = this.campaigns.filter((x) => x.id !== c.id);
            this.toast.add({
              severity: "success",
              summary: "Eliminada",
              detail: r.mensaje,
            });
          },
          error: (e: Error) => {
            this.eliminandoId = null;
            this.toast.add({
              severity: "error",
              summary: "No se pudo eliminar",
              detail: e.message,
            });
            this.cargar();
          },
        });
      },
    });
  }

  /** "2026-11-05" → "05/11/2026" (sin pasar por Date, para no correr el día). */
  fechaDMY(fecha: string): string {
    const [a, m, d] = fecha.split("-");
    return d && m && a ? `${d}/${m}/${a}` : fecha;
  }

  onImageError(event: Event): void {
    const el = event.target as HTMLImageElement;
    if (el && el.src !== DEFAULT_CAMPAIGN_PLACEHOLDER) {
      el.src = DEFAULT_CAMPAIGN_PLACEHOLDER;
    }
  }
}

function vacio(): CampanaInput {
  return { titulo: "", texto: "", imagen_url: "", fecha: "" };
}
