# 0009 — Migración total de notificaciones a Pub/Sub y eliminación de HTTP sincrónico

**Fecha:** 2026-11-04 · **Estado:** Aceptada

## Contexto

Históricamente, mientras Educación y Promoción Humana utilizaban eventos de dominio en Pub/Sub, los servicios de Autenticación, Inserción Laboral y Campañas interactuaban con Notificaciones mediante despachos HTTP sincrónicos contra el endpoint `POST /api/v1/notifications/send`, autenticados por un secreto HMAC compartido (`NOTIFICATION_HMAC_SECRET`).

Esto presentaba múltiples problemas:
1. **Acoplamiento temporal:** Una degradación o caída del servicio de notificaciones bloqueaba o degradaba flujos críticos de negocio (registro de usuarios, moderación de ofertas, postulaciones).
2. **Pérdida de mensajes:** Los fallos de red durante el despacho HTTP carecían de reintentos durables fuera de la memoria del proceso emisor.
3. **Deuda técnica:** Coexistían dos mecanismos de comunicación en el sistema (HTTP directo con firmas HMAC y eventos asíncronos en Pub/Sub).
4. **Seguridad y tokens sensibles:** Despachar enlaces de verificación o restablecimiento por HTTP o eventos requería cuidar que los tokens de seguridad no se filtraran en colas o registros.

## Decisión

1. **Migración total a Pub/Sub + Outbox:** Todos los disparadores de notificaciones se migraron a eventos de dominio publicados mediante el patrón Outbox transaccional en cada microservicio (Postgres en `auth`, Firestore en `labor-integration`, MongoDB en `campaigns`).
2. **Eliminación completa del endpoint HTTP y HMAC:** Se eliminó de raíz la ruta `POST /api/v1/notifications/send`, el middleware de verificación HMAC (`hmac_auth.go`), y todas las implementaciones de `NotificationDispatcher` HTTP en los servicios emisores.
3. **Generación segura de enlaces a demanda (gRPC):** En estricto cumplimiento del [ADR 0007](0007-sin-pii-en-eventos.md), los eventos en el broker (`usuarios.usuario.creado`, `usuarios.email.solicitado`) transportan únicamente identificadores y tipo de acción, sin tokens ni credenciales. Al momento de despachar el correo, Notificaciones solicita el link de acción directamente a `auth` mediante gRPC (`IssueActionLink`).

## Consecuencias

- **Resiliencia:** Si Notificaciones o el proveedor de correo están caídos, los eventos esperan en el Outbox o en las suscripciones de Pub/Sub con reintentos exponenciales y DLQ sin afectar a los emisores.
- **Seguridad robusta:** Se elimina el secreto HMAC compartido entre servicios. Ningún token de recuperación ni enlace de verificación queda persistido en tópicos de Pub/Sub ni en colas de mensajes muertos (DLQ).
- **Consistencia arquitectónica:** Todos los servicios del sistema utilizan el mismo estándar de eventos ([ADR 0005](0005-formato-unico-de-eventos.md)) y entrega desacoplada.

