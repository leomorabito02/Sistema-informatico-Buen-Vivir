import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { InputTextModule } from "primeng/inputtext";
import { CardModule } from "primeng/card";
import { DialogModule } from "primeng/dialog";
import { ToastModule } from "primeng/toast";
import { ConfirmDialogModule } from "primeng/confirmdialog";
import { ConfirmationService, MessageService } from "primeng/api";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { SkeletonModule } from "primeng/skeleton";
import { AuthService } from "../../../services/auth.service";
import { UsersService } from "../../../services/users.service";
import { EducationService } from "../../../services/education.service";
import { User } from "../../../core/models/auth.model";
import { Router } from "@angular/router";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";
import { RxResumeService } from "../../../services/rxresume.service";
import { ImageService } from "../../../services/image.service";
import { validarFotoPerfil } from "../../../shared/utils/foto-perfil";

export interface DashboardUser {
  id: string;
  name: string;
  email: string;
  phone: string;
  role: string;
  status: string; // 'Activo' | 'Pendiente' | 'Bloqueado'
  roleApprovalStatus?: string; // 'approved' | 'pending' | 'rejected'
  image_url?: string;
  coursesStatus: string;
  originalUser?: any; // To store full backend data if needed
  profileData?: any; // Full profile data fetched on demand
  coursesList?: any[]; // Enrolled courses or taught courses
}

/** Valor del selector para el docente externo (role_teacher + is_external). */
export const ROLE_TEACHER_EXTERNAL = "role_teacher_external";

const NEW_ROLE_LABELS: Record<string, string> = {
  role_participant: "Participante",
  role_volunteer: "Voluntario",
  role_company: "Empresa",
  role_teacher: "Docente",
  [ROLE_TEACHER_EXTERNAL]: "Docente externo",
  role_admin: "Responsable (administración)",
};

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
    ToastModule,
    ConfirmDialogModule,
    ErrorStateComponent,
    ProgressSpinnerModule,
    SkeletonModule,
  ],
  providers: [MessageService, ConfirmationService],
  templateUrl: "./usuarios.component.html",
  styleUrls: ["./usuarios.component.css"],
})
export class UsuariosComponent implements OnInit {
  readonly ROLE_TEACHER_EXTERNAL = ROLE_TEACHER_EXTERNAL;

  private readonly authService = inject(AuthService);
  private readonly usersService = inject(UsersService);
  private readonly educationService = inject(EducationService);
  private readonly router = inject(Router);
  private readonly messages = inject(MessageService);
  private readonly confirmation = inject(ConfirmationService);

  /** Filas de muestra para el skeleton de la tabla. */
  readonly skeletonRows = [0, 1, 2, 3, 4, 5];

  /** Carga inicial de la lista (para skeleton y estado de error). */
  loading = true;
  loadError = "";
  private readonly rxResumeService = inject(RxResumeService);
  private readonly imageService = inject(ImageService);

  searchTerm: string = "";
  /** Mostrar solo las cuentas pendientes de aprobación (HU-12, HU-38). */
  soloPendientes = false;
  private busquedaTimer?: ReturnType<typeof setTimeout>;
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
    dni: "",
    address: "",
  };
  availableNewRoles: string[] = [];
  /** approval_status con el que se abrió la edición (para saber si cambió). */
  originalApprovalStatus = "";

  users: DashboardUser[] = [];
  filteredUsers: DashboardUser[] = [];

  showDetailModal: boolean = false;
  showEditModal: boolean = false;
  selectedUser: DashboardUser | null = null;
  editForm: any = {};
  participantDocuments: any[] = [];
  uploadingDocument: boolean = false;
  editingDocId: string | null = null;
  editingDocName: string = "";
  editingDocFile: File | null = null;

  ngOnInit() {
    this.checkCurrentUserPermissions();
  }

  /** "approved" o "rejected" si la edición decide sobre una cuenta pendiente. */
  approvalDecision(): "approved" | "rejected" | null {
    const nuevo = this.editForm?.approval_status;
    if (!this.isRolePending(this.originalApprovalStatus)) return null;
    return nuevo === "approved" || nuevo === "rejected" ? nuevo : null;
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

  /** Reintento desde el estado de error. */
  reload() {
    this.loadUsers();
  }

  private loadUsers() {
    this.loading = true;
    this.loadError = "";
    this.usersService.listUsers({ q: this.searchTerm }).subscribe({
      next: (data: User[]) => {
        if (data && data.length > 0) {
          this.users = data.map((u) => this.mapUserToDashboardUser(u));
          this.loadUserCoursesBatch();
        } else {
          this.loading = false;
          this.users = [];
          this.filterUsers();
        }
      },
      error: (err) => {
        console.error("Error loading users:", err);
        this.loading = false;
        this.loadError =
          err?.message || "No se pudo obtener el listado de usuarios.";
        this.users = [];
        this.filterUsers();
      },
    });
  }

  private loadUserCoursesBatch() {
    const userIds = this.users.map((u) => u.id);
    if (userIds.length === 0) {
      this.loading = false;
      return;
    }

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
        this.loading = false;
        this.filterUsers();
      },
      error: (err) => {
        console.error("Error cargando cursos batch:", err);
        this.loading = false;
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
      image_url: u.image_url,
      coursesStatus: "—",
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

  /**
   * El texto lo busca auth-service (`?q=`: nombre, apellido, email o
   * teléfono). Acá solo se espera a que se deje de tipear para no pedir
   * la lista en cada tecla.
   */
  buscar() {
    clearTimeout(this.busquedaTimer);
    this.busquedaTimer = setTimeout(() => this.loadUsers(), 300);
  }

  /** Filtros sobre la lista que ya mandó auth-service: rol elegido y restricción del voluntario. */
  filterUsers() {
    const roleFilter = this.selectedRole;

    this.filteredUsers = this.users.filter((u) => {
      // 1. Volunteer role restriction enforcement
      if (this.isVolunteerOnly && u.role !== "Participante") {
        return false;
      }

      // 2. Solo pendientes de aprobación
      if (this.soloPendientes && !this.isRolePending(u.roleApprovalStatus)) {
        return false;
      }

      // 3. Dropdown role filter
      return roleFilter === "Todos los roles" || u.role === roleFilter;
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
      dni: "",
      address: "",
    };

    if (this.isVolunteerOnly) {
      // auth-service solo deja al voluntario crear participantes (HU-07):
      // ofrecer otros roles terminaba en un error al guardar.
      this.availableNewRoles = ["role_participant"];
    } else {
      this.availableNewRoles = [
        "role_participant",
        "role_company",
        "role_teacher",
        ROLE_TEACHER_EXTERNAL,
        "role_volunteer",
        "role_admin",
      ];
    }

    this.showNewUserModal = true;
  }

  /** Nombre visible del rol en el alta (antes se veía "role_participant"). */
  newRoleLabel(roleId: string): string {
    return NEW_ROLE_LABELS[roleId] ?? this.formatRoleName(roleId);
  }

  createUser() {
    if (
      !this.newUserForm.email ||
      !this.newUserForm.password ||
      !this.newUserForm.role_id
    ) {
      this.messages.add({
        severity: "warn",
        summary: "Faltan datos",
        detail: "Completar todos los campos obligatorios para continuar",
      });
      return;
    }

    const participantError = validateNewParticipant(this.newUserForm);
    if (participantError) {
      this.messages.add({
        severity: "warn",
        summary: "Faltan datos del participante",
        detail: participantError,
      });
      return;
    }

    // "Docente externo" no es un rol aparte: es role_teacher con
    // is_external (columna que ya existe en teachers).
    const esExterno = this.newUserForm.role_id === ROLE_TEACHER_EXTERNAL;
    const payload: Record<string, string | boolean> = {
      first_name: this.newUserForm.first_name,
      last_name: this.newUserForm.last_name,
      email: this.newUserForm.email,
      phone: this.newUserForm.phone,
      password: this.newUserForm.password,
      role_id: esExterno ? "role_teacher" : this.newUserForm.role_id,
    };
    if (esExterno) payload["is_external"] = true;
    if (this.newUserForm.role_id === "role_participant") {
      payload["dni"] = String(this.newUserForm.dni ?? "").trim();
      payload["address"] = String(this.newUserForm.address ?? "").trim();
    }

    this.usersService.createUser(payload).subscribe({
      next: () => {
        this.messages.add({
          severity: "success",
          summary: "Listo",
          detail: "Usuario creado correctamente",
        });
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
        this.messages.add({
          severity: "error",
          summary: "No se pudo crear el usuario",
          detail: errorMsg,
        });
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
      this.messages.add({
        severity: "info",
        summary: "Sin permiso",
        detail: "Como voluntario solo podés editar participantes.",
      });
      return;
    }

    this.showDetailModal = false;
    this.selectedUser = user;
    this.cancelEditDocument();

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
    const isParticipant =
      user.originalUser.role_id === "role_participant" ||
      user.role === "Participante";

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
          this.originalApprovalStatus = this.editForm.approval_status ?? "";

          // Formatear fechas si existen para input type="date"
          if (this.editForm.birthdate) {
            this.editForm.birthdate = new Date(this.editForm.birthdate)
              .toISOString()
              .split("T")[0];
          }
        }

        const participantId = profile.role_data?.participant_id;
        if (participantId) {
          this.usersService.getParticipantDocuments(participantId).subscribe({
            next: (docs) => (this.participantDocuments = docs || []),
            error: (err) => console.error("Error fetching docs", err),
          });
        } else {
          this.participantDocuments = [];
        }

        if (isParticipant) {
          this.loadParticipantEnrollments(user, participantId || userId);
          this.rxResumeService.getUserResumeByAdmin(userId).subscribe({
            next: (resume) => {
              const cvPic = resume?.data?.basics?.picture;
              if (cvPic) {
                user.image_url = cvPic;
                if (user.profileData?.role_data) {
                  user.profileData.role_data.image_url = cvPic;
                }
                if (populateEditForm) {
                  this.editForm.image_url = cvPic;
                }
              }
            },
            error: () => {},
          });
        }

        if (onComplete) onComplete();
      },
      error: (err) => {
        console.error("Error fetching user profile:", err);
        if (isParticipant) {
          this.loadParticipantEnrollments(user, userId);
        }
        if (onComplete) onComplete();
      },
    });

    if (user.originalUser.role_id === "role_teacher") {
      this.educationService.getTeacherCourses(userId).subscribe({
        next: (courses) => {
          user.coursesList = courses || [];
        },
        error: (err) => console.error("Error fetching teacher courses:", err),
      });
    }
  }

  private loadParticipantEnrollments(
    user: DashboardUser,
    participantOrUserId: string,
  ) {
    this.educationService
      .getParticipantEnrollments(participantOrUserId)
      .subscribe({
        next: (enrollments) => {
          user.coursesList = (enrollments || []).filter(
            (e) => e.estado !== "baja",
          );
        },
        error: (err) => console.error("Error fetching enrollments:", err),
      });
  }

  /**
   * Baja de usuario (DELETE /api/v1/users/{id}). Pide confirmación antes;
   * el backend solo deja borrar cuentas ajenas a role_admin / role_dev,
   * por eso el botón se muestra solo a quien no es voluntario.
   */
  /**
   * Cuentas pendientes que quien está mirando puede aprobar o rechazar:
   * admin, cualquiera; voluntario, solo participantes (igual que el
   * backend, canApprove).
   */
  get pendientes(): DashboardUser[] {
    return this.users.filter((u) => this.puedeAprobar(u));
  }

  puedeAprobar(user: DashboardUser): boolean {
    if (!this.isRolePending(user.roleApprovalStatus)) return false;
    if (user.originalUser?.role_id === "role_admin") return false;
    return !this.isVolunteerOnly || user.role === "Participante";
  }

  togglePendientes() {
    this.soloPendientes = !this.soloPendientes;
    this.filterUsers();
  }

  /** HU-12 y HU-38: aprobar o rechazar una cuenta desde la lista. */
  decidirAprobacion(
    user: DashboardUser,
    decision: "approved" | "rejected",
    event: Event,
  ) {
    event.stopPropagation();
    const userId = user.originalUser?.user_id;
    if (!userId) return;
    const aprobar = decision === "approved";
    this.confirmation.confirm({
      header: aprobar ? "Aprobar cuenta" : "Rechazar cuenta",
      message: aprobar
        ? `¿Aprobar la cuenta de ${user.name} (${user.role})? Va a poder iniciar sesión y recibe un email de aviso.`
        : `¿Rechazar la cuenta de ${user.name} (${user.role})? No va a poder iniciar sesión y recibe un email de aviso.`,
      icon: aprobar ? "pi pi-check-circle" : "pi pi-exclamation-triangle",
      acceptLabel: aprobar ? "Aprobar" : "Rechazar",
      rejectLabel: "Cancelar",
      acceptButtonStyleClass: aprobar ? "p-button-success" : "p-button-danger",
      rejectButtonStyleClass: "p-button-text",
      defaultFocus: "reject",
      accept: () => {
        this.usersService.setApproval(userId, decision).subscribe({
          next: () => {
            this.messages.add({
              severity: "success",
              summary: aprobar ? "Cuenta aprobada" : "Cuenta rechazada",
              detail: `${user.name} recibe un email con la novedad.`,
            });
            this.loadUsers();
          },
          error: (err: Error) =>
            this.messages.add({
              severity: "error",
              summary: aprobar
                ? "No se pudo aprobar la cuenta"
                : "No se pudo rechazar la cuenta",
              detail: err.message,
            }),
        });
      },
    });
  }

  deleteUser(user: DashboardUser, event: Event) {
    event.stopPropagation();
    const userId = user.originalUser?.user_id;
    if (!userId) return;

    this.confirmation.confirm({
      header: "Eliminar usuario",
      message: `¿Seguro que querés eliminar a ${user.name}? Esta acción no se puede deshacer.`,
      icon: "pi pi-exclamation-triangle",
      acceptLabel: "Eliminar",
      rejectLabel: "Cancelar",
      acceptButtonStyleClass: "p-button-danger",
      rejectButtonStyleClass: "p-button-text",
      defaultFocus: "reject",
      accept: () => {
        this.usersService.deleteUser(userId).subscribe({
          next: () => {
            this.messages.add({
              severity: "success",
              summary: "Listo",
              detail: "El usuario se eliminó correctamente",
            });
            this.loadUsers();
          },
          error: (err: Error) => {
            this.messages.add({
              severity: "error",
              summary: "No se pudo eliminar el usuario",
              detail: err.message,
            });
          },
        });
      },
    });
  }

  uploadingPicture = false;

  resolveImageUrl(url?: string): string {
    return this.imageService.resolveImageUrl(url);
  }

  canEditPicture(): boolean {
    if (!this.selectedUser) return false;
    const isParticipant =
      this.selectedUser.originalUser?.role_id === "role_participant" ||
      this.selectedUser.role === "Participante";
    return (
      isParticipant &&
      (this.currentUserRoleId === "role_admin" ||
        this.currentUserRoleId === "admin" ||
        this.isVolunteerOnly)
    );
  }

  onDetailPictureSelected(event: any) {
    const file = event.target?.files?.[0];
    if (file && this.selectedUser?.originalUser?.user_id) {
      // HU-11 y D9: tipo y tamaño (el backend también lo controla).
      const invalida = validarFotoPerfil(file);
      if (invalida) {
        this.messages.add({
          severity: "warn",
          summary: "Imagen no válida",
          detail: invalida,
        });
        return;
      }
      this.uploadingPicture = true;
      const userId = this.selectedUser.originalUser.user_id;

      this.rxResumeService.uploadUserPictureByAdmin(userId, file).subscribe({
        next: (resume) => {
          this.uploadingPicture = false;
          const newPic = resume.data?.basics?.picture;
          if (newPic && this.selectedUser) {
            this.selectedUser.image_url = newPic;
            if (this.selectedUser.profileData?.role_data) {
              this.selectedUser.profileData.role_data.image_url = newPic;
            }
            const inList = this.users.find((u) => u.id === userId);
            if (inList) inList.image_url = newPic;

            const partId =
              this.selectedUser.profileData?.role_data?.participant_id;
            if (partId) {
              this.usersService
                .updateRoleData("participants", partId, { image_url: newPic })
                .subscribe({
                  error: (err) =>
                    console.error("Error sincronizando imagen en rol:", err),
                });
            }
          }
        },
        error: (err) => {
          console.error("Error al subir foto:", err);
          this.uploadingPicture = false;
          this.messages.add({
            severity: "error",
            summary: "No se pudo subir la foto",
            detail: err.message || "Probá de nuevo en unos minutos.",
          });
        },
      });
    }
  }

  onPictureSelected(event: any) {
    const file = event.target.files[0];
    if (file && this.selectedUser?.originalUser?.user_id) {
      const invalida = validarFotoPerfil(file);
      if (invalida) {
        this.messages.add({
          severity: "warn",
          summary: "Imagen no válida",
          detail: invalida,
        });
        return;
      }
      this.uploadingPicture = true;
      this.rxResumeService
        .uploadUserPictureByAdmin(this.selectedUser.originalUser.user_id, file)
        .subscribe({
          next: (resume) => {
            this.uploadingPicture = false;
            if (resume.data?.basics?.picture) {
              this.editForm.image_url = resume.data.basics.picture;
              // Update immediately in db
              this.saveEdit();
            }
          },
          error: (err) => {
            console.error("Error al subir foto:", err);
            this.uploadingPicture = false;
            this.messages.add({
              severity: "error",
              summary: "No se pudo subir la foto",
              detail: err.message || "Probá de nuevo en unos minutos.",
            });
          },
        });
    }
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
        this.messages.add({
          severity: "error",
          summary: "No se pudieron guardar los cambios",
          detail: err.message,
        });
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

  private buildParticipantPayload(): Record<string, any> {
    const raw = this.editForm;
    const payload: Record<string, any> = {
      birthdate: raw.birthdate
        ? new Date(raw.birthdate).toISOString()
        : undefined,
      dni: raw.dni || "",
      address: raw.address || "",
      gender: raw.gender || "",
      emergency_contact_name: raw.emergency_contact_name || "",
      emergency_contact_phone: raw.emergency_contact_phone || "",
      image_url: raw.image_url || "",
    };

    if (raw.approval_status) {
      payload["approval_status"] = raw.approval_status;
      if (raw.approval_status === "approved" && this.currentUserId) {
        payload["approved_by"] = this.currentUserId;
      }
    }
    return payload;
  }

  private buildGenericRolePayload(): Record<string, any> {
    const rolePayload: any = { ...this.editForm };
    delete rolePayload.first_name;
    delete rolePayload.last_name;
    delete rolePayload.phone;
    delete rolePayload.status;

    if (rolePayload.approval_status === "approved" && this.currentUserId) {
      rolePayload.approved_by = this.currentUserId;
    }
    rolePayload.birthdate = rolePayload.birthdate
      ? new Date(rolePayload.birthdate).toISOString()
      : null;

    return rolePayload;
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

    if (key && !this.editForm[key]) {
      this.finishSave();
      return;
    }

    const targetRoleId = key ? this.editForm[key] : userId;
    const rolePayload =
      tableName === "participants"
        ? this.buildParticipantPayload()
        : this.buildGenericRolePayload();

    // HU-12 y HU-38: aprobar o rechazar una cuenta pendiente va por
    // PATCH /users/{id}/approval, que además manda el email al usuario.
    const decision = this.approvalDecision();
    if (decision) {
      delete rolePayload["approval_status"];
      delete rolePayload["approved_by"];
    }

    // Enviar petición para actualizar el rol
    this.usersService
      .updateRoleData(tableName, targetRoleId, rolePayload)
      .subscribe({
        next: () => {
          if (!decision) {
            this.finishSave();
            return;
          }
          this.usersService.setApproval(userId, decision).subscribe({
            next: () => this.finishSave(),
            error: (err: Error) =>
              this.messages.add({
                severity: "error",
                summary:
                  decision === "rejected"
                    ? "No se pudo rechazar la cuenta"
                    : "No se pudo aprobar la cuenta",
                detail: err.message,
              }),
          });
        },
        error: (err) => {
          console.error("Error updating role data:", err);
          this.messages.add({
            severity: "error",
            summary: "No se pudo guardar la información del rol",
            detail: err.message,
          });
        },
      });
  }

  private finishSave() {
    this.messages.add({
      severity: "success",
      summary: "Listo",
      detail: "Cambios guardados correctamente",
    });
    this.showEditModal = false;
    this.loadUsers(); // Recargar la lista para reflejar los cambios
  }

  goToEditCV(user: DashboardUser) {
    if (user?.originalUser?.user_id) {
      void this.router.navigate(["/dashboard/cv"], {
        queryParams: { userId: user.originalUser.user_id },
      });
    }
  }

  onDocumentSelected(event: any) {
    const file = event.target.files[0];
    const participantId =
      this.selectedUser?.profileData?.role_data?.participant_id;
    if (file && file.size > 5 * 1024 * 1024) {
      // HU-11: máximo 5 MB.
      this.messages.add({
        severity: "warn",
        summary: "Archivo muy grande",
        detail: "El documento no debe superar los 5 MB.",
      });
      event.target.value = "";
      return;
    }
    if (file && participantId) {
      this.uploadingDocument = true;
      this.usersService
        .uploadParticipantDocument(participantId, file, file.name)
        .subscribe({
          next: (doc) => {
            this.uploadingDocument = false;
            if (doc) {
              this.participantDocuments.push(doc);
            }
            event.target.value = ""; // Reset input
          },
          error: (err) => {
            console.error("Error al subir documento:", err);
            this.uploadingDocument = false;
            const errorMsg =
              err.error?.error ||
              err.error?.message ||
              err.message ||
              "Error al subir el documento.";
            alert(errorMsg);
          },
        });
    }
  }

  startEditDocument(doc: any) {
    this.editingDocId = doc.id;
    this.editingDocName = doc.name || "";
    this.editingDocFile = null;
  }

  cancelEditDocument() {
    this.editingDocId = null;
    this.editingDocName = "";
    this.editingDocFile = null;
  }

  onEditDocFileSelected(event: any) {
    this.editingDocFile = event.target?.files?.[0] || null;
  }

  saveEditDocument(docId: string) {
    const participantId =
      this.selectedUser?.profileData?.role_data?.participant_id;
    if (!participantId) return;

    if (!this.editingDocName.trim()) {
      alert("El nombre del documento no puede estar vacío.");
      return;
    }

    this.uploadingDocument = true;
    this.usersService
      .updateParticipantDocument(
        participantId,
        docId,
        this.editingDocFile || undefined,
        this.editingDocName.trim(),
      )
      .subscribe({
        next: (updatedDoc) => {
          this.uploadingDocument = false;
          const idx = this.participantDocuments.findIndex(
            (d) => d.id === docId,
          );
          if (idx !== -1) {
            this.participantDocuments[idx] = updatedDoc;
          }
          this.cancelEditDocument();
        },
        error: (err) => {
          console.error("Error al actualizar documento:", err);
          this.uploadingDocument = false;
          const errorMsg =
            err.error?.error ||
            err.error?.message ||
            err.message ||
            "Error al actualizar el documento.";
          alert(errorMsg);
        },
      });
  }

  deleteDocument(docId: string) {
    const participantId =
      this.selectedUser?.profileData?.role_data?.participant_id;
    if (
      participantId &&
      confirm("¿Estás seguro de que deseas eliminar este documento?")
    ) {
      this.usersService
        .deleteParticipantDocument(participantId, docId)
        .subscribe({
          next: () => {
            this.participantDocuments = this.participantDocuments.filter(
              (d) => d.id !== docId,
            );
            if (this.editingDocId === docId) {
              this.cancelEditDocument();
            }
          },
          error: (err) => {
            console.error("Error al eliminar documento:", err);
            const errorMsg =
              err.error?.error ||
              err.error?.message ||
              err.message ||
              "Error al eliminar el documento.";
            alert(errorMsg);
          },
        });
    }
  }

  downloadDocument(docId: string, docName: string, event: Event) {
    event.preventDefault(); // Prevent navigating to href
    const participantId =
      this.selectedUser?.profileData?.role_data?.participant_id;
    if (participantId) {
      this.usersService
        .downloadParticipantDocument(participantId, docId)
        .subscribe({
          next: (blob: Blob) => {
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = docName;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            a.remove();
          },
          error: (err) => {
            console.error("Error al descargar documento:", err);
            const errorMsg =
              err.error?.error ||
              err.error?.message ||
              err.message ||
              "No se pudo descargar el documento.";
            alert(errorMsg);
          },
        });
    }
  }
}

/**
 * HU-07 y HU-01: el alta de un participante exige Gmail, DNI (7 u 8
 * números), dirección y teléfono. Devuelve el mensaje de error o "".
 */
export function validateNewParticipant(form: {
  role_id?: string;
  email?: string;
  dni?: string;
  address?: string;
  phone?: string;
}): string {
  if (form.role_id !== "role_participant") return "";
  const email = String(form.email ?? "")
    .trim()
    .toLowerCase();
  if (!email.endsWith("@gmail.com")) {
    return "El correo del participante tiene que ser una cuenta de Gmail (@gmail.com).";
  }
  if (!/^\d{7,8}$/.test(String(form.dni ?? "").trim())) {
    return "El DNI es obligatorio y debe tener 7 u 8 números, sin puntos.";
  }
  if (String(form.address ?? "").trim().length < 3) {
    return "La dirección es obligatoria.";
  }
  if (!String(form.phone ?? "").trim()) {
    return "El teléfono es obligatorio.";
  }
  return "";
}
