import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { ActivatedRoute } from "@angular/router";
import { RxResumeService, Resume } from "../../services/rxresume.service";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { ButtonModule } from "primeng/button";

import { environment } from "../../../environments/environment";

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

  resume: Resume | null = null;
  loading = true;
  error = false;
  userId = "";

  resolveImageUrl(url?: string | null): string {
    if (!url || typeof url !== "string") return "";
    if (url.startsWith("http://localhost:8080/api/v1/")) {
      const path = url.replace("http://localhost:8080", "");
      return `${environment.apiUrl || ""}${path}`;
    }
    if (url.startsWith("/")) {
      return `${environment.apiUrl || ""}${url}`;
    }
    return url;
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

  downloadPDF() {
    if (this.userId) {
      this.rxResumeService.downloadPublicPDF(this.userId);
    }
  }
}
