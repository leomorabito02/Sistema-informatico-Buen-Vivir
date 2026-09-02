import { ComponentFixture, TestBed } from "@angular/core/testing";
import { SeguimientoComponent } from "./seguimiento.component";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { EducationService } from "../../../services/education.service";
import { of, throwError } from "rxjs";
import {
  Course,
  Enrollment,
  ParticipantProgress,
} from "../../../core/models/education.model";

describe("SeguimientoComponent", () => {
  let component: SeguimientoComponent;
  let fixture: ComponentFixture<SeguimientoComponent>;
  let educationServiceSpy: jasmine.SpyObj<EducationService>;

  const mockCourses: Course[] = [
    { id: "c1", title: "Course 1", description: null, mode: "Virtual" },
    { id: "c2", title: "Course 2", description: null, mode: "Presencial" },
  ];

  const mockEnrollments: Enrollment[] = [
    {
      id: "e1",
      cursoId: "c1",
      participanteId: "p1",
      estado: "activa",
      fechaInscripcion: "2026-01-01",
    },
    {
      id: "e2",
      cursoId: "c1",
      participanteId: "p2",
      estado: "activa",
      fechaInscripcion: "2026-01-01",
    },
    {
      id: "e3",
      cursoId: "c1",
      participanteId: "p3",
      estado: "activa",
      fechaInscripcion: "2026-01-01",
    },
    {
      id: "e4",
      cursoId: "c1",
      participanteId: "p4",
      estado: "activa",
      fechaInscripcion: "2026-01-01",
    },
  ];

  const mockProgress: ParticipantProgress[] = [
    {
      participanteId: "p1",
      porcentajeAsistencia: 40,
      actividadesAsistidas: 2,
      actividadesTotal: 5,
    },
    {
      participanteId: "p2",
      porcentajeAsistencia: 60,
      actividadesAsistidas: 3,
      actividadesTotal: 5,
    },
    {
      participanteId: "p3",
      porcentajeAsistencia: 80,
      actividadesAsistidas: 4,
      actividadesTotal: 5,
    },
  ];

  beforeEach(async () => {
    educationServiceSpy = jasmine.createSpyObj("EducationService", [
      "listMyCourses",
      "listEnrollments",
      "getCourseProgress",
    ]);

    educationServiceSpy.listMyCourses.and.returnValue(of(mockCourses));
    educationServiceSpy.listEnrollments.and.returnValue(of(mockEnrollments));
    educationServiceSpy.getCourseProgress.and.returnValue(of(mockProgress));

    await TestBed.configureTestingModule({
      imports: [SeguimientoComponent, BrowserAnimationsModule],
      providers: [{ provide: EducationService, useValue: educationServiceSpy }],
    }).compileComponents();

    fixture = TestBed.createComponent(SeguimientoComponent);
    component = fixture.componentInstance;
  });

  it("should create", () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it("should load courses on init", () => {
    fixture.detectChanges();
    expect(component.courseOptions).toHaveSize(2);
    expect(component.loadingCourses).toBeFalse();
  });

  it("should auto-select course if only 1 course is returned", () => {
    educationServiceSpy.listMyCourses.and.returnValue(of([mockCourses[0]]));
    fixture.detectChanges();
    expect(component.selectedCourseId).toBe("c1");
    expect(educationServiceSpy.listEnrollments).toHaveBeenCalledWith("c1");
  });

  it("should handle error when loading courses", () => {
    educationServiceSpy.listMyCourses.and.returnValue(
      throwError(() => new Error("Error loading courses")),
    );
    fixture.detectChanges();
    expect(component.error).toBe("Error loading courses");
    expect(component.loadingCourses).toBeFalse();
  });

  it("should return early onCourseChange if no course selected", () => {
    component.selectedCourseId = null;
    component.onCourseChange();
    expect(educationServiceSpy.listEnrollments).not.toHaveBeenCalled();
  });

  it("should fetch and calculate risk rows onCourseChange", () => {
    component.selectedCourseId = "c1";
    component.onCourseChange();

    expect(component.rows).toHaveSize(4);

    // p1: 40% -> En riesgo
    expect(component.rows[0].riesgo).toBe("En riesgo");

    // p2: 60% -> Regular
    expect(component.rows[1].riesgo).toBe("Regular");

    // p3: 80% -> Buen avance
    expect(component.rows[2].riesgo).toBe("Buen avance");

    // p4: no progress data -> Sin datos
    expect(component.rows[3].riesgo).toBe("Sin datos");
    expect(component.rows[3].porcentajeAsistencia).toBe(0);
    expect(component.rows[3].actividadesAsistidas).toBe(0);

    expect(component.loadingData).toBeFalse();
  });

  it("should handle error in onCourseChange", () => {
    educationServiceSpy.listEnrollments.and.returnValue(
      throwError(() => new Error("Error enrollments")),
    );
    component.selectedCourseId = "c1";
    component.onCourseChange();

    expect(component.error).toBe("Error enrollments");
    expect(component.loadingData).toBeFalse();
  });

  it("should return correct risk severities", () => {
    expect(component.riesgoSeverity("Buen avance")).toBe("success");
    expect(component.riesgoSeverity("Regular")).toBe("info");
    expect(component.riesgoSeverity("En riesgo")).toBe("danger");
    expect(component.riesgoSeverity("Sin datos")).toBe("secondary");
    expect(component.riesgoSeverity("Otro")).toBe("secondary");
  });
});
