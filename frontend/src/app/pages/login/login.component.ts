import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { LucideAngularModule, Mail, Lock, Shield, AlertCircle } from 'lucide-angular';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink, LucideAngularModule],
  templateUrl: './login.component.html'
})
export class LoginComponent {
  readonly MailIcon = Mail;
  readonly LockIcon = Lock;
  readonly ShieldIcon = Shield;
  readonly AlertCircleIcon = AlertCircle;

  private fb = inject(FormBuilder);
  private authService = inject(AuthService);
  private router = inject(Router);

  loginForm: FormGroup;
  selectedRole = 'participante';
  loginError = '';
  locked = false;
  failedAttempts = 0;
  MAX_ATTEMPTS = 5;

  roles = [
    { label: 'Responsable (ve todo)', value: 'admin' },
    { label: 'Voluntario (seguimiento y campañas)', value: 'voluntario' },
    { label: 'Participante (cursos y CV)', value: 'participante' },
    { label: 'Docente (subir contenido)', value: 'docente' },
    { label: 'Empresa', value: 'empresa' }
  ];

  constructor() {
    this.loginForm = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
      password: ['', [Validators.required, Validators.minLength(4)]]
    });
  }

  onRoleChange(event: any) {
    this.selectedRole = event.target.value;
  }

  onSubmit() {
    if (this.locked) return;

    if (this.loginForm.invalid) {
      this.handleFailedAttempt();
      return;
    }

    const { email, password } = this.loginForm.value;
    
    this.authService.loginWithEmail(email, password, this.selectedRole).subscribe({
      next: (res) => {
        this.loginError = '';
        this.router.navigate(['/dashboard']);
      },
      error: (err) => {
        this.handleFailedAttempt();
      }
    });
  }

  handleFailedAttempt() {
    this.failedAttempts++;
    if (this.failedAttempts >= this.MAX_ATTEMPTS) {
      this.locked = true;
      this.loginError = 'Tu cuenta fue bloqueada temporalmente por demasiados intentos fallidos. Intentá nuevamente más tarde.';
    } else {
      this.loginError = 'Email o contraseña incorrectos.';
    }
  }

  handleGoogleLogin() {
    this.authService.loginWithGoogle(this.selectedRole).subscribe({
      next: (res) => {
        this.loginError = '';
        this.router.navigate(['/dashboard']);
      },
      error: (err) => {
        this.loginError = 'Hubo un error al iniciar sesión con Google.';
      }
    });
  }
}
