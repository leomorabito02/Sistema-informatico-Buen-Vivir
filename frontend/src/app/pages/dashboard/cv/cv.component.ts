import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule, NgForm } from "@angular/forms";
import { CardModule } from "primeng/card";
import { ButtonModule } from "primeng/button";
import { InputTextModule } from "primeng/inputtext";
import { Textarea } from "primeng/textarea";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { DropdownModule } from "primeng/dropdown";
import { CheckboxModule } from "primeng/checkbox";
import {
  RxResumeService,
  Resume,
  WorkExperience,
  EducationItem,
  Skill,
  LanguageItem,
} from "../../../services/rxresume.service";
import { ProfileService } from "../../../services/profile.service";

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
  ],
  templateUrl: "./cv.component.html",
  styleUrls: ["./cv.component.css"],
})
export class CvComponent implements OnInit {
  private readonly rxResumeService = inject(RxResumeService);
  private readonly profileService = inject(ProfileService);

  resume: Resume | null = null;
  loading = false;
  generating = false;
  uploadingPicture = false;
  message = "";

  templates = [
    { label: "Onyx", value: "onyx" },
    { label: "Pikachu", value: "pikachu" },
    { label: "Gengar", value: "gengar" },
    { label: "Castform", value: "castform" },
    { label: "Glalie", value: "glalie" },
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
    this.loadData();
  }

  loadData() {
    this.loading = true;

    // Load resume from DB (without RxResume API call)
    this.rxResumeService.getMyResume().subscribe({
      next: (res) => {
        this.resume = res;

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

    this.rxResumeService.updateMyResume(this.resume.data).subscribe({
      next: (res) => {
        this.resume = res;
        this.generating = false;
        this.message = "¡Currículum guardado y generado exitosamente!";
      },
      error: (err) => {
        this.generating = false;
        this.message = err.message || "Error al generar el currículum.";
      },
    });
  }

  downloadPdf() {
    if (this.resume?.pdf_url && this.resume?.user_id) {
      this.rxResumeService.downloadPublicPDF(this.resume.user_id);
    } else {
      alert("Debes generar el currículum primero.");
    }
  }

  shareResume() {
    if (this.resume?.public_url) {
      navigator.clipboard
        .writeText(this.resume.public_url)
        .then(() => {
          alert("Enlace copiado al portapapeles: " + this.resume!.public_url);
        })
        .catch(() => {
          alert("Error al copiar el enlace.");
        });
    } else {
      alert("Debes generar el currículum primero para poder compartirlo.");
    }
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
        alert("La imagen no debe superar los 10MB.");
        return;
      }
      this.uploadingPicture = true;
      this.rxResumeService.uploadPicture(file).subscribe({
        next: (res) => {
          this.resume = res;
          this.uploadingPicture = false;
        },
        error: (err) => {
          this.uploadingPicture = false;
          alert("Error al subir la imagen: " + (err.message || "Desconocido"));
        },
      });
    }
  }
}
