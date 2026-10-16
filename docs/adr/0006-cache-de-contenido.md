# 0006 — Caché de Contentful con contenido stale

**Fecha:** 2026-10-16 · **Estado:** Aceptada

## Contexto

HU-17 y el riesgo R2 piden que, si Contentful falla (5xx o timeout), se
siga sirviendo el último contenido válido. Hasta ahora Educación leía la
Content Delivery API en cada request, sin caché: si Contentful caía, los
`GET` de cursos, áreas y materiales fallaban.

## Decisión

Una caché en memoria dentro de `ContentfulClient`, solo para las
lecturas de la CDA (`repositories/content_cache.go`):

- Respuesta fresca (TTL `CONTENT_CACHE_TTL`, 5 min por defecto): se sirve
  sin llamar a Contentful.
- Vencida o invalidada: se le pide a Contentful. Si responde bien, se
  guarda. Si falla con 5xx o sin respuesta, **se sirve la versión
  guardada** y queda un aviso en el log. Un 4xx no se tapa con contenido
  viejo.
- **Invalidar no borra**: vence las entradas pero conserva el cuerpo como
  respaldo. Invalidan los eventos `contenido.publicado/despublicado`
  (webhook → Pub/Sub) y cada escritura propia por la CMA, para que quien
  edita vea su cambio enseguida.
- Se invalida todo, no una entry puntual: los listados incluyen entries
  y referencias resueltas, y saber cuáles se ven afectadas no vale la
  complejidad.

## Consecuencias

- Menos llamadas a Contentful y tolerancia a sus caídas.
- Es **por instancia**: el push invalida la instancia que lo recibe, y las
  demás se ponen al día al vencer el TTL. Con una instancia (lo habitual
  con este tráfico) no se nota. Si hiciera falta consistencia inmediata
  entre instancias, el siguiente paso sería una caché compartida
  (Memorystore) o la Sync API de Contentful.
- Tope de 500 respuestas guardadas.
