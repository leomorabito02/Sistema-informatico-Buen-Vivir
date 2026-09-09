import { Component, OnInit, PLATFORM_ID, inject } from "@angular/core";
import { CommonModule, isPlatformBrowser } from "@angular/common";
import { RouterLink } from "@angular/router";
import {
  LucideAngularModule,
  FileCheck,
  UserCheck,
  Scale,
  TriangleAlert,
  CircleQuestionMark,
  ArrowLeft,
} from "lucide-angular";
import { FooterComponent } from "../../components/layout/footer/footer.component";

@Component({
  selector: "app-terms-conditions",
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule, FooterComponent],
  templateUrl: "./terms-conditions.component.html",
  styleUrl: "./terms-conditions.component.css",
})
export class TermsConditionsComponent implements OnInit {
  readonly FileCheckIcon = FileCheck;
  readonly UserCheckIcon = UserCheck;
  readonly ScaleIcon = Scale;
  readonly AlertTriangleIcon = TriangleAlert;
  readonly HelpCircleIcon = CircleQuestionMark;
  readonly ArrowLeftIcon = ArrowLeft;

  readonly lastUpdated = "03 de Septiembre de 2026";
  private readonly platformId = inject(PLATFORM_ID);

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      window.scrollTo(0, 0);
    }
  }
}
