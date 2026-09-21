import { Component, inject, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import {
  FormBuilder,
  FormGroup,
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
          this.router.navigate(["/account-notice"]);
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
      this.router.navigate(["/select-role"], {
        state: { accounts, loginNonce },
      });
    }
  }

  /**
   * Initiates step 2: select account/role and receive the JWT.
   */
  private loginWithAccount(account: AvailableAccount, loginNonce: string) {
    if (account.status !== "active") {
      this.router.navigate(["/account-notice"]);
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
            this.router.navigate(["/account-notice"]);
          } else if (res.approval_status === "pending") {
            this.router.navigate(["/pending-approval"]);
          } else {
            this.router.navigate(["/dashboard"]);
          }
        },
        error: (err) => {
          this.loading = false;
          const errorMsg = err.error?.error || "";
          if (errorMsg.includes("activa") || errorMsg.includes("bloque")) {
            this.router.navigate(["/account-notice"]);
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
