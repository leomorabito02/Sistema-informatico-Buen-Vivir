import { ComponentFixture, TestBed } from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { of, throwError } from "rxjs";
import { MisContenidosComponent } from "./mis-contenidos.component";
import {
  ContenidoDocente,
  ContenidosService,
} from "../../../services/contenidos.service";

describe("MisContenidosComponent", () => {
  let fixture: ComponentFixture<MisContenidosComponent>;
  let component: MisContenidosComponent;
  let service: jasmine.SpyObj<ContenidosService>;

  const base: ContenidoDocente = {
    id: "c1",
    docenteId: "usr-doc",
    titulo: "Guía de compost",
    descripcion: "",
    archivoNombre: "compost.pdf",
    archivoTipo: "application/pdf",
    archivoTamano: 2048,
    estadoModeracion: "pendiente",
    fechaSubida: "2026-10-22T12:00:00Z",
  };

  function setup(lista: ContenidoDocente[]) {
    service = jasmine.createSpyObj("ContenidosService", [
      "misContenidos",
      "biblioteca",
      "subir",
      "editar",
      "eliminar",
      "archivo",
    ]);
    service.misContenidos.and.returnValue(of(lista));
    service.biblioteca.and.returnValue(of([]));
    TestBed.configureTestingModule({
      imports: [MisContenidosComponent, NoopAnimationsModule],
      providers: [{ provide: ContenidosService, useValue: service }],
    });
    fixture = TestBed.createComponent(MisContenidosComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  const texto = () => fixture.nativeElement.textContent as string;

  it("shows the state and allows editing only while pending", () => {
    setup([
      base,
      { ...base, id: "c2", titulo: "Rechazado", estadoModeracion: "rechazado", motivoRechazo: "Falta la fuente" },
      { ...base, id: "c3", titulo: "Aprobado", estadoModeracion: "aprobado" },
    ]);
    expect(texto()).toContain("Pendiente de moderación");
    expect(texto()).toContain("Motivo del rechazo: Falta la fuente");
    const editar = fixture.nativeElement.querySelectorAll("p-button[label='Editar']");
    expect(editar.length).toBe(1);
  });

  it("requires a file to upload a new material", () => {
    setup([]);
    expect(texto()).toContain("Todavía no subiste material");
    component.nuevo();
    component.titulo = "Guía";
    component.guardar();
    expect(service.subir).not.toHaveBeenCalled();
    expect(component.errorForm).toContain("Adjuntá");
  });

  it("rejects files over 10 MB before sending", () => {
    setup([]);
    const grande = new File([new Uint8Array(10 * 1024 * 1024 + 1)], "g.pdf");
    const input = { files: [grande], value: "x" } as unknown as HTMLInputElement;
    component.elegirArchivo({ target: input } as unknown as Event);
    expect(component.archivo).toBeNull();
    expect(component.errorForm).toContain("10 MB");
  });

  it("uploads and reloads the list", () => {
    setup([]);
    service.subir.and.returnValue(of(base));
    component.nuevo();
    component.titulo = "Guía de compost";
    component.archivo = new File(["%PDF"], "compost.pdf");
    component.guardar();
    expect(service.subir).toHaveBeenCalledWith(
      jasmine.objectContaining({ titulo: "Guía de compost" }),
    );
    expect(component.dialogo).toBeFalse();
    expect(service.misContenidos).toHaveBeenCalledTimes(2);
  });

  it("edits keeping the file when none is chosen", () => {
    setup([base]);
    service.editar.and.returnValue(of(base));
    component.editar(base);
    component.titulo = "Guía nueva";
    component.guardar();
    expect(service.editar).toHaveBeenCalledWith("c1", jasmine.objectContaining({ titulo: "Guía nueva", archivo: null }));
  });

  it("shows the backend error in the form", () => {
    setup([]);
    service.subir.and.returnValue(throwError(() => new Error("El archivo tiene que ser PDF")));
    component.nuevo();
    component.titulo = "Guía";
    component.archivo = new File(["x"], "x.exe");
    component.guardar();
    expect(component.errorForm).toBe("El archivo tiene que ser PDF");
    expect(component.dialogo).toBeTrue();
  });
});
