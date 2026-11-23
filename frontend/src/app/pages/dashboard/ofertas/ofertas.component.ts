import { Component, DestroyRef, OnInit, inject } from "@angular/core";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { ActivatedRoute, Router, RouterLink } from "@angular/router";
import { CardModule } from "primeng/card";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { InputTextModule } from "primeng/inputtext";
import { DialogModule } from "primeng/dialog";
import { ConfirmDialogModule } from "primeng/confirmdialog";
import { ConfirmationService, MessageService } from "primeng/api";
import { ToastModule } from "primeng/toast";
import { DropdownModule } from "primeng/dropdown";
import {
  JobOfferService,
  JobOffer,
  JobApplication,
} from "../../../services/job-offer.service";
import { AuthService } from "../../../services/auth.service";
import { RxResumeService, Resume } from "../../../services/rxresume.service";
import { UsersService } from "../../../services/users.service";
import { ProfileService } from "../../../services/profile.service";
import { User } from "../../../core/models/auth.model";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";

export const CENTRO_CONTACTO = {
  nombre: 'Equipo de Inserción Laboral — Centro "La Otra Orilla"',
  email: "ceibuenvivir7@gmail.com",
  telefonos: [
    { label: "+54 9 3547 61 2118", href: "tel:+5493547612118" },
    { label: "+54 9 3547 45 6184", href: "tel:+5493547456184" },
  ],
  horario: "Lunes a viernes de 09:00 a 18:00 hs",
} as const;

@Component({
  selector: "app-ofertas",
  standalone: true,
  imports: [
    EmptyStateComponent,
    CommonModule,
    FormsModule,
    RouterLink,
    CardModule,
    ButtonModule,
    TagModule,
    InputTextModule,
    DialogModule,
    ConfirmDialogModule,
    ToastModule,
    DropdownModule,
  ],
  providers: [ConfirmationService, MessageService],
  templateUrl: "./ofertas.component.html",
  styleUrls: ["./ofertas.component.css"],
})
export class OfertasComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);

  readonly centroContacto = CENTRO_CONTACTO;

  searchTerm: string = "";
  private busquedaTimer?: ReturnType<typeof setTimeout>;
  showDialog = false;
  isEditing = false;
  currentRequisito = "";

  // Renewal dialog state
  showRenewDialog = false;
  renewDate = "";
  renewVacancies = 1;
  selectedJobForRenew: JobOffer | null = null;

  newJob: JobOffer = {
    empresa_id: "",
    titulo: "",
    descripcion: "",
    ubicacion: "",
    modalidad: "Presencial",
    tipo_empleo: "Tiempo Completo",
    salario: "",
    horario: "",
    vacantes: 1,
    requisitos: [],
    fecha_limite: "",
    estado: "Activa",
  };

  jobs: JobOffer[] = [];
  filteredJobs: JobOffer[] = [];
  isCompany = false;
  isParticipant = false;
  isAdminOrVolunteer = false;
  /** Approving and rejecting offers is exclusive to administrators (HU-41). */
  isAdmin = false;
  userId = "";
  currentUserName = "";

  // Section and tab filters
  activeSection: "vigentes" | "historial" = "vigentes";
  currentTab = "Disponibles"; // For participants
  companyTab = "Todas"; // For companies in vigentes
  adminTab = "Pendiente"; // For admin/volunteer in vigentes
  historyTab = "Todas"; // For company/staff in historial

  employmentTypes = [
    { label: "Todos", value: null },
    { label: "Tiempo completo", value: "Tiempo Completo" },
    { label: "Medio tiempo", value: "Medio Tiempo" },
    { label: "Pasantía", value: "Pasantía" },
  ];
  selectedEmploymentType: string | null = null;

  applicationsDialogVisible = false;
  currentJob: JobOffer | null = null;
  currentJobTitle = "";
  currentCandidates: { app: JobApplication; resume?: Resume }[] = [];
  appliedJobs: Set<string> = new Set();
  myApplicationsMap: Map<string, string> = new Map(); // oferta_id -> estado

  // Cached applications for finalized offers follow-up
  applicationsByOffer: Map<string, JobApplication[]> = new Map();
  loadingApplicationsFor: Set<string> = new Set();
  applicationsErrorFor: Set<string> = new Set();

  // Admin / Volunteer Apply
  participants: User[] = [];
  selectedParticipant: string = "";
  showApplyDialog = false;
  jobToApply: JobOffer | null = null;

  showAcceptanceModal = false;
  acceptanceMessage = "";
  acceptedOfferForContact: JobOffer | null = null;

  constructor(
    private readonly jobOfferService: JobOfferService,
    private readonly authService: AuthService,
    private readonly rxResumeService: RxResumeService,
    private readonly usersService: UsersService,
    private readonly profileService: ProfileService,
    private readonly confirmationService: ConfirmationService,
    private readonly messageService: MessageService,
  ) {}

  ngOnInit() {
    this.route.queryParamMap
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((params) => {
        const nextSection =
          params.get("seccion") === "historial" ? "historial" : "vigentes";
        if (this.activeSection !== nextSection) {
          this.activeSection = nextSection;
          this.filterJobs();
          if (nextSection === "historial") {
            this.loadHistoryApplications();
          }
        }
      });

    this.authService.isAuthenticated().subscribe({
      next: (res) => {
        const role = res?.role_id ?? res?.user?.role_id;
        if (role) {
          this.isCompany = role === "role_company";
          this.isParticipant = role === "role_participant";
          this.isAdmin = role === "role_admin" || role === "role_dev";
          this.isAdminOrVolunteer = this.isAdmin || role === "role_volunteer";
        }
        const r = res as { user_id?: string };
        const uid = r.user_id || res?.user?.user_id;
        if (uid) {
          this.userId = uid;
        }

        if (this.isParticipant) {
          this.activeSection = "vigentes";
          this.loadMyApplications();
        }
        if (this.isAdminOrVolunteer) {
          this.loadParticipants();
        }
        this.filterJobs();
        if (
          this.activeSection === "historial" &&
          (this.isCompany || this.isAdminOrVolunteer)
        ) {
          this.loadHistoryApplications();
        }

        this.profileService.getMyProfile().subscribe({
          next: (profile) => {
            if (this.isCompany && profile.role_data?.business_name) {
              this.currentUserName = profile.role_data.business_name;
            } else if (profile.user?.first_name && profile.user?.last_name) {
              this.currentUserName = `${profile.user.first_name} ${profile.user.last_name}`;
            }
          },
        });
      },
    });
    this.loadJobs();
  }

  get vigentesCount(): number {
    return this.jobs.filter((j) => this.isVigentStatus(j.estado)).length;
  }

  get historialCount(): number {
    return this.jobs.filter((j) => this.isHistoryStatus(j.estado)).length;
  }

  get modalAcceptedCount(): number {
    return this.currentCandidates.filter((c) =>
      this.isAcceptedStatus(c.app.estado),
    ).length;
  }

  get modalPendingCount(): number {
    return this.currentCandidates.filter(
      (c) =>
        !this.isAcceptedStatus(c.app.estado) &&
        !this.isRejectedStatus(c.app.estado),
    ).length;
  }

  isEstado(actual?: string, expected?: string): boolean {
    return (
      (actual ?? "").trim().toLowerCase() ===
      (expected ?? "").trim().toLowerCase()
    );
  }

  isVigentStatus(estado?: string): boolean {
    const e = (estado ?? "").trim().toLowerCase();
    return e === "activa" || e === "pendiente";
  }

  isHistoryStatus(estado?: string): boolean {
    const e = (estado ?? "").trim().toLowerCase();
    return e === "vencida" || e === "cerrada" || e === "rechazada";
  }

  isFollowUpStatus(estado?: string): boolean {
    const e = (estado ?? "").trim().toLowerCase();
    return e === "vencida" || e === "cerrada";
  }

  isAcceptedStatus(estado?: string): boolean {
    const e = (estado ?? "").trim().toLowerCase();
    return [
      "aceptado",
      "aceptada",
      "accepted",
      "aprobada",
      "aprobado",
      "seleccionado",
      "contratado",
    ].includes(e);
  }

  isRejectedStatus(estado?: string): boolean {
    const e = (estado ?? "").trim().toLowerCase();
    return e === "rechazado" || e === "rechazada";
  }

  offerStatusSeverity(
    estado?: string,
  ): "success" | "warn" | "secondary" | "danger" {
    const e = (estado ?? "").trim().toLowerCase();
    if (e === "activa") return "success";
    if (e === "pendiente") return "warn";
    if (e === "cerrada") return "secondary";
    return "danger";
  }

  vacantesTexto(job: JobOffer): string {
    const v = Number(job.vacantes ?? 0);
    if (!Number.isFinite(v) || v < 1) return "Vacantes sin especificar";
    return v === 1 ? "1 vacante" : `${v} vacantes`;
  }

  loadParticipants() {
    this.usersService.listByRole("role_participant").subscribe({
      next: (users) => (this.participants = users),
      error: () => console.warn("No se pudieron cargar los participantes"),
    });
  }

  loadMyApplications() {
    this.jobOfferService.getMyApplications().subscribe({
      next: (apps) => {
        this.appliedJobs.clear();
        this.myApplicationsMap.clear();
        (apps || []).forEach((app) => {
          this.appliedJobs.add(app.oferta_id);
          this.myApplicationsMap.set(app.oferta_id, app.estado || "Pendiente");
        });
        this.filterJobs();
      },
      error: () =>
        console.warn("No se pudieron cargar las postulaciones del usuario."),
    });
  }

  /** Waits until typing pauses before requesting filtered offers from labor-integration. */
  buscar() {
    clearTimeout(this.busquedaTimer);
    this.busquedaTimer = setTimeout(() => this.loadJobs(), 300);
  }

  /** True while offers are loading from the backend. */
  cargando = true;

  loadJobs() {
    this.cargando = true;
    this.jobOfferService.getOffers(this.searchTerm).subscribe({
      next: (data) => {
        this.cargando = false;
        this.jobs = data || [];
        this.applicationsByOffer.clear();
        this.applicationsErrorFor.clear();
        this.filterJobs();
        if (
          this.activeSection === "historial" &&
          (this.isCompany || this.isAdminOrVolunteer)
        ) {
          this.loadHistoryApplications();
        }
      },
      error: () => {
        this.cargando = false;
        this.messageService.add({
          severity: "error",
          summary: "Error",
          detail: "No se pudieron cargar las ofertas.",
        });
      },
    });
  }

  loadHistoryApplications(forceOfferId?: string) {
    if (!this.isCompany && !this.isAdminOrVolunteer) return;
    const targets = this.jobs.filter((j) => {
      if (!j.id || !this.isFollowUpStatus(j.estado)) return false;
      if (forceOfferId) return j.id === forceOfferId;
      return (
        !this.applicationsByOffer.has(j.id) &&
        !this.loadingApplicationsFor.has(j.id)
      );
    });

    targets.forEach((job) => {
      const offerId = job.id!;
      this.loadingApplicationsFor.add(offerId);
      this.applicationsErrorFor.delete(offerId);
      this.jobOfferService.getApplicationsForOffer(offerId).subscribe({
        next: (apps) => {
          this.loadingApplicationsFor.delete(offerId);
          this.applicationsByOffer.set(offerId, apps || []);
        },
        error: () => {
          this.loadingApplicationsFor.delete(offerId);
          this.applicationsErrorFor.add(offerId);
        },
      });
    });
  }

  followUpCandidates(offerId?: string): {
    accepted: JobApplication[];
    pending: JobApplication[];
  } {
    if (!offerId) return { accepted: [], pending: [] };
    const apps = this.applicationsByOffer.get(offerId) ?? [];
    const accepted: JobApplication[] = [];
    const pending: JobApplication[] = [];
    for (const app of apps) {
      if (this.isAcceptedStatus(app.estado)) {
        accepted.push(app);
      } else if (!this.isRejectedStatus(app.estado)) {
        pending.push(app);
      }
    }
    return { accepted, pending };
  }

  applicantDisplayName(app: JobApplication): string {
    return (
      app.nombre?.trim() ||
      `Participante (${app.usuario_id || app.participante_id})`
    );
  }

  private followUpNamesSummary(job: JobOffer): string {
    const { accepted, pending } = this.followUpCandidates(job.id);
    const parts: string[] = [];
    if (accepted.length) {
      parts.push(
        `Aceptados: ${accepted.map((a) => this.applicantDisplayName(a)).join(", ")}`,
      );
    }
    if (pending.length) {
      parts.push(
        `Pendientes sin resolución: ${pending.map((a) => this.applicantDisplayName(a)).join(", ")}`,
      );
    }
    return parts.join("\n");
  }

  companyContactMailto(job: JobOffer): string {
    if (!job.empresa_email) return "";
    const subject = `Seguimiento de oferta "${job.titulo}" (${job.estado})`;
    const summary = this.followUpNamesSummary(job);
    const body = summary
      ? `Hola,\n\nNos comunicamos desde el Centro Buen Vivir por la oferta "${job.titulo}" (${job.estado}) para consultar sobre:\n${summary}\n\nMuchas gracias.`
      : `Hola,\n\nNos comunicamos desde el Centro Buen Vivir por la oferta "${job.titulo}" (${job.estado}).\n\nMuchas gracias.`;
    return `mailto:${encodeURIComponent(job.empresa_email)}?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`;
  }

  centerContactQueryParams(job: JobOffer): { asunto: string; mensaje: string } {
    const summary = this.followUpNamesSummary(job);
    const base = `Consulta sobre la oferta "${job.titulo}" (estado: ${job.estado}).`;
    const mensaje = summary
      ? `${base}\nPostulantes a consultar:\n${summary}`
      : `${base} Quisiera coordinar el seguimiento con los responsables.`;
    return {
      asunto: "empleo",
      mensaje: mensaje.slice(0, 3000),
    };
  }

  participantContactQueryParams(job: JobOffer): {
    asunto: string;
    mensaje: string;
  } {
    const appStatus = job.id
      ? this.myApplicationsMap.get(job.id) || "Pendiente"
      : "Pendiente";
    return {
      asunto: "empleo",
      mensaje:
        `Hola, quiero consultar por mi postulación (${appStatus}) a la oferta "${job.titulo}" (${job.estado}).`.slice(
          0,
          3000,
        ),
    };
  }

  filterJobs() {
    let filtered = [...this.jobs];

    if (this.selectedEmploymentType) {
      filtered = filtered.filter(
        (j) => j.tipo_empleo === this.selectedEmploymentType,
      );
    }

    if (this.isParticipant) {
      if (this.currentTab === "Disponibles") {
        filtered = filtered.filter(
          (j) =>
            this.isEstado(j.estado, "Activa") && !this.appliedJobs.has(j.id!),
        );
      } else if (this.currentTab === "Postuladas") {
        filtered = filtered.filter((j) => this.appliedJobs.has(j.id!));
      }
    } else if (this.isCompany || this.isAdminOrVolunteer) {
      if (this.activeSection === "historial") {
        filtered = filtered.filter((j) => this.isHistoryStatus(j.estado));
        if (this.historyTab !== "Todas") {
          filtered = filtered.filter((j) =>
            this.isEstado(j.estado, this.historyTab),
          );
        }
      } else {
        filtered = filtered.filter((j) => this.isVigentStatus(j.estado));
        const activeSubTab = this.isCompany ? this.companyTab : this.adminTab;
        if (activeSubTab !== "Todas") {
          filtered = filtered.filter((j) =>
            this.isEstado(j.estado, activeSubTab),
          );
        }
      }
    }

    this.filteredJobs = filtered;
  }

  setSection(section: "vigentes" | "historial") {
    this.activeSection = section;
    void this.router.navigate([], {
      relativeTo: this.route,
      queryParams: { seccion: section === "historial" ? "historial" : null },
      queryParamsHandling: "merge",
      replaceUrl: true,
    });
    this.filterJobs();
    if (section === "historial") {
      this.loadHistoryApplications();
    }
  }

  setTab(tab: string) {
    if (this.isParticipant) this.currentTab = tab;
    else if (this.isCompany) this.companyTab = tab;
    else if (this.isAdminOrVolunteer) this.adminTab = tab;
    this.filterJobs();
  }

  setHistoryTab(tab: string) {
    this.historyTab = tab;
    this.filterJobs();
  }

  openNewJobDialog() {
    this.isEditing = false;
    this.newJob = {
      empresa_id: "",
      creador_nombre: this.currentUserName,
      titulo: "",
      descripcion: "",
      ubicacion: "",
      modalidad: "Presencial",
      tipo_empleo: "Tiempo Completo",
      salario: "",
      horario: "",
      vacantes: 1,
      requisitos: [],
      fecha_limite: "",
      estado: this.isAdminOrVolunteer ? "Activa" : "Pendiente",
    };
    this.currentRequisito = "";
    this.intentoGuardar = false;
    this.showDialog = true;
    if (this.isCompany) {
      this.messageService.add({
        severity: "info",
        summary: "Aviso",
        detail: "La oferta quedará en estado Pendiente hasta ser aprobada.",
      });
    }
  }

  openEditJobDialog(job: JobOffer) {
    this.isEditing = true;
    this.newJob = {
      ...job,
      vacantes: job.vacantes && job.vacantes > 0 ? job.vacantes : 1,
      requisitos: [...(job.requisitos || [])],
    };
    if (this.newJob.fecha_limite?.includes("T")) {
      const date = new Date(this.newJob.fecha_limite);
      const utc3Date = new Date(date.getTime() - 3 * 60 * 60 * 1000);
      this.newJob.fecha_limite = utc3Date.toISOString().split("T")[0];
    }
    this.currentRequisito = "";
    this.intentoGuardar = false;
    this.showDialog = true;
  }

  addRequisito() {
    if (this.currentRequisito.trim()) {
      if (!this.newJob.requisitos) this.newJob.requisitos = [];
      this.newJob.requisitos.push(this.currentRequisito.trim());
      this.currentRequisito = "";
    }
  }

  removeRequisito(index: number) {
    this.newJob.requisitos.splice(index, 1);
  }

  openRenewDialog(job: JobOffer) {
    this.selectedJobForRenew = job;
    this.renewDate = "";
    this.renewVacancies = job.vacantes && job.vacantes > 0 ? job.vacantes : 1;
    this.showRenewDialog = true;
  }

  renewJob() {
    if (!this.renewDate || !this.selectedJobForRenew?.id) {
      this.messageService.add({
        severity: "warn",
        summary: "Aviso",
        detail: "La nueva fecha es obligatoria.",
      });
      return;
    }
    if (this.renewDate < hoyYMD()) {
      this.messageService.add({
        severity: "warn",
        summary: "Aviso",
        detail: "La nueva fecha límite debe ser hoy o posterior.",
      });
      return;
    }
    const vacantes = Math.floor(Number(this.renewVacancies));
    if (!Number.isFinite(vacantes) || vacantes < 1) {
      this.messageService.add({
        severity: "warn",
        summary: "Aviso",
        detail: "Indicá al menos 1 vacante.",
      });
      return;
    }

    const updatedJob = {
      ...this.selectedJobForRenew,
      vacantes,
      estado: "Activa",
      fecha_limite: `${this.renewDate}T23:59:59-03:00`,
    } as JobOffer;
    this.jobOfferService
      .updateOffer(updatedJob.id as string, updatedJob)
      .subscribe({
        next: () => {
          this.messageService.add({
            severity: "success",
            summary: "Éxito",
            detail: "Oferta renovada correctamente.",
          });
          this.showRenewDialog = false;
          this.loadJobs();
        },
        error: () =>
          this.messageService.add({
            severity: "error",
            summary: "Error",
            detail:
              "Hubo un problema al renovar la oferta. Si le faltan datos obligatorios (ubicación, salario, horario, requisitos…), editala y completalos.",
          }),
      });
  }

  confirmCloseOffer(job: JobOffer) {
    if (!job.id) return;
    this.confirmationService.confirm({
      message:
        "¿Querés cerrar esta oferta antes de su fecha límite? Dejará de recibir postulaciones y pasará a Finalizadas e historial.",
      header: "Confirmar cierre anticipado",
      icon: "pi pi-lock",
      acceptLabel: "Sí, cerrar oferta",
      rejectLabel: "Cancelar",
      accept: () => {
        if (!job.id) return;
        this.jobOfferService.closeOffer(job.id).subscribe({
          next: (closed) => {
            this.messageService.add({
              severity: "success",
              summary: "Éxito",
              detail: "Oferta cerrada correctamente.",
            });
            if (this.currentJob?.id === job.id) {
              this.currentJob = {
                ...this.currentJob,
                ...closed,
                estado: "Cerrada",
              };
            }
            this.loadJobs();
          },
          error: (err) =>
            this.messageService.add({
              severity: "error",
              summary: "Error",
              detail: mensajeErrorOferta(
                err,
                "Hubo un problema al cerrar la oferta.",
              ),
            }),
        });
      },
    });
  }

  /**
   * True after the first save attempt so empty required fields highlight inline
   * without showing red errors as soon as the dialog opens.
   */
  intentoGuardar = false;

  /** Field-level errors matching backend offer validation. */
  get errores(): ErroresOferta {
    return this.intentoGuardar
      ? erroresOferta(this.newJob, { exigirFechaFutura: !this.isEditing })
      : {};
  }

  saveJob() {
    if (this.currentRequisito.trim()) {
      this.addRequisito();
    }
    this.intentoGuardar = true;
    const errores = erroresOferta(this.newJob, {
      exigirFechaFutura: !this.isEditing,
    });
    const faltan = Object.values(errores);
    if (faltan.length) {
      this.messageService.add({
        severity: "warn",
        summary: "Revisá los campos marcados",
        detail: faltan.join(" "),
      });
      return;
    }

    const jobToSave: JobOffer = {
      ...this.newJob,
      vacantes: Math.floor(Number(this.newJob.vacantes)),
    };
    if (jobToSave.fecha_limite && !jobToSave.fecha_limite.includes("T")) {
      jobToSave.fecha_limite = `${jobToSave.fecha_limite}T23:59:59-03:00`;
    } else if (!jobToSave.fecha_limite) {
      delete jobToSave.fecha_limite;
    }

    if (this.isEditing && jobToSave.id) {
      this.jobOfferService.updateOffer(jobToSave.id, jobToSave).subscribe({
        next: () => {
          this.messageService.add({
            severity: "success",
            summary: "Éxito",
            detail: "Oferta actualizada correctamente.",
          });
          this.showDialog = false;
          this.loadJobs();
        },
        error: (err) =>
          this.messageService.add({
            severity: "error",
            summary: "Error",
            detail: mensajeErrorOferta(
              err,
              "Hubo un problema al actualizar la oferta.",
            ),
          }),
      });
    } else {
      this.jobOfferService.createOffer(jobToSave).subscribe({
        next: () => {
          this.messageService.add({
            severity: "success",
            summary: "Éxito",
            detail: "Oferta creada correctamente.",
          });
          this.showDialog = false;
          this.loadJobs();
        },
        error: (err) =>
          this.messageService.add({
            severity: "error",
            summary: "Error",
            detail: mensajeErrorOferta(
              err,
              "Hubo un problema al crear la oferta.",
            ),
          }),
      });
    }
  }

  confirmDelete(job: JobOffer) {
    this.confirmationService.confirm({
      message:
        "¿Querés eliminar esta oferta laboral? Esta acción no se puede deshacer.",
      header: "Confirmar eliminación",
      icon: "pi-exclamation-triangle",
      acceptLabel: "Sí",
      rejectLabel: "No",
      accept: () => {
        if (job.id) {
          this.jobOfferService.deleteOffer(job.id).subscribe({
            next: () => {
              this.messageService.add({
                severity: "success",
                summary: "Éxito",
                detail: "Oferta eliminada correctamente.",
              });
              this.loadJobs();
            },
            error: () =>
              this.messageService.add({
                severity: "error",
                summary: "Error",
                detail: "Hubo un problema al eliminar la oferta.",
              }),
          });
        }
      },
    });
  }

  applyToJob(job: JobOffer) {
    if (!job.id || !this.userId) return;

    this.confirmationService.confirm({
      message:
        "¿Querés postularte a esta oferta laboral? Antes, revisá que tu CV esté actualizado en Mi CV.",
      header: "Confirmar postulación",
      icon: "pi pi-check-circle",
      acceptLabel: "Sí, postularme",
      rejectLabel: "Cancelar",
      accept: () => {
        this.jobOfferService
          .applyToOffer({ oferta_id: job.id, participante_id: this.userId })
          .subscribe({
            next: () => {
              this.messageService.add({
                severity: "success",
                summary: "Éxito",
                detail: "Te has postulado correctamente.",
              });
              this.appliedJobs.add(job.id as string);
              this.myApplicationsMap.set(job.id as string, "Pendiente");
              this.filterJobs();
            },
            error: (err) => {
              if (err.status === 409) {
                this.messageService.add({
                  severity: "warn",
                  summary: "Aviso",
                  detail: "Ya te encuentras postulado a esta oferta.",
                });
                this.appliedJobs.add(job.id as string);
                this.myApplicationsMap.set(job.id as string, "Pendiente");
                this.filterJobs();
              } else {
                this.messageService.add({
                  severity: "error",
                  summary: "Error",
                  detail: mensajeErrorOferta(
                    err,
                    "Hubo un problema al postularte.",
                  ),
                });
              }
            },
          });
      },
    });
  }

  openApplyDialog(job: JobOffer) {
    this.jobToApply = job;
    this.selectedParticipant = "";
    this.showApplyDialog = true;
  }

  submitAdminApplication() {
    if (!this.jobToApply?.id || !this.selectedParticipant) {
      this.messageService.add({
        severity: "warn",
        summary: "Aviso",
        detail: "Seleccione un participante.",
      });
      return;
    }

    this.jobOfferService
      .applyToOffer({
        oferta_id: this.jobToApply.id,
        participante_id: this.selectedParticipant,
      })
      .subscribe({
        next: () => {
          this.messageService.add({
            severity: "success",
            summary: "Éxito",
            detail: "Participante postulado correctamente.",
          });
          this.showApplyDialog = false;
        },
        error: (err) => {
          if (err.status === 409) {
            this.messageService.add({
              severity: "warn",
              summary: "Aviso",
              detail: "El participante ya está postulado.",
            });
          } else {
            this.messageService.add({
              severity: "error",
              summary: "Error",
              detail: mensajeErrorOferta(err, "Hubo un problema al postular."),
            });
          }
        },
      });
  }

  showApplications(job: JobOffer) {
    if (!job.id) return;
    this.currentJob = job;
    this.currentJobTitle = job.titulo;
    this.currentCandidates = [];
    this.applicationsDialogVisible = true;

    this.jobOfferService.getApplicationsForOffer(job.id).subscribe({
      next: (apps) => {
        const applications = apps || [];
        this.applicationsByOffer.set(job.id!, applications);
        applications.forEach((app) => {
          const candidateData = {
            app,
            resume: undefined as Resume | undefined,
          };
          this.currentCandidates.push(candidateData);

          this.rxResumeService.getPublicResume(this.cvId(app)).subscribe({
            next: (resume) => {
              candidateData.resume = resume;
            },
            error: () =>
              console.warn(`No se pudo cargar el CV para ${this.cvId(app)}`),
          });
        });
      },
      error: () =>
        this.messageService.add({
          severity: "error",
          summary: "Error",
          detail: "No se pudieron cargar los postulantes.",
        }),
    });
  }

  /** Applicant name: from labor-integration or CV basics fallback. */
  candidateName(candidate: { app: JobApplication; resume?: Resume }): string {
    return (
      candidate.app.nombre?.trim() ||
      candidate.resume?.data?.basics?.name?.trim() ||
      "Participante sin nombre cargado"
    );
  }

  /** Display name for the "Postular participante" selector. */
  participantName(p: User): string {
    const nombre = `${p.first_name ?? ""} ${p.last_name ?? ""}`.trim();
    return nombre || p.email;
  }

  /** CV documents are indexed by user_id. */
  cvId(app: JobApplication): string {
    return app.usuario_id || app.participante_id;
  }

  updateApplicationStatus(
    candidate: { app: JobApplication; resume?: Resume },
    status: string,
  ) {
    if (!candidate.app.id) return;
    this.jobOfferService
      .updateApplicationStatus(candidate.app.id, status)
      .subscribe({
        next: () => {
          candidate.app.estado = status;
          if (this.currentJob?.id) {
            const cached = this.applicationsByOffer.get(this.currentJob.id);
            if (cached) {
              const target = cached.find((a) => a.id === candidate.app.id);
              if (target) target.estado = status;
            }
          }
          if (this.isCompany && status === "Aceptado") {
            const participantName = this.candidateName(candidate);
            this.acceptanceMessage = `Has aceptado a ${participantName}. Comunicate con los responsables del centro para coordinar el ingreso.`;
            this.acceptedOfferForContact = this.currentJob;
            this.showAcceptanceModal = true;
          } else {
            this.messageService.add({
              severity: "success",
              summary: "Éxito",
              detail: `Postulación marcada como ${status}`,
            });
          }
        },
        error: () =>
          this.messageService.add({
            severity: "error",
            summary: "Error",
            detail: "Hubo un problema al actualizar el estado.",
          }),
      });
  }

  approveJob(job: JobOffer) {
    if (!job.id) return;
    this.jobOfferService.approveOffer(job.id).subscribe({
      next: () => {
        this.messageService.add({
          severity: "success",
          summary: "Éxito",
          detail: "Oferta aprobada.",
        });
        this.loadJobs();
      },
      error: () =>
        this.messageService.add({
          severity: "error",
          summary: "Error",
          detail: "Hubo un problema al aprobar.",
        }),
    });
  }

  rejectJob(job: JobOffer) {
    if (!job.id) return;
    this.jobOfferService.rejectOffer(job.id).subscribe({
      next: () => {
        this.messageService.add({
          severity: "success",
          summary: "Éxito",
          detail: "Oferta rechazada.",
        });
        this.loadJobs();
      },
      error: () =>
        this.messageService.add({
          severity: "error",
          summary: "Error",
          detail: "Hubo un problema al rechazar.",
        }),
    });
  }
}

/**
 * HU-39: required offer fields (mirrors backend validation).
 */
export function camposFaltantesOferta(job: {
  titulo?: string;
  descripcion?: string;
  ubicacion?: string;
  modalidad?: string;
  tipo_empleo?: string;
  salario?: string;
  horario?: string;
  vacantes?: number;
  requisitos?: string[];
  fecha_limite?: string;
}): string[] {
  const vacio = (v?: string) => !v || !String(v).trim();
  const faltan: string[] = [];
  if (vacio(job.titulo)) faltan.push("título");
  if (vacio(job.descripcion)) faltan.push("descripción");
  if (vacio(job.ubicacion)) faltan.push("ubicación");
  if (vacio(job.modalidad)) faltan.push("modalidad");
  if (vacio(job.tipo_empleo)) faltan.push("tipo de empleo");
  if (vacio(job.salario)) faltan.push("salario");
  if (vacio(job.horario)) faltan.push("horario");
  const vacantes = Number(job.vacantes);
  if (!Number.isInteger(vacantes) || vacantes < 1) {
    faltan.push("vacantes (al menos 1)");
  }
  if (!(job.requisitos ?? []).some((r) => !vacio(r)))
    faltan.push("al menos un requisito");
  if (vacio(job.fecha_limite)) faltan.push("fecha límite");
  return faltan;
}

/** Form field keys that can carry inline validation messages. */
export type CampoOferta =
  | "titulo"
  | "descripcion"
  | "ubicacion"
  | "modalidad"
  | "tipo_empleo"
  | "salario"
  | "horario"
  | "vacantes"
  | "requisitos"
  | "fecha_limite";

export type ErroresOferta = Partial<Record<CampoOferta, string>>;

/** Today's date (YYYY-MM-DD) in UTC-3 for date input comparison. */
function hoyYMD(): string {
  return new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString().slice(0, 10);
}

/**
 * Per-field validation matching labor-integration `validateOfferFields`.
 */
export function erroresOferta(
  job: Parameters<typeof camposFaltantesOferta>[0],
  opciones: { exigirFechaFutura?: boolean } = {},
): ErroresOferta {
  const vacio = (v?: string) => !v || !String(v).trim();
  const e: ErroresOferta = {};
  if (vacio(job.titulo)) e.titulo = "Ingresá el título del puesto.";
  if (vacio(job.ubicacion)) e.ubicacion = "Ingresá la ubicación.";
  if (vacio(job.tipo_empleo)) e.tipo_empleo = "Elegí el tipo de empleo.";
  if (vacio(job.modalidad)) e.modalidad = "Elegí la modalidad.";
  if (vacio(job.salario)) e.salario = 'Ingresá el salario (o "A convenir").';
  if (vacio(job.horario)) e.horario = "Ingresá el horario.";
  const vacantes = Number(job.vacantes);
  if (!Number.isInteger(vacantes) || vacantes < 1) {
    e.vacantes = "Indicá al menos 1 vacante.";
  }
  if (vacio(job.fecha_limite)) {
    e.fecha_limite = "Elegí la fecha límite.";
  } else if (
    opciones.exigirFechaFutura &&
    String(job.fecha_limite).slice(0, 10) < hoyYMD()
  ) {
    e.fecha_limite = "La fecha límite ya pasó.";
  }
  if (vacio(job.descripcion)) e.descripcion = "Ingresá la descripción.";
  if (!(job.requisitos ?? []).some((r) => !vacio(r)))
    e.requisitos = "Agregá al menos un requisito.";
  return e;
}

/** Extracts user-facing 400 error message from backend response when available. */
export function mensajeErrorOferta(err: unknown, fallback: string): string {
  const e = err as { status?: number; error?: unknown };
  if (e?.status === 400 && typeof e.error === "string" && e.error.trim()) {
    return e.error.trim();
  }
  return fallback;
}
