import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { InputTextModule } from "primeng/inputtext";
import { CardModule } from "primeng/card";
import { DialogModule } from "primeng/dialog";
import { AuthService } from "../../../services/auth.service";
import { UsersService } from "../../../services/users.service";
import { EducationService } from "../../../services/education.service";
import { User } from "../../../core/models/auth.model";

export interface DashboardUser {
  id: string;
  name: string;
  email: string;
  phone: string;
  role: string;
  status: string; // 'Activo' | 'Pendiente' | 'Bloqueado'
  roleApprovalStatus?: string; // 'approved' | 'pending' | 'rejected'
  coursesStatus: string;
  originalUser?: any; // To store full backend data if needed
  profileData?: any; // Full profile data fetched on demand
  coursesList?: any[]; // Enrolled courses or taught courses
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
    DialogModule,
  ],
  templateUrl: "./usuarios.component.html",
  styleUrls: ["./usuarios.component.css"],
})
export class UsuariosComponent implements OnInit {
  private readonly authService = inject(AuthService);
  private readonly usersService = inject(UsersService);
  private readonly educationService = inject(EducationService);

  searchTerm: string = "";
  selectedRole: string = "Todos los roles";
  currentUserRoleId: string = "";
  currentUserId: string = "";
  isVolunteerOnly: boolean = false;

  readonly rolesList: string[] = [
    "Todos los roles",
    "Participante",
    "Voluntario",
    "Docente",
    "Responsable",
    "Empresa",
  ];

  showNewUserModal: boolean = false;
  newUserForm: any = {
    first_name: "",
    last_name: "",
    email: "",
    phone: "",
    password: "",
    role_id: "",
  };
  availableNewRoles: string[] = [];

  users: DashboardUser[] = [];
  filteredUsers: DashboardUser[] = [];

  showDetailModal: boolean = false;
  showEditModal: boolean = false;
  selectedUser: DashboardUser | null = null;
  editForm: any = {};

  ngOnInit() {
    this.checkCurrentUserPermissions();
  }

  isRolePending(status?: string): boolean {
    if (!status) return false;
    const s = status.toLowerCase();
    return s === "pending" || s === "pendiente";
  }

  private checkCurrentUserPermissions() {
    this.authService.isAuthenticated().subscribe({
      next: (me) => {
        this.currentUserId = me.user?.user_id || (me as any).user_id || "";
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
          this.loadUserCoursesBatch();
        } else {
          this.users = [];
          this.filterUsers();
        }
      },
      error: (err) => {
        console.error("Error loading users:", err);
        this.users = [];
        this.filterUsers();
      },
    });
  }

  private loadUserCoursesBatch() {
    const userIds = this.users.map((u) => u.id);
    if (userIds.length === 0) return;

    this.educationService.getBulkUserCourses(userIds).subscribe({
      next: (data) => {
        const courseMap = new Map<string, string[]>();
        data.forEach((item) => courseMap.set(item.user_id, item.courses));

        this.users = this.users.map((u) => {
          const courses = courseMap.get(u.id) || [];
          return {
            ...u,
            coursesStatus:
              courses.length > 0 ? courses.join(", ") : "Sin cursos",
          };
        });
        this.filterUsers();
      },
      error: (err) => {
        console.error("Error cargando cursos batch:", err);
        this.filterUsers();
      },
    });
  }

  private mapUserToDashboardUser(u: User): DashboardUser {
    const roleName = this.formatRoleName(u.role?.name ?? u.role_id);
    const fullName =
      `${u.first_name || ""} ${u.last_name || ""}`.trim() || u.email;

    // Real status mapped to Spanish for UI
    let displayStatus = "Activo";
    if (u.status?.toLowerCase() === "blocked") {
      displayStatus = "Bloqueado";
    }

    return {
      id: u.user_id,
      name: fullName,
      email: u.email || "-",
      phone: u.phone || "-",
      role: roleName,
      status: displayStatus,
      roleApprovalStatus: (u as any).role_approval_status,
      coursesStatus: "No implementado aún",
      originalUser: u,
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

  openNewUserModal() {
    this.newUserForm = {
      first_name: "",
      last_name: "",
      email: "",
      phone: "",
      password: "",
      role_id: "",
    };

    if (this.isVolunteerOnly) {
      this.availableNewRoles = [
        "role_participant",
        "role_company",
        "role_teacher",
        "role_volunteer",
      ];
    } else {
      this.availableNewRoles = [
        "role_participant",
        "role_company",
        "role_teacher",
        "role_volunteer",
        "role_admin",
      ];
    }

    this.showNewUserModal = true;
  }

  createUser() {
    if (
      !this.newUserForm.email ||
      !this.newUserForm.password ||
      !this.newUserForm.role_id
    ) {
      return;
    }

    const payload = {
      first_name: this.newUserForm.first_name,
      last_name: this.newUserForm.last_name,
      email: this.newUserForm.email,
      phone: this.newUserForm.phone,
      password: this.newUserForm.password,
      role_id: this.newUserForm.role_id,
    };

    this.usersService.createUser(payload).subscribe({
      next: () => {
        alert("Usuario creado correctamente");
        this.showNewUserModal = false;
        this.loadUsers();
      },
      error: (err: any) => {
        console.error("Error creating user:", err);
        const errorMsg =
          err.message ||
          err.error?.message ||
          err.error?.error ||
          "Hubo un error al crear el usuario. Por favor verifica los datos e intentalo de nuevo.";
        alert(errorMsg);
      },
    });
  }

  openUserDetail(user: DashboardUser) {
    this.selectedUser = user;
    this.fetchUserDetails(user, false, () => {
      this.showDetailModal = true;
    });
  }

  openEditUser(user: DashboardUser, event: Event) {
    event.stopPropagation();

    // Verificación de seguridad: Voluntarios solo pueden editar participantes
    if (
      this.isVolunteerOnly &&
      user.originalUser?.role_id !== "role_participant"
    ) {
      // Si el rol es distinto a role_participant, bloqueamos.
      alert("Como voluntario solo tienes permisos para editar participantes.");
      return;
    }

    this.selectedUser = user;

    // Inicializar el formulario con los datos básicos
    const userStatusRaw = user.originalUser?.status || "active";
    this.editForm = {
      first_name: user.originalUser?.first_name || "",
      last_name: user.originalUser?.last_name || "",
      phone: user.originalUser?.phone || "",
      status: userStatusRaw,
    };

    this.fetchUserDetails(user, true, () => {
      this.showEditModal = true;
    });
  }

  private fetchUserDetails(
    user: DashboardUser,
    populateEditForm: boolean = false,
    onComplete?: () => void,
  ) {
    if (!user.originalUser?.user_id) {
      if (onComplete) onComplete();
      return;
    }
    const userId = user.originalUser.user_id;

    this.usersService.getUserProfile(userId).subscribe({
      next: (profile) => {
        user.profileData = profile;
        if (populateEditForm && profile.role_data) {
          // Fusionar los datos del rol en el editForm para los inputs correspondientes
          this.editForm = {
            ...this.editForm,
            ...profile.role_data,
          };

          // "accepted" es el valor legado de aprobación (hay perfiles
          // guardados así); el select solo ofrece "approved".
          if (this.editForm.approval_status === "accepted") {
            this.editForm.approval_status = "approved";
          }

          // Formatear fechas si existen para input type="date"
          if (this.editForm.birthdate) {
            this.editForm.birthdate = new Date(this.editForm.birthdate)
              .toISOString()
              .split("T")[0];
          }
        }
        if (onComplete) onComplete();
      },
      error: (err) => {
        console.error("Error fetching user profile:", err);
        if (onComplete) onComplete();
      },
    });

    if (
      user.originalUser.role_id === "role_participant" ||
      user.originalUser.role_id === "role_volunteer"
    ) {
      this.educationService.getParticipantEnrollments(userId).subscribe({
        next: (enrollments) => {
          user.coursesList = enrollments || [];
        },
        error: (err) => console.error("Error fetching enrollments:", err),
      });
    } else if (user.originalUser.role_id === "role_teacher") {
      this.educationService.getTeacherCourses(userId).subscribe({
        next: (courses) => {
          user.coursesList = courses || [];
        },
        error: (err) => console.error("Error fetching teacher courses:", err),
      });
    }
  }

  deleteUser(user: DashboardUser, event: Event) {
    event.stopPropagation();
    // Lógica para eliminar usuario
  }

  saveEdit() {
    if (!this.selectedUser?.originalUser?.user_id) return;

    const userId = this.selectedUser.originalUser.user_id;
    const roleId = this.selectedUser.originalUser.role_id;

    // 1. Actualizar usuario general
    const userPayload = {
      first_name: this.editForm.first_name,
      last_name: this.editForm.last_name,
      phone: this.editForm.phone,
      status: this.editForm.status,
    };

    this.usersService.updateUser(userId, userPayload).subscribe({
      next: () => {
        // 2. Si hay rol específico, actualizar la información del rol
        if (roleId && roleId !== "role_admin" && roleId !== "admin") {
          this.updateRoleSpecificData(userId, roleId);
        } else {
          this.finishSave();
        }
      },
      error: (err) => {
        console.error("Error updating user:", err);
        alert("Error al actualizar usuario base: " + err.message);
      },
    });
  }

  private getRoleTableName(roleId: string): string {
    const tableMap: Record<string, string> = {
      role_participant: "participants",
      participant: "participants",
      role_company: "companies",
      company: "companies",
      role_teacher: "teachers",
      teacher: "teachers",
      role_volunteer: "volunteers",
      volunteer: "volunteers",
    };
    return tableMap[roleId] || "";
  }

  private updateRoleSpecificData(userId: string, roleId: string) {
    const tableName = this.getRoleTableName(roleId);
    if (!tableName) {
      this.finishSave();
      return;
    }

    const idKeyMap: Record<string, string> = {
      participants: "participant_id",
      companies: "company_id",
      teachers: "teacher_id",
      volunteers: "volunteer_id",
    };
    const key = idKeyMap[tableName];

    // Si no existe la ID del rol (ej. volunteer_id) en el formulario, significa que
    // el registro del rol no existe en la base de datos para este usuario.
    // No podemos hacer un PUT.
    if (key && !this.editForm[key]) {
      this.finishSave();
      return;
    }

    const targetRoleId = key ? this.editForm[key] : userId;

    // Creamos un payload limpio con las propiedades del role
    const rolePayload = { ...this.editForm };
    delete rolePayload.first_name;
    delete rolePayload.last_name;
    delete rolePayload.phone;
    delete rolePayload.status;

    if (rolePayload.approval_status === "approved" && this.currentUserId) {
      rolePayload.approved_by = this.currentUserId;
    }

    // Formatear fechas para enviarlas como RFC3339 / ISO
    if (rolePayload.birthdate) {
      rolePayload.birthdate = new Date(rolePayload.birthdate).toISOString();
    } else {
      rolePayload.birthdate = null;
    }

    // Enviar petición para actualizar el rol
    this.usersService
      .updateRoleData(tableName, targetRoleId, rolePayload)
      .subscribe({
        next: () => {
          this.finishSave();
        },
        error: (err) => {
          console.error("Error updating role data:", err);
          alert("Error al guardar la información del rol: " + err.message);
        },
      });
  }

  private finishSave() {
    this.showEditModal = false;
    this.loadUsers(); // Recargar la lista para reflejar los cambios
  }
}
