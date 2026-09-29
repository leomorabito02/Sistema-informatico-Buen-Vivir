import { ComponentFixture, TestBed } from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { MessageService } from "primeng/api";
import { of, throwError } from "rxjs";

import { AsistenciaGeneralComponent } from "./asistencia-general.component";
import { EducationService } from "../../../services/education.service";
import { GeneralAttendance } from "../../../core/models/education.model";
import { todayYMD } from "../../../shared/utils/date-ymd";

describe("AsistenciaGeneralComponent", () => {
  let fixture: ComponentFixture<AsistenciaGeneralComponent>;
  let component: AsistenciaGeneralComponent;
  let education: jasmine.SpyObj<EducationService>;
  let toast: MessageService;

  const list: GeneralAttendance[] = [
    {
      participanteId: "p2",
      nombre: "Zoe",
      apellido: "Ríos",
      presente: true,
      observacion: "tarde",
    },
    { participanteId: "p1", nombre: "Ana", apellido: "Paz", presente: null },
    { participanteId: "p3", presente: false },
  ];

  beforeEach(() => {
    education = jasmine.createSpyObj<EducationService>("EducationService", [
      "getGeneralAttendance",
      "registerGeneralAttendance",
    ]);
    education.getGeneralAttendance.and.returnValue(of(list));

    TestBed.configureTestingModule({
      imports: [AsistenciaGeneralComponent, NoopAnimationsModule],
      providers: [{ provide: EducationService, useValue: education }],
    });
    fixture = TestBed.createComponent(AsistenciaGeneralComponent);
    component = fixture.componentInstance;
    toast = fixture.debugElement.injector.get(MessageService);
    spyOn(toast, "add");
    fixture.detectChanges();
  });

  it("loads today's attendance sorted by name", () => {
    expect(education.getGeneralAttendance).toHaveBeenCalledWith(todayYMD());
    expect(component.loading).toBeFalse();
    expect(component.rows.map((r) => r.nombreCompleto)).toEqual([
      "Ana Paz",
      "p3",
      "Zoe Ríos",
    ]);
    expect(component.rows[2].observacion).toBe("tarde");
    expect(component.pendientesDeGuardar).toBe(2);
  });

  it("reloads when the date changes", () => {
    component.selectedDate = new Date(2026, 8, 1);
    component.onDateChange();
    expect(component.fecha).toBe("2026-09-01");
    expect(education.getGeneralAttendance).toHaveBeenCalledWith("2026-09-01");
  });

  it("shows the error when loading fails", () => {
    education.getGeneralAttendance.and.returnValue(
      throwError(() => new Error("auth caído")),
    );
    component.cargar();
    expect(component.loading).toBeFalse();
    expect(component.error).toBe("auth caído");
  });

  it("only sends rows marked present or absent", () => {
    education.registerGeneralAttendance.and.returnValue(
      of({ fecha: todayYMD(), status: "ok" }),
    );
    const ana = component.rows[0];
    component.marcar(ana, true);
    ana.observacion = "  ";

    component.guardar();

    expect(education.registerGeneralAttendance).toHaveBeenCalledWith(
      todayYMD(),
      [
        { participanteId: "p1", presente: true, observacion: undefined },
        { participanteId: "p3", presente: false, observacion: undefined },
        { participanteId: "p2", presente: true, observacion: "tarde" },
      ],
    );
    expect(toast.add).toHaveBeenCalledWith(
      jasmine.objectContaining({ severity: "success" }),
    );
    expect(component.guardando).toBeFalse();
  });

  it("does not call the backend when nothing is marked", () => {
    component.rows.forEach((r) => (r.presente = null));
    component.guardar();
    expect(education.registerGeneralAttendance).not.toHaveBeenCalled();
  });

  it("shows a toast when saving fails", () => {
    education.registerGeneralAttendance.and.returnValue(
      throwError(() => new Error("El participante p3 no está activo")),
    );
    component.guardar();
    expect(component.guardando).toBeFalse();
    expect(toast.add).toHaveBeenCalledWith(
      jasmine.objectContaining({
        severity: "error",
        detail: "El participante p3 no está activo",
      }),
    );
  });
});
