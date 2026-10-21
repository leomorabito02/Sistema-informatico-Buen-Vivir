import { ComponentFixture, TestBed } from "@angular/core/testing";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import {
  ActivatedRoute,
  convertToParamMap,
  provideRouter,
} from "@angular/router";
import { of, throwError } from "rxjs";
import { FichaComponent } from "./ficha.component";
import {
  HumanPromotionError,
  HumanPromotionService,
} from "../../../../services/human-promotion.service";
import { AuthService } from "../../../../services/auth.service";
import { UsersService } from "../../../../services/users.service";
import { EducationService } from "../../../../services/education.service";
import {
  EntradaActividad,
  FichaDetalle,
  HitoDetalle,
} from "../../../../core/models/human-promotion.model";
import { Course } from "../../../../core/models/education.model";
import { MeResponse, User } from "../../../../core/models/auth.model";

const ficha: FichaDetalle = {
  id: "f1",
  participante_id: "p1",
  situacion_actual: "Busca trabajo",
  avances: "Terminó la secundaria",
  creado_por: "u1",
  updated_at: "2026-09-01T00:00:00Z",
  participante: { nombre: "Ana", apellido: "Pérez", edad: 30 },
  habilidades: ["Compostaje"],
};

const hito: HitoDetalle = {
  id: "h1",
  ficha_id: "f1",
  fecha: "2026-09-20",
  descripcion: "Empezó el taller de cocina",
  responsable_id: "u1",
  responsable_nombre: "Laura Díaz",
  visible_participante: false,
};

const inscripcion: EntradaActividad = {
  id: "i1",
  tipo: "inscripcion",
  fecha: "2026-09-02",
  descripcion: "Se inscribió en un curso",
  curso_id: "c1",
};

const yo = {
  user_id: "u1",
  first_name: "Laura",
  last_name: "Díaz",
  role_id: "role_volunteer",
} as User;

describe("FichaComponent", () => {
  let fixture: ComponentFixture<FichaComponent>;
  let component: FichaComponent;
  let service: jasmine.SpyObj<HumanPromotionService>;

  beforeEach(async () => {
    service = jasmine.createSpyObj<HumanPromotionService>(
      "HumanPromotionService",
      [
        "getFichaPorParticipante",
        "listarHitos",
        "listarActividad",
        "crearFicha",
        "editarFicha",
        "crearHito",
        "editarHito",
        "descargarPdf",
      ],
    );
    service.getFichaPorParticipante.and.returnValue(of(ficha));
    service.listarHitos.and.returnValue(of([hito]));
    service.listarActividad.and.returnValue(of([inscripcion]));

    const auth = jasmine.createSpyObj<AuthService>("AuthService", [
      "isAuthenticated",
    ]);
    auth.isAuthenticated.and.returnValue(of({ user: yo } as MeResponse));
    const users = jasmine.createSpyObj<UsersService>("UsersService", [
      "listUsers",
    ]);
    users.listUsers.and.returnValue(of([]));
    const education = jasmine.createSpyObj<EducationService>(
      "EducationService",
      ["listCourses"],
    );
    education.listCourses.and.returnValue(
      of([{ id: "c1", title: "Huerta agroecológica" } as Course]),
    );

    await TestBed.configureTestingModule({
      imports: [FichaComponent, BrowserAnimationsModule],
      providers: [
        provideRouter([]),
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: { paramMap: convertToParamMap({ participanteId: "p1" }) },
          },
        },
        { provide: HumanPromotionService, useValue: service },
        { provide: AuthService, useValue: auth },
        { provide: UsersService, useValue: users },
        { provide: EducationService, useValue: education },
      ],
    }).compileComponents();
  });

  function crear(): HTMLElement {
    fixture = TestBed.createComponent(FichaComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
    return fixture.nativeElement;
  }

  it("muestra la ficha con los datos de Auth y las habilidades (HU-31)", () => {
    const el = crear();
    expect(service.getFichaPorParticipante).toHaveBeenCalledWith("p1");
    expect(el.textContent).toContain("Ana Pérez");
    expect(el.textContent).toContain("Busca trabajo");
    expect(el.textContent).toContain("Compostaje");
  });

  it("mezcla por fecha los hitos con la actividad en el centro (HU-33)", () => {
    const el = crear();
    const items = el.querySelectorAll(".timeline-item");
    expect(items.length).toBe(2);
    expect(items[0].classList).toContain("timeline-item--actividad");
    expect(items[0].textContent).toContain("Huerta agroecológica");
    expect(items[0].textContent).not.toContain("Editar");
    expect(items[1].textContent).toContain("Empezó el taller de cocina");
    expect(items[1].textContent).toContain("Confidencial");
    expect(items[1].textContent).toContain("Laura Díaz");
  });

  it("si la actividad no se puede leer, muestra igual los hitos", () => {
    service.listarActividad.and.returnValue(
      throwError(() => new Error("caído")),
    );
    const el = crear();
    expect(el.querySelectorAll(".timeline-item").length).toBe(1);
    expect(el.textContent).not.toContain("No se pudo");
  });

  it("sin ficha (404) ofrece el alta", () => {
    service.getFichaPorParticipante.and.returnValue(
      throwError(() => new HumanPromotionError("recurso no encontrado", 404)),
    );
    crear();
    expect(component.sinFicha).toBeTrue();
    expect(component.editando).toBeTrue();
    expect(service.listarHitos).not.toHaveBeenCalled();
  });

  it("otro error muestra el mensaje", () => {
    service.getFichaPorParticipante.and.returnValue(
      throwError(() => new HumanPromotionError("No tenés permisos", 403)),
    );
    const el = crear();
    expect(component.error).toBe("No tenés permisos");
    expect(el.textContent).toContain("No tenés permisos");
  });

  it("no guarda la ficha si faltan campos obligatorios (HU-29)", () => {
    crear();
    component.empezarEdicion();
    component.situacionActual = "  ";
    component.guardarFicha();
    expect(service.editarFicha).not.toHaveBeenCalled();
  });

  it("edita la ficha con los textos recortados", () => {
    service.editarFicha.and.returnValue(
      of({
        message: "Cambios guardados",
        ficha: { ...ficha, avances: "Nuevo" },
      }),
    );
    crear();
    component.empezarEdicion();
    component.avances = "  Nuevo  ";
    component.guardarFicha();
    expect(service.editarFicha).toHaveBeenCalledWith("f1", {
      situacion_actual: "Busca trabajo",
      avances: "Nuevo",
    });
    expect(component.editando).toBeFalse();
    expect(component.ficha?.avances).toBe("Nuevo");
  });

  it("registra un hito con el usuario logueado como responsable (HU-32)", () => {
    service.crearHito.and.returnValue(of(hito));
    crear();
    component.nuevoHito();
    component.hitoForm.descripcion = "  Primera entrevista ";
    component.guardarHito();
    const [fichaId, input] = service.crearHito.calls.mostRecent().args;
    expect(fichaId).toBe("f1");
    expect(input.descripcion).toBe("Primera entrevista");
    expect(input.responsable_id).toBe("u1");
    expect(input.visible_participante).toBeFalse();
    // Recarga la línea de tiempo.
    expect(service.listarHitos).toHaveBeenCalledTimes(2);
  });

  it("no registra un hito sin descripción", () => {
    crear();
    component.nuevoHito();
    component.guardarHito();
    expect(service.crearHito).not.toHaveBeenCalled();
  });

  it("al editar un hito precarga sus datos", () => {
    service.editarHito.and.returnValue(of(hito));
    crear();
    component.editarHito(hito);
    expect(component.hitoEditandoId).toBe("h1");
    expect(component.hitoForm.descripcion).toBe("Empezó el taller de cocina");
    component.guardarHito();
    expect(service.editarHito).toHaveBeenCalledWith(
      "f1",
      "h1",
      jasmine.objectContaining({ fecha: "2026-09-20" }),
    );
  });
});
