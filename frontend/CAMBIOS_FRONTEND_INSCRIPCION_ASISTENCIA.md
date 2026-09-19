# Cambios de frontend — Inscripción y Asistencia (2026-09-17)

Documenta lo agregado al frontend Angular para consumir los endpoints de
`education-service` implementados en la etapa anterior (ver
`education/CONTEXTO_INSCRIPCIONES_ASISTENCIA.md` y
`education/GUIA_PRUEBAS_INSCRIPCION_ASISTENCIA.md`). **No se modificó
ninguna pantalla ya existente** — todo lo de abajo es código nuevo, salvo
4 archivos que recibieron cambios puramente aditivos (se detalla exactamente
qué se agregó en cada uno, sin tocar una sola línea de lo que ya había).

## Resumen

3 pantallas nuevas, exclusivas de `role_admin` y `role_volunteer`:

| Pantalla | Ruta | HU/RF | Qué hace |
|---|---|---|---|
| Inscripciones | `/dashboard/inscripciones` | HU-18/RF-18 | Elegir un curso publicado, ver quién está inscripto, inscribir un participante nuevo eligiéndolo por nombre |
| Asistencia por curso | `/dashboard/asistencia-cursos` | HU-19/RF-19 | Elegir curso → actividad → tildar presente/ausente de los inscriptos activos |
| Asistencia general | `/dashboard/asistencia-general` | HU-19bis/RF-64 | Elegir una fecha y marcar Presente/Ausente + observación para todos los participantes aprobados del centro |

Las tres aparecen en el menú lateral solo para `role_admin` y
`role_volunteer` (y `role_dev` por el bypass global del backend).

## Archivos nuevos

**Pantallas** (cada una con `.ts` + `.html` + `.css`, componentes
standalone con PrimeNG, calcados del estilo de `asistencia.component.ts`
y `seguimiento.component.ts` que ya existían):
- `src/app/pages/dashboard/inscripciones/`
- `src/app/pages/dashboard/asistencia-cursos/`
- `src/app/pages/dashboard/asistencia-general/`

**Guard nuevo:**
- `src/app/core/guards/admin-volunteer.guard.ts` — deja pasar
  `role_admin`/`role_volunteer`/`role_dev`, calcado literal de
  `admin.guard.ts` (que ya existía y no se tocó).

**Utilidad nueva:**
- `src/app/shared/utils/date-ymd.ts` — formatea/parsea fechas como
  `"YYYY-MM-DD"` usando getters locales (no `toISOString()`, que corre el
  día en husos horarios negativos). Necesario porque
  `GET/POST /api/asistencia-general` espera ese formato exacto de
  calendario, no un instante — ver `general_attendance.go`.

## Archivos existentes — cambios puramente aditivos

Ninguna línea de lo que ya existía fue borrada ni reescrita en estos 4
archivos; solo se agregó código nuevo.

### `src/app/core/models/education.model.ts`
Se agregaron 3 interfaces al final (antes de las funciones
`richTextToPlain`/`plainToRichText`, que no se tocaron):
`EnrollmentInput`, `GeneralAttendance`, `GeneralAttendanceInput`.

### `src/app/services/education.service.ts`
Se agregaron 6 métodos nuevos en una sección propia
("Inscripción, asistencia por actividad y asistencia general"), antes
del `handleError` privado que ya existía:
- `enrollParticipant(cursoId, input)` → `POST /api/cursos/{id}/inscripciones`
- `listCourseEnrollments(cursoId)` → `GET /api/cursos/{id}/inscripciones`
- `listActivityAttendance(actividadId)` → `GET /api/actividades/{id}/asistencia`
- `registerActivityAttendance(actividadId, inputs)` → `POST /api/actividades/{id}/asistencia`
- `getGeneralAttendance(fecha)` → `GET /api/asistencia-general?fecha=...`
- `registerGeneralAttendance(fecha, inputs)` → `POST /api/asistencia-general?fecha=...`

Los métodos existentes (`listEnrollments`, `listActivities`,
`registerAttendance`, etc., todos bajo `/api/docente/...`) quedaron
intactos — son otro camino del backend (con scope "mis cursos", para el
docente responsable) y la pantalla `asistencia.component.ts` del docente
los sigue usando sin cambios.

### `src/app/app.routes.ts`
Se agregó el `import` del guard nuevo y 3 objetos de ruta nuevos (hijos
de `/dashboard`, junto a las rutas de `cursos`/`mis-cursos`/etc. que ya
estaban), cada uno con `canActivate: [adminOrVolunteerGuard]` y carga
diferida (`loadComponent`). Ninguna ruta existente cambió de path,
componente ni guard.

### `src/app/components/layout/sidebar/sidebar.component.ts`
Se agregaron 3 objetos al array `allItems` (Inscripciones, Asistencia
por curso, Asistencia general, con `roles: ["role_admin", "role_volunteer"]`),
entre "Usuarios" y "Mis cursos". Los demás ítems del menú no se tocaron.

## Decisiones de diseño

- **Selector de participantes por nombre, no por id a mano.** El
  `CONTEXTO_INSCRIPCIONES_ASISTENCIA.md` señalaba que no existe en el
  frontend ningún servicio que resuelva participantes por nombre, y
  sugería que el selector de inscripción "muestre nombre+apellido y
  guarde `participant_id` por dentro" para evitar el error de
  confundirlo con `user_id` (documentado como problema conocido en la
  guía de pruebas). Se resolvió reutilizando
  `GET /api/asistencia-general` (fecha de hoy): es, hoy, el único
  endpoint del backend que devuelve el directorio de participantes
  aprobados con nombre/apellido ya resueltos contra auth-service. Las
  3 pantallas nuevas arman ese directorio una vez al cargar y lo usan
  para mostrar nombres en vez de ids crudos.

- **Asistencia por curso usa los inscriptos reales, no ids a mano.** La
  pantalla del docente (`asistencia.component.html`, no tocada) todavía
  carga participantes escribiendo el id a mano, con la nota "hasta que
  exista el módulo de inscripciones". La pantalla nueva sí puede
  resolverlo: usa `GET /api/cursos/{id}/inscripciones` (inscriptos
  activos del curso) + el directorio de nombres, así que solo hay que
  tildar presente/ausente.

- **Asistencia general distingue "sin registrar" de "ausente".** El
  backend devuelve `presente: null` para quien todavía no tiene
  registro ese día (`GeneralAttendanceView`, ver
  `general_attendance.go`). La pantalla solo manda al backend las filas
  que el usuario tocó explícitamente (Presente o Ausente) — dejar una
  fila "sin registrar" no la pisa con `false`.

## Limitación conocida (documentada, no resuelta)

**No hay endpoint de backend para listar las actividades de un curso
abierto a `role_volunteer`.** Solo existe
`GET /api/docente/cursos/{cursoId}/actividades`, exclusivo de
`role_teacher` interno. `role_admin` igual lo puede usar por el bypass
global de rol que ya tiene `RequireRole(...)` en el backend (ver
`education/CLAUDE.md`), pero `role_volunteer` no tiene ese bypass y
recibe `403`.

La pantalla "Asistencia por curso" lo maneja así: intenta listar las
actividades con ese endpoint; si falla por permisos (403), cae
automáticamente a un campo de texto para ingresar el id de la actividad
a mano (mismo criterio que ya usaba la pantalla del docente antes de
que existiera el módulo de inscripciones), con una nota explicando que
hay que pedirle el id al docente responsable. Para admin, la lista
aparece normalmente.

**Si esto molesta en el uso real**, la solución de fondo es agregar al
backend un endpoint tipo `GET /api/actividades?cursoId=...` en el mismo
grupo de rutas `role_admin`/`role_volunteer` que ya tienen inscripción y
asistencia (no se agregó en esta etapa porque la tarea pedida era
"programar el front" sobre el backend ya cerrado, sin tocarlo).

## Cómo probarlo

Igual que el resto del frontend en desarrollo (sin tocar Docker, ver
memoria `local-dev-frontend-4300-proxy`):

```bash
cd frontend
pnpm start:local   # ng serve -c local, http://localhost:4300
```

Necesitás un usuario `role_admin` o `role_volunteer` logueado, y al
menos un curso publicado + un participante aprobado en auth-service
(ver precondiciones de `education/GUIA_PRUEBAS_INSCRIPCION_ASISTENCIA.md`,
mismas que usa el backend).

`pnpm run build` corrido para verificar: compila sin errores (solo el
warning de presupuesto de bundle inicial que ya existía antes de este
cambio, 1.03 MB de un máximo de error de 1.6 MB — no bloqueante).

## Pendiente (no implementado a propósito, fuera de lo pedido)

- Tests automáticos (`.spec.ts`) de las 3 pantallas nuevas — no se
  pidieron y el resto del módulo de educación en frontend tampoco tiene
  cobertura más allá de los specs por defecto de Angular CLI.
- El endpoint de backend de listado de actividades para `role_volunteer`
  (ver "Limitación conocida" arriba).
