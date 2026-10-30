import { Injectable, inject, signal } from "@angular/core";
import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { firstValueFrom } from "rxjs";
import { environment } from "../../environments/environment";
import {
  FiltrosAplicados,
  FiltrosDashboard,
  Indicador,
  ReporteExportado,
} from "../core/models/dashboard.model";
import { formatoFecha } from "./dashboard.fechas";

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

/** Resultado de exportar: el PDF siempre se descarga; guardarlo puede fallar. */
export interface ResultadoExportar {
  reporte: ReporteExportado | null;
  guardado: boolean;
  /** Por qué no se guardó en el historial (si no se guardó). */
  error?: string;
}

const MENSAJE_HISTORIAL = "No se pudo guardar el reporte en el historial.";

/**
 * Exportación del dashboard a PDF e historial de reportes (HU-55).
 *
 * El PDF se arma en el navegador y se descarga en el momento. Además se
 * manda al microservicio de Dashboard, que guarda los datos en Firestore
 * (tabla ReporteExportado del DER) y el PDF en Cloud Storage, para poder
 * volver a descargarlo después.
 */
@Injectable({ providedIn: "root" })
export class ReportesService {
  private http = inject(HttpClient);
  private readonly base = `${environment.apiUrl ?? ""}/api/v1/dashboard/reportes`;

  readonly historial = signal<ReporteExportado[]>([]);
  readonly cargandoHistorial = signal(false);
  readonly errorHistorial = signal<string | null>(null);

  /** Trae el historial guardado. */
  cargarHistorial(): void {
    this.cargandoHistorial.set(true);
    this.errorHistorial.set(null);
    this.http
      .get<ReporteExportado[]>(this.base, { withCredentials: true })
      .subscribe({
        next: (h) => {
          this.historial.set(h);
          this.cargandoHistorial.set(false);
        },
        error: (e: HttpErrorResponse) => {
          this.errorHistorial.set(
            mensajeDe(e, "No se pudo cargar el historial de reportes."),
          );
          this.cargandoHistorial.set(false);
        },
      });
  }

  async exportar(input: ExportarInput): Promise<ResultadoExportar> {
    const blob = await this.generarPdf(input);
    descargarBlob(blob, `reporte-dashboard-${aFechaIso(new Date())}.pdf`);
    try {
      const reporte = await firstValueFrom(this.guardar(input, blob));
      this.historial.update((h) => [reporte, ...h]);
      return { reporte, guardado: true };
    } catch (e) {
      return {
        reporte: null,
        guardado: false,
        error:
          e instanceof HttpErrorResponse
            ? mensajeDe(e, MENSAJE_HISTORIAL)
            : MENSAJE_HISTORIAL,
      };
    }
  }

  /** Vuelve a descargar un PDF guardado. */
  async descargar(reporte: ReporteExportado): Promise<void> {
    const blob = await firstValueFrom(
      this.http.get(`${this.base}/${encodeURIComponent(reporte.id)}/pdf`, {
        responseType: "blob",
        withCredentials: true,
      }),
    );
    descargarBlob(
      blob,
      `reporte-dashboard-${reporte.fecha_generacion.slice(0, 10)}.pdf`,
    );
  }

  /** Datos del reporte para el backend (el autor sale del token). */
  metadatos(input: ExportarInput) {
    const filtros: FiltrosAplicados = {
      modulo: input.filtros.modulo,
      areaId: input.filtros.areaId,
      areaNombre: input.areaNombre,
    };
    return {
      periodo_desde: aFechaIso(input.filtros.desde),
      periodo_hasta: aFechaIso(input.filtros.hasta),
      filtros_aplicados: filtros,
      generado_por_nombre: input.generadoPor,
    };
  }

  private guardar(input: ExportarInput, pdf: Blob) {
    const form = new FormData();
    form.append("metadatos", JSON.stringify(this.metadatos(input)));
    form.append("archivo", pdf, "reporte.pdf");
    return this.http.post<ReporteExportado>(this.base, form, {
      withCredentials: true,
    });
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

    // Indicadores: tarjetas en filas de hasta cuatro.
    if (input.indicadores.length) {
      const gap = 4;
      const porFila = 4;
      const n = Math.min(input.indicadores.length, porFila);
      const ancho = (anchoUtil - gap * (n - 1)) / n;
      input.indicadores.forEach((ind, i) => {
        if (i > 0 && i % porFila === 0) y += 30;
        const x = margen + (i % porFila) * (ancho + gap);
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

function descargarBlob(blob: Blob, nombre: string): void {
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = nombre;
  a.click();
  setTimeout(() => URL.revokeObjectURL(url), 10_000);
}

function mensajeDe(e: HttpErrorResponse, porDefecto: string): string {
  const body = e.error as { error?: string } | null;
  if (body && typeof body === "object" && body.error) return body.error;
  if (e.status === 401) return "Tu sesión expiró. Volvé a iniciar sesión.";
  return porDefecto;
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
