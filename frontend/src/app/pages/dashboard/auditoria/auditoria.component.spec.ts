import { ComponentFixture, TestBed } from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { of, throwError } from "rxjs";

import { AuditoriaComponent } from "./auditoria.component";
import { EducationService } from "../../../services/education.service";
import { HumanPromotionService } from "../../../services/human-promotion.service";
import { RegistroAuditoria } from "../../../core/models/education.model";

describe("AuditoriaComponent", () => {
  let fixture: ComponentFixture<AuditoriaComponent>;
  let component: AuditoriaComponent;
  let education: jasmine.SpyObj<EducationService>;
  let promocion: jasmine.SpyObj<HumanPromotionService>;

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
    education.listAuditoria.and.returnValue(
      of([{ ...fila, accion: "baja", recurso: "curso", recursoId: "c1" }]),
    );

    TestBed.configureTestingModule({
      imports: [AuditoriaComponent, NoopAnimationsModule],
      providers: [
        { provide: EducationService, useValue: education },
        { provide: HumanPromotionService, useValue: promocion },
      ],
    });
    fixture = TestBed.createComponent(AuditoriaComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("starts with the Promoción Humana log of the last 30 days", () => {
    expect(promocion.listarAuditoria).toHaveBeenCalledWith(
      jasmine.objectContaining({ desde: component.desde, hasta: component.hasta }),
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
});
