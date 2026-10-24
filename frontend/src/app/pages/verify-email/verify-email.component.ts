import { Component, inject, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import { ActivatedRoute, Router, RouterLink } from "@angular/router";
import { FormsModule } from "@angular/forms";
import {
  LucideAngularModule,
  Mail,
  CircleCheck,
  CircleAlert,
  ArrowLeft,
  Send,
  RefreshCw,
} from "lucide-angular";
import { AuthService } from "../../services/auth.service";

@Component({
  selector: "app-verify-email",
  standalone: true,
  imports: [CommonModule, RouterLink, FormsModule, LucideAngularModule],
  templateUrl: "./verify-email.component.html",
  styleUrl: "./verify-email.component.css",
})
export class VerifyEmailComponent implements OnInit {
  readonly MailIcon = Mail;
  readonly CheckCircleIcon = CircleCheck;
  readonly AlertCircleIcon = CircleAlert;
  readonly ArrowLeftIcon = ArrowLeft;
  readonly SendIcon = Send;
  readonly RefreshCwIcon = RefreshCw;

  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly authService = inject(AuthService);

  email = "";
  token = "";

  verifying = false;
  verified = false;
  verifyError = "";

  resending = false;
  resendSuccess = "";
  resendError = "";

  ngOnInit(): void {
    this.token = this.route.snapshot.queryParamMap.get("token") || "";
    this.email =
      this.route.snapshot.queryParamMap.get("email") ||
      history.state?.email ||
      "";

    if (this.token) {
      this.executeTokenVerification(this.token);
    }
  }

  executeTokenVerification(token: string): void {
    this.verifying = true;
    this.verifyError = "";
    this.authService.verifyEmail(token).subscribe({
      next: () => {
        this.verifying = false;
        this.verified = true;
      },
      error: (err) => {
        this.verifying = false;
        this.verified = false;
        this.verifyError =
          err?.message ||
          "El enlace de verificación es inválido, ya fue utilizado o ha expirado.";
      },
    });
  }

  handleResendEmail(): void {
    if (!this.email || this.resending) return;

    this.resending = true;
    this.resendSuccess = "";
    this.resendError = "";

    this.authService.resendVerificationEmail(this.email.trim()).subscribe({
      next: () => {
        this.resending = false;
        this.resendSuccess =
          "¡Correo de verificación enviado! Por favor, revisá tu bandeja de entrada o spam.";
      },
      error: (err) => {
        this.resending = false;
        this.resendError =
          err?.message ||
          "No se pudo enviar el correo de verificación. Intentá nuevamente más tarde.";
      },
    });
  }

  goToLogin(): void {
    void this.router.navigate(["/login"]);
  }
}
