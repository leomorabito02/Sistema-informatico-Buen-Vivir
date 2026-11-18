import { TestBed } from "@angular/core/testing";
import { provideHttpClient } from "@angular/common/http";
import {
  HttpTestingController,
  provideHttpClientTesting,
} from "@angular/common/http/testing";
import { ContenidosService, tamanoLegible } from "./contenidos.service";
import { environment } from "../../environments/environment";

describe("ContenidosService", () => {
  let service: ContenidosService;
  let http: HttpTestingController;
  const base = environment.educationApiUrl || environment.apiUrl || "";

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(ContenidosService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it("uploads the material as multipart", () => {
    const archivo = new File(["%PDF"], "guia.pdf", { type: "application/pdf" });
    service.subir({ titulo: "Guía", descripcion: "Desc", archivo }).subscribe();
    const req = http.expectOne(`${base}/api/contenidos-docentes`);
    expect(req.request.method).toBe("POST");
    const fd = req.request.body as FormData;
    expect(fd.get("titulo")).toBe("Guía");
    expect((fd.get("archivo") as File).name).toBe("guia.pdf");
    expect(req.request.withCredentials).toBeTrue();
    req.flush({});
  });

  it("edits without a file when none is chosen", () => {
    service.editar("c1", { titulo: "Guía", descripcion: "" }).subscribe();
    const req = http.expectOne(`${base}/api/contenidos-docentes/c1`);
    expect(req.request.method).toBe("PUT");
    expect((req.request.body as FormData).has("archivo")).toBeFalse();
    req.flush({});
  });

  it("lists by state and moderates", () => {
    service.listar("pendiente").subscribe();
    http
      .expectOne(`${base}/api/contenidos-docentes?estado=pendiente`)
      .flush([]);

    service.moderar("c1", "rechazado", "Falta la fuente").subscribe();
    const req = http.expectOne(`${base}/api/contenidos-docentes/c1/moderacion`);
    expect(req.request.body).toEqual({
      estado: "rechazado",
      motivo: "Falta la fuente",
    });
    req.flush({});
  });

  it("public download sends the visitor data without credentials", () => {
    const datos = {
      nombre: "Ana",
      apellido: "Gómez",
      email: "ana@gmail.com",
      telefono: "3547123456",
    };
    service.descargarPublico("c1", datos).subscribe();
    const req = http.expectOne(`${base}/api/biblioteca/c1/descargas`);
    expect(req.request.body).toEqual(datos);
    expect(req.request.withCredentials).toBeFalse();
    expect(req.request.responseType).toBe("blob");
    req.flush(new Blob(["%PDF"]));
  });

  it("reads the error message even when the response is a blob", (done) => {
    service
      .descargarPublico("c1", {
        nombre: "",
        apellido: "",
        email: "",
        telefono: "",
      })
      .subscribe({
        error: (e: Error) => {
          expect(e.message).toBe("El email no es válido");
          done();
        },
      });
    http
      .expectOne(`${base}/api/biblioteca/c1/descargas`)
      .flush(new Blob([JSON.stringify({ error: "El email no es válido" })]), {
        status: 400,
        statusText: "Bad Request",
      });
  });

  it("explains a rate limit", (done) => {
    service.biblioteca().subscribe({
      error: (e: Error) => {
        expect(e.message).toContain("muchos pedidos");
        done();
      },
    });
    http
      .expectOne(`${base}/api/biblioteca`)
      .flush(null, { status: 429, statusText: "Too Many" });
  });

  it("formats sizes", () => {
    expect(tamanoLegible(512)).toBe("1 KB");
    expect(tamanoLegible(850 * 1024)).toBe("850 KB");
    expect(tamanoLegible(2.5 * 1024 * 1024)).toBe("2,5 MB");
  });
});
