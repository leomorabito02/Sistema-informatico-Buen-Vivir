import { Injectable, inject } from "@angular/core";
import {
  HttpClient,
  HttpErrorResponse,
  HttpParams,
} from "@angular/common/http";
import { Observable, catchError, from, switchMap, throwError } from "rxjs";
import { environment } from "../../environments/environment";

export type EstadoModeracion = "pendiente" | "aprobado" | "rechazado";

/** Material de un docente (HU-23, DER ContenidoDocente). */
export interface ContenidoDocente {
  id: string;
  docenteId: string;
  titulo: string;
  descripcion: string;
  archivoNombre: string;
  archivoTipo: string;
  archivoTamano: number;
  estadoModeracion: EstadoModeracion;
  motivoRechazo?: string;
  fechaSubida: string;
  fechaModeracion?: string;
}

/** Lo que muestra la biblioteca pública: sin docente ni estado. */
export interface ContenidoPublico {
  id: string;
  titulo: string;
  descripcion: string;
  archivoNombre: string;
  archivoTipo: string;
  archivoTamano: number;
  fechaSubida: string;
}

/** Quién descargó de la biblioteca (HU-25). Solo lo ve la administración. */
export interface DescargaBiblioteca {
  id: string;
  contenidoId: string;
  nombre: string;
  apellido: string;
  email: string;
  telefono: string;
  fecha: string;
}

export interface DatosDescarga {
  nombre: string;
  apellido: string;
  email: string;
  telefono: string;
}

export interface MaterialForm {
  titulo: string;
  descripcion: string;
  archivo?: File | null;
}

/**
 * Contenido de docentes con moderación (HU-23) y biblioteca pública
 * (HU-25), en education-service.
 */
@Injectable({ providedIn: "root" })
export class ContenidosService {
  private readonly http = inject(HttpClient);
  private readonly base =
    environment.educationApiUrl || environment.apiUrl || "";
  private readonly opts = { withCredentials: true };

  // ── Docente ──
  misContenidos(): Observable<ContenidoDocente[]> {
    return this.http
      .get<ContenidoDocente[]>(`${this.base}/api/contenidos-docentes/mios`, this.opts)
      .pipe(catchError(this.error));
  }

  subir(m: MaterialForm): Observable<ContenidoDocente> {
    return this.http
      .post<ContenidoDocente>(`${this.base}/api/contenidos-docentes`, this.form(m), this.opts)
      .pipe(catchError(this.error));
  }

  editar(id: string, m: MaterialForm): Observable<ContenidoDocente> {
    return this.http
      .put<ContenidoDocente>(
        `${this.base}/api/contenidos-docentes/${encodeURIComponent(id)}`,
        this.form(m),
        this.opts,
      )
      .pipe(catchError(this.error));
  }

  eliminar(id: string): Observable<void> {
    return this.http
      .delete<void>(`${this.base}/api/contenidos-docentes/${encodeURIComponent(id)}`, this.opts)
      .pipe(catchError(this.error));
  }

  /** Archivo de un contenido, para usuarios con sesión. */
  archivo(id: string): Observable<Blob> {
    return this.http
      .get(`${this.base}/api/contenidos-docentes/${encodeURIComponent(id)}/archivo`, {
        ...this.opts,
        responseType: "blob",
      })
      .pipe(catchError(this.errorBlob));
  }

  // ── Administración ──
  listar(estado: EstadoModeracion | "" = ""): Observable<ContenidoDocente[]> {
    const params = estado ? new HttpParams().set("estado", estado) : undefined;
    return this.http
      .get<ContenidoDocente[]>(`${this.base}/api/contenidos-docentes`, { ...this.opts, params })
      .pipe(catchError(this.error));
  }

  moderar(id: string, estado: "aprobado" | "rechazado", motivo = ""): Observable<ContenidoDocente> {
    return this.http
      .post<ContenidoDocente>(
        `${this.base}/api/contenidos-docentes/${encodeURIComponent(id)}/moderacion`,
        { estado, motivo },
        this.opts,
      )
      .pipe(catchError(this.error));
  }

  descargas(): Observable<DescargaBiblioteca[]> {
    return this.http
      .get<DescargaBiblioteca[]>(`${this.base}/api/biblioteca/descargas`, this.opts)
      .pipe(catchError(this.error));
  }

  // ── Biblioteca pública (sin sesión) ──
  biblioteca(): Observable<ContenidoPublico[]> {
    return this.http
      .get<ContenidoPublico[]>(`${this.base}/api/biblioteca`)
      .pipe(catchError(this.error));
  }

  /** Registra los datos de quien descarga (HU-25) y devuelve el archivo. */
  descargarPublico(id: string, datos: DatosDescarga): Observable<Blob> {
    return this.http
      .post(`${this.base}/api/biblioteca/${encodeURIComponent(id)}/descargas`, datos, {
        responseType: "blob",
      })
      .pipe(catchError(this.errorBlob));
  }

  private form(m: MaterialForm): FormData {
    const fd = new FormData();
    fd.append("titulo", m.titulo);
    fd.append("descripcion", m.descripcion ?? "");
    if (m.archivo) fd.append("archivo", m.archivo, m.archivo.name);
    return fd;
  }

  private mensaje(status: number, texto?: string): string {
    if (status === 429) return "Hiciste muchos pedidos seguidos. Probá de nuevo en un minuto.";
    if (status === 404) return "Ese material ya no está disponible.";
    return texto || "No se pudo completar la acción. Probá de nuevo.";
  }

  private readonly error = (e: HttpErrorResponse) =>
    throwError(() => new Error(this.mensaje(e.status, e.error?.error)));

  /** Con responseType blob el error también llega como Blob: se lee el JSON. */
  private readonly errorBlob = (e: HttpErrorResponse) => {
    if (e.error instanceof Blob) {
      return from(e.error.text()).pipe(
        switchMap((t) => {
          let texto: string | undefined;
          try {
            texto = JSON.parse(t)?.error;
          } catch {
            texto = undefined;
          }
          return throwError(() => new Error(this.mensaje(e.status, texto)));
        }),
      );
    }
    return this.error(e);
  };
}

/** Guarda un Blob como archivo descargado. */
export function guardarArchivo(blob: Blob, nombre: string): void {
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = nombre;
  a.click();
  URL.revokeObjectURL(url);
}

/** Tamaño legible: 2,3 MB, 850 KB. */
export function tamanoLegible(bytes: number): string {
  if (bytes >= 1 << 20) return `${(bytes / (1 << 20)).toLocaleString("es-AR", { maximumFractionDigits: 1 })} MB`;
  return `${Math.max(1, Math.round(bytes / 1024))} KB`;
}
