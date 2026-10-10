import { TestBed } from "@angular/core/testing";
import { ExportarInput, ReportesService } from "./reportes.service";

describe("ReportesService", () => {
  let service: ReportesService;

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

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(ReportesService);
  });

  it("registra el ReporteExportado con período, filtros y autor", () => {
    const r = service.registrar(input, "blob:1");
    expect(r.tipo_reporte).toBe("dashboard");
    expect(r.periodo_desde).toBe("2026-01-01");
    expect(r.periodo_hasta).toBe("2026-06-30");
    expect(r.filtros_aplicados).toEqual({
      modulo: "educacion",
      areaId: "a1",
      areaNombre: "Oficios",
    });
    expect(r.generado_por).toBe("Sofía Cuozzo");
    expect(r.url_archivo).toBe("blob:1");
    expect(service.historial()[0]).toBe(r);

    const r2 = service.registrar(input, "blob:2");
    expect(service.historial().map((x) => x.id)).toEqual([r2.id, r.id]);
  });

  it("genera un PDF y lo descarga al exportar", async () => {
    const click = spyOn(HTMLAnchorElement.prototype, "click");
    const r = await service.exportar(input);
    expect(r.url_archivo.startsWith("blob:")).toBeTrue();
    expect(click).toHaveBeenCalled();
    expect(service.historial().length).toBe(1);

    const blob = await service.generarPdf(input);
    expect(blob.type).toBe("application/pdf");
    expect(blob.size).toBeGreaterThan(1000);
  });
});
