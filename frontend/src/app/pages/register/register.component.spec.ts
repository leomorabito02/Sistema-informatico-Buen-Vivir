import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RegisterComponent } from './register.component';
import { ReactiveFormsModule, FormControl, FormGroup } from '@angular/forms';
import { Router, ActivatedRoute, provideRouter } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { of, Subject } from 'rxjs';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

describe('RegisterComponent', () => {
  let component: RegisterComponent;
  let fixture: ComponentFixture<RegisterComponent>;
  let mockAuthService: jasmine.SpyObj<AuthService>;
  let mockRouter: any;
  let queryParamsSubject: Subject<any>;

  beforeEach(async () => {
    mockAuthService = jasmine.createSpyObj('AuthService', ['registerWithEmail']);
    queryParamsSubject = new Subject<any>();

    await TestBed.configureTestingModule({
      imports: [RegisterComponent, ReactiveFormsModule, BrowserAnimationsModule],
      providers: [
        { provide: AuthService, useValue: mockAuthService },
        provideRouter([])
      ]
    })
    .overrideProvider(ActivatedRoute, { useValue: { queryParams: queryParamsSubject.asObservable() } })
    .compileComponents();
    
    mockRouter = TestBed.inject(Router);
    spyOn(mockRouter, 'navigate');

    fixture = TestBed.createComponent(RegisterComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create with correct default states and role options', () => {
    expect(component).toBeTruthy();
    expect(component.registerForm.get('role')?.value).toBe('participant');
    expect(component.registered).toBeFalse();
    expect(component.registerError).toBe('');
    expect(component.loading).toBeFalse();
    expect(component.showPassword).toBeFalse();
    expect(component.showConfirmPassword).toBeFalse();

    expect(component.roles.length).toBe(4);
    expect(component.roles[0]).toEqual({ label: 'Participante', value: 'participant' });
    expect(component.roles[1]).toEqual({ label: 'Voluntario', value: 'volunteer' });
    expect(component.roles[2]).toEqual({ label: 'Docente', value: 'teacher' });
    expect(component.roles[3]).toEqual({ label: 'Empresa', value: 'company' });
  });

  it('should patch role when queryParams contain role', () => {
    queryParamsSubject.next({ role: 'teacher' });
    expect(component.registerForm.get('role')?.value).toBe('teacher');
  });

  it('should not patch role when queryParams omit role', () => {
    queryParamsSubject.next({});
    expect(component.registerForm.get('role')?.value).toBe('participant');
  });

  it('should toggle password visibility', () => {
    expect(component.showPassword).toBeFalse();
    component.togglePasswordVisibility();
    expect(component.showPassword).toBeTrue();
    component.togglePasswordVisibility();
    expect(component.showPassword).toBeFalse();
  });

  it('should toggle confirm password visibility', () => {
    expect(component.showConfirmPassword).toBeFalse();
    component.toggleConfirmPasswordVisibility();
    expect(component.showConfirmPassword).toBeTrue();
    component.toggleConfirmPasswordVisibility();
    expect(component.showConfirmPassword).toBeFalse();
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

    it('should return false for participant and other values', () => {
      component.registerForm.controls['role'].setValue('participant');
      expect(component.needsApproval).toBeFalse();
      component.registerForm.controls['role'].setValue('unknown');
      expect(component.needsApproval).toBeFalse();
    });
  });

  describe('Form validations', () => {
    it('should set error when required fields are empty independently and mark fields touched', () => {
      const validForm = {
        firstName: 'John', lastName: 'Doe', phone: '123456', email: 'john@doe.com',
        password: 'Password1!', confirmPassword: 'Password1!', terms: true, role: 'volunteer'
      };

      const fieldsToTest = ['firstName', 'lastName', 'email', 'phone', 'password'];
      for (const field of fieldsToTest) {
        component.registerForm.patchValue({ ...validForm, [field]: '' });
        component.onSubmit();
        expect(component.registerError).toBe('Por favor, completá todos los campos obligatorios.');
        expect(component.registerForm.touched).toBeTrue();
      }
    });

    it('should validate email format when fields are otherwise valid', () => {
      component.registerForm.patchValue({
        firstName: 'John', lastName: 'Doe', phone: '123456', password: 'Password1!', confirmPassword: 'Password1!', terms: true, email: 'invalid-email'
      });
      component.onSubmit();
      expect(component.registerError).toBe('Por favor, ingresá una dirección de correo electrónico válida.');
    });

    it('should validate password complexity rules sequentially', () => {
      const baseForm = {
        firstName: 'John', lastName: 'Doe', phone: '123456', email: 'john@doe.com', terms: true, role: 'volunteer'
      };

      // minlength
      component.registerForm.patchValue({ ...baseForm, password: 'short', confirmPassword: 'short' });
      component.onSubmit();
      expect(component.registerError).toBe('La contraseña debe tener al menos 8 caracteres.');

      // missingUpper
      component.registerForm.patchValue({ ...baseForm, password: 'nouppercase1!', confirmPassword: 'nouppercase1!' });
      component.onSubmit();
      expect(component.registerError).toBe('La contraseña debe incluir al menos una letra mayúscula.');

      // missingLower
      component.registerForm.patchValue({ ...baseForm, password: 'NOLOWERCASE1!', confirmPassword: 'NOLOWERCASE1!' });
      component.onSubmit();
      expect(component.registerError).toBe('La contraseña debe incluir al menos una letra minúscula.');

      // missingSpecial
      component.registerForm.patchValue({ ...baseForm, password: 'NoSpecialChar123', confirmPassword: 'NoSpecialChar123' });
      component.onSubmit();
      expect(component.registerError).toBe('La contraseña debe incluir al menos un carácter especial (ej: !@#$%^&*).');
    });

    it('should validate password match', () => {
      component.registerForm.patchValue({
        firstName: 'John', lastName: 'Doe', phone: '123456', email: 'john@doe.com', terms: true,
        password: 'Password1!', confirmPassword: 'Password2@'
      });
      component.onSubmit();
      expect(component.registerError).toBe('Las contraseñas no coinciden.');
    });

    it('should validate terms acceptance', () => {
      component.registerForm.patchValue({
        firstName: 'John', lastName: 'Doe', phone: '123456', email: 'john@doe.com',
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

    it('should set loading to true while registering and set registered to true on success', () => {
      const subject = new Subject<any>();
      mockAuthService.registerWithEmail.and.returnValue(subject.asObservable());
      
      component.onSubmit();
      
      expect(mockAuthService.registerWithEmail).toHaveBeenCalledWith({
        first_name: 'John',
        last_name: 'Doe',
        email: 'john@doe.com',
        phone: '123456',
        password: 'Password1!',
        role: 'volunteer'
      });
      expect(component.loading).toBeTrue();
      expect(component.registerError).toBe('');
      
      subject.next({ token: 'jwt', user: {} as any });
      
      expect(component.registered).toBeTrue();
      expect(component.loading).toBeFalse();
    });

    it('should handle registration error with custom server message', () => {
      const subject = new Subject<any>();
      spyOn(console, 'error');
      mockAuthService.registerWithEmail.and.returnValue(subject.asObservable());
      
      component.onSubmit();
      expect(component.loading).toBeTrue();
      
      subject.error({ error: { error: 'Email exists' } });
      
      expect(component.registerError).toBe('Email exists');
      expect(component.loading).toBeFalse();
      expect(console.error).toHaveBeenCalledWith('Registration error', { error: { error: 'Email exists' } });
    });

    it('should handle registration error with fallback default message when err.error.error is missing', () => {
      const subject = new Subject<any>();
      spyOn(console, 'error');
      mockAuthService.registerWithEmail.and.returnValue(subject.asObservable());
      
      component.onSubmit();
      expect(component.loading).toBeTrue();
      
      subject.error({ error: {} });
      
      expect(component.registerError).toBe('Ocurrió un error al procesar el registro. Verificá los datos e intentá nuevamente.');
      expect(component.loading).toBeFalse();
    });
  });
});
