import { ComponentFixture, TestBed } from "@angular/core/testing";
import { provideNoopAnimations } from "@angular/platform-browser/animations";
import {
  DashboardFiltrosComponent,
  filtrosPorDefecto,
  rangoRapido,
} from "./dashboard-filtros.component";
import { FiltrosDashboard } from "../../../core/models/dashboard.model";

describe("DashboardFiltrosComponent", () => {
  let fixture: ComponentFixture<DashboardFiltrosComponent>;
  let component: DashboardFiltrosComponent;
  let emitidos: FiltrosDashboard[];
  const hoy = new Date(2026, 9, 8);

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DashboardFiltrosComponent],
      providers: [provideNoopAnimations()],
    }).compileComponents();
    fixture = TestBed.createComponent(DashboardFiltrosComponent);
    component = fixture.componentInstance;
    (component as { hoy: Date }).hoy = hoy;
    component.areas = [
      { id: "area-huerta", nombre: "Huerta y agroecología" },
      { id: "area-oficios", nombre: "Oficios" },
    ];
    fixture.componentRef.setInput("filtros", filtrosPorDefecto(hoy));
    emitidos = [];
    component.filtrosChange.subscribe((f) => emitidos.push(f));
    fixture.detectChanges();
  });

  const ultimo = () => emitidos[emitidos.length - 1];

  it("los accesos rápidos terminan hoy y abarcan 1, 3 y 12 meses", () => {
    const mes = rangoRapido("mes", hoy);
    expect(mes.hasta).toEqual(new Date(2026, 9, 8));
    expect(mes.desde).toEqual(new Date(2026, 8, 9));
    expect(rangoRapido("trimestre", hoy).desde).toEqual(new Date(2026, 6, 9));
    expect(rangoRapido("anio", hoy).desde).toEqual(new Date(2025, 9, 9));
  });

  it("por defecto: último año, todos los módulos y todas las áreas", () => {
    const def = filtrosPorDefecto(hoy);
    expect(def.modulo).toBe("todos");
    expect(def.areaId).toBeNull();
    expect(component.rangoActivo("anio")).toBeTrue();
    expect(component.rangoActivo("mes")).toBeFalse();
  });

  it("aplicar un acceso rápido emite el nuevo período", () => {
    component.aplicarRango("trimestre");
    expect(ultimo().desde).toEqual(new Date(2026, 6, 9));
    expect(ultimo().hasta).toEqual(hoy);
  });

  it("cambiar módulo o área emite el filtro completo", () => {
    component.modulo = "educacion";
    component.areaId = "area-oficios";
    component.emitir();
    expect(ultimo()).toEqual(
      jasmine.objectContaining({ modulo: "educacion", areaId: "area-oficios" }),
    );
  });

  it("no emite si 'Desde' es posterior a 'Hasta' y muestra el error", () => {
    component.desde = new Date(2026, 5, 1);
    component.hasta = new Date(2026, 0, 1);
    component.emitir();
    fixture.detectChanges();
    expect(emitidos.length).toBe(0);
    expect(component.errorPeriodo).toContain("anterior");
    expect(fixture.nativeElement.querySelector("[role=alert]")).not.toBeNull();
  });

  it("no emite si falta una fecha", () => {
    component.hasta = null;
    component.emitir();
    expect(emitidos.length).toBe(0);
    expect(component.errorPeriodo).toContain("dos fechas");
  });

  it("muestra los filtros activos como chips y los quita de a uno", () => {
    fixture.componentRef.setInput("filtros", {
      desde: new Date(2026, 0, 1),
      hasta: new Date(2026, 5, 30),
      modulo: "insercion",
      areaId: "area-huerta",
    } satisfies FiltrosDashboard);
    fixture.detectChanges();

    const chips = component.chips;
    expect(chips.map((c) => c.texto)).toEqual([
      "01/01/2026 al 30/06/2026",
      "Módulo: Inserción Laboral",
      "Área: Huerta y agroecología",
    ]);
    expect(chips.every((c) => c.quitable)).toBeTrue();
    expect(fixture.nativeElement.querySelectorAll("p-chip").length).toBe(3);

    component.quitar(chips[2]);
    expect(ultimo().areaId).toBeNull();
    expect(ultimo().modulo).toBe("insercion");

    component.quitar(chips[1]);
    expect(ultimo().modulo).toBe("todos");

    component.quitar(chips[0]);
    expect(ultimo().desde).toEqual(filtrosPorDefecto(hoy).desde);
  });

  it("con los filtros por defecto el período no se puede quitar", () => {
    expect(component.chips.length).toBe(1);
    expect(component.chips[0].quitable).toBeFalse();
    expect(component.hayFiltrosQuitables).toBeFalse();
  });

  it("'Limpiar filtros' vuelve a los valores por defecto", () => {
    component.modulo = "usuarios";
    component.areaId = "area-oficios";
    component.limpiar();
    expect(ultimo()).toEqual(filtrosPorDefecto(hoy));
  });

  it("las opciones de área incluyen 'Todas' y se deshabilitan si no cargaron", () => {
    expect(component.opcionesArea[0]).toEqual({
      label: "Todas las áreas",
      value: null,
    });
    expect(component.opcionesArea.length).toBe(3);
    fixture.componentRef.setInput("areasNoDisponibles", true);
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain(
      "No se pudieron cargar las áreas",
    );
  });

  it("cada campo tiene su etiqueta", () => {
    const el: HTMLElement = fixture.nativeElement;
    for (const id of [
      "filtroDesde",
      "filtroHasta",
      "filtroModulo",
      "filtroArea",
    ]) {
      expect(el.querySelector(`label[for=${id}]`))
        .withContext(id)
        .not.toBeNull();
    }
  });
});
