import { ComponentFixture, TestBed } from "@angular/core/testing";
import { AsistenciaComponent } from "./asistencia.component";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { provideRouter } from "@angular/router";
import { EducationService } from "../../../services/education.service";
import { of, throwError } from "rxjs";
import {
  Activity,
  Attendance,
  Course,
  Enrollment,
} from "../../../core/models/education.model";

describe("AsistenciaComponent", () => {
  let component: AsistenciaComponent;
  let fixture: ComponentFixture<AsistenciaComponent>;
  let educationServiceSpy: jasmine.SpyObj<EducationService>;

  const mockCourses: Course[] = [
    { id: "c1", title: "Course 1", description: null, mode: "Virtual" },
    { id: "c2", title: "", description: null, mode: "Presencial" },
  ];

  const act = (id: string, fecha: string): Activity => ({
    id,
    cursoId: "c1",
    titulo: `Act ${id}`,
    consigna: "Hacer",
    fecha,
    createdAt: fecha,
  });

  const enrollments: Enrollment[] = [
    {
      id: "e1",
      participanteId: "p1",
      cursoId: "c1",
      fechaInscripcion: "",
      estado: "activa",
    },
    {
      id: "e2",
      participanteId: "p2",
      cursoId: "c1",
      fechaInscripcion: "",
      estado: "activa",
    },
    {
      id: "e3",
      participanteId: "p3",
      cursoId: "c1",
      fechaInscripcion: "",
      estado: "baja",
    },
  ];

  const presente = (
    actividadId: string,
    participanteId: string,
  ): Attendance => ({
    id: `${actividadId}_${participanteId}`,
    actividadId,
    participanteId,
    presente: true,
    fecha: "",
  });

  beforeEach(async () => {
    educationServiceSpy = jasmine.createSpyObj("EducationService", [
      "listMyCourses",
      "listActivities",
      "listEnrollments",
      "listAttendance",
    ]);

    educationServiceSpy.listMyCourses.and.returnValue(of(mockCourses));
    educationServiceSpy.listActivities.and.returnValue(
      of([
        act("a1", "2026-10-05T00:00:00Z"),
        act("a2", "2026-10-12T00:00:00Z"),
      ]),
    );
    educationServiceSpy.listEnrollments.and.returnValue(of(enrollments));
    educationServiceSpy.listAttendance.and.callFake((id: string) =>
      of(
        id === "a1"
          ? [presente("a1", "p1"), presente("a1", "p2")]
          : [presente("a2", "p1")],
      ),
    );

    await TestBed.configureTestingModule({
      imports: [AsistenciaComponent, BrowserAnimationsModule],
      providers: [
        { provide: EducationService, useValue: educationServiceSpy },
        provideRouter([]),
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(AsistenciaComponent);
    component = fixture.componentInstance;
  });

  it("should load courses on init (>1 courses)", () => {
    fixture.detectChanges();
    expect(educationServiceSpy.listMyCourses).toHaveBeenCalled();
    expect(component.courseOptions).toHaveSize(2);
    expect(component.courseOptions[1].label).toBe("c2");
    expect(component.selectedCourseId).toBeNull();
  });

  it("should auto-select when there is a single course", () => {
    educationServiceSpy.listMyCourses.and.returnValue(of([mockCourses[0]]));
    fixture.detectChanges();
    expect(component.selectedCourseId).toBe("c1");
    expect(educationServiceSpy.listActivities).toHaveBeenCalledWith("c1");
  });

  it("builds the read-only matrix from enrollments and attendance", () => {
    fixture.detectChanges();
    component.selectedCourseId = "c1";
    component.onCourseChange();

    expect(component.activities.map((a) => a.id)).toEqual(["a1", "a2"]);
    expect(component.rows.map((r) => r.participanteId)).toEqual(["p1", "p2"]);
    const p1 = component.rows.find((r) => r.participanteId === "p1")!;
    const p2 = component.rows.find((r) => r.participanteId === "p2")!;
    expect([...p1.presentes].sort()).toEqual(["a1", "a2"]);
    expect([...p2.presentes]).toEqual(["a1"]);
    expect(component.loadingMatrix).toBeFalse();
  });

  it("muestra nombre y apellido en vez del id", () => {
    educationServiceSpy.listEnrollments.and.returnValue(
      of([{ ...enrollments[0], nombre: "Ana", apellido: "Pérez" }]),
    );
    fixture.detectChanges();
    component.selectedCourseId = "c1";
    component.onCourseChange();

    expect(component.rows[0].nombre).toBe("Ana Pérez");
  });

  it("handles a course without activities", () => {
    educationServiceSpy.listActivities.and.returnValue(of([]));
    fixture.detectChanges();
    component.selectedCourseId = "c1";
    component.onCourseChange();

    expect(educationServiceSpy.listAttendance).not.toHaveBeenCalled();
    expect(component.activities).toEqual([]);
    expect(component.rows).toHaveSize(2);
  });

  it("shows the error when loading fails", () => {
    educationServiceSpy.listEnrollments.and.returnValue(
      throwError(() => new Error("403")),
    );
    fixture.detectChanges();
    component.selectedCourseId = "c1";
    component.onCourseChange();

    expect(component.error).toBe("403");
    expect(component.loadingMatrix).toBeFalse();
  });

  it("does nothing without a selected course", () => {
    fixture.detectChanges();
    component.selectedCourseId = null;
    component.onCourseChange();
    expect(educationServiceSpy.listEnrollments).not.toHaveBeenCalled();
  });
});
