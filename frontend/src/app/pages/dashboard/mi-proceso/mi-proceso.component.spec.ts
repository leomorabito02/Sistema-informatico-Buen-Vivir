import { ComponentFixture, TestBed } from "@angular/core/testing";
import { of, throwError } from "rxjs";
import { MiProcesoComponent } from "./mi-proceso.component";
import { HumanPromotionService } from "../../../services/human-promotion.service";
import { EducationService } from "../../../services/education.service";
import { LineaDeTiempoPropia } from "../../../core/models/human-promotion.model";
import { Course } from "../../../core/models/education.model";

const linea: LineaDeTiempoPropia = {
  hitos: [
    {
      id: "h1",
      fecha: "2026-09-20",
      descripcion: "Empezó el taller de cocina",
    },
  ],
  actividad: [
    {
      id: "i1",
      tipo: "inscripcion",
      fecha: "2026-09-02",
      descripcion: "Se inscribió en un curso",
      curso_id: "c1",
    },
  ],
};

describe("MiProcesoComponent", () => {
  let fixture: ComponentFixture<MiProcesoComponent>;
  let service: jasmine.SpyObj<HumanPromotionService>;
  let education: jasmine.SpyObj<EducationService>;

  beforeEach(async () => {
    service = jasmine.createSpyObj<HumanPromotionService>(
      "HumanPromotionService",
      ["miLineaDeTiempo"],
    );
    education = jasmine.createSpyObj<EducationService>("EducationService", [
      "listCourses",
    ]);
    education.listCourses.and.returnValue(
      of([{ id: "c1", title: "Huerta agroecológica" } as Course]),
    );
    await TestBed.configureTestingModule({
      imports: [MiProcesoComponent],
      providers: [
        { provide: HumanPromotionService, useValue: service },
        { provide: EducationService, useValue: education },
      ],
    }).compileComponents();
  });

  function crear(): HTMLElement {
    fixture = TestBed.createComponent(MiProcesoComponent);
    fixture.detectChanges();
    return fixture.nativeElement;
  }

  it("mezcla por fecha los hitos y la actividad en el centro", () => {
    service.miLineaDeTiempo.and.returnValue(of(linea));
    const el = crear();
    const items = el.querySelectorAll(".timeline-item");
    expect(items.length).toBe(2);
    // Primero la inscripción (02/09), después el hito (20/09).
    expect(items[0].classList).toContain("timeline-item--actividad");
    expect(items[0].textContent).toContain("02/09/2026");
    expect(items[0].textContent).toContain("Huerta agroecológica");
    expect(items[0].textContent).toContain("Actividad en el centro");
    expect(items[1].textContent).toContain("Empezó el taller de cocina");
    expect(items[1].textContent).toContain("Equipo");
  });

  it("si Educación no responde, muestra la actividad sin el nombre del curso", () => {
    education.listCourses.and.returnValue(throwError(() => new Error("caído")));
    service.miLineaDeTiempo.and.returnValue(of(linea));
    const el = crear();
    expect(el.querySelectorAll(".timeline-item").length).toBe(2);
    expect(el.textContent).not.toContain("Huerta");
    expect(el.querySelector(".timeline-curso")).toBeNull();
  });

  it("avisa cuando todavía no hay nada registrado", () => {
    service.miLineaDeTiempo.and.returnValue(of({ hitos: [], actividad: [] }));
    const el = crear();
    expect(el.textContent).toContain("Todavía no hay momentos registrados");
  });

  it("muestra el error y permite reintentar", () => {
    service.miLineaDeTiempo.and.returnValue(
      throwError(() => new Error("No se pudo conectar")),
    );
    const el = crear();
    expect(el.textContent).toContain("No se pudo conectar");

    service.miLineaDeTiempo.and.returnValue(of(linea));
    fixture.componentInstance.cargar();
    fixture.detectChanges();
    expect(el.querySelectorAll(".timeline-item").length).toBe(2);
  });

  it("no muestra números de asistencia: una entrada por mes", () => {
    service.miLineaDeTiempo.and.returnValue(
      of({
        hitos: [],
        actividad: [
          {
            id: "asistencia-2026-09",
            tipo: "asistencia_general",
            fecha: "2026-09-05",
            descripcion: "Asistió al centro en septiembre de 2026",
          },
        ],
      }),
    );
    const el = crear();
    expect(el.querySelectorAll(".timeline-item").length).toBe(1);
    expect(el.textContent).toContain("Asistió al centro en septiembre de 2026");
  });
});
