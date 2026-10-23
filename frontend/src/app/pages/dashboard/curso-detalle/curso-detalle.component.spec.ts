import { ComponentFixture, TestBed } from "@angular/core/testing";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import {
  ActivatedRoute,
  convertToParamMap,
  provideRouter,
} from "@angular/router";
import { ConfirmationService } from "primeng/api";
import { of, throwError } from "rxjs";

import { CursoDetalleComponent } from "./curso-detalle.component";
import { AuthService } from "../../../services/auth.service";
import { EducationService } from "../../../services/education.service";
import { UsersService } from "../../../services/users.service";
import { SkillsService } from "../../../services/skills.service";
import { Skill } from "../../../core/models/skills.model";
import { Activity, Course } from "../../../core/models/education.model";
import { ComentariosService } from "../../../services/comentarios.service";

describe("CursoDetalleComponent", () => {
  let fixture: ComponentFixture<CursoDetalleComponent>;
  let component: CursoDetalleComponent;
  let education: jasmine.SpyObj<EducationService>;
  let auth: jasmine.SpyObj<AuthService>;
  let skills: jasmine.SpyObj<SkillsService>;
  const excel: Skill = {
    id: "uso-de-excel",
    nombre: "Uso de Excel",
    categoria: "Tecnología",
  };

  const curso: Course = {
    id: "c1",
    title: "Panadería",
    description: null,
    mode: "Presencial",
    responsibleTeacher: "t1",
  };
  const act = (id: string, fecha: string): Activity => ({
    id,
    cursoId: "c1",
    titulo: `Actividad ${id}`,
    consigna: "Traer una receta",
    fecha,
    createdAt: fecha,
  });

  function setup(role: string) {
    auth.isAuthenticated.and.returnValue(
      of({ role_id: role, user: { user_id: "u1" } } as any),
    );
    fixture = TestBed.createComponent(CursoDetalleComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  /** Acepta en el acto cualquier confirmación del componente. */
  function autoConfirm() {
    const confirm = fixture.debugElement.injector.get(ConfirmationService);
    spyOn(confirm, "confirm").and.callFake((c: any) => {
      c.accept();
      return confirm;
    });
    return confirm;
  }

  beforeEach(async () => {
    education = jasmine.createSpyObj("EducationService", [
      "getCourse",
      "listDraftCourses",
      "listCourseActivities",
      "listActivities",
      "createCourseActivity",
      "createActivity",
      "updateCourseActivity",
      "updateActivity",
      "deleteCourseActivity",
      "deleteActivity",
      "listCourseActivitySubmissions",
      "listActivitySubmissions",
      "getMyCourse",
      "submitActivity",
      "downloadSubmissionFile",
    ]);
    auth = jasmine.createSpyObj("AuthService", ["isAuthenticated"]);
    skills = jasmine.createSpyObj("SkillsService", [
      "getCourseSkills",
      "setActivitySkills",
      "suggest",
      "searchCatalog",
    ]);
    skills.getCourseSkills.and.returnValue(
      of({ habilidades: [excel], porActividad: { a1: [excel] } }),
    );
    const users = jasmine.createSpyObj("UsersService", ["listUsers"]);
    users.listUsers.and.returnValue(
      of([{ user_id: "t1", first_name: "María", last_name: "Rodríguez" }]),
    );

    education.getCourse.and.returnValue(of(curso));
    education.listCourseActivities.and.returnValue(
      of([act("a1", "2026-10-05T00:00:00Z")]),
    );
    education.listActivities.and.returnValue(
      of([act("a1", "2026-10-05T00:00:00Z")]),
    );

    await TestBed.configureTestingModule({
      imports: [CursoDetalleComponent, BrowserAnimationsModule],
      providers: [
        provideRouter([]),
        { provide: EducationService, useValue: education },
        { provide: AuthService, useValue: auth },
        { provide: UsersService, useValue: users },
        { provide: SkillsService, useValue: skills },
        {
          provide: ComentariosService,
          useValue: { listar: () => of([]) },
        },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: convertToParamMap({ id: "c1" }) } },
        },
      ],
    }).compileComponents();
  });

  it("solo el administrador modera los comentarios del curso", () => {
    setup("role_volunteer");
    expect(component.esAdmin).toBeFalse();
    setup("role_admin");
    expect(component.esAdmin).toBeTrue();
  });

  it("gestión: carga el curso, sus actividades y el nombre del docente", () => {
    setup("role_volunteer");
    expect(component.puedeGestionar).toBeTrue();
    expect(component.curso?.title).toBe("Panadería");
    expect(component.actividades).toHaveSize(1);
    expect(component.docenteNombre).toBe("María Rodríguez");
    expect(education.listCourseActivities).toHaveBeenCalledWith("c1");
  });

  it("gestión: un admin puede abrir un curso en borrador", () => {
    education.getCourse.and.returnValue(
      throwError(() => new Error("no existe")),
    );
    education.listDraftCourses.and.returnValue(
      of([{ ...curso, title: "Borrador" }]),
    );
    setup("role_admin");
    expect(component.curso?.title).toBe("Borrador");
  });

  it("gestión: crea una actividad con la fecha como día calendario UTC", () => {
    setup("role_admin");
    education.createCourseActivity.and.returnValue(
      of(act("a0", "2026-10-01T00:00:00Z")),
    );

    component.nuevaActividad();
    component.form = {
      titulo: " Práctica ",
      consigna: "Amasar",
      fecha: new Date(2026, 9, 1, 23, 30),
    };
    component.guardarActividad();

    expect(education.createCourseActivity).toHaveBeenCalledWith("c1", {
      titulo: "Práctica",
      consigna: "Amasar",
      fecha: "2026-10-01T00:00:00Z",
    });
    expect(component.actividades.map((a) => a.id)).toEqual(["a0", "a1"]);
    expect(component.formVisible).toBeFalse();
  });

  it("gestión: no guarda sin consigna", () => {
    setup("role_admin");
    component.form = { titulo: "t", consigna: " ", fecha: new Date() };
    expect(component.formValido).toBeFalse();
    component.guardarActividad();
    expect(education.createCourseActivity).not.toHaveBeenCalled();
  });

  it("gestión: edita una actividad existente", () => {
    setup("role_admin");
    education.updateCourseActivity.and.returnValue(
      of({ ...act("a1", "2026-10-06T00:00:00Z"), titulo: "Editada" }),
    );
    component.editarActividad(component.actividades[0]);
    expect(component.form.fecha?.getDate()).toBe(5);

    component.guardarActividad();
    expect(education.updateCourseActivity).toHaveBeenCalled();
    expect(component.actividades[0].titulo).toBe("Editada");
  });

  it("gestión: elimina una actividad después de confirmar", () => {
    setup("role_admin");
    education.deleteCourseActivity.and.returnValue(of(undefined));
    const confirm = autoConfirm();

    component.eliminarActividad(component.actividades[0]);

    expect(confirm.confirm).toHaveBeenCalled();
    expect(education.deleteCourseActivity).toHaveBeenCalledWith("a1");
    expect(component.actividades).toEqual([]);
  });

  it("gestión: lista las entregas de una actividad", () => {
    setup("role_volunteer");
    education.listCourseActivitySubmissions.and.returnValue(
      of([
        {
          participanteId: "p1",
          nombre: "Ana",
          apellido: "Pérez",
          estado: "entregado",
        },
        { participanteId: "p2", estado: "sin_entregar" },
      ] as any),
    );
    component.verEntregas(component.actividades[0]);
    expect(component.entregas).toHaveSize(2);
    expect(component.entregadasCount).toBe(1);
    expect(component.nombreParticipante(component.entregas[0])).toBe(
      "Ana Pérez",
    );
    expect(component.nombreParticipante(component.entregas[1])).toBe("p2");
  });

  it("docente: usa los endpoints de /api/docente", () => {
    setup("role_teacher");
    education.deleteActivity.and.returnValue(of(undefined));
    autoConfirm();

    expect(education.listActivities).toHaveBeenCalledWith("c1");
    component.eliminarActividad(component.actividades[0]);
    expect(education.deleteActivity).toHaveBeenCalledWith("a1");
    expect(education.deleteCourseActivity).not.toHaveBeenCalled();
  });

  it("docente: curso ajeno muestra un mensaje claro", () => {
    education.listActivities.and.returnValue(
      throwError(
        () => new Error("no sos el docente responsable de este curso"),
      ),
    );
    setup("role_teacher");
    expect(component.error).toContain("docente responsable");
  });

  it("participante: ve sus actividades, su progreso y entrega", () => {
    const pendiente = act("a2", "2026-10-12T00:00:00Z");
    education.getMyCourse.and.returnValue(
      of({
        curso,
        actividades: [
          {
            ...act("a1", "2026-10-05T00:00:00Z"),
            entrega: { id: "a1_p1" } as any,
          },
          pendiente,
        ],
        progreso: {
          participanteId: "p1",
          actividadesTotal: 2,
          actividadesAsistidas: 1,
          porcentajeAsistencia: 50,
        },
      }),
    );
    setup("role_participant");
    expect(component.esParticipante).toBeTrue();
    expect(component.puedeGestionar).toBeFalse();
    expect(component.progreso?.porcentajeAsistencia).toBe(50);

    education.submitActivity.and.returnValue(
      of({
        id: "a2_p1",
        fechaEntrega: "2026-10-12T15:00:00Z",
        archivos: [],
        texto: "hecho",
      } as any),
    );
    autoConfirm();
    component.abrirEntrega(component.actividades[1]);
    component.entregaTexto = " hecho ";
    component.confirmarEntrega();

    expect(education.submitActivity).toHaveBeenCalledWith("a2", "hecho", []);
    expect(component.actividades[1].entrega?.id).toBe("a2_p1");
    expect(component.progreso?.porcentajeAsistencia).toBe(100);
  });

  it("participante: filtra adjuntos no permitidos o demasiado grandes", () => {
    education.getMyCourse.and.returnValue(
      of({ curso, actividades: [], progreso: {} as any }),
    );
    setup("role_participant");

    const input = document.createElement("input");
    const grande = new File(["x"], "grande.pdf");
    Object.defineProperty(grande, "size", { value: 11 * 1024 * 1024 });
    const files = [
      new File(["x"], "tp.pdf"),
      new File(["x"], "virus.exe"),
      grande,
    ];
    Object.defineProperty(input, "files", { value: files });
    component.onArchivosSeleccionados({ target: input } as any);

    expect(component.entregaArchivos.map((f) => f.name)).toEqual(["tp.pdf"]);
    component.quitarArchivo(0);
    expect(component.entregaArchivos).toEqual([]);
  });

  it("participante no inscripto: muestra el error del backend", () => {
    education.getMyCourse.and.returnValue(
      throwError(() => new Error("no estás inscripto en este curso")),
    );
    setup("role_participant");
    expect(component.error).toBe("no estás inscripto en este curso");
  });

  it("empresa: sin acceso", () => {
    setup("role_company");
    expect(component.error).toContain("no tiene acceso");
  });

  it("formatea tamaños e íconos de archivo", () => {
    setup("role_admin");
    expect(component.tamanoLegible(500)).toBe("500 B");
    expect(component.tamanoLegible(2048)).toBe("2 KB");
    expect(component.tamanoLegible(3 * 1024 * 1024)).toBe("3.0 MB");
    expect(component.iconoArchivo("application/pdf")).toBe("pi-file-pdf");
    expect(component.iconoArchivo("image/png")).toBe("pi-image");
  });

  describe("habilidades (HU-21)", () => {
    it("muestra las del curso y las de cada actividad", () => {
      setup("role_admin");
      expect(skills.getCourseSkills).toHaveBeenCalledWith("c1");
      expect(component.habilidadesCurso).toEqual([excel]);
      expect(component.habilidadesDe("a1")).toEqual([excel]);
      expect(component.habilidadesDe("otra")).toEqual([]);
      const texto = fixture.nativeElement.textContent as string;
      expect(texto).toContain("Habilidades que desarrolla este curso");
    });

    it("si no se pueden cargar, el detalle se muestra igual", () => {
      skills.getCourseSkills.and.returnValue(
        throwError(() => new Error("caído")),
      );
      setup("role_admin");
      expect(component.error).toBe("");
      expect(component.habilidadesCurso).toEqual([]);
    });

    it("al editar, carga las de la actividad y guarda solo si cambiaron", () => {
      setup("role_admin");
      education.updateCourseActivity.and.returnValue(
        of(act("a1", "2026-10-05T00:00:00Z")),
      );

      component.editarActividad(component.actividades[0]);
      expect(component.formSkills).toEqual([excel]);
      component.guardarActividad();
      expect(skills.setActivitySkills).not.toHaveBeenCalled();
      expect(component.formVisible).toBeFalse();

      skills.setActivitySkills.and.returnValue(of([]));
      component.editarActividad(component.actividades[0]);
      component.formSkills = [];
      component.guardarActividad();
      expect(skills.setActivitySkills).toHaveBeenCalledWith("a1", []);
    });

    it("una actividad nueva guarda sus habilidades con el id recién creado", () => {
      setup("role_teacher");
      education.createActivity.and.returnValue(
        of(act("nueva", "2026-10-20T00:00:00Z")),
      );
      skills.setActivitySkills.and.returnValue(of([excel]));

      component.nuevaActividad();
      component.form = { titulo: "t", consigna: "c", fecha: new Date() };
      component.formSkills = [
        { nombre: "Uso de Excel", categoria: "Tecnología" },
      ];
      component.guardarActividad();

      expect(skills.setActivitySkills).toHaveBeenCalledWith("nueva", [
        { nombre: "Uso de Excel", categoria: "Tecnología" },
      ]);
      expect(component.guardando).toBeFalse();
      expect(component.formVisible).toBeFalse();
    });

    it("si fallan las habilidades, avisa que la actividad sí se guardó", () => {
      setup("role_admin");
      education.createCourseActivity.and.returnValue(
        of(act("nueva", "2026-10-20T00:00:00Z")),
      );
      skills.setActivitySkills.and.returnValue(
        throwError(() => new Error("categoría inválida.")),
      );
      component.nuevaActividad();
      component.form = { titulo: "t", consigna: "c", fecha: new Date() };
      component.formSkills = [{ nombre: "X", categoria: "Oficio" }];
      component.guardarActividad();

      expect(component.actividades.map((a) => a.id)).toContain("nueva");
      expect(component.guardando).toBeFalse();
      expect(component.formVisible).toBeFalse();
    });
  });

  describe("Volver sin pantalla previa", () => {
    it("el voluntario vuelve al catálogo, no a gestión (solo admin)", () => {
      setup("role_volunteer");
      expect(component.destinoVolver()).toBe("/dashboard/cursos");
    });

    it("el administrador vuelve a gestión de cursos", () => {
      setup("role_admin");
      expect(component.destinoVolver()).toBe("/dashboard/gestion-cursos");
    });

    it("el docente vuelve a sus cursos", () => {
      setup("role_teacher");
      expect(component.destinoVolver()).toBe("/dashboard/mis-cursos");
    });

    it("el participante vuelve a sus cursos del catálogo", () => {
      education.getMyCourse.and.returnValue(
        of({
          curso,
          actividades: [],
          progreso: {
            participanteId: "p1",
            actividadesTotal: 0,
            actividadesAsistidas: 0,
            porcentajeAsistencia: 0,
          },
        }),
      );
      setup("role_participant");
      expect(component.destinoVolver()).toBe("/dashboard/cursos");
    });
  });
});
