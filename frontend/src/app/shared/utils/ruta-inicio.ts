/**
 * Destino de "Volver al inicio" en las páginas públicas: con sesión, el inicio del
 * panel de su rol; sin sesión, la landing. Usa la marca local de sesión (si la cookie
 * venció, el guard lo manda a iniciar sesión).
 */
export function rutaInicio(): string {
  try {
    return localStorage.getItem("isAuthenticated") === "true"
      ? "/dashboard"
      : "/";
  } catch {
    return "/";
  }
}
