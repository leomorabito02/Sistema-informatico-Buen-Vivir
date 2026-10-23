import { ComponentFixture, TestBed } from "@angular/core/testing";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { of, throwError } from "rxjs";
import { CampanasComponent } from "./campanas.component";
import { CampaignsService } from "../../../services/campaigns.service";
import { Campana } from "../../../core/models/campaign.model";

const borrador: Campana = {
  id: "c1",
  titulo: "Jornada de reciclaje",
  texto: "Traé tus botellas.",
  imagen_url: "https://example.org/a.jpg",
  fecha: "2026-11-05",
  estado: "borrador",
};
const publicada: Campana = { ...borrador, id: "c2", titulo: "Campaña del árbol", estado: "publicada" };

describe("CampanasComponent", () => {
  let fixture: ComponentFixture<CampanasComponent>;
  let component: CampanasComponent;
  let service: jasmine.SpyObj<CampaignsService>;

  function crear(lista: Campana[] = [borrador, publicada]) {
    service.listar.and.returnValue(of(lista));
    fixture = TestBed.createComponent(CampanasComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  beforeEach(async () => {
    service = jasmine.createSpyObj<CampaignsService>("CampaignsService", ["listar", "crear", "editar", "publicar"]);
    await TestBed.configureTestingModule({
      imports: [CampanasComponent, BrowserAnimationsModule],
      providers: [{ provide: CampaignsService, useValue: service }],
    }).compileComponents();
  });

  it("muestra las campañas con su estado y solo ofrece publicar los borradores", () => {
    crear();
    const el: HTMLElement = fixture.nativeElement;
    const tarjetas = el.querySelectorAll(".campana-card");
    expect(tarjetas.length).toBe(2);
    expect(tarjetas[0].textContent).toContain("Borrador");
    expect(tarjetas[0].textContent).toContain("05/11/2026");
    expect(tarjetas[0].textContent).toContain("Publicar");
    expect(tarjetas[1].textContent).toContain("Publicada");
    expect(tarjetas[1].textContent).not.toContain("Publicar");
  });

  it("muestra el estado vacío y el error del servicio", () => {
    crear([]);
    expect(fixture.nativeElement.textContent).toContain("Todavía no se cargó ninguna campaña");

    service.listar.and.returnValue(throwError(() => new Error("Firestore caído")));
    component.cargar();
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain("Firestore caído");
  });

  it("filtra con la búsqueda y el estado elegidos", () => {
    crear();
    component.busqueda = "árbol";
    component.estado = "publicada";
    component.cargar();
    expect(service.listar).toHaveBeenCalledWith("árbol", "publicada");
  });

  it("no deja guardar hasta completar los cuatro datos", () => {
    crear();
    component.nueva();
    expect(component.formCompleto).toBeFalse();
    component.guardar();
    expect(service.crear).not.toHaveBeenCalled();
    component.form = { titulo: "Nueva", texto: "Texto", imagen_url: "https://example.org/x.jpg", fecha: "2026-12-01" };
    expect(component.formCompleto).toBeTrue();
  });

  it("crea, cierra el diálogo, avisa con el mensaje del backend y recarga", () => {
    crear();
    service.crear.and.returnValue(of({ mensaje: "Campaña creada correctamente", campana: borrador }));
    const toast = spyOn((component as unknown as { toast: { add: (m: unknown) => void } }).toast, "add");
    component.nueva();
    component.form = { titulo: "Nueva", texto: "Texto", imagen_url: "https://example.org/x.jpg", fecha: "2026-12-01" };
    component.guardar();
    expect(service.crear).toHaveBeenCalledWith(component.form);
    expect(component.dialogo).toBeFalse();
    expect(toast).toHaveBeenCalledWith(jasmine.objectContaining({ detail: "Campaña creada correctamente" }));
    expect(service.listar).toHaveBeenCalledTimes(2);
  });

  it("edita con los datos de la campaña elegida", () => {
    crear();
    service.editar.and.returnValue(of({ mensaje: "Campaña editada correctamente", campana: borrador }));
    component.editar(borrador);
    expect(component.form.titulo).toBe(borrador.titulo);
    component.guardar();
    expect(service.editar).toHaveBeenCalledWith("c1", jasmine.objectContaining({ titulo: borrador.titulo }));
  });

  it("si el backend rechaza, deja el diálogo abierto y muestra el error", () => {
    crear();
    service.crear.and.returnValue(throwError(() => new Error("La imagen debe ser una dirección web válida (http o https)")));
    const toast = spyOn((component as unknown as { toast: { add: (m: unknown) => void } }).toast, "add");
    component.nueva();
    component.form = { titulo: "N", texto: "T", imagen_url: "ftp://x", fecha: "2026-12-01" };
    component.guardar();
    expect(component.dialogo).toBeTrue();
    expect(toast).toHaveBeenCalledWith(jasmine.objectContaining({ severity: "error" }));
  });

  it("publica y recarga", () => {
    crear();
    service.publicar.and.returnValue(of({ mensaje: "Campaña publicada correctamente", campana: { ...borrador, estado: "publicada" } }));
    component.publicar(borrador);
    expect(service.publicar).toHaveBeenCalledWith("c1");
    expect(service.listar).toHaveBeenCalledTimes(2);
  });
});
