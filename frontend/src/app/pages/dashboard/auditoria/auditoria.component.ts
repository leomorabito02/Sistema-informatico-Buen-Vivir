import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { Observable } from "rxjs";
import { ButtonModule } from "primeng/button";
import { InputTextModule } from "primeng/inputtext";
import { SelectModule } from "primeng/select";
import { TableModule } from "primeng/table";
import { TagModule } from "primeng/tag";

import { EducationService } from "../../../services/education.service";
import { HumanPromotionService } from "../../../services/human-promotion.service";
import {
  FiltroAuditoria,
  RegistroAuditoria,
} from "../../../core/models/education.model";
import { formatYMD, todayYMD } from "../../../shared/utils/date-ymd";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";

type Modulo = "promocion" | "educacion";

interface Option<T = string> {
  label: string;
  value: T;
}

/** Cómo se nombra cada acción y cada recurso en pantalla. */
const ACCIONES: Record<string, string> = {
  crear: "Alta",
  registrar: "Carga",
  actualizar: "Edición",
  baja: "Baja",
};
const RECURSOS: Record<string, string> = {
  ficha: "Ficha personal",
  hito: "Hito",
  curso: "Curso",
  asistencia_general: "Asistencia general",
};

/**
 * Registro de auditoría, solo lectura y solo administración (HU-30 y
 * HU-35 en Promoción Humana; HU-16 y HU-19bis en Educación): quién dio de
 * alta, editó o dio de baja algo, cuándo y desde qué IP. Nunca muestra el
 * contenido (las fichas son confidenciales). Cada servicio guarda su
 * propio registro, así que se elige el módulo.
 */
@Component({
  selector: "app-auditoria",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ButtonModule,
    InputTextModule,
    SelectModule,
    TableModule,
    TagModule,
    LoadingStateComponent,
    ErrorStateComponent,
  ],
  templateUrl: "./auditoria.component.html",
  styleUrls: ["./auditoria.component.css"],
})
export class AuditoriaComponent implements OnInit {
  private readonly education = inject(EducationService);
  private readonly promocion = inject(HumanPromotionService);

  readonly modulos: Option<Modulo>[] = [
    { label: "Promoción Humana (fichas e hitos)", value: "promocion" },
    { label: "Educación (cursos y asistencia)", value: "educacion" },
  ];
  modulo: Modulo = "promocion";
  desde = formatYMD(haceDias(29));
  hasta = todayYMD();
  accion = "";

  readonly acciones: Option[] = [
    { label: "Todas las acciones", value: "" },
    ...Object.entries(ACCIONES).map(([value, label]) => ({ label, value })),
  ];

  loading = false;
  error = "";
  registros: RegistroAuditoria[] = [];

  ngOnInit() {
    this.buscar();
  }

  buscar() {
    const filtro: FiltroAuditoria = {
      desde: this.desde,
      hasta: this.hasta,
      accion: this.accion,
    };
    const consulta$: Observable<RegistroAuditoria[]> =
      this.modulo === "educacion"
        ? this.education.listAuditoria(filtro)
        : this.promocion.listarAuditoria(filtro);
    this.loading = true;
    this.error = "";
    consulta$.subscribe({
      next: (rows) => {
        this.registros = rows;
        this.loading = false;
      },
      error: (err: Error) => {
        this.registros = [];
        this.error = err.message;
        this.loading = false;
      },
    });
  }

  accionLabel(a: string): string {
    return ACCIONES[a] ?? a;
  }

  recursoLabel(r: string): string {
    return RECURSOS[r] ?? r;
  }

  accionSeverity(a: string): "success" | "info" | "danger" | "secondary" {
    switch (a) {
      case "crear":
      case "registrar":
        return "success";
      case "actualizar":
        return "info";
      case "baja":
        return "danger";
      default:
        return "secondary";
    }
  }
}

function haceDias(n: number): Date {
  const d = new Date();
  d.setDate(d.getDate() - n);
  return d;
}
