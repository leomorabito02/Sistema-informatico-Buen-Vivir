/**
 * Texto para comparar en los buscadores de listados: minúsculas, sin
 * acentos y con los espacios colapsados ("José  Pérez" -> "jose perez").
 */
export function normalizarBusqueda(texto: string | null | undefined): string {
  return (texto ?? "")
    .normalize("NFD")
    .replace(/[̀-ͯ]/g, "")
    .toLowerCase()
    .replace(/\s+/g, " ")
    .trim();
}

/**
 * true si todas las palabras buscadas aparecen en alguno de los campos.
 * Una búsqueda vacía coincide con todo.
 */
export function coincideBusqueda(
  busqueda: string,
  ...campos: (string | null | undefined)[]
): boolean {
  const palabras = normalizarBusqueda(busqueda).split(" ").filter(Boolean);
  if (palabras.length === 0) return true;
  const texto = normalizarBusqueda(campos.join(" "));
  return palabras.every((p) => texto.includes(p));
}
