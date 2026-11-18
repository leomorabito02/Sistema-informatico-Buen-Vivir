import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ChangePasswordComponent } from "./change-password.component";
import { Router, ActivatedRoute, provideRouter } from "@angular/router";
import { of, throwError } from "rxjs";
import { AuthService } from "../../services/auth.service";

describe("ChangePasswordComponent", () => {
  let component: ChangePasswordComponent;
  let fixture: ComponentFixture<ChangePasswordComponent>;
  let router: Router;
  let authServiceSpy: jasmine.SpyObj<AuthService>;

  beforeEach(async () => {
    authServiceSpy = jasmine.createSpyObj("AuthService", [
      "forgotPassword",
      "resetPassword",
    ]);

    await TestBed.configureTestingModule({
      imports: [ChangePasswordComponent],
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: authServiceSpy },
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              queryParamMap: {
                get: (key: string) => {
                  if (key === "email") return "test@example.com";
                  if (key === "token") return "sample-token-123";
                  return null;
                },
              },
            },
          },
        },
      ],
    }).compileComponents();

    router = TestBed.inject(Router);
    spyOn(router, "navigate");

    fixture = TestBed.createComponent(ChangePasswordComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create component and read token and email from query params", () => {
    // Assert
    expect(component).toBeTruthy();
    expect(component.token).toBe("sample-token-123");
    expect(component.requestEmail).toBe("test@example.com");
  });

  it("should validate password complexity correctly", () => {
    // Arrange & Act
    component.newPassword = "weak";
    expect(component.isPasswordValid).toBeFalse();

    component.newPassword = "ValidPassword1!";
    component.confirmPassword = "ValidPassword1!";

    // Assert
    expect(component.hasMinLength).toBeTrue();
    expect(component.hasUpper).toBeTrue();
    expect(component.hasLower).toBeTrue();
    expect(component.hasSpecial).toBeTrue();
    expect(component.isPasswordValid).toBeTrue();
    expect(component.passwordsMatch).toBeTrue();
  });

  it("should call authService.forgotPassword on handleRequestReset", () => {
    // Arrange
    authServiceSpy.forgotPassword.and.returnValue(
      of({ message: "Enlace enviado" }),
    );
    component.requestEmail = "user@example.com";

    // Act
    component.handleRequestReset();

    // Assert
    expect(authServiceSpy.forgotPassword).toHaveBeenCalledWith(
      "user@example.com",
    );
    expect(component.requestSuccess).toBe("Enlace enviado");
    expect(component.requesting).toBeFalse();
  });

  it("should handle error in handleRequestReset", () => {
    // Arrange
    authServiceSpy.forgotPassword.and.returnValue(
      throwError(() => new Error("Error al enviar")),
    );
    component.requestEmail = "user@example.com";

    // Act
    component.handleRequestReset();

    // Assert
    expect(component.requestError).toBe("Error al enviar");
    expect(component.requesting).toBeFalse();
  });

  it("should call authService.resetPassword on handleResetPassword", () => {
    // Arrange
    authServiceSpy.resetPassword.and.returnValue(
      of({ message: "Contraseña actualizada" }),
    );
    component.token = "token-abc";
    component.newPassword = "ValidPassword1!";
    component.confirmPassword = "ValidPassword1!";

    // Act
    component.handleResetPassword();

    // Assert
    expect(authServiceSpy.resetPassword).toHaveBeenCalledWith(
      "token-abc",
      "ValidPassword1!",
    );
    expect(component.resetSuccess).toBe("Contraseña actualizada");
    expect(component.resetting).toBeFalse();
  });

  it("should handle error in handleResetPassword", () => {
    // Arrange
    authServiceSpy.resetPassword.and.returnValue(
      throwError(() => new Error("El enlace ha expirado")),
    );
    component.token = "token-abc";
    component.newPassword = "ValidPassword1!";
    component.confirmPassword = "ValidPassword1!";

    // Act
    component.handleResetPassword();

    // Assert
    expect(component.resetError).toBe("El enlace ha expirado");
    expect(component.resetting).toBeFalse();
  });

  it("should clear token and navigate on clearTokenAndRetry", () => {
    // Arrange
    component.token = "old-token";
    component.resetError = "Some error";

    // Act
    component.clearTokenAndRetry();

    // Assert
    expect(component.token).toBe("");
    expect(component.resetError).toBe("");
    expect(router.navigate).toHaveBeenCalledWith(["/change-password"]);
  });

  it("should navigate to /login when goToLogin is called", () => {
    // Act
    component.goToLogin();

    // Assert
    expect(router.navigate).toHaveBeenCalledWith(["/login"]);
  });
});
