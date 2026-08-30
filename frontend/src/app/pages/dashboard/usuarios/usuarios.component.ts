import { Component, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { InputTextModule } from "primeng/inputtext";
import { CardModule } from "primeng/card";

interface DashboardUser {
  name: string;
  email: string;
  role: string;
  status: string;
  createdAt?: string;
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
  searchTerm: string = "";

  users: DashboardUser[] = [];
  filteredUsers: DashboardUser[] = [];

  ngOnInit() {}

  filterUsers() {
    const term = this.searchTerm.toLowerCase().trim();
    this.filteredUsers = this.users.filter(
      (u) =>
        u.name.toLowerCase().includes(term) ||
        u.email.toLowerCase().includes(term) ||
        u.role.toLowerCase().includes(term),
    );
  }

  getRoleSeverity(role: string): string | undefined {
    switch (role) {
      case "Responsable":
        return "danger";
      case "Docente":
        return "warning";
      case "Empresa":
        return "contrast";
      case "Voluntario":
        return "info";
      case "Participante":
        return "success";
      default:
        return undefined;
    }
  }

  getStatusSeverity(status: string): string {
    return status === "Activo" ? "success" : "warning";
  }
}
