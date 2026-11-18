import { Component, OnInit, PLATFORM_ID, inject } from "@angular/core";
import { CommonModule, isPlatformBrowser } from "@angular/common";
import { RouterLink } from "@angular/router";
import {
  LucideAngularModule,
  Shield,
  FileText,
  Lock,
  UserCheck,
  Building2,
  Mail,
  ArrowLeft,
} from "lucide-angular";
import { FooterComponent } from "../../components/layout/footer/footer.component";
import { rutaInicio } from "../../shared/utils/ruta-inicio";

@Component({
  selector: "app-privacy-policy",
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule, FooterComponent],
  templateUrl: "./privacy-policy.component.html",
  styleUrl: "./privacy-policy.component.css",
})
export class PrivacyPolicyComponent implements OnInit {
  /** "Volver al inicio": panel si hay sesión, landing si no. */
  readonly inicio = rutaInicio();
  readonly ShieldIcon = Shield;
  readonly FileTextIcon = FileText;
  readonly LockIcon = Lock;
  readonly UserCheckIcon = UserCheck;
  readonly BuildingIcon = Building2;
  readonly MailIcon = Mail;
  readonly ArrowLeftIcon = ArrowLeft;

  readonly lastUpdated = "03 de Septiembre de 2026";
  private readonly platformId = inject(PLATFORM_ID);

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      window.scrollTo(0, 0);
    }
  }
}
