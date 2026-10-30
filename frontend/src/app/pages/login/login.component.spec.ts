import { ComponentFixture, TestBed } from "@angular/core/testing";
import { LoginComponent } from "./login.component";
import { AuthService } from "../../services/auth.service";
import { Router, ActivatedRoute, provideRouter } from "@angular/router";
import { of, throwError } from "rxjs";
import { ReactiveFormsModule } from "@angular/forms";
import { By } from "@angular/platform-browser";

describe("LoginComponent", () => {
  let component: LoginComponent;
  let fixture: ComponentFixture<LoginComponent>;
  let authServiceSpy: jasmine.SpyObj<AuthService>;
  let routerSpy: jasmine.SpyObj<Router>;

  beforeEach(async () => {
    authServiceSpy = jasmine.createSpyObj("AuthService", [
      "checkCredentials",
      "checkGoogleCredentials",
      "loginWithAccount",
    ]);

    await TestBed.configureTestingModule({
      imports: [LoginComponent, ReactiveFormsModule],
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: authServiceSpy },
        { provide: ActivatedRoute, useValue: { queryParams: of({}) } },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(LoginComponent);
    component = fixture.componentInstance;
    routerSpy = TestBed.inject(Router) as jasmine.SpyObj<Router>;
    spyOn(routerSpy, "navigate");
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should toggle password visibility", () => {
    expect(component.showPassword).toBeFalse();
    component.togglePasswordVisibility();
    expect(component.showPassword).toBeTrue();
  });

  it("should handle invalid form on submit", () => {
    component.onSubmit();
    expect(component.loginForm.invalid).toBeTrue();
    expect(component.loginError).toContain(
      "Por favor, ingresá tu email y contraseña",
    );
  });

  describe("Form Validation", () => {
    it("should validate email format", () => {
      component.loginForm.controls["email"].setValue("invalid");
      component.loginForm.controls["password"].setValue("1234");
      component.onSubmit();
      expect(component.loginError).toContain("correo electrónico válido");
    });

    it("should validate password length", () => {
      component.loginForm.controls["email"].setValue("test@test.com");
      component.loginForm.controls["password"].setValue("123");
      component.onSubmit();
      expect(component.loginError).toContain("al menos 4 caracteres");
    });
  });

  describe("Successful Login", () => {
    it("should login directly if single account", () => {
      const accounts = [
        {
          user_id: "u1",
          role_id: "r1",
          status: "active",
          role_name: "admin",
          entity_name: "",
        },
      ];
      authServiceSpy.checkCredentials.and.returnValue(
        of({ accounts, login_nonce: "nonce" }),
      );
      authServiceSpy.loginWithAccount.and.returnValue(
        of({ user_status: "active", approval_status: "approved" } as any),
      );

      component.loginForm.setValue({
        email: "test@test.com",
        password: "password",
        rememberMe: false,
      });
      component.onSubmit();

      expect(authServiceSpy.loginWithAccount).toHaveBeenCalledWith(
        "u1",
        "r1",
        "nonce",
      );
      expect(routerSpy.navigate).toHaveBeenCalledWith(["/dashboard"]);
    });

    it("should navigate to role select if multiple accounts", () => {
      const accounts = [
        {
          user_id: "u1",
          role_id: "r1",
          status: "active",
          role_name: "admin",
          entity_name: "",
        },
        {
          user_id: "u2",
          role_id: "r2",
          status: "active",
          role_name: "user",
          entity_name: "",
        },
      ];
      authServiceSpy.checkCredentials.and.returnValue(
        of({ accounts, login_nonce: "nonce" }),
      );

      component.loginForm.setValue({
        email: "test@test.com",
        password: "password",
        rememberMe: false,
      });
      component.onSubmit();

      expect(routerSpy.navigate).toHaveBeenCalledWith(["/select-role"], {
        state: { accounts, loginNonce: "nonce" },
      });
    });
  });

  describe("Error Handling & Account Status", () => {
    it("should handle blocked status on checkCredentials", () => {
      authServiceSpy.checkCredentials.and.returnValue(
        throwError(() => ({ error: { user_status: "blocked" } })),
      );

      component.loginForm.setValue({
        email: "test@test.com",
        password: "password",
        rememberMe: false,
      });
      component.onSubmit();

      expect(routerSpy.navigate).toHaveBeenCalledWith(["/account-notice"]);
    });

    it("should lock out after max attempts", () => {
      authServiceSpy.checkCredentials.and.returnValue(
        throwError(() => ({ error: { error: "Invalid" } })),
      );
      component.loginForm.setValue({
        email: "test@test.com",
        password: "password",
        rememberMe: false,
      });

      for (let i = 0; i < 5; i++) {
        component.onSubmit();
      }

      expect(component.locked).toBeTrue();
      expect(component.loginError).toContain("bloqueada temporalmente");
    });

    it("should handle pending approval status", () => {
      const accounts = [
        {
          user_id: "u1",
          role_id: "r1",
          status: "active",
          role_name: "admin",
          entity_name: "",
        },
      ];
      authServiceSpy.checkCredentials.and.returnValue(
        of({ accounts, login_nonce: "nonce" }),
      );
      authServiceSpy.loginWithAccount.and.returnValue(
        of({ user_status: "active", approval_status: "pending" } as any),
      );

      component.loginForm.setValue({
        email: "test@test.com",
        password: "password",
        rememberMe: false,
      });
      component.onSubmit();

      expect(routerSpy.navigate).toHaveBeenCalledWith(["/pending-approval"]);
    });
  });

  describe("Google Login", () => {
    it("should handle google login success", () => {
      const accounts = [
        {
          user_id: "u1",
          role_id: "r1",
          status: "active",
          role_name: "admin",
          entity_name: "",
        },
      ];
      authServiceSpy.checkGoogleCredentials.and.returnValue(
        of({ accounts, login_nonce: "nonce" }),
      );
      authServiceSpy.loginWithAccount.and.returnValue(
        of({ user_status: "active", approval_status: "approved" } as any),
      );

      component.handleGoogleLogin();

      expect(authServiceSpy.checkGoogleCredentials).toHaveBeenCalled();
      expect(routerSpy.navigate).toHaveBeenCalledWith(["/dashboard"]);
    });

    it("should handle google login error", () => {
      authServiceSpy.checkGoogleCredentials.and.returnValue(
        throwError(() => ({ error: { error: "Google Error" } })),
      );

      component.handleGoogleLogin();

      expect(component.loginError).toBe("Google Error");
    });

    it("should navigate to /verify-email when google login account has is_email_verified = false", () => {
      const accounts = [
        {
          user_id: "u1",
          role_id: "r1",
          status: "active",
          role_name: "admin",
          entity_name: "",
        },
      ];
      authServiceSpy.checkGoogleCredentials.and.returnValue(
        of({
          accounts,
          login_nonce: "nonce",
          is_email_verified: false,
          email: "google@example.com",
        }),
      );

      component.handleGoogleLogin();

      expect(routerSpy.navigate).toHaveBeenCalledWith(["/verify-email"], {
        queryParams: { email: "google@example.com" },
      });
    });

    it("should navigate to /verify-email when checkCredentials has is_email_verified = false", () => {
      component.loginForm.setValue({
        email: "test@example.com",
        password: "Password123!",
        rememberMe: false,
      });

      authServiceSpy.checkCredentials.and.returnValue(
        of({
          accounts: [],
          login_nonce: "nonce",
          is_email_verified: false,
          email: "test@example.com",
        }),
      );

      component.onSubmit();

      expect(routerSpy.navigate).toHaveBeenCalledWith(["/verify-email"], {
        queryParams: { email: "test@example.com" },
      });
    });
  });
});
