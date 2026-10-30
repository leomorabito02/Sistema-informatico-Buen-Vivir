# 0008 — Receptor de webhooks de Contentful dentro de Educación

**Fecha:** 2026-10-16 · **Estado:** Aceptada (decisión de la usuaria)

## Contexto

Hacía falta un receptor `POST /webhooks/contentful` que valide el
secreto y traduzca `Entry.publish` / `unpublish` / `delete` a eventos de
`contenido-eventos`. Podía ser un servicio nuevo de Cloud Run o una ruta
más de un servicio existente. El sistema ya tiene más contenedores que
cajas en el diagrama de arquitectura.

## Decisión

Una ruta dentro de **Educación**, el único servicio que ya habla con
Contentful. Valida el header `X-Webhook-Secret` en tiempo constante (el
secreto viene de Secret Manager como variable de entorno) y responde 401
si no coincide. Deja el evento en el outbox y responde **202 enseguida**:
todo el procesamiento ocurre en los consumidores. Para que los
reintentos de Contentful no generen eventos distintos, el `event_id` es
determinístico (entry + acción + revisión).

## Consecuencias

- No suma servicios ni contenedores.
- Educación publica en dos topics: `educacion-eventos` y
  `contenido-eventos`. Su service account necesita permiso de publicación
  en ambos (`setup.sh` lo da).
- Educación consume sus propios eventos de contenido. Es intencional:
  así el aviso llega a la instancia que Pub/Sub elija, no solo a la que
  recibió el webhook.
