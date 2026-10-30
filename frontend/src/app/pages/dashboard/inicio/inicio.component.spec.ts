import { ComponentFixture, TestBed } from "@angular/core/testing";
import { Router, provideRouter } from "@angular/router";
import { of, throwError } from "rxjs";
import { InicioComponent } from "./inicio.component";
import { AuthService } from "../../../services/auth.service";

describe("InicioComponent", () => {
  let fixture: ComponentFixture<InicioComponent>;
  let component: InicioComponent;
  let auth: jasmine.SpyObj<AuthService>;
  let router: Router;

  function create(me: unknown | Error) {
    auth.isAuthenticated.and.returnValue(
      me instanceof Error ? throwError(() => me) : of(me as any),
    );
    fixture = TestBed.createComponent(InicioComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  beforeEach(async () => {
    auth = jasmine.createSpyObj("AuthService", ["isAuthenticated"]);
    await TestBed.configureTestingModule({
      imports: [InicioComponent],
      providers: [{ provide: AuthService, useValue: auth }, provideRouter([])],
    }).compileComponents();
    router = TestBed.inject(Router);
    spyOn(router, "navigateByUrl").and.resolveTo(true);
  });

  it("saluda al participante por su nombre", () => {
    create({ role_id: "role_participant", user: { first_name: " Ana " } });

    expect(component.roleLabel).toBe("Participante");
    expect(component.firstName).toBe("Ana");
    expect(router.navigateByUrl).not.toHaveBeenCalled();
    expect(fixture.nativeElement.textContent).toContain("Ana");
  });

  it("toma el rol desde user cuando no viene en la raíz", () => {
    create({ user: { role_id: "role_admin", first_name: "Sofía" } });
    expect(component.roleLabel).toBe("Responsable");
  });

  it("lleva al docente directo a Mis cursos, reemplazando el historial", () => {
    create({ role_id: "role_teacher" });

    expect(router.navigateByUrl).toHaveBeenCalledWith("/dashboard/mis-cursos", {
      replaceUrl: true,
    });
    expect(component.roleLabel).toBe("");
  });

  it("lleva a la empresa directo a Ofertas laborales", () => {
    create({ role_id: "role_company" });

    expect(router.navigateByUrl).toHaveBeenCalledWith("/dashboard/ofertas", {
      replaceUrl: true,
    });
  });

  it("no muestra nada si no se puede leer la sesión", () => {
    create(new Error("401"));

    expect(component.roleLabel).toBe("");
    expect(fixture.nativeElement.querySelector(".welcome")).toBeNull();
  });
});
