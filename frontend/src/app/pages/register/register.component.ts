import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators, AbstractControl, ValidationErrors } from '@angular/forms';
import { Router, RouterLink, ActivatedRoute } from '@angular/router';
import { LucideAngularModule, User, Mail, Lock, Phone, CheckCircle2, Clock, AlertCircle } from 'lucide-angular';
import { AuthService } from '../../services/auth.service';

function passwordMatchValidator(control: AbstractControl): ValidationErrors | null {
  const password = control.get('password');
  const confirmPassword = control.get('confirmPassword');
  
  if (password && confirmPassword && password.value !== confirmPassword.value) {
    return { passwordMismatch: true };
  }
  return null;
}

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink, LucideAngularModule],
  templateUrl: './register.component.html',
  styleUrl: './register.component.css'
})
export class RegisterComponent {
  readonly UserIcon = User;
  readonly MailIcon = Mail;
  readonly LockIcon = Lock;
  readonly PhoneIcon = Phone;
  readonly CheckCircle2Icon = CheckCircle2;
  readonly ClockIcon = Clock;
  readonly AlertCircleIcon = AlertCircle;

  private fb = inject(FormBuilder);
  private authService = inject(AuthService);
  private route = inject(ActivatedRoute);

  registerForm: FormGroup;
  registered = false;
  registerError = '';
  loading = false;

  roles = [
    { label: 'Participante', value: 'participant' },
    { label: 'Voluntario', value: 'volunteer' },
    { label: 'Docente', value: 'teacher' },
    { label: 'Empresa', value: 'company' }
  ];

  constructor() {
    this.registerForm = this.fb.group({
      firstName: ['', [Validators.required, Validators.minLength(2)]],
      lastName: ['', [Validators.required, Validators.minLength(2)]],
      email: ['', [Validators.required, Validators.email]],
      phone: ['', [Validators.required]],
      role: ['participant', [Validators.required]],
      // Password must be at least 8 chars long. Can add pattern validator for stronger checks.
      password: ['', [Validators.required, Validators.minLength(8)]],
      confirmPassword: ['', [Validators.required]],
      terms: [false, [Validators.requiredTrue]]
    }, { validators: passwordMatchValidator });
  }

  ngOnInit() {
    this.route.queryParams.subscribe(params => {
      if (params['role']) {
        this.registerForm.patchValue({ role: params['role'] });
      }
    });
  }

  get needsApproval() {
    const role = this.registerForm.get('role')?.value;
    return role === 'volunteer' || role === 'teacher' || role === 'company';
  }

  onSubmit() {
    if (this.registerForm.invalid) {
      this.registerForm.markAllAsTouched();
      return;
    }

    this.loading = true;
    this.registerError = '';
    
    // Only send backend-required fields
    const formData = {
      first_name: this.registerForm.value.firstName,
      last_name: this.registerForm.value.lastName,
      email: this.registerForm.value.email,
      phone: this.registerForm.value.phone,
      password: this.registerForm.value.password,
      role: this.registerForm.value.role
    };

    this.authService.registerWithEmail(formData).subscribe({
      next: (res) => {
        this.registered = true;
        this.loading = false;
      },
      error: (err) => {
        this.loading = false;
        // Generic error message as per security best practices
        this.registerError = 'Ocurrió un error al registrarse. Revise los datos o intente nuevamente.';
        console.error('Registration error', err);
      }
    });
  }
}
