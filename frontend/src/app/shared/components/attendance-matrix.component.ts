import { Component, Input } from "@angular/core";
import { CommonModule } from "@angular/common";
import { TooltipModule } from "primeng/tooltip";

import {
  Activity,
  Attendance,
  Enrollment,
} from "../../core/models/education.model";

export interface AttendanceMatrixRow {
  participanteId: string;
  nombre: string;
  /** ids de las actividades con asistencia presente (= entregadas). */
  presentes: Set<string>;
}

/**
 * Arma las filas de la matriz: los inscriptos activos del curso más
 * cualquiera que tenga asistencia registrada (por ejemplo, alguien que
 * entregó y después se dio de baja), ordenados por nombre.
 */
export function buildAttendanceMatrixRows(
  enrollments: Enrollment[],
  attendance: Attendance[],
  nombres: Map<string, string>,
): AttendanceMatrixRow[] {
  const rows = new Map<string, AttendanceMatrixRow>();
  const fila = (id: string) => {
    let r = rows.get(id);
    if (!r) {
      r = {
        participanteId: id,
        nombre: nombres.get(id) ?? id,
        presentes: new Set(),
      };
      rows.set(id, r);
    }
    return r;
  };
  enrollments
    .filter((e) => e.estado === "activa")
    .forEach((e) => fila(e.participanteId));
  attendance
    .filter((a) => a.presente)
    .forEach((a) => fila(a.participanteId).presentes.add(a.actividadId));
  return [...rows.values()].sort((a, b) => a.nombre.localeCompare(b.nombre));
}

/**
 * Historial de asistencia de un curso (HU-19), solo lectura: una fila
 * por participante, una columna por actividad (en orden de fecha) y el
 * porcentaje de actividades cumplidas (HU-20). La asistencia no se
 * carga a mano: cada entrega de una actividad la marca como presente.
 */
@Component({
  selector: "app-attendance-matrix",
  standalone: true,
  imports: [CommonModule, TooltipModule],
  template: `
    <div
      class="matrix-scroll"
      *ngIf="rows.length && activities.length; else vacio"
    >
      <table class="matrix">
        <thead>
          <tr>
            <th class="sticky">Participante</th>
            <th
              *ngFor="let a of activities"
              class="act-col"
              [pTooltip]="a.titulo"
              tooltipPosition="top"
            >
              <span class="act-date">{{
                a.fecha | date: "dd/MM/yyyy" : "UTC"
              }}</span>
              <span class="act-title">{{ a.titulo }}</span>
            </th>
            <th>Progreso</th>
          </tr>
        </thead>
        <tbody>
          <tr *ngFor="let r of rows">
            <td class="sticky">{{ r.nombre }}</td>
            <td *ngFor="let a of activities" class="cell">
              <i
                *ngIf="r.presentes.has(a.id); else ausente"
                class="pi pi-check-circle presente"
                [attr.aria-label]="'Entregó ' + a.titulo"
              ></i>
              <ng-template #ausente>
                <i
                  class="pi pi-minus ausente"
                  [attr.aria-label]="'No entregó ' + a.titulo"
                ></i>
              </ng-template>
            </td>
            <td class="pct">{{ porcentaje(r) | number: "1.0-0" }}%</td>
          </tr>
        </tbody>
      </table>
    </div>
    <ng-template #vacio>
      <p class="empty">
        {{
          activities.length === 0
            ? "Este curso todavía no tiene actividades."
            : "No hay participantes inscriptos en este curso."
        }}
      </p>
    </ng-template>
  `,
  styles: [
    `
      .matrix-scroll {
        overflow-x: auto;
        border: 1px solid var(--border);
        border-radius: 10px;
      }
      .matrix {
        border-collapse: collapse;
        width: 100%;
        font-size: 0.9rem;
      }
      th,
      td {
        padding: 0.6rem 0.75rem;
        border-bottom: 1px solid var(--border);
        text-align: center;
        white-space: nowrap;
      }
      thead th {
        background: #fafaf6;
        color: var(--primary);
        font-weight: 600;
        vertical-align: bottom;
      }
      .sticky {
        position: sticky;
        left: 0;
        background: var(--bg-white);
        text-align: left;
        z-index: 1;
      }
      thead .sticky {
        background: #fafaf6;
      }
      .act-col {
        max-width: 9rem;
      }
      .act-date {
        display: block;
      }
      .act-title {
        display: block;
        font-weight: 400;
        color: var(--text-muted);
        font-size: 0.8rem;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 9rem;
      }
      .presente {
        color: var(--primary);
      }
      .ausente {
        color: #c7c7bd;
      }
      .pct {
        font-weight: 700;
        color: var(--primary);
      }
      .empty {
        color: var(--text-muted);
        text-align: center;
        padding: 1rem 0;
      }
    `,
  ],
})
export class AttendanceMatrixComponent {
  @Input() activities: Activity[] = [];
  @Input() rows: AttendanceMatrixRow[] = [];

  porcentaje(r: AttendanceMatrixRow): number {
    if (this.activities.length === 0) return 0;
    const cumplidas = this.activities.filter((a) =>
      r.presentes.has(a.id),
    ).length;
    return (cumplidas / this.activities.length) * 100;
  }
}
