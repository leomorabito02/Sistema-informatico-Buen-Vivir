/**
 * Destino de "Volver al inicio" en las páginas públicas (contacto,
 * privacidad, términos, cookies, campañas): con sesión iniciada vuelve al
 * inicio del panel de su rol, sin sesión a la landing. Antes siempre iba a
 * la landing, aunque el usuario (por ejemplo una empresa) estuviera
 * logueado (pedido del 17/11/2026).
 *
 * Usa la marca local de sesión (la misma que usan el guard y el header):
 * si la cookie venció, el guard del panel lo manda a iniciar sesión.
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
