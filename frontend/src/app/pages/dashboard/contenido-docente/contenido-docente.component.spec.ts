import { ComponentFixture, TestBed } from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { of } from "rxjs";
import { ContenidoDocenteComponent } from "./contenido-docente.component";
import {
  ContenidoDocente,
  ContenidosService,
} from "../../../services/contenidos.service";
import { UsersService } from "../../../services/users.service";

describe("ContenidoDocenteComponent", () => {
  let fixture: ComponentFixture<ContenidoDocenteComponent>;
  let component: ContenidoDocenteComponent;
  let service: jasmine.SpyObj<ContenidosService>;

  const pendiente: ContenidoDocente = {
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

  beforeEach(() => {
    service = jasmine.createSpyObj("ContenidosService", [
      "listar",
      "moderar",
      "descargas",
      "archivo",
    ]);
    service.listar.and.returnValue(of([pendiente]));
    service.descargas.and.returnValue(
      of([
        {
          id: "d1",
          contenidoId: "c1",
          nombre: "Ana",
          apellido: "Gómez",
          email: "ana@gmail.com",
          telefono: "3547123456",
          fecha: "2026-10-22T13:00:00Z",
        },
      ]),
    );
    service.moderar.and.returnValue(of(pendiente));
    const users = jasmine.createSpyObj("UsersService", ["listByRole"]);
    users.listByRole.and.returnValue(
      of([{ user_id: "usr-doc", first_name: "Luis", last_name: "Paz" }]),
    );

    TestBed.configureTestingModule({
      imports: [ContenidoDocenteComponent, NoopAnimationsModule],
      providers: [
        { provide: ContenidosService, useValue: service },
        { provide: UsersService, useValue: users },
      ],
    });
    fixture = TestBed.createComponent(ContenidoDocenteComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  const texto = () => fixture.nativeElement.textContent as string;

  it("starts with the pending material, with the teacher's name", () => {
    expect(service.listar).toHaveBeenCalledWith("pendiente");
    expect(texto()).toContain("Guía de compost");
    expect(texto()).toContain("Luis Paz");
  });

  it("lists who downloaded from the library, with the material title", () => {
    expect(texto()).toContain("ana@gmail.com");
    expect(texto()).toContain("3547123456");
    expect(component.titulos.get("c1")).toBe("Guía de compost");
  });

  it("approves", () => {
    component.aprobar(pendiente);
    expect(service.moderar).toHaveBeenCalledWith("c1", "aprobado");
  });

  it("rejects with a reason", () => {
    component.abrirRechazo(pendiente);
    component.motivo = "Falta la fuente";
    component.rechazar();
    expect(service.moderar).toHaveBeenCalledWith(
      "c1",
      "rechazado",
      "Falta la fuente",
    );
    expect(component.rechazando).toBeNull();
  });
});
