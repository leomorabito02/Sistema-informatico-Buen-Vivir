import { Component, inject, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import {
  FormBuilder,
  FormGroup,
  FormsModule,
  ReactiveFormsModule,
  Validators,
} from "@angular/forms";
import { Router, RouterLink } from "@angular/router";
import {
  LucideAngularModule,
  Mail,
  Lock,
  Eye,
  EyeOff,
  ShieldCheck,
} from "lucide-angular";
import { AuthService } from "../../services/auth.service";
import { AvailableAccount } from "../../core/models/auth.model";

import { ButtonModule } from "primeng/button";
import { CardModule } from "primeng/card";
import { InputTextModule } from "primeng/inputtext";
import { MessageModule } from "primeng/message";

@Component({
  selector: "app-login",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    RouterLink,
    LucideAngularModule,
    ButtonModule,
    CardModule,
    InputTextModule,
    MessageModule,
  ],
  templateUrl: "./login.component.html",
  styleUrl: "./login.component.css",
})
export class LoginComponent implements OnInit {
  readonly MailIcon = Mail;
  readonly LockIcon = Lock;
  readonly EyeIcon = Eye;
  readonly EyeOffIcon = EyeOff;
  readonly ShieldIcon = ShieldCheck;

  private readonly fb = inject(FormBuilder);
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  loginForm: FormGroup;
  loginError = "";
  locked = false;
  loading = false;
  googleLoading = false;
  failedAttempts = 0;
  MAX_ATTEMPTS = 5;
  showPassword = false;

  // 2FA state
  is2faStep = false;
  twofaMethod = "email";
  twofaNonce = "";
  twofaCode = "";
  trustDevice = false;
  twofaEmail = "";
  resendSuccess = "";
  resendingCode = false;

  constructor() {
    this.loginForm = this.fb.group({
      email: ["", [Validators.required, Validators.email]],
      password: ["", [Validators.required, Validators.minLength(4)]],
      rememberMe: [false],
    });
  }

  ngOnInit(): void {
    const savedEmail = localStorage.getItem("rememberedEmail");
    if (savedEmail) {
      this.loginForm.patchValue({
        email: savedEmail,
        rememberMe: true,
      });
    }
  }

  togglePasswordVisibility(): void {
    this.showPassword = !this.showPassword;
  }

  private getFirstFormError(): string {
    const c = this.loginForm.controls;
    if (c["email"].hasError("required") || c["password"].hasError("required")) {
      return "Por favor, ingresá tu email y contraseña.";
    }
    if (c["email"].hasError("email")) {
      return "Por favor, ingresá un correo electrónico válido.";
    }
    if (c["password"].hasError("minlength")) {
      return "La contraseña debe tener al menos 4 caracteres.";
    }
    return "Por favor, completá los campos del formulario correctamente.";
  }

  onSubmit() {
    if (this.locked || this.loading || this.googleLoading) return;

    if (this.loginForm.invalid) {
      this.loginForm.markAllAsTouched();
      this.loginError = this.getFirstFormError();
      return;
    }

    const { email, password, rememberMe } = this.loginForm.value;

    if (rememberMe) {
      localStorage.setItem("rememberedEmail", email);
    } else {
      localStorage.removeItem("rememberedEmail");
    }

    this.loading = true;
    this.loginError = "";

    this.authService.checkCredentials(email, password).subscribe({
      next: (res) => {
        this.loading = false;
        this.loginError = "";
        if (res.requires_2fa) {
          this.is2faStep = true;
          this.twofaNonce = res.twofa_nonce || "";
          this.twofaMethod = res.twofa_method || "email";
          this.twofaEmail = res.email || email;
          this.twofaCode = "";
          return;
        }
        if (res.is_email_verified === false) {
          void this.router.navigate(["/verify-email"], {
            queryParams: { email: res.email || email },
          });
          return;
        }
        this.handleAccountsResponse(res.accounts, res.login_nonce);
      },
      error: (err) => {
        this.loading = false;
        const errorMsg = err.error?.error || "";
        if (
          err.error?.user_status === "blocked" ||
          errorMsg.includes("activa") ||
          errorMsg.includes("bloque")
        ) {
          this.loginError = "";
          void this.router.navigate(["/account-notice"]);
        } else {
          this.handleFailedAttempt(errorMsg);
        }
      },
    });
  }

  handleGoogleLogin() {
    if (this.loading || this.googleLoading) return;
    this.googleLoading = true;
    this.loginError = "";

    this.authService.checkGoogleCredentials().subscribe({
      next: (res) => {
        this.googleLoading = false;
        this.loginError = "";
        if (res.requires_2fa) {
          this.is2faStep = true;
          this.twofaNonce = res.twofa_nonce || "";
          this.twofaMethod = res.twofa_method || "email";
          this.twofaEmail = res.email || "";
          this.twofaCode = "";
          return;
        }
        if (res.is_email_verified === false) {
          void this.router.navigate(["/verify-email"], {
            queryParams: { email: res.email },
          });
          return;
        }
        this.handleAccountsResponse(res.accounts, res.login_nonce);
      },
      error: (err) => {
        this.googleLoading = false;
        const errorMsg = err.error?.error;
        this.loginError =
          errorMsg ||
          "Hubo un error al iniciar sesión con Google. Intentá nuevamente.";
      },
    });
  }

  onVerifyTwofa(): void {
    if (this.loading || this.twofaCode?.trim().length !== 6) return;
    this.loading = true;
    this.loginError = "";
    this.resendSuccess = "";

    this.authService
      .verifyTwoFactor(this.twofaNonce, this.twofaCode.trim(), this.trustDevice)
      .subscribe({
        next: (res) => {
          this.loading = false;
          this.is2faStep = false;
          this.handleAccountsResponse(res.accounts, res.login_nonce);
        },
        error: (err) => {
          this.loading = false;
          const errorMsg = err.error?.error || err.message || "";
          this.loginError = errorMsg || "Código de verificación incorrecto.";
        },
      });
  }

  onResendTwofa(): void {
    if (this.resendingCode || !this.twofaNonce) return;
    this.resendingCode = true;
    this.loginError = "";
    this.resendSuccess = "";

    this.authService.resendTwoFactorCode(this.twofaNonce).subscribe({
      next: (res) => {
        this.resendingCode = false;
        this.resendSuccess =
          res.message || "Se ha reenviado el código a tu correo electrónico.";
      },
      error: (err) => {
        this.resendingCode = false;
        const errorMsg = err.error?.error || err.message || "";
        this.loginError = errorMsg || "Error al reenviar el código.";
      },
    });
  }

  cancelTwofa(): void {
    this.is2faStep = false;
    this.twofaNonce = "";
    this.twofaCode = "";
    this.trustDevice = false;
    this.loginError = "";
    this.resendSuccess = "";
  }

  /**
   * If only one account is available, log in directly.
   * If multiple accounts exist, navigate to role selection screen.
   * The loginNonce must be forwarded to Step 2 in both paths.
   */
  private handleAccountsResponse(
    accounts: AvailableAccount[],
    loginNonce: string,
  ) {
    if (!accounts || accounts.length === 0) {
      this.handleFailedAttempt(
        "No se encontraron cuentas activas asociadas a este correo electrónico.",
      );
      return;
    }

    if (accounts.length === 1) {
      // Single account — log in directly
      this.loginWithAccount(accounts[0], loginNonce);
    } else {
      // Multiple accounts — navigate to role selection, carry nonce in router state
      void this.router.navigate(["/select-role"], {
        state: { accounts, loginNonce },
      });
    }
  }

  /**
   * Initiates step 2: select account/role and receive the JWT.
   */
  private loginWithAccount(account: AvailableAccount, loginNonce: string) {
    if (account.status !== "active") {
      void this.router.navigate(["/account-notice"]);
      return;
    }

    this.loading = true;
    this.authService
      .loginWithAccount(account.user_id, account.role_id, loginNonce)
      .subscribe({
        next: (res) => {
          this.loading = false;
          this.loginError = "";
          if (res.user_status === "blocked") {
            void this.router.navigate(["/account-notice"]);
          } else if (res.approval_status === "pending") {
            void this.router.navigate(["/pending-approval"]);
          } else {
            void this.router.navigate(["/dashboard"]);
          }
        },
        error: (err) => {
          this.loading = false;
          const errorMsg = err.error?.error || err.message || "";
          if (
            errorMsg === "EMAIL_NOT_VERIFIED" ||
            errorMsg.includes("verificar tu correo")
          ) {
            const email = this.loginForm.get("email")?.value;
            void this.router.navigate(["/verify-email"], {
              queryParams: email ? { email } : undefined,
            });
            return;
          }
          if (errorMsg.includes("activa") || errorMsg.includes("bloque")) {
            void this.router.navigate(["/account-notice"]);
          } else {
            this.handleFailedAttempt(errorMsg);
          }
        },
      });
  }

  handleFailedAttempt(specificError?: string) {
    this.failedAttempts++;
    if (this.failedAttempts >= this.MAX_ATTEMPTS) {
      this.locked = true;
      this.loginError =
        "Tu cuenta fue bloqueada temporalmente por demasiados intentos fallidos. Intentá nuevamente más tarde.";
    } else {
      this.loginError = specificError || "Email o contraseña incorrectos.";
    }
  }
}
