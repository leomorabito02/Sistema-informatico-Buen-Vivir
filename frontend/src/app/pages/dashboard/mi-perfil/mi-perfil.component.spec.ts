import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MiPerfilComponent } from './mi-perfil.component';
import { ProfileService } from '../../../services/profile.service';
import { ReactiveFormsModule, FormGroup, FormControl } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { Auth } from '@angular/fire/auth';
import { of, throwError } from 'rxjs';

describe('MiPerfilComponent', () => {
  let component: MiPerfilComponent;
  let fixture: ComponentFixture<MiPerfilComponent>;
  let mockProfileService: jasmine.SpyObj<ProfileService>;

  beforeEach(async () => {
    mockProfileService = jasmine.createSpyObj('ProfileService', [
      'getMyProfile',
      'updateMyPassword',
      'linkGoogleAccount',
      'unlinkGoogleAccount'
    ]);

    mockProfileService.getMyProfile.and.returnValue(of({
      user: { id: 'usr-1', email: 'test@test.com', first_name: 'John', last_name: 'Doe' }
    } as any));

    await TestBed.configureTestingModule({
      imports: [MiPerfilComponent, ReactiveFormsModule, BrowserAnimationsModule],
      providers: [
        { provide: ProfileService, useValue: mockProfileService },
        { provide: Auth, useValue: {} }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(MiPerfilComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create and load profile with correct initial default states', () => {
    expect(component).toBeTruthy();
    expect(component.profileData).toBeDefined();
    expect(component.loading).toBeFalse();
    expect(component.error).toBe('');
    expect(component.googleLoading).toBeFalse();
    expect(component.googleSuccess).toBe('');
    expect(component.googleError).toBe('');
    expect(component.passwordLoading).toBeFalse();
    expect(component.passwordSuccess).toBe('');
    expect(component.passwordError).toBe('');
    expect(component.showOldPassword).toBeFalse();
    expect(component.showNewPassword).toBeFalse();
    expect(component.showConfirmPassword).toBeFalse();
  });

  it('should set error state if loadProfile fails', () => {
    mockProfileService.getMyProfile.and.returnValue(throwError(() => new Error('Load failed')));
    component.loadProfile();
    expect(component.error).toBe('Error al cargar el perfil');
    expect(component.loading).toBeFalse();
  });

  describe('getRoleName', () => {
    it('should return correct role label for known and unknown roles', () => {
      expect(component.getRoleName('')).toBe('');
      expect(component.getRoleName('role_participant')).toBe('Participante');
      expect(component.getRoleName('role_company')).toBe('Empresa');
      expect(component.getRoleName('role_teacher')).toBe('Docente');
      expect(component.getRoleName('role_volunteer')).toBe('Voluntario');
      expect(component.getRoleName('role_admin')).toBe('Responsable');
      expect(component.getRoleName('role_dev')).toBe('Desarrollador');
      expect(component.getRoleName('participant')).toBe('Participante');
      expect(component.getRoleName('company')).toBe('Empresa');
      expect(component.getRoleName('teacher')).toBe('Docente');
      expect(component.getRoleName('volunteer')).toBe('Voluntario');
      expect(component.getRoleName('admin')).toBe('Responsable');
      expect(component.getRoleName('dev')).toBe('Desarrollador');
      expect(component.getRoleName('custom_role')).toBe('custom_role');
    });
  });

  describe('getRoleKeyLabel', () => {
    it('should return label from map or format snake_case key', () => {
      expect(component.getRoleKeyLabel('volunteer_id')).toBe('ID de Voluntario');
      expect(component.getRoleKeyLabel('participant_id')).toBe('ID de Participante');
      expect(component.getRoleKeyLabel('teacher_id')).toBe('ID de Docente');
      expect(component.getRoleKeyLabel('company_id')).toBe('ID de Empresa');
      expect(component.getRoleKeyLabel('approval_status')).toBe('Estado de Aprobación');
      expect(component.getRoleKeyLabel('approval_date')).toBe('Fecha de Aprobación');
      expect(component.getRoleKeyLabel('approved_by')).toBe('Aprobado Por');
      expect(component.getRoleKeyLabel('created_at')).toBe('Fecha de Creación');
      expect(component.getRoleKeyLabel('updated_at')).toBe('Última Actualización');
      expect(component.getRoleKeyLabel('deleted_at')).toBe('Fecha de Eliminación');
      expect(component.getRoleKeyLabel('specialization')).toBe('Especialidad');
      expect(component.getRoleKeyLabel('is_external')).toBe('Docente Externo');
      expect(component.getRoleKeyLabel('business_name')).toBe('Razón Social');
      expect(component.getRoleKeyLabel('cuit')).toBe('CUIT');
      expect(component.getRoleKeyLabel('birthdate')).toBe('Fecha de Nacimiento');
      expect(component.getRoleKeyLabel('dni')).toBe('DNI');
      expect(component.getRoleKeyLabel('address')).toBe('Dirección');
      expect(component.getRoleKeyLabel('gender')).toBe('Género');
      expect(component.getRoleKeyLabel('image_url')).toBe('Imagen de Perfil');
      expect(component.getRoleKeyLabel('doc_url')).toBe('Documentación');
      expect(component.getRoleKeyLabel('emergency_contact_name')).toBe('Contacto de Emergencia');
      expect(component.getRoleKeyLabel('emergency_contact_phone')).toBe('Teléfono de Emergencia');
      expect(component.getRoleKeyLabel('custom_field_name')).toBe('Custom Field Name');
    });
  });

  describe('formatRoleValue', () => {
    it('should format null/undefined/empty as No especificado', () => {
      expect(component.formatRoleValue('dni', null)).toBe('No especificado');
      expect(component.formatRoleValue('dni', undefined)).toBe('No especificado');
      expect(component.formatRoleValue('dni', '')).toBe('No especificado');
    });

    it('should format booleans', () => {
      expect(component.formatRoleValue('is_external', true)).toBe('Sí');
      expect(component.formatRoleValue('is_external', false)).toBe('No');
    });

    it('should format approval_status', () => {
      expect(component.formatRoleValue('approval_status', 'approved')).toBe('Aprobado');
      expect(component.formatRoleValue('approval_status', 'pending')).toBe('Pendiente');
      expect(component.formatRoleValue('approval_status', 'rejected')).toBe('Rechazado');
      expect(component.formatRoleValue('approval_status', 'other')).toBe('other');
    });

    it('should format gender', () => {
      expect(component.formatRoleValue('gender', 'male')).toBe('Masculino');
      expect(component.formatRoleValue('gender', 'female')).toBe('Femenino');
      expect(component.formatRoleValue('gender', 'other')).toBe('Otro');
      expect(component.formatRoleValue('gender', 'unknown')).toBe('unknown');
    });

    it('should format dates ending with _at or date', () => {
      const atFormatted = component.formatRoleValue('created_at', '2023-10-15T00:00:00Z');
      expect(atFormatted).toContain('2023');

      const dateFormatted = component.formatRoleValue('birthdate', '1995-05-15T00:00:00Z');
      expect(dateFormatted).toContain('1995');

      const invalidDate = component.formatRoleValue('created_at', 'not-a-date');
      expect(invalidDate).toBe('not-a-date');
    });

    it('should return string representation for generic keys', () => {
      expect(component.formatRoleValue('dni', '12345678')).toBe('12345678');
    });
  });

  describe('passwordMatchValidator', () => {
    it('should validate matching passwords', () => {
      const group = new FormGroup({
        newPassword: new FormControl('pass123'),
        confirmPassword: new FormControl('pass123')
      });
      expect(component.passwordMatchValidator(group)).toBeNull();
    });

    it('should return mismatch error for mismatching passwords', () => {
      const group = new FormGroup({
        newPassword: new FormControl('pass123'),
        confirmPassword: new FormControl('pass456')
      });
      expect(component.passwordMatchValidator(group)).toEqual({ mismatch: true });
    });
  });

  describe('togglePasswordVisibility', () => {
    it('should toggle visibility flags for old, new, and confirm fields', () => {
      component.togglePasswordVisibility('old');
      expect(component.showOldPassword).toBeTrue();
      component.togglePasswordVisibility('old');
      expect(component.showOldPassword).toBeFalse();

      component.togglePasswordVisibility('new');
      expect(component.showNewPassword).toBeTrue();
      component.togglePasswordVisibility('new');
      expect(component.showNewPassword).toBeFalse();

      component.togglePasswordVisibility('confirm');
      expect(component.showConfirmPassword).toBeTrue();
      component.togglePasswordVisibility('confirm');
      expect(component.showConfirmPassword).toBeFalse();
    });
  });

  describe('changePassword', () => {
    it('should validate password match before calling service', () => {
      component.passwordForm.patchValue({
        oldPassword: 'OldPassword123!',
        newPassword: 'NewPassword123!',
        confirmPassword: 'DifferentPassword123!'
      });
      expect(component.passwordForm.invalid).toBeTrue();

      component.changePassword();
      expect(mockProfileService.updateMyPassword).not.toHaveBeenCalled();
    });

    it('should call updateMyPassword and handle success', () => {
      mockProfileService.updateMyPassword.and.returnValue(of(undefined as any));

      component.passwordForm.patchValue({
        oldPassword: 'OldPassword123!',
        newPassword: 'NewPassword123!',
        confirmPassword: 'NewPassword123!'
      });

      component.changePassword();

      expect(mockProfileService.updateMyPassword).toHaveBeenCalledWith('OldPassword123!', 'NewPassword123!');
      expect(component.passwordSuccess).toBe('Contraseña actualizada correctamente.');
      expect(component.passwordLoading).toBeFalse();
    });

    it('should handle updateMyPassword error', () => {
      mockProfileService.updateMyPassword.and.returnValue(throwError(() => ({ error: { error: 'Wrong password' } })));

      component.passwordForm.patchValue({
        oldPassword: 'WrongPassword123!',
        newPassword: 'NewPassword123!',
        confirmPassword: 'NewPassword123!'
      });

      component.changePassword();

      expect(component.passwordError).toBe('Wrong password');
      expect(component.passwordLoading).toBeFalse();
    });

    it('should handle updateMyPassword error with default message', () => {
      mockProfileService.updateMyPassword.and.returnValue(throwError(() => ({ error: {} })));

      component.passwordForm.patchValue({
        oldPassword: 'WrongPassword123!',
        newPassword: 'NewPassword123!',
        confirmPassword: 'NewPassword123!'
      });

      component.changePassword();

      expect(component.passwordError).toBe('Error al actualizar la contraseña.');
    });
  });

  describe('unlinkGoogle', () => {
    it('should call unlinkGoogleAccount and handle success', () => {
      mockProfileService.unlinkGoogleAccount.and.returnValue(of({} as any));

      component.unlinkGoogle();

      expect(mockProfileService.unlinkGoogleAccount).toHaveBeenCalled();
      expect(component.googleSuccess).toBe('Cuenta de Google desvinculada correctamente.');
      expect(component.googleLoading).toBeFalse();
    });

    it('should handle unlinkGoogleAccount error', () => {
      mockProfileService.unlinkGoogleAccount.and.returnValue(throwError(() => ({ error: { error: 'Unlink error' } })));

      component.unlinkGoogle();

      expect(component.googleError).toBe('Unlink error');
      expect(component.googleLoading).toBeFalse();
    });

    it('should handle unlinkGoogleAccount error with default message', () => {
      mockProfileService.unlinkGoogleAccount.and.returnValue(throwError(() => ({ error: {} })));

      component.unlinkGoogle();

      expect(component.googleError).toBe('Error al desvincular la cuenta de Google.');
    });
  });
});
