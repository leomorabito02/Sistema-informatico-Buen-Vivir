import { Component, OnInit, OnDestroy, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { ActivatedRoute } from "@angular/router";
import { FormsModule, NgForm } from "@angular/forms";
import { CardModule } from "primeng/card";
import { ButtonModule } from "primeng/button";
import { InputTextModule } from "primeng/inputtext";
import { Textarea } from "primeng/textarea";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { DropdownModule } from "primeng/dropdown";
import { CheckboxModule } from "primeng/checkbox";
import { ToastModule } from "primeng/toast";
import { MessageService } from "primeng/api";
import { LucideAngularModule, TriangleAlert } from "lucide-angular";
import {
  RxResumeService,
  Resume,
  WorkExperience,
  EducationItem,
  Skill,
  LanguageItem,
} from "../../../services/rxresume.service";
import { ProfileService } from "../../../services/profile.service";
import { UsersService } from "../../../services/users.service";
import { environment } from "../../../../environments/environment";

@Component({
  selector: "app-cv",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    CardModule,
    ButtonModule,
    InputTextModule,
    Textarea,
    ProgressSpinnerModule,
    DropdownModule,
    CheckboxModule,
    ToastModule,
    LucideAngularModule,
  ],
  providers: [MessageService],
  templateUrl: "./cv.component.html",
  styleUrls: ["./cv.component.css"],
})
export class CvComponent implements OnInit, OnDestroy {
  private readonly rxResumeService = inject(RxResumeService);
  private readonly profileService = inject(ProfileService);
  private readonly usersService = inject(UsersService);
  private readonly route = inject(ActivatedRoute);
  private readonly messages = inject(MessageService);

  readonly TriangleAlert = TriangleAlert;

  userId: string | null = null;
  resume: Resume | null = null;
  loading = false;
  generating = false;
  uploadingPicture = false;
  pdfReady = false;
  pdfPolling = false;
  message = "";
  private pdfPollInterval: ReturnType<typeof setInterval> | null = null;

  resolveImageUrl(url?: string): string {
    if (!url) return "";
    if (url.startsWith("http://localhost:8080/api/v1/")) {
      const path = url.replace("http://localhost:8080", "");
      return `${environment.apiUrl || ""}${path}`;
    }
    if (url.startsWith("/")) {
      return `${environment.apiUrl || ""}${url}`;
    }
    return url;
  }

  templates = [
    { label: "Azurill", value: "azurill" },
    { label: "Bronzor", value: "bronzor" },
    { label: "Chikorita", value: "chikorita" },
    { label: "Ditgar", value: "ditgar" },
    { label: "Ditto", value: "ditto" },
    { label: "Gengar", value: "gengar" },
    { label: "Glalie", value: "glalie" },
    { label: "Kakuna", value: "kakuna" },
    { label: "Lapras", value: "lapras" },
    { label: "Leafish", value: "leafish" },
    { label: "Meowth", value: "meowth" },
    { label: "Onyx", value: "onyx" },
    { label: "Pikachu", value: "pikachu" },
    { label: "Rhyhorn", value: "rhyhorn" },
    { label: "Scizor", value: "scizor" },
  ];

  educationStatuses = [
    { label: "En curso", value: "En curso" },
    { label: "Completo", value: "Completo" },
    { label: "Incompleto", value: "Incompleto" },
  ];

  skillLevels = [
    { label: "Básico", value: "Básico" },
    { label: "Intermedio", value: "Intermedio" },
    { label: "Avanzado", value: "Avanzado" },
  ];

  ngOnInit() {
    this.route.queryParams.subscribe((params) => {
      this.userId = params["userId"] || null;
      this.loadData();
    });
  }

  loadData() {
    this.loading = true;

    const resumeObservable = this.userId
      ? this.rxResumeService.getUserResumeByAdmin(this.userId)
      : this.rxResumeService.getMyResume();

    // Load resume from DB (without RxResume API call)
    resumeObservable.subscribe({
      next: (res) => {
        this.resume = res;
        this.checkInitialPDFStatus();

        // Auto-load missing fields from user profile
        this.profileService.getMyProfile().subscribe({
          next: (profileRes) => {
            const user = profileRes.user;
            if (this.resume?.data?.basics) {
              if (!this.resume.data.basics.name && user) {
                this.resume.data.basics.name =
                  `${user.first_name} ${user.last_name}`.trim();
              }
              if (!this.resume.data.basics.email && user) {
                this.resume.data.basics.email = user.email;
              }
              if (!this.resume.data.basics.phone && user) {
                this.resume.data.basics.phone = user.phone;
              }
              if (
                !this.resume.data.basics.location &&
                profileRes.role_data?.address
              ) {
                this.resume.data.basics.location = profileRes.role_data.address;
              }
              if (!this.resume.data.metadata) {
                this.resume.data.metadata = { template: "onyx" };
              }
            }
            this.loading = false;
          },
          error: () => {
            this.loading = false; // Ignore profile error, we still have resume
          },
        });
      },
      error: () => {
        this.loading = false;
      },
    });
  }

  private checkInitialPDFStatus() {
    const statusObs = this.userId
      ? this.rxResumeService.getUserPDFStatusByAdmin(this.userId)
      : this.rxResumeService.getPDFStatus();

    statusObs.subscribe({
      next: (res) => {
        this.pdfReady = res.ready;
      },
    });
  }

  // Llamar después de guardar exitosamente
  private startPDFPolling() {
    this.pdfReady = false;
    this.pdfPolling = true;

    // Limpiar cualquier poll previo
    if (this.pdfPollInterval) clearInterval(this.pdfPollInterval);

    this.pdfPollInterval = setInterval(() => {
      const statusObs = this.userId
        ? this.rxResumeService.getUserPDFStatusByAdmin(this.userId)
        : this.rxResumeService.getPDFStatus();

      statusObs.subscribe({
        next: (res) => {
          if (res.ready) {
            this.pdfReady = true;
            this.pdfPolling = false;
            clearInterval(this.pdfPollInterval!);
            this.pdfPollInterval = null;
            this.message =
              'PDF listo. Hacé click en "Descargar PDF" para obtenerlo.';
          }
        },
        error: (err) => {
          if (err.status === 401 || err.status === 403) {
            this.pdfPolling = false;
            clearInterval(this.pdfPollInterval!);
            this.pdfPollInterval = null;
            this.message =
              "Tu sesión expiró. Actualizá la página y volvé a iniciar sesión.";
          }
          // Ignorar otros errores de polling, seguir intentando
        },
      });
    }, 3000); // Poll cada 3 segundos
  }

  generateResume(form?: NgForm) {
    if (form?.invalid) {
      this.message = "Por favor, completa los campos requeridos correctamente.";
      // Mark all as touched to show errors
      form.control.markAllAsTouched();
      return;
    }
    if (!this.resume?.data) return;

    this.generating = true;
    this.message = "";

    const updateObservable = this.userId
      ? this.rxResumeService.updateUserResumeByAdmin(
          this.userId,
          this.resume.data,
        )
      : this.rxResumeService.updateMyResume(this.resume.data);

    updateObservable.subscribe({
      next: (res) => {
        this.resume = res;
        this.generating = false;
        this.message = "¡Currículum guardado! Generando PDF...";
        this.startPDFPolling();
      },
      error: (err) => {
        this.generating = false;
        this.message = err.message || "Error al generar el currículum.";
      },
    });
  }

  ngOnDestroy() {
    if (this.pdfPollInterval) clearInterval(this.pdfPollInterval);
  }

  downloadPdf() {
    if (this.resume?.pdf_url && this.resume?.user_id) {
      this.rxResumeService.downloadPublicPDF(this.resume.user_id);
    } else {
      this.avisar(
        "warn",
        "Primero generá el currículum para poder descargarlo.",
      );
    }
  }

  shareResume() {
    if (this.resume?.public_url) {
      navigator.clipboard
        .writeText(this.resume.public_url)
        .then(() => {
          this.avisar("success", "Enlace copiado. Ya podés compartirlo.");
        })
        .catch(() => {
          this.avisar(
            "error",
            "No se pudo copiar el enlace. Intentá de nuevo.",
          );
        });
    } else {
      this.avisar(
        "warn",
        "Primero generá el currículum para poder compartirlo.",
      );
    }
  }

  private avisar(severity: "success" | "warn" | "error", detail: string) {
    this.messages.add({ severity, detail, life: 4000 });
  }

  // --- Helpers for arrays ---
  addWork() {
    if (!this.resume?.data) return;
    if (!this.resume.data.work) this.resume.data.work = [];
    const newWork: WorkExperience = {
      id: crypto.randomUUID(),
      company: "",
      position: "",
      startDate: "",
      endDate: "",
      description: "",
    };
    this.resume.data.work.push(newWork);
  }

  removeWork(index: number) {
    if (this.resume?.data?.work) {
      this.resume.data.work.splice(index, 1);
    }
  }

  addEducation() {
    if (!this.resume?.data) return;
    if (!this.resume.data.education) this.resume.data.education = [];
    const newEdu: EducationItem = {
      id: crypto.randomUUID(),
      institution: "",
      degree: "",
      field: "",
      startDate: "",
      endDate: "",
      grade: "",
      status: "En curso",
    };
    this.resume.data.education.push(newEdu);
  }

  removeEducation(index: number) {
    if (this.resume?.data?.education) {
      this.resume.data.education.splice(index, 1);
    }
  }

  addSkill() {
    if (!this.resume?.data) return;
    if (!this.resume.data.skills) this.resume.data.skills = [];
    const newSkill: Skill = {
      id: crypto.randomUUID(),
      name: "",
      level: "Intermedio",
    };
    this.resume.data.skills.push(newSkill);
  }

  removeSkill(index: number) {
    if (this.resume?.data?.skills) {
      this.resume.data.skills.splice(index, 1);
    }
  }

  addLanguage() {
    if (!this.resume?.data) return;
    if (!this.resume.data.languages) this.resume.data.languages = [];
    const newLang: LanguageItem = {
      id: crypto.randomUUID(),
      name: "",
      level: "Básico",
    };
    this.resume.data.languages.push(newLang);
  }

  removeLanguage(index: number) {
    if (this.resume?.data?.languages) {
      this.resume.data.languages.splice(index, 1);
    }
  }

  onPictureSelected(event: any) {
    const file = event.target.files[0];
    if (file) {
      if (file.size > 10 * 1024 * 1024) {
        this.avisar("warn", "La imagen no debe superar los 10 MB.");
        return;
      }
      this.uploadingPicture = true;

      const uploadObservable = this.userId
        ? this.rxResumeService.uploadUserPictureByAdmin(this.userId, file)
        : this.rxResumeService.uploadPicture(file);

      uploadObservable.subscribe({
        next: (res) => {
          this.resume = res;
          this.uploadingPicture = false;
          const newPic = res.data?.basics?.picture;
          if (newPic) {
            if (this.userId) {
              this.usersService.getUserProfile(this.userId).subscribe({
                next: (profile) => {
                  const partId = profile.role_data?.participant_id;
                  if (partId) {
                    this.usersService
                      .updateRoleData("participants", partId, { image_url: newPic })
                      .subscribe({
                        error: (err) => console.error("Error sincronizando foto en rol:", err),
                      });
                  }
                },
                error: (err) => console.error("Error obteniendo perfil para sincronizar foto:", err),
              });
            } else {
              this.profileService.updateProfile({ image_url: newPic }).subscribe({
                error: (err) => console.error("Error sincronizando foto en perfil:", err),
              });
            }
          }
        },
        error: (err) => {
          this.uploadingPicture = false;
          this.avisar(
            "error",
            "No se pudo subir la imagen" +
              (err.message ? `: ${err.message}` : "."),
          );
        },
      });
    }
  }
}
