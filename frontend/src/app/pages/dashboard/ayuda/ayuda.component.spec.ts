import { ComponentFixture, TestBed } from "@angular/core/testing";
import { of, throwError } from "rxjs";
import { AyudaComponent } from "./ayuda.component";
import { AYUDA_GENERAL, AYUDA_POR_ROL } from "./ayuda.content";
import { AuthService } from "../../../services/auth.service";
import { MeResponse } from "../../../core/models/auth.model";

describe("AyudaComponent", () => {
  let fixture: ComponentFixture<AyudaComponent>;
  let component: AyudaComponent;
  let auth: jasmine.SpyObj<AuthService>;

  function create(me: unknown) {
    auth.isAuthenticated.and.returnValue(of(me as MeResponse));
    fixture = TestBed.createComponent(AyudaComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  beforeEach(async () => {
    auth = jasmine.createSpyObj("AuthService", ["isAuthenticated"]);
    await TestBed.configureTestingModule({
      imports: [AyudaComponent],
      providers: [{ provide: AuthService, useValue: auth }],
    }).compileComponents();
  });

  it("muestra las tareas del rol y las preguntas generales", () => {
    create({ role_id: "role_participant" });
    expect(component.delRol).toBe(AYUDA_POR_ROL["role_participant"]);

    const el: HTMLElement = fixture.nativeElement;
    const preguntas = el.querySelectorAll("details.ayuda-item").length;
    expect(preguntas).toBe(
      AYUDA_POR_ROL["role_participant"].length + AYUDA_GENERAL.length,
    );
    expect(el.textContent).toContain("¿Cómo entrego una actividad?");
    expect(el.querySelectorAll("kbd").length).toBeGreaterThan(0);
  });

  it("todos los roles del menú tienen ayuda propia", () => {
    for (const rol of [
      "role_admin",
      "role_dev",
      "role_volunteer",
      "role_participant",
      "role_teacher",
      "role_company",
    ]) {
      expect(AYUDA_POR_ROL[rol]?.length).withContext(rol).toBeGreaterThan(0);
    }
  });

  it("busca sin importar tildes ni mayúsculas", () => {
    create({ user: { role_id: "role_admin" } });
    component.query = "INSCRIPCION";
    const encontrados = component.filtrar(component.delRol);
    expect(encontrados.length).toBeGreaterThan(0);
    expect(
      encontrados.every((t) =>
        [t.pregunta, ...t.pasos].join(" ").toLowerCase().includes("inscrip"),
      ),
    ).toBeTrue();
  });

  it("avisa cuando la búsqueda no encuentra nada", () => {
    create({ role_id: "role_company" });
    component.query = "zzzz";
    fixture.detectChanges();
    expect(component.sinResultados).toBeTrue();
    expect(fixture.nativeElement.textContent).toContain("Sin resultados");
  });

  it("sin sesión muestra solo la ayuda general", () => {
    auth.isAuthenticated.and.returnValue(throwError(() => new Error("401")));
    fixture = TestBed.createComponent(AyudaComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
    expect(component.delRol).toEqual([]);
    expect(component.filtrar(component.general).length).toBe(
      AYUDA_GENERAL.length,
    );
  });
});
