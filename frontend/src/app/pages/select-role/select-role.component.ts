import { Component, inject, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import { Router, RouterLink } from "@angular/router";
import { AuthService } from "../../services/auth.service";
import { AvailableAccount } from "../../core/models/auth.model";
import {
  LucideAngularModule,
  BookOpen,
  Building2,
  CircleAlert,
  Code,
  Handshake,
  ShieldCheck,
  Sprout,
  User,
} from "lucide-angular";

@Component({
  selector: "app-select-role",
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule],
  templateUrl: "./select-role.component.html",
  styleUrl: "./select-role.component.css",
})
export class SelectRoleComponent implements OnInit {
  private authService = inject(AuthService);
  private router = inject(Router);

  accounts!: AvailableAccount[];
  loginNonce!: string;
  loginError = "";
  loadingAccountId = "";

  /** Ícono (Lucide) de cada rol en la lista de cuentas */
  readonly roleIcons: Record<string, typeof User> = {
    role_admin: ShieldCheck,
    role_dev: Code,
    role_participant: Sprout,
    role_company: Building2,
    role_teacher: BookOpen,
    role_volunteer: Handshake,
  };
  readonly defaultRoleIcon = User;
  readonly CircleAlert = CircleAlert;

  ngOnInit() {
    const nav = this.router.getCurrentNavigation();
    const state = nav?.extras?.state as
      { accounts: AvailableAccount[]; loginNonce: string } | undefined;
    const histState = history.state as {
      accounts?: AvailableAccount[];
      loginNonce?: string;
    };
    const accounts = state?.accounts ?? histState?.accounts;
    const loginNonce = state?.loginNonce ?? histState?.loginNonce ?? "";

    if (!accounts || accounts.length === 0 || !loginNonce) {
      this.router.navigate(["/login"]);
      return;
    }

    this.accounts = accounts;
    this.loginNonce = loginNonce;
  }

  selectAccount(account: AvailableAccount) {
    if (account.status !== "active") {
      this.router.navigate(["/account-notice"]);
      return;
    }

    this.loadingAccountId = account.user_id;
    this.loginError = "";

    this.authService
      .loginWithAccount(account.user_id, account.role_id, this.loginNonce)
      .subscribe({
        next: (res) => {
          this.loadingAccountId = "";
          if (res.user_status === "blocked") {
            this.router.navigate(["/account-notice"]);
          } else if (res.approval_status === "pending") {
            this.router.navigate(["/pending-approval"]);
          } else {
            this.router.navigate(["/dashboard"]);
          }
        },
        error: (err) => {
          this.loadingAccountId = "";
          const errorMsg = err.error?.error || err.message || "";
          if (errorMsg === "EMAIL_NOT_VERIFIED" || errorMsg.includes("verificar tu correo")) {
            void this.router.navigate(["/verify-email"]);
            return;
          }
          if (errorMsg.includes("activa") || errorMsg.includes("bloque")) {
            void this.router.navigate(["/account-notice"]);
          } else {
            this.loginError =
              "Ocurrió un error al iniciar sesión. Intentá nuevamente.";
          }
        },
      });
  }
}
