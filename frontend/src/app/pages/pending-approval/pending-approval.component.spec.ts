import { ComponentFixture, TestBed } from "@angular/core/testing";
import { PendingApprovalComponent } from "./pending-approval.component";
import { Router, provideRouter } from "@angular/router";
import { AuthService } from "../../services/auth.service";
import { ProfileService } from "../../services/profile.service";
import { ReactiveFormsModule } from "@angular/forms";
import { of, throwError } from "rxjs";

describe("PendingApprovalComponent", () => {
  let component: PendingApprovalComponent;
  let fixture: ComponentFixture<PendingApprovalComponent>;
  let mockAuthService: jasmine.SpyObj<AuthService>;
  let mockProfileService: jasmine.SpyObj<ProfileService>;
  let router: Router;

  beforeEach(async () => {
    mockAuthService = jasmine.createSpyObj("AuthService", [
      "isAuthenticated",
      "updateProfile",
      "logout",
    ]);
    mockProfileService = jasmine.createSpyObj("ProfileService", [
      "getMyProfile",
    ]);

    await TestBed.configureTestingModule({
      imports: [PendingApprovalComponent, ReactiveFormsModule],
      providers: [
        { provide: AuthService, useValue: mockAuthService },
        { provide: ProfileService, useValue: mockProfileService },
        provideRouter([]),
      ],
    }).compileComponents();

    router = TestBed.inject(Router);
    spyOn(router, "navigate");
  });

  it("should initialize and build form for participant role", () => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ role_id: "role_participant" } as any),
    );
    mockProfileService.getMyProfile.and.returnValue(
      of({
        role_data: { birthdate: "1990-01-01T00:00:00Z", dni: "123" },
      } as any),
    );

    fixture = TestBed.createComponent(PendingApprovalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    expect(component.roleId).toBe("role_participant");
    expect(component.getRoleName()).toBe("Participante");
    expect(component.profileForm.get("dni")?.value).toBe("123");
    expect(component.profileForm.get("birthdate")?.value).toBe("1990-01-01");
  });

  it("should initialize and build form for company role", () => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ user: { role_id: "role_company" } } as any),
    );
    mockProfileService.getMyProfile.and.returnValue(
      of({ role_data: { business_name: "Acme Corp" } } as any),
    );

    fixture = TestBed.createComponent(PendingApprovalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    expect(component.roleId).toBe("role_company");
    expect(component.getRoleName()).toBe("Empresa");
    expect(component.profileForm.get("business_name")?.value).toBe("Acme Corp");
  });

  it("should initialize and build form for teacher role", () => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ role_id: "role_teacher" } as any),
    );
    mockProfileService.getMyProfile.and.returnValue(
      of({ role_data: { specialization: "Math" } } as any),
    );

    fixture = TestBed.createComponent(PendingApprovalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    expect(component.roleId).toBe("role_teacher");
    expect(component.getRoleName()).toBe("Docente");
    expect(component.profileForm.get("specialization")?.value).toBe("Math");
  });

  it("should initialize and build form for volunteer role", () => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ role_id: "role_volunteer" } as any),
    );
    mockProfileService.getMyProfile.and.returnValue(
      of({ role_data: { phone: "123456" } } as any),
    );

    fixture = TestBed.createComponent(PendingApprovalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    expect(component.roleId).toBe("role_volunteer");
    expect(component.getRoleName()).toBe("Voluntario");
    expect(component.profileForm.get("phone")?.value).toBe("123456");
  });

  it("should fallback to default form for unknown role", () => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ role_id: "unknown_role" } as any),
    );
    mockProfileService.getMyProfile.and.returnValue(of({} as any));

    fixture = TestBed.createComponent(PendingApprovalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    expect(component.roleId).toBe("unknown_role");
    expect(component.getRoleName()).toBe("Usuario");
    expect(component.profileForm.get("notes")).toBeDefined();
  });

  it("should redirect to /login if isAuthenticated fails on init", () => {
    mockAuthService.isAuthenticated.and.returnValue(
      throwError(() => new Error("Unauthorized")),
    );

    fixture = TestBed.createComponent(PendingApprovalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    expect(router.navigate).toHaveBeenCalledWith(["/login"]);
  });

  it("should toggle form visibility", () => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ role_id: "role_participant" } as any),
    );
    mockProfileService.getMyProfile.and.returnValue(of({} as any));

    fixture = TestBed.createComponent(PendingApprovalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    expect(component.showForm).toBeFalse();
    component.toggleForm();
    expect(component.showForm).toBeTrue();
  });

  it("should handle onSubmit success", () => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ role_id: "role_participant" } as any),
    );
    mockProfileService.getMyProfile.and.returnValue(of({} as any));
    mockAuthService.updateProfile.and.returnValue(of({} as any));

    fixture = TestBed.createComponent(PendingApprovalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    component.profileForm.patchValue({ birthdate: "1990-01-01", dni: "123" });
    component.onSubmit();

    expect(component.loading).toBeFalse();
    expect(component.formSaved).toBeTrue();
    expect(component.successMessage).toBe(
      "Datos del perfil actualizados correctamente.",
    );
    expect(component.showForm).toBeFalse();
  });

  it("should handle onSubmit error", () => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ role_id: "role_participant" } as any),
    );
    mockProfileService.getMyProfile.and.returnValue(of({} as any));
    mockAuthService.updateProfile.and.returnValue(
      throwError(() => ({ error: { error: "Error guardando" } })),
    );

    fixture = TestBed.createComponent(PendingApprovalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    component.onSubmit();

    expect(component.loading).toBeFalse();
    expect(component.errorMessage).toBe("Error guardando");
  });

  it("should handle onSubmit error with default message when err.error.error is missing", () => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ role_id: "role_participant" } as any),
    );
    mockProfileService.getMyProfile.and.returnValue(of({} as any));
    mockAuthService.updateProfile.and.returnValue(
      throwError(() => ({ error: {} })),
    );

    fixture = TestBed.createComponent(PendingApprovalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    component.onSubmit();

    expect(component.loading).toBeFalse();
    expect(component.errorMessage).toBe(
      "Ocurrió un error al guardar los datos.",
    );
  });

  it("should navigate home when goBack is called", () => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ role_id: "role_participant" } as any),
    );
    mockProfileService.getMyProfile.and.returnValue(of({} as any));

    fixture = TestBed.createComponent(PendingApprovalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    component.goBack();
    expect(router.navigate).toHaveBeenCalledWith(["/"]);
  });

  it("should call logout and navigate to login", () => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ role_id: "role_participant" } as any),
    );
    mockProfileService.getMyProfile.and.returnValue(of({} as any));

    fixture = TestBed.createComponent(PendingApprovalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    component.logout();
    expect(mockAuthService.logout).toHaveBeenCalled();
    expect(router.navigate).toHaveBeenCalledWith(["/login"]);
  });
});
