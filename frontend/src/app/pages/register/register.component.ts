import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators, AbstractControl, ValidationErrors } from '@angular/forms';
import { RouterLink, ActivatedRoute } from '@angular/router';
import { LucideAngularModule, User, Mail, Lock, Phone, CheckCircle2, Clock, AlertCircle, Eye, EyeOff } from 'lucide-angular';
import { AuthService } from '../../services/auth.service';

function passwordMatchValidator(control: AbstractControl): ValidationErrors | null {
  const password = control.get('password');
  const confirmPassword = control.get('confirmPassword');
  
  if (password && confirmPassword && password.value !== confirmPassword.value) {
    return { passwordMismatch: true };
  }
  return null;
}

function passwordComplexityValidator(control: AbstractControl): ValidationErrors | null {
  const value = control.value || '';
  if (!value) return null;

  if (value.length < 8) {
    return { minlength: { requiredLength: 8, actualLength: value.length } };
  }
  if (!/[A-Z]/.test(value)) {
    return { missingUpper: true };
  }
  if (!/[a-z]/.test(value)) {
    return { missingLower: true };
  }
  if (!/[^a-zA-Z0-9]/.test(value)) {
    return { missingSpecial: true };
  }
  return null;
}

import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { InputTextModule } from 'primeng/inputtext';
import { DropdownModule } from 'primeng/dropdown';
import { MessageModule } from 'primeng/message';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink, LucideAngularModule, ButtonModule, CardModule, InputTextModule, DropdownModule, MessageModule],
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
  readonly EyeIcon = Eye;
  readonly EyeOffIcon = EyeOff;

  private fb = inject(FormBuilder);
  private authService = inject(AuthService);
  private route = inject(ActivatedRoute);

  registerForm: FormGroup;
  registered = false;
  registerError = '';
  loading = false;
  showPassword = false;
  showConfirmPassword = false;

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
      password: ['', [Validators.required, passwordComplexityValidator]],
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

  get needsApproval(): boolean {
    const role = this.registerForm.controls['role'].value;
    return role === 'volunteer' || role === 'teacher' || role === 'company';
  }

  private getFirstFormError(): string {
    const c = this.registerForm.controls;
    const reqFields = [c['firstName'], c['lastName'], c['email'], c['phone'], c['password']];
    if (reqFields.some(ctrl => ctrl.hasError('required'))) {
      return 'Por favor, completá todos los campos obligatorios.';
    }
    if (c['email'].hasError('email')) {
      return 'Por favor, ingresá una dirección de correo electrónico válida.';
    }
    if (c['password'].hasError('minlength')) {
      return 'La contraseña debe tener al menos 8 caracteres.';
    }
    if (c['password'].hasError('missingUpper')) {
      return 'La contraseña debe incluir al menos una letra mayúscula.';
    }
    if (c['password'].hasError('missingLower')) {
      return 'La contraseña debe incluir al menos una letra minúscula.';
    }
    if (c['password'].hasError('missingSpecial')) {
      return 'La contraseña debe incluir al menos un carácter especial (ej: !@#$%^&*).';
    }
    if (this.registerForm.hasError('passwordMismatch')) {
      return 'Las contraseñas no coinciden.';
    }
    if (c['terms'].hasError('required')) {
      return 'Debés aceptar los términos y condiciones para continuar.';
    }
    return 'Por favor, revisá los campos del formulario.';
  }

  onSubmit() {
    if (this.registerForm.invalid) {
      this.registerForm.markAllAsTouched();
      this.registerError = this.getFirstFormError();
      return;
    }

    this.loading = true;
    this.registerError = '';
    
    const formData = {
      first_name: this.registerForm.value.firstName,
      last_name: this.registerForm.value.lastName,
      email: this.registerForm.value.email,
      phone: this.registerForm.value.phone,
      password: this.registerForm.value.password,
      role: this.registerForm.value.role
    };

    this.authService.registerWithEmail(formData).subscribe({
      next: () => {
        this.registered = true;
        this.loading = false;
      },
      error: (err) => {
        this.loading = false;
        const apiError = err.error?.error;
        this.registerError = apiError || 'Ocurrió un error al procesar el registro. Verificá los datos e intentá nuevamente.';
        console.error('Registration error', err);
      }
    });
  }

  togglePasswordVisibility(): void {
    this.showPassword = !this.showPassword;
  }

  toggleConfirmPasswordVisibility(): void {
    this.showConfirmPassword = !this.showConfirmPassword;
  }
}
