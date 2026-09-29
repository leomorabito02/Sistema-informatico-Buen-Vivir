import { ComponentFixture, TestBed } from "@angular/core/testing";
import { CursosComponent } from "./cursos.component";
import { EducationService } from "../../../services/education.service";
import { of, throwError } from "rxjs";
import { provideHttpClient } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { Router, provideRouter } from "@angular/router";
import { AuthService } from "../../../services/auth.service";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { Auth } from "@angular/fire/auth";

describe("CursosComponent", () => {
  let component: CursosComponent;
  let fixture: ComponentFixture<CursosComponent>;
  let educationServiceSpy: jasmine.SpyObj<EducationService>;
  let authServiceSpy: jasmine.SpyObj<AuthService>;

  const me = (role_id: string, user_id = "u1"): any => ({
    role_id,
    user: { user_id, role_id },
  });

  beforeEach(async () => {
    educationServiceSpy = jasmine.createSpyObj("EducationService", [
      "listCourses",
      "listMyEnrolledCourses",
      "getGeneralAttendance",
    ]);
    educationServiceSpy.listCourses.and.returnValue(of([]));
    educationServiceSpy.listMyEnrolledCourses.and.returnValue(of([]));
    educationServiceSpy.getGeneralAttendance.and.returnValue(of([]));
    authServiceSpy = jasmine.createSpyObj("AuthService", ["isAuthenticated"]);
    authServiceSpy.isAuthenticated.and.returnValue(of(me("role_company")));

    await TestBed.configureTestingModule({
      imports: [CursosComponent, BrowserAnimationsModule],
      providers: [
        { provide: EducationService, useValue: educationServiceSpy },
        { provide: AuthService, useValue: authServiceSpy },
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(CursosComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should load courses on init", () => {
    const mockCourses: any[] = [
      {
        id: "1",
        title: "Course 1",
        description: { content: [{ content: [{ text: "desc 1" }] }] },
        mode: "Virtual",
        trainingArea: { name: "Area 1" },
      },
    ];
    educationServiceSpy.listCourses.and.returnValue(of(mockCourses));

    component.ngOnInit();

    expect(educationServiceSpy.listCourses).toHaveBeenCalled();
    expect(component.cursos).toHaveSize(1);
    expect(component.cursos[0].title).toBe("Course 1");
    expect(component.cursos[0].area).toBe("Area 1");
    expect(component.cursos[0].mode).toBe("Virtual");
    expect(component.loading).toBeFalse();
  });

  it("should handle error when loading courses", () => {
    educationServiceSpy.listCourses.and.returnValue(
      throwError(() => new Error("Load error")),
    );

    component.ngOnInit();

    expect(component.error).toBe("Load error");
    expect(component.loading).toBeFalse();
  });

  it("should filter courses correctly", () => {
    const c1: any = {
      id: "1",
      title: "Angular Basics",
      area: "Web",
      descripcion: "learn angular",
      mode: "Virtual",
    };
    const c2: any = {
      id: "2",
      title: "React Basics",
      area: "Web",
      descripcion: "learn react",
      mode: "Virtual",
    };
    component.cursos = [c1, c2];

    component.searchTerm = "angular";
    component.filterCursos();

    expect(component.filteredCursos).toHaveSize(1);
    expect(component.filteredCursos[0].title).toBe("Angular Basics");
  });

  it("should open the summary dialog for roles without the detail page", () => {
    const c: any = { id: "1", title: "Course 1", teacherId: "otro" };
    component.openDetail(c);
    expect(component.detail).toBe(c);
    expect(component.detailVisible).toBeTrue();
  });

  it("participant sees only enrolled courses and goes to the detail page", () => {
    authServiceSpy.isAuthenticated.and.returnValue(of(me("role_participant")));
    educationServiceSpy.listMyEnrolledCourses.and.returnValue(
      of([{ id: "c9", title: "Mi curso", mode: "Virtual" } as any]),
    );
    educationServiceSpy.listCourses.calls.reset();
    const navigate = spyOn(TestBed.inject(Router), "navigate").and.resolveTo(
      true,
    );

    component.ngOnInit();

    expect(component.esParticipante).toBeTrue();
    expect(educationServiceSpy.listCourses).not.toHaveBeenCalled();
    expect(component.cursos.map((c) => c.id)).toEqual(["c9"]);

    component.openDetail(component.cursos[0]);
    expect(navigate).toHaveBeenCalledWith(["/dashboard/cursos", "c9"]);
  });

  it("teacher goes to the detail page only for their own courses", () => {
    authServiceSpy.isAuthenticated.and.returnValue(
      of(me("role_teacher", "t1")),
    );
    const navigate = spyOn(TestBed.inject(Router), "navigate").and.resolveTo(
      true,
    );
    component.ngOnInit();

    component.openDetail({ id: "c1", teacherId: "t1" } as any);
    expect(navigate).toHaveBeenCalledWith(["/dashboard/cursos", "c1"]);

    component.openDetail({ id: "c2", teacherId: "t2" } as any);
    expect(navigate).toHaveBeenCalledTimes(1);
    expect(component.detailVisible).toBeTrue();
  });

  it("should return correct severity for mode", () => {
    expect(component.modeSeverity("Presencial")).toBe("success");
    expect(component.modeSeverity("Virtual")).toBe("info");
    expect(component.modeSeverity("Mixta")).toBe("warn");
    expect(component.modeSeverity("Unknown")).toBe("secondary");
  });
});
