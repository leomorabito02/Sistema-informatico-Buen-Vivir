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
