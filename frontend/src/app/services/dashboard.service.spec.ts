import { TestBed } from "@angular/core/testing";
import { of, throwError } from "rxjs";
import {
  CursoOrigen,
  FiltrosDashboard,
  ParticipanteHabilidadOrigen,
  ParticipanteOrigen,
  PostulacionOrigen,
} from "../core/models/dashboard.model";
import {
  DashboardService,
  DatosOrigen,
  calcularDashboard,
  visibleEnModulo,
} from "./dashboard.service";
import { DashboardSourcesService } from "./dashboard-sources.service";
import { provideHttpClient } from "@angular/common/http";
import {
  HttpTestingController,
  provideHttpClientTesting,
} from "@angular/common/http/testing";
import {
  estadoCurso,
  mesesDelPeriodo,
  participanteActivoEn,
} from "./dashboard.calculos";

const d = (s: string) => new Date(s + "T12:00:00");

function filtros(over: Partial<FiltrosDashboard> = {}): FiltrosDashboard {
  return {
    desde: new Date(2026, 0, 1),
    hasta: new Date(2026, 5, 30),
    modulo: "todos",
    areaId: null,
    ...over,
  };
}

const participantes: ParticipanteOrigen[] = [
  {
    id: "p1",
    rol: "participante",
    estado: "activo",
    fechaAlta: d("2025-12-01").toISOString(),
  },
  {
    id: "p2",
    rol: "participante",
    estado: "activo",
    fechaAlta: d("2026-03-10").toISOString(),
  },
  // Baja lógica en abril: cuenta hasta marzo.
  {
    id: "p3",
    rol: "participante",
    estado: "baja",
    fechaAlta: d("2025-11-01").toISOString(),
    fechaBaja: d("2026-04-15").toISOString(),
  },
  // Alta posterior al período: no cuenta.
  {
    id: "p4",
    rol: "participante",
    estado: "activo",
    fechaAlta: d("2026-08-01").toISOString(),
  },
];

const curso = (
  id: string,
  areaId: string,
  fechas: string[],
  estado: CursoOrigen["estado"] = "activo",
  creacion = "2026-01-05",
): CursoOrigen => ({
  id,
  titulo: id,
  areaId,
  estado,
  fechaCreacion: d(creacion).toISOString(),
  actividades: fechas.map((f, i) => ({
    id: `${id}-a${i}`,
    fecha: d(f).toISOString(),
  })),
});

const cursos: CursoOrigen[] = [
  curso("c1", "huerta", ["2026-02-01", "2026-03-01"]), // finaliza en marzo
  curso("c2", "huerta", ["2026-02-01", "2026-09-01"]), // pendiente al 30/06
  curso("c3", "oficios", ["2026-04-01"]), // finaliza en abril
  curso("c4", "oficios", ["2026-02-01"], "borrador"), // borrador: no cuenta
];

const hab = (
  id: string,
  areaId: string,
  fecha: string,
  estado: ParticipanteHabilidadOrigen["estado"] = "aceptada",
  origen: ParticipanteHabilidadOrigen["origen"] = "actividad",
): ParticipanteHabilidadOrigen => ({
  id,
  participanteId: "p1",
  cursoId: "c1",
  areaId,
  origen,
  estado,
  fecha: d(fecha).toISOString(),
});

const habilidades: ParticipanteHabilidadOrigen[] = [
  hab("h1", "huerta", "2026-02-10"),
  hab("h2", "huerta", "2026-02-20", "aceptada", "manual"),
  hab("h3", "oficios", "2026-05-02"),
  hab("h4", "huerta", "2026-03-01", "pendiente"),
  hab("h5", "huerta", "2026-03-01", "rechazada"),
  hab("h6", "huerta", "2025-12-01"), // fuera del período
];

const post = (
  id: string,
  estado: PostulacionOrigen["estado"],
  fecha: string,
): PostulacionOrigen => ({
  id,
  participanteId: "p1",
  estado,
  fechaEstado: d(fecha).toISOString(),
});

const postulaciones: PostulacionOrigen[] = [
  post("s1", "seleccionado", "2026-01-20"),
  post("s2", "seleccionado", "2026-06-29"),
  post("s3", "rechazado", "2026-02-01"),
  post("s4", "postulado", "2026-03-01"),
  post("s5", "seleccionado", "2026-07-15"), // fuera del período
];

const todos: DatosOrigen = {
  participantes,
  cursos,
  habilidades,
  postulaciones,
};

const valor = (data: ReturnType<typeof calcularDashboard>, id: string) =>
  data.indicadores.find((i) => i.id === id)?.valor;

describe("dashboard.calculos", () => {
  it("estadoCurso: en desarrollo con actividades pendientes, finalizado si ya ocurrieron todas", () => {
    const c = curso("x", "a", ["2026-02-01", "2026-03-01"]);
    expect(estadoCurso(c, d("2026-02-15"))).toBe("en_desarrollo");
    expect(estadoCurso(c, d("2026-03-02"))).toBe("finalizado");
  });

  it("estadoCurso: borradores y cursos que todavía no existían no cuentan", () => {
    expect(
      estadoCurso(curso("b", "a", ["2026-02-01"], "borrador"), d("2026-06-01")),
    ).toBe("fuera");
    expect(
      estadoCurso(
        curso("f", "a", ["2026-09-01"], "activo", "2026-08-01"),
        d("2026-06-01"),
      ),
    ).toBe("fuera");
  });

  it("estadoCurso: un curso activo sin actividades está en desarrollo", () => {
    expect(estadoCurso(curso("v", "a", []), d("2026-06-01"))).toBe(
      "en_desarrollo",
    );
  });

  it("participanteActivoEn excluye bajas lógicas desde su fecha de baja", () => {
    const p3 = participantes[2];
    expect(participanteActivoEn(p3, d("2026-03-31"))).toBeTrue();
    expect(participanteActivoEn(p3, d("2026-04-30"))).toBeFalse();
    expect(
      participanteActivoEn({ ...p3, fechaBaja: null }, d("2026-01-01")),
    ).toBeFalse();
  });

  it("mesesDelPeriodo devuelve un mes por punto y corta en la fecha 'hasta'", () => {
    const meses = mesesDelPeriodo(new Date(2026, 0, 15), new Date(2026, 2, 10));
    expect(meses.map((m) => m.mes)).toEqual(["2026-01", "2026-02", "2026-03"]);
    expect(meses[2].corte.getDate()).toBe(10);
    expect(meses[0].inicio.getDate()).toBe(15);
    expect(mesesDelPeriodo(new Date(2026, 5, 1), new Date(2026, 0, 1))).toEqual(
      [],
    );
  });
});

describe("calcularDashboard", () => {
  it("calcula los cinco indicadores (HU-52)", () => {
    const data = calcularDashboard(todos, filtros());
    expect(data.indicadores.map((i) => i.id)).toEqual([
      "participantesActivos",
      "cursosEnDesarrollo",
      "cursosFinalizados",
      "habilidadesDesarrolladas",
      "insercionesLaborales",
    ]);
    expect(valor(data, "participantesActivos")).toBe(2); // p1, p2
    expect(valor(data, "cursosEnDesarrollo")).toBe(1); // c2
    expect(valor(data, "cursosFinalizados")).toBe(2); // c1, c3
    expect(valor(data, "habilidadesDesarrolladas")).toBe(3); // h1, h2 (manual), h3
    expect(valor(data, "insercionesLaborales")).toBe(2); // s1, s2
    expect(data.indicadores.every((i) => i.disponible)).toBeTrue();
  });

  it("arma una gráfica por indicador con un punto por mes (HU-53)", () => {
    const data = calcularDashboard(todos, filtros());
    expect(data.graficas.map((g) => g.id)).toEqual([
      "participantes",
      "cursos",
      "habilidades",
      "inserciones",
    ]);
    const cursosG = data.graficas.find((g) => g.id === "cursos")!;
    expect(cursosG.tipo).toBe("bar");
    expect(cursosG.series!.map((s) => s.nombre)).toEqual([
      "En desarrollo",
      "Finalizados",
    ]);
    expect(cursosG.series![0].puntos.length).toBe(6);
    // Febrero: c1, c2 y c3 en desarrollo; marzo: c1 ya finalizó.
    expect(cursosG.series![0].puntos[1].valor).toBe(3);
    expect(cursosG.series![1].puntos[2].valor).toBe(1);

    const part = data.graficas.find((g) => g.id === "participantes")!;
    // Enero: p1 y p3; marzo: p1, p2, p3; abril: p3 se dio de baja.
    expect(part.series![0].puntos.map((p) => p.valor)).toEqual([
      2, 2, 3, 2, 2, 2,
    ]);

    const habG = data.graficas.find((g) => g.id === "habilidades")!;
    expect(habG.series![0].puntos.map((p) => p.valor)).toEqual([
      0, 2, 0, 0, 1, 0,
    ]);
  });

  it("el período cambia indicadores y gráficas (HU-54)", () => {
    const data = calcularDashboard(
      todos,
      filtros({ desde: new Date(2026, 0, 1), hasta: new Date(2026, 1, 28) }),
    );
    expect(valor(data, "cursosFinalizados")).toBe(0);
    expect(valor(data, "cursosEnDesarrollo")).toBe(3);
    expect(valor(data, "habilidadesDesarrolladas")).toBe(2);
    expect(valor(data, "insercionesLaborales")).toBe(1);
    expect(data.graficas[0].series![0].puntos.length).toBe(2);
  });

  it("el área formativa filtra cursos y habilidades, no participantes ni inserciones", () => {
    const data = calcularDashboard(todos, filtros({ areaId: "oficios" }));
    expect(valor(data, "cursosEnDesarrollo")).toBe(0);
    expect(valor(data, "cursosFinalizados")).toBe(1);
    expect(valor(data, "habilidadesDesarrolladas")).toBe(1);
    expect(valor(data, "participantesActivos")).toBe(2);
    expect(valor(data, "insercionesLaborales")).toBe(2);
  });

  it("el módulo muestra solo sus indicadores y gráficas", () => {
    const edu = calcularDashboard(todos, filtros({ modulo: "educacion" }));
    expect(edu.indicadores.map((i) => i.id)).toEqual([
      "cursosEnDesarrollo",
      "cursosFinalizados",
      "habilidadesDesarrolladas",
    ]);
    expect(edu.graficas.map((g) => g.id)).toEqual(["cursos", "habilidades"]);

    const ins = calcularDashboard(todos, filtros({ modulo: "insercion" }));
    expect(ins.indicadores.map((i) => i.id)).toEqual(["insercionesLaborales"]);

    const pro = calcularDashboard(todos, filtros({ modulo: "promocion" }));
    expect(pro.indicadores).toEqual([]);
    expect(pro.graficas).toEqual([]);

    expect(visibleEnModulo("usuarios", "todos")).toBeTrue();
    expect(visibleEnModulo("usuarios", "educacion")).toBeFalse();
  });

  it("si una fuente falla, solo sus indicadores quedan como no disponibles", () => {
    const data = calcularDashboard({ ...todos, cursos: null }, filtros());
    const nd = data.indicadores.filter((i) => !i.disponible).map((i) => i.id);
    expect(nd).toEqual(["cursosEnDesarrollo", "cursosFinalizados"]);
    expect(valor(data, "cursosEnDesarrollo")).toBeNull();
    expect(valor(data, "participantesActivos")).toBe(2);
    expect(data.graficas.find((g) => g.id === "cursos")!.series).toBeNull();
    expect(
      data.graficas.find((g) => g.id === "participantes")!.series,
    ).not.toBeNull();
  });
});

describe("DashboardService", () => {
  let sources: jasmine.SpyObj<DashboardSourcesService>;
  let service: DashboardService;

  beforeEach(() => {
    sources = jasmine.createSpyObj("DashboardSourcesService", [
      "participantes",
      "cursos",
      "habilidades",
      "postulaciones",
      "areas",
    ]);
    sources.participantes.and.returnValue(of(participantes));
    sources.cursos.and.returnValue(of(cursos));
    sources.habilidades.and.returnValue(of(habilidades));
    sources.postulaciones.and.returnValue(of(postulaciones));
    sources.areas.and.returnValue(of([]));
    TestBed.configureTestingModule({
      providers: [{ provide: DashboardSourcesService, useValue: sources }],
    });
    service = TestBed.inject(DashboardService);
  });

  it("combina las cuatro fuentes y calcula", (done) => {
    service.cargar(filtros()).subscribe((data) => {
      expect(data.indicadores.length).toBe(5);
      expect(valor(data, "insercionesLaborales")).toBe(2);
      done();
    });
  });

  it("no se rompe si un servicio de origen falla", (done) => {
    sources.postulaciones.and.returnValue(throwError(() => new Error("502")));
    service.cargar(filtros()).subscribe((data) => {
      const ins = data.indicadores.find(
        (i) => i.id === "insercionesLaborales",
      )!;
      expect(ins.disponible).toBeFalse();
      expect(data.indicadores.filter((i) => i.disponible).length).toBe(4);
      done();
    });
  });

  it("expone las áreas de la fuente de Educación", (done) => {
    service.areas().subscribe((a) => {
      expect(a).toEqual([]);
      expect(sources.areas).toHaveBeenCalled();
      done();
    });
  });
});

describe("DashboardSourcesService", () => {
  let http: HttpTestingController;
  let sources: DashboardSourcesService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    http = TestBed.inject(HttpTestingController);
    sources = TestBed.inject(DashboardSourcesService);
  });

  afterEach(() => http.verify());

  it("pide cada fuente al microservicio de Dashboard", () => {
    const llamadas: [() => unknown, string][] = [
      [() => sources.participantes(), "participantes"],
      [() => sources.cursos(), "cursos"],
      [() => sources.habilidades(), "habilidades"],
      [() => sources.postulaciones(), "postulaciones"],
      [() => sources.areas(), "areas"],
    ];
    for (const [llamar, ruta] of llamadas) {
      let recibido: unknown;
      (llamar() as ReturnType<typeof sources.areas>).subscribe(
        (r) => (recibido = r),
      );
      const req = http.expectOne((r) =>
        r.url.endsWith(`/api/v1/dashboard/${ruta}`),
      );
      expect(req.request.method).toBe("GET");
      // En Docker la API está en otro puerto: sin esto no viaja la cookie.
      expect(req.request.withCredentials).toBeTrue();
      req.flush([{ id: ruta }]);
      expect(recibido).toEqual([{ id: ruta }]);
    }
  });

  it("si una fuente falla, el error llega a DashboardService", () => {
    let error: unknown;
    sources.cursos().subscribe({ error: (e) => (error = e) });
    http
      .expectOne((r) => r.url.endsWith("/api/v1/dashboard/cursos"))
      .flush(
        { error: "El servicio de origen no respondió." },
        { status: 502, statusText: "Bad Gateway" },
      );
    expect(error).toBeTruthy();
  });
});
