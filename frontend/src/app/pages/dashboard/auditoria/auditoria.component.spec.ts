import { ComponentFixture, TestBed } from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { of, throwError } from "rxjs";

import { AuditoriaComponent } from "./auditoria.component";
import { EducationService } from "../../../services/education.service";
import { HumanPromotionService } from "../../../services/human-promotion.service";
import { RegistroAuditoria } from "../../../core/models/education.model";
import {
  AuditoriaService,
  aRegistro,
} from "../../../services/auditoria.service";

describe("AuditoriaComponent", () => {
  let fixture: ComponentFixture<AuditoriaComponent>;
  let component: AuditoriaComponent;
  let education: jasmine.SpyObj<EducationService>;
  let promocion: jasmine.SpyObj<HumanPromotionService>;
  let auditoria: jasmine.SpyObj<AuditoriaService>;

  const fila: RegistroAuditoria = {
    id: "1",
    usuarioId: "u1",
    usuarioNombre: "Ana Paz",
    accion: "actualizar",
    recurso: "ficha",
    recursoId: "f1",
    fecha: "2026-10-20T15:00:00Z",
  };

  beforeEach(() => {
    education = jasmine.createSpyObj<EducationService>("EducationService", [
      "listAuditoria",
    ]);
    promocion = jasmine.createSpyObj<HumanPromotionService>(
      "HumanPromotionService",
      ["listarAuditoria"],
    );
    promocion.listarAuditoria.and.returnValue(of([fila]));
    auditoria = jasmine.createSpyObj<AuditoriaService>("AuditoriaService", [
      "listarUsuarios",
      "listarLaboral",
    ]);
    auditoria.listarUsuarios.and.returnValue(
      of([
        {
          ...fila,
          accion: "acceso_denegado",
          recurso: "perfil_participante",
          detalle: "Denegado: PUT /api/v1/participants/{id}",
        },
      ]),
    );
    auditoria.listarLaboral.and.returnValue(
      of([
        { ...fila, accion: "aprobar", recurso: "oferta", recursoId: "of-1" },
      ]),
    );
    education.listAuditoria.and.returnValue(
      of([{ ...fila, accion: "baja", recurso: "curso", recursoId: "c1" }]),
    );

    TestBed.configureTestingModule({
      imports: [AuditoriaComponent, NoopAnimationsModule],
      providers: [
        { provide: EducationService, useValue: education },
        { provide: HumanPromotionService, useValue: promocion },
        { provide: AuditoriaService, useValue: auditoria },
      ],
    });
    fixture = TestBed.createComponent(AuditoriaComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("starts with the Promoción Humana log of the last 30 days", () => {
    expect(promocion.listarAuditoria).toHaveBeenCalledWith(
      jasmine.objectContaining({
        desde: component.desde,
        hasta: component.hasta,
      }),
    );
    expect(component.registros).toEqual([fila]);
    const text = fixture.nativeElement.textContent;
    expect(text).toContain("Ana Paz");
    expect(text).toContain("Edición");
    expect(text).toContain("Ficha personal");
  });

  it("switches to the Educación log", () => {
    component.modulo = "educacion";
    component.accion = "baja";
    component.buscar();
    fixture.detectChanges();

    expect(education.listAuditoria).toHaveBeenCalledWith(
      jasmine.objectContaining({ accion: "baja" }),
    );
    expect(fixture.nativeElement.textContent).toContain("Curso");
  });

  it("shows the backend error", () => {
    promocion.listarAuditoria.and.returnValue(
      throwError(() => new Error("La fecha desde no puede ser posterior")),
    );
    component.buscar();
    expect(component.loading).toBeFalse();
    expect(component.registros).toEqual([]);
    expect(component.error).toContain("posterior");
  });

  it("labels unknown actions and resources with their raw value", () => {
    expect(component.accionLabel("otra")).toBe("otra");
    expect(component.recursoLabel("otro")).toBe("otro");
    expect(component.accionSeverity("crear")).toBe("success");
    expect(component.accionSeverity("actualizar")).toBe("info");
    expect(component.accionSeverity("baja")).toBe("danger");
    expect(component.accionSeverity("x")).toBe("secondary");
  });

  it("switches to the Autenticación log (HU-10, HU-06)", () => {
    component.modulo = "usuarios";
    component.buscar();
    fixture.detectChanges();

    expect(auditoria.listarUsuarios).toHaveBeenCalled();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain("Acceso denegado");
    expect(text).toContain("Perfil de participante");
    expect(text).toContain("Denegado: PUT");
  });

  it("switches to the Inserción laboral log (HU-42)", () => {
    component.modulo = "laboral";
    component.buscar();
    fixture.detectChanges();

    expect(auditoria.listarLaboral).toHaveBeenCalled();
    const text = fixture.nativeElement.textContent;
    expect(text).toContain("Aprobación");
    expect(text).toContain("Oferta laboral");
  });
});

describe("aRegistro", () => {
  it("maps the snake_case API row and the changed fields", () => {
    const r = aRegistro({
      id: 7,
      usuario_id: "usr-1",
      accion: "actualizar",
      recurso: "perfil_participante",
      recurso_id: "part-1",
      campos: "address,dni",
      resultado: "ok",
      fecha: "2026-11-15T10:00:00Z",
    });
    expect(r.id).toBe("7");
    expect(r.recursoId).toBe("part-1");
    expect(r.detalle).toBe("Campos: address, dni");
  });
});
