import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ProgresoComponent } from "./progreso.component";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { EducationService } from "../../../services/education.service";
import { of, throwError } from "rxjs";
import {
  Course,
  Enrollment,
  ParticipantProgress,
} from "../../../core/models/education.model";

describe("ProgresoComponent", () => {
  let component: ProgresoComponent;
  let fixture: ComponentFixture<ProgresoComponent>;
  let educationServiceSpy: jasmine.SpyObj<EducationService>;

  const mockCourses: Course[] = [
    { id: "c1", title: "Course 1", description: null, mode: "Virtual" },
    { id: "c2", title: "Course 2", description: null, mode: "Presencial" },
  ];

  const mockProgress: ParticipantProgress[] = [
    {
      participanteId: "p1",
      porcentajeAsistencia: 50,
      actividadesAsistidas: 2,
      actividadesTotal: 4,
    },
    {
      participanteId: "p2",
      porcentajeAsistencia: 100,
      actividadesAsistidas: 4,
      actividadesTotal: 4,
    },
  ];

  beforeEach(async () => {
    educationServiceSpy = jasmine.createSpyObj("EducationService", [
      "listMyCourses",
      "getCourseProgress",
      "listEnrollments",
    ]);

    educationServiceSpy.listMyCourses.and.returnValue(of(mockCourses));
    educationServiceSpy.getCourseProgress.and.returnValue(of(mockProgress));
    educationServiceSpy.listEnrollments.and.returnValue(of([]));

    await TestBed.configureTestingModule({
      imports: [ProgresoComponent, BrowserAnimationsModule],
      providers: [{ provide: EducationService, useValue: educationServiceSpy }],
    }).compileComponents();

    fixture = TestBed.createComponent(ProgresoComponent);
    component = fixture.componentInstance;
  });

  it("should create", () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it("should calculate overallProgress correctly", () => {
    expect(component.overallProgress).toBe(0);

    component.rows = mockProgress.map((p) => ({ ...p, nombre: p.participanteId }));
    // (50 + 100) / 2 = 75
    expect(component.overallProgress).toBe(75);
  });

  it("should load courses on init (>1 courses)", () => {
    fixture.detectChanges();
    expect(component.courseOptions).toHaveSize(2);
    expect(component.loadingCourses).toBeFalse();
  });

  it("should auto-select course if only 1 course is returned", () => {
    educationServiceSpy.listMyCourses.and.returnValue(of([mockCourses[0]]));
    fixture.detectChanges();
    expect(component.selectedCourseId).toBe("c1");
    expect(educationServiceSpy.getCourseProgress).toHaveBeenCalledWith("c1");
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
    expect(educationServiceSpy.getCourseProgress).not.toHaveBeenCalled();
  });

  it("should fetch progress onCourseChange", () => {
    educationServiceSpy.listEnrollments.and.returnValue(
      of([
        { participanteId: mockProgress[0].participanteId, nombre: "Ana", apellido: "Paz" },
      ] as Enrollment[]),
    );
    component.selectedCourseId = "c1";
    component.onCourseChange();

    expect(component.rows.length).toBe(mockProgress.length);
    expect(component.rows.map((r) => r.nombre)).toContain("Ana Paz");
    expect(component.loadingProgress).toBeFalse();
  });

  it("should handle error in onCourseChange", () => {
    educationServiceSpy.getCourseProgress.and.returnValue(
      throwError(() => new Error("Error progress")),
    );
    component.selectedCourseId = "c1";
    component.onCourseChange();

    expect(component.error).toBe("Error progress");
    expect(component.loadingProgress).toBeFalse();
  });
});
