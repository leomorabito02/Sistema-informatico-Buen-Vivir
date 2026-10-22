import { provideHttpClientTesting } from "@angular/common/http/testing";
import { provideHttpClient } from "@angular/common/http";
import { provideRouter } from "@angular/router";
import { ComponentFixture, TestBed } from "@angular/core/testing";

import { of } from "rxjs";
import { LibraryComponent } from "./library.component";
import { ContenidosService } from "../../../services/contenidos.service";

describe("LibraryComponent", () => {
  let component: LibraryComponent;
  let fixture: ComponentFixture<LibraryComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
      imports: [LibraryComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(LibraryComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("validates the visitor data before downloading (HU-25)", () => {
    const service = TestBed.inject(ContenidosService);
    const spy = spyOn(service, "descargarPublico").and.returnValue(of(new Blob(["x"])));
    const recurso = {
      id: "c1",
      titulo: "Guía",
      descripcion: "",
      archivoNombre: "guia.pdf",
      archivoTipo: "application/pdf",
      archivoTamano: 1024,
      fechaSubida: "",
    };
    component.abrir(recurso);

    component.datos = { nombre: "Ana", apellido: "Gómez", email: "ana", telefono: "3547123456" };
    component.descargar();
    expect(component.errorForm).toContain("email");

    component.datos.email = "ana@gmail.com";
    component.datos.telefono = "1234";
    component.descargar();
    expect(component.errorForm).toContain("teléfono");
    expect(spy).not.toHaveBeenCalled();

    component.datos.telefono = "(3547) 12-3456";
    spyOn(URL, "createObjectURL").and.returnValue("blob:x");
    spyOn(URL, "revokeObjectURL");
    spyOn(HTMLAnchorElement.prototype, "click");
    component.descargar();
    expect(spy).toHaveBeenCalledWith("c1", component.datos);
    expect(component.elegido).toBeNull();
  });

  it("shows the file kind", () => {
    const r = { archivoTipo: "image/png" } as any;
    expect(component.formato(r)).toBe("Imagen");
    expect(component.formato({ archivoTipo: "application/pdf" } as any)).toBe("PDF");
  });
});
