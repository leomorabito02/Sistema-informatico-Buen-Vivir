import { ComponentFixture, TestBed } from "@angular/core/testing";
import { CursosComponent } from "./cursos.component";
import { EducationService } from "../../../services/education.service";
import { of, throwError } from "rxjs";
import { provideHttpClient } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { provideRouter } from "@angular/router";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { Auth } from "@angular/fire/auth";

describe("CursosComponent", () => {
  let component: CursosComponent;
  let fixture: ComponentFixture<CursosComponent>;
  let educationServiceSpy: jasmine.SpyObj<EducationService>;

  beforeEach(async () => {
    educationServiceSpy = jasmine.createSpyObj("EducationService", [
      "listCourses",
    ]);
    educationServiceSpy.listCourses.and.returnValue(of([]));

    await TestBed.configureTestingModule({
      imports: [CursosComponent, BrowserAnimationsModule],
      providers: [
        { provide: EducationService, useValue: educationServiceSpy },
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

  it("should open detail dialog", () => {
    const c: any = { id: "1", title: "Course 1" };
    component.openDetail(c);
    expect(component.detail).toBe(c);
    expect(component.detailVisible).toBeTrue();
  });

  it("should return correct severity for mode", () => {
    expect(component.modeSeverity("Presencial")).toBe("success");
    expect(component.modeSeverity("Virtual")).toBe("info");
    expect(component.modeSeverity("Mixta")).toBe("warn");
    expect(component.modeSeverity("Unknown")).toBe("secondary");
  });
});
