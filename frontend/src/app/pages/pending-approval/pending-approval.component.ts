import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { FormBuilder, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { LucideAngularModule, Clock, ArrowLeft, Save, FileEdit, CheckCircle, AlertCircle, LogOut } from 'lucide-angular';
import { AuthService } from '../../services/auth.service';
import { ProfileService } from '../../services/profile.service';
import { MeResponse } from '../../core/models/auth.model';

@Component({
  selector: 'app-pending-approval',
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule, ReactiveFormsModule],
  templateUrl: './pending-approval.component.html',
  styleUrl: './pending-approval.component.css'
})
export class PendingApprovalComponent implements OnInit {
  readonly ClockIcon = Clock;
  readonly ArrowLeftIcon = ArrowLeft;
  readonly SaveIcon = Save;
  readonly FileEditIcon = FileEdit;
  readonly CheckCircleIcon = CheckCircle;
  readonly AlertCircleIcon = AlertCircle;
  readonly LogOutIcon = LogOut;

  private router = inject(Router);
  private authService = inject(AuthService);
  private profileService = inject(ProfileService);
  private fb = inject(FormBuilder);

  showForm = false;
  roleId = '';
  profileForm!: FormGroup;
  loading = false;
  successMessage = '';
  errorMessage = '';
  formSaved = false;

  roleLabels: Record<string, string> = {
    'role_participant': 'Participante',
    'role_company': 'Empresa',
    'role_teacher': 'Docente',
    'role_volunteer': 'Voluntario',
    'role_admin': 'Responsable',
    'role_dev': 'Desarrollador'
  };

  ngOnInit() {
    this.authService.isAuthenticated().subscribe({
      next: (res: MeResponse) => {
        this.roleId = res.role_id ?? res.user?.role_id ?? '';
        this.buildForm();
        this.loadExistingData();
      },
      error: () => {
        this.router.navigate(['/login']);
      }
    });
  }

  loadExistingData() {
    this.profileService.getMyProfile().subscribe({
      next: (data) => {
        if (data && data.role_data) {
          const roleData = { ...data.role_data };
          if (roleData.birthdate) {
            roleData.birthdate = roleData.birthdate.split('T')[0];
          }
          this.profileForm.patchValue(roleData);
        }
      },
      error: () => {
        // Ignore error, form stays empty
      }
    });
  }

  buildForm() {
    if (this.roleId === 'role_participant') {
      this.profileForm = this.fb.group({
        birthdate: [''],
        dni: [''],
        address: [''],
        gender: [''],
        emergency_contact_name: [''],
        emergency_contact_phone: ['']
      });
    } else if (this.roleId === 'role_company') {
      this.profileForm = this.fb.group({
        business_name: [''],
        cuit: ['']
      });
    } else if (this.roleId === 'role_teacher') {
      this.profileForm = this.fb.group({
        specialization: [''],
        is_external: [false]
      });
    } else if (this.roleId === 'role_volunteer') {
      this.profileForm = this.fb.group({
        phone: [''],
        notes: ['']
      });
    } else {
      this.profileForm = this.fb.group({
        phone: [''],
        notes: ['']
      });
    }
  }

  getRoleName(): string {
    return this.roleLabels[this.roleId] || 'Usuario';
  }

  needsExtraData(): boolean {
    return true;
  }

  toggleForm() {
    this.showForm = !this.showForm;
  }

  onSubmit() {
    if (!this.profileForm || this.profileForm.invalid) return;

    this.loading = true;
    this.errorMessage = '';
    this.successMessage = '';

    const data = { ...this.profileForm.value };
    if (data.birthdate) {
      data.birthdate = new Date(data.birthdate).toISOString();
    }

    this.authService.updateProfile(data).subscribe({
      next: () => {
        this.loading = false;
        this.formSaved = true;
        this.successMessage = 'Datos del perfil actualizados correctamente.';
        this.showForm = false;
      },
      error: (err) => {
        this.loading = false;
        this.errorMessage = err.error?.error || 'Ocurrió un error al guardar los datos.';
      }
    });
  }

  goBack() {
    this.router.navigate(['/']);
  }

  logout() {
    this.authService.logout();
    this.router.navigate(['/login']);
  }
}
