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

  describe("por participante (HU-19)", () => {
    const historial = [
      {
        cursoId: "c1",
        cursoTitulo: "Huerta",
        estadoInscripcion: "activa",
        fechaInscripcion: "2026-10-01T00:00:00Z",
        actividadesTotal: 2,
        actividadesAsistidas: 1,
        porcentajeAsistencia: 50,
        actividades: [
          {
            actividadId: "a1",
            titulo: "TP 1",
            fecha: "2026-10-01T00:00:00Z",
            presente: true,
          },
          {
            actividadId: "a2",
            titulo: "TP 2",
            fecha: "2026-10-08T00:00:00Z",
            presente: false,
          },
        ],
      },
    ] as any;

    it("loads the history of the chosen participant", () => {
      create();
      education.getParticipantAttendanceHistory.and.returnValue(of(historial));
      component.vista = "participante";
      component.onParticipanteChange({
        participanteId: "p1",
        nombre: "Ana",
        apellido: "Paz",
      });
      fixture.detectChanges();

      expect(education.getParticipantAttendanceHistory).toHaveBeenCalledWith(
        "p1",
      );
      expect(component.loadingHistorial).toBeFalse();
      expect(component.nombreParticipante).toBe("Paz, Ana");
      const text = fixture.nativeElement.textContent;
      expect(text).toContain("Huerta");
      expect(text).toContain("asistió a 1 de");
      expect(text).toContain("Sin entregar");
    });

    it("clears the history when the participant is cleared", () => {
      create();
      component.historial = historial;
      component.onParticipanteChange(null);
      expect(component.historial).toEqual([]);
      expect(education.getParticipantAttendanceHistory).not.toHaveBeenCalled();
    });

    it("shows the error when the history fails", () => {
      create();
      education.getParticipantAttendanceHistory.and.returnValue(
        throwError(() => new Error("firestore caído")),
      );
      component.onParticipanteChange({
        participanteId: "p1",
        nombre: "",
        apellido: "",
      });
      expect(component.loadingHistorial).toBeFalse();
      expect(component.error).toBe("firestore caído");
    });

    it("maps enrollment states to tag severities", () => {
      create();
      expect(component.estadoSeverity("activa")).toBe("success");
      expect(component.estadoSeverity("finalizada")).toBe("info");
      expect(component.estadoSeverity("baja")).toBe("danger");
      expect(component.estadoSeverity("x")).toBe("secondary");
    });
  });
});
