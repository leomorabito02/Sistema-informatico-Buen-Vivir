import { ComponentFixture, TestBed } from "@angular/core/testing";
import { SeguimientoComponent } from "./seguimiento.component";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { EducationService } from "../../../services/education.service";
import { AuthService } from "../../../services/auth.service";
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
  let authSpy: jasmine.SpyObj<AuthService>;

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
      "listCourses",
      "listCourseEnrollments",
      "getCourseProgressGestion",
    ]);
    authSpy = jasmine.createSpyObj("AuthService", ["isAuthenticated"]);
    authSpy.isAuthenticated.and.returnValue(
      of({ role_id: "role_teacher", user: { user_id: "t1" } } as any),
    );

    educationServiceSpy.listMyCourses.and.returnValue(of(mockCourses));
    educationServiceSpy.listEnrollments.and.returnValue(of(mockEnrollments));
    educationServiceSpy.getCourseProgress.and.returnValue(of(mockProgress));

    await TestBed.configureTestingModule({
      imports: [SeguimientoComponent, BrowserAnimationsModule],
      providers: [
        { provide: EducationService, useValue: educationServiceSpy },
        { provide: AuthService, useValue: authSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(SeguimientoComponent);
    component = fixture.componentInstance;
  });

  it("should create", () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it("al entrar muestra todos los participantes de todos los cursos", () => {
    educationServiceSpy.listEnrollments.and.callFake((id: string) =>
      of(
        id === "c1"
          ? mockEnrollments
          : [{ ...mockEnrollments[0], id: "e9", cursoId: "c2" }],
      ),
    );
    fixture.detectChanges();

    // "Todos los cursos" + los 2 cursos
    expect(component.courseOptions).toHaveSize(3);
    expect(component.selectedCourseId).toBe("");
    expect(educationServiceSpy.listEnrollments).toHaveBeenCalledWith("c1", {
      silent: true,
    });
    expect(educationServiceSpy.listEnrollments).toHaveBeenCalledWith("c2", {
      silent: true,
    });
    expect(component.loading).toBeFalse();
    // 4 inscripciones en c1 + 1 en c2: una fila por inscripción
    expect(component.filasFiltradas).toHaveSize(5);
    const text: string = fixture.nativeElement.textContent;
    expect(text).toContain("Course 1");
    expect(text).toContain("Course 2");
    expect(text).not.toContain("Cruce de inscripciones");
  });

  it("calcula el riesgo según la asistencia", () => {
    educationServiceSpy.listMyCourses.and.returnValue(of([mockCourses[0]]));
    fixture.detectChanges();
    const porId = new Map(component.rows.map((r) => [r.participanteId, r]));
    expect(porId.get("p1")!.riesgo).toBe("En riesgo"); // 40 %
    expect(porId.get("p2")!.riesgo).toBe("Regular"); // 60 %
    expect(porId.get("p3")!.riesgo).toBe("Buen avance"); // 80 %
    expect(porId.get("p4")!.riesgo).toBe("Sin datos");
    expect(porId.get("p4")!.porcentajeAsistencia).toBe(0);
    expect(porId.get("p4")!.actividadesAsistidas).toBe(0);
  });

  it("el curso y el buscador filtran sin cambiar de vista", () => {
    educationServiceSpy.listEnrollments.and.callFake((id: string) =>
      of(
        id === "c1"
          ? [{ ...mockEnrollments[0], nombre: "José", apellido: "Pérez" }]
          : [
              {
                ...mockEnrollments[1],
                cursoId: "c2",
                nombre: "Ana",
                apellido: "Gómez",
              },
            ],
      ),
    );
    fixture.detectChanges();
    expect(component.filasFiltradas).toHaveSize(2);

    component.selectedCourseId = "c2";
    expect(component.filasFiltradas.map((r) => r.nombre)).toEqual([
      "Ana Gómez",
    ]);

    component.selectedCourseId = "";
    component.busqueda = "jose perez";
    expect(component.filasFiltradas.map((r) => r.nombre)).toEqual([
      "José Pérez",
    ]);
    expect(component.hayFiltros).toBeTrue();

    component.limpiarFiltros();
    expect(component.filasFiltradas).toHaveSize(2);
  });

  it("muestra el error si no se pueden cargar los cursos", () => {
    educationServiceSpy.listMyCourses.and.returnValue(
      throwError(() => new Error("Error loading courses")),
    );
    fixture.detectChanges();
    expect(component.error).toBe("Error loading courses");
    expect(component.loading).toBeFalse();
  });

  it("si un curso falla, muestra los demás y avisa cuál faltó", () => {
    educationServiceSpy.listEnrollments.and.callFake((id: string) =>
      id === "c1"
        ? throwError(() => new Error("Error enrollments"))
        : of([{ ...mockEnrollments[0], cursoId: "c2" }]),
    );
    fixture.detectChanges();
    expect(component.cursosConError).toEqual(["Course 1"]);
    expect(component.rows).toHaveSize(1);
    expect(component.rows[0].cursoId).toBe("c2");
    expect(component.loading).toBeFalse();
  });

  it("sin cursos muestra el aviso vacío", () => {
    educationServiceSpy.listMyCourses.and.returnValue(of([]));
    fixture.detectChanges();
    expect(component.rows).toEqual([]);
    expect(fixture.nativeElement.textContent).toContain(
      "No tenés cursos asignados",
    );
  });

  it("should return correct risk severities", () => {
    expect(component.riesgoSeverity("Buen avance")).toBe("success");
    expect(component.riesgoSeverity("Regular")).toBe("info");
    expect(component.riesgoSeverity("En riesgo")).toBe("danger");
    expect(component.riesgoSeverity("Sin datos")).toBe("secondary");
    expect(component.riesgoSeverity("Otro")).toBe("secondary");
  });

  // HU-20: admin y voluntario usan el catálogo y los endpoints de gestión, no /api/docente.
  for (const rol of ["role_volunteer", "role_admin"]) {
    it(`${rol}: usa el catálogo y el progreso de gestión, con nombres`, () => {
      authSpy.isAuthenticated.and.returnValue(of({ role_id: rol } as any));
      educationServiceSpy.listCourses.and.returnValue(of([mockCourses[0]]));
      educationServiceSpy.listCourseEnrollments.and.returnValue(
        of([{ ...mockEnrollments[0], nombre: "Ana", apellido: "Pérez" }]),
      );
      educationServiceSpy.getCourseProgressGestion.and.returnValue(
        of([mockProgress[0]]),
      );

      fixture.detectChanges();

      expect(component.gestion).toBeTrue();
      expect(educationServiceSpy.listMyCourses).not.toHaveBeenCalled();
      expect(educationServiceSpy.listEnrollments).not.toHaveBeenCalled();
      expect(educationServiceSpy.getCourseProgressGestion).toHaveBeenCalledWith(
        "c1",
        { silent: true },
      );
      expect(component.rows[0].nombre).toBe("Ana Pérez");
      expect(component.rows[0].riesgo).toBe("En riesgo");
    });
  }

  it("docente: sigue usando sus cursos (/api/docente)", () => {
    fixture.detectChanges();
    expect(component.gestion).toBeFalse();
    expect(educationServiceSpy.listMyCourses).toHaveBeenCalled();
    expect(educationServiceSpy.listCourses).not.toHaveBeenCalled();
  });
});
