import { ComponentFixture, TestBed } from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { MessageService } from "primeng/api";
import { of, throwError } from "rxjs";

import { InscripcionesComponent } from "./inscripciones.component";
import { EducationService } from "../../../services/education.service";
import { Course, Enrollment } from "../../../core/models/education.model";

describe("InscripcionesComponent", () => {
  let fixture: ComponentFixture<InscripcionesComponent>;
  let component: InscripcionesComponent;
  let education: jasmine.SpyObj<EducationService>;
  let toast: MessageService;

  const courses = [
    { id: "c1", title: "Huerta" },
    { id: "c2", title: "" },
  ] as Course[];
  const enrollments: Enrollment[] = [
    {
      id: "e1",
      participanteId: "p1",
      cursoId: "c1",
      fechaInscripcion: "2026-09-01",
      estado: "activa",
      nombre: "Ana",
      apellido: "Paz",
    },
    {
      id: "e2",
      participanteId: "p2",
      cursoId: "c1",
      fechaInscripcion: "2026-09-01",
      estado: "baja",
      nombre: "Zoe",
      apellido: "Ríos",
    },
    {
      id: "e3",
      participanteId: "px",
      cursoId: "c1",
      fechaInscripcion: "2026-09-01",
      estado: "finalizada",
    },
  ];

  function create(opts: { failInit?: boolean } = {}) {
    education = jasmine.createSpyObj<EducationService>("EducationService", [
      "listCourses",
      "searchParticipants",
      "listCourseEnrollments",
      "enrollParticipant",
    ]);
    education.listCourses.and.returnValue(
      opts.failInit
        ? throwError(() => new Error("Contentful caído"))
        : of(courses),
    );
    education.searchParticipants.and.returnValue(of([]));
    education.listCourseEnrollments.and.returnValue(of(enrollments));

    TestBed.configureTestingModule({
      imports: [InscripcionesComponent, NoopAnimationsModule],
      providers: [{ provide: EducationService, useValue: education }],
    });
    fixture = TestBed.createComponent(InscripcionesComponent);
    component = fixture.componentInstance;
    toast = fixture.debugElement.injector.get(MessageService);
    spyOn(toast, "add");
    fixture.detectChanges();
  }

  it("loads only the courses: participants are searched in the backend", () => {
    create();

    expect(component.loading).toBeFalse();
    expect(component.courseOptions).toEqual([
      { label: "Huerta", value: "c1" },
      { label: "c2", value: "c2" },
    ]);
    expect(education.searchParticipants).not.toHaveBeenCalled();
  });

  it("shows the error when the initial load fails", () => {
    create({ failInit: true });

    expect(component.loading).toBeFalse();
    expect(component.error).toBe("Contentful caído");
  });

  it("loads enrollments of the selected course with resolved names", () => {
    create();
    component.selectedParticipant = {
      participanteId: "p3",
      nombre: "Leo",
      apellido: "Sosa",
    };
    component.selectedCourseId = "c1";
    component.onCourseChange();
    fixture.detectChanges();

    expect(education.listCourseEnrollments).toHaveBeenCalledWith("c1");
    expect(component.selectedParticipant).toBeNull();
    expect(component.enrollments.map((e) => e.nombreCompleto)).toEqual([
      "Paz, Ana",
      "Ríos, Zoe",
      "px",
    ]);
  });

  it("clears the list when no course is selected", () => {
    create();
    component.selectedCourseId = null;
    component.onCourseChange();

    expect(component.enrollments).toEqual([]);
    expect(education.listCourseEnrollments).not.toHaveBeenCalled();
  });

  it("shows the error when enrollments fail to load", () => {
    create();
    education.listCourseEnrollments.and.returnValue(
      throwError(() => new Error("firestore caído")),
    );
    component.selectedCourseId = "c1";
    component.onCourseChange();

    expect(component.loadingEnrollments).toBeFalse();
    expect(component.error).toBe("firestore caído");
  });

  it("enrolls the selected participant and reloads the list", () => {
    create();
    education.enrollParticipant.and.returnValue(of(enrollments[0]));
    component.selectedCourseId = "c1";
    component.selectedParticipant = {
      participanteId: "p1",
      nombre: "Ana",
      apellido: "Paz",
    };

    component.inscribir();

    expect(education.enrollParticipant).toHaveBeenCalledWith("c1", {
      participanteId: "p1",
    });
    expect(toast.add).toHaveBeenCalledWith(
      jasmine.objectContaining({ severity: "success", detail: "Paz, Ana" }),
    );
    expect(component.inscribiendo).toBeFalse();
    expect(component.selectedParticipant).toBeNull();
    expect(education.listCourseEnrollments).toHaveBeenCalledWith("c1");
  });

  it("shows the backend message when the enrollment is rejected", () => {
    create();
    education.enrollParticipant.and.returnValue(
      throwError(() => new Error("El curso no tiene cupo disponible")),
    );
    component.selectedCourseId = "c1";
    component.selectedParticipant = {
      participanteId: "p1",
      nombre: "Ana",
      apellido: "Paz",
    };

    component.inscribir();

    expect(component.inscribiendo).toBeFalse();
    expect(toast.add).toHaveBeenCalledWith(
      jasmine.objectContaining({
        severity: "error",
        detail: "El curso no tiene cupo disponible",
      }),
    );
  });

  it("does nothing without course and participant", () => {
    create();
    component.inscribir();
    expect(education.enrollParticipant).not.toHaveBeenCalled();
  });

  it("maps enrollment states to tag severities", () => {
    create();
    expect(component.estadoSeverity("activa")).toBe("success");
    expect(component.estadoSeverity("finalizada")).toBe("info");
    expect(component.estadoSeverity("baja")).toBe("danger");
    expect(component.estadoSeverity("otro")).toBe("secondary");
  });
});
