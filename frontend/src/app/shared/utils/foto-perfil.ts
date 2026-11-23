/** Tope de la foto de perfil (HU-11); el backend también lo controla. */
export const MAX_FOTO_BYTES = 5 * 1024 * 1024;

const TIPOS_FOTO = ["image/jpeg", "image/png", "image/webp", "image/gif"];
const EXTENSIONES_FOTO = /\.(jpe?g|png|webp|gif)$/i;

/** Valor para el atributo accept del <input type="file">. */
export const ACCEPT_FOTO = ".jpg,.jpeg,.png,.webp,.gif";

/**
 * Valida una foto de perfil antes de subirla. Devuelve el mensaje para mostrar, o null
 * si se puede subir. El backend (rxresume) repite el control con el contenido real.
 */
export function validarFotoPerfil(
  file: File | null | undefined,
): string | null {
  if (!file) return "Elegí una imagen.";
  const tipoOk = TIPOS_FOTO.includes(file.type);
  const extensionOk = EXTENSIONES_FOTO.test(file.name);
  if (!tipoOk || !extensionOk) {
    return "La foto tiene que ser una imagen JPG, PNG, WEBP o GIF.";
  }
  if (file.size > MAX_FOTO_BYTES) {
    return "La imagen no debe superar los 5 MB.";
  }
  return null;
}
