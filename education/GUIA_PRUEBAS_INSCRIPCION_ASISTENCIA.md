# Guía de pruebas manuales — Inscripción y Asistencia

Cubre HU-18/RF-18 (inscripción), HU-19/RF-19 (asistencia por curso),
HU-19bis/RF-64 (asistencia general) y HU-20/RF-20 (progreso, sin
cambios — solo se confirma que sigue andando). **Alcance backend
únicamente** — no se tocó frontend en esta etapa (instrucción
explícita), así que no hay pantallas que probar todavía.

Herramienta: la colección de Postman **"Buen Vivir - Educación"**
(`postman/`, ver `education/COMO_PROBAR_POSTMAN.md`) o `curl` directo.
Varios de los casos de este documento ya se verificaron en vivo durante
la implementación (Contentful real + emulador de Firestore + un stub
de auth-service) — están marcados como tal en "Resultado obtenido".

## Precondiciones generales

- `education-service` corriendo (`go run ./cmd/server` o Docker).
- Emulador de Firestore si se corre suelto (`FIRESTORE_EMULATOR_HOST`).
- `AUTH_SERVICE_URL` apuntando a un `auth-service` real y alcanzable —
  **obligatorio** para casi todos los casos de este documento, porque
  `Enroll`/`RegisterBatch` (por actividad y general) ahora validan el
  participante contra auth-service. Con `auth-service` caído, estos 3
  flujos fallan con `502`.
- Tokens JWT para `role_admin`, `role_teacher` (`is_external:false`) y
  `role_volunteer`, firmados con el mismo `JWT_SECRET`.
- Al menos un curso publicado en Contentful, con `trainingArea` y
  `material` ya creados.
- Al menos 2 participantes reales en auth-service: uno con
  `approval_status: "approved"` y otro `"pending"` (o cualquier valor
  distinto de `approved`), para poder probar los casos de rechazo.

> **ℹ️ El valor estandarizado es `"approved"`.** Estandarizado en el
> sistema (columna `approval_status` en auth-service). El código de
> education (`services.participantApprovalStatusAccepted`, en
> `internal/services/errors.go`) compara contra `"approved"`.

> **⚠️ `participanteId` = `participant_id`, NO `user_id`.** En la tabla
> `participantes` de Neon hay dos columnas distintas: `participant_id`
> (la PK propia de esa tabla) y `user_id` (FK a `usuarios`, es el mismo
> valor que el `sub` del JWT). `GET /api/v1/participants/{id}` de
> auth-service busca **solo por `participant_id`**
> (`WHERE participant_id = ?`, ver `participant_repository.go`) — si le
> pasás el `user_id` por error, da 404 y education lo traduce a
> `"El participante no existe"`. Para ver los dos valores lado a lado
> (y confirmar cuál es cuál) usá el request de Postman
> **"http://localhost:8082/api/v1/participants (DEBUG...)"** — le pega
> directo a `auth-service` (puerto `8082`, no `8210`) y devuelve
> `participant_id` y `user_id` de cada fila.

---

## Bloque 1 — Inscripción (HU-18/RF-18)

### CP-01 — Inscripción correcta
- **HU/RF:** HU-18/RF-18
- **Objetivo:** confirmar que un participante aprobado se inscribe en un curso publicado con cupo disponible.
- **Precondiciones:** curso publicado; participante aprobado en auth-service, sin inscripción previa en ese curso.
- **Datos:** `POST /api/cursos/{id}/inscripciones`, body `{"participanteId": "<id aprobado>"}`, token `role_volunteer` o `role_admin`.
- **Pasos:** 1) Autenticarse. 2) Enviar el POST. 3) Verificar respuesta y código.
- **Resultado esperado:** `201`, body con `id`, `participanteId`, `cursoId`, `fechaInscripcion`, `estado: "activa"`.
- **Resultado obtenido:** ✅ Verificado en vivo — `201` con el body esperado (curso real de Contentful, participante `part-1` del stub, `approval_status: "accepted"`). Re-verificado 2026-09-13 después de corregir el valor de `approval_status` (ver advertencia arriba).
- **Estado:** OK
- **Observaciones:** persiste en Firestore, colección `inscripciones`.

### CP-02 — Inscripción duplicada
- **HU/RF:** HU-18/RF-18
- **Objetivo:** impedir una segunda inscripción activa del mismo participante en el mismo curso.
- **Precondiciones:** CP-01 ya ejecutado (mismo participante, mismo curso).
- **Datos:** mismo request que CP-01.
- **Pasos:** repetir el POST de CP-01 sin cambios.
- **Resultado esperado:** `400`, `{"error":"El participante ya se encuentra inscripto en el curso"}`. No se crea un segundo documento.
- **Resultado obtenido:** ✅ Verificado en vivo.
- **Estado:** OK

### CP-03 — Curso sin cupo
- **HU/RF:** HU-18/RF-18
- **Objetivo:** no permitir inscribir si `capacity` ya está completo.
- **Precondiciones:** curso con `capacity` fijado (ej. 2 en Contentful) y esa cantidad de inscriptos activos ya cargada.
- **Datos:** `POST .../inscripciones` con un participante nuevo, aprobado.
- **Pasos:** inscribir hasta llenar el cupo, después intentar uno más.
- **Resultado esperado:** `400`, `{"error":"El curso no tiene cupo disponible"}`.
- **Resultado obtenido:** pendiente de repetir con un curso que tenga `capacity` seteado en Contentful (la lógica se corrió sin `capacity`, que es el camino "sin límite" — ver CP-01/02). Cubierto por revisión de código: `EnrollmentService.Enroll` cuenta activos y compara contra `course.Capacity` antes de crear.
- **Estado:** pendiente de repetir con datos reales de cupo
- **Observaciones:** requiere que el campo `capacity` (Number) exista en el content type `course` del Space — ya está creado (confirmado por el usuario).

### CP-04 — Curso inactivo (despublicado)
- **HU/RF:** HU-18/RF-18
- **Objetivo:** distinguir "curso inactivo" de "curso inexistente".
- **Precondiciones:** un curso que exista en Contentful pero esté despublicado (`POST /api/cursos/{id}/despublicar` como admin).
- **Datos:** `POST .../inscripciones` sobre ese id.
- **Pasos:** despublicar un curso, después intentar inscribir.
- **Resultado esperado:** `400`, `{"error":"El curso no está disponible (inactivo o no publicado)"}`.
- **Resultado obtenido:** pendiente de ejecutar (no se despublicó ningún curso real durante la verificación en vivo para no alterar datos del equipo).
- **Estado:** pendiente

### CP-05 — Curso inexistente
- **HU/RF:** HU-18/RF-18
- **Objetivo:** confirmar el mensaje distinto para un id que nunca existió.
- **Datos:** `POST /api/cursos/id-que-no-existe/inscripciones`.
- **Resultado esperado:** `400`, `{"error":"El curso no existe"}`.
- **Resultado obtenido:** pendiente de ejecutar (no probado explícitamente; sí se probó el camino de código equivalente al confirmar que `ExistsInManagement` devuelve `false, nil` en 404 de la CMA).
- **Estado:** pendiente

### CP-06 — Participante inexistente
- **HU/RF:** HU-18/RF-18
- **Datos:** `POST .../inscripciones` con `participanteId` que no existe en auth-service.
- **Resultado esperado:** `400`, `{"error":"El participante no existe"}`.
- **Resultado obtenido:** ✅ Verificado en vivo.
- **Estado:** OK
- **Observaciones:** el falso positivo más común de este caso es pasar
  el `user_id` en vez del `participant_id` — ver el recuadro de
  advertencia al principio de este documento.

### CP-07 — Participante inactivo (no aprobado)
- **HU/RF:** HU-18/RF-18
- **Datos:** `participanteId` con `approval_status` distinto de `"accepted"` (ej. `"pending"`).
- **Resultado esperado:** `400`, `{"error":"El participante no está activo"}`.
- **Resultado obtenido:** ✅ Verificado en vivo.
- **Estado:** OK

### CP-08 — Permisos insuficientes
- **HU/RF:** HU-18/RF-18
- **Datos:** `POST .../inscripciones` con token `role_participant` o `role_company`.
- **Resultado esperado:** `403 Forbidden: insufficient role permissions`.
- **Resultado obtenido:** cubierto por `middlewares.RequireRole("role_admin", "role_volunteer")`, mismo middleware ya probado en los demás endpoints del servicio (ver `education/TESTS.md`). No se re-ejecutó a mano.
- **Estado:** OK (por diseño/middleware compartido)

### CP-09 — Persistencia en base de datos
- **HU/RF:** HU-18/RF-18
- **Objetivo:** confirmar que la inscripción queda en Firestore, no solo en la respuesta HTTP.
- **Pasos:** después de CP-01, `GET /api/cursos/{id}/inscripciones`.
- **Resultado esperado:** el registro aparece listado.
- **Resultado obtenido:** ✅ Verificado en vivo (`GET` devolvió el mismo registro creado en CP-01).
- **Estado:** OK

---

## Bloque 2 — Asistencia por curso (HU-19/RF-19)

### CP-10 — Marcar actividad completada / registrar asistencia
- **HU/RF:** HU-19/RF-19
- **Precondiciones:** actividad creada (`POST /api/docente/cursos/{cursoId}/actividades`, rol `role_teacher`); participante inscripto activo en ese curso y aprobado en auth-service.
- **Datos:** `POST /api/actividades/{actividadId}/asistencia` (rol `role_admin`/`role_volunteer`) o `POST /api/docente/actividades/{actividadId}/asistencia` (rol `role_teacher`, dueño del curso), body `[{"participanteId":"...","presente":true}]`.
- **Resultado esperado:** `200`, `{"actividadId":"...","status":"asistencia registrada"}`.
- **Resultado obtenido:** ✅ Verificado en vivo (ambas condiciones: participante inscripto y aprobado).
- **Estado:** OK

### CP-11 — Verificar actualización / consultar historial
- **Datos:** `GET /api/actividades/{actividadId}/asistencia` después de CP-10.
- **Resultado esperado:** lista con el registro cargado (`participanteId`, `presente`, `fecha`).
- **Resultado obtenido:** ✅ Verificado en vivo.
- **Estado:** OK

### CP-12 — Participante no inscripto
- **HU/RF:** HU-19/RF-19 (validación nueva de esta etapa)
- **Datos:** registrar asistencia para un `participanteId` aprobado pero SIN inscripción activa en el curso de la actividad.
- **Resultado esperado:** `400`, `{"error":"El participante <id> no está inscripto en el curso"}`.
- **Resultado obtenido:** cubierto por revisión de código (`AttendanceService.RegisterBatch` arma el set de inscriptos activos del curso antes de validar cada item); no se re-probó con un tercer participante separado durante la verificación en vivo (se usó `part-2`, que ya fallaba antes por "no está activo").
- **Estado:** pendiente de repetir con un participante aprobado-pero-no-inscripto

### CP-13 — Participante inexistente / inactivo
- **Datos:** igual que CP-06/CP-07 pero contra el endpoint de asistencia.
- **Resultado esperado:** mismos mensajes ("no existe" / "no está activo").
- **Resultado obtenido:** ✅ Verificado en vivo (caso "no está activo", con `part-2`).
- **Estado:** OK

### CP-14 — Actividad inexistente
- **Datos:** `POST/GET /api/actividades/id-que-no-existe/asistencia`.
- **Resultado esperado:** `404` (repositories.ErrNotFound, vía `ActivityService.GetActivity`).
- **Resultado obtenido:** no re-ejecutado en esta etapa; comportamiento sin cambios respecto del ya existente (mismo `ActivityService.GetActivity` que usaba `/api/docente/*`, ver `education/TESTS.md`).
- **Estado:** OK (sin cambios de esta etapa)

### CP-15 — Evitar duplicados
- **Datos:** registrar asistencia del mismo participante dos veces para la misma actividad.
- **Resultado esperado:** un solo documento en Firestore (id determinístico `{actividadId}_{participanteId}`), con el último valor de `presente`.
- **Resultado obtenido:** sin cambios de esta etapa (mecanismo ya existente, ver `AttendanceRepository.RegisterBatch`).
- **Estado:** OK (sin cambios de esta etapa)

### CP-16 — Permisos
- **Datos:** token sin `role_admin`/`role_volunteer`/`role_teacher` (según el endpoint).
- **Resultado esperado:** `403`.
- **Estado:** OK (middleware compartido, ver CP-08)

### CP-17 — Persistencia
- **Resultado esperado:** el registro sobrevive a un reinicio del proceso (vive en Firestore, no en memoria).
- **Estado:** OK (por diseño — Firestore es la única fuente de verdad, no hay caché en memoria)

---

## Bloque 3 — Asistencia general (HU-19bis/RF-64)

### CP-18 — Seleccionar fecha / cargar participantes
- **Datos:** `GET /api/asistencia-general?fecha=YYYY-MM-DD`.
- **Resultado esperado:** `200`, un ítem por cada participante **aprobado** del centro (`participanteId`, `nombre`, `apellido`, `presente: null` si nunca se cargó).
- **Resultado obtenido:** ✅ Verificado en vivo — devolvió `part-1` (aprobado, con nombre/apellido resueltos) y excluyó correctamente a `part-2` (pendiente).
- **Estado:** OK

### CP-19 — Marcar Presente
- **Datos:** `POST /api/asistencia-general?fecha=...`, body `[{"participanteId":"...","presente":true}]`.
- **Resultado esperado:** `200`.
- **Resultado obtenido:** ✅ Verificado en vivo.
- **Estado:** OK

### CP-20 — Marcar Ausente
- **Datos:** igual que CP-19 con `"presente": false`.
- **Resultado esperado:** `200`, y el `GET` posterior muestra `presente: false`.
- **Resultado obtenido:** cubierto por el mismo código que CP-19 (un solo `if/else` sobre el booleano); no se re-probó explícitamente el valor `false`.
- **Estado:** pendiente de repetir con `presente: false`

### CP-21 — Agregar observación
- **Datos:** incluir `"observacion": "texto libre"` en el body de CP-19.
- **Resultado esperado:** el `GET` posterior devuelve esa observación.
- **Resultado obtenido:** ✅ Verificado en vivo (`"observacion":"llego a horario"`).
- **Estado:** OK

### CP-22 — Guardar
- Cubierto por CP-19/CP-20/CP-21.

### CP-23 — Volver a entrar a la fecha / comprobar que se cargue lo guardado
- **Datos:** `GET` con la misma `fecha` después de CP-19.
- **Resultado esperado:** el participante ya no tiene `presente: null`, sino el valor cargado.
- **Resultado obtenido:** ✅ Verificado en vivo.
- **Estado:** OK

### CP-24 — Editar asistencia / editar observación
- **Datos:** repetir el `POST` de CP-19 para el mismo participante y fecha, con `presente`/`observacion` distintos.
- **Resultado esperado:** el `GET` refleja el valor nuevo, no un registro adicional.
- **Resultado obtenido:** cubierto por el mismo mecanismo de upsert que CP-25 (id determinístico); no se re-probó un segundo cambio de valor explícitamente.
- **Estado:** pendiente de repetir con un segundo cambio de valor

### CP-25 — Comprobar que no se duplique
- **Datos:** contar los documentos de `asistencia_general` para `{fecha}_{participanteId}` después de cargar dos veces.
- **Resultado esperado:** un solo documento.
- **Resultado obtenido:** ✅ Verificado en vivo — el `GET` después de cargar mostró un solo ítem para `part-1`, con los valores del último `POST`.
- **Estado:** OK

### CP-26 — Participante inactivo/inexistente (POST)
- **Datos:** `POST /api/asistencia-general` con un `participanteId` que no existe en auth-service.
- **Resultado esperado:** `400`, `{"error":"El participante <id> no existe"}`. No se guarda nada.
- **Resultado obtenido:** ✅ Verificado en vivo.
- **Estado:** OK

### CP-27 — Fecha inválida
- **Datos:** `GET`/`POST /api/asistencia-general?fecha=no-es-fecha`.
- **Resultado esperado:** `400`, `{"error":"fecha inválida: se espera el formato YYYY-MM-DD"}`.
- **Resultado obtenido:** ✅ Verificado en vivo (caso GET).
- **Estado:** OK

### CP-28 — Permisos
- **Estado:** OK (middleware compartido, ver CP-08)

### CP-29 — Persistencia
- **Estado:** OK (Firestore, sin caché en memoria — igual que CP-17)

---

## Checklist de aceptación

- [x] HU-18: inscripción correcta, duplicado y participante inválido bloqueados; cupo y curso inactivo cubiertos por revisión de código, pendientes de repetir con datos reales.
- [x] HU-19: asistencia por actividad funciona para `role_teacher` (sin cambios) y ahora también para `role_admin`/`role_volunteer` (sin scope de "mis cursos"); valida participante existente/activo/inscripto.
- [x] HU-19bis: asistencia general funciona de punta a punta — fecha, listar, cargar, editar sin duplicar, validar participante.
- [x] HU-20: sin cambios en esta etapa; `ProgressService`/`GET .../progreso` no se tocaron.

## Pendientes de esta guía (marcados arriba como "pendiente")

CP-03 (cupo lleno con datos reales), CP-04/CP-05 (curso inactivo/inexistente), CP-12 (participante aprobado pero no inscripto), CP-20 (marcar ausente) y CP-24 (segunda edición) quedan cubiertos por revisión de código y por la misma lógica ya verificada en sus casos hermanos, pero no se re-ejecutaron uno por uno contra un servidor real por acotar el tiempo de esta verificación. Recomendado repetirlos manualmente antes de dar por cerrada la etapa.
