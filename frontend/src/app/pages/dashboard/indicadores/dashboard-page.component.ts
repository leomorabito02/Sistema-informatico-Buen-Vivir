import {
  Component,
  DestroyRef,
  OnInit,
  QueryList,
  ViewChildren,
  inject,
} from "@angular/core";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import { BehaviorSubject, switchMap, tap } from "rxjs";
import { ButtonModule } from "primeng/button";
import { ToastModule } from "primeng/toast";
import { MessageService } from "primeng/api";
import {
  AreaFormativa,
  DashboardData,
  FiltrosDashboard,
} from "../../../core/models/dashboard.model";
import {
  DashboardService,
  INDICADORES,
} from "../../../services/dashboard.service";
import { ReportesService } from "../../../services/reportes.service";
import { AuthService } from "../../../services/auth.service";
import { BackLinkComponent } from "../../../shared/components/back-link.component";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import { KpiCardComponent } from "./kpi-card.component";
import { GraficaEvolucionComponent } from "./grafica-evolucion.component";
import {
  DashboardFiltrosComponent,
  filtrosPorDefecto,
} from "./dashboard-filtros.component";
import { HistorialReportesComponent } from "./historial-reportes.component";

/**
 * Dashboard y Reportes (HU-52 a HU-55), solo para el administrador.
 * Los filtros recalculan indicadores y gráficas; "Exportar PDF" genera el
 * reporte con lo que se ve y lo suma al historial.
 */
@Component({
  selector: "app-dashboard-page",
  standalone: true,
  imports: [
    ButtonModule,
    ToastModule,
    BackLinkComponent,
    EmptyStateComponent,
    KpiCardComponent,
    GraficaEvolucionComponent,
    DashboardFiltrosComponent,
    HistorialReportesComponent,
  ],
  providers: [MessageService],
  templateUrl: "./dashboard-page.component.html",
  styleUrl: "./dashboard-page.component.css",
})
export class DashboardPageComponent implements OnInit {
  private dashboard = inject(DashboardService);
  private reportes = inject(ReportesService);
  private auth = inject(AuthService);
  private messages = inject(MessageService);
  private destroyRef = inject(DestroyRef);

  @ViewChildren(GraficaEvolucionComponent)
  private graficasVista?: QueryList<GraficaEvolucionComponent>;

  filtros: FiltrosDashboard = filtrosPorDefecto();
  areas: AreaFormativa[] = [];
  areasNoDisponibles = false;

  datos: DashboardData | null = null;
  cargando = true;
  exportando = false;
  private usuario = "Administración";

  /** Para los esqueletos de carga: uno por indicador. */
  readonly placeholders = INDICADORES;

  private filtros$ = new BehaviorSubject<FiltrosDashboard>(this.filtros);

  ngOnInit(): void {
    this.reportes.cargarHistorial();

    this.dashboard
      .areas()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (a) => (this.areas = a),
        error: () => (this.areasNoDisponibles = true),
      });

    this.auth
      .isAuthenticated()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (res) => {
          const u = res.user;
          const nombre = [u?.first_name, u?.last_name]
            .filter(Boolean)
            .join(" ")
            .trim();
          this.usuario = nombre || u?.email || this.usuario;
        },
        error: () => undefined,
      });

    // switchMap: si cambian los filtros mientras carga, gana el último.
    this.filtros$
      .pipe(
        tap(() => (this.cargando = true)),
        switchMap((f) => this.dashboard.cargar(f)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe((d) => {
        this.datos = d;
        this.cargando = false;
      });
  }

  onFiltros(f: FiltrosDashboard): void {
    this.filtros = f;
    this.filtros$.next(f);
  }

  /** El módulo elegido no tiene indicadores (por ejemplo, Promoción Humana). */
  get moduloSinIndicadores(): boolean {
    return (
      !this.cargando && !!this.datos && this.datos.indicadores.length === 0
    );
  }

  get puedeExportar(): boolean {
    return !this.cargando && !this.exportando && !this.moduloSinIndicadores;
  }

  async exportar(): Promise<void> {
    if (!this.datos || !this.puedeExportar) return;
    this.exportando = true;
    try {
      const area = this.areas.find((a) => a.id === this.filtros.areaId);
      const r = await this.reportes.exportar({
        filtros: this.filtros,
        areaNombre: area?.nombre ?? null,
        indicadores: this.datos.indicadores,
        graficas: (this.graficasVista?.toArray() ?? []).map((g) => g.paraPdf()),
        generadoPor: this.usuario,
      });
      this.messages.add(
        r.guardado
          ? {
              severity: "success",
              summary: "Reporte exportado",
              detail: "Se descargó el PDF y quedó en el historial de reportes.",
            }
          : {
              severity: "warn",
              summary: "Reporte descargado",
              detail: `Se descargó el PDF, pero no quedó en el historial: ${r.error}`,
            },
      );
    } catch {
      this.messages.add({
        severity: "error",
        summary: "No se pudo exportar",
        detail: "Volvé a intentar en unos minutos.",
      });
    } finally {
      this.exportando = false;
    }
  }
}
