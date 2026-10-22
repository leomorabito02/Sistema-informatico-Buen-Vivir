import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { DialogModule } from "primeng/dialog";
import { SelectModule } from "primeng/select";
import { TextareaModule } from "primeng/textarea";
import { ToastModule } from "primeng/toast";
import { MessageService } from "primeng/api";
import { catchError, of } from "rxjs";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import { UsersService } from "../../../services/users.service";
import {
  ContenidoDocente,
  ContenidosService,
  DescargaBiblioteca,
  EstadoModeracion,
  guardarArchivo,
  tamanoLegible,
} from "../../../services/contenidos.service";

/**
 * Moderación del material de docentes (HU-23) y registro de quién
 * descargó de la biblioteca pública (HU-25). Solo administración.
 */
@Component({
  selector: "app-contenido-docente",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ButtonModule,
    TagModule,
    DialogModule,
    SelectModule,
    TextareaModule,
    ToastModule,
    EmptyStateComponent,
  ],
  providers: [MessageService],
  templateUrl: "./contenido-docente.component.html",
  styleUrls: ["../mis-contenidos/mis-contenidos.component.css", "./contenido-docente.component.css"],
})
export class ContenidoDocenteComponent implements OnInit {
  private readonly service = inject(ContenidosService);
  private readonly users = inject(UsersService);
  private readonly toast = inject(MessageService);

  readonly tamano = tamanoLegible;
  readonly estados = [
    { label: "Pendientes", value: "pendiente" },
    { label: "Aprobados", value: "aprobado" },
    { label: "Rechazados", value: "rechazado" },
    { label: "Todos", value: "" },
  ];

  estado: EstadoModeracion | "" = "pendiente";
  contenidos: ContenidoDocente[] = [];
  cargando = true;
  error = "";

  /** user_id del docente -> "Nombre Apellido". */
  docentes = new Map<string, string>();
  /** id del contenido -> título, para la tabla de descargas. */
  titulos = new Map<string, string>();

  descargas: DescargaBiblioteca[] = [];
  errorDescargas = "";

  rechazando: ContenidoDocente | null = null;
  motivo = "";
  enviando = false;

  ngOnInit(): void {
    this.users
      .listByRole("role_teacher")
      .pipe(catchError(() => of([])))
      .subscribe((lista) => {
        for (const u of lista) this.docentes.set(u.user_id, `${u.first_name} ${u.last_name}`.trim());
      });
    this.cargar();
    this.cargarDescargas();
  }

  cargar(): void {
    this.cargando = true;
    this.error = "";
    this.service.listar(this.estado).subscribe({
      next: (lista) => {
        this.contenidos = lista ?? [];
        for (const c of this.contenidos) this.titulos.set(c.id, c.titulo);
        this.cargando = false;
      },
      error: (e: Error) => {
        this.error = e.message;
        this.cargando = false;
      },
    });
  }

  cargarDescargas(): void {
    // Todos los contenidos, para poner el título en cada descarga.
    this.service
      .listar("")
      .pipe(catchError(() => of([] as ContenidoDocente[])))
      .subscribe((todos) => todos.forEach((c) => this.titulos.set(c.id, c.titulo)));
    this.service.descargas().subscribe({
      next: (lista) => (this.descargas = lista ?? []),
      error: (e: Error) => (this.errorDescargas = e.message),
    });
  }

  docente(id: string): string {
    return this.docentes.get(id) || "Docente";
  }

  etiqueta(e: EstadoModeracion): string {
    return { pendiente: "Pendiente", aprobado: "Aprobado", rechazado: "Rechazado" }[e];
  }

  severidad(e: EstadoModeracion): "warn" | "success" | "danger" {
    return ({ pendiente: "warn", aprobado: "success", rechazado: "danger" } as const)[e];
  }

  fecha(iso: string): string {
    const d = new Date(iso);
    return isNaN(d.getTime()) ? "" : d.toLocaleDateString("es-AR", { day: "2-digit", month: "2-digit", year: "numeric" });
  }

  descargar(c: ContenidoDocente): void {
    this.service.archivo(c.id).subscribe({
      next: (blob) => guardarArchivo(blob, c.archivoNombre),
      error: (e: Error) => this.toast.add({ severity: "error", summary: "No se pudo descargar", detail: e.message }),
    });
  }

  aprobar(c: ContenidoDocente): void {
    this.service.moderar(c.id, "aprobado").subscribe({
      next: () => {
        this.toast.add({ severity: "success", summary: "Material aprobado", detail: "Ya se ve en la biblioteca pública." });
        this.cargar();
      },
      error: (e: Error) => this.toast.add({ severity: "error", summary: "No se pudo aprobar", detail: e.message }),
    });
  }

  abrirRechazo(c: ContenidoDocente): void {
    this.rechazando = c;
    this.motivo = "";
  }

  rechazar(): void {
    const c = this.rechazando;
    if (!c) return;
    this.enviando = true;
    this.service.moderar(c.id, "rechazado", this.motivo).subscribe({
      next: () => {
        this.enviando = false;
        this.rechazando = null;
        this.toast.add({ severity: "success", summary: "Material rechazado", detail: "El docente ve el motivo en «Mis contenidos»." });
        this.cargar();
      },
      error: (e: Error) => {
        this.enviando = false;
        this.toast.add({ severity: "error", summary: "No se pudo rechazar", detail: e.message });
      },
    });
  }
}
