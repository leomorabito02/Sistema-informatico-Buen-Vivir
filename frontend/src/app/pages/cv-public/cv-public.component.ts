import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { ActivatedRoute } from "@angular/router";
import { RxResumeService, Resume } from "../../services/rxresume.service";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { ButtonModule } from "primeng/button";

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
