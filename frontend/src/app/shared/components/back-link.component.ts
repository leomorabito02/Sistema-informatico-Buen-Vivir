import { Component, EventEmitter, Input, Output } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterLink } from "@angular/router";
import { LucideAngularModule, ArrowLeft } from "lucide-angular";

/**
 * Botón "Volver" único para toda la app.
 *
 * - Con `to`: es un link a esa ruta (destino explícito y predecible).
 * - Sin `to`: es un botón que emite `back`, para pantallas que deciden el
 *   destino en su propio código (por ejemplo, según el rol).
 *
 * El texto siempre arranca con "Volver" y nombra el destino real
 * ("Volver a mis cursos"), nunca "Volver al dashboard".
 */
@Component({
  selector: "app-back-link",
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule],
  template: `
    <a *ngIf="to; else boton" class="back-link" [routerLink]="to">
      <lucide-icon
        [img]="ArrowLeft"
        [size]="18"
        aria-hidden="true"
      ></lucide-icon>
      <span>{{ label }}</span>
    </a>
    <ng-template #boton>
      <button type="button" class="back-link" (click)="back.emit()">
        <lucide-icon
          [img]="ArrowLeft"
          [size]="18"
          aria-hidden="true"
        ></lucide-icon>
        <span>{{ label }}</span>
      </button>
    </ng-template>
  `,
  styles: [
    `
      :host {
        display: inline-block;
      }
      .back-link {
        display: inline-flex;
        align-items: center;
        gap: var(--space-2);
        min-height: var(--tap-target);
        padding: 0 var(--space-3) 0 var(--space-2);
        margin-left: calc(var(--space-2) * -1);
        border: none;
        border-radius: var(--radius-md);
        background: transparent;
        color: var(--color-link);
        font: inherit;
        font-size: var(--font-size-sm);
        font-weight: var(--font-weight-medium);
        text-decoration: none;
        cursor: pointer;
        transition: var(--transition-interactive);
      }
      .back-link:hover {
        background: var(--color-primary-soft);
      }
      .back-link:hover lucide-icon {
        transform: translateX(-2px);
      }
      lucide-icon {
        display: inline-flex;
        transition: transform var(--duration-fast) var(--ease-standard);
      }
    `,
  ],
})
export class BackLinkComponent {
  /** Ruta de destino. Si no se pasa, el componente emite `back`. */
  @Input() to: string | null = null;
  @Input() label = "Volver";
  @Output() back = new EventEmitter<void>();

  readonly ArrowLeft = ArrowLeft;
}
