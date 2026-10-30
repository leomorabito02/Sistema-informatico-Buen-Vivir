import { Component, Input, OnChanges, ViewChild } from "@angular/core";
import { ChartModule, UIChart } from "primeng/chart";
import { SkeletonModule } from "primeng/skeleton";
import { Grafica } from "../../../core/models/dashboard.model";
import { etiquetaMes } from "../../../services/dashboard.fechas";
import { GraficaParaPdf } from "../../../services/reportes.service";

/** Colores de las series, desde los tokens de la app. */
const TOKENS_SERIES = ["--green-600", "--earth-600"];

/**
 * Gráfica de evolución mensual de un indicador (HU-53): línea para
 * evoluciones, barras agrupadas para cursos en desarrollo vs. finalizados.
 * Incluye un resumen en texto y la tabla de datos para lectores de
 * pantalla (WCAG 1.1.1).
 */
@Component({
  selector: "app-grafica-evolucion",
  standalone: true,
  imports: [ChartModule, SkeletonModule],
  templateUrl: "./grafica-evolucion.component.html",
  styleUrl: "./grafica-evolucion.component.css",
})
export class GraficaEvolucionComponent implements OnChanges {
  @Input({ required: true }) grafica!: Grafica;
  @Input() loading = false;

  @ViewChild(UIChart) private chart?: UIChart;

  data: unknown = null;
  options: unknown = null;
  resumen = "";
  meses: string[] = [];

  get disponible(): boolean {
    return this.grafica?.series !== null;
  }

  /** Sin puntos o todos en cero: no hay nada que mostrar. */
  get vacia(): boolean {
    const series = this.grafica?.series ?? [];
    return (
      series.length === 0 ||
      series.every(
        (s) => s.puntos.length === 0 || s.puntos.every((p) => p.valor === 0),
      )
    );
  }

  ngOnChanges(): void {
    if (!this.grafica?.series) return;
    const series = this.grafica.series;
    this.meses = (series[0]?.puntos ?? []).map((p) => etiquetaMes(p.mes));
    const colores = TOKENS_SERIES.map((t, i) =>
      leerToken(t, i === 0 ? "#485839" : "#9a5b43"),
    );
    const texto = leerToken("--color-text-muted", "#66655b");
    const grilla = leerToken("--color-border", "#e7e6da");
    const esLinea = this.grafica.tipo === "line";

    this.data = {
      labels: this.meses,
      datasets: series.map((s, i) => ({
        label: s.nombre,
        data: s.puntos.map((p) => p.valor),
        borderColor: colores[i % colores.length],
        backgroundColor: esLinea
          ? colores[i % colores.length] + "22"
          : colores[i % colores.length],
        fill: esLinea,
        tension: 0.3,
        pointRadius: 3,
        borderRadius: esLinea ? 0 : 4,
        maxBarThickness: 28,
      })),
    };
    this.options = {
      maintainAspectRatio: false,
      animation: {
        duration: matchMedia?.("(prefers-reduced-motion: reduce)")?.matches
          ? 0
          : 400,
      },
      interaction: { mode: "index", intersect: false },
      plugins: {
        legend: {
          position: "bottom",
          labels: { color: texto, usePointStyle: true, boxWidth: 8 },
        },
        tooltip: { enabled: true },
      },
      scales: {
        x: { ticks: { color: texto }, grid: { display: false } },
        y: {
          beginAtZero: true,
          ticks: { color: texto, precision: 0 },
          grid: { color: grilla },
        },
      },
    };
    this.resumen = resumirSeries(this.grafica);
  }

  /** La gráfica como imagen, para el PDF (HU-55). */
  paraPdf(): GraficaParaPdf {
    const canvas: HTMLCanvasElement | undefined = this.chart?.chart?.canvas;
    const conImagen = !!canvas && this.disponible && !this.vacia;
    return {
      titulo: this.grafica.titulo,
      imagen: conImagen ? canvas!.toDataURL("image/png") : null,
      ancho: canvas?.width || 1,
      alto: canvas?.height || 1,
      resumen: !this.disponible
        ? "No disponible: el servicio de origen no respondió."
        : this.vacia
          ? "No hay datos para el período seleccionado."
          : this.resumen,
    };
  }
}

function leerToken(nombre: string, respaldo: string): string {
  const v = getComputedStyle(document.documentElement)
    .getPropertyValue(nombre)
    .trim();
  return /^#[0-9a-f]{6}$/i.test(v) ? v : respaldo;
}

/** Texto alternativo: punto de partida, de llegada y máximo de cada serie. */
export function resumirSeries(g: Grafica): string {
  if (!g.series) return "";
  return g.series
    .map((s) => {
      const p = s.puntos;
      if (p.length === 0) return `${s.nombre}: sin datos.`;
      const max = p.reduce((a, b) => (b.valor > a.valor ? b : a));
      const ini = p[0];
      const fin = p[p.length - 1];
      return (
        `${s.nombre}: ${ini.valor} en ${etiquetaMes(ini.mes)} y ` +
        `${fin.valor} en ${etiquetaMes(fin.mes)}; ` +
        `máximo de ${max.valor} en ${etiquetaMes(max.mes)}.`
      );
    })
    .join(" ");
}
