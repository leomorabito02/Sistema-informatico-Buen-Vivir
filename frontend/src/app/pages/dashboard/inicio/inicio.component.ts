import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { Router } from "@angular/router";
import { AuthService } from "../../../services/auth.service";
import { MeResponse } from "../../../core/models/auth.model";

/**
 * Vista inicial de /dashboard (ruta hija vacía).
 *
 * - Responsable, Voluntario y Participante: saludo con el nombre.
 * - Docente y Empresa entran directo a su portal (Mis cursos / Ofertas
 *   laborales): para ellos esta pantalla no tiene contenido propio, así
 *   que se reemplaza en el historial y "Atrás" no vuelve a una página vacía.
 *   El acceso a cada destino lo siguen controlando los guards.
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
  private router = inject(Router);

  private static readonly ROLE_LABELS: Record<string, string> = {
    role_admin: "Responsable",
    role_volunteer: "Voluntario",
    role_participant: "Participante",
  };

  static readonly PORTAL_POR_ROL: Record<string, string> = {
    role_teacher: "/dashboard/mis-cursos",
    role_company: "/dashboard/ofertas",
  };

  firstName = "";
  roleLabel = "";

  ngOnInit(): void {
    this.authService.isAuthenticated().subscribe({
      next: (res: MeResponse) => {
        const roleId = res.role_id ?? res.user?.role_id ?? "";
        const portal =
          roleId === "role_teacher" && res.is_external
            ? "/dashboard/mis-contenidos"
            : InicioComponent.PORTAL_POR_ROL[roleId];
        if (portal) {
          this.router.navigateByUrl(portal, { replaceUrl: true });
          return;
        }
        this.roleLabel = InicioComponent.ROLE_LABELS[roleId] ?? "";
        this.firstName = res.user?.first_name?.trim() ?? "";
      },
      error: () => {
        this.roleLabel = "";
      },
    });
  }
}
