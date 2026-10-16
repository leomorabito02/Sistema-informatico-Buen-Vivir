# 0005 — Envelope único y relación con los servicios existentes

**Fecha:** 2026-10-16 · **Estado:** Aceptada

## Contexto

La regla era reutilizar los eventos que ya publicaran Autenticación e
Inserción Laboral y adaptar los consumidores a su formato. Revisión del
código (2026-10-16):

- **Autenticación** y **Inserción Laboral** no publican ningún evento:
  no hay cliente de Pub/Sub, ni topics, ni nombres de evento.
- El único uso de Pub/Sub es **rxresume**: `rxresume-sync-topic`, una
  cola de trabajo interna con cuerpo `{"user_id": "..."}`, sin envelope,
  sin atributos, pull, sin DLQ.

## Decisión

1. Todos los eventos nuevos usan **un envelope único** (`event_id`,
   `event_type`, `version`, `occurred_at`, `producer`, `correlation_id`,
   `actor_id`, `payload`), con `event_type`, `version`, `producer` y
   `correlation_id` también como atributos. Los nombres siguen la forma
   `<dominio>.<entidad>.<acción>`.
2. Como Autenticación e Inserción Laboral no publican, **no se agregó
   nada en su código**. Los eventos que necesitan los consumidores
   (`usuarios.participante.creado/eliminado`, `usuarios.usuario.actualizado`,
   `insercion_laboral.postulacion.seleccionada`) quedan **definidos y
   documentados como propuesta**: payload, topic y punto exacto del
   código donde publicarlos (`docs/eventos/README.md`, sección 5). Los
   consumidores ya están implementados y probados con ese formato.
3. **`rxresume-sync-topic` no se toca.** Es una cola de trabajo, no un
   evento de dominio, y no tiene otros consumidores. Unificarlo no le
   aporta nada a nadie hoy.

## Propuesta de unificación (a consultar, no aplicada)

- Si otro servicio necesita saber que un CV cambió: publicar
  `curriculum.cv.actualizado` con el envelope estándar en un topic propio,
  y dejar la cola interna como está.
- Mejora opcional para la cola de rxresume: DLQ y `max-delivery-attempts`,
  para que un CV que falla siempre no se reintente indefinidamente.

## Consecuencias

- Un solo formato para documentar (`asyncapi.yaml`), validar
  (`events.Decode`) y filtrar.
- Hasta que Autenticación e Inserción Laboral publiquen, los consumidores
  de `usuarios-eventos` e `insercion-laboral-eventos` no reciben nada. El
  sistema funciona igual que antes en esas partes.
