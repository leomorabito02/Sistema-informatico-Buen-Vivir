import { Component, Input } from "@angular/core";
import { SkeletonModule } from "primeng/skeleton";

export type LoadingVariant = "list" | "cards" | "table" | "detail";

/**
 * Estado de carga con skeletons (PrimeNG) que imitan la forma del
 * contenido que va a aparecer: así la pantalla no "salta" al cargar.
 *
 * - list:   filas con dos líneas (listas, buscadores, línea de tiempo)
 * - cards:  grilla de tarjetas (catálogo de cursos)
 * - table:  encabezado + filas (tablas y matrices)
 * - detail: título, párrafo y bloque (pantallas de detalle)
 *
 * Lectores de pantalla: anuncia `label` una sola vez (role="status").
 */
@Component({
  selector: "app-loading-state",
  standalone: true,
  imports: [SkeletonModule],
  template: `
    <div class="loading-state" role="status" aria-live="polite">
      <span class="sr-only">{{ label }}</span>
      @switch (variant) {
        @case ("cards") {
          <div class="ls-cards" aria-hidden="true">
            @for (i of filas; track i) {
              <div class="ls-card">
                <p-skeleton height="8rem" borderRadius="var(--radius-md)" />
                <p-skeleton width="75%" height="1.25rem" />
                <p-skeleton width="95%" />
                <p-skeleton width="45%" />
              </div>
            }
          </div>
        }
        @case ("table") {
          <div class="ls-table" aria-hidden="true">
            <p-skeleton height="2.5rem" borderRadius="var(--radius-sm)" />
            @for (i of filas; track i) {
              <div class="ls-row">
                <p-skeleton width="30%" />
                <p-skeleton width="25%" />
                <p-skeleton width="20%" />
              </div>
            }
          </div>
        }
        @case ("detail") {
          <div class="ls-detail" aria-hidden="true">
            <p-skeleton width="55%" height="2rem" />
            <p-skeleton width="90%" />
            <p-skeleton width="80%" />
            <p-skeleton height="10rem" borderRadius="var(--radius-lg)" />
          </div>
        }
        @default {
          <div class="ls-list" aria-hidden="true">
            @for (i of filas; track i) {
              <div class="ls-item">
                <p-skeleton shape="circle" size="2.5rem" />
                <div class="ls-lines">
                  <p-skeleton width="60%" />
                  <p-skeleton width="35%" height="0.75rem" />
                </div>
              </div>
            }
          </div>
        }
      }
    </div>
  `,
  styles: [
    `
      :host {
        display: block;
      }
      .loading-state {
        padding: var(--space-4) 0;
        --p-skeleton-background: var(--color-surface-alt);
        --p-skeleton-animation-background: color-mix(
          in srgb,
          var(--color-surface) 60%,
          transparent
        );
      }
      .ls-cards {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(16rem, 1fr));
        gap: var(--space-6);
      }
      .ls-card {
        display: flex;
        flex-direction: column;
        gap: var(--space-3);
        padding: var(--space-4);
        background: var(--color-surface);
        border: 1px solid var(--color-border);
        border-radius: var(--radius-lg);
      }
      .ls-table,
      .ls-detail,
      .ls-list {
        display: flex;
        flex-direction: column;
        gap: var(--space-4);
      }
      .ls-row {
        display: grid;
        grid-template-columns: 2fr 1.5fr 1fr;
        gap: var(--space-6);
        padding: var(--space-2) var(--space-3);
      }
      .ls-item {
        display: flex;
        align-items: center;
        gap: var(--space-3);
      }
      .ls-lines {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: var(--space-2);
      }
    `,
  ],
})
export class LoadingStateComponent {
  @Input() variant: LoadingVariant = "list";
  /** Cantidad de filas o tarjetas de muestra. */
  @Input() rows = 3;
  /** Texto para lectores de pantalla. */
  @Input() label = "Cargando…";

  get filas(): number[] {
    return Array.from({ length: Math.max(1, this.rows) }, (_, i) => i);
  }
}
