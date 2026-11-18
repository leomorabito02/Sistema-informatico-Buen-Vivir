import { Component, inject } from "@angular/core";
import { DialogModule } from "primeng/dialog";
import { ButtonModule } from "primeng/button";
import { SessionPromptService } from "../../../services/session-prompt.service";

/** Diálogo de sesión vencida (ver SessionPromptService). */
@Component({
  selector: "app-session-expired-dialog",
  standalone: true,
  imports: [DialogModule, ButtonModule],
  template: `
    <p-dialog
      header="Tu sesión venció"
      [visible]="prompt.visible()"
      [modal]="true"
      [closable]="false"
      [draggable]="false"
      [style]="{ width: '420px' }"
      [breakpoints]="{ '640px': '92vw' }"
    >
      <p class="session-text">
        Por seguridad, la sesión se cierra después de un rato sin actividad.
        ¿Querés seguir conectado/a? Si salís, lo que no hayas guardado se
        pierde.
      </p>
      <ng-template #footer>
        <p-button
          label="Cerrar sesión"
          severity="secondary"
          [outlined]="true"
          (onClick)="prompt.answer(false)"
        ></p-button>
        <p-button
          label="Seguir conectado/a"
          (onClick)="prompt.answer(true)"
        ></p-button>
      </ng-template>
    </p-dialog>
  `,
  styles: [
    `
      .session-text {
        margin: 0;
      }
    `,
  ],
})
export class SessionExpiredDialogComponent {
  readonly prompt = inject(SessionPromptService);
}
