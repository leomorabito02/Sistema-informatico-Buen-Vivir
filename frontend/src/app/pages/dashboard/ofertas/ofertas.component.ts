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
import { JobOfferService, JobOffer, JobApplication } from "../../../services/job-offer.service";
import { AuthService } from "../../../services/auth.service";
import { RxResumeService, Resume } from "../../../services/rxresume.service";

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
    DropdownModule
  ],
  providers: [ConfirmationService, MessageService],
  templateUrl: "./ofertas.component.html",
  styleUrls: ["./ofertas.component.css"],
})
export class OfertasComponent implements OnInit {
  searchTerm: string = "";
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
  userId = "";

  employmentTypes = [
    { label: 'Todos', value: null },
    { label: 'Tiempo completo', value: 'Tiempo Completo' },
    { label: 'Medio tiempo', value: 'Medio Tiempo' },
    { label: 'Pasantía', value: 'Pasantía' }
  ];
  selectedEmploymentType: string | null = null;

  applicationsDialogVisible = false;
  currentJobTitle = "";
  currentCandidates: { app: JobApplication, resume?: Resume }[] = [];
  appliedJobs: Set<string> = new Set();
  myApplicationsMap: Map<string, string> = new Map(); // oferta_id -> estado


  constructor(
    private readonly jobOfferService: JobOfferService,
    private readonly authService: AuthService,
    private readonly rxResumeService: RxResumeService,
    private readonly confirmationService: ConfirmationService,
    private readonly messageService: MessageService
  ) {}

  ngOnInit() {
    this.authService.isAuthenticated().subscribe({
      next: (res) => {
        if (res?.role_id) {
          this.isCompany = res.role_id === "role_company";
        }
        const r = res as any;
        const uid = r.user_id || res?.user?.user_id;
        if (uid) {
          this.userId = uid;
        }
        
        if (!this.isCompany) {
          this.loadMyApplications();
        }
      }
    });
    this.loadJobs();
  }

  loadMyApplications() {
    this.jobOfferService.getMyApplications().subscribe({
      next: (apps) => {
        apps.forEach(app => {
          this.appliedJobs.add(app.oferta_id);
          this.myApplicationsMap.set(app.oferta_id, app.estado || 'Pendiente');
        });
      },
      error: () => console.warn('No se pudieron cargar las postulaciones del usuario.')
    });
  }

  loadJobs() {
    this.jobOfferService.getOffers().subscribe({
      next: (data) => {
        this.jobs = data || [];
        this.filterJobs();
      },
      error: (err) => {
        this.messageService.add({severity:'error', summary:'Error', detail:'No se pudieron cargar las ofertas.'});
      }
    });
  }

  filterJobs() {
    let filtered = [...this.jobs];
    
    if (this.selectedEmploymentType) {
      filtered = filtered.filter(j => j.tipo_empleo === this.selectedEmploymentType);
    }

    const term = this.searchTerm.toLowerCase().trim();
    if (term) {
      filtered = filtered.filter(
        (j) =>
          j.titulo.toLowerCase().includes(term) ||
          j.ubicacion.toLowerCase().includes(term) ||
          j.tipo_empleo.toLowerCase().includes(term)
      );
    }
    
    this.filteredJobs = filtered;
  }

  openNewJobDialog() {
    this.isEditing = false;
    this.newJob = {
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
    this.currentRequisito = "";
    this.showDialog = true;
  }

  openEditJobDialog(job: JobOffer) {
    this.isEditing = true;
    this.newJob = { ...job, requisitos: [...(job.requisitos || [])] };
    if (this.newJob.fecha_limite?.includes('T')) {
      const date = new Date(this.newJob.fecha_limite);
      const utc3Date = new Date(date.getTime() - (3 * 60 * 60 * 1000));
      this.newJob.fecha_limite = utc3Date.toISOString().split('T')[0];
    }
    this.currentRequisito = "";
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
      this.messageService.add({severity:'warn', summary:'Aviso', detail:'La nueva fecha es obligatoria.'});
      return;
    }

    const updatedJob = { ...this.selectedJobForRenew, estado: 'Activa', fecha_limite: `${this.renewDate}T23:59:59-03:00` } as JobOffer;
    this.jobOfferService.updateOffer(updatedJob.id as string, updatedJob).subscribe({
      next: () => {
        this.messageService.add({severity:'success', summary:'Éxito', detail:'Oferta renovada correctamente.'});
        this.showRenewDialog = false;
        this.loadJobs();
      },
      error: () => this.messageService.add({severity:'error', summary:'Error', detail:'Hubo un problema al renovar la oferta.'})
    });
  }

  saveJob() {
    if (!this.newJob.titulo) {
      this.messageService.add({severity:'warn', summary:'Aviso', detail:'El título es obligatorio.'});
      return;
    }

    const jobToSave = { ...this.newJob };
    if (jobToSave.fecha_limite && !jobToSave.fecha_limite.includes('T')) {
      jobToSave.fecha_limite = `${jobToSave.fecha_limite}T23:59:59-03:00`;
    } else if (!jobToSave.fecha_limite) {
      delete jobToSave.fecha_limite;
    }

    if (this.isEditing && jobToSave.id) {
      this.jobOfferService.updateOffer(jobToSave.id, jobToSave).subscribe({
        next: () => {
          this.messageService.add({severity:'success', summary:'Éxito', detail:'Oferta actualizada correctamente.'});
          this.showDialog = false;
          this.loadJobs();
        },
        error: () => this.messageService.add({severity:'error', summary:'Error', detail:'Hubo un problema al actualizar la oferta.'})
      });
    } else {
      this.jobOfferService.createOffer(jobToSave).subscribe({
        next: () => {
          this.messageService.add({severity:'success', summary:'Éxito', detail:'Oferta creada correctamente.'});
          this.showDialog = false;
          this.loadJobs();
        },
        error: () => this.messageService.add({severity:'error', summary:'Error', detail:'Hubo un problema al crear la oferta.'})
      });
    }
  }

  confirmDelete(job: JobOffer) {
    this.confirmationService.confirm({
      message: '¿Estás seguro de que deseas eliminar esta oferta laboral?',
      header: 'Confirmar Eliminación',
      icon: 'pi-exclamation-triangle',
      acceptLabel: 'Sí',
      rejectLabel: 'No',
      accept: () => {
        if (job.id) {
          this.jobOfferService.deleteOffer(job.id).subscribe({
            next: () => {
              this.messageService.add({severity:'success', summary:'Éxito', detail:'Oferta eliminada correctamente.'});
              this.loadJobs();
            },
            error: () => this.messageService.add({severity:'error', summary:'Error', detail:'Hubo un problema al eliminar la oferta.'})
          });
        }
      }
    });
  }

  applyToJob(job: JobOffer) {
    if (!job.id || !this.userId) return;

    this.confirmationService.confirm({
      message: '¿Estás seguro que deseas postularte a esta oferta laboral? Asegúrate de tener tu CV actualizado en tu perfil.',
      header: 'Confirmar Postulación',
      icon: 'pi pi-check-circle',
      acceptLabel: 'Sí, postularme',
      rejectLabel: 'Cancelar',
      accept: () => {
        this.jobOfferService.applyToOffer({ oferta_id: job.id, participante_id: this.userId }).subscribe({
          next: () => {
            this.messageService.add({severity:'success', summary:'Éxito', detail:'Te has postulado correctamente.'});
            this.appliedJobs.add(job.id as string);
            this.myApplicationsMap.set(job.id as string, 'Pendiente');
          },
          error: (err) => {
            if (err.status === 409) {
              this.messageService.add({severity:'warn', summary:'Aviso', detail:'Ya te encuentras postulado a esta oferta.'});
              this.appliedJobs.add(job.id as string);
              this.myApplicationsMap.set(job.id as string, 'Pendiente');
            } else {
              this.messageService.add({severity:'error', summary:'Error', detail:'Hubo un problema al postularte.'});
            }
          }
        });
      }
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
        applications.forEach(app => {
          const candidateData = { app, resume: undefined as Resume | undefined };
          this.currentCandidates.push(candidateData);
          
          this.rxResumeService.getPublicResume(app.participante_id).subscribe({
            next: (resume) => {
              candidateData.resume = resume;
            },
            error: () => console.warn(`No se pudo cargar el CV para ${app.participante_id}`)
          });
        });
      },
      error: () => this.messageService.add({severity:'error', summary:'Error', detail:'No se pudieron cargar los postulantes.'})
    });
  }

  updateApplicationStatus(candidate: { app: JobApplication, resume?: Resume }, status: string) {
    if (!candidate.app.id) return;
    this.jobOfferService.updateApplicationStatus(candidate.app.id, status).subscribe({
      next: () => {
        candidate.app.estado = status;
        this.messageService.add({severity:'success', summary:'Éxito', detail:`Postulación marcada como ${status}`});
      },
      error: () => this.messageService.add({severity:'error', summary:'Error', detail:'Hubo un problema al actualizar el estado.'})
    });
  }
}
