import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RegisterComponent } from './register.component';
import { ReactiveFormsModule } from '@angular/forms';
import { Router, ActivatedRoute, provideRouter } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { of, throwError } from 'rxjs';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

describe('RegisterComponent', () => {
  let component: RegisterComponent;
  let fixture: ComponentFixture<RegisterComponent>;
  let mockAuthService: jasmine.SpyObj<AuthService>;
  let mockRouter: any;
  let mockActivatedRoute: any;

  beforeEach(async () => {
    mockAuthService = jasmine.createSpyObj('AuthService', ['registerWithEmail']);
    mockActivatedRoute = {
      queryParams: of({ role: 'volunteer' })
    };

    await TestBed.configureTestingModule({
      imports: [RegisterComponent, ReactiveFormsModule, BrowserAnimationsModule],
      providers: [
        { provide: AuthService, useValue: mockAuthService },
        provideRouter([])
      ]
    })
    .overrideProvider(ActivatedRoute, { useValue: mockActivatedRoute })
    .compileComponents();
    
    mockRouter = TestBed.inject(Router);
    spyOn(mockRouter, 'navigate');

    fixture = TestBed.createComponent(RegisterComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create and patch role from query params', () => {
    expect(component).toBeTruthy();
    expect(component.registerForm.get('role')?.value).toBe('volunteer');
  });

  it('should toggle password visibility', () => {
    expect(component.showPassword).toBeFalse();
    component.togglePasswordVisibility();
    expect(component.showPassword).toBeTrue();
  });

  it('should toggle confirm password visibility', () => {
    expect(component.showConfirmPassword).toBeFalse();
    component.toggleConfirmPasswordVisibility();
    expect(component.showConfirmPassword).toBeTrue();
  });

  describe('needsApproval getter', () => {
    it('should return true for volunteer, teacher, company', () => {
      component.registerForm.controls['role'].setValue('volunteer');
      expect(component.needsApproval).toBeTrue();
      component.registerForm.controls['role'].setValue('teacher');
      expect(component.needsApproval).toBeTrue();
      component.registerForm.controls['role'].setValue('company');
      expect(component.needsApproval).toBeTrue();
    });

    it('should return false for participant', () => {
      component.registerForm.controls['role'].setValue('participant');
      expect(component.needsApproval).toBeFalse();
    });
  });

  describe('Form validations', () => {
    it('should set error when required fields are empty', () => {
      component.registerForm.reset();
      component.onSubmit();
      expect(component.registerError).toBe('Por favor, completá todos los campos obligatorios.');
    });

    it('should validate email format', () => {
      component.registerForm.patchValue({
        firstName: 'a', lastName: 'b', phone: '1', password: 'Password1!', confirmPassword: 'Password1!', terms: true, email: 'invalid'
      });
      component.onSubmit();
      expect(component.registerError).toBe('Por favor, ingresá una dirección de correo electrónico válida.');
    });

    it('should validate password complexity', () => {
      component.registerForm.patchValue({
        firstName: 'a', lastName: 'b', phone: '1', email: 'a@a.com', terms: true,
        password: 'short', confirmPassword: 'short'
      });
      component.onSubmit();
      expect(component.registerError).toBe('La contraseña debe tener al menos 8 caracteres.');

      component.registerForm.patchValue({ password: 'nouppercase1!', confirmPassword: 'nouppercase1!' });
      component.onSubmit();
      expect(component.registerError).toBe('La contraseña debe incluir al menos una letra mayúscula.');

      component.registerForm.patchValue({ password: 'NOLOWERCASE1!', confirmPassword: 'NOLOWERCASE1!' });
      component.onSubmit();
      expect(component.registerError).toBe('La contraseña debe incluir al menos una letra minúscula.');

      component.registerForm.patchValue({ password: 'NoSpecialChar123', confirmPassword: 'NoSpecialChar123' });
      component.onSubmit();
      expect(component.registerError).toBe('La contraseña debe incluir al menos un carácter especial (ej: !@#$%^&*).');
    });

    it('should validate password match', () => {
      component.registerForm.patchValue({
        firstName: 'a', lastName: 'b', phone: '1', email: 'a@a.com', terms: true,
        password: 'Password1!', confirmPassword: 'Password2@'
      });
      component.onSubmit();
      expect(component.registerError).toBe('Las contraseñas no coinciden.');
    });

    it('should validate terms acceptance', () => {
      component.registerForm.patchValue({
        firstName: 'a', lastName: 'b', phone: '1', email: 'a@a.com',
        password: 'Password1!', confirmPassword: 'Password1!', terms: false
      });
      component.onSubmit();
      expect(component.registerError).toBe('Debés aceptar los términos y condiciones para continuar.');
    });
  });

  describe('onSubmit success/error', () => {
    beforeEach(() => {
      component.registerForm.patchValue({
        firstName: 'John', lastName: 'Doe', phone: '123456', email: 'john@doe.com',
        password: 'Password1!', confirmPassword: 'Password1!', terms: true, role: 'volunteer'
      });
    });

    it('should register successfully', () => {
      mockAuthService.registerWithEmail.and.returnValue(of({ token: 'jwt', user: {} as any }));
      component.onSubmit();
      
      expect(mockAuthService.registerWithEmail).toHaveBeenCalledWith({
        first_name: 'John',
        last_name: 'Doe',
        email: 'john@doe.com',
        phone: '123456',
        password: 'Password1!',
        role: 'volunteer'
      });
      expect(component.registered).toBeTrue();
      expect(component.loading).toBeFalse();
    });

    it('should handle registration error', () => {
      mockAuthService.registerWithEmail.and.returnValue(throwError(() => ({ error: { error: 'Email exists' } })));
      component.onSubmit();
      
      expect(component.registerError).toBe('Email exists');
      expect(component.loading).toBeFalse();
    });
  });
});
