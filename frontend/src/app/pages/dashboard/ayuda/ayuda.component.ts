import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { LucideAngularModule, Search } from "lucide-angular";
import { AuthService } from "../../../services/auth.service";
import { MeResponse } from "../../../core/models/auth.model";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import {
  ATAJOS,
  AYUDA_GENERAL,
  AYUDA_POR_ROL,
  AyudaTema,
} from "./ayuda.content";

/**
 * Ayuda del panel: guías cortas de las tareas de cada rol, preguntas
 * generales y atajos de teclado, con un buscador.
 */
@Component({
  selector: "app-ayuda",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    LucideAngularModule,
    EmptyStateComponent,
  ],
  templateUrl: "./ayuda.component.html",
  styleUrl: "./ayuda.component.css",
})
export class AyudaComponent implements OnInit {
  private authService = inject(AuthService);

  readonly atajos = ATAJOS;
  readonly Search = Search;

  query = "";
  delRol: AyudaTema[] = [];
  readonly general = AYUDA_GENERAL;

  ngOnInit(): void {
    this.authService.isAuthenticated().subscribe({
      next: (res: MeResponse) => {
        const roleId = (res.role_id ?? res.user?.role_id ?? "").toLowerCase();
        this.delRol = AYUDA_POR_ROL[roleId] ?? [];
      },
      error: () => (this.delRol = []),
    });
  }

  filtrar(temas: AyudaTema[]): AyudaTema[] {
    const q = this.normalizar(this.query.trim());
    if (!q) return temas;
    return temas.filter((t) =>
      this.normalizar([t.pregunta, ...t.pasos].join(" ")).includes(q),
    );
  }

  get sinResultados(): boolean {
    return (
      this.filtrar(this.delRol).length === 0 &&
      this.filtrar(this.general).length === 0
    );
  }

  /** Minúsculas y sin tildes, para que "inscripcion" encuentre "inscripción". */
  private normalizar(s: string): string {
    return s
      .toLowerCase()
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "");
  }
}
