import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { AuthService } from "../../../services/auth.service";
import { MeResponse } from "../../../core/models/auth.model";

/**
 * Vista inicial de /dashboard (ruta hija vacía). Muestra un saludo con
 * el nombre del usuario logueado para Responsable, Voluntario y
 * Participante. Docente y Empresa no ven nada acá (mismo comportamiento
 * que antes): su navegación sigue siendo solo desde el header.
 */
@Component({
  selector: "app-inicio",
  standalone: true,
  imports: [CommonModule],
  templateUrl: "./inicio.component.html",
  styleUrl: "./inicio.component.css",
})
export class InicioComponent implements OnInit {
  private authService = inject(AuthService);

  private static readonly ROLE_LABELS: Record<string, string> = {
    role_admin: "Responsable",
    role_volunteer: "Voluntario",
    role_participant: "Participante",
  };

  firstName = "";
  roleLabel = "";

  ngOnInit(): void {
    this.authService.isAuthenticated().subscribe({
      next: (res: MeResponse) => {
        const roleId = res.role_id ?? res.user?.role_id ?? "";
        this.roleLabel = InicioComponent.ROLE_LABELS[roleId] ?? "";
        this.firstName = res.user?.first_name?.trim() ?? "";
      },
      error: () => {
        this.roleLabel = "";
      },
    });
  }
}
