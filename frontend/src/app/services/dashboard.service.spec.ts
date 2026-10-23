import { TestBed } from "@angular/core/testing";
import { provideHttpClient } from "@angular/common/http";
import {
  HttpTestingController,
  provideHttpClientTesting,
} from "@angular/common/http/testing";
import { DashboardService } from "./dashboard.service";
import {
  etiquetaMes,
  fechaLocal,
  formatoFecha,
  inicioDelDia,
} from "./dashboard.fechas";
import { DashboardData } from "../core/models/dashboard.model";

// Los cálculos se prueban en el backend (dashboard/internal/services/
// indicadores_test.go). Acá solo se prueba que el frontend pida bien el
// resumen y lo entregue tal cual.

const respuesta: DashboardData = {
  indicadores: [
    {
      id: "insercionesLaborales",
      titulo: "Inserciones laborales",
      descripcion: "desc",
      modulo: "insercion",
      fuente: "insercion",
      valor: 2,
      disponible: true,
    },
  ],
  graficas: [],
};

describe("DashboardService", () => {
  let http: HttpTestingController;
  let service: DashboardService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    http = TestBed.inject(HttpTestingController);
    service = TestBed.inject(DashboardService);
  });

  afterEach(() => http.verify());

  it("pide el resumen con los filtros como fechas AAAA-MM-DD", () => {
    let recibido: DashboardData | undefined;
    service
      .cargar({
        desde: new Date(2026, 0, 5),
        hasta: new Date(2026, 5, 30, 23, 59),
        modulo: "educacion",
        areaId: "huerta",
      })
      .subscribe((d) => (recibido = d));
    const req = http.expectOne((r) =>
      r.url.endsWith("/api/v1/dashboard/resumen"),
    );
    expect(req.request.method).toBe("GET");
    expect(req.request.params.get("desde")).toBe("2026-01-05");
    expect(req.request.params.get("hasta")).toBe("2026-06-30");
    expect(req.request.params.get("modulo")).toBe("educacion");
    expect(req.request.params.get("area")).toBe("huerta");
    // En Docker la API está en otro puerto: sin esto no viaja la cookie.
    expect(req.request.withCredentials).toBeTrue();
    req.flush(respuesta);
    expect(recibido).toEqual(respuesta);
  });

  it("sin área no manda el parámetro", () => {
    service
      .cargar({
        desde: new Date(2026, 0, 1),
        hasta: new Date(2026, 0, 31),
        modulo: "todos",
        areaId: null,
      })
      .subscribe();
    const req = http.expectOne((r) =>
      r.url.endsWith("/api/v1/dashboard/resumen"),
    );
    expect(req.request.params.has("area")).toBeFalse();
    req.flush(respuesta);
  });

  it("si el Dashboard falla, el error llega a la página", () => {
    let error: unknown;
    service
      .cargar({
        desde: new Date(2026, 0, 1),
        hasta: new Date(2026, 0, 31),
        modulo: "todos",
        areaId: null,
      })
      .subscribe({ error: (e) => (error = e) });
    http
      .expectOne((r) => r.url.endsWith("/api/v1/dashboard/resumen"))
      .flush(
        { error: "El servicio de origen no respondió." },
        { status: 502, statusText: "Bad Gateway" },
      );
    expect(error).toBeTruthy();
  });

  it("pide las áreas formativas", () => {
    service.areas().subscribe((a) => expect(a.length).toBe(1));
    const req = http.expectOne((r) =>
      r.url.endsWith("/api/v1/dashboard/areas"),
    );
    expect(req.request.withCredentials).toBeTrue();
    req.flush([{ id: "a1", nombre: "Oficios" }]);
  });
});

describe("dashboard.fechas", () => {
  it("formatea fechas para los filtros, la pantalla y el PDF", () => {
    expect(fechaLocal(new Date(2026, 2, 9, 23, 30))).toBe("2026-03-09");
    expect(formatoFecha(new Date(2026, 2, 9))).toBe("09/03/2026");
    expect(inicioDelDia(new Date(2026, 2, 9, 15)).getHours()).toBe(0);
    expect(etiquetaMes("2026-03")).toBe("mar 2026");
  });
});
