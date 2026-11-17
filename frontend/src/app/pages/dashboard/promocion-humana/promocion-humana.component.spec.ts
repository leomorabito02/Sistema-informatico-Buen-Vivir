import { ComponentFixture, TestBed } from "@angular/core/testing";
import { provideRouter } from "@angular/router";
import { of, throwError } from "rxjs";
import { PromocionHumanaComponent } from "./promocion-humana.component";
import { HumanPromotionService } from "../../../services/human-promotion.service";
import { ResultadoBusqueda } from "../../../core/models/human-promotion.model";

const conFicha: ResultadoBusqueda = {
  participante_id: "p1",
  nombre: "Ana",
  apellido: "Pérez",
  ficha_id: "f1",
};
const sinFicha: ResultadoBusqueda = {
  participante_id: "p2",
  nombre: "Bruno",
  apellido: "Gómez",
  ficha_id: null,
};

describe("PromocionHumanaComponent", () => {
  let fixture: ComponentFixture<PromocionHumanaComponent>;
  let component: PromocionHumanaComponent;
  let service: jasmine.SpyObj<HumanPromotionService>;

  function crear(respuesta = of([conFicha, sinFicha])) {
    service.listarParticipantes.and.returnValue(respuesta);
    fixture = TestBed.createComponent(PromocionHumanaComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  function filas(): NodeListOf<HTMLElement> {
    fixture.detectChanges();
    return fixture.nativeElement.querySelectorAll(".result-item");
  }

  beforeEach(async () => {
    service = jasmine.createSpyObj<HumanPromotionService>(
      "HumanPromotionService",
      ["listarParticipantes", "buscarParticipantes"],
    );
    await TestBed.configureTestingModule({
      imports: [PromocionHumanaComponent],
      providers: [
        provideRouter([]),
        { provide: HumanPromotionService, useValue: service },
      ],
    }).compileComponents();
  });

  it("al entrar muestra el listado completo, sin buscar (17/11/2026)", () => {
    crear();
    expect(service.listarParticipantes).toHaveBeenCalledTimes(1);
    expect(service.buscarParticipantes).not.toHaveBeenCalled();
    expect(filas().length).toBe(2);
    expect(fixture.nativeElement.textContent).toContain("2 de 2 participantes");
  });

  it("el buscador filtra el listado sin acentos ni mayúsculas", () => {
    crear();
    component.texto = "PEREZ";
    expect(filas().length).toBe(1);
    expect(filas()[0].textContent).toContain("Ana Pérez");

    component.texto = "";
    expect(filas().length).toBe(2);
    // El filtro es local: no se vuelve a consultar al backend.
    expect(service.listarParticipantes).toHaveBeenCalledTimes(1);
  });

  it("indica quién tiene ficha y ofrece verla o crearla", () => {
    crear();
    const f = filas();
    expect(f[0].textContent).toContain("Con ficha");
    expect(f[0].querySelector("a")?.getAttribute("aria-label")).toBe(
      "Ver ficha de Ana Pérez",
    );
    expect(f[1].textContent).toContain("Sin ficha");
    expect(f[1].querySelector("a")?.getAttribute("aria-label")).toBe(
      "Crear ficha para Bruno Gómez",
    );
  });

  it("avisa cuando la búsqueda no coincide con nadie", () => {
    crear();
    component.texto = "zzz";
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain(
      "Ningún participante coincide",
    );
  });

  it("avisa cuando no hay participantes", () => {
    crear(of([]));
    expect(fixture.nativeElement.textContent).toContain(
      "Todavía no hay participantes",
    );
  });

  it("muestra el error del servicio", () => {
    crear(throwError(() => new Error("No tenés permisos")));
    expect(component.error).toBe("No tenés permisos");
    expect(component.cargando).toBeFalse();
    expect(fixture.nativeElement.textContent).toContain("No tenés permisos");
  });

  it("usa el id si el participante no tiene nombre cargado", () => {
    crear();
    expect(
      component.nombreCompleto({ ...sinFicha, nombre: "", apellido: "" }),
    ).toBe("p2");
  });
});
