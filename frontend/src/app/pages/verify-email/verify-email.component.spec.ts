import { ComponentFixture, TestBed } from "@angular/core/testing";
import { VerifyEmailComponent } from "./verify-email.component";
import { Router, ActivatedRoute, provideRouter } from "@angular/router";
import { of, throwError } from "rxjs";
import { AuthService } from "../../services/auth.service";

describe("VerifyEmailComponent", () => {
  let component: VerifyEmailComponent;
  let fixture: ComponentFixture<VerifyEmailComponent>;
  let router: Router;
  let authServiceSpy: jasmine.SpyObj<AuthService>;

  beforeEach(async () => {
    authServiceSpy = jasmine.createSpyObj("AuthService", [
      "verifyEmail",
      "resendVerificationEmail",
    ]);

    await TestBed.configureTestingModule({
      imports: [VerifyEmailComponent],
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: authServiceSpy },
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              queryParamMap: {
                get: (key: string) => (key === "email" ? "test@example.com" : null),
              },
            },
          },
        },
      ],
    }).compileComponents();

    router = TestBed.inject(Router);
    spyOn(router, "navigate");

    fixture = TestBed.createComponent(VerifyEmailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create component and populate email from query params", () => {
    expect(component).toBeTruthy();
    expect(component.email).toBe("test@example.com");
  });

  it("should navigate to /login when goToLogin is called", () => {
    component.goToLogin();
    expect(router.navigate).toHaveBeenCalledWith(["/login"]);
  });

  it("should call resendVerificationEmail on handleResendEmail", () => {
    authServiceSpy.resendVerificationEmail.and.returnValue(
      of({ message: "sent" }),
    );

    component.email = "test@example.com";
    component.handleResendEmail();

    expect(authServiceSpy.resendVerificationEmail).toHaveBeenCalledWith(
      "test@example.com",
    );
    expect(component.resendSuccess).toContain("¡Correo de verificación enviado!");
  });

  it("should set resendError on handleResendEmail failure", () => {
    authServiceSpy.resendVerificationEmail.and.returnValue(
      throwError(() => new Error("Network error")),
    );

    component.email = "test@example.com";
    component.handleResendEmail();

    expect(component.resendError).toBe("Network error");
  });
});
