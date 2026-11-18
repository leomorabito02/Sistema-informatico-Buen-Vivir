import { of, throwError } from "rxjs";
import { mezclarLineaDeTiempo, nombresDeCursos } from "./linea-tiempo";
import { EntradaActividad } from "../../core/models/human-promotion.model";
import { EducationService } from "../../services/education.service";
import { Course } from "../../core/models/education.model";

const inscripcion: EntradaActividad = {
  id: "i1",
  tipo: "inscripcion",
  fecha: "2026-09-10",
  descripcion: "Se inscribió en un curso",
  curso_id: "c1",
};
const asistencia: EntradaActividad = {
  id: "asistencia-2026-09",
  tipo: "asistencia_general",
  fecha: "2026-09-05",
  descripcion: "Asistió al centro en septiembre de 2026",
};

describe("mezclarLineaDeTiempo", () => {
  it("ordena hitos y actividad por fecha", () => {
    const items = mezclarLineaDeTiempo(
      [
        { id: "h1", fecha: "2026-09-01" },
        { id: "h2", fecha: "2026-09-20" },
      ],
      [inscripcion, asistencia],
    );
    expect(items.map((i) => i.key)).toEqual([
      "h-h1",
      "a-asistencia-2026-09",
      "a-i1",
      "h-h2",
    ]);
  });

  it("en la misma fecha pone primero la actividad", () => {
    const items = mezclarLineaDeTiempo(
      [{ id: "h1", fecha: "2026-09-10" }],
      [inscripcion],
    );
    expect(items.map((i) => i.tipo)).toEqual(["actividad", "hito"]);
  });

  it("agrega el nombre del curso cuando se conoce", () => {
    const [conNombre] = mezclarLineaDeTiempo(
      [],
      [inscripcion],
      new Map([["c1", "Huerta"]]),
    );
    expect(conNombre.tipo === "actividad" && conNombre.curso).toBe("Huerta");

    const [sinNombre] = mezclarLineaDeTiempo([], [inscripcion]);
    expect(sinNombre.tipo === "actividad" && sinNombre.curso).toBeUndefined();
  });

  it("sin datos devuelve una lista vacía", () => {
    expect(mezclarLineaDeTiempo([], [])).toEqual([]);
  });
});

describe("nombresDeCursos", () => {
  it("arma id -> título", (done) => {
    const education = jasmine.createSpyObj<EducationService>(
      "EducationService",
      ["listCourses"],
    );
    education.listCourses.and.returnValue(
      of([{ id: "c1", title: "Huerta" } as Course]),
    );
    nombresDeCursos(education).subscribe((m) => {
      expect(m.get("c1")).toBe("Huerta");
      done();
    });
  });

  it("si Educación falla, devuelve un mapa vacío en vez de error", (done) => {
    const education = jasmine.createSpyObj<EducationService>(
      "EducationService",
      ["listCourses"],
    );
    education.listCourses.and.returnValue(throwError(() => new Error("caído")));
    nombresDeCursos(education).subscribe((m) => {
      expect(m.size).toBe(0);
      done();
    });
  });
});
