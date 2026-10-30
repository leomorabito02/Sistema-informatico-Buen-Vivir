import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { RouterLink } from "@angular/router";
import {
  LucideAngularModule,
  UtensilsCrossed,
  Hammer,
  Sprout,
  Laptop,
  Scissors,
  HeartPulse,
  Briefcase,
  Mic,
  FileText,
} from "lucide-angular";
import {
  ContenidoPublico,
  ContenidosService,
  DatosDescarga,
  guardarArchivo,
  tamanoLegible,
} from "../../../services/contenidos.service";

@Component({
  selector: "app-library",
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, LucideAngularModule],
  templateUrl: "./library.component.html",
  styleUrl: "./library.component.css",
})
/**
 * Biblioteca de la portada: categorías y los recursos que aprobó la
 * administración (HU-23). Para descargar se piden nombre, apellido, email
 * y teléfono (HU-25); el backend los registra y devuelve el archivo.
 */
export class LibraryComponent implements OnInit {
  private readonly service = inject(ContenidosService);

  readonly FileTextIcon = FileText;
  readonly tamano = tamanoLegible;

  recursos: ContenidoPublico[] = [];
  cargando = true;
  error = "";

  elegido: ContenidoPublico | null = null;
  datos: DatosDescarga = { nombre: "", apellido: "", email: "", telefono: "" };
  errorForm = "";
  enviando = false;

  ngOnInit(): void {
    this.service.biblioteca().subscribe({
      next: (lista) => {
        this.recursos = lista ?? [];
        this.cargando = false;
      },
      error: () => {
        this.error = "No se pudieron cargar los recursos. Probá más tarde.";
        this.cargando = false;
      },
    });
  }

  /** PDF, Word, Imagen… según el tipo del archivo. */
  formato(r: ContenidoPublico): string {
    const t = r.archivoTipo;
    if (t === "application/pdf") return "PDF";
    if (t.startsWith("image/")) return "Imagen";
    if (t.includes("word")) return "Word";
    if (t.includes("sheet") || t.includes("excel")) return "Excel";
    if (t.includes("presentation") || t.includes("powerpoint"))
      return "Presentación";
    return "Documento";
  }

  abrir(r: ContenidoPublico): void {
    this.elegido = r;
    this.errorForm = "";
  }

  cerrar(): void {
    this.elegido = null;
  }

  descargar(): void {
    const r = this.elegido;
    if (!r) return;
    const d = this.datos;
    if (!d.nombre.trim() || !d.apellido.trim()) {
      this.errorForm = "Completá tu nombre y apellido";
      return;
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(d.email.trim())) {
      this.errorForm = "El email no es válido";
      return;
    }
    const digitos = d.telefono.replace(/[\s.()-]/g, "");
    if (!/^\+?[0-9]{8,15}$/.test(digitos)) {
      this.errorForm =
        "El teléfono tiene que tener entre 8 y 15 números, por ejemplo 3547 123456";
      return;
    }
    this.errorForm = "";
    this.enviando = true;
    this.service.descargarPublico(r.id, d).subscribe({
      next: (blob) => {
        this.enviando = false;
        guardarArchivo(blob, r.archivoNombre);
        this.elegido = null;
      },
      error: (e: Error) => {
        this.enviando = false;
        this.errorForm = e.message;
      },
    });
  }

  readonly UtensilsCrossedIcon = UtensilsCrossed;
  readonly HammerIcon = Hammer;
  readonly SproutIcon = Sprout;
  readonly LaptopIcon = Laptop;
  readonly ScissorsIcon = Scissors;
  readonly HeartPulseIcon = HeartPulse;
  readonly BriefcaseIcon = Briefcase;
  readonly MicIcon = Mic;
}
