import { Component, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
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

@Component({
  selector: "app-ofertas",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
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
  searchTerm: string = "";
  private busquedaTimer?: ReturnType<typeof setTimeout>;
  showDialog = false;
  isEditing = false;
  currentRequisito = "";

  // Variables para la renovación de ofertas
  showRenewDialog = false;
  renewDate = "";
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
    requisitos: [],
    fecha_limite: "",
    estado: "Activa",
  };

  jobs: JobOffer[] = [];
  filteredJobs: JobOffer[] = [];
  isCompany = false;
  isParticipant = false;
  isAdminOrVolunteer = false;
  /** Aprobar y rechazar ofertas es solo del administrador (HU-41). */
  isAdmin = false;
  userId = "";
  currentUserName = "";

  // Tabs for filtering
  currentTab = "Disponibles"; // For participants
  companyTab = "Todas"; // For companies
  adminTab = "Pendiente"; // For admin/volunteer (mismo valor que el estado de la oferta)

  employmentTypes = [
    { label: "Todos", value: null },
    { label: "Tiempo completo", value: "Tiempo Completo" },
    { label: "Medio tiempo", value: "Medio Tiempo" },
    { label: "Pasantía", value: "Pasantía" },
  ];
  selectedEmploymentType: string | null = null;

  applicationsDialogVisible = false;
  currentJobTitle = "";
  currentCandidates: { app: JobApplication; resume?: Resume }[] = [];
  appliedJobs: Set<string> = new Set();
  myApplicationsMap: Map<string, string> = new Map(); // oferta_id -> estado

  // Admin Apply
  participants: User[] = [];
  selectedParticipant: string = "";
  showApplyDialog = false;
  jobToApply: JobOffer | null = null;

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
    this.authService.isAuthenticated().subscribe({
      next: (res) => {
        const role = res?.role_id ?? res?.user?.role_id;
        if (role) {
          this.isCompany = role === "role_company";
          this.isParticipant = role === "role_participant";
          this.isAdmin = role === "role_admin" || role === "role_dev";
          this.isAdminOrVolunteer = this.isAdmin || role === "role_volunteer";
        }
        const r = res as any;
        const uid = r.user_id || res?.user?.user_id;
        if (uid) {
          this.userId = uid;
        }

        if (this.isParticipant) {
          this.loadMyApplications();
        }
        if (this.isAdminOrVolunteer) {
          this.loadParticipants();
        }
        this.filterJobs();

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

  loadParticipants() {
    this.usersService.listByRole("role_participant").subscribe({
      next: (users) => (this.participants = users),
      error: () => console.warn("No se pudieron cargar los participantes"),
    });
  }

  loadMyApplications() {
    this.jobOfferService.getMyApplications().subscribe({
      next: (apps) => {
        apps.forEach((app) => {
          this.appliedJobs.add(app.oferta_id);
          this.myApplicationsMap.set(app.oferta_id, app.estado || "Pendiente");
        });
      },
      error: () =>
        console.warn("No se pudieron cargar las postulaciones del usuario."),
    });
  }

  /** Espera a que se deje de tipear y pide las ofertas con el texto a labor-integration. */
  buscar() {
    clearTimeout(this.busquedaTimer);
    this.busquedaTimer = setTimeout(() => this.loadJobs(), 300);
  }

  loadJobs() {
    this.jobOfferService.getOffers(this.searchTerm).subscribe({
      next: (data) => {
        this.jobs = data || [];
        this.filterJobs();
      },
      error: (err) => {
        this.messageService.add({
          severity: "error",
          summary: "Error",
          detail: "No se pudieron cargar las ofertas.",
        });
      },
    });
  }

  filterJobs() {
    let filtered = [...this.jobs];

    if (this.selectedEmploymentType) {
      filtered = filtered.filter(
        (j) => j.tipo_empleo === this.selectedEmploymentType,
      );
    }

    // Tab filters
    if (this.isParticipant) {
      if (this.currentTab === "Disponibles") {
        filtered = filtered.filter((j) => !this.appliedJobs.has(j.id!));
      } else if (this.currentTab === "Postuladas") {
        filtered = filtered.filter((j) => this.appliedJobs.has(j.id!));
      }
    } else if (this.isCompany) {
      if (this.companyTab !== "Todas") {
        filtered = filtered.filter(
          (j) => j.estado?.toLowerCase() === this.companyTab.toLowerCase(),
        );
      }
    } else if (this.isAdminOrVolunteer) {
      if (this.adminTab !== "Todas") {
        filtered = filtered.filter(
          (j) => j.estado?.toLowerCase() === this.adminTab.toLowerCase(),
        );
      }
    }

    this.filteredJobs = filtered;
  }

  setTab(tab: string) {
    if (this.isParticipant) this.currentTab = tab;
    else if (this.isCompany) this.companyTab = tab;
    else if (this.isAdminOrVolunteer) this.adminTab = tab;
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
      requisitos: [],
      fecha_limite: "",
      estado: "Pendiente",
    };
    this.currentRequisito = "";
    this.intentoGuardar = false;
    this.showDialog = true;
    this.messageService.add({
      severity: "info",
      summary: "Aviso",
      detail: "La oferta quedará en estado Pendiente hasta ser aprobada.",
    });
  }

  openEditJobDialog(job: JobOffer) {
    this.isEditing = true;
    this.newJob = { ...job, requisitos: [...(job.requisitos || [])] };
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

    const updatedJob = {
      ...this.selectedJobForRenew,
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

  /**
   * true después del primer intento de guardar: desde ahí cada campo
   * obligatorio vacío muestra su error al lado (no antes, para no marcar
   * en rojo un formulario recién abierto).
   */
  intentoGuardar = false;

  /** Errores por campo del formulario de oferta (mismos criterios que el backend). */
  get errores(): ErroresOferta {
    return this.intentoGuardar
      ? erroresOferta(this.newJob, { exigirFechaFutura: !this.isEditing })
      : {};
  }

  saveJob() {
    // Un requisito escrito pero sin confirmar con Enter o "+" también cuenta.
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

    const jobToSave = { ...this.newJob };
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
              } else {
                this.messageService.add({
                  severity: "error",
                  summary: "Error",
                  detail: "Hubo un problema al postularte.",
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
              detail: "Hubo un problema al postular.",
            });
          }
        },
      });
  }

  showApplications(job: JobOffer) {
    if (!job.id) return;
    this.currentJobTitle = job.titulo;
    this.currentCandidates = [];
    this.applicationsDialogVisible = true;

    this.jobOfferService.getApplicationsForOffer(job.id).subscribe({
      next: (apps) => {
        const applications = apps || [];
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

  showAcceptanceModal = false;
  acceptanceMessage = "";

  /** El CV está guardado con el id de usuario del postulante. */
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
          if (this.isCompany && status === "Aceptado") {
            const participantName =
              candidate.resume?.data?.basics?.name || "este participante";
            this.acceptanceMessage = `Has aceptado a ${participantName}. Ponte en contacto con los responsables para continuar.`;
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
 * HU-39: campos obligatorios de una oferta (el backend valida lo mismo).
 * Devuelve los nombres de los que faltan, en el orden del formulario.
 */
export function camposFaltantesOferta(job: {
  titulo?: string;
  descripcion?: string;
  ubicacion?: string;
  modalidad?: string;
  tipo_empleo?: string;
  salario?: string;
  horario?: string;
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
  if (!(job.requisitos ?? []).some((r) => !vacio(r)))
    faltan.push("al menos un requisito");
  if (vacio(job.fecha_limite)) faltan.push("fecha límite");
  return faltan;
}

/** Campos del formulario de oferta que pueden tener error. */
export type CampoOferta =
  | "titulo"
  | "descripcion"
  | "ubicacion"
  | "modalidad"
  | "tipo_empleo"
  | "salario"
  | "horario"
  | "requisitos"
  | "fecha_limite";

export type ErroresOferta = Partial<Record<CampoOferta, string>>;

/** Fecha de hoy (YYYY-MM-DD) en Argentina, para comparar con el input date. */
function hoyYMD(): string {
  return new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString().slice(0, 10);
}

/**
 * Errores por campo de una oferta, con los mismos criterios que el backend
 * (labor-integration `validateOfferFields`): todos los campos obligatorios
 * y, al publicar, una fecha límite que no haya pasado. Así el formulario
 * no deja enviar algo que el backend va a rechazar.
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

/** Usa el texto que devuelve el backend (400 con el campo faltante). */
export function mensajeErrorOferta(err: unknown, fallback: string): string {
  const e = err as { status?: number; error?: unknown };
  if (e?.status === 400 && typeof e.error === "string" && e.error.trim()) {
    return e.error.trim();
  }
  return fallback;
}
