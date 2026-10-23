import { ComponentFixture, TestBed } from "@angular/core/testing";
import { of, throwError } from "rxjs";
import { ComentariosComponent } from "./comentarios.component";
import {
  Comentario,
  ComentariosService,
} from "../../services/comentarios.service";

describe("ComentariosComponent", () => {
  let fixture: ComponentFixture<ComentariosComponent>;
  let component: ComentariosComponent;
  let service: jasmine.SpyObj<ComentariosService>;

  const existente: Comentario = {
    id: "com-1",
    entidadTipo: "campana",
    entidadId: "camp-1",
    nombre: "Ana",
    apellido: "Gómez",
    texto: "<b>Muy buena</b>",
    fecha: "2026-10-22T15:00:00Z",
  };

  function setup(lista: Comentario[] = [existente], puedeModerar = false) {
    service = jasmine.createSpyObj("ComentariosService", [
      "listar",
      "crear",
      "eliminar",
    ]);
    service.listar.and.returnValue(of(lista));
    TestBed.configureTestingModule({
      imports: [ComentariosComponent],
      providers: [{ provide: ComentariosService, useValue: service }],
    });
    fixture = TestBed.createComponent(ComentariosComponent);
    component = fixture.componentInstance;
    fixture.componentRef.setInput("entidadTipo", "campana");
    fixture.componentRef.setInput("entidadId", "camp-1");
    fixture.componentRef.setInput("puedeModerar", puedeModerar);
    fixture.detectChanges();
  }

  const texto = () => fixture.nativeElement.textContent as string;

  it("loads the comments of the content and shows the text escaped", () => {
    setup();
    expect(service.listar).toHaveBeenCalledWith("campana", "camp-1");
    expect(texto()).toContain("Ana Gómez");
    expect(fixture.nativeElement.querySelector(".comentario-texto b")).toBeNull();
    expect(texto()).toContain("<b>Muy buena</b>");
  });

  it("shows the empty message", () => {
    setup([]);
    expect(texto()).toContain("Todavía no hay comentarios");
  });

  it("asks for name and surname before sending", () => {
    setup([]);
    component.texto = "Hola";
    component.enviar();
    expect(service.crear).not.toHaveBeenCalled();
    expect(component.errorForm).toContain("nombre y apellido");
  });

  it("publishes the comment and puts it first", () => {
    setup();
    const nuevo = { ...existente, id: "com-2", texto: "Genial" };
    service.crear.and.returnValue(of(nuevo));
    component.nombre = "Luis";
    component.apellido = "Paz";
    component.texto = "Genial";
    component.enviar();

    expect(service.crear).toHaveBeenCalledWith({
      entidadTipo: "campana",
      entidadId: "camp-1",
      nombre: "Luis",
      apellido: "Paz",
      texto: "Genial",
    });
    expect(component.comentarios.map((c) => c.id)).toEqual(["com-2", "com-1"]);
    expect(component.texto).toBe("");
    expect(component.enviado).toBeTrue();
  });

  it("shows the backend error", () => {
    setup();
    service.crear.and.returnValue(
      throwError(() => new Error("Mandaste muchos comentarios seguidos.")),
    );
    component.nombre = "Luis";
    component.apellido = "Paz";
    component.texto = "Hola";
    component.enviar();
    expect(component.errorForm).toContain("muchos comentarios");
    expect(component.enviando).toBeFalse();
  });

  it("only the moderator sees Eliminar, and it removes the comment", () => {
    setup([existente], false);
    expect(fixture.nativeElement.querySelector(".comentario-eliminar")).toBeNull();

    fixture.componentRef.setInput("puedeModerar", true);
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector(".comentario-eliminar")).not.toBeNull();

    spyOn(window, "confirm").and.returnValue(true);
    service.eliminar.and.returnValue(of(undefined));
    component.eliminar(existente);
    expect(service.eliminar).toHaveBeenCalledWith("com-1");
    expect(component.comentarios).toEqual([]);
  });
});
