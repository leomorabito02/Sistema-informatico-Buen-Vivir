import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { ButtonModule } from "primeng/button";
import { CardModule } from "primeng/card";
import { DatePickerModule } from "primeng/datepicker";
import { InputTextModule } from "primeng/inputtext";
import { ButtonGroupModule } from "primeng/buttongroup";
import { ToastModule } from "primeng/toast";
import { MessageService } from "primeng/api";

import { EducationService } from "../../../services/education.service";
import { formatYMD, parseYMD, todayYMD } from "../../../shared/utils/date-ymd";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";

interface AttendanceRow {
  participanteId: string;
  nombreCompleto: string;
  /** null = todavía sin registrar ese día. */
  presente: boolean | null;
  observacion: string;
}

/**
 * Asistencia general del centro (HU-19bis): control diario, independiente de los
 * cursos. Solo se envían las filas marcadas Presente/Ausente, para no pisar con false a
 * quien todavía no se tocó.
 */
@Component({
  selector: "app-asistencia-general",
  standalone: true,
  imports: [
    LoadingStateComponent,
    ErrorStateComponent,
    CommonModule,
    FormsModule,
    TableModule,
    ButtonModule,
    CardModule,
    DatePickerModule,
    InputTextModule,
    ButtonGroupModule,
    ToastModule,
  ],
  providers: [MessageService],
  templateUrl: "./asistencia-general.component.html",
  styleUrls: ["./asistencia-general.component.css"],
})
export class AsistenciaGeneralComponent implements OnInit {
  private education = inject(EducationService);
  private toast = inject(MessageService);

  selectedDate: Date = parseYMD(todayYMD());
  loading = false;
  error = "";
  guardando = false;

  rows: AttendanceRow[] = [];

  ngOnInit() {
    this.cargar();
  }

  get fecha(): string {
    return formatYMD(this.selectedDate);
  }

  onDateChange() {
    this.cargar();
  }

  cargar() {
    this.loading = true;
    this.error = "";
    this.education.getGeneralAttendance(this.fecha).subscribe({
      next: (list) => {
        this.rows = list
          .map((p) => ({
            participanteId: p.participanteId,
            nombreCompleto:
              `${p.nombre ?? ""} ${p.apellido ?? ""}`.trim() ||
              "Participante sin nombre",
            presente: p.presente,
            observacion: p.observacion ?? "",
          }))
          .sort((a, b) => a.nombreCompleto.localeCompare(b.nombreCompleto));
        this.loading = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loading = false;
      },
    });
  }

  marcar(row: AttendanceRow, presente: boolean) {
    row.presente = presente;
  }

  get pendientesDeGuardar(): number {
    return this.rows.filter((r) => r.presente !== null).length;
  }

  guardar() {
    const payload = this.rows
      .filter((r) => r.presente !== null)
      .map((r) => ({
        participanteId: r.participanteId,
        presente: r.presente as boolean,
        observacion: r.observacion.trim() || undefined,
      }));
    if (payload.length === 0) return;
    this.guardando = true;
    this.error = "";
    this.education.registerGeneralAttendance(this.fecha, payload).subscribe({
      next: () => {
        this.guardando = false;
        this.toast.add({
          severity: "success",
          summary: "Asistencia general guardada",
          detail: `${payload.length} participante(s) el ${this.fecha}`,
        });
        this.cargar();
      },
      error: (err: Error) => {
        this.guardando = false;
        this.toast.add({
          severity: "error",
          summary: "No se pudo guardar",
          detail: err.message,
        });
      },
    });
  }
}
