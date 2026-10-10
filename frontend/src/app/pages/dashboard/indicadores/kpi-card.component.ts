import { Component, Input } from "@angular/core";
import { SkeletonModule } from "primeng/skeleton";
import { Indicador } from "../../../core/models/dashboard.model";

/** Tarjeta de un indicador (HU-52). */
@Component({
  selector: "app-kpi-card",
  standalone: true,
  imports: [SkeletonModule],
  template: `
    <article
      class="kpi"
      [class.no-disponible]="indicador && !indicador.disponible"
    >
      @if (loading || !indicador) {
        <p-skeleton width="60%" height="1rem"></p-skeleton>
        <p-skeleton
          width="40%"
          height="2.25rem"
          styleClass="kpi-sk-valor"
        ></p-skeleton>
        <p-skeleton width="90%" height="0.875rem"></p-skeleton>
      } @else {
        <h3 class="kpi-titulo">{{ indicador.titulo }}</h3>
        @if (indicador.disponible) {
          <p class="kpi-valor">{{ indicador.valor }}</p>
          <p class="kpi-desc">{{ indicador.descripcion }}</p>
        } @else {
          <p class="kpi-valor kpi-valor-nd">
            <i class="pi pi-exclamation-circle" aria-hidden="true"></i>
            No disponible
          </p>
          <p class="kpi-desc">
            El servicio de {{ nombreFuente }} no respondió. Probá actualizar en
            unos minutos.
          </p>
        }
      }
    </article>
  `,
  styles: [
    `
      :host {
        display: block;
      }
      .kpi {
        display: flex;
        flex-direction: column;
        gap: var(--space-2);
        height: 100%;
        padding: var(--space-5);
        background: var(--color-surface);
        border: 1px solid var(--color-border);
        border-radius: var(--radius-xl);
        box-shadow: var(--shadow-xs);
      }
      .kpi-titulo {
        margin: 0;
        font-size: var(--font-size-sm);
        font-weight: var(--font-weight-medium);
        color: var(--color-text-muted);
        letter-spacing: 0;
      }
      .kpi-valor {
        margin: 0;
        font-size: var(--font-size-3xl);
        font-weight: var(--font-weight-semibold);
        line-height: var(--line-height-tight);
        color: var(--color-text);
        font-variant-numeric: tabular-nums;
      }
      .kpi-valor-nd {
        display: flex;
        align-items: center;
        gap: var(--space-2);
        font-size: var(--font-size-lg);
        color: var(--color-warning);
      }
      .kpi-desc {
        margin: 0;
        font-size: var(--font-size-xs);
        color: var(--color-text-muted);
      }
      .no-disponible {
        background: var(--color-warning-soft);
      }
      :host ::ng-deep .kpi-sk-valor {
        margin: var(--space-1) 0;
      }
    `,
  ],
})
export class KpiCardComponent {
  @Input() indicador: Indicador | null = null;
  @Input() loading = false;

  get nombreFuente(): string {
    switch (this.indicador?.fuente) {
      case "autenticacion":
        return "usuarios";
      case "educacion":
        return "cursos";
      default:
        return "inserción laboral";
    }
  }
}
