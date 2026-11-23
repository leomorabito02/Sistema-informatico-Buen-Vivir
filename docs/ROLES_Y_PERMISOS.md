# Roles y permisos

Basado en HU-06 y verificado en el código el 11/11/2026 (tests de rutas de
cada servicio y sonda de seguridad, ver `TESTING.md`). Regla general: **el
backend valida siempre**; el menú y los guards del front son solo
navegación.

Roles (`role_id` en el JWT): `role_admin`, `role_dev` (igual que admin),
`role_volunteer`, `role_participant`, `role_company`, `role_teacher`
(con `is_external` para el docente externo). Sin sesión: visitante.

## Matriz

| Rol | Puede | No puede | Pantallas (menú) |
|---|---|---|---|
| Administrador | Todo: usuarios y aprobaciones, cursos/áreas/materiales y su publicación, moderación de contenido docente y comentarios, fichas e hitos, auditoría, ofertas (aprobar, rechazar, cerrar, borrar, resolver postulaciones, ver contacto de empresas en historial), campañas (incluye eliminar), dashboard y reportes | - | Dashboard, Cursos (catálogo, gestión, inscripciones, contenido docente), Asistencia, Participantes, Campañas, Inserción laboral, Auditoría |
| Voluntario | Alta/edición/baja de participantes; inscribir y ver asistencia; fichas, hitos y PDF; campañas (crear, editar, publicar, eliminar); ver ofertas, postular participantes, editar/cerrar/renovar/borrar ofertas, resolver postulaciones y ver contacto de empresas en historial | Crear cursos, aprobar ofertas, dashboard, auditoría, RBAC | Cursos (catálogo, inscripciones), Asistencia, Participantes, Seguimiento, Campañas, Inserción laboral |
| Participante | Sus cursos, actividades y entregas, su progreso, su línea de tiempo, ofertas activas disponibles y sus postulaciones (incluyendo vencidas/cerradas), su CV, preferencias de notificación, comentar | Ver datos de otros, contacto de empresas, fichas, gestionar | Mis cursos, Mi proceso, Ofertas, Mi CV |
| Empresa (aprobada) | Publicar ofertas (nacen pendientes, con vacantes), ver y editar/cerrar/renovar/borrar las suyas, ver y resolver sus postulantes | Aprobar ofertas, ver pendientes de otras empresas, perfiles y fichas | Ofertas laborales, Contacto |
| Docente interno | Sus cursos publicados, actividades, entregas, asistencia, progreso, materiales de sus cursos, habilidades de sus actividades | Fichas, dashboard, gestión de usuarios, moderación | Mis cursos, Asistencia, Progreso, Seguimiento |
| Docente externo | Subir contenido (queda pendiente), ver su estado, editar/borrar mientras está pendiente | Todo lo demás (403 en `/api/docente`) | Mis contenidos |
| Visitante | Campañas publicadas, biblioteca (dejando sus datos), comentarios, mapa | Todo lo autenticado (401) | Landing, campañas públicas, contacto |

## Cómo se aplica en el backend

| Servicio | Mecanismo |
|---|---|
| auth | `RequireRole` por grupo de rutas + controles por dueño en los servicios (`checkParticipantUpdateAccess`, `checkCreateProfileOwnership`, `canApprove`) |
| education | `RequireRole` por grupo, `RequireInternalTeacher`, ownership de cursos para el docente (`VerifyOwnership`) |
| human-promotion | Solo admin/voluntario en fichas; indicadores y auditoría solo admin; `mi-linea-de-tiempo` solo participante |
| campaigns | Públicas sin token; gestión (incluye DELETE) admin/voluntario |
| labor-integration | Identidad del JWT (si hay `JWT_SECRET` descarta los headers del cliente); `puedeGestionarOferta`, `estadoPermitido`, aprobar/rechazar solo admin |
| dashboard, notifications | Dashboard solo admin; notificaciones solo del propio usuario |

## Problemas corregidos el 11/11/2026

1. Labor confiaba en `X-User-Role`/`X-User-ID` del cliente sin token válido (crítico).
2. Editar y borrar ofertas sin control de rol ni dueño; la empresa podía autoaprobarse.
3. Pendientes de ofertas visibles a cualquier rol no participante.
4. Perfiles (participante, empresa, docente, voluntario) creados a nombre de otra cuenta.
5. El participante podía dar de baja su perfil de participante (HU-09).
6. Seguimiento en el menú de admin/voluntario usaba endpoints de docente (403).

## Cambios del 22/11/2026

1. `GET /api/v1/users/{id}/profile` (perfil completo: contacto, DNI, datos del rol) queda solo para la propia cuenta, administración y voluntariado. Antes lo podía leer cualquier usuario logueado. El nombre del docente de un curso lo resuelve Educación (`responsibleTeacherName`).
2. Auditoría de Campañas (`/api/v1/campaigns/auditoria`) y de las acciones de gestión sobre CV (`/api/v1/rxresume/auditoria`): solo administración.
3. Labor acepta `approval_status = "accepted"` (valor viejo) igual que `"approved"` al crear ofertas y postularse.

## Diferencias abiertas con HU-06

- No hay panel de configuración RBAC en el front (solo la API `/roles` y `/permissions`).
- No se auditan los intentos de acceso no autorizados.
- El docente puede crear cursos en borrador (`POST /api/docente/cursos`).
- Las ofertas creadas por admin o voluntario nacen publicadas.
- Rutas del front sin guard: `usuarios`, `seguimiento`, `cv`, `ofertas` (el backend igual rechaza).
