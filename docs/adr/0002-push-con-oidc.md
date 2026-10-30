# 0002 — Suscripciones push con OIDC (no pull)

**Fecha:** 2026-10-16 · **Estado:** Aceptada

## Contexto

Los servicios corren en Cloud Run con *scale to zero*: si no hay tráfico,
no hay ninguna instancia corriendo. Una suscripción **pull** necesita un
proceso escuchando todo el tiempo (como hace hoy rxresume con `Receive`),
lo que obligaría a tener al menos una instancia siempre prendida y con CPU
asignada, para cada consumidor.

## Decisión

Suscripciones **push**: Pub/Sub hace un `POST` al endpoint del
consumidor (`/eventos/...`) y Cloud Run levanta la instancia si hace
falta.

- Cada push lleva un **token OIDC** emitido para una **service account
  dedicada por consumidor** (`sa-push-*`, ver `infra/pubsub/setup.sh`).
  El servicio verifica firma (claves públicas de Google), issuer,
  audience, vencimiento y que la cuenta sea una de las autorizadas
  (`PUBSUB_PUSH_SERVICE_ACCOUNTS`). Si no, responde 401.
- Ack deadline de 30 s, reintentos con backoff exponencial (10 s a
  10 min) y **dead-letter topic después de 5 intentos**.
- Los endpoints de eventos se montan **fuera** del router de la API: no
  pasan por CORS, por el JWT de usuario ni por el rate limit por IP
  (Pub/Sub manda los push desde IPs compartidas de Google).
- El barrido del outbox (`/internal/outbox/flush`) usa el mismo esquema:
  Cloud Scheduler con OIDC, una vez por minuto.
- Elegí `gcloud` y no Terraform para crear la infraestructura: el repo no
  tiene Terraform ni estado remoto, y el CD ya usa gcloud. El script es
  idempotente.

## Consecuencias

- Sin costo de instancias ociosas, y la entrega no depende de que el
  servicio esté despierto.
- La verificación OIDC es obligatoria en producción: el servicio **no
  arranca** con `PUBSUB_PUSH_AUTH=disabled` si `APP_ENV=production`.
- El emulador de Pub/Sub no firma los push: en local se usa
  `PUBSUB_PUSH_AUTH=disabled`, y los tests leen en pull y entregan al
  mismo handler HTTP.
