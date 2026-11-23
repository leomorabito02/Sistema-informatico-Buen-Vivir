import { ComponentFixture, TestBed } from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { provideRouter } from "@angular/router";
import { of, throwError } from "rxjs";

import { AsistenciaCursosComponent } from "./asistencia-cursos.component";
import { EducationService } from "../../../services/education.service";
import {
  Activity,
  Attendance,
  Course,
  Enrollment,
} from "../../../core/models/education.model";

describe("AsistenciaCursosComponent", () => {
  let fixture: ComponentFixture<AsistenciaCursosComponent>;
  let component: AsistenciaCursosComponent;
  let education: jasmine.SpyObj<EducationService>;

  const activities = [
    {
      id: "a1",
      cursoId: "c1",
      titulo: "TP 1",
      consigna: "",
      fecha: "2026-10-01T00:00:00Z",
      createdAt: "",
    },
    {
      id: "a2",
      cursoId: "c1",
      titulo: "TP 2",
      consigna: "",
      fecha: "2026-10-08T00:00:00Z",
      createdAt: "",
    },
  ] as Activity[];
  const enrollments = [
    {
      id: "e1",
      participanteId: "p1",
      cursoId: "c1",
      fechaInscripcion: "",
      estado: "activa",
      nombre: "Ana",
      apellido: "Paz",
    },
    {
      id: "e2",
      participanteId: "p2",
      cursoId: "c1",
      fechaInscripcion: "",
      estado: "baja",
    },
  ] as Enrollment[];
  const att = (actividadId: string, participanteId: string): Attendance => ({
    id: `${actividadId}_${participanteId}`,
    actividadId,
    participanteId,
    presente: true,
    fecha: "",
  });

  function create(opts: { coursesFail?: boolean } = {}) {
    education = jasmine.createSpyObj<EducationService>("EducationService", [
      "listCourses",
      "getParticipantAttendanceHistory",
      "searchParticipants",
      "listCourseEnrollments",
      "listCourseActivities",
      "listActivityAttendance",
    ]);
    education.listCourses.and.returnValue(
      opts.coursesFail
        ? throwError(() => new Error("Contentful caído"))
        : of([{ id: "c1", title: "Huerta" }] as Course[]),
    );
    education.listCourseEnrollments.and.returnValue(of(enrollments));
    education.listCourseActivities.and.returnValue(of(activities));
    education.listActivityAttendance.and.callFake((id: string) =>
      of(id === "a1" ? [att("a1", "p1"), att("a1", "p2")] : [att("a2", "p1")]),
    );

    TestBed.configureTestingModule({
      imports: [AsistenciaCursosComponent, NoopAnimationsModule],
      providers: [
        { provide: EducationService, useValue: education },
        provideRouter([]),
      ],
    });
    fixture = TestBed.createComponent(AsistenciaCursosComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  it("loads the course options", () => {
    create();
    expect(component.loadingCourses).toBeFalse();
    expect(component.courseOptions).toEqual([{ label: "Huerta", value: "c1" }]);
  });

  it("shows the error when courses fail to load", () => {
    create({ coursesFail: true });
    expect(component.loadingCourses).toBeFalse();
    expect(component.error).toBe("Contentful caído");
  });

  it("builds the attendance matrix of the selected course", () => {
    create();
    component.selectedCourseId = "c1";
    component.onCourseChange();
    fixture.detectChanges();

    expect(education.listActivityAttendance).toHaveBeenCalledTimes(2);
    expect(component.loadingMatrix).toBeFalse();
    expect(component.activities).toEqual(activities);
    // p2 está de baja pero entregó a1: igual aparece; sin nombre resuelto queda el id
    expect(component.rows.map((r) => r.nombre)).toEqual(["p2", "Paz, Ana"]);
    expect([...component.rows[1].presentes]).toEqual(["a1", "a2"]);
    expect([...component.rows[0].presentes]).toEqual(["a1"]);
  });

  it("handles a course without activities", () => {
    create();
    education.listCourseActivities.and.returnValue(of([]));
    component.selectedCourseId = "c1";
    component.onCourseChange();

    expect(education.listActivityAttendance).not.toHaveBeenCalled();
    expect(component.rows.map((r) => r.participanteId)).toEqual(["p1"]);
  });

  it("shows the error when the matrix fails to load", () => {
    create();
    education.listCourseEnrollments.and.returnValue(
      throwError(() => new Error("firestore caído")),
    );
    component.selectedCourseId = "c1";
    component.onCourseChange();

    expect(component.loadingMatrix).toBeFalse();
    expect(component.error).toBe("firestore caído");
  });

  it("clears the matrix when the course is deselected", () => {
    create();
    component.selectedCourseId = null;
    component.onCourseChange();
    expect(component.rows).toEqual([]);
    expect(education.listCourseEnrollments).not.toHaveBeenCalled();
  });

  // No hay botón "Por participante".
  it("ya no ofrece la vista por participante", () => {
    create();
    fixture.detectChanges();
    const text: string = fixture.nativeElement.textContent;
    expect(text).not.toContain("Por participante");
    expect(fixture.nativeElement.querySelector("p-selectbutton")).toBeNull();
    expect(
      fixture.nativeElement.querySelector("app-participant-picker"),
    ).toBeNull();
  });
});
