import {
  ComponentFixture,
  TestBed,
  fakeAsync,
  tick,
} from "@angular/core/testing";
import { provideRouter } from "@angular/router";
import { of, throwError } from "rxjs";
import { CampanasPublicoComponent } from "./campanas-publico.component";
import { CampaignsService } from "../../services/campaigns.service";
import { Campana } from "../../core/models/campaign.model";
import { ComentariosService } from "../../services/comentarios.service";
import { AuthService } from "../../services/auth.service";
import { ProfileService } from "../../services/profile.service";

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
    service = jasmine.createSpyObj<CampaignsService>("CampaignsService", [
      "publicasPagina",
    ]);
    await TestBed.configureTestingModule({
      imports: [CampanasPublicoComponent],
      providers: [
        provideRouter([]),
        { provide: CampaignsService, useValue: service },
        { provide: ComentariosService, useValue: { listar: () => of([]) } },
        {
          provide: AuthService,
          useValue: { isLocallyAuthenticated: () => false },
        },
        {
          provide: ProfileService,
          useValue: { getMyProfile: () => of({ user: {} }) },
        },
      ],
    }).compileComponents();
  });

  function crear() {
    fixture = TestBed.createComponent(CampanasPublicoComponent);
    fixture.detectChanges();
  }

  it("muestra las campañas publicadas con su fecha en DD/MM/AAAA", fakeAsync(() => {
    service.publicasPagina.and.returnValue(
      of({ items: [campana], total: [campana].length }),
    );
    crear();
    tick(300);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(service.publicasPagina).toHaveBeenCalledWith("", 0);
    expect(el.querySelectorAll(".tarjeta")).toHaveSize(1);
    expect(el.textContent).toContain("Campaña del árbol");
    expect(el.textContent).toContain("21/09/2026");
  }));

  it("al seleccionar una campaña abre su detalle con los comentarios (HU-24)", fakeAsync(() => {
    service.publicasPagina.and.returnValue(
      of({ items: [campana], total: [campana].length }),
    );
    crear();
    tick(300);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector(".detalle")).toBeNull();
    expect(el.querySelector("app-comentarios")).toBeNull();

    (el.querySelector(".tarjeta") as HTMLElement).click();
    fixture.detectChanges();
    const detalle = el.querySelector(".detalle") as HTMLElement;
    expect(detalle.textContent).toContain("Campaña del árbol");
    expect(detalle.textContent).toContain("Plantamos 50 árboles.");
    expect(detalle.querySelector("app-comentarios")).not.toBeNull();

    (el.querySelector(".detalle-cerrar") as HTMLButtonElement).click();
    fixture.detectChanges();
    expect(el.querySelector(".detalle")).toBeNull();
  }));

  it("cierra el detalle con Escape", fakeAsync(() => {
    service.publicasPagina.and.returnValue(
      of({ items: [campana], total: [campana].length }),
    );
    crear();
    tick(300);
    const component = fixture.componentInstance;
    component.verDetalle(campana);
    fixture.detectChanges();

    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    fixture.detectChanges();
    expect(component.seleccionada).toBeNull();
  }));

  it("busca por título esperando a que se deje de escribir", fakeAsync(() => {
    service.publicasPagina.and.returnValue(of({ items: [], total: [].length }));
    crear();
    tick(300);
    const component = fixture.componentInstance;
    component.buscar("ar");
    component.buscar("arb");
    component.buscar(" arbol ");
    tick(300);
    expect(service.publicasPagina).toHaveBeenCalledTimes(2);
    expect(service.publicasPagina.calls.mostRecent().args[0]).toBe("arbol");
  }));

  it("avisa cuando no hay coincidencias", fakeAsync(() => {
    service.publicasPagina.and.returnValue(of({ items: [], total: [].length }));
    crear();
    tick(300);
    fixture.componentInstance.busqueda = "zzz";
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain(
      "No hay campañas que coincidan con «zzz»",
    );
  }));

  it("muestra el error sin romper la página", fakeAsync(() => {
    service.publicasPagina.and.returnValue(
      throwError(() => new Error("No se pudieron cargar las campañas.")),
    );
    crear();
    tick(300);
    fixture.detectChanges();
    expect(
      fixture.nativeElement.querySelector("[role=alert]")?.textContent,
    ).toContain("No se pudieron cargar las campañas.");
  }));

  it("reemplaza la imagen por el placeholder ante error de carga", fakeAsync(() => {
    service.publicasPagina.and.returnValue(of({ items: [], total: [].length }));
    crear();
    tick(300);
    const img = document.createElement("img");
    img.src = "https://broken.example.com/fail.jpg";
    fixture.componentInstance.onImageError({ target: img } as unknown as Event);
    expect(img.src).toContain("data:image/svg+xml");
  }));

  it("trae la página siguiente con Ver más (HU-50)", fakeAsync(() => {
    service.publicasPagina.and.returnValues(
      of({ items: [campana], total: 2 }),
      of({ items: [{ ...campana, id: "c2" }], total: 2 }),
    );
    crear();
    tick(300);
    const component = fixture.componentInstance;
    expect(component.hayMas).toBeTrue();

    component.verMas();

    expect(service.publicasPagina.calls.mostRecent().args[1]).toBe(1);
    expect(component.campanas).toHaveSize(2);
    expect(component.hayMas).toBeFalse();
  }));
});
