import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import {
  AbstractControlOptions,
  FormBuilder,
  FormGroup,
  ReactiveFormsModule,
  Validators,
} from "@angular/forms";
import { Auth, signInWithPopup, GoogleAuthProvider } from "@angular/fire/auth";
import { from, switchMap, firstValueFrom, of, catchError } from "rxjs";
import { ProfileService } from "../../../services/profile.service";
import { ProfileResponse } from "../../../core/models/auth.model";
import {
  LucideAngularModule,
  User as UserIcon,
  Shield,
  Briefcase,
  Lock,
  CircleCheck,
  CircleAlert,
  Eye,
  EyeOff,
  Link,
  Unlink,
  Download,
  Trash2,
  ShieldAlert,
  FileText,
} from "lucide-angular";
import { RxResumeService } from "../../../services/rxresume.service";
import { UsersService } from "../../../services/users.service";
import { EducationService } from "../../../services/education.service";
import { JobOfferService } from "../../../services/job-offer.service";
import { ImageService } from "../../../services/image.service";

@Component({
  selector: "app-mi-perfil",
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, LucideAngularModule],
  templateUrl: "./mi-perfil.component.html",
  styleUrls: ["./mi-perfil.component.css"],
})
export class MiPerfilComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly profileService = inject(ProfileService);
  private readonly auth = inject(Auth);
  private readonly rxResumeService = inject(RxResumeService);
  private readonly usersService = inject(UsersService);
  private readonly educationService = inject(EducationService);
  private readonly jobOfferService = inject(JobOfferService);
  private readonly imageService = inject(ImageService);

  resolveImageUrl(url?: string | boolean | null): string {
    return this.imageService.resolveImageUrl(url);
  }

  readonly UserIcon = UserIcon;
  readonly ShieldIcon = Shield;
  readonly BriefcaseIcon = Briefcase;
  readonly LockIcon = Lock;
  readonly CheckCircleIcon = CircleCheck;
  readonly AlertCircleIcon = CircleAlert;
  readonly EyeIcon = Eye;
  readonly EyeOffIcon = EyeOff;
  readonly LinkIcon = Link;
  readonly UnlinkIcon = Unlink;
  readonly DownloadIcon = Download;
  readonly TrashIcon = Trash2;
  readonly ShieldAlertIcon = ShieldAlert;
  readonly FileTextIcon = FileText;

  isExported = false;
  isExporting = false;

  profileData: ProfileResponse | null = null;
  loading = true;
  error = "";

  googleLoading = false;
  googleSuccess = "";
  googleError = "";

  passwordForm: FormGroup;
  passwordLoading = false;
  passwordSuccess = "";
  passwordError = "";

  showOldPassword = false;
  showNewPassword = false;

  participantDocuments: any[] = [];
  uploadingDocument = false;
  showConfirmPassword = false;

  isEditingProfile = false;
  profileForm!: FormGroup;
  profileUpdateLoading = false;
  profileUpdateSuccess = "";
  profileUpdateError = "";

  constructor() {
    this.passwordForm = this.fb.group(
      {
        oldPassword: ["", Validators.required],
        newPassword: ["", [Validators.required, Validators.minLength(8)]],
        confirmPassword: ["", Validators.required],
      },
      { validators: [this.passwordMatchValidator] } as AbstractControlOptions,
    );
  }

  ngOnInit() {
    this.loadProfile();
  }

  loadProfile() {
    this.profileService.getMyProfile().subscribe({
      next: (data) => {
        this.profileData = data;
        this.initProfileForm();
        if (data.role_data?.["participant_id"]) {
          this.usersService
            .getParticipantDocuments(data.role_data["participant_id"] as string)
            .subscribe({
              next: (docs) => (this.participantDocuments = docs || []),
              error: (err) => console.error("Error fetching docs", err),
            });
        }
        this.loading = false;
      },
      error: (_err) => {
        this.error = "Error al cargar el perfil";
        this.loading = false;
      },
    });
  }

  initProfileForm() {
    if (!this.profileData) return;
    const user = this.profileData.user;
    const roleData = this.profileData.role_data || {};

    const controls: any = {
      email: [user.email || "", [Validators.required, Validators.email]],
      phone: [user.phone || ""],
    };

    const role = user.role_id;
    if (role === "role_company") {
      if (this.canEditCompanyFields()) {
        controls["cuit"] = [roleData["cuit"] || ""];
        controls["business_name"] = [roleData["business_name"] || ""];
      }
    } else if (role === "role_participant") {
      controls["image_url"] = [user.image_url || roleData["image_url"] || ""];
      controls["emergency_contact_name"] = [
        roleData["emergency_contact_name"] || "",
      ];
      controls["emergency_contact_phone"] = [
        roleData["emergency_contact_phone"] || "",
      ];
      controls["address"] = [roleData["address"] || ""];
    } else if (role === "role_teacher") {
      controls["specialization"] = [roleData["specialization"] || ""];
    }

    this.profileForm = this.fb.group(controls);
  }

  canEditCompanyFields(): boolean {
    if (!this.profileData?.role_data) return false;
    const status = this.profileData.role_data["approval_status"];
    const cuit = this.profileData.role_data["cuit"];
    const businessName = this.profileData.role_data["business_name"];
    return status === "pending" || !cuit || !businessName;
  }

  toggleEditProfile() {
    this.isEditingProfile = !this.isEditingProfile;
    if (!this.isEditingProfile && this.profileData) {
      this.initProfileForm();
      this.profileUpdateSuccess = "";
      this.profileUpdateError = "";
    }
  }

  saveProfile() {
    if (this.profileForm.invalid) return;
    this.profileUpdateLoading = true;
    this.profileUpdateSuccess = "";
    this.profileUpdateError = "";

    const data = this.profileForm.value;

    this.profileService.updateProfile(data).subscribe({
      next: () => {
        this.profileUpdateSuccess = "Perfil actualizado correctamente.";
        this.profileUpdateLoading = false;
        this.isEditingProfile = false;
        this.loadProfile();
      },
      error: (err) => {
        this.profileUpdateError =
          err.error?.error || "Error al actualizar el perfil.";
        this.profileUpdateLoading = false;
      },
    });
  }

  uploadingPicture = false;

  onPictureSelected(event: any) {
    const file = event.target.files[0];
    if (file) {
      this.uploadingPicture = true;
      this.rxResumeService.uploadPicture(file).subscribe({
        next: (resume) => {
          this.uploadingPicture = false;
          if (resume.data?.basics?.picture) {
            this.profileForm.patchValue({
              image_url: resume.data.basics.picture,
            });
            this.saveProfile();
          }
        },
        error: (err) => {
          console.error("Error al subir foto:", err);
          this.uploadingPicture = false;
          alert("Error al subir la foto de perfil.");
        },
      });
    }
  }

  onDocumentSelected(event: any) {
    const file = event.target.files[0];
    const participantId = this.profileData?.role_data?.[
      "participant_id"
    ] as string;
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
            event.target.value = "";
          },
          error: (err) => {
            console.error("Error al subir documento:", err);
            this.uploadingDocument = false;
            alert("Error al subir el documento.");
          },
        });
    }
  }

  deleteDocument(docId: string) {
    const participantId = this.profileData?.role_data?.[
      "participant_id"
    ] as string;
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
          },
          error: (err) => {
            console.error("Error al eliminar documento:", err);
            alert("Error al eliminar el documento.");
          },
        });
    }
  }

  downloadDocument(docId: string, docName: string, event: Event) {
    event.preventDefault();
    const participantId = this.profileData?.role_data?.[
      "participant_id"
    ] as string;
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
            alert("No se pudo descargar el documento.");
          },
        });
    }
  }

  getRoleName(roleId?: string): string {
    if (!roleId) return "";
    const roleMap: Record<string, string> = {
      role_participant: "Participante",
      role_company: "Empresa",
      role_teacher: "Docente",
      role_volunteer: "Voluntario",
      role_admin: "Responsable",
      role_dev: "Desarrollador",
      participant: "Participante",
      company: "Empresa",
      teacher: "Docente",
      volunteer: "Voluntario",
      admin: "Responsable",
      dev: "Desarrollador",
    };
    return roleMap[roleId.toLowerCase()] || roleId;
  }

  getRoleKeyLabel(key: string): string {
    const k = String(key);
    const labelMap: Record<string, string> = {
      volunteer_id: "ID de Voluntario",
      participant_id: "ID de Participante",
      teacher_id: "ID de Docente",
      company_id: "ID de Empresa",
      approval_status: "Estado de Aprobación",
      approval_date: "Fecha de Aprobación",
      approved_by: "Aprobado Por",
      created_at: "Fecha de Creación",
      updated_at: "Última Actualización",
      deleted_at: "Fecha de Eliminación",
      specialization: "Especialidad",
      is_external: "Docente Externo",
      business_name: "Razón Social",
      cuit: "CUIT",
      birthdate: "Fecha de Nacimiento",
      dni: "DNI",
      address: "Dirección",
      gender: "Género",
      image_url: "Imagen de Perfil",
      doc_url: "Documentación",
      emergency_contact_name: "Contacto de Emergencia",
      emergency_contact_phone: "Teléfono de Emergencia",
    };
    return (
      labelMap[k] ||
      k.replaceAll("_", " ").replace(/\b\w/g, (l) => l.toUpperCase())
    );
  }

  formatRoleValue(
    key: string,
    value: string | boolean | null | undefined,
  ): string {
    const k = key;
    if (value === null || value === undefined || value === "")
      return "No especificado";
    if (typeof value === "boolean") return value ? "Sí" : "No";

    if (k === "approval_status") {
      const statusMap: Record<string, string> = {
        approved: "Aprobado",
        pending: "Pendiente",
        rejected: "Rechazado",
      };
      return statusMap[value] || value;
    }

    if (k === "gender") {
      const genderMap: Record<string, string> = {
        male: "Masculino",
        female: "Femenino",
        other: "Otro",
      };
      return genderMap[value] || value;
    }

    if (
      (k.endsWith("_at") || k.endsWith("date")) &&
      typeof value === "string" &&
      !Number.isNaN(Date.parse(value))
    ) {
      try {
        return new Date(value).toLocaleDateString("es-AR", {
          year: "numeric",
          month: "long",
          day: "numeric",
        });
      } catch {
        return value;
      }
    }

    return String(value);
  }

  passwordMatchValidator(g: FormGroup) {
    return g.get("newPassword")?.value === g.get("confirmPassword")?.value
      ? null
      : { mismatch: true };
  }

  togglePasswordVisibility(field: "old" | "new" | "confirm") {
    if (field === "old") this.showOldPassword = !this.showOldPassword;
    if (field === "new") this.showNewPassword = !this.showNewPassword;
    if (field === "confirm")
      this.showConfirmPassword = !this.showConfirmPassword;
  }

  changePassword() {
    if (this.passwordForm.invalid) return;

    this.passwordLoading = true;
    this.passwordSuccess = "";
    this.passwordError = "";

    const { oldPassword, newPassword } = this.passwordForm.value;

    this.profileService.updateMyPassword(oldPassword, newPassword).subscribe({
      next: () => {
        this.passwordSuccess = "Contraseña actualizada correctamente.";
        this.passwordForm.reset();
        this.passwordLoading = false;
      },
      error: (err) => {
        this.passwordError =
          err.error?.error || "Error al actualizar la contraseña.";
        this.passwordLoading = false;
      },
    });
  }

  linkGoogle() {
    this.googleLoading = true;
    this.googleSuccess = "";
    this.googleError = "";

    const provider = new GoogleAuthProvider();
    from(signInWithPopup(this.auth, provider))
      .pipe(
        switchMap((credential) =>
          from(
            credential.user
              .getIdToken()
              .then((idToken) => ({ idToken, googleId: credential.user.uid })),
          ),
        ),
        switchMap(({ idToken, googleId }) =>
          this.profileService.linkGoogleAccount(idToken, googleId),
        ),
      )
      .subscribe({
        next: () => {
          this.googleSuccess = "Cuenta de Google vinculada correctamente.";
          this.googleLoading = false;
          this.loadProfile();
        },
        error: (err) => {
          this.googleError =
            err.error?.error || "Error al vincular la cuenta de Google.";
          this.googleLoading = false;
        },
      });
  }

  unlinkGoogle() {
    this.googleLoading = true;
    this.googleSuccess = "";
    this.googleError = "";

    this.profileService.unlinkGoogleAccount().subscribe({
      next: () => {
        this.googleSuccess = "Cuenta de Google desvinculada correctamente.";
        this.googleLoading = false;
        this.loadProfile();
      },
      error: (err) => {
        this.googleError =
          err.error?.error || "Error al desvincular la cuenta de Google.";
        this.googleLoading = false;
      },
    });
  }

  private async fetchParticipantData(): Promise<{
    enrolledCourses: any[];
    jobApplications: any[];
    userResume: any;
  }> {
    const [courses, jobApplications, userResume] = await Promise.all([
      firstValueFrom(
        this.educationService
          .listMyEnrolledCourses()
          .pipe(catchError(() => of([]))),
      ),
      firstValueFrom(
        this.jobOfferService.getMyApplications().pipe(catchError(() => of([]))),
      ),
      firstValueFrom(
        this.rxResumeService.getMyResume().pipe(catchError(() => of(null))),
      ),
    ]);

    let enrolledCourses: any[] = [];
    if (courses?.length) {
      enrolledCourses = await Promise.all(
        courses.map(async (c) => {
          try {
            const detail = await firstValueFrom(
              this.educationService
                .getMyCourse(c.id)
                .pipe(catchError(() => of(null))),
            );
            return detail || { curso: c, actividades: [], progreso: null };
          } catch {
            return { curso: c, actividades: [], progreso: null };
          }
        }),
      );
    }

    return {
      enrolledCourses,
      jobApplications: jobApplications || [],
      userResume,
    };
  }

  private async fetchTeacherData(): Promise<any[]> {
    const courses = await firstValueFrom(
      this.educationService.listMyCourses().pipe(catchError(() => of([]))),
    );
    if (!courses?.length) return [];

    return Promise.all(
      courses.map(async (c) => {
        try {
          const acts = await firstValueFrom(
            this.educationService
              .listActivities(c.id)
              .pipe(catchError(() => of([]))),
          );
          return { curso: c, actividades: acts || [] };
        } catch {
          return { curso: c, actividades: [] };
        }
      }),
    );
  }

  private async fetchCompanyData(
    empresaId: string,
    userId: string,
  ): Promise<any[]> {
    const allOffers = await firstValueFrom(
      this.jobOfferService.getOffers().pipe(catchError(() => of([]))),
    );
    return (allOffers || []).filter(
      (o) => o.empresa_id === empresaId || o.empresa_id === userId,
    );
  }

  private renderPdfHeader(
    doc: any,
    margen: number,
    anchoUtil: number,
    y: number,
  ): number {
    doc.setFont("helvetica", "bold");
    doc.setFontSize(16);
    doc.setTextColor("#485839");
    doc.text("Centro de Ecología Integral Buen Vivir", margen, y);
    y += 6;

    doc.setFontSize(11);
    doc.setTextColor("#1d1d1b");
    doc.text("INFORME DE DATOS PERSONALES - DERECHO DE ACCESO", margen, y);
    y += 5;

    doc.setFont("helvetica", "normal");
    doc.setFontSize(9);
    doc.setTextColor("#66655b");
    doc.text(
      "Conforme a los Arts. 14 y 15 de la Ley Nacional Nº 25.326 de Protección de los Datos Personales",
      margen,
      y,
    );
    y += 4;
    doc.text(
      `Fecha de emisión: ${new Date().toLocaleString("es-AR")}`,
      margen,
      y,
    );
    y += 4;

    doc.setDrawColor("#d6d4c6");
    doc.line(margen, y, margen + anchoUtil, y);
    return y + 6;
  }

  private renderPdfAccountAndRole(
    doc: any,
    user: any,
    roleData: any,
    margen: number,
    y: number,
    checkPageBreak: (h: number) => void,
  ): number {
    doc.setFont("helvetica", "bold");
    doc.setFontSize(12);
    doc.setTextColor("#485839");
    doc.text("1. Información de la Cuenta y del Titular", margen, y);
    y += 6;

    doc.setFont("helvetica", "normal");
    doc.setFontSize(9);
    doc.setTextColor("#1d1d1b");

    const items = [
      `Nombre y Apellido: ${user.first_name || ""} ${user.last_name || ""}`.trim(),
      `Correo Electrónico: ${user.email || "No registrado"}`,
      `Teléfono: ${user.phone || "No registrado"}`,
      `Rol del Sistema: ${user.role?.name || user.role_id || "No especificado"}`,
      `Estado de Cuenta: ${user.status || "Activo"}`,
      `Fecha de Alta en el Sistema: ${user.created_at ? new Date(user.created_at).toLocaleDateString("es-AR") : "No disponible"}`,
    ];

    for (const item of items) {
      checkPageBreak(5);
      doc.text(`• ${item}`, margen + 2, y);
      y += 5;
    }
    y += 3;

    const roleKeys = Object.keys(roleData).filter(
      (k) =>
        !["approval_date", "approved_by", "deleted_at"].includes(k) &&
        roleData[k] !== undefined &&
        roleData[k] !== null &&
        roleData[k] !== "",
    );

    if (roleKeys.length > 0) {
      checkPageBreak(12);
      doc.setFont("helvetica", "bold");
      doc.setFontSize(12);
      doc.setTextColor("#485839");
      doc.text("2. Datos Específicos del Rol Registrado", margen, y);
      y += 6;

      doc.setFont("helvetica", "normal");
      doc.setFontSize(9);
      doc.setTextColor("#1d1d1b");

      for (const k of roleKeys) {
        checkPageBreak(5);
        doc.text(
          `• ${this.getRoleKeyLabel(k)}: ${this.formatRoleValue(k, roleData[k])}`,
          margen + 2,
          y,
        );
        y += 5;
      }
      y += 3;
    }

    return y;
  }

  private renderEnrolledCourseItem(
    doc: any,
    item: any,
    margen: number,
    y: number,
    checkPageBreak: (h: number) => void,
  ): number {
    checkPageBreak(12);
    doc.setFont("helvetica", "bold");
    doc.setFontSize(10);
    doc.setTextColor("#1d1d1b");
    doc.text(`Curso: ${item.curso?.title || "Sin título"}`, margen + 2, y);
    y += 5;

    doc.setFont("helvetica", "normal");
    doc.setFontSize(9);
    if (item.progreso) {
      doc.text(
        `  Progreso formativo: ${item.progreso.porcentajeAsistencia || 0}% de asistencia (${item.progreso.actividadesAsistidas || 0} de ${item.progreso.actividadesTotal || 0} actividades entregadas)`,
        margen + 2,
        y,
      );
      y += 5;
    }

    if (item.actividades?.length) {
      doc.setFont("helvetica", "bold");
      doc.setFontSize(8.5);
      doc.setTextColor("#66655b");
      doc.text("  Actividades y Entregas:", margen + 2, y);
      y += 4;
      doc.setFont("helvetica", "normal");
      doc.setTextColor("#1d1d1b");

      for (const act of item.actividades) {
        checkPageBreak(6);
        const estado = act.entrega
          ? `Entregado el ${new Date(act.entrega.fechaEntrega).toLocaleDateString("es-AR")}`
          : "Sin entregar";
        doc.text(`    - ${act.titulo}: ${estado}`, margen + 4, y);
        y += 4;
      }
    }
    return y + 3;
  }

  private renderParticipantCourses(
    doc: any,
    courses: any[],
    margen: number,
    y: number,
    checkPageBreak: (h: number) => void,
  ): number {
    checkPageBreak(14);
    doc.setFont("helvetica", "bold");
    doc.setFontSize(12);
    doc.setTextColor("#485839");
    doc.text("3. Cursos Inscriptos, Actividades y Progreso", margen, y);
    y += 6;

    if (!courses.length) {
      doc.setFont("helvetica", "italic");
      doc.setFontSize(9);
      doc.setTextColor("#66655b");
      doc.text(
        "No se registran inscripciones a cursos actualmente.",
        margen + 2,
        y,
      );
      return y + 6;
    }

    for (const item of courses) {
      y = this.renderEnrolledCourseItem(doc, item, margen, y, checkPageBreak);
    }
    return y;
  }

  private renderTeacherCourses(
    doc: any,
    courses: any[],
    margen: number,
    y: number,
    checkPageBreak: (h: number) => void,
  ): number {
    checkPageBreak(14);
    doc.setFont("helvetica", "bold");
    doc.setFontSize(12);
    doc.setTextColor("#485839");
    doc.text("3. Cursos a Cargo y Actividades Docentes", margen, y);
    y += 6;

    if (!courses.length) {
      doc.setFont("helvetica", "italic");
      doc.setFontSize(9);
      doc.setTextColor("#66655b");
      doc.text("No se registran cursos dictados actualmente.", margen + 2, y);
      return y + 6;
    }

    for (const tc of courses) {
      checkPageBreak(10);
      doc.setFont("helvetica", "bold");
      doc.setFontSize(10);
      doc.setTextColor("#1d1d1b");
      doc.text(
        `Curso: ${tc.curso?.title || "Sin título"} (${tc.curso?.mode || "Modalidad general"})`,
        margen + 2,
        y,
      );
      y += 5;

      doc.setFont("helvetica", "normal");
      doc.setFontSize(9);
      doc.text(
        `  Total de actividades creadas: ${tc.actividades?.length || 0}`,
        margen + 2,
        y,
      );
      y += 6;
    }
    return y;
  }

  private renderPdfResumeBlock(
    doc: any,
    userResume: any,
    margen: number,
    y: number,
    checkPageBreak: (h: number) => void,
  ): number {
    if (!userResume?.data?.basics) return y;

    checkPageBreak(14);
    doc.setFont("helvetica", "bold");
    doc.setFontSize(10);
    doc.setTextColor("#1d1d1b");
    doc.text("Currículum Vitae Registrado:", margen + 2, y);
    y += 5;

    doc.setFont("helvetica", "normal");
    doc.setFontSize(9);
    doc.text(`• Titular: ${userResume.data.basics.name || ""}`, margen + 4, y);
    y += 4.5;
    if (userResume.data.basics.headline) {
      doc.text(
        `• Titular / Especialidad: ${userResume.data.basics.headline}`,
        margen + 4,
        y,
      );
      y += 4.5;
    }
    if (userResume.data.work?.length) {
      doc.text(
        `• Experiencias laborales registradas: ${userResume.data.work.length}`,
        margen + 4,
        y,
      );
      y += 4.5;
    }
    if (userResume.data.education?.length) {
      doc.text(
        `• Estudios y certificaciones: ${userResume.data.education.length}`,
        margen + 4,
        y,
      );
      y += 4.5;
    }
    return y + 3;
  }

  private renderParticipantLabor(
    doc: any,
    jobApplications: any[],
    userResume: any,
    margen: number,
    y: number,
    checkPageBreak: (h: number) => void,
  ): number {
    checkPageBreak(14);
    doc.setFont("helvetica", "bold");
    doc.setFontSize(12);
    doc.setTextColor("#485839");
    doc.text("4. Inserción Laboral y Postulaciones", margen, y);
    y += 6;

    if (!jobApplications.length) {
      doc.setFont("helvetica", "italic");
      doc.setFontSize(9);
      doc.setTextColor("#66655b");
      doc.text(
        "No se registran postulaciones a ofertas laborales.",
        margen + 2,
        y,
      );
      y += 6;
    } else {
      doc.setFont("helvetica", "normal");
      doc.setFontSize(9);
      doc.setTextColor("#1d1d1b");
      for (const app of jobApplications) {
        checkPageBreak(5);
        const fecha = app.fecha_postulacion
          ? new Date(app.fecha_postulacion).toLocaleDateString("es-AR")
          : "Fecha no registrada";
        doc.text(
          `• Postulación a Oferta ID ${app.oferta_id || ""}: Estado ${app.estado || "Pendiente"} (${fecha})`,
          margen + 2,
          y,
        );
        y += 5;
      }
      y += 3;
    }

    return this.renderPdfResumeBlock(
      doc,
      userResume,
      margen,
      y,
      checkPageBreak,
    );
  }

  private renderCompanyOffers(
    doc: any,
    companyOffers: any[],
    margen: number,
    y: number,
    checkPageBreak: (h: number) => void,
  ): number {
    checkPageBreak(14);
    doc.setFont("helvetica", "bold");
    doc.setFontSize(12);
    doc.setTextColor("#485839");
    doc.text("4. Ofertas Laborales Publicadas", margen, y);
    y += 6;

    if (!companyOffers.length) {
      doc.setFont("helvetica", "italic");
      doc.setFontSize(9);
      doc.setTextColor("#66655b");
      doc.text("No se registran ofertas laborales publicadas.", margen + 2, y);
      return y + 6;
    }

    doc.setFont("helvetica", "normal");
    doc.setFontSize(9);
    doc.setTextColor("#1d1d1b");
    for (const off of companyOffers) {
      checkPageBreak(5);
      doc.text(
        `• ${off.titulo} - Modalidad: ${off.modalidad} | Estado: ${off.estado}`,
        margen + 2,
        y,
      );
      y += 5;
    }
    return y + 3;
  }

  private renderParticipantDocumentsSection(
    doc: any,
    margen: number,
    y: number,
    checkPageBreak: (h: number) => void,
  ): number {
    if (!this.participantDocuments?.length) return y;

    checkPageBreak(14);
    doc.setFont("helvetica", "bold");
    doc.setFontSize(12);
    doc.setTextColor("#485839");
    doc.text("5. Documentación de Legajo Registrada", margen, y);
    y += 6;

    doc.setFont("helvetica", "normal");
    doc.setFontSize(9);
    doc.setTextColor("#1d1d1b");
    for (const docItem of this.participantDocuments) {
      checkPageBreak(5);
      const docName =
        docItem.nombre || docItem.name || docItem.type || "Documento";
      const subido = docItem.uploaded_at
        ? new Date(docItem.uploaded_at).toLocaleDateString("es-AR")
        : "Registrado";
      doc.text(`• ${docName} (${subido})`, margen + 2, y);
      y += 5;
    }
    return y + 3;
  }

  async exportUserData(): Promise<void> {
    if (!this.profileData || this.isExporting) return;

    this.isExporting = true;
    try {
      const user = this.profileData.user;
      const roleId = user?.role_id || "";
      const roleData = this.profileData.role_data || {};

      let enrolledCourses: any[] = [];
      let teacherCourses: any[] = [];
      let jobApplications: any[] = [];
      let companyOffers: any[] = [];
      let userResume: any = null;

      if (roleId === "role_participant") {
        const participantData = await this.fetchParticipantData();
        enrolledCourses = participantData.enrolledCourses;
        jobApplications = participantData.jobApplications;
        userResume = participantData.userResume;
      } else if (roleId === "role_teacher") {
        teacherCourses = await this.fetchTeacherData();
      } else if (roleId === "role_company") {
        const empresaId = (roleData["company_id"] as string) || user.user_id;
        companyOffers = await this.fetchCompanyData(empresaId, user.user_id);
      }

      const { jsPDF } = await import("jspdf");
      const doc = new jsPDF({ unit: "mm", format: "a4" });
      const margen = 15;
      const anchoUtil = 210 - margen * 2;
      let y = margen;

      const checkPageBreak = (neededHeight: number) => {
        if (y + neededHeight > 280) {
          doc.addPage();
          y = margen;
        }
      };

      y = this.renderPdfHeader(doc, margen, anchoUtil, y);
      y = this.renderPdfAccountAndRole(
        doc,
        user,
        roleData,
        margen,
        y,
        checkPageBreak,
      );

      if (roleId === "role_participant") {
        y = this.renderParticipantCourses(
          doc,
          enrolledCourses,
          margen,
          y,
          checkPageBreak,
        );
        y = this.renderParticipantLabor(
          doc,
          jobApplications,
          userResume,
          margen,
          y,
          checkPageBreak,
        );
      } else if (roleId === "role_teacher") {
        y = this.renderTeacherCourses(
          doc,
          teacherCourses,
          margen,
          y,
          checkPageBreak,
        );
      } else if (roleId === "role_company") {
        y = this.renderCompanyOffers(
          doc,
          companyOffers,
          margen,
          y,
          checkPageBreak,
        );
      }

      this.renderParticipantDocumentsSection(doc, margen, y, checkPageBreak);

      const pageCount = doc.getNumberOfPages();
      for (let i = 1; i <= pageCount; i++) {
        doc.setPage(i);
        doc.setFont("helvetica", "normal");
        doc.setFontSize(8);
        doc.setTextColor("#66655b");
        doc.text(
          "Centro de Ecología Integral Buen Vivir - Documento generado en ejercicio del Derecho de Acceso (Ley 25.326)",
          margen,
          290,
        );
        doc.text(`Página ${i} de ${pageCount}`, 210 - margen - 20, 290);
      }

      const dni = (roleData["dni"] as string) || user.user_id;
      doc.save(`mis_datos_buen_vivir_${dni}.pdf`);

      this.isExported = true;
      setTimeout(() => (this.isExported = false), 5000);
    } catch (err) {
      console.error("Error al generar el informe en PDF:", err);
    } finally {
      this.isExporting = false;
    }
  }
}
