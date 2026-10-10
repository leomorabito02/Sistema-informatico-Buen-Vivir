import { Injectable, signal } from "@angular/core";
import {
  FiltrosAplicados,
  FiltrosDashboard,
  Indicador,
  ReporteExportado,
} from "../core/models/dashboard.model";
import { formatoFecha } from "./dashboard.calculos";

/** Una gráfica tal como se ve en pantalla, lista para el PDF. */
export interface GraficaParaPdf {
  titulo: string;
  /** PNG en data URL (canvas de Chart.js), o null si no hay datos. */
  imagen: string | null;
  ancho: number;
  alto: number;
  resumen: string;
}

export interface ExportarInput {
  filtros: FiltrosDashboard;
  areaNombre: string | null;
  indicadores: Indicador[];
  graficas: GraficaParaPdf[];
  generadoPor: string;
}

export const NOMBRE_MODULO: Record<string, string> = {
  todos: "Todos",
  usuarios: "Usuarios",
  educacion: "Educación",
  promocion: "Promoción Humana",
  insercion: "Inserción Laboral",
};

/**
 * Exportación del dashboard a PDF (HU-55) e historial de reportes.
 *
 * HOY el historial vive en memoria y `url_archivo` es una URL local del
 * navegador (válida mientras la pestaña esté abierta). Con backend, el
 * PDF se sube al bucket y `registrar()` pasa a ser un POST que devuelve
 * el ReporteExportado con la URL definitiva.
 */
@Injectable({ providedIn: "root" })
export class ReportesService {
  readonly historial = signal<ReporteExportado[]>([]);

  async exportar(input: ExportarInput): Promise<ReporteExportado> {
    const blob = await this.generarPdf(input);
    const reporte = this.registrar(input, URL.createObjectURL(blob));
    this.descargar(reporte);
    return reporte;
  }

  descargar(reporte: ReporteExportado): void {
    const a = document.createElement("a");
    a.href = reporte.url_archivo;
    a.download = `reporte-dashboard-${reporte.fecha_generacion.slice(0, 10)}.pdf`;
    a.click();
  }

  registrar(input: ExportarInput, url: string): ReporteExportado {
    const filtros: FiltrosAplicados = {
      modulo: input.filtros.modulo,
      areaId: input.filtros.areaId,
      areaNombre: input.areaNombre,
    };
    const reporte: ReporteExportado = {
      id: crypto.randomUUID(),
      tipo_reporte: "dashboard",
      periodo_desde: aFechaIso(input.filtros.desde),
      periodo_hasta: aFechaIso(input.filtros.hasta),
      filtros_aplicados: filtros,
      generado_por: input.generadoPor,
      fecha_generacion: new Date().toISOString(),
      url_archivo: url,
    };
    this.historial.update((h) => [reporte, ...h]);
    return reporte;
  }

  /** Arma el PDF: logo, fecha, filtros, indicadores y gráficas. */
  async generarPdf(input: ExportarInput): Promise<Blob> {
    // Carga diferida: jsPDF solo se descarga cuando alguien exporta.
    const { jsPDF } = await import("jspdf");
    const doc = new jsPDF({ unit: "mm", format: "a4" });
    const verde = colorToken("--color-primary", "#485839");
    const gris = colorToken("--color-text-muted", "#66655b");
    const margen = 15;
    const anchoUtil = 210 - margen * 2;
    let y = margen;

    const logo = await cargarImagen("assets/logo.png");
    if (logo) doc.addImage(logo, "PNG", margen, y, 16, 16);
    doc.setFont("helvetica", "bold");
    doc.setFontSize(16);
    doc.setTextColor(verde);
    doc.text("Reporte de indicadores", margen + 20, y + 7);
    doc.setFont("helvetica", "normal");
    doc.setFontSize(10);
    doc.setTextColor(gris);
    doc.text("Centro de Ecología Integral - Buen Vivir", margen + 20, y + 13);
    y += 24;

    doc.setTextColor("#1d1d1b");
    const { filtros } = input;
    const lineas = [
      `Generado el ${formatoFecha(new Date())} por ${input.generadoPor}`,
      `Período: ${formatoFecha(filtros.desde)} al ${formatoFecha(filtros.hasta)}`,
      `Módulo: ${NOMBRE_MODULO[filtros.modulo]}    Área formativa: ${input.areaNombre ?? "Todas"}`,
    ];
    for (const l of lineas) {
      doc.text(l, margen, y);
      y += 6;
    }
    y += 4;

    // Indicadores: una fila de tarjetas.
    if (input.indicadores.length) {
      const gap = 4;
      const n = input.indicadores.length;
      const ancho = (anchoUtil - gap * (n - 1)) / n;
      input.indicadores.forEach((ind, i) => {
        const x = margen + i * (ancho + gap);
        doc.setDrawColor("#d6d4c6");
        doc.roundedRect(x, y, ancho, 26, 2, 2);
        doc.setFont("helvetica", "bold");
        doc.setFontSize(ind.disponible ? 18 : 10);
        doc.setTextColor(ind.disponible ? verde : gris);
        doc.text(
          ind.disponible ? String(ind.valor) : "No disponible",
          x + 3,
          y + 10,
        );
        doc.setFont("helvetica", "normal");
        doc.setFontSize(8);
        doc.setTextColor("#1d1d1b");
        doc.text(doc.splitTextToSize(ind.titulo, ancho - 6), x + 3, y + 17);
      });
      y += 34;
    }

    // Gráficas: una por fila, a todo el ancho.
    for (const g of input.graficas) {
      const alto = g.imagen ? (anchoUtil * g.alto) / g.ancho : 10;
      if (y + alto + 18 > 297 - margen) {
        doc.addPage();
        y = margen;
      }
      doc.setFont("helvetica", "bold");
      doc.setFontSize(11);
      doc.setTextColor("#1d1d1b");
      doc.text(g.titulo, margen, y);
      y += 4;
      if (g.imagen) {
        doc.addImage(g.imagen, "PNG", margen, y, anchoUtil, alto);
        y += alto + 3;
      } else {
        y += 3;
      }
      doc.setFont("helvetica", "normal");
      doc.setFontSize(9);
      doc.setTextColor(gris);
      const resumen = doc.splitTextToSize(g.resumen, anchoUtil);
      doc.text(resumen, margen, y + 3);
      y += 6 + resumen.length * 4;
    }

    return doc.output("blob");
  }
}

function aFechaIso(d: Date): string {
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${d.getFullYear()}-${mm}-${dd}`;
}

function colorToken(nombre: string, respaldo: string): string {
  const v = getComputedStyle(document.documentElement)
    .getPropertyValue(nombre)
    .trim();
  return /^#[0-9a-f]{6}$/i.test(v) ? v : respaldo;
}

async function cargarImagen(url: string): Promise<string | null> {
  try {
    const res = await fetch(url);
    if (!res.ok) return null;
    const blob = await res.blob();
    return await new Promise<string>((resolve, reject) => {
      const r = new FileReader();
      r.onload = () => resolve(r.result as string);
      r.onerror = reject;
      r.readAsDataURL(blob);
    });
  } catch {
    return null;
  }
}
