import { Component, OnInit, PLATFORM_ID, inject } from "@angular/core";
import { CommonModule, isPlatformBrowser } from "@angular/common";
import { RouterLink } from "@angular/router";
import {
  LucideAngularModule,
  Cookie,
  ShieldCheck,
  CircleCheck,
  Settings,
  CircleQuestionMark,
  ArrowLeft,
} from "lucide-angular";
import { FooterComponent } from "../../components/layout/footer/footer.component";
import { rutaInicio } from "../../shared/utils/ruta-inicio";

@Component({
  selector: "app-cookie-policy",
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule, FooterComponent],
  templateUrl: "./cookie-policy.component.html",
  styleUrl: "./cookie-policy.component.css",
})
export class CookiePolicyComponent implements OnInit {
  /** "Volver al inicio": panel si hay sesión, landing si no. */
  readonly inicio = rutaInicio();
  readonly CookieIcon = Cookie;
  readonly ShieldCheckIcon = ShieldCheck;
  readonly CheckCircleIcon = CircleCheck;
  readonly SettingsIcon = Settings;
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
