import { Component, Input } from "@angular/core";
import { LucideAngularModule, Inbox } from "lucide-angular";

/**
 * Estado vacío: ícono de línea, título opcional y un mensaje útil (lo que
 * va entre las etiquetas). Las acciones van en un elemento con el
 * atributo `actions`:
 *
 *   <app-empty-state title="Sin cursos">
 *     Todavía no hay cursos publicados.
 *     <p-button actions label="Crear curso" />
 *   </app-empty-state>
 *
 * `compact` es para espacios chicos (debajo de un selector, en un form).
 */
@Component({
  selector: "app-empty-state",
  standalone: true,
  imports: [LucideAngularModule],
  template: `
    <div class="empty-state-box" [class.compact]="compact">
      @if (!compact) {
        <span class="es-icon" aria-hidden="true">
          <lucide-icon [img]="icon" [size]="28"></lucide-icon>
        </span>
      }
      @if (title) {
        <p class="es-title">{{ title }}</p>
      }
      <p class="es-message"><ng-content></ng-content></p>
      <div class="es-actions"><ng-content select="[actions]"></ng-content></div>
    </div>
  `,
  styles: [
    `
      :host {
        display: block;
      }
      .empty-state-box {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: var(--space-2);
        padding: var(--space-10) var(--space-6);
        text-align: center;
        background: var(--color-surface);
        border: 1px dashed var(--color-border-strong);
        border-radius: var(--radius-lg);
      }
      .es-icon {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 3.5rem;
        height: 3.5rem;
        margin-bottom: var(--space-2);
        border-radius: var(--radius-full);
        background: var(--color-primary-soft);
        color: var(--color-primary);
      }
      .es-title {
        margin: 0;
        font-size: var(--font-size-lg);
        font-weight: var(--font-weight-semibold);
        color: var(--color-text);
      }
      .es-message {
        margin: 0;
        max-width: 48ch;
        color: var(--color-text-muted);
      }
      .es-actions:empty {
        display: none;
      }
      .es-actions {
        margin-top: var(--space-3);
      }
      .empty-state-box.compact {
        align-items: flex-start;
        padding: var(--space-3) 0;
        text-align: left;
        background: none;
        border: none;
      }
      .compact .es-message {
        font-size: var(--font-size-sm);
      }
    `,
  ],
})
export class EmptyStateComponent {
  @Input() title = "";
  @Input() icon: typeof Inbox = Inbox;
  @Input() compact = false;
}
