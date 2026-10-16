# 0004 — Patrón outbox en los productores

**Fecha:** 2026-10-16 · **Estado:** Aceptada

## Contexto

Si un servicio guarda un cambio y *después* publica el evento en
Pub/Sub, cualquier falla entre los dos pasos rompe la consistencia: o
queda el cambio sin evento (los demás servicios nunca se enteran), o el
evento sin el cambio.

## Decisión

El evento se escribe en la colección `outbox_eventos` **en la misma
transacción de Firestore** que el cambio de dominio (`outbox.Writer.AddTx`).
Un publicador (`outbox.Relay`) lo manda a Pub/Sub y lo borra cuando
Pub/Sub confirma. Si falla, lo reprograma con backoff exponencial (de 5 s
a 10 min) y lo vuelve a intentar sin límite: el evento no se pierde.

- El Relay corre en segundo plano y se despierta después de cada commit
  (`Notify`). Como en Cloud Run la instancia no tiene CPU entre requests,
  un job de Cloud Scheduler llama cada minuto a
  `POST /internal/outbox/flush`.
- Para que dos instancias no publiquen lo mismo a la vez, cada evento
  queda reservado 1 minuto (*lease*) por la que lo toma. Si igual sale
  duplicado (por ejemplo, la instancia se cae después de publicar), el
  consumidor lo descarta por `event_id`.
- Con *ordering key*, si falla un evento, los siguientes de la misma
  clave se postergan junto con él.

## Excepción: cursos

Los cursos viven en **Contentful**, que no puede participar de una
transacción de Firestore. El evento se escribe en el outbox apenas
Contentful confirma el cambio. Si esa escritura falla, el error queda en
el log con el `correlation_id` y el request no falla, porque el curso ya
está creado. Es el único caso sin garantía total, y está documentado.

## Consecuencias

- Hay una escritura más por transacción. Con la asistencia general en
  lote, el tope práctico baja a ~250 participantes por fecha (límite de
  500 escrituras de Firestore).
- Autenticación (Postgres) ya tiene un outbox propio
  (`StorageCleanupTask`) que sirve de modelo para cuando publique.
