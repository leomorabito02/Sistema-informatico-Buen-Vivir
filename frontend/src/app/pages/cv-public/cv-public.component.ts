import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { ActivatedRoute } from "@angular/router";
import { RxResumeService, Resume } from "../../services/rxresume.service";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { ButtonModule } from "primeng/button";

import { ImageService } from "../../services/image.service";
import { CvPdfService } from "../../services/cv-pdf.service";

@Component({
  selector: "app-cv-public",
  standalone: true,
  imports: [CommonModule, ProgressSpinnerModule, ButtonModule],
  templateUrl: "./cv-public.component.html",
  styleUrls: ["./cv-public.component.css"],
})
export class CvPublicComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly rxResumeService = inject(RxResumeService);
  private readonly imageService = inject(ImageService);
  private readonly cvPdfService = inject(CvPdfService);

  resume: Resume | null = null;
  loading = true;
  error = false;
  userId = "";

  resolveImageUrl(url?: string | null): string {
    return this.imageService.resolveImageUrl(url);
  }

  ngOnInit() {
    this.userId = this.route.snapshot.paramMap.get("id") || "";
    if (this.userId) {
      this.loadResume();
    } else {
      this.error = true;
      this.loading = false;
    }
  }

  loadResume() {
    this.rxResumeService.getPublicResume(this.userId).subscribe({
      next: (res) => {
        this.resume = res;
        this.loading = false;
        if (this.resume?.data?.basics) {
          document.title = `CV - ${this.resume.data.basics.name}`;
        }
      },
      error: () => {
        this.error = true;
        this.loading = false;
      },
    });
  }

  async downloadPDF() {
    if (!this.userId) return;

    if (this.resume?.data) {
      const filename = `cv-${this.userId}.pdf`;
      const picture = this.resolveImageUrl(this.resume.data.basics?.picture);

      try {
        await this.cvPdfService.downloadPdf(this.resume, filename, {
          primaryColor: this.getPrimaryColor(),
          pictureUrl: picture,
        });
        return;
      } catch {
        // Fallback to backend PDF generation
      }
    }

    this.rxResumeService.downloadPublicPDF(this.userId);
  }

  getPrimaryColor(): string {
    return this.resume?.data?.metadata?.design?.colors?.primary || "#485839";
  }
}
