# 0001 — Librería de eventos copiada en cada servicio

**Fecha:** 2026-10-16 · **Estado:** Aceptada

## Contexto

Todos los servicios necesitan el mismo código para la mensajería:
envelope, catálogo de eventos, handler push con idempotencia,
verificación OIDC y outbox. Cada servicio es un módulo Go aparte, y su
`Dockerfile` (y el `docker-compose.yml`) compila **solo con su propia
carpeta** como contexto: no ve nada fuera de ella.

Opciones:

1. **Un módulo compartido `pkg/`** importado por todos. Obliga a cambiar
   el contexto de build de todos los Dockerfiles, del `docker-compose.yml`
   y del pipeline, incluidos los servicios de otro integrante del equipo.
2. **Una copia en cada servicio**, generada desde una fuente única. Es lo
   que el repo ya hace con el proto de Autenticación (`api/proto/authpb`
   copiado en labor-integration y human-promotion).

## Decisión

Opción 2. La fuente canónica es `pkg/events` (solo librería estándar) y
`pkg/outbox` (Firestore + Pub/Sub), con sus tests. `scripts/sync-eventlib.sh`
copia los archivos a `<servicio>/internal/events` e `internal/outbox` y
les agrega el encabezado `// Code generated ... DO NOT EDIT.`.
`--check` falla si alguna copia quedó desactualizada.

`pkg/events` usa solo la librería estándar, incluida la verificación del
token OIDC, para que la pueda copiar cualquier servicio, también
Dashboard (Go 1.22, sin dependencias).

## Consecuencias

- No se tocó ningún Dockerfile ni el `docker-compose.yml`.
- El código está repetido: **nunca editar las copias**, editar `pkg/` y
  correr el script. Conviene sumar `scripts/sync-eventlib.sh --check` al CI.
- Para sumar un servicio (por ejemplo, Inserción Laboral), alcanza con
  agregarlo a `TARGETS` en el script.
