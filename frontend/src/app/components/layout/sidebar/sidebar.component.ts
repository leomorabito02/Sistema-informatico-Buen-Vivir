import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterLink, RouterLinkActive } from "@angular/router";
import { AuthService } from "../../../services/auth.service";
import { MeResponse } from "../../../core/models/auth.model";

interface NavItem {
  label: string;
  icon: string;
  route: string;
  roles: string[]; // vacío = todos los autenticados
}

@Component({
  selector: "app-sidebar",
  standalone: true,
  imports: [CommonModule, RouterLink, RouterLinkActive],
  templateUrl: "./sidebar.component.html",
  styleUrl: "./sidebar.component.css",
})
export class SidebarComponent implements OnInit {
  private authService = inject(AuthService);

  roleId = "";

  private readonly allItems: NavItem[] = [
    { label: "Catálogo de cursos", icon: "pi-book", route: "/dashboard/cursos", roles: [] },
    { label: "Mi perfil", icon: "pi-user", route: "/dashboard/mi-perfil", roles: [] },
    // Administración
    { label: "Gestión de cursos", icon: "pi-cog", route: "/dashboard/gestion-cursos", roles: ["role_admin", "role_dev"] },
    { label: "Usuarios", icon: "pi-users", route: "/dashboard/usuarios", roles: ["role_admin", "role_dev"] },
    // Docente responsable
    { label: "Mis cursos", icon: "pi-briefcase", route: "/dashboard/mis-cursos", roles: ["role_teacher"] },
    { label: "Asistencia", icon: "pi-calendar-plus", route: "/dashboard/asistencia", roles: ["role_teacher"] },
    { label: "Progreso", icon: "pi-chart-line", route: "/dashboard/progreso", roles: ["role_teacher"] },
    { label: "Participantes", icon: "pi-id-card", route: "/dashboard/participantes", roles: ["role_teacher"] },
    { label: "Seguimiento", icon: "pi-eye", route: "/dashboard/seguimiento", roles: ["role_teacher", "role_volunteer"] },
  ];

  items: NavItem[] = [];

  ngOnInit(): void {
    this.authService.isAuthenticated().subscribe({
      next: (res: MeResponse) => {
        this.roleId = res.role_id ?? res.user?.role_id ?? "";
        this.items = this.allItems.filter(
          (it) => it.roles.length === 0 || it.roles.includes(this.roleId),
        );
      },
      error: () => {
        this.items = this.allItems.filter((it) => it.roles.length === 0);
      },
    });
  }
}
