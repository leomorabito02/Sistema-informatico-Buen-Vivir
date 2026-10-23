/**
 * Ayudas de fechas para mostrar el dashboard. Los indicadores y las series
 * los calcula el microservicio de Dashboard (dashboard/internal/services/
 * indicadores.go): acá solo se arman fechas para los filtros y se
 * formatean para la pantalla y el PDF.
 */

/** "AAAA-MM-DD" de una fecha local (formato de los filtros del backend). */
export function fechaLocal(d: Date): string {
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${d.getFullYear()}-${mm}-${dd}`;
}

export function inicioDelDia(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate());
}

/** "2026-03" → "mar 2026" */
export function etiquetaMes(mes: string): string {
  const [anio, m] = mes.split("-").map(Number);
  const nombre = new Date(anio, m - 1, 1).toLocaleDateString("es-AR", {
    month: "short",
  });
  return `${nombre.replace(".", "")} ${anio}`;
}

/** DD/MM/AAAA */
export function formatoFecha(d: Date): string {
  const dd = String(d.getDate()).padStart(2, "0");
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  return `${dd}/${mm}/${d.getFullYear()}`;
}
