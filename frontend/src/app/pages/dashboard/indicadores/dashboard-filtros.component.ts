import {
  Component,
  EventEmitter,
  Input,
  OnChanges,
  Output,
} from "@angular/core";
import { FormsModule } from "@angular/forms";
import { DatePickerModule } from "primeng/datepicker";
import { SelectModule } from "primeng/select";
import { ButtonModule } from "primeng/button";
import { ChipModule } from "primeng/chip";
import {
  AreaFormativa,
  FiltrosDashboard,
  ModuloDashboard,
} from "../../../core/models/dashboard.model";
import {
  formatoFecha,
  inicioDelDia,
} from "../../../services/dashboard.calculos";
import { NOMBRE_MODULO } from "../../../services/reportes.service";

export type RangoRapido = "mes" | "trimestre" | "anio";

/** Período por defecto: el último año hasta hoy. */
export function filtrosPorDefecto(hoy = new Date()): FiltrosDashboard {
  return {
    ...rangoRapido("anio", hoy),
    modulo: "todos",
    areaId: null,
  };
}

/** Último mes / últimos 3 meses / último año, terminando hoy. */
export function rangoRapido(
  rango: RangoRapido,
  hoy = new Date(),
): { desde: Date; hasta: Date } {
  const meses = rango === "mes" ? 1 : rango === "trimestre" ? 3 : 12;
  const hasta = inicioDelDia(hoy);
  const desde = new Date(hasta);
  desde.setMonth(desde.getMonth() - meses);
  desde.setDate(desde.getDate() + 1);
  return { desde, hasta };
}

export interface ChipFiltro {
  clave: "periodo" | "modulo" | "area";
  texto: string;
  /** El período siempre está; se puede volver al de por defecto. */
  quitable: boolean;
}

/**
 * Filtros del dashboard (HU-54): período, módulo y área formativa.
 * Emite el filtro completo cada vez que algo cambia y es válido; el
 * dashboard recalcula indicadores y gráficas con eso.
 */
@Component({
  selector: "app-dashboard-filtros",
  standalone: true,
  imports: [
    FormsModule,
    DatePickerModule,
    SelectModule,
    ButtonModule,
    ChipModule,
  ],
  templateUrl: "./dashboard-filtros.component.html",
  styleUrl: "./dashboard-filtros.component.css",
})
export class DashboardFiltrosComponent implements OnChanges {
  @Input({ required: true }) filtros!: FiltrosDashboard;
  @Input() areas: AreaFormativa[] = [];
  /** Las áreas no se pudieron cargar: el selector queda deshabilitado. */
  @Input() areasNoDisponibles = false;
  @Output() filtrosChange = new EventEmitter<FiltrosDashboard>();

  readonly hoy = new Date();
  readonly modulos: { label: string; value: ModuloDashboard }[] = (
    ["todos", "usuarios", "educacion", "promocion", "insercion"] as const
  ).map((value) => ({ label: NOMBRE_MODULO[value], value }));
  readonly rangos: { label: string; value: RangoRapido }[] = [
    { label: "Último mes", value: "mes" },
    { label: "Últimos 3 meses", value: "trimestre" },
    { label: "Último año", value: "anio" },
  ];

  desde: Date | null = null;
  hasta: Date | null = null;
  modulo: ModuloDashboard = "todos";
  areaId: string | null = null;
  errorPeriodo = "";

  get opcionesArea(): { label: string; value: string | null }[] {
    return [
      { label: "Todas las áreas", value: null },
      ...this.areas.map((a) => ({ label: a.nombre, value: a.id })),
    ];
  }

  ngOnChanges(): void {
    this.desde = this.filtros.desde;
    this.hasta = this.filtros.hasta;
    this.modulo = this.filtros.modulo;
    this.areaId = this.filtros.areaId;
  }

  /** ¿El período actual coincide con un acceso rápido? (para marcarlo) */
  rangoActivo(r: RangoRapido): boolean {
    if (!this.desde || !this.hasta) return false;
    const { desde, hasta } = rangoRapido(r, this.hoy);
    return (
      desde.getTime() === inicioDelDia(this.desde).getTime() &&
      hasta.getTime() === inicioDelDia(this.hasta).getTime()
    );
  }

  aplicarRango(r: RangoRapido): void {
    const { desde, hasta } = rangoRapido(r, this.hoy);
    this.desde = desde;
    this.hasta = hasta;
    this.emitir();
  }

  emitir(): void {
    if (!this.desde || !this.hasta) {
      this.errorPeriodo = "Elegí las dos fechas del período.";
      return;
    }
    if (this.desde > this.hasta) {
      this.errorPeriodo = "La fecha «Desde» tiene que ser anterior a «Hasta».";
      return;
    }
    this.errorPeriodo = "";
    this.filtrosChange.emit({
      desde: inicioDelDia(this.desde),
      hasta: inicioDelDia(this.hasta),
      modulo: this.modulo,
      areaId: this.areaId,
    });
  }

  limpiar(): void {
    const def = filtrosPorDefecto(this.hoy);
    this.desde = def.desde;
    this.hasta = def.hasta;
    this.modulo = def.modulo;
    this.areaId = def.areaId;
    this.emitir();
  }

  get hayFiltrosQuitables(): boolean {
    return this.chips.some((c) => c.quitable);
  }

  get chips(): ChipFiltro[] {
    const chips: ChipFiltro[] = [];
    const def = filtrosPorDefecto(this.hoy);
    const periodoPorDefecto =
      this.filtros.desde.getTime() === def.desde.getTime() &&
      this.filtros.hasta.getTime() === def.hasta.getTime();
    chips.push({
      clave: "periodo",
      texto: `${formatoFecha(this.filtros.desde)} al ${formatoFecha(this.filtros.hasta)}`,
      quitable: !periodoPorDefecto,
    });
    if (this.filtros.modulo !== "todos") {
      chips.push({
        clave: "modulo",
        texto: `Módulo: ${NOMBRE_MODULO[this.filtros.modulo]}`,
        quitable: true,
      });
    }
    if (this.filtros.areaId) {
      const area = this.areas.find((a) => a.id === this.filtros.areaId);
      chips.push({
        clave: "area",
        texto: `Área: ${area?.nombre ?? this.filtros.areaId}`,
        quitable: true,
      });
    }
    return chips;
  }

  quitar(chip: ChipFiltro): void {
    const def = filtrosPorDefecto(this.hoy);
    if (chip.clave === "periodo") {
      this.desde = def.desde;
      this.hasta = def.hasta;
    } else if (chip.clave === "modulo") {
      this.modulo = "todos";
    } else {
      this.areaId = null;
    }
    this.emitir();
  }
}
