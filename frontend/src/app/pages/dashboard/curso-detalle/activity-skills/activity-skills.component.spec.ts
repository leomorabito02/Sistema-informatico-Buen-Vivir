import { ComponentFixture, TestBed } from "@angular/core/testing";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { MessageService } from "primeng/api";
import { of, throwError } from "rxjs";

import {
  ActivitySkillsComponent,
  MAX_SKILLS_PER_ACTIVITY,
} from "./activity-skills.component";
import {
  SkillsService,
  SUGGESTIONS_UNAVAILABLE,
} from "../../../../services/skills.service";
import {
  Skill,
  SkillInput,
  SkillSuggestion,
} from "../../../../core/models/skills.model";

describe("ActivitySkillsComponent", () => {
  let fixture: ComponentFixture<ActivitySkillsComponent>;
  let component: ActivitySkillsComponent;
  let api: jasmine.SpyObj<SkillsService>;
  let toast: MessageService;
  let emitted: SkillInput[] | null;

  const huerta: SkillSuggestion = {
    nombre: "Armado de huerta",
    categoria: "Cuidado de la naturaleza",
    justificacion: "Arman un cantero.",
    existente: false,
  };
  const equipo: SkillSuggestion = {
    id: "trabajo-en-equipo",
    nombre: "Trabajo en equipo",
    categoria: "Socioemocional",
    justificacion: "Lo hacen en grupo.",
    existente: true,
  };
  const excel: Skill = {
    id: "uso-de-excel",
    nombre: "Uso de Excel",
    categoria: "Tecnología",
  };

  beforeEach(async () => {
    api = jasmine.createSpyObj("SkillsService", ["suggest", "searchCatalog"]);
    await TestBed.configureTestingModule({
      imports: [ActivitySkillsComponent, BrowserAnimationsModule],
      providers: [MessageService, { provide: SkillsService, useValue: api }],
    }).compileComponents();

    fixture = TestBed.createComponent(ActivitySkillsComponent);
    component = fixture.componentInstance;
    toast = TestBed.inject(MessageService);
    spyOn(toast, "add");
    component.cursoId = "c1";
    emitted = null;
    component.skillsChange.subscribe((l) => (emitted = l));
    fixture.detectChanges();
  });

  function completar() {
    component.titulo = "Canteros";
    component.consigna = "Armen un cantero en grupo";
  }

  it("el botón de la IA se habilita recién con título y consigna", () => {
    expect(component.puedeSugerir).toBeFalse();
    component.titulo = "Canteros";
    expect(component.puedeSugerir).toBeFalse();
    component.consigna = "  ";
    expect(component.puedeSugerir).toBeFalse();
    component.consigna = "Armen un cantero";
    expect(component.puedeSugerir).toBeTrue();
  });

  it("no pide sugerencias si falta la consigna", () => {
    component.titulo = "Canteros";
    component.sugerir();
    expect(api.suggest).not.toHaveBeenCalled();
  });

  it("pide sugerencias excluyendo las que ya están en el formulario", () => {
    completar();
    component.actividadId = "a1";
    component.skills = [{ nombre: "Uso de Excel", categoria: "Tecnología" }];
    api.suggest.and.returnValue(
      of({
        proveedor: "gemini",
        sugerencias: [
          huerta,
          { ...equipo },
          { ...huerta, nombre: "uso de excel" },
        ],
      }),
    );

    component.sugerir();

    expect(api.suggest).toHaveBeenCalledWith({
      cursoId: "c1",
      actividadId: "a1",
      titulo: "Canteros",
      consigna: "Armen un cantero en grupo",
      excluir: ["Uso de Excel"],
    });
    expect(component.sugerencias.map((s) => s.nombre)).toEqual([
      "Armado de huerta",
      "Trabajo en equipo",
    ]);
    expect(component.cargando).toBeFalse();
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain("Lo hacen en grupo.");
  });

  it("muestra el vacío si la IA no encontró nada nuevo", () => {
    completar();
    api.suggest.and.returnValue(of({ proveedor: "openai", sugerencias: [] }));
    component.sugerir();
    expect(component.sinSugerencias).toBeTrue();
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain(
      "No encontramos habilidades nuevas para sugerir",
    );
  });

  it("muestra el error si ninguna IA respondió", () => {
    completar();
    api.suggest.and.returnValue(
      throwError(() => new Error(SUGGESTIONS_UNAVAILABLE)),
    );
    component.sugerir();
    expect(component.errorSugerencias).toBe(SUGGESTIONS_UNAVAILABLE);
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain(
      "Podés cargarlas a mano",
    );
  });

  it("aceptar agrega la sugerencia (con el id del catálogo) y descartar la saca", () => {
    component.sugerencias = [huerta, equipo];
    component.aceptar(equipo);
    expect(emitted).toEqual([
      {
        id: "trabajo-en-equipo",
        nombre: "Trabajo en equipo",
        categoria: "Socioemocional",
      },
    ]);
    expect(component.sugerencias).toEqual([huerta]);

    component.descartar(huerta);
    expect(component.sugerencias).toEqual([]);
    expect(component.skills).toHaveSize(1);
  });

  it("aceptar todas respeta el máximo por actividad", () => {
    component.skills = Array.from(
      { length: MAX_SKILLS_PER_ACTIVITY - 1 },
      (_, i) => ({
        nombre: "H" + i,
        categoria: "Oficio" as const,
      }),
    );
    component.sugerencias = [huerta, equipo];

    component.aceptarTodas();

    expect(component.skills).toHaveSize(MAX_SKILLS_PER_ACTIVITY);
    expect(component.sugerencias).toEqual([equipo]);
    expect(toast.add).toHaveBeenCalledWith(
      jasmine.objectContaining({ summary: "Límite de habilidades" }),
    );

    component.aceptar(equipo);
    expect(component.skills).toHaveSize(MAX_SKILLS_PER_ACTIVITY);
  });

  it("quitar saca una habilidad asignada", () => {
    const a = { nombre: "A", categoria: "Oficio" as const };
    const b = { nombre: "B", categoria: "Oficio" as const };
    component.skills = [a, b];
    component.quitar(a);
    expect(emitted).toEqual([b]);
  });

  it("el autocompletado no ofrece las que ya están asignadas", () => {
    component.skills = [{ nombre: "Uso de Excel", categoria: "Tecnología" }];
    api.searchCatalog.and.returnValue(
      of([
        excel,
        { id: "compostaje", nombre: "Compostaje", categoria: "Oficio" },
      ]),
    );
    component.buscarCatalogo({ query: "o" });
    expect(component.catalogoFiltrado.map((s) => s.nombre)).toEqual([
      "Compostaje",
    ]);

    api.searchCatalog.and.returnValue(throwError(() => new Error("caído")));
    component.buscarCatalogo({ query: "x" });
    expect(component.catalogoFiltrado).toEqual([]);
  });

  it("agregar a mano una del catálogo reusa su categoría", () => {
    component.manual = excel;
    expect(component.puedeAgregarManual).toBeTrue();
    component.agregarManual();
    expect(emitted).toEqual([
      { id: "uso-de-excel", nombre: "Uso de Excel", categoria: "Tecnología" },
    ]);
    expect(component.manual).toBeNull();
  });

  it("escribir el nombre de una del catálogo también la reusa", () => {
    component.catalogoFiltrado = [excel];
    component.manual = "uso de  excel";
    expect(component.manualDelCatalogo).toEqual(excel);
    component.agregarManual();
    expect(emitted![0].id).toBe("uso-de-excel");
  });

  it("una habilidad nueva pide la categoría", () => {
    component.manual = "Compostaje";
    expect(component.puedeAgregarManual).toBeFalse();
    component.manualCategoria = "Cuidado de la naturaleza";
    expect(component.puedeAgregarManual).toBeTrue();
    component.agregarManual();
    expect(emitted).toEqual([
      { nombre: "Compostaje", categoria: "Cuidado de la naturaleza" },
    ]);
    expect(component.manualCategoria).toBeNull();
  });

  it("no agrega una repetida escrita distinto", () => {
    component.skills = [{ nombre: "Compostaje", categoria: "Oficio" }];
    component.manual = "COMPOSTAJE";
    component.manualCategoria = "Oficio";
    component.agregarManual();
    expect(emitted).toBeNull();
    expect(toast.add).toHaveBeenCalledWith(
      jasmine.objectContaining({ summary: "Ya está en la lista" }),
    );
  });

  it("sin nombre no se puede agregar", () => {
    component.manual = "   ";
    expect(component.manualNombre).toBe("");
    expect(component.puedeAgregarManual).toBeFalse();
    component.agregarManual();
    expect(emitted).toBeNull();
  });
});
