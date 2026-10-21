import {
  ComponentFixture,
  TestBed,
  fakeAsync,
  tick,
} from "@angular/core/testing";
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

  beforeEach(async () => {
    service = jasmine.createSpyObj<HumanPromotionService>(
      "HumanPromotionService",
      ["buscarParticipantes"],
    );
    await TestBed.configureTestingModule({
      imports: [PromocionHumanaComponent],
      providers: [
        provideRouter([]),
        { provide: HumanPromotionService, useValue: service },
      ],
    }).compileComponents();
    fixture = TestBed.createComponent(PromocionHumanaComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("espera a que se deje de escribir y busca una sola vez (HU-36)", fakeAsync(() => {
    service.buscarParticipantes.and.returnValue(of([conFicha]));
    component.onTextoChange("a");
    component.onTextoChange("an");
    component.onTextoChange(" ana ");
    tick(400);
    expect(service.buscarParticipantes).toHaveBeenCalledOnceWith("ana");
  }));

  it("con menos de 2 letras no consulta y limpia los resultados", fakeAsync(() => {
    service.buscarParticipantes.and.returnValue(of([conFicha]));
    component.onTextoChange("ana");
    tick(400);
    component.onTextoChange("a");
    tick(400);
    expect(service.buscarParticipantes).toHaveBeenCalledTimes(1);
    expect(component.resultados).toEqual([]);
    expect(component.buscado).toBeFalse();
  }));

  it("indica quién tiene ficha y ofrece verla o crearla", fakeAsync(() => {
    service.buscarParticipantes.and.returnValue(of([conFicha, sinFicha]));
    component.onTextoChange("pe");
    tick(400);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    const filas = el.querySelectorAll(".result-item");
    expect(filas.length).toBe(2);
    expect(filas[0].textContent).toContain("Ana Pérez");
    expect(filas[0].textContent).toContain("Con ficha");
    expect(filas[0].querySelector("a")?.getAttribute("aria-label")).toBe(
      "Ver ficha de Ana Pérez",
    );
    expect(filas[1].textContent).toContain("Sin ficha");
    expect(filas[1].querySelector("a")?.getAttribute("aria-label")).toBe(
      "Crear ficha para Bruno Gómez",
    );
  }));

  it("avisa cuando no hay coincidencias", fakeAsync(() => {
    service.buscarParticipantes.and.returnValue(of([]));
    component.onTextoChange("zzz");
    tick(400);
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain(
      "No se encontraron participantes",
    );
  }));

  it("muestra el error del servicio", fakeAsync(() => {
    service.buscarParticipantes.and.returnValue(
      throwError(() => new Error("No tenés permisos")),
    );
    component.onTextoChange("ana");
    tick(400);
    fixture.detectChanges();
    expect(component.error).toBe("No tenés permisos");
    expect(fixture.nativeElement.textContent).toContain("No tenés permisos");
  }));

  it("usa el id si el participante no tiene nombre cargado", () => {
    expect(
      component.nombreCompleto({ ...sinFicha, nombre: "", apellido: "" }),
    ).toBe("p2");
  });
});
