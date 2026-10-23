import {
  ComponentFixture,
  TestBed,
  fakeAsync,
  tick,
} from "@angular/core/testing";
import { provideNoopAnimations } from "@angular/platform-browser/animations";
import { provideRouter } from "@angular/router";
import { Subject, of, throwError } from "rxjs";
import { DashboardPageComponent } from "./dashboard-page.component";
import { DashboardService } from "../../../services/dashboard.service";
import { ReportesService } from "../../../services/reportes.service";
import { AuthService } from "../../../services/auth.service";
import {
  DashboardData,
  FiltrosDashboard,
  Grafica,
  Indicador,
  ReporteExportado,
} from "../../../core/models/dashboard.model";
import { KpiCardComponent } from "./kpi-card.component";
import {
  GraficaEvolucionComponent,
  resumirSeries,
} from "./grafica-evolucion.component";
import { HistorialReportesComponent } from "./historial-reportes.component";
import { MeResponse } from "../../../core/models/auth.model";

// Lo que devuelve el backend (GET /api/v1/dashboard/resumen) para un
// período de dos meses sin datos.
const ceros = [
  { mes: "2026-01", valor: 0 },
  { mes: "2026-02", valor: 0 },
];
const indicador = (
  id: Indicador["id"],
  titulo: string,
  modulo: Indicador["modulo"],
  fuente: Indicador["fuente"],
): Indicador => ({
  id,
  titulo,
  descripcion: titulo,
  modulo,
  fuente,
  valor: 0,
  disponible: true,
});
const grafica = (
  id: Grafica["id"],
  modulo: Grafica["modulo"],
  series: string[],
): Grafica => ({
  id,
  titulo: id,
  tipo: "line",
  modulo,
  indicadores: [],
  series: series.map((nombre) => ({ nombre, puntos: ceros })),
});
const vacio: DashboardData = {
  indicadores: [
    indicador(
      "participantesActivos",
      "Participantes activos",
      "usuarios",
      "autenticacion",
    ),
    indicador(
      "cursosEnDesarrollo",
      "Cursos en desarrollo",
      "educacion",
      "educacion",
    ),
    indicador(
      "cursosFinalizados",
      "Cursos finalizados",
      "educacion",
      "educacion",
    ),
    indicador(
      "habilidadesDesarrolladas",
      "Habilidades desarrolladas",
      "educacion",
      "educacion",
    ),
    indicador(
      "insercionesLaborales",
      "Inserciones laborales",
      "insercion",
      "insercion",
    ),
    indicador("fichasActivas", "Fichas activas", "promocion", "promocion"),
    indicador(
      "participantesConSeguimiento",
      "Participantes con seguimiento",
      "promocion",
      "promocion",
    ),
  ],
  graficas: [
    grafica("participantes", "usuarios", ["Participantes activos"]),
    grafica("cursos", "educacion", ["En desarrollo", "Finalizados"]),
    grafica("habilidades", "educacion", ["Habilidades desarrolladas"]),
    grafica("inserciones", "insercion", ["Inserciones laborales"]),
    grafica("promocion", "promocion", ["Fichas activas", "Con seguimiento"]),
  ],
};

describe("DashboardPageComponent", () => {
  let fixture: ComponentFixture<DashboardPageComponent>;
  let component: DashboardPageComponent;
  let dashboard: jasmine.SpyObj<DashboardService>;
  let reportes: jasmine.SpyObj<ReportesService>;
  let respuesta$: Subject<DashboardData>;

  beforeEach(async () => {
    respuesta$ = new Subject<DashboardData>();
    dashboard = jasmine.createSpyObj("DashboardService", ["cargar", "areas"]);
    dashboard.cargar.and.callFake(() => respuesta$);
    dashboard.areas.and.returnValue(of([{ id: "a1", nombre: "Oficios" }]));
    reportes = jasmine.createSpyObj(
      "ReportesService",
      ["exportar", "descargar", "cargarHistorial"],
      {
        historial: () => [],
        cargandoHistorial: () => false,
        errorHistorial: () => null,
      },
    );
    reportes.exportar.and.resolveTo({
      reporte: {} as ReporteExportado,
      guardado: true,
    });
    const auth = jasmine.createSpyObj("AuthService", ["isAuthenticated"]);
    auth.isAuthenticated.and.returnValue(
      of({ user: { first_name: "Sofía", last_name: "Cuozzo" } } as MeResponse),
    );

    await TestBed.configureTestingModule({
      imports: [DashboardPageComponent],
      providers: [
        provideRouter([]),
        provideNoopAnimations(),
        { provide: DashboardService, useValue: dashboard },
        { provide: ReportesService, useValue: reportes },
        { provide: AuthService, useValue: auth },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardPageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("muestra esqueletos mientras carga y después los siete indicadores", () => {
    const el: HTMLElement = fixture.nativeElement;
    expect(component.cargando).toBeTrue();
    expect(reportes.cargarHistorial).toHaveBeenCalled();
    expect(el.querySelectorAll("app-kpi-card").length).toBe(7);
    expect(el.querySelector("p-skeleton")).not.toBeNull();

    respuesta$.next(vacio);
    fixture.detectChanges();
    expect(component.cargando).toBeFalse();
    expect(el.textContent).toContain("Participantes activos");
    expect(el.querySelectorAll("app-grafica-evolucion").length).toBe(5);
    expect(el.textContent).toContain("Fichas activas");
    expect(el.textContent).toContain(
      "No hay datos para el período seleccionado",
    );
    expect(el.querySelector("app-back-link")).not.toBeNull();
  });

  it("al cambiar los filtros vuelve a cargar con los nuevos", () => {
    const f: FiltrosDashboard = {
      desde: new Date(2026, 0, 1),
      hasta: new Date(2026, 2, 31),
      modulo: "educacion",
      areaId: "a1",
    };
    component.onFiltros(f);
    expect(dashboard.cargar).toHaveBeenCalledWith(f);
    expect(component.cargando).toBeTrue();
  });

  it("si el módulo no tiene indicadores, lo explica y no deja exportar", () => {
    respuesta$.next({ indicadores: [], graficas: [] });
    fixture.detectChanges();
    expect(component.moduloSinIndicadores).toBeTrue();
    expect(component.puedeExportar).toBeFalse();
    expect(fixture.nativeElement.textContent).toContain(
      "Este módulo todavía no tiene indicadores",
    );
  });

  it("exporta con los filtros, indicadores y gráficas actuales", fakeAsync(() => {
    respuesta$.next(vacio);
    fixture.detectChanges();
    component.exportar();
    tick();
    const arg = reportes.exportar.calls.mostRecent().args[0];
    expect(arg.filtros).toBe(component.filtros);
    expect(arg.indicadores.length).toBe(7);
    expect(arg.graficas.length).toBe(5);
    expect(arg.generadoPor).toBe("Sofía Cuozzo");
    expect(component.exportando).toBeFalse();
  }));

  it("si el PDF se descargó pero no se guardó, avisa sin decir que falló todo", fakeAsync(() => {
    reportes.exportar.and.resolveTo({
      reporte: null,
      guardado: false,
      error: "El historial de reportes no está configurado en este entorno",
    });
    respuesta$.next(vacio);
    fixture.detectChanges();
    component.exportar();
    tick();
    fixture.detectChanges();
    expect(document.body.textContent).toContain("no quedó en el historial");
    expect(document.body.textContent).not.toContain("No se pudo exportar");
  }));

  it("avisa si la exportación falla", fakeAsync(() => {
    reportes.exportar.and.rejectWith(new Error("x"));
    respuesta$.next(vacio);
    fixture.detectChanges();
    component.exportar();
    tick();
    fixture.detectChanges();
    expect(component.exportando).toBeFalse();
    expect(document.body.textContent).toContain("No se pudo exportar");
  }));

  it("si el Dashboard no responde, muestra el error y permite reintentar", () => {
    respuesta$.error(new Error("502"));
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(component.error).toContain("No se pudieron cargar los indicadores");
    expect(el.querySelector("app-error-state")).not.toBeNull();
    expect(component.puedeExportar).toBeFalse();

    respuesta$ = new Subject<DashboardData>();
    component.reintentar();
    expect(dashboard.cargar).toHaveBeenCalledTimes(2);
    respuesta$.next(vacio);
    fixture.detectChanges();
    expect(component.error).toBe("");
    expect(el.querySelectorAll("app-kpi-card").length).toBe(7);
  });

  it("si las áreas no cargan, deshabilita ese filtro", () => {
    dashboard.areas.and.returnValue(throwError(() => new Error("x")));
    const f2 = TestBed.createComponent(DashboardPageComponent);
    f2.detectChanges();
    expect(f2.componentInstance.areasNoDisponibles).toBeTrue();
  });
});

describe("KpiCardComponent", () => {
  const ind: Indicador = {
    id: "insercionesLaborales",
    titulo: "Inserciones laborales",
    descripcion: "desc",
    modulo: "insercion",
    fuente: "insercion",
    valor: 7,
    disponible: true,
  };

  it("muestra valor, o 'No disponible' con el servicio que falló", () => {
    const f = TestBed.createComponent(KpiCardComponent);
    f.componentRef.setInput("indicador", ind);
    f.detectChanges();
    expect(f.nativeElement.querySelector(".kpi-valor").textContent).toContain(
      "7",
    );

    f.componentRef.setInput("indicador", {
      ...ind,
      valor: null,
      disponible: false,
    });
    f.detectChanges();
    expect(f.nativeElement.textContent).toContain("No disponible");
    expect(f.nativeElement.textContent).toContain("inserción laboral");

    for (const [fuente, texto] of [
      ["autenticacion", "usuarios"],
      ["educacion", "cursos"],
    ] as const) {
      f.componentRef.setInput("indicador", {
        ...ind,
        fuente,
        disponible: false,
      });
      expect(f.componentInstance.nombreFuente).toBe(texto);
    }
  });
});

describe("GraficaEvolucionComponent", () => {
  const grafica: Grafica = {
    id: "inserciones",
    titulo: "Inserciones laborales por mes",
    tipo: "line",
    modulo: "insercion",
    indicadores: ["insercionesLaborales"],
    series: [
      {
        nombre: "Inserciones",
        puntos: [
          { mes: "2026-01", valor: 1 },
          { mes: "2026-02", valor: 4 },
          { mes: "2026-03", valor: 2 },
        ],
      },
    ],
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [GraficaEvolucionComponent],
      providers: [provideNoopAnimations()],
    }).compileComponents();
  });

  it("resume la serie en texto (alternativa a la gráfica)", () => {
    const r = resumirSeries(grafica);
    expect(r).toContain("1 en ene 2026");
    expect(r).toContain("2 en mar 2026");
    expect(r).toContain("máximo de 4 en feb 2026");
    expect(resumirSeries({ ...grafica, series: null })).toBe("");
    expect(
      resumirSeries({ ...grafica, series: [{ nombre: "X", puntos: [] }] }),
    ).toBe("X: sin datos.");
  });

  it("dibuja la gráfica con la tabla de datos y la exporta", () => {
    const f = TestBed.createComponent(GraficaEvolucionComponent);
    f.componentRef.setInput("grafica", grafica);
    f.detectChanges();
    const el: HTMLElement = f.nativeElement;
    expect(el.querySelector("p-chart")).not.toBeNull();
    expect(el.querySelectorAll("tbody tr").length).toBe(3);
    const pdf = f.componentInstance.paraPdf();
    expect(pdf.titulo).toBe(grafica.titulo);
    expect(pdf.resumen).toContain("máximo de 4");
  });

  it("estados: no disponible y sin datos", () => {
    const f = TestBed.createComponent(GraficaEvolucionComponent);
    f.componentRef.setInput("grafica", { ...grafica, series: null });
    f.detectChanges();
    expect(f.nativeElement.textContent).toContain("No disponible");
    expect(f.componentInstance.paraPdf().imagen).toBeNull();

    f.componentRef.setInput("grafica", {
      ...grafica,
      series: [{ nombre: "X", puntos: [{ mes: "2026-01", valor: 0 }] }],
    });
    f.detectChanges();
    expect(f.nativeElement.textContent).toContain(
      "No hay datos para el período seleccionado",
    );
    expect(f.componentInstance.paraPdf().resumen).toContain("No hay datos");
  });
});

describe("HistorialReportesComponent", () => {
  it("lista los reportes y permite volver a descargarlos", async () => {
    const reporte: ReporteExportado = {
      id: "r1",
      tipo_reporte: "dashboard",
      periodo_desde: "2026-01-01",
      periodo_hasta: "2026-06-30",
      filtros_aplicados: {
        modulo: "educacion",
        areaId: "a1",
        areaNombre: "Oficios",
      },
      generado_por: "admin-1",
      generado_por_nombre: "Sofía",
      fecha_generacion: "2026-10-08T15:00:00.000Z",
      url_archivo: "blob:x",
    };
    const reportes = jasmine.createSpyObj("ReportesService", ["descargar"], {
      historial: () => [reporte],
      cargandoHistorial: () => false,
      errorHistorial: () => null,
    });
    reportes.descargar.and.resolveTo();
    await TestBed.configureTestingModule({
      imports: [HistorialReportesComponent],
      providers: [
        provideNoopAnimations(),
        { provide: ReportesService, useValue: reportes },
      ],
    }).compileComponents();
    const f = TestBed.createComponent(HistorialReportesComponent);
    f.detectChanges();
    const el: HTMLElement = f.nativeElement;
    expect(el.textContent).toContain("01/01/2026 al 30/06/2026");
    expect(el.textContent).toContain("Módulo: Educación · Área: Oficios");
    expect(el.textContent).toContain("Sofía");
    (el.querySelector("p-button button") as HTMLButtonElement).click();
    expect(reportes.descargar).toHaveBeenCalledWith(reporte);
  });
});
