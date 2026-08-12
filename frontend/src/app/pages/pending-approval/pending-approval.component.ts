import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { FormBuilder, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { LucideAngularModule, Clock, ArrowLeft, Save, FileEdit, CheckCircle } from 'lucide-angular';
import { AuthService } from '../../services/auth.service';

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
  
  private router = inject(Router);
  private authService = inject(AuthService);
  private fb = inject(FormBuilder);

  showForm = false;
  roleId = '';
  profileForm!: FormGroup;
  loading = false;
  successMessage = '';
  errorMessage = '';
  formSaved = false;

  ngOnInit() {
    this.authService.isAuthenticated().subscribe({
      next: (res: any) => {
        this.roleId = res.role_id;
        this.buildForm();
      },
      error: () => {
        this.router.navigate(['/login']);
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
        specialization: ['']
      });
    }
  }

  needsExtraData(): boolean {
    return this.roleId === 'role_participant' || this.roleId === 'role_company' || this.roleId === 'role_teacher';
  }

  toggleForm() {
    this.showForm = !this.showForm;
  }

  onSubmit() {
    if (this.profileForm.invalid) return;

    this.loading = true;
    this.errorMessage = '';
    this.successMessage = '';

    // Convert date string to ISO date format for backend if necessary
    const data = { ...this.profileForm.value };
    if (data.birthdate) {
      data.birthdate = new Date(data.birthdate).toISOString();
    }

    this.authService.updateProfile(data).subscribe({
      next: () => {
        this.loading = false;
        this.formSaved = true;
        this.successMessage = 'Datos guardados correctamente.';
        this.showForm = false;
      },
      error: (err) => {
        this.loading = false;
        this.errorMessage = err.error?.error || 'Ocurrió un error al guardar los datos.';
      }
    });
  }

  goBack() {
    this.router.navigate(['/login']);
  }

  logout() {
    this.authService.logout();
    this.router.navigate(['/login']);
  }
}
