import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Auth, signInWithPopup, GoogleAuthProvider } from '@angular/fire/auth';
import { from, switchMap } from 'rxjs';
import { ProfileService } from '../../../services/profile.service';
import { User, ProfileResponse } from '../../../core/models/auth.model';
import { LucideAngularModule, User as UserIcon, Shield, Briefcase, Lock, CheckCircle, AlertCircle, Eye, EyeOff, Link, Unlink } from 'lucide-angular';

@Component({
  selector: 'app-mi-perfil',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, LucideAngularModule],
  templateUrl: './mi-perfil.component.html',
  styleUrls: ['./mi-perfil.component.css']
})
export class MiPerfilComponent implements OnInit {
  private fb = inject(FormBuilder);
  private profileService = inject(ProfileService);
  private auth = inject(Auth);

  readonly UserIcon = UserIcon;
  readonly ShieldIcon = Shield;
  readonly BriefcaseIcon = Briefcase;
  readonly LockIcon = Lock;
  readonly CheckCircleIcon = CheckCircle;
  readonly AlertCircleIcon = AlertCircle;
  readonly EyeIcon = Eye;
  readonly EyeOffIcon = EyeOff;
  readonly LinkIcon = Link;
  readonly UnlinkIcon = Unlink;

  profileData: ProfileResponse | null = null;
  loading = true;
  error = '';

  googleLoading = false;
  googleSuccess = '';
  googleError = '';

  passwordForm: FormGroup;
  passwordLoading = false;
  passwordSuccess = '';
  passwordError = '';

  showOldPassword = false;
  showNewPassword = false;
  showConfirmPassword = false;

  constructor() {
    this.passwordForm = this.fb.group({
      oldPassword: ['', Validators.required],
      newPassword: ['', [Validators.required, Validators.minLength(8)]],
      confirmPassword: ['', Validators.required]
    }, { validators: this.passwordMatchValidator });
  }

  ngOnInit() {
    this.loadProfile();
  }

  loadProfile() {
    this.profileService.getMyProfile().subscribe({
      next: (data) => {
        this.profileData = data;
        this.loading = false;
      },
      error: (_err) => {
        this.error = 'Error al cargar el perfil';
        this.loading = false;
      }
    });
  }

  getRoleName(roleId?: string): string {
    if (!roleId) return '';
    const roleMap: Record<string, string> = {
      'role_participant': 'Participante',
      'role_company': 'Empresa',
      'role_teacher': 'Docente',
      'role_volunteer': 'Voluntario',
      'role_admin': 'Responsable',
      'role_dev': 'Desarrollador',
      'participant': 'Participante',
      'company': 'Empresa',
      'teacher': 'Docente',
      'volunteer': 'Voluntario',
      'admin': 'Responsable',
      'dev': 'Desarrollador'
    };
    return roleMap[roleId.toLowerCase()] || roleId;
  }

  getRoleKeyLabel(key: string): string {
    const k = String(key);
    const labelMap: Record<string, string> = {
      'volunteer_id': 'ID de Voluntario',
      'participant_id': 'ID de Participante',
      'teacher_id': 'ID de Docente',
      'company_id': 'ID de Empresa',
      'approval_status': 'Estado de Aprobación',
      'approval_date': 'Fecha de Aprobación',
      'approved_by': 'Aprobado Por',
      'created_at': 'Fecha de Creación',
      'updated_at': 'Última Actualización',
      'deleted_at': 'Fecha de Eliminación',
      'specialization': 'Especialidad',
      'is_external': 'Docente Externo',
      'business_name': 'Razón Social',
      'cuit': 'CUIT',
      'birthdate': 'Fecha de Nacimiento',
      'dni': 'DNI',
      'address': 'Dirección',
      'gender': 'Género',
      'image_url': 'Imagen de Perfil',
      'doc_url': 'Documentación',
      'emergency_contact_name': 'Contacto de Emergencia',
      'emergency_contact_phone': 'Teléfono de Emergencia'
    };
    return labelMap[k] || k.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
  }

  formatRoleValue(key: string, value: string | boolean | null | undefined): string {
    const k = key;
    if (value === null || value === undefined || value === '') return 'No especificado';
    if (typeof value === 'boolean') return value ? 'Sí' : 'No';

    if (k === 'approval_status') {
      const statusMap: Record<string, string> = {
        'approved': 'Aprobado',
        'pending': 'Pendiente',
        'rejected': 'Rechazado'
      };
      return statusMap[value] || value;
    }

    if (k === 'gender') {
      const genderMap: Record<string, string> = {
        'male': 'Masculino',
        'female': 'Femenino',
        'other': 'Otro'
      };
      return genderMap[value] || value;
    }

    if ((k.endsWith('_at') || k.endsWith('date')) && typeof value === 'string' && !isNaN(Date.parse(value))) {
      try {
        return new Date(value).toLocaleDateString('es-AR', {
          year: 'numeric',
          month: 'long',
          day: 'numeric'
        });
      } catch {
        return value;
      }
    }

    return String(value);
  }

  passwordMatchValidator(g: FormGroup) {
    return g.get('newPassword')?.value === g.get('confirmPassword')?.value
      ? null : { mismatch: true };
  }

  togglePasswordVisibility(field: 'old' | 'new' | 'confirm') {
    if (field === 'old') this.showOldPassword = !this.showOldPassword;
    if (field === 'new') this.showNewPassword = !this.showNewPassword;
    if (field === 'confirm') this.showConfirmPassword = !this.showConfirmPassword;
  }

  changePassword() {
    if (this.passwordForm.invalid) return;

    this.passwordLoading = true;
    this.passwordSuccess = '';
    this.passwordError = '';

    const { oldPassword, newPassword } = this.passwordForm.value;

    this.profileService.updateMyPassword(oldPassword, newPassword).subscribe({
      next: () => {
        this.passwordSuccess = 'Contraseña actualizada correctamente.';
        this.passwordForm.reset();
        this.passwordLoading = false;
      },
      error: (err) => {
        this.passwordError = err.error?.error || 'Error al actualizar la contraseña.';
        this.passwordLoading = false;
      }
    });
  }

  linkGoogle() {
    this.googleLoading = true;
    this.googleSuccess = '';
    this.googleError = '';

    const provider = new GoogleAuthProvider();
    from(signInWithPopup(this.auth, provider)).pipe(
      switchMap(credential => from(credential.user.getIdToken().then(idToken => ({ idToken, googleId: credential.user.uid })))),
      switchMap(({ idToken, googleId }) => this.profileService.linkGoogleAccount(idToken, googleId))
    ).subscribe({
      next: () => {
        this.googleSuccess = 'Cuenta de Google vinculada correctamente.';
        this.googleLoading = false;
        this.loadProfile();
      },
      error: (err) => {
        this.googleError = err.error?.error || 'Error al vincular la cuenta de Google.';
        this.googleLoading = false;
      }
    });
  }

  unlinkGoogle() {
    this.googleLoading = true;
    this.googleSuccess = '';
    this.googleError = '';

    this.profileService.unlinkGoogleAccount().subscribe({
      next: () => {
        this.googleSuccess = 'Cuenta de Google desvinculada correctamente.';
        this.googleLoading = false;
        this.loadProfile();
      },
      error: (err) => {
        this.googleError = err.error?.error || 'Error al desvincular la cuenta de Google.';
        this.googleLoading = false;
      }
    });
  }
}
