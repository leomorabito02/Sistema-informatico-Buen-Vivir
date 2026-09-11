import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { InputTextModule } from "primeng/inputtext";
import { CardModule } from "primeng/card";
import { AuthService } from "../../../services/auth.service";
import { UsersService } from "../../../services/users.service";
import { User } from "../../../core/models/auth.model";

export interface DashboardUser {
  id?: string;
  name: string;
  email: string;
  phone: string;
  role: string;
  status: string;
  coursesStatus?: string;
}

@Component({
  selector: "app-usuarios",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    TableModule,
    ButtonModule,
    TagModule,
    InputTextModule,
    CardModule,
  ],
  templateUrl: "./usuarios.component.html",
  styleUrls: ["./usuarios.component.css"],
})
export class UsuariosComponent implements OnInit {
  private readonly authService = inject(AuthService);
  private readonly usersService = inject(UsersService);

  searchTerm: string = "";
  selectedRole: string = "Todos los roles";
  currentUserRoleId: string = "";
  isVolunteerOnly: boolean = false;

  readonly rolesList: string[] = [
    "Todos los roles",
    "Participante",
    "Voluntario",
    "Docente",
    "Responsable",
    "Empresa",
  ];

  readonly initialMockUsers: DashboardUser[] = [
    {
      name: "María González",
      email: "maria@gmail.com",
      phone: "+54 9 11 1234-5678",
      role: "Participante",
      status: "Activo",
      coursesStatus: "No implementado aún",
    },
    {
      name: "Juan Pérez",
      email: "juan@gmail.com",
      phone: "+54 9 11 2345-6789",
      role: "Participante",
      status: "Activo",
      coursesStatus: "No implementado aún",
    },
    {
      name: "Ana Martinez",
      email: "ana@gmail.com",
      phone: "+54 9 11 3456-7890",
      role: "Voluntario",
      status: "Pendiente",
      coursesStatus: "No implementado aún",
    },
    {
      name: "Carlos López",
      email: "carlos@gmail.com",
      phone: "+54 9 11 4567-8901",
      role: "Docente",
      status: "Activo",
      coursesStatus: "No implementado aún",
    },
    {
      name: "Laura Rodríguez",
      email: "laura@gmail.com",
      phone: "+54 9 11 5678-9012",
      role: "Participante",
      status: "Activo",
      coursesStatus: "No implementado aún",
    },
  ];

  users: DashboardUser[] = [];
  filteredUsers: DashboardUser[] = [];

  ngOnInit() {
    this.checkCurrentUserPermissions();
  }

  private checkCurrentUserPermissions() {
    this.authService.isAuthenticated().subscribe({
      next: (me) => {
        const roleId = (me.role_id ?? me.user?.role_id ?? "").toLowerCase();
        this.currentUserRoleId = roleId;
        this.isVolunteerOnly =
          roleId === "role_volunteer" ||
          roleId === "voluntario" ||
          roleId === "volunteer";

        this.loadUsers();
      },
      error: () => {
        // Fallback for standalone/demo execution
        this.loadUsers();
      },
    });
  }

  private loadUsers() {
    this.usersService.listUsers().subscribe({
      next: (data: User[]) => {
        if (data && data.length > 0) {
          this.users = data.map((u) => this.mapUserToDashboardUser(u));
        } else {
          this.users = [...this.initialMockUsers];
        }
        this.filterUsers();
      },
      error: () => {
        this.users = [...this.initialMockUsers];
        this.filterUsers();
      },
    });
  }

  private mapUserToDashboardUser(u: User): DashboardUser {
    const roleName = this.formatRoleName(u.role?.name ?? u.role_id);
    const fullName =
      `${u.first_name || ""} ${u.last_name || ""}`.trim() || u.email;
    const isPending =
      u.status?.toLowerCase() === "pending" ||
      u.status?.toLowerCase() === "pending_approval" ||
      u.status?.toLowerCase() === "pendiente";

    return {
      id: u.user_id,
      name: fullName,
      email: u.email || "-",
      phone: u.phone || "-",
      role: roleName,
      status: isPending ? "Pendiente" : "Activo",
      coursesStatus: "No implementado aún",
    };
  }

  private formatRoleName(roleRaw: string): string {
    if (!roleRaw) return "Participante";
    const r = roleRaw.toLowerCase();
    if (r.includes("admin")) return "Responsable";
    if (r.includes("teacher") || r.includes("docente")) return "Docente";
    if (r.includes("volunteer") || r.includes("voluntario"))
      return "Voluntario";
    if (r.includes("company") || r.includes("empresa")) return "Empresa";
    return "Participante";
  }

  filterUsers() {
    const term = this.searchTerm.toLowerCase().trim();
    const roleFilter = this.selectedRole;

    this.filteredUsers = this.users.filter((u) => {
      // 1. Volunteer role restriction enforcement
      if (this.isVolunteerOnly && u.role !== "Participante") {
        return false;
      }

      // 2. Dropdown role filter
      if (roleFilter !== "Todos los roles" && u.role !== roleFilter) {
        return false;
      }

      // 3. Search query matching
      if (!term) return true;
      return (
        u.name.toLowerCase().includes(term) ||
        u.email.toLowerCase().includes(term) ||
        u.phone.toLowerCase().includes(term) ||
        u.role.toLowerCase().includes(term)
      );
    });
  }
}
