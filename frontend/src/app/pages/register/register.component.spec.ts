import { ComponentFixture, TestBed } from "@angular/core/testing";
import {
  RegisterComponent,
  participantGmailValidator,
} from "./register.component";
import { AuthService } from "../../services/auth.service";
import { ActivatedRoute } from "@angular/router";
import { of, throwError } from "rxjs";
import { FormControl, FormGroup, ReactiveFormsModule } from "@angular/forms";

describe("RegisterComponent", () => {
  let component: RegisterComponent;
  let fixture: ComponentFixture<RegisterComponent>;
  let authServiceSpy: jasmine.SpyObj<AuthService>;

  beforeEach(async () => {
    authServiceSpy = jasmine.createSpyObj("AuthService", ["registerWithEmail"]);

    await TestBed.configureTestingModule({
      imports: [RegisterComponent, ReactiveFormsModule],
      providers: [
        { provide: AuthService, useValue: authServiceSpy },
        {
          provide: ActivatedRoute,
          useValue: { queryParams: of({ role: "teacher" }) },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(RegisterComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should set role from query params on init", () => {
    expect(component.registerForm.controls["role"].value).toBe("teacher");
  });

  it("should compute needsApproval correctly", () => {
    component.registerForm.controls["role"].setValue("participant");
    expect(component.needsApproval).toBeFalse();

    component.registerForm.controls["role"].setValue("volunteer");
    expect(component.needsApproval).toBeTrue();
  });

  it("should toggle password visibility", () => {
    expect(component.showPassword).toBeFalse();
    component.togglePasswordVisibility();
    expect(component.showPassword).toBeTrue();

    expect(component.showConfirmPassword).toBeFalse();
    component.toggleConfirmPasswordVisibility();
    expect(component.showConfirmPassword).toBeTrue();
  });

  describe("Form Validation", () => {
    it("should validate required fields", () => {
      component.onSubmit();
      expect(component.registerForm.invalid).toBeTrue();
      expect(component.registerError).toContain(
        "completá todos los campos obligatorios",
      );
    });

    it("should validate email format", () => {
      component.registerForm.patchValue({
        firstName: "John",
        lastName: "Doe",
        phone: "123456",
        password: "Password1!",
        confirmPassword: "Password1!",
        terms: true,
      });
      component.registerForm.controls["email"].setValue("invalid");
      component.onSubmit();
      expect(component.registerError).toContain("correo electrónico válida");
    });

    it("should validate password complexity", () => {
      component.registerForm.patchValue({
        firstName: "John",
        lastName: "Doe",
        email: "test@test.com",
        phone: "123456",
        terms: true,
      });

      component.registerForm.controls["password"].setValue("short");
      component.onSubmit();
      expect(component.registerError).toContain("al menos 8 caracteres");

      component.registerForm.controls["password"].setValue("password!");
      component.onSubmit();
      expect(component.registerError).toContain("letra mayúscula");

      component.registerForm.controls["password"].setValue("PASSWORD!");
      component.onSubmit();
      expect(component.registerError).toContain("letra minúscula");

      component.registerForm.controls["password"].setValue("Password123");
      component.onSubmit();
      expect(component.registerError).toContain("carácter especial");
    });

    it("should validate password match", () => {
      component.registerForm.patchValue({
        firstName: "John",
        lastName: "Doe",
        email: "test@test.com",
        phone: "123456",
        terms: true,
        password: "Password1!",
        confirmPassword: "Password2!",
      });
      component.onSubmit();
      expect(component.registerError).toContain("no coinciden");
    });

    it("should validate terms acceptance", () => {
      component.registerForm.patchValue({
        firstName: "John",
        lastName: "Doe",
        email: "test@test.com",
        phone: "123456",
        password: "Password1!",
        confirmPassword: "Password1!",
        terms: false,
      });
      component.onSubmit();
      expect(component.registerError).toContain("términos y condiciones");
    });
  });

  describe("Successful Registration", () => {
    it("should call register api and set registered flag", () => {
      component.registerForm.patchValue({
        firstName: "John",
        lastName: "Doe",
        email: "test@test.com",
        phone: "123456",
        role: "teacher",
        password: "Password1!",
        confirmPassword: "Password1!",
        terms: true,
      });

      authServiceSpy.registerWithEmail.and.returnValue(of({} as any));

      component.onSubmit();

      expect(authServiceSpy.registerWithEmail).toHaveBeenCalledWith({
        first_name: "John",
        last_name: "Doe",
        email: "test@test.com",
        phone: "123456",
        password: "Password1!",
        role: "teacher",
      });
      expect(component.registered).toBeTrue();
      expect(component.loading).toBeFalse();
    });

    it("should handle registration error", () => {
      component.registerForm.patchValue({
        firstName: "John",
        lastName: "Doe",
        email: "test@test.com",
        phone: "123456",
        role: "teacher",
        password: "Password1!",
        confirmPassword: "Password1!",
        terms: true,
      });

      authServiceSpy.registerWithEmail.and.returnValue(
        throwError(() => ({ error: { error: "API Error" } })),
      );

      component.onSubmit();

      expect(component.registerError).toBe("API Error");
      expect(component.registered).toBeFalse();
      expect(component.loading).toBeFalse();
    });
  });
});

describe("participantGmailValidator (HU-01, decisión 15/11)", () => {
  const form = (role: string, email: string) =>
    new FormGroup({
      role: new FormControl(role),
      email: new FormControl(email),
    });

  it("rechaza un participante sin Gmail", () => {
    expect(
      participantGmailValidator(form("participant", "ana@hotmail.com")),
    ).toEqual({ participantGmail: true });
  });

  it("acepta Gmail para participantes y cualquier dominio para otros roles", () => {
    expect(
      participantGmailValidator(form("participant", "Ana@Gmail.com ")),
    ).toBeNull();
    expect(
      participantGmailValidator(form("company", "rrhh@empresa.com.ar")),
    ).toBeNull();
  });
});
