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

/**
 * Arma el `fecha` de una actividad a partir del día elegido en el
 * datepicker: medianoche UTC de ese día calendario ("2026-10-05T00:00:00Z").
 * Así la fecha no se corre de día por el huso horario, ni al guardarla
 * ni al mostrarla con `date: 'dd/MM/yyyy' : 'UTC'`.
 */
export function toActivityFecha(d: Date): string {
  return `${formatYMD(d)}T00:00:00Z`;
}

/** Inversa de `toActivityFecha`: el día calendario de la actividad como `Date` local (para el datepicker). */
export function activityFechaToDate(iso: string): Date {
  return parseYMD(iso.slice(0, 10));
}

/** Parsea "YYYY-MM-DD" a un `Date` local (evita el corrimiento de `new Date(str)`, que lo interpreta como UTC medianoche). */
export function parseYMD(s: string): Date {
  const [y, m, d] = s.split("-").map(Number);
  return new Date(y, m - 1, d);
}
