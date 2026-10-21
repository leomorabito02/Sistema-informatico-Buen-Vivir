import { TestBed } from "@angular/core/testing";
import { provideHttpClient } from "@angular/common/http";
import {
  HttpTestingController,
  provideHttpClientTesting,
} from "@angular/common/http/testing";
import { ExportarInput, ReportesService } from "./reportes.service";
import { ReporteExportado } from "../core/models/dashboard.model";

const base = "/api/v1/dashboard/reportes";

describe("ReportesService", () => {
  let service: ReportesService;
  let http: HttpTestingController;

  const input: ExportarInput = {
    filtros: {
      desde: new Date(2026, 0, 1),
      hasta: new Date(2026, 5, 30),
      modulo: "educacion",
      areaId: "a1",
    },
    areaNombre: "Oficios",
    indicadores: [
      {
        id: "cursosEnDesarrollo",
        titulo: "Cursos en desarrollo",
        descripcion: "",
        modulo: "educacion",
        fuente: "educacion",
        valor: 3,
        disponible: true,
      },
      {
        id: "cursosFinalizados",
        titulo: "Cursos finalizados",
        descripcion: "",
        modulo: "educacion",
        fuente: "educacion",
        valor: null,
        disponible: false,
      },
    ],
    graficas: [
      {
        titulo: "Cursos",
        imagen: null,
        ancho: 1,
        alto: 1,
        resumen: "Sin datos.",
      },
    ],
    generadoPor: "Sofía Cuozzo",
  };

  const guardado: ReporteExportado = {
    id: "r1",
    tipo_reporte: "dashboard",
    periodo_desde: "2026-01-01",
    periodo_hasta: "2026-06-30",
    filtros_aplicados: { modulo: "educacion", areaId: "a1", areaNombre: "Oficios" },
    generado_por: "admin-1",
    generado_por_nombre: "Sofía Cuozzo",
    fecha_generacion: "2026-10-20T15:00:00Z",
    url_archivo: "gs://ceibuenvivir-bucket/reportes/r1.pdf",
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(ReportesService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it("arma los datos del reporte con período, filtros y nombre", () => {
    expect(service.metadatos(input)).toEqual({
      periodo_desde: "2026-01-01",
      periodo_hasta: "2026-06-30",
      filtros_aplicados: { modulo: "educacion", areaId: "a1", areaNombre: "Oficios" },
      generado_por_nombre: "Sofía Cuozzo",
    });
  });

  it("carga el historial guardado y avisa si falla", () => {
    service.cargarHistorial();
    expect(service.cargandoHistorial()).toBeTrue();
    const req = http.expectOne((r) => r.url.endsWith(base));
    expect(req.request.withCredentials).toBeTrue();
    req.flush([guardado]);
    expect(service.historial()).toEqual([guardado]);
    expect(service.cargandoHistorial()).toBeFalse();

    service.cargarHistorial();
    http
      .expectOne((r) => r.url.endsWith(base))
      .flush({ error: "El historial de reportes no está configurado en este entorno" }, { status: 503, statusText: "Unavailable" });
    expect(service.errorHistorial()).toBe("El historial de reportes no está configurado en este entorno");
  });

  it("al exportar descarga el PDF y lo manda al backend como multipart", async () => {
    const click = spyOn(HTMLAnchorElement.prototype, "click");
    const promesa = service.exportar(input);
    // El PDF se arma con jsPDF (carga diferida): esperar a que salga el pedido.
    let req;
    for (let i = 0; i < 50 && !req; i++) {
      await new Promise((r) => setTimeout(r, 20));
      req = http.match((r) => r.url.endsWith(base) && r.method === "POST")[0];
    }
    expect(click).toHaveBeenCalled();
    expect(req).toBeDefined();
    const form = req!.request.body as FormData;
    expect(JSON.parse(form.get("metadatos") as string).periodo_desde).toBe("2026-01-01");
    const archivo = form.get("archivo") as Blob;
    expect(archivo.type).toBe("application/pdf");
    req!.flush(guardado);

    const r = await promesa;
    expect(r.guardado).toBeTrue();
    expect(service.historial()[0]).toEqual(guardado);
  });

  it("si no se puede guardar, el PDF igual se descargó y se informa por qué", async () => {
    const click = spyOn(HTMLAnchorElement.prototype, "click");
    const promesa = service.exportar(input);
    let req;
    for (let i = 0; i < 50 && !req; i++) {
      await new Promise((r) => setTimeout(r, 20));
      req = http.match((r) => r.url.endsWith(base) && r.method === "POST")[0];
    }
    req!.flush({ error: "El PDF no puede superar los 10 MB" }, { status: 400, statusText: "Bad Request" });
    const r = await promesa;
    expect(click).toHaveBeenCalled();
    expect(r.guardado).toBeFalse();
    expect(r.error).toBe("El PDF no puede superar los 10 MB");
    expect(service.historial().length).toBe(0);
  });

  it("vuelve a descargar un PDF guardado", async () => {
    const click = spyOn(HTMLAnchorElement.prototype, "click");
    const promesa = service.descargar(guardado);
    const req = http.expectOne((r) => r.url.endsWith(`${base}/r1/pdf`));
    expect(req.request.responseType).toBe("blob");
    req.flush(new Blob(["%PDF-1.7"], { type: "application/pdf" }));
    await promesa;
    expect(click).toHaveBeenCalled();
  });

  it("genera un PDF real", async () => {
    const blob = await service.generarPdf(input);
    expect(blob.type).toBe("application/pdf");
    expect(blob.size).toBeGreaterThan(1000);
  });
});
