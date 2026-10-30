import { ComponentFixture, TestBed } from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { of, throwError } from "rxjs";

import {
  ParticipantPickerComponent,
  etiquetaParticipante,
} from "./participant-picker.component";
import { EducationService } from "../../services/education.service";
import { ParticipanteResumen } from "../../core/models/education.model";

describe("ParticipantPickerComponent", () => {
  let fixture: ComponentFixture<ParticipantPickerComponent>;
  let component: ParticipantPickerComponent;
  let education: jasmine.SpyObj<EducationService>;

  beforeEach(() => {
    education = jasmine.createSpyObj<EducationService>("EducationService", [
      "searchParticipants",
    ]);
    education.searchParticipants.and.returnValue(
      of([{ participanteId: "p1", nombre: "Ana", apellido: "Paz" }]),
    );
    TestBed.configureTestingModule({
      imports: [ParticipantPickerComponent, NoopAnimationsModule],
      providers: [{ provide: EducationService, useValue: education }],
    });
    fixture = TestBed.createComponent(ParticipantPickerComponent);
    component = fixture.componentInstance;
    component.cursoId = "c1";
    fixture.detectChanges();
  });

  it("searches in the backend, excluding those already enrolled in the course", () => {
    component.buscar({ query: "an" });
    expect(education.searchParticipants).toHaveBeenCalledWith("an", "c1");
    expect(component.sugerencias.map((s) => s.etiqueta)).toEqual(["Paz, Ana"]);
  });

  it("shows the backend message when the search fails", () => {
    education.searchParticipants.and.returnValue(
      throwError(() => new Error("Escribí al menos 2 letras")),
    );
    component.buscar({ query: "a" });
    expect(component.sugerencias).toEqual([]);
    expect(component.mensajeVacio).toBe("Escribí al menos 2 letras");
  });

  it("emits the chosen participant and null when cleared", () => {
    const emitidos: (ParticipanteResumen | null)[] = [];
    component.seleccionado.subscribe((p) => emitidos.push(p));
    const ana = {
      participanteId: "p1",
      nombre: "Ana",
      apellido: "Paz",
      etiqueta: "Paz, Ana",
    };
    component.elegir(ana);
    component.limpiar();
    expect(emitidos).toEqual([ana, null]);
    expect(component.valor).toBeNull();
  });

  it("builds the label with what is available", () => {
    expect(
      etiquetaParticipante({
        participanteId: "p1",
        nombre: "Ana",
        apellido: "Paz",
      }),
    ).toBe("Paz, Ana");
    expect(
      etiquetaParticipante({
        participanteId: "p1",
        nombre: "Ana",
        apellido: "",
      }),
    ).toBe("Ana");
    expect(etiquetaParticipante({ participanteId: "p9" })).toBe("p9");
  });
});
