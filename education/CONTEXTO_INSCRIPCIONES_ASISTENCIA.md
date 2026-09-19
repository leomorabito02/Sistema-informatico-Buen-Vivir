
# Contexto — Inscripción y Asistencia (education-service)

Documento para retomar esta tarea en una sesión futura sin tener que
re-explicar el contexto. Última actualización: 2026-09-13.

## Objetivo

Implementar, **solo en el backend** de `education-service` (instrucción
explícita: no tocar frontend en esta etapa), las funcionalidades de:

1. Inscripción de participantes a cursos.
2. Asistencia por actividad/curso, abierta también a `role_volunteer`
   (antes exclusiva de `role_teacher` bajo `/api/docente`).
3. Asistencia general diaria del centro (nueva por completo).
4. Validaciones de negocio sobre participantes (existe/activo,
   inscripto) para las tres, consultando a `auth-service`.

## Historias involucradas

- **HU-18/RF-18** — inscribir participantes a cursos desde el catálogo.
- **HU-19/RF-19** — registrar asistencia mediante cumplimiento de
  actividades.
- **HU-19bis/RF-64** — asistencia general del centro, independiente de
  cursos.
- **HU-20/RF-20** — progreso formativo. **No se tocó**: ya estaba
  completo de antes (`ProgressService`, `GET .../progreso`, y la
  pantalla `seguimiento.component.ts` del frontend que ya lo consume).
  Se revisó como parte del diagnóstico, no como parte de la
  implementación.

## Arquitectura

### Backend (Go, `education/`)

Capas sin cambios: `repositories` (HTTP/Firestore crudo) → `services`
(reglas de negocio) → `controllers` (handlers HTTP) → `routes`.

**Archivos nuevos:**
- `internal/models/general_attendance.go` — `GeneralAttendance`,
  `GeneralAttendanceInput`, `GeneralAttendanceView`.
- `internal/models/auth_participant.go` — `AuthParticipant`, `AuthUser`
  (formas recortadas de lo que devuelve auth-service).
- `internal/repositories/general_attendance_repository.go` — Firestore,
  colección `asistencia_general`.
- `internal/repositories/auth_service_client.go` — cliente HTTP a
  auth-service: `ListParticipants`, `GetParticipant` (lookup puntual,
  usado para validar), `ListUsers`.
- `internal/services/general_attendance_service.go`.
- `internal/controllers/enrollment_controller.go`,
  `attendance_controller.go`, `general_attendance_controller.go`.

**Archivos modificados:**
- `internal/models/course.go` — campo `Capacity *int` en `Course` y
  `CourseInput`.
- `internal/repositories/contentful_fields.go` — helper `rawInt`.
- `internal/repositories/course_repository.go` — lee/escribe
  `fields.capacity`; agrega `ExistsInManagement` (distingue curso
  inactivo de inexistente).
- `internal/repositories/enrollment_repository.go` — agrega `Create`.
- `internal/middlewares/auth_middleware.go` — agrega `TokenKey` al
  contexto (JWT crudo, para reenviar a auth-service).
- `internal/config/config.go` — nueva variable `AUTH_SERVICE_URL`.
- `internal/services/enrollment_service.go` — `Enroll(...)` gana
  validaciones (ver "Reglas de negocio").
- `internal/services/attendance_service.go` — `RegisterBatch(...)` gana
  `cursoID` + validaciones; nuevas dependencias
  (`EnrollmentRepository`, `AuthServiceClient`).
- `internal/controllers/docente_controller.go` — `verifyActivityOwnership`
  devuelve `(*models.Activity, error)` en vez de solo `error`, para
  reusar `activity.CursoID` sin releer Firestore.
- `internal/routes/routes.go` — nuevo grupo `role_admin`/`role_volunteer`.
- `cmd/server/main.go` — wiring de todo lo anterior.

**Tocado fuera de `education/` (con aprobación explícita del usuario):**
- `auth/cmd/server/main.go` — se agregó `role_volunteer` a
  `RequireRole(...)` del bloque `/participants` (antes solo
  `role_participant`). Es la única línea tocada de `auth/`. Ver
  "Reglas de negocio" y "No modificar" más abajo.
- `api-gateway/nginx.conf` — la regex de ruteo a education-service ahora
  incluye `actividades` y `asistencia-general`.
- `api-gateway/openapi.yaml` (gateway de GCP) — 3 rutas nuevas,
  regeneradas con `add_education_routes.py` (raíz del repo, idempotente).

### Frontend

**No implementado — pendiente para una etapa futura.** Instrucción
explícita del usuario en esta etapa: no tocar frontend. Diagnóstico de
lo que hace falta (hecho, no implementado):

- `EducationService` (`frontend/src/app/services/education.service.ts`)
  no tiene métodos para `POST /api/cursos/{id}/inscripciones`,
  `/api/actividades/{id}/asistencia` (la variante sin *ownership*) ni
  `/api/asistencia-general`.
- `education.model.ts` no tiene `EnrollmentInput` ni los tipos
  `GeneralAttendance*`.
- No existe ninguna página `inscripciones/` ni `asistencia-general/`
  bajo `pages/dashboard/`.
- La página `asistencia.component.ts` que ya existe es **solo para
  docente** (usa `listMyCourses`, endpoints de `/api/docente`) — no
  sirve tal cual para `role_volunteer`.
- No existe en todo el frontend un servicio que resuelva
  participantes por nombre (ni siquiera `seguimiento.component.ts`
  muestra nombre, solo `participanteId` crudo) — un selector real de
  participante para la pantalla de inscripción se construye de cero.
- Ya existe el patrón de guard por rol a calcar
  (`core/guards/admin.guard.ts`) para uno nuevo que deje pasar
  `role_admin` **o** `role_volunteer`.

## Base de datos

Sin entidades nuevas fuera de las ya previstas en el DER. Todo vive en
Firestore (`DB Educación — NoSQL Documental`), no en Contentful ni en
una base relacional nueva.

| Colección | Campos clave | Restricción de duplicado |
|---|---|---|
| `inscripciones` | `participante_id`, `curso_id`, `fecha_inscripcion`, `estado` | **Solo de aplicación** — `EnrollmentService.Enroll` lista y compara antes de crear. No hay índice único en Firestore para esto; dos requests simultáneas podrían crear 2 inscripciones activas (riesgo de carrera conocido, no resuelto en esta etapa) |
| `asistencias` | `participante_id`, `actividad_id`, `presente`, `fecha` | **Estructural** — id determinístico `{actividadId}_{participanteId}` (sin cambios de esta etapa) |
| `asistencia_general` | `participante_id`, `fecha` (string `YYYY-MM-DD`), `presente`, `observacion`, `updated_at` | **Estructural** — id determinístico `{fecha}_{participanteId}`, imposible duplicar aunque haya concurrencia |

`Participante` y `Usuario` (nombre/apellido, `approval_status`) viven
en auth-service (Postgres), no acá — se resuelven en tiempo real vía
`AuthServiceClient`, sin caché ni duplicación local.

## Endpoints

Todos nuevos, grupo `role_admin` / `role_volunteer` (sin bypass de
ownership):

| Método | Ruta | Objetivo | Body | Permisos |
|---|---|---|---|---|
| POST | `/api/cursos/{id}/inscripciones` | Inscribir un participante | `{"participanteId":"..."}` | `role_admin`, `role_volunteer` |
| GET | `/api/cursos/{id}/inscripciones` | Listar inscriptos del curso | — | `role_admin`, `role_volunteer` |
| POST | `/api/actividades/{actividadId}/asistencia` | Registrar asistencia (sin *ownership*) | `[{"participanteId":"...","presente":bool}]` | `role_admin`, `role_volunteer` |
| GET | `/api/actividades/{actividadId}/asistencia` | Listar asistencia de la actividad | — | `role_admin`, `role_volunteer` |
| GET | `/api/asistencia-general?fecha=YYYY-MM-DD` | Listado combinado (participantes + asistencia del día) | — | `role_admin`, `role_volunteer` |
| POST | `/api/asistencia-general?fecha=YYYY-MM-DD` | Upsert de asistencia general en lote | `[{"participanteId":"...","presente":bool,"observacion":"..."}]` | `role_admin`, `role_volunteer` |

Respuestas de error: `400` con `{"error": "<mensaje>"}` para toda
validación de negocio (ver tabla de reglas abajo); `404` si el
recurso (actividad) no existe; `403` si el rol no alcanza; `502` si
Contentful/Firestore/auth-service no responden.

## Reglas de negocio implementadas

| Validación | Dónde | Mensaje |
|---|---|---|
| Curso no existe | `EnrollmentService.Enroll` | `"El curso no existe"` |
| Curso existe pero no publicado | `EnrollmentService.Enroll` (usa `CourseRepository.ExistsInManagement` contra la CMA) | `"El curso no está disponible (inactivo o no publicado)"` |
| Participante no existe (auth-service) | `Enroll`, `AttendanceService.RegisterBatch`, `GeneralAttendanceService.RegisterBatch` | `"El participante [<id>] no existe"` |
| Participante no aprobado | ídem | `"El participante [<id>] no está activo"` |
| Inscripción duplicada | `Enroll` | `"El participante ya se encuentra inscripto en el curso"` |
| Cupo completo (`course.Capacity`) | `Enroll` | `"El curso no tiene cupo disponible"` |
| Participante no inscripto en el curso de la actividad | `AttendanceService.RegisterBatch` | `"El participante [<id>] no está inscripto en el curso"` |
| Fecha inválida (asistencia general) | `GeneralAttendanceService` | `"fecha inválida: se espera el formato YYYY-MM-DD"` |
| Asistencia general independiente de asistencia por curso | Colecciones separadas (`asistencia_general` vs `asistencias`), sin ningún código que escriba en una a partir de la otra | — |

**Cómo se resuelve la validación contra auth-service:** cada
service recibe el JWT crudo del usuario que llamó a education
(`middlewares.TokenKey`, puesto por `AuthMiddleware`) y lo reenvía tal
cual a `auth-service` vía `AuthServiceClient.GetParticipant` — no hay
un secreto de servicio propio; auth-service aplica su propio RBAC sobre
ese mismo usuario.

## Estado

### Implementado y verificado
Todo lo de la tabla de endpoints, verificado en vivo (no con tests
automáticos — instrucción explícita) contra: Contentful real (cursos
reales del Space), el emulador de Firestore, y un stub HTTP mínimo de
auth-service (`/api/v1/participants[/:id]`, `/api/v1/users`). Casos
confirmados en vivo: inscripción correcta, duplicado, participante
inexistente/inactivo (en los 3 endpoints), listado de inscriptos,
asistencia por actividad (con participante inscripto y no-activo),
asistencia general completa (listar con nombre resuelto, cargar,
re-consultar sin duplicar, fecha inválida, participante inexistente).
Detalle caso por caso: `education/GUIA_PRUEBAS_INSCRIPCION_ASISTENCIA.md`.

`go build ./...` y `go vet ./...` (código de producción) limpios en
`education/` y en `auth/`.

### Pendiente
- **Frontend completo** — ver sección "Frontend" arriba.
- Repetir a mano (no bloqueante, cubierto por revisión de código): cupo
  lleno con `capacity` real, curso inactivo/inexistente, participante
  aprobado-pero-no-inscripto, marcar ausente, segunda edición de
  asistencia general — ver "Pendientes" en la guía de pruebas.
- **Email/notificaciones** (HU-18, HU-26) — fuera de alcance a
  propósito, no implementado.
- **Integración con línea de tiempo** (HU-19bis) — fuera de alcance a
  propósito, no se tocó ningún componente de timeline.
- **Auditoría** — fuera de alcance a propósito, no se agregó ningún
  registro ni se tocó el sistema de auditoría existente.
- **Tests automáticos** — fuera de alcance a propósito en esta etapa.
  Los mocks existentes (`enrollment_service_test.go`,
  `docente_controller_test.go`) están desactualizados respecto de las
  firmas nuevas (`go vet ./...` los marca; no rompen `go build ./...`
  porque los tests no compilan junto al binario). Cuando se retome:
  actualizar esos mocks + sumar cobertura de todo lo nuevo de esta
  etapa.
- **Rutas de education en `nginx.conf`/`openapi.yaml`**: ya están
  agregadas (`actividades`, `asistencia-general`), pero la SA del
  backend del gateway de GCP todavía necesita `roles/run.invoker`
  sobre `education-*` — ver `docs/PIPELINE_EDUCATION.md` (de una tarea
  anterior, no de esta).
- **`GET /api/v1/participants`/`GET /api/v1/participants/{id}` con
  service accounts o participantes en volumen**: `GeneralAttendanceService`
  hace `ListParticipants` (trae TODOS los participantes) en cada
  `GET /api/asistencia-general`. Para un centro chico es aceptable; si
  la base de participantes crece mucho, conviene paginar o cachear.

### Problemas conocidos
- **El valor real de `approval_status` es `"accepted"`, no `"approved"`**
  (encontrado y corregido 2026-09-13, probando con datos reales de
  Neon): `auth/internal/dto/participant_dto.go` valida el campo con
  `oneof=pending approved rejected`, pero lo que auth-service devuelve
  de verdad en participantes reales es `"accepted"` (o `"pending"`) —
  el validador de auth y el dato real están desalineados (no se tocó
  `auth/` para investigar/corregir esto, fuera de alcance). Los 3
  puntos de esta etapa que comparaban contra `"approved"` (`Enroll`,
  `AttendanceService.RegisterBatch`, `GeneralAttendanceService`) se
  corrigieron para comparar contra `"accepted"`, centralizado en una
  sola constante (`services.participantApprovalStatusAccepted`,
  `internal/services/errors.go`) para que no vuelva a divergir entre
  los 3 servicios. Antes del fix, esto hacía que **cualquier**
  participante real (con `approval_status: "accepted"`) fallara como
  "no está activo" al inscribir/registrar asistencia, aunque
  estuviera perfectamente habilitado.
- **`participanteId` se confunde fácil con `user_id`** (encontrado
  2026-09-13, probando con un id real de Neon): la tabla `participantes`
  tiene `participant_id` (PK propia) y `user_id` (FK a `usuarios`, el
  mismo valor que el `sub` del JWT). Todo lo de esta etapa
  (`Inscripcion`, `Asistencia`, `AsistenciaGeneral` en Firestore, y por
  lo tanto `GetParticipant`) espera **`participant_id`**, no `user_id`.
  `GET /api/v1/participants/{id}` de auth-service busca solo por
  `participant_id` (`WHERE participant_id = ?`) — pasarle el `user_id`
  da 404, que education traduce a `"El participante no existe"` aunque
  el usuario sí exista. No se resolvió con código (sería cambiar el
  contrato de auth-service, fuera de alcance) — se documentó en
  `GUIA_PRUEBAS_INSCRIPCION_ASISTENCIA.md` y se agregó un request de
  Postman (`GET http://localhost:8082/api/v1/participants`, directo a
  auth-service) para poder ver ambos ids de cada fila y no confundirlos
  al probar. **Si esto se vuelve un problema recurrente para quien use
  la futura pantalla de inscripción, considerar que el selector de
  participantes muestre nombre+apellido y guarde `participant_id` por
  dentro, para que nadie tenga que manipular ids a mano.**
- **Condición de carrera en inscripción**: `Enroll` hace
  "leer inscripciones → contar → decidir → crear" sin transacción.
  Firestore no tiene un índice único multi-campo nativo para
  `curso_id + participante_id + estado`. Dos requests simultáneas para
  el mismo participante/curso podrían, en teoría, crear 2 inscripciones
  activas. No se resolvió en esta etapa (bajo riesgo real: la app la
  usan pocas personas del centro a la vez, no un público masivo).
- **`Enroll`/`RegisterBatch` (asistencia por curso y general) ahora
  dependen de que `auth-service` esté arriba** — antes de esta etapa,
  inscripción y asistencia funcionaban sin ninguna dependencia externa
  más que Contentful/Firestore. Es una consecuencia directa de validar
  "participante inexistente/inactivo": si `auth-service` está caído o
  `AUTH_SERVICE_URL` no está configurado, estas 3 escrituras fallan con
  `502` en vez de funcionar de forma aislada.
- **Códigos HTTP de "curso inexistente"/"curso inactivo"**: antes de
  esta etapa, un curso no encontrado devolvía `404` (mapeo genérico de
  `repositories.IsNotFound`). Ahora, dentro de `Enroll` específicamente,
  esos dos casos se convierten en `ValidationError` (`400`) con mensaje
  explícito, para ser consistentes con el resto de las validaciones de
  negocio de la misma función (duplicado, cupo). Es un cambio de
  código de estado a tener en cuenta si algo del lado que consuma esta
  API llegara a depender del `404` viejo (hoy nada lo hace, porque el
  frontend no está tocado).

### Decisiones tomadas
- **Reenviar el JWT del caller a auth-service**, en vez de un secreto
  de servicio propio — decisión tomada explícitamente con el usuario en
  la sesión anterior, para que auth-service siga siendo la única
  autoridad de RBAC sobre sus propios datos.
- **`GET /api/v1/participants` necesitaba `role_volunteer`** en
  auth-service (antes solo `role_participant` + bypass admin/dev) — se
  agregó, con aprobación explícita, como la única línea tocada de
  `auth/`.
- **"curso inactivo" = despublicado**, no un tercer estado en
  Contentful — se distingue de "inexistente" solo consultando la CMA
  (que ve borradores) cuando la CDA (solo publicados) da 404.
- **Validar participante contra auth-service en 3 puntos** (inscripción,
  asistencia por actividad, asistencia general) en vez de solo en
  asistencia general (que es donde ya existía la integración) — pedido
  explícito de la ETAPA 2 de esta tarea ("que en el back estén hechas
  las validaciones").
- **`AttendanceService.RegisterBatch` recibe `cursoID` como parámetro**
  en vez de resolverlo internamente desde Firestore — el controller ya
  tiene la actividad resuelta (para verificar que existe/ownership), así
  se evita una lectura duplicada.

### No modificar
**AUTENTICACIÓN.** Regla absoluta para quien continúe este trabajo: no
tocar login, registro, JWT, refresh tokens, 2FA, Google Login,
contraseñas, middleware de autenticación, roles existentes ni
configuración de autenticación en `auth/`. La única excepción ya hecha
y aprobada explícitamente es la línea de `RequireRole` en
`auth/cmd/server/main.go` descripta arriba — no se tocó ni se debe
tocar nada más de esa carpeta sin pedir permiso primero.
