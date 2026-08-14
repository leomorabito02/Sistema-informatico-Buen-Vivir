import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { AvailableAccount } from '../../core/models/auth.model';

@Component({
  selector: 'app-select-role',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './select-role.component.html',
  styleUrl: './select-role.component.css'
})
export class SelectRoleComponent implements OnInit {
  private authService = inject(AuthService);
  private router = inject(Router);

  accounts: AvailableAccount[] = [];
  loginNonce = '';
  loginError = '';
  loadingAccountId = '';

  /** Maps role IDs to icon names for display */
  readonly roleIcons: Record<string, string> = {
    'role_admin': '🛡️',
    'role_dev': '💻',
    'role_participant': '🌱',
    'role_company': '🏢',
    'role_teacher': '📚',
    'role_volunteer': '🤝',
  };

  ngOnInit() {
    const nav = this.router.getCurrentNavigation();
    const state = nav?.extras?.state as { accounts: AvailableAccount[]; loginNonce: string } | undefined;
    const histState = history.state as { accounts?: AvailableAccount[]; loginNonce?: string };
    const accounts = state?.accounts ?? histState?.accounts;
    const loginNonce = state?.loginNonce ?? histState?.loginNonce ?? '';

    if (!accounts || accounts.length === 0 || !loginNonce) {
      this.router.navigate(['/login']);
      return;
    }

    this.accounts = accounts;
    this.loginNonce = loginNonce;
  }

  selectAccount(account: AvailableAccount) {
    if (account.status !== 'active') {
      this.router.navigate(['/account-notice']);
      return;
    }

    this.loadingAccountId = account.user_id;
    this.loginError = '';

    this.authService.loginWithAccount(account.user_id, account.role_id, this.loginNonce).subscribe({
      next: (res) => {
        this.loadingAccountId = '';
        if (res.user_status === 'blocked') {
          this.router.navigate(['/account-notice']);
        } else if (res.approval_status === 'pending') {
          this.router.navigate(['/pending-approval']);
        } else {
          this.router.navigate(['/dashboard']);
        }
      },
      error: (err) => {
        this.loadingAccountId = '';
        const errorMsg = err.error?.error || '';
        if (errorMsg.includes('activa') || errorMsg.includes('bloque')) {
          this.router.navigate(['/account-notice']);
        } else {
          this.loginError = 'Ocurrió un error al iniciar sesión. Intentá nuevamente.';
        }
      }
    });
  }
}
