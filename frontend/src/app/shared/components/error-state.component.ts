import { Component, EventEmitter, Input, Output } from "@angular/core";
import { LucideAngularModule, CircleAlert, RotateCw } from "lucide-angular";

/**
 * Estado de error claro: qué pasó (título) y el detalle que dio la
 * pantalla (message). Si la pantalla escucha `(retry)`, aparece el botón
 * "Reintentar". Se anuncia a lectores de pantalla (role="alert").
 */
@Component({
  selector: "app-error-state",
  standalone: true,
  imports: [LucideAngularModule],
  template: `
    <div class="error-state" role="alert">
      <lucide-icon
        [img]="CircleAlert"
        [size]="22"
        class="err-icon"
        aria-hidden="true"
      ></lucide-icon>
      <div class="err-body">
        <p class="err-title">{{ title }}</p>
        @if (message) {
          <p class="err-message">{{ message }}</p>
        }
      </div>
      @if (retry.observed) {
        <button type="button" class="err-retry" (click)="retry.emit()">
          <lucide-icon
            [img]="RotateCw"
            [size]="16"
            aria-hidden="true"
          ></lucide-icon>
          Reintentar
        </button>
      }
    </div>
  `,
  styles: [
    `
      :host {
        display: block;
        margin: var(--space-4) 0;
      }
      .error-state {
        display: flex;
        align-items: flex-start;
        gap: var(--space-3);
        padding: var(--space-4);
        background: var(--color-danger-soft);
        border: 1px solid var(--red-200);
        border-left: 4px solid var(--color-danger);
        border-radius: var(--radius-md);
        color: var(--red-800);
      }
      .err-icon {
        flex-shrink: 0;
        margin-top: 2px;
        color: var(--color-danger);
      }
      .err-body {
        flex: 1;
        min-width: 0;
      }
      .err-title {
        margin: 0;
        font-weight: var(--font-weight-semibold);
      }
      .err-message {
        margin: var(--space-1) 0 0;
        font-size: var(--font-size-sm);
        overflow-wrap: anywhere;
      }
      .err-retry {
        display: inline-flex;
        align-items: center;
        gap: var(--space-2);
        min-height: var(--tap-target);
        padding: 0 var(--space-4);
        border: 1px solid var(--color-danger);
        border-radius: var(--radius-md);
        background: var(--color-surface);
        color: var(--color-danger);
        font: inherit;
        font-size: var(--font-size-sm);
        font-weight: var(--font-weight-medium);
        cursor: pointer;
        transition: var(--transition-interactive);
      }
      .err-retry:hover {
        background: var(--red-100);
      }
    `,
  ],
})
export class ErrorStateComponent {
  @Input() title = "No pudimos cargar la información";
  @Input() message = "";
  @Output() retry = new EventEmitter<void>();

  readonly CircleAlert = CircleAlert;
  readonly RotateCw = RotateCw;
}
