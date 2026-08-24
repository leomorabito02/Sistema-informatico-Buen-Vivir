import { ComponentFixture, TestBed } from '@angular/core/testing';
import { LoginComponent } from './login.component';
import { ReactiveFormsModule } from '@angular/forms';
import { Router, provideRouter } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { of, throwError } from 'rxjs';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

describe('LoginComponent', () => {
  let component: LoginComponent;
  let fixture: ComponentFixture<LoginComponent>;
  let mockAuthService: jasmine.SpyObj<AuthService>;
  let mockRouter: any;

  beforeEach(async () => {
    mockAuthService = jasmine.createSpyObj('AuthService', ['checkCredentials', 'checkGoogleCredentials', 'loginWithAccount']);

    await TestBed.configureTestingModule({
      imports: [LoginComponent, ReactiveFormsModule, BrowserAnimationsModule],
      providers: [
        { provide: AuthService, useValue: mockAuthService },
        provideRouter([])
      ]
    }).compileComponents();
    
    mockRouter = TestBed.inject(Router);
    spyOn(mockRouter, 'navigate');

    fixture = TestBed.createComponent(LoginComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should toggle password visibility', () => {
    expect(component.showPassword).toBeFalse();
    component.togglePasswordVisibility();
    expect(component.showPassword).toBeTrue();
  });

  describe('onSubmit validations', () => {
    it('should set error when fields are empty', () => {
      component.onSubmit();
      expect(component.loginError).toBe('Por favor, ingresá tu email y contraseña.');
    });

    it('should set error for invalid email', () => {
      component.loginForm.controls['email'].setValue('invalid');
      component.loginForm.controls['password'].setValue('1234');
      component.onSubmit();
      expect(component.loginError).toBe('Por favor, ingresá un correo electrónico válido.');
    });

    it('should set error for short password', () => {
      component.loginForm.controls['email'].setValue('test@test.com');
      component.loginForm.controls['password'].setValue('123');
      component.onSubmit();
      expect(component.loginError).toBe('La contraseña debe tener al menos 4 caracteres.');
    });
  });

  describe('onSubmit success', () => {
    beforeEach(() => {
      component.loginForm.controls['email'].setValue('test@test.com');
      component.loginForm.controls['password'].setValue('1234');
    });

    it('should handle missing accounts', () => {
      mockAuthService.checkCredentials.and.returnValue(of({ accounts: [], login_nonce: 'nonce' }));
      component.onSubmit();
      expect(component.loginError).toBe('No se encontraron cuentas activas asociadas a este correo electrónico.');
    });

    it('should navigate to select-role if multiple accounts exist', () => {
      const accounts = [
        { user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'active' },
        { user_id: '1', role_id: 'r2', role_name: 'Role 2', status: 'active' }
      ];
      mockAuthService.checkCredentials.and.returnValue(of({ accounts, login_nonce: 'nonce' }));
      component.onSubmit();
      expect(mockRouter.navigate).toHaveBeenCalledWith(['/select-role'], { state: { accounts, loginNonce: 'nonce' } });
    });

    it('should login directly if exactly one account exists and is active', () => {
      const account = { user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'active' };
      mockAuthService.checkCredentials.and.returnValue(of({ accounts: [account], login_nonce: 'nonce' }));
      mockAuthService.loginWithAccount.and.returnValue(of({ token: 'jwt', user: {} as any }));
      
      component.onSubmit();
      
      expect(mockAuthService.loginWithAccount).toHaveBeenCalledWith('1', 'r1', 'nonce');
      expect(mockRouter.navigate).toHaveBeenCalledWith(['/dashboard']);
    });

    it('should redirect to account-notice if account status is not active before login', () => {
      const account = { user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'inactive' };
      mockAuthService.checkCredentials.and.returnValue(of({ accounts: [account], login_nonce: 'nonce' }));
      
      component.onSubmit();
      
      expect(mockRouter.navigate).toHaveBeenCalledWith(['/account-notice']);
    });
    
    it('should redirect to account-notice if login response returns blocked status', () => {
      const account = { user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'active' };
      mockAuthService.checkCredentials.and.returnValue(of({ accounts: [account], login_nonce: 'nonce' }));
      mockAuthService.loginWithAccount.and.returnValue(of({ token: 'jwt', user: {} as any, user_status: 'blocked' }));
      
      component.onSubmit();
      
      expect(mockRouter.navigate).toHaveBeenCalledWith(['/account-notice']);
    });

    it('should redirect to pending-approval if login response returns pending status', () => {
      const account = { user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'active' };
      mockAuthService.checkCredentials.and.returnValue(of({ accounts: [account], login_nonce: 'nonce' }));
      mockAuthService.loginWithAccount.and.returnValue(of({ token: 'jwt', user: {} as any, approval_status: 'pending' }));
      
      component.onSubmit();
      
      expect(mockRouter.navigate).toHaveBeenCalledWith(['/pending-approval']);
    });
  });

  describe('onSubmit errors and lockout', () => {
    beforeEach(() => {
      component.loginForm.controls['email'].setValue('test@test.com');
      component.loginForm.controls['password'].setValue('1234');
    });

    it('should navigate to account-notice if error user_status is blocked', () => {
      mockAuthService.checkCredentials.and.returnValue(throwError(() => ({ error: { user_status: 'blocked' } })));
      component.onSubmit();
      expect(mockRouter.navigate).toHaveBeenCalledWith(['/account-notice']);
    });

    it('should handle failed attempts and lock out after MAX_ATTEMPTS', () => {
      mockAuthService.checkCredentials.and.returnValue(throwError(() => ({ error: { error: 'Bad login' } })));
      
      for(let i=0; i<5; i++) {
        component.onSubmit();
      }
      
      expect(component.locked).toBeTrue();
      expect(component.loginError).toContain('bloqueada temporalmente');
      
      // Should not call service again after lock
      mockAuthService.checkCredentials.calls.reset();
      component.onSubmit();
      expect(mockAuthService.checkCredentials).not.toHaveBeenCalled();
    });
  });
  
  describe('Google Login', () => {
    it('should handle google login success', () => {
      const account = { user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'active' };
      mockAuthService.checkGoogleCredentials.and.returnValue(of({ accounts: [account], login_nonce: 'nonce' }));
      mockAuthService.loginWithAccount.and.returnValue(of({ token: 'jwt', user: {} as any }));
      
      component.handleGoogleLogin();
      
      expect(mockRouter.navigate).toHaveBeenCalledWith(['/dashboard']);
    });

    it('should handle google login error', () => {
      mockAuthService.checkGoogleCredentials.and.returnValue(throwError(() => ({ error: { error: 'Google err' } })));
      component.handleGoogleLogin();
      expect(component.loginError).toBe('Google err');
    });
  });
});
