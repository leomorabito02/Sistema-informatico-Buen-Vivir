import { rutaInicio } from "./ruta-inicio";

describe("rutaInicio", () => {
  afterEach(() => localStorage.removeItem("isAuthenticated"));

  it("con sesión vuelve al panel", () => {
    localStorage.setItem("isAuthenticated", "true");
    expect(rutaInicio()).toBe("/dashboard");
  });

  it("sin sesión vuelve a la landing", () => {
    localStorage.removeItem("isAuthenticated");
    expect(rutaInicio()).toBe("/");
  });
});
