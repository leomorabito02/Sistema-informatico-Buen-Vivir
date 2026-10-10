import { Injectable, signal } from "@angular/core";
import { Observable, Subject, take } from "rxjs";

/**
 * Pregunta "¿Seguir conectado?" cuando la sesión vence.
 *
 * El interceptor llama a ask() y espera la respuesta; el diálogo
 * (SessionExpiredDialogComponent, montado en AppComponent) se muestra
 * mientras `visible` es true y responde con answer(). Reemplaza al
 * confirm() nativo del navegador para que se vea como el resto de la app.
 */
@Injectable({ providedIn: "root" })
export class SessionPromptService {
  readonly visible = signal(false);
  private pending: Subject<boolean> | null = null;

  ask(): Observable<boolean> {
    if (!this.pending) {
      this.pending = new Subject<boolean>();
      this.visible.set(true);
    }
    return this.pending.pipe(take(1));
  }

  answer(extend: boolean): void {
    const pending = this.pending;
    this.pending = null;
    this.visible.set(false);
    pending?.next(extend);
    pending?.complete();
  }
}
