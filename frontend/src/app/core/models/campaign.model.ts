/**
 * Campañas de concientización (HU-47 a HU-51). Mismos nombres de campo
 * que campaigns-service (tabla Campaña del DER).
 */
export type EstadoCampana = "borrador" | "publicada";

export interface Campana {
  id: string;
  titulo: string;
  texto: string;
  imagen_url: string;
  /** Fecha calendario AAAA-MM-DD. */
  fecha: string;
  estado: EstadoCampana;
  /** Solo viene en las rutas de gestión, no en el listado público. */
  creado_por?: string;
}

export interface CampanaInput {
  titulo: string;
  texto: string;
  imagen_url: string;
  fecha: string;
}

/** Respuesta de crear, editar y publicar: la campaña y el mensaje de la HU. */
export interface CampanaEscritura {
  mensaje: string;
  campana: Campana;
}

export const DEFAULT_CAMPAIGN_PLACEHOLDER =
  "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='800' height='450' viewBox='0 0 800 450'%3E%3Crect width='100%25' height='100%25' fill='%23e2e8f0'/%3E%3Ctext x='50%25' y='50%25' dominant-baseline='middle' text-anchor='middle' fill='%2364748b' font-family='sans-serif' font-size='24'%3ECampa%C3%B1a%3C/text%3E%3C/svg%3E";

