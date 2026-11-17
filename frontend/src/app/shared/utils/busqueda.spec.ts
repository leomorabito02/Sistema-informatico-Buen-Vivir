import { coincideBusqueda, normalizarBusqueda } from "./busqueda";

describe("busqueda", () => {
  it("normaliza acentos, mayúsculas y espacios", () => {
    expect(normalizarBusqueda("  José   PÉREZ ")).toBe("jose perez");
    expect(normalizarBusqueda(null)).toBe("");
  });

  it("coincide cuando todas las palabras aparecen en algún campo", () => {
    expect(coincideBusqueda("", "Ana")).toBeTrue();
    expect(coincideBusqueda("perez jose", "José", "Pérez")).toBeTrue();
    expect(coincideBusqueda("ana", "Mariana", "Gómez")).toBeTrue();
    expect(coincideBusqueda("luis", "José", "Pérez")).toBeFalse();
  });
});
