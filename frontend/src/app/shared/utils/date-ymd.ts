/**
 * Formatea una fecha como "YYYY-MM-DD" usando los getters locales (no
 * `toISOString()`, que convierte a UTC y puede correr el día en husos
 * horarios negativos). El backend de Educación espera este formato
 * exacto para la asistencia general (`fecha` es un string calendario,
 * no un instante — ver education/internal/models/general_attendance.go).
 */
export function formatYMD(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

/** Atajo para la fecha de hoy en formato "YYYY-MM-DD". */
export function todayYMD(): string {
  return formatYMD(new Date());
}

/** Parsea "YYYY-MM-DD" a un `Date` local (evita el corrimiento de `new Date(str)`, que lo interpreta como UTC medianoche). */
export function parseYMD(s: string): Date {
  const [y, m, d] = s.split("-").map(Number);
  return new Date(y, m - 1, d);
}
