# Decisiones de arquitectura (ADR)

Registro corto de decisiones importantes: qué se decidió, por qué y qué
consecuencias tiene. Un ADR no se edita para cambiar la decisión: si
cambia, se escribe uno nuevo que lo reemplace.

| # | Decisión | Fecha | Estado |
|---|---|---|---|
| [0001](0001-libreria-de-eventos-copiada.md) | Librería de eventos copiada en cada servicio | 2026-10-16 | Aceptada |
| [0002](0002-push-con-oidc.md) | Suscripciones push con OIDC (no pull) | 2026-10-16 | Aceptada |
| [0003](0003-dashboard-sin-snapshots.md) | Dashboard sin snapshots: eventos solo para invalidar caché | 2026-10-16 | Aceptada |
| [0004](0004-outbox.md) | Patrón outbox en los productores | 2026-10-16 | Aceptada |
| [0005](0005-formato-unico-de-eventos.md) | Envelope único y relación con los servicios existentes | 2026-10-16 | Aceptada |
| [0006](0006-cache-de-contenido.md) | Caché de Contentful con contenido stale | 2026-10-16 | Aceptada |
| [0007](0007-sin-pii-en-eventos.md) | Sin datos personales en los eventos | 2026-10-16 | Aceptada |
| [0008](0008-webhook-contentful-en-education.md) | Receptor de webhooks de Contentful dentro de Educación | 2026-10-16 | Aceptada |
| [0009](0009-migracion-notificaciones-pubsub-total.md) | Migración total de notificaciones a Pub/Sub y eliminación de HTTP sincrónico | 2026-11-04 | Aceptada |
