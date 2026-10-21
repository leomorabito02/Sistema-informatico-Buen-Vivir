import { TestBed } from "@angular/core/testing";
import { provideHttpClient } from "@angular/common/http";
import {
  HttpTestingController,
  provideHttpClientTesting,
} from "@angular/common/http/testing";
import { CampaignsService } from "./campaigns.service";
import { Campana, CampanaInput } from "../core/models/campaign.model";

const base = "/api/v1/campaigns";

const campana: Campana = {
  id: "c1",
  titulo: "Jornada de reciclaje",
  texto: "Traé tus botellas.",
  imagen_url: "https://example.org/a.jpg",
  fecha: "2026-11-05",
  estado: "borrador",
};

const input: CampanaInput = {
  titulo: "Jornada de reciclaje",
  texto: "Traé tus botellas.",
  imagen_url: "https://example.org/a.jpg",
  fecha: "2026-11-05",
};

describe("CampaignsService", () => {
  let service: CampaignsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(CampaignsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it("pide las públicas sin cookie y manda la búsqueda recortada", () => {
    service.publicas("  arbol ").subscribe((c) => expect(c).toEqual([campana]));
    const req = http.expectOne((r) => r.url.endsWith(`${base}/publicas`));
    expect(req.request.method).toBe("GET");
    expect(req.request.params.get("q")).toBe("arbol");
    expect(req.request.withCredentials).toBeFalse();
    req.flush([campana]);
  });

  it("no manda q si la búsqueda está vacía", () => {
    service.publicas("   ").subscribe();
    const req = http.expectOne((r) => r.url.endsWith(`${base}/publicas`));
    expect(req.request.params.has("q")).toBeFalse();
    req.flush([]);
  });

  it("lista para la gestión con cookie, búsqueda y estado", () => {
    service.listar("reciclaje", "publicada").subscribe();
    const req = http.expectOne((r) => r.url.endsWith(base));
    expect(req.request.withCredentials).toBeTrue();
    expect(req.request.params.get("q")).toBe("reciclaje");
    expect(req.request.params.get("estado")).toBe("publicada");
    req.flush([]);
  });

  it("crea, edita y publica con los métodos y rutas correctos", () => {
    const respuesta = { mensaje: "ok", campana };
    service.crear(input).subscribe((r) => expect(r.campana.id).toBe("c1"));
    const crear = http.expectOne((r) => r.url.endsWith(base) && r.method === "POST");
    expect(crear.request.body).toEqual(input);
    crear.flush(respuesta);

    service.editar("c1", input).subscribe();
    const editar = http.expectOne((r) => r.url.endsWith(`${base}/c1`));
    expect(editar.request.method).toBe("PUT");
    editar.flush(respuesta);

    service.publicar("c1").subscribe();
    const publicar = http.expectOne((r) => r.url.endsWith(`${base}/c1/publicar`));
    expect(publicar.request.method).toBe("POST");
    expect(publicar.request.withCredentials).toBeTrue();
    publicar.flush(respuesta);
  });

  it("usa el mensaje de error del backend", (done) => {
    service.crear(input).subscribe({
      error: (e: Error) => {
        expect(e.message).toBe("Completar los campos obligatorios: imagen");
        done();
      },
    });
    http
      .expectOne((r) => r.url.endsWith(base))
      .flush({ error: "Completar los campos obligatorios: imagen" }, { status: 400, statusText: "Bad Request" });
  });

  it("traduce 401 y 403 sin cuerpo", () => {
    const mensajes: string[] = [];
    service.listar().subscribe({ error: (e: Error) => mensajes.push(e.message) });
    http.expectOne((r) => r.url.endsWith(base)).flush("", { status: 401, statusText: "Unauthorized" });
    service.listar().subscribe({ error: (e: Error) => mensajes.push(e.message) });
    http.expectOne((r) => r.url.endsWith(base)).flush("", { status: 403, statusText: "Forbidden" });
    expect(mensajes[0]).toContain("sesión");
    expect(mensajes[1]).toContain("permisos");
  });
});
