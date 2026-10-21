import { ComponentFixture, TestBed, fakeAsync, tick } from "@angular/core/testing";
import { provideRouter } from "@angular/router";
import { of, throwError } from "rxjs";
import { CampanasPublicoComponent } from "./campanas-publico.component";
import { CampaignsService } from "../../services/campaigns.service";
import { Campana } from "../../core/models/campaign.model";

const campana: Campana = {
  id: "c1",
  titulo: "Campaña del árbol",
  texto: "Plantamos 50 árboles.",
  imagen_url: "https://example.org/arbol.jpg",
  fecha: "2026-09-21",
  estado: "publicada",
};

describe("CampanasPublicoComponent", () => {
  let fixture: ComponentFixture<CampanasPublicoComponent>;
  let service: jasmine.SpyObj<CampaignsService>;

  beforeEach(async () => {
    service = jasmine.createSpyObj<CampaignsService>("CampaignsService", ["publicas"]);
    await TestBed.configureTestingModule({
      imports: [CampanasPublicoComponent],
      providers: [provideRouter([]), { provide: CampaignsService, useValue: service }],
    }).compileComponents();
  });

  function crear() {
    fixture = TestBed.createComponent(CampanasPublicoComponent);
    fixture.detectChanges();
  }

  it("muestra las campañas publicadas con su fecha en DD/MM/AAAA", fakeAsync(() => {
    service.publicas.and.returnValue(of([campana]));
    crear();
    tick(300);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(service.publicas).toHaveBeenCalledWith("");
    expect(el.querySelectorAll(".tarjeta").length).toBe(1);
    expect(el.textContent).toContain("Campaña del árbol");
    expect(el.textContent).toContain("21/09/2026");
  }));

  it("busca por título esperando a que se deje de escribir", fakeAsync(() => {
    service.publicas.and.returnValue(of([]));
    crear();
    tick(300);
    const component = fixture.componentInstance;
    component.buscar("ar");
    component.buscar("arb");
    component.buscar(" arbol ");
    tick(300);
    expect(service.publicas).toHaveBeenCalledTimes(2);
    expect(service.publicas.calls.mostRecent().args[0]).toBe("arbol");
  }));

  it("avisa cuando no hay coincidencias", fakeAsync(() => {
    service.publicas.and.returnValue(of([]));
    crear();
    tick(300);
    fixture.componentInstance.busqueda = "zzz";
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain("No hay campañas que coincidan con «zzz»");
  }));

  it("muestra el error sin romper la página", fakeAsync(() => {
    service.publicas.and.returnValue(throwError(() => new Error("No se pudieron cargar las campañas.")));
    crear();
    tick(300);
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector("[role=alert]")?.textContent).toContain("No se pudieron cargar las campañas.");
  }));
});
