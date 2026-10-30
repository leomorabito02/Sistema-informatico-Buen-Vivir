import { TestBed } from "@angular/core/testing";
import { provideHttpClient } from "@angular/common/http";
import {
  HttpTestingController,
  provideHttpClientTesting,
} from "@angular/common/http/testing";

import { SkillsService, SUGGESTIONS_UNAVAILABLE } from "./skills.service";
import { environment } from "../../environments/environment";
import { skillKey } from "../core/models/skills.model";

describe("SkillsService", () => {
  let service: SkillsService;
  let http: HttpTestingController;
  // Se calcula en cada test: otro spec puede cambiar environment.apiUrl.
  const base = () => environment.educationApiUrl || environment.apiUrl || "";

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(SkillsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it("busca en el catálogo con credenciales", () => {
    service.searchCatalog("equipo").subscribe();
    const req = http.expectOne(`${base()}/api/habilidades?q=equipo`);
    expect(req.request.withCredentials).toBeTrue();
    req.flush([]);
  });

  it("trae las habilidades del curso", () => {
    let got: unknown;
    service.getCourseSkills("c1").subscribe((r) => (got = r));
    http
      .expectOne(`${base()}/api/cursos/c1/habilidades`)
      .flush({ habilidades: [], porActividad: {} });
    expect(got).toEqual({ habilidades: [], porActividad: {} });
  });

  it("guarda la lista completa de una actividad", () => {
    service
      .setActivitySkills("a1", [{ nombre: "Compostaje", categoria: "Oficio" }])
      .subscribe();
    const req = http.expectOne(`${base()}/api/actividades/a1/habilidades`);
    expect(req.request.method).toBe("PUT");
    expect(req.request.body).toEqual({
      habilidades: [{ nombre: "Compostaje", categoria: "Oficio" }],
    });
    req.flush([]);
  });

  it("pide sugerencias con POST", () => {
    service.suggest({ cursoId: "c1", titulo: "t", consigna: "c" }).subscribe();
    const req = http.expectOne(`${base()}/api/actividades/sugerir-habilidades`);
    expect(req.request.method).toBe("POST");
    req.flush({ proveedor: "gemini", sugerencias: [] });
  });

  it("trae las habilidades del participante", () => {
    service.mySkills().subscribe();
    http.expectOne(`${base()}/api/participante/habilidades`).flush([]);
  });

  it("si ninguna IA responde, muestra el mensaje para cargar a mano", () => {
    let message = "";
    service
      .suggest({ cursoId: "c1", titulo: "t", consigna: "c" })
      .subscribe({ error: (e: Error) => (message = e.message) });
    http
      .expectOne(`${base()}/api/actividades/sugerir-habilidades`)
      .flush(
        { error: "otro texto" },
        { status: 503, statusText: "Service Unavailable" },
      );
    expect(message).toBe(SUGGESTIONS_UNAVAILABLE);
  });

  it("usa el mensaje del backend y explica 401 y 403", () => {
    const errores: string[] = [];
    const onError = { error: (e: Error) => errores.push(e.message) };
    service.searchCatalog("").subscribe(onError);
    service.searchCatalog("x").subscribe(onError);
    service.searchCatalog("y").subscribe(onError);
    service.searchCatalog("z").subscribe(onError);

    http
      .expectOne(`${base()}/api/habilidades?q=`)
      .flush(
        { error: "categoría inválida" },
        { status: 400, statusText: "Bad Request" },
      );
    http
      .expectOne(`${base()}/api/habilidades?q=x`)
      .flush(null, { status: 401, statusText: "Unauthorized" });
    http
      .expectOne(`${base()}/api/habilidades?q=y`)
      .flush(null, { status: 403, statusText: "Forbidden" });
    http
      .expectOne(`${base()}/api/habilidades?q=z`)
      .flush(null, { status: 500, statusText: "Error" });

    expect(errores[0]).toBe("categoría inválida");
    expect(errores[1]).toContain("sesión expiró");
    expect(errores[2]).toContain("permisos");
    expect(errores[3]).toBe("No se pudieron cargar las habilidades.");
  });
});

describe("skillKey", () => {
  it("normaliza como el backend", () => {
    expect(skillKey("  Trabajo en Équipo. ")).toBe("trabajo-en-equipo");
    expect(skillKey("Diseño de huerta")).toBe("diseno-de-huerta");
    expect(skillKey("...")).toBe("");
  });
});
