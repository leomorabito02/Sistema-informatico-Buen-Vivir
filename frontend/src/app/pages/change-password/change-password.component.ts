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
  KeyRound,
  Eye,
  EyeOff,
  Lock,
} from "lucide-angular";
import { AuthService } from "../../services/auth.service";

@Component({
  selector: "app-change-password",
  standalone: true,
  imports: [CommonModule, RouterLink, FormsModule, LucideAngularModule],
  templateUrl: "./change-password.component.html",
  styleUrl: "./change-password.component.css",
})
export class ChangePasswordComponent implements OnInit {
  readonly MailIcon = Mail;
  readonly CheckCircleIcon = CircleCheck;
  readonly AlertCircleIcon = CircleAlert;
  readonly ArrowLeftIcon = ArrowLeft;
  readonly SendIcon = Send;
  readonly RefreshCwIcon = RefreshCw;
  readonly KeyRoundIcon = KeyRound;
  readonly EyeIcon = Eye;
  readonly EyeOffIcon = EyeOff;
  readonly LockIcon = Lock;

  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly authService = inject(AuthService);

  token = "";

  // Step 1: Request reset email
  requestEmail = "";
  requesting = false;
  requestSuccess = "";
  requestError = "";

  // Step 2: Set new password
  newPassword = "";
  confirmPassword = "";
  showNewPassword = false;
  showConfirmPassword = false;
  resetting = false;
  resetSuccess = "";
  resetError = "";

  ngOnInit(): void {
    this.token = this.route.snapshot.queryParamMap.get("token") || "";
    this.requestEmail =
      this.route.snapshot.queryParamMap.get("email") ||
      history.state?.email ||
      "";
  }

  get hasMinLength(): boolean {
    return this.newPassword.length >= 8;
  }

  get hasUpper(): boolean {
    return /[A-Z]/.test(this.newPassword);
  }

  get hasLower(): boolean {
    return /[a-z]/.test(this.newPassword);
  }

  get hasSpecial(): boolean {
    return /[^A-Za-z0-9]/.test(this.newPassword);
  }

  get isPasswordValid(): boolean {
    return (
      this.hasMinLength && this.hasUpper && this.hasLower && this.hasSpecial
    );
  }

  get passwordsMatch(): boolean {
    return (
      this.newPassword.length > 0 &&
      this.newPassword === this.confirmPassword
    );
  }

  toggleNewPassword(): void {
    this.showNewPassword = !this.showNewPassword;
  }

  toggleConfirmPassword(): void {
    this.showConfirmPassword = !this.showConfirmPassword;
  }

  handleRequestReset(): void {
    const email = this.requestEmail.trim();
    if (!email || this.requesting) return;

    this.requesting = true;
    this.requestSuccess = "";
    this.requestError = "";

    this.authService.forgotPassword(email).subscribe({
      next: (res) => {
        this.requesting = false;
        this.requestSuccess =
          res?.message ||
          "Si la dirección de correo electrónico está registrada, recibirás un enlace para restablecer tu contraseña.";
      },
      error: (err) => {
        this.requesting = false;
        this.requestError =
          err?.message ||
          "Ocurrió un error al procesar la solicitud. Por favor, intente nuevamente.";
      },
    });
  }

  handleResetPassword(): void {
    if (!this.token || !this.isPasswordValid || !this.passwordsMatch || this.resetting) {
      return;
    }

    this.resetting = true;
    this.resetSuccess = "";
    this.resetError = "";

    this.authService.resetPassword(this.token.trim(), this.newPassword).subscribe({
      next: (res) => {
        this.resetting = false;
        this.resetSuccess =
          res?.message ||
          "Contraseña restablecida exitosamente. Ya podés iniciar sesión con tu nueva contraseña.";
      },
      error: (err) => {
        this.resetting = false;
        this.resetError =
          err?.message ||
          "El enlace de restablecimiento es inválido, ya fue utilizado o ha expirado.";
      },
    });
  }

  clearTokenAndRetry(): void {
    this.token = "";
    this.resetError = "";
    this.resetSuccess = "";
    this.newPassword = "";
    this.confirmPassword = "";
    void this.router.navigate(["/change-password"]);
  }

  goToLogin(): void {
    void this.router.navigate(["/login"]);
  }
}
