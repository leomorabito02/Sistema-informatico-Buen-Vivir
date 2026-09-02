import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { MiPerfilComponent } from './mi-perfil.component';
import { ProfileService } from '../../../services/profile.service';
import { of, throwError } from 'rxjs';
import { ReactiveFormsModule } from '@angular/forms';
import { Auth, GoogleAuthProvider, signInWithPopup } from '@angular/fire/auth';

describe('MiPerfilComponent', () => {
  let component: MiPerfilComponent;
  let fixture: ComponentFixture<MiPerfilComponent>;
  let profileServiceSpy: jasmine.SpyObj<ProfileService>;
  let mockAuth: any;

  beforeEach(async () => {
    profileServiceSpy = jasmine.createSpyObj('ProfileService', [
      'getMyProfile', 'updateMyPassword', 'linkGoogleAccount', 'unlinkGoogleAccount'
    ]);
    profileServiceSpy.getMyProfile.and.returnValue(of({ user: { email: 'test@test.com' }, roles: [] } as any));

    mockAuth = {};

    await TestBed.configureTestingModule({
      imports: [MiPerfilComponent, ReactiveFormsModule],
      providers: [
        { provide: ProfileService, useValue: profileServiceSpy },
        { provide: Auth, useValue: mockAuth }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(MiPerfilComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should load profile on init', () => {
    expect(profileServiceSpy.getMyProfile).toHaveBeenCalled();
    expect(component.profileData?.user.email).toBe('test@test.com');
    expect(component.loading).toBeFalse();
  });

  it('should format role names', () => {
    expect(component.getRoleName('role_admin')).toBe('Responsable');
    expect(component.getRoleName('role_teacher')).toBe('Docente');
    expect(component.getRoleName('unknown')).toBe('unknown');
    expect(component.getRoleName(undefined)).toBe('');
  });

  it('should format role key labels', () => {
    expect(component.getRoleKeyLabel('approval_status')).toBe('Estado de Aprobación');
    expect(component.getRoleKeyLabel('some_other_key')).toBe('Some Other Key');
  });

  it('should format role values', () => {
    expect(component.formatRoleValue('test', null)).toBe('No especificado');
    expect(component.formatRoleValue('test', true)).toBe('Sí');
    expect(component.formatRoleValue('approval_status', 'pending')).toBe('Pendiente');
    expect(component.formatRoleValue('gender', 'male')).toBe('Masculino');
    expect(component.formatRoleValue('created_at', '2023-01-01T12:00:00Z')).toContain('2023');
  });

  describe('Password Change', () => {
    it('should validate password match', () => {
      component.passwordForm.patchValue({ newPassword: 'password1', confirmPassword: 'password1' });
      expect(component.passwordForm.errors).toBeNull();
      
      component.passwordForm.patchValue({ newPassword: 'password1', confirmPassword: 'password123' });
      expect(component.passwordForm.errors).toEqual({ mismatch: true });
    });

    it('should change password successfully', () => {
      component.passwordForm.patchValue({ oldPassword: 'old', newPassword: 'newPassword', confirmPassword: 'newPassword' });
      profileServiceSpy.updateMyPassword.and.returnValue(of(undefined as any));
      
      component.changePassword();
      
      expect(profileServiceSpy.updateMyPassword).toHaveBeenCalledWith('old', 'newPassword');
      expect(component.passwordSuccess).toContain('correctamente');
    });

    it('should handle password change error', () => {
      component.passwordForm.patchValue({ oldPassword: 'old', newPassword: 'newPassword', confirmPassword: 'newPassword' });
      profileServiceSpy.updateMyPassword.and.returnValue(throwError(() => ({ error: { error: 'Invalid old password' } })));
      
      component.changePassword();
      
      expect(component.passwordError).toBe('Invalid old password');
    });
  });

  describe('Google Link/Unlink', () => {
    it('should unlink google account', () => {
      profileServiceSpy.unlinkGoogleAccount.and.returnValue(of({} as any));
      component.unlinkGoogle();
      
      expect(profileServiceSpy.unlinkGoogleAccount).toHaveBeenCalled();
      expect(component.googleSuccess).toContain('desvinculada correctamente');
      expect(profileServiceSpy.getMyProfile).toHaveBeenCalledTimes(2); // Init + reload
    });

    it('should handle unlink google error', () => {
      profileServiceSpy.unlinkGoogleAccount.and.returnValue(throwError(() => ({ error: { error: 'Unlink Error' } })));
      component.unlinkGoogle();
      
      expect(component.googleError).toBe('Unlink Error');
    });
  });
});
