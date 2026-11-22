import { Component, OnInit, ChangeDetectorRef, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { ActivatedRoute } from "@angular/router";
import { FormsModule, NgForm } from "@angular/forms";
import { forkJoin, of } from "rxjs";
import { catchError } from "rxjs/operators";
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
import { CvPdfService } from "../../../services/cv-pdf.service";
import { SkillsService } from "../../../services/skills.service";
import { User, RoleData } from "../../../core/models/auth.model";
import { skillKey } from "../../../core/models/skills.model";

/** Nivel con el que figuran en el PDF las habilidades ganadas en el centro. */
export const NIVEL_HABILIDAD_CENTRO = "Adquirida en el centro";

/**
 * Suma al CV las habilidades ganadas en el centro que la persona no cargó
 * a mano (misma habilidad escrita distinto = una sola). No modifica el
 * CV guardado: se usa al armar el PDF.
 */
export function conHabilidadesDelCentro(
  skills: Skill[] | undefined,
  delCentro: string[],
): Skill[] {
  const manuales = skills ?? [];
  const vistas = new Set(manuales.map((s) => skillKey(s.name || "")));
  const extra: Skill[] = [];
  for (const nombre of delCentro) {
    const key = skillKey(nombre);
    if (!nombre.trim() || vistas.has(key)) continue;
    vistas.add(key);
    extra.push({
      id: `centro-${key}`,
      name: nombre,
      level: NIVEL_HABILIDAD_CENTRO,
    });
  }
  return [...manuales, ...extra];
}

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
export class CvComponent implements OnInit {
  private readonly rxResumeService = inject(RxResumeService);
  private readonly profileService = inject(ProfileService);
  private readonly usersService = inject(UsersService);
  private readonly imageService = inject(ImageService);
  private readonly cvPdfService = inject(CvPdfService);
  private readonly skillsService = inject(SkillsService);
  private readonly route = inject(ActivatedRoute);
  private readonly messages = inject(MessageService);
  private readonly cdr = inject(ChangeDetectorRef);

  readonly TriangleAlert = TriangleAlert;

  /**
   * Habilidades que la persona ganó en el centro al completar actividades
   * (22/11/2026). Se leen solas y van al PDF; no se editan a mano.
   */
  habilidadesCentro: string[] = [];

  userId: string | null = null;
  resume: Resume | null = null;
  loading = false;
  generating = false;
  pdfReady = false;
  pdfPolling = false;
  message = "";
  userProfilePicture = "";

  resolveImageUrl(url?: string): string {
    return this.imageService.resolveImageUrl(url);
  }

  readonly templates = [
    {
      label: "Azurill",
      value: "azurill",
      image: "assets/cv-templates/azurill.png",
    },
    {
      label: "Bronzor",
      value: "bronzor",
      image: "assets/cv-templates/bronzor.png",
    },
    {
      label: "Chikorita",
      value: "chikorita",
      image: "assets/cv-templates/chikorita.png",
    },
    {
      label: "Ditgar",
      value: "ditgar",
      image: "assets/cv-templates/ditgar.png",
    },
    { label: "Ditto", value: "ditto", image: "assets/cv-templates/ditto.png" },
    {
      label: "Gengar",
      value: "gengar",
      image: "assets/cv-templates/gengar.png",
    },
    {
      label: "Glalie",
      value: "glalie",
      image: "assets/cv-templates/glalie.png",
    },
    {
      label: "Kakuna",
      value: "kakuna",
      image: "assets/cv-templates/kakuna.png",
    },
    {
      label: "Lapras",
      value: "lapras",
      image: "assets/cv-templates/lapras.png",
    },
    {
      label: "Leafish",
      value: "leafish",
      image: "assets/cv-templates/leafish.png",
    },
    {
      label: "Meowth",
      value: "meowth",
      image: "assets/cv-templates/meowth.png",
    },
    { label: "Onyx", value: "onyx", image: "assets/cv-templates/onyx.png" },
    {
      label: "Pikachu",
      value: "pikachu",
      image: "assets/cv-templates/pikachu.png",
    },
    {
      label: "Rhyhorn",
      value: "rhyhorn",
      image: "assets/cv-templates/rhyhorn.png",
    },
    {
      label: "Scizor",
      value: "scizor",
      image: "assets/cv-templates/scizor.png",
    },
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

    const resume$ = this.userId
      ? this.rxResumeService.getUserResumeByAdmin(this.userId)
      : this.rxResumeService.getMyResume();

    const profile$ = this.userId
      ? this.usersService
          .getUserProfile(this.userId)
          .pipe(catchError(() => of(null)))
      : this.profileService.getMyProfile().pipe(catchError(() => of(null)));

    // Solo el CV propio: el endpoint toma al participante del token. Si
    // falla (o no es participante) el CV se arma igual, sin esta sección.
    const habilidades$ = this.userId
      ? of([])
      : this.skillsService.mySkills().pipe(catchError(() => of([])));

    forkJoin({
      resume: resume$,
      profile: profile$,
      habilidades: habilidades$,
    }).subscribe({
      next: ({ resume, profile, habilidades }) => {
        this.resume = resume;
        this.habilidadesCentro = [
          ...new Set(
            habilidades.map((h) => (h.nombre || "").trim()).filter(Boolean),
          ),
        ].sort((a, b) => a.localeCompare(b, "es"));
        this.checkInitialPDFStatus();
        if (profile) {
          this.populateBasicsFromProfile(profile);
        }
        this.ensureResumeMetadata();
        this.loading = false;
        this.cdr.detectChanges();
      },
      error: () => {
        this.loading = false;
        this.cdr.detectChanges();
      },
    });
  }

  private populateBasicsFromProfile(profileRes: {
    user?: Partial<User>;
    role_data?: Partial<RoleData>;
  }): void {
    const basics = this.resume?.data?.basics;
    const user = profileRes.user;
    if (!basics) return;

    if (!basics.name && user) {
      basics.name = `${user.first_name || ""} ${user.last_name || ""}`.trim();
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

    const userPic =
      user?.image_url ||
      (profileRes.role_data?.["image_url"] as string | undefined) ||
      (profileRes.role_data?.["profile_picture_url"] as string | undefined);
    if (userPic) {
      this.userProfilePicture = userPic;
      basics.picture = userPic;
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
    this.pdfReady = true;
  }

  generateResume(form?: NgForm) {
    if (form?.invalid) {
      this.message = "Por favor, completa los campos requeridos correctamente.";
      form.control.markAllAsTouched();
      return;
    }
    if (!this.resume?.data) return;

    if (this.resume.data.basics) {
      if (!this.resume.data.basics.picture && this.userProfilePicture) {
        this.resume.data.basics.picture = this.userProfilePicture;
      }
    }

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
        this.pdfReady = true;
        this.message = "¡Currículum guardado y listo para descargar!";
        this.avisar("success", "Currículum guardado correctamente.");
      },
      error: (err) => {
        this.generating = false;
        this.message = err.message || "Error al generar el currículum.";
      },
    });
  }

  async downloadPdf() {
    if (!this.resume?.data) {
      this.avisar(
        "warn",
        "Primero generá el currículum para poder descargarlo.",
      );
      return;
    }

    const id = this.userId || this.resume.user_id || "usuario";
    const filename = `cv-${id}.pdf`;
    const picToUse = this.resume.data.basics.picture || this.userProfilePicture;
    const picture = this.resolveImageUrl(picToUse);

    // El PDF lleva también las habilidades ganadas en el centro.
    const resumeParaPdf = {
      ...this.resume,
      data: {
        ...this.resume.data,
        skills: conHabilidadesDelCentro(
          this.resume.data.skills,
          this.habilidadesCentro,
        ),
      },
    };

    try {
      await this.cvPdfService.downloadPdf(resumeParaPdf, filename, {
        primaryColor: this.getPrimaryColor(),
        pictureUrl: picture,
      });
    } catch {
      if (this.resume?.user_id) {
        this.rxResumeService.downloadPublicPDF(this.resume.user_id);
      }
    }
  }

  shareResume() {
    if (this.resume?.public_url) {
      const path = this.resume.public_url;
      const origin =
        typeof window !== "undefined" ? window.location.origin : "";
      let fullUrl = path;
      if (!path.startsWith("http")) {
        const normalizedPath = path.startsWith("/") ? path : `/${path}`;
        fullUrl = `${origin}${normalizedPath}`;
      }

      navigator.clipboard
        .writeText(fullUrl)
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
