# 0003 — Dashboard sin snapshots: los eventos solo invalidan la caché

**Fecha:** 2026-10-16 · **Estado:** Aceptada (decisión de la usuaria)

## Contexto

El DER dice que Dashboard no tiene base de dominio propia y que consulta
en vivo a los demás servicios. Las gráficas de evolución temporal
(HU-53) necesitan historia, y la propuesta inicial era una colección
mínima de *snapshots* de indicadores alimentada por eventos, más
`ReporteExportado`.

Revisando el código: **Dashboard ya reconstruye la evolución sin
snapshots**, con las fechas que guardan los servicios de origen:

- participantes: alta (fecha de aprobación) y baja (`deleted_at`), desde
  `GET /api/v1/users/participants/history` de Autenticación;
- habilidades: la fecha en que se ganó cada una (Educación);
- postulaciones: la fecha de cada cambio de estado (Inserción Laboral);
- cursos y actividades: sus fechas en Contentful y Firestore.

## Decisión

**No agregar snapshots por ahora.** Dashboard sigue sin base de datos.
Los eventos (`participante.creado/eliminado`, `curso.creado/eliminado`,
`curso.completado_por_participante`, `habilidad.otorgada`,
`postulacion.seleccionada`) **solo invalidan una caché en memoria** de la
fuente afectada. La caché tiene un TTL de 5 minutos como respaldo.

Como invalidar es idempotente por naturaleza, la marca de eventos
procesados de Dashboard puede vivir en memoria
(`events.MemoryStore`): un duplicado, en el peor caso, invalida dos veces.

## Cuándo habría que revisar esta decisión

Hacen falta snapshots si se necesita historia que **no queda** en el
origen. El caso conocido: un curso **borrado** en Contentful desaparece
(borrado físico) y deja de contar en el histórico de cursos. Si eso
importa para HU-53, se crea una colección `snapshots_indicadores`
(fecha, indicador, área formativa, módulo, valor) alimentada por estos
mismos eventos, y ahí sí hay que actualizar el DER.

## Consecuencias

- El DER de Dashboard **no cambia**: sigue sin entidades propias.
- `ReporteExportado` no se implementó; queda para cuando se pida
  persistir las exportaciones (HU-55).
- La caché es por instancia (ver la sección 13 de `docs/eventos/README.md`).
