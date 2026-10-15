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
import { CarouselModule } from "primeng/carousel";
import { ColorPickerModule } from "primeng/colorpicker";
import { DialogModule } from "primeng/dialog";
import { MessageService } from "primeng/api";
import { LucideAngularModule, TriangleAlert } from "lucide-angular";
import {
  RxResumeService,
  Resume,
  WorkExperience,
  EducationItem,
  Skill,
  LanguageItem,
  Reference,
} from "../../../services/rxresume.service";
import { ProfileService } from "../../../services/profile.service";
import { UsersService } from "../../../services/users.service";
import { ImageService } from "../../../services/image.service";
import { ProfileResponse } from "../../../core/models/auth.model";

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
    CarouselModule,
    ColorPickerModule,
    DialogModule,
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
  private readonly imageService = inject(ImageService);
  private readonly route = inject(ActivatedRoute);
  private readonly messages = inject(MessageService);

  readonly TriangleAlert = TriangleAlert;

  userId: string | null = null;
  resume: Resume | null = null;
  loading = false;
  generating = false;
  pdfReady = false;
  pdfPolling = false;
  message = "";
  private pdfPollInterval: ReturnType<typeof setInterval> | null = null;

  resolveImageUrl(url?: string): string {
    return this.imageService.resolveImageUrl(url);
  }

  templates = [
    { label: "Azurill", value: "azurill", image: "assets/cv-templates/azurill.png" },
    { label: "Bronzor", value: "bronzor", image: "assets/cv-templates/bronzor.png" },
    { label: "Chikorita", value: "chikorita", image: "assets/cv-templates/chikorita.png" },
    { label: "Ditgar", value: "ditgar", image: "assets/cv-templates/ditgar.png" },
    { label: "Ditto", value: "ditto", image: "assets/cv-templates/ditto.png" },
    { label: "Gengar", value: "gengar", image: "assets/cv-templates/gengar.png" },
    { label: "Glalie", value: "glalie", image: "assets/cv-templates/glalie.png" },
    { label: "Kakuna", value: "kakuna", image: "assets/cv-templates/kakuna.png" },
    { label: "Lapras", value: "lapras", image: "assets/cv-templates/lapras.png" },
    { label: "Leafish", value: "leafish", image: "assets/cv-templates/leafish.png" },
    { label: "Meowth", value: "meowth", image: "assets/cv-templates/meowth.png" },
    { label: "Onyx", value: "onyx", image: "assets/cv-templates/onyx.png" },
    { label: "Pikachu", value: "pikachu", image: "assets/cv-templates/pikachu.png" },
    { label: "Rhyhorn", value: "rhyhorn", image: "assets/cv-templates/rhyhorn.png" },
    { label: "Scizor", value: "scizor", image: "assets/cv-templates/scizor.png" },
  ];

  responsiveOptions = [
    { breakpoint: "1400px", numVisible: 3, numScroll: 1 },
    { breakpoint: "1024px", numVisible: 2, numScroll: 1 },
    { breakpoint: "768px", numVisible: 1, numScroll: 1 },
  ];

  colorPresets = [
    { name: "Verde Institucional", value: "#485839" },
    { name: "Terracota", value: "#9a5b43" },
    { name: "Azul Marino", value: "#1e3a8a" },
    { name: "Grafito", value: "#334155" },
    { name: "Borgoña", value: "#881337" },
    { name: "Verde Bosque", value: "#2d3b23" },
  ];

  showTemplateModal = false;

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
            this.populateBasicsFromProfile(profileRes);
            this.ensureResumeMetadata();
            this.loading = false;
          },
          error: () => {
            this.ensureResumeMetadata();
            this.loading = false; // Ignore profile error, we still have resume
          },
        });
      },
      error: () => {
        this.loading = false;
      },
    });
  }

  private populateBasicsFromProfile(profileRes: ProfileResponse): void {
    const basics = this.resume?.data?.basics;
    const user = profileRes.user;
    if (!basics) return;

    if (!basics.name && user) {
      basics.name = `${user.first_name} ${user.last_name}`.trim();
    }
    if (!basics.email && user?.email) {
      basics.email = user.email;
    }
    if (!basics.phone && user?.phone) {
      basics.phone = user.phone;
    }
    if (!basics.location && profileRes.role_data?.address) {
      basics.location = profileRes.role_data.address;
    }
  }

  private ensureResumeMetadata(): void {
    if (!this.resume?.data) return;

    const defaultColors = {
      primary: "#485839",
      text: "rgba(0, 0, 0, 1)",
      background: "rgba(255, 255, 255, 1)",
    };

    if (!this.resume.data.metadata) {
      this.resume.data.metadata = {
        template: "onyx",
        design: { colors: defaultColors },
      };
    } else {
      const meta = this.resume.data.metadata;
      meta.template = meta.template || "onyx";
      meta.design = meta.design || { colors: defaultColors };
      meta.design.colors = meta.design.colors || defaultColors;
      meta.design.colors.primary = meta.design.colors.primary || "#485839";
    }

    if (!this.resume.data.references) {
      this.resume.data.references = [];
    }
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

  // --- Template & Color customization ---
  openTemplateModal(): void {
    this.showTemplateModal = true;
  }

  selectTemplate(templateValue: string): void {
    if (!this.resume?.data) return;
    if (!this.resume.data.metadata) this.resume.data.metadata = {};
    this.resume.data.metadata.template = templateValue;
  }

  getPrimaryColor(): string {
    return this.resume?.data?.metadata?.design?.colors?.primary || "#485839";
  }

  setPrimaryColor(color: string): void {
    if (!this.resume?.data) return;
    if (!this.resume.data.metadata) this.resume.data.metadata = {};
    if (!this.resume.data.metadata.design) {
      this.resume.data.metadata.design = {
        colors: {
          primary: color,
          text: "rgba(0, 0, 0, 1)",
          background: "rgba(255, 255, 255, 1)",
        },
      };
    } else if (!this.resume.data.metadata.design.colors) {
      this.resume.data.metadata.design.colors = {
        primary: color,
        text: "rgba(0, 0, 0, 1)",
        background: "rgba(255, 255, 255, 1)",
      };
    } else {
      this.resume.data.metadata.design.colors.primary = color;
    }
  }

  getSelectedTemplateImage(): string {
    const tplName = this.resume?.data?.metadata?.template || "onyx";
    const found = this.templates.find((t) => t.value === tplName);
    return found ? found.image : "assets/cv-templates/onyx.png";
  }

  getSelectedTemplateLabel(): string {
    const tplName = this.resume?.data?.metadata?.template || "onyx";
    const found = this.templates.find((t) => t.value === tplName);
    return found ? found.label : "Onyx";
  }

  // --- References ---
  addReference(): void {
    if (!this.resume?.data) return;
    if (!this.resume.data.references) this.resume.data.references = [];
    const newRef: Reference = {
      id: crypto.randomUUID(),
      name: "",
      description: "",
    };
    this.resume.data.references.push(newRef);
  }

  removeReference(index: number): void {
    if (this.resume?.data?.references) {
      this.resume.data.references.splice(index, 1);
    }
  }
}
