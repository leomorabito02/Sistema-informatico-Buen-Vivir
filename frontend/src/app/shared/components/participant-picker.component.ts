import {
  Component,
  EventEmitter,
  Input,
  Output,
  inject,
} from "@angular/core";
import { FormsModule } from "@angular/forms";
import { AutoCompleteModule } from "primeng/autocomplete";

import { EducationService } from "../../services/education.service";
import { ParticipanteResumen } from "../../core/models/education.model";

/** Resultado del buscador con el texto que muestra el selector. */
export type OpcionParticipante = ParticipanteResumen & { etiqueta: string };

/** "Apellido, Nombre" (o lo que haya) para mostrar en el selector. */
export function etiquetaParticipante(p: {
  nombre?: string;
  apellido?: string;
  participanteId: string;
}): string {
  const partes = [p.apellido, p.nombre].filter((x) => x && x.trim());
  return partes.length ? partes.join(", ") : p.participanteId;
}

/**
 * Selector de participantes que busca en el backend
 * (`GET /api/participantes?q=`): se escriben 2 letras del nombre, el
 * apellido o el DNI y education devuelve hasta 20 participantes
 * aprobados. Con `cursoId` deja afuera a los ya inscriptos en ese curso.
 * Reemplaza a los selectores que bajaban el directorio entero y
 * filtraban en el navegador.
 */
@Component({
  selector: "app-participant-picker",
  standalone: true,
  imports: [FormsModule, AutoCompleteModule],
  template: `
    <p-autoComplete
      [inputId]="inputId"
      [(ngModel)]="valor"
      [suggestions]="sugerencias"
      optionLabel="etiqueta"
      [minLength]="2"
      [delay]="300"
      [forceSelection]="true"
      [showEmptyMessage]="true"
      [emptyMessage]="mensajeVacio"
      [placeholder]="placeholder"
      appendTo="body"
      [style]="{ width: '100%' }"
      [inputStyle]="{ width: '100%' }"
      (completeMethod)="buscar($event)"
      (onSelect)="elegir($event.value)"
      (onClear)="limpiar()"
    ></p-autoComplete>
  `,
  styles: [
    `
      :host {
        display: block;
        width: 100%;
      }
    `,
  ],
})
export class ParticipantPickerComponent {
  private readonly education = inject(EducationService);

  /** Curso para el que se inscribe: oculta a quienes ya están inscriptos. */
  @Input() cursoId = "";
  @Input() placeholder = "Escribí nombre, apellido o DNI";
  @Input() inputId = "participant-picker";
  @Output() readonly seleccionado =
    new EventEmitter<ParticipanteResumen | null>();

  valor: OpcionParticipante | string | null = null;
  sugerencias: OpcionParticipante[] = [];
  mensajeVacio = "No encontramos participantes con ese nombre o DNI";

  buscar(event: { query: string }) {
    this.education.searchParticipants(event.query, this.cursoId).subscribe({
      next: (lista) => {
        this.mensajeVacio = "No encontramos participantes con ese nombre o DNI";
        this.sugerencias = lista.map((p) => ({
          ...p,
          etiqueta: etiquetaParticipante(p),
        }));
      },
      error: (err: Error) => {
        this.mensajeVacio = err.message;
        this.sugerencias = [];
      },
    });
  }

  elegir(p: OpcionParticipante) {
    this.seleccionado.emit(p);
  }

  /** Vacía el selector (por ejemplo, después de inscribir). */
  limpiar() {
    this.valor = null;
    this.sugerencias = [];
    this.seleccionado.emit(null);
  }
}
