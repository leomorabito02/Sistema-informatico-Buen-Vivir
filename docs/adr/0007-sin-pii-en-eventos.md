# 0007 — Sin datos personales en los eventos

**Fecha:** 2026-10-16 · **Estado:** Aceptada

## Contexto

Los eventos se replican en Pub/Sub, en las DLQ (hasta 7 días), en el
outbox y en los logs. Los datos del centro incluyen información sensible
(situación personal, avances, DNI, contacto), protegida por la Ley 25.326
y por RNF-06/RNF-08. Las fichas de Promoción Humana, además, se guardan
cifradas.

## Decisión

Los payloads llevan **solo ids y metadatos**: ids de participante,
curso, inscripción, actividad y habilidad, fechas, estados y booleanos.
**Nunca** nombres, DNI, teléfono, email, texto libre (observaciones de
asistencia, textos de entregas) ni el contenido de fichas o hitos. Si un
consumidor necesita un dato personal, se lo pide al servicio dueño por
gRPC/HTTP, con sus controles de acceso.

Ejemplos concretos:
- `inscripcion.creada` no lleva el email: Notificaciones lo resuelve con
  `participante_id` contra Autenticación.
- `asistencia_general.*` no lleva la observación.
- `hito.*` no lleva la descripción, aunque esté cifrada.

## Consecuencias

- Una DLQ, un log o un outbox expuesto no filtran información de las
  personas.
- Los consumidores que necesitan datos hacen una llamada extra.
- Hay tests que verifican que esos campos no aparezcan en los payloads
  (`education/internal/events/producers/producers_test.go`,
  `*/internal/repositories/events_test.go`).
