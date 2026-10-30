import { Component, inject } from "@angular/core";
import { DatePipe } from "@angular/common";
import { TableModule } from "primeng/table";
import { ButtonModule } from "primeng/button";
import { ReporteExportado } from "../../../core/models/dashboard.model";
import { formatoFecha } from "../../../services/dashboard.fechas";
import {
  NOMBRE_MODULO,
  ReportesService,
} from "../../../services/reportes.service";

/** Historial de reportes exportados (HU-55), con descarga de nuevo. */
@Component({
  selector: "app-historial-reportes",
  standalone: true,
  imports: [TableModule, ButtonModule, DatePipe],
  template: `
    <section class="historial" aria-labelledby="historial-titulo">
      <h2 id="historial-titulo" class="historial-titulo">
        Historial de reportes
      </h2>
      @if (reportes.errorHistorial(); as error) {
        <p class="historial-error" role="alert">{{ error }}</p>
      }
      @if (errorDescarga) {
        <p class="historial-error" role="alert">{{ errorDescarga }}</p>
      }
      <p-table
        [value]="reportes.historial()"
        [loading]="reportes.cargandoHistorial()"
        [paginator]="reportes.historial().length > 10"
        [rows]="10"
        responsiveLayout="stack"
        [breakpoint]="'768px'"
        [tableStyle]="{ 'min-width': '40rem' }"
      >
        <ng-template pTemplate="header">
          <tr>
            <th scope="col">Generado</th>
            <th scope="col">Período</th>
            <th scope="col">Filtros</th>
            <th scope="col">Generado por</th>
            <th scope="col"><span class="sr-only">Acciones</span></th>
          </tr>
        </ng-template>
        <ng-template pTemplate="body" let-r>
          <tr>
            <td>{{ r.fecha_generacion | date: "dd/MM/yyyy HH:mm" }}</td>
            <td>{{ periodo(r) }}</td>
            <td>{{ filtros(r) }}</td>
            <td>{{ r.generado_por_nombre || r.generado_por }}</td>
            <td class="accion">
              <p-button
                icon="pi pi-download"
                label="Descargar"
                size="small"
                [outlined]="true"
                [ariaLabel]="'Descargar el reporte del ' + periodo(r)"
                (onClick)="descargar(r)"
              ></p-button>
            </td>
          </tr>
        </ng-template>
        <ng-template pTemplate="emptymessage">
          <tr>
            <td colspan="5" class="empty-state">
              Todavía no se exportaron reportes. Usá «Exportar PDF» para generar
              el primero.
            </td>
          </tr>
        </ng-template>
      </p-table>
    </section>
  `,
  styles: [
    `
      .historial {
        margin-top: var(--space-10);
      }
      .historial-titulo {
        margin: 0 0 var(--space-4);
        font-size: var(--font-size-xl);
      }
      .accion {
        text-align: right;
      }
      .historial-error {
        color: var(--color-danger, #b91c1c);
      }
    `,
  ],
})
export class HistorialReportesComponent {
  readonly reportes = inject(ReportesService);
  errorDescarga = "";

  async descargar(r: ReporteExportado): Promise<void> {
    this.errorDescarga = "";
    try {
      await this.reportes.descargar(r);
    } catch {
      this.errorDescarga =
        "No se pudo descargar el reporte. Probá de nuevo en unos minutos.";
    }
  }

  periodo(r: ReporteExportado): string {
    return `${formatoFecha(soloFecha(r.periodo_desde))} al ${formatoFecha(soloFecha(r.periodo_hasta))}`;
  }

  filtros(r: ReporteExportado): string {
    const f = r.filtros_aplicados;
    return `Módulo: ${NOMBRE_MODULO[f.modulo]} · Área: ${f.areaNombre ?? "Todas"}`;
  }
}

/** "2026-03-01" como fecha local (sin correrse un día por la zona horaria). */
function soloFecha(iso: string): Date {
  const [a, m, d] = iso.slice(0, 10).split("-").map(Number);
  return new Date(a, m - 1, d);
}
