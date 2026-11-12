# Historias de usuario: estado de implementación

Auditoría del 11/11/2026 (rama develop, commit base cc8dbfb + cambios sin commitear de esta auditoría). Generado por `scripts/informe-auditoria/generar.py` desde `datos.py`: no editar a mano. La fuente de las HU es `docs/referencias/HU DE BUEN VIVIR APP.pdf`.

## Resumen

| Estado | Cantidad |
|---|---|
| Implementado | 41 |
| Parcialmente implementado | 17 |
| No implementado | 2 |
| Implementado con errores | 0 |
| Pendiente de verificación | 4 |
| **Total** | **64** |

El estado es el final, después de las correcciones de la auditoría. Detalle completo en el informe PDF.

## Tabla general

| HU | Módulo | Historia | Estado | Problema o pendiente principal |
|---|---|---|---|---|
| HU-01 | Autenticación y usuarios | Registrar usuarios con nombre, email, contraseña, teléfono y rol | Parcialmente implementado | No se valida que el email sea de dominio gmail (criterio de la HU). |
| HU-02 | Autenticación y usuarios | Iniciar sesión con email y contraseña | Parcialmente implementado | No existe el bloqueo temporal tras 5 intentos fallidos de login (solo hay límite de intentos en el código 2FA y rate limit por IP). |
| HU-02bis | Autenticación y usuarios | Recuperar contraseña por enlace al email | Implementado | Verificar el envío real del email en testing (Resend). |
| HU-03 | Autenticación y usuarios | Validar el JWT en cada solicitud protegida | Parcialmente implementado | Las validaciones fallidas van al log técnico (slog), no a un log de auditoría consultable. |
| HU-04 | Autenticación y usuarios | Expiración configurable y refresh tokens de uso único | Parcialmente implementado | Reutilizar un refresh token ya usado se rechaza, pero no invalida el resto de la sesión (la HU lo pide). |
| HU-05 | Autenticación y usuarios | Autenticación de doble factor (2FA) | Implementado | - |
| HU-06 | Autenticación y usuarios | Control de acceso RBAC con vistas por rol | Parcialmente implementado | (1) No hay panel de configuración de permisos por rol en el front (solo la API /roles y /permissions). (2) No se registran los intentos no autorizados en auditoría. (3) El docente puede crear cursos en borrador por /api/docente/cursos, y la HU dice que no. |
| HU-07 | Autenticación y usuarios | Crear perfiles de participantes | Parcialmente implementado | El backend no exige DNI, dirección ni teléfono del participante (la HU los pide). No hay registro en el log de auditoría. |
| HU-08 | Autenticación y usuarios | Editar perfiles de participantes | Parcialmente implementado | La edición no queda en un log de auditoría con el detalle de campos modificados. |
| HU-09 | Autenticación y usuarios | Eliminar perfiles de participantes (baja lógica) | Parcialmente implementado | La baja no queda en un log de auditoría. |
| HU-10 | Autenticación y usuarios | Log de auditoría de acciones sobre perfiles | No implementado | Las altas, bajas y modificaciones de perfiles (auth) no generan registro de auditoría. Sí existe auditoría inmutable y filtrable en Educación y Promoción Humana. |
| HU-11 | Autenticación y usuarios | Foto y documentación del participante | Parcialmente implementado | Los cambios no quedan en auditoría. |
| HU-12 | Autenticación y usuarios | Aprobar el registro de voluntarios | Parcialmente implementado | Se notifica la aprobación (email 'cuenta_aprobada'), pero no el rechazo. |
| HU-13 | Autenticación y usuarios | Iniciar sesión con Google | Implementado | - |
| HU-14 | Educación | Crear cursos (con materiales opcionales) | Implementado | Antes de la auditoría no se podía crear un curso sin materiales (formulario, backend y Contentful lo exigían) y no se podía subir material desde el formulario del curso. |
| HU-15 | Educación | Editar cursos | Implementado | - |
| HU-16 | Educación | Eliminar cursos (baja lógica) | Implementado | - |
| HU-17 | Educación | Contenido de cursos gestionado en Contentful | Implementado | - |
| HU-18 | Educación | Inscribir participantes a cursos | Pendiente de verificación | El email de confirmación depende de Notificaciones vía Pub/Sub; localmente el emulador había quedado sin suscripciones (corregido) y en testing no se pudo verificar. |
| HU-19 | Educación | Asistencia por cumplimiento de actividades | Implementado | - |
| HU-19bis | Educación | Asistencia general diaria del centro | Implementado | - |
| HU-20 | Educación | Progreso formativo por actividades | Implementado | La pantalla Seguimiento estaba en el menú de admin y voluntario, pero solo usaba endpoints de docente: 403 al voluntario, lista vacía al admin. |
| HU-21 | Educación | Sugerencia de habilidades con IA | Implementado | Diferencia con la HU: se sugiere por actividad y no al crear el curso (decisión del equipo, docs/ia-habilidades.md). |
| HU-22 | Educación | Docente externo consulta contenidos | Implementado | - |
| HU-23 | Educación | Docente externo sube contenido con moderación | Pendiente de verificación | La notificación del resultado depende de Pub/Sub -> Notificaciones; no se pudo verificar en testing. |
| HU-24 | Educación | Comentarios en contenido público | Implementado | - |
| HU-25 | Educación | Registro de datos al descargar de la biblioteca | Implementado | - |
| HU-26 | Educación | Email al inscribirse y recordatorios de sesión | Pendiente de verificación | Depende de Pub/Sub, Cloud Scheduler y Resend; no verificado en testing. |
| HU-27 | Educación | Buscar cursos por nombre | Implementado | - |
| HU-28 | Educación | Buscar cursos por área formativa | Implementado | - |
| HU-29 | Promoción Humana | Crear ficha de proceso personal | Implementado | - |
| HU-30 | Promoción Humana | Editar ficha | Implementado | - |
| HU-31 | Promoción Humana | Campos estructurados de la ficha | Parcialmente implementado | Nombre, apellido y edad se muestran desde el perfil (auth) y las habilidades desde Educación; no son campos editables de la ficha como pide la HU. Es una decisión para no duplicar datos personales, pero difiere del texto de la HU. |
| HU-32 | Promoción Humana | Registrar hitos de avance | Implementado | - |
| HU-33 | Promoción Humana | Línea de tiempo del participante | Implementado | - |
| HU-34 | Promoción Humana | Línea de tiempo propia del participante | Implementado | - |
| HU-35 | Promoción Humana | Auditoría de modificaciones de fichas | Implementado | La HU menciona 'eliminar' fichas: no existe eliminación de fichas (no hay nada que auditar). |
| HU-36 | Promoción Humana | Buscar participantes por nombre | Implementado | - |
| HU-37 | Promoción Humana | Exportar ficha en PDF | Implementado | - |
| HU-38 | Inserción Laboral | Cuenta de empresa con aprobación | Parcialmente implementado | Se notifica la aprobación, no el rechazo. |
| HU-39 | Inserción Laboral | Publicar ofertas laborales | Parcialmente implementado | Solo se exigen título y descripción; la HU pide todos los campos obligatorios (ubicación, modalidad, tipo, salario, horario, requisitos, fecha límite). |
| HU-40 | Inserción Laboral | Ofertas pendientes hasta aprobación | Implementado | Antes: cualquier rol distinto de participante veía las pendientes, y la empresa podía autoaprobarse editando el estado. Excepción vigente: las ofertas creadas por admin o voluntario nacen 'Activa'. |
| HU-41 | Inserción Laboral | Aprobar, rechazar o eliminar ofertas | Parcialmente implementado | Eliminar no notifica a la empresa. Antes de la auditoría cualquier usuario con sesión podía borrar ofertas. |
| HU-42 | Inserción Laboral | Eliminar ofertas rechazadas o publicadas | Parcialmente implementado | Inserción laboral no tiene log de auditoría. |
| HU-43 | Inserción Laboral | Generar CV desde el perfil | Implementado | - |
| HU-44 | Inserción Laboral | Notificación de oferta compatible | No implementado | No hay cruce de compatibilidad entre ofertas y perfiles; los participantes no reciben aviso. |
| HU-45 | Inserción Laboral | Buscar ofertas por título | Implementado | - |
| HU-46 | Inserción Laboral | Buscar ofertas por empresa | Implementado | - |
| HU-47 | Campañas | Crear campañas | Implementado | La imagen es una URL, no se sube un archivo. |
| HU-48 | Campañas | Editar campañas | Implementado | - |
| HU-49 | Campañas | Publicar campañas | Implementado | Antes avisaba dos veces (al crear el borrador y al publicar). |
| HU-50 | Campañas | Ver campañas publicadas sin sesión | Implementado | - |
| HU-51 | Campañas | Buscar campañas por título | Implementado | - |
| HU-52 | Dashboard y reportes | Indicadores clave | Implementado | - |
| HU-53 | Dashboard y reportes | Gráficas de evolución temporal | Implementado | - |
| HU-54 | Dashboard y reportes | Filtros por período y módulo | Implementado | - |
| HU-55 | Dashboard y reportes | Exportar reportes en PDF | Implementado | Permisos de la SA del dashboard en GCP (run.invoker, storage.objectAdmin). |
| HU-56 | Observabilidad | Módulo de observabilidad | Parcialmente implementado | No hay panel propio en una URL privada ni métricas por servicio exportadas por la app; se depende de la consola de GCP. |
| HU-57 | Integraciones externas | Mapa con la ubicación del centro | Implementado | - |
| HU-58 | Integraciones externas | Integración con Gemini para habilidades | Implementado | - |
| HU-59 | Integraciones externas | Servicio externo de CV | Implementado | - |
| HU-60 | Notificaciones | Emails ante eventos clave | Pendiente de verificación | Localmente el emulador no tenía suscripciones y el contenedor de notifications corre sin AUTH_GRPC_URL (stack viejo). En testing no hay variables PUBSUB_PUSH_* en el CD. |
| HU-61 | Notificaciones | Notificaciones push (PWA) | Parcialmente implementado | El push queda inactivo si no hay credenciales de Firebase (localmente no arranca; FIREBASE_PROJECT_ID no existe en el environment de testing). |
| HU-62 | Notificaciones | Configurar qué notificaciones recibir | Implementado | - |

## Detalle por historia

### HU-01: Registrar usuarios con nombre, email, contraseña, teléfono y rol

- **Actor:** Administrador
- **Estado:** Parcialmente implementado
- **Frontend:** pages/register, dashboard/usuarios
- **Backend:** auth: UserService.CreateUser, AuthService.Register
- **Endpoints:** POST /api/v1/users, POST /api/v1/auth/register
- **Persistencia:** PostgreSQL (Neon), tabla users + perfil por rol
- **Tests existentes:** auth services/controllers (validador de contraseña, email duplicado, roles permitidos)
- **Tests que faltan:** Test de dominio de email
- **Problemas detectados:** No se valida que el email sea de dominio gmail (criterio de la HU).
- **Correcciones (11/11/2026):** -
- **Pendiente:** Agregar la validación de dominio gmail en el DTO de registro y en el formulario (o confirmar con el equipo si el criterio sigue vigente).

### HU-02: Iniciar sesión con email y contraseña

- **Actor:** Usuario registrado
- **Estado:** Parcialmente implementado
- **Frontend:** pages/login
- **Backend:** auth: AuthService.Login
- **Endpoints:** POST /api/v1/auth/login
- **Persistencia:** users, refresh_tokens
- **Tests existentes:** auth_service_test, auth_controller_test, k6_auth_load_test.js (CD)
- **Tests que faltan:** Test de bloqueo tras 5 intentos
- **Problemas detectados:** No existe el bloqueo temporal tras 5 intentos fallidos de login (solo hay límite de intentos en el código 2FA y rate limit por IP).
- **Correcciones (11/11/2026):** -
- **Pendiente:** Contador de intentos fallidos por cuenta con bloqueo temporal (por ejemplo 15 minutos) y su test.

### HU-02bis: Recuperar contraseña por enlace al email

- **Actor:** Usuario registrado
- **Estado:** Implementado
- **Frontend:** pages/change-password
- **Backend:** auth: PasswordRecoveryService
- **Endpoints:** POST /api/v1/auth/forgot-password, POST /api/v1/auth/reset-password
- **Persistencia:** password_recoveries (vence a los 30 min, uso único)
- **Tests existentes:** password_recovery_service tests
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** Verificar el envío real del email en testing (Resend).

### HU-03: Validar el JWT en cada solicitud protegida

- **Actor:** Sistema
- **Estado:** Parcialmente implementado
- **Frontend:** interceptors (sesión expirada)
- **Backend:** AuthMiddleware en cada servicio (HS256, type=access, vencimiento)
- **Endpoints:** Todos los protegidos
- **Persistencia:** -
- **Tests existentes:** middlewares_test por servicio; sonda de seguridad: alg=none, otra clave, vencido y refresh rechazados con 401
- **Tests que faltan:** -
- **Problemas detectados:** Las validaciones fallidas van al log técnico (slog), no a un log de auditoría consultable.
- **Correcciones (11/11/2026):** -
- **Pendiente:** Registrar los rechazos de token en una auditoría central (ver HU-10).

### HU-04: Expiración configurable y refresh tokens de uso único

- **Actor:** Responsable de seguridad
- **Estado:** Parcialmente implementado
- **Frontend:** auth.service (refresh)
- **Backend:** JWTService (JWT_EXPIRATION_MINUTES, JWT_REFRESH_EXPIRATION_DAYS), AuthService.Refresh
- **Endpoints:** POST /api/v1/auth/refresh
- **Persistencia:** refresh_tokens (revocado al rotar)
- **Tests existentes:** auth_service_refresh_test
- **Tests que faltan:** Test de reutilización que invalide la sesión
- **Problemas detectados:** Reutilizar un refresh token ya usado se rechaza, pero no invalida el resto de la sesión (la HU lo pide).
- **Correcciones (11/11/2026):** -
- **Pendiente:** Ante reutilización, revocar todos los refresh tokens del usuario (RevokeAllForUser ya existe).

### HU-05: Autenticación de doble factor (2FA)

- **Actor:** Usuario
- **Estado:** Implementado
- **Frontend:** login, mi-perfil
- **Backend:** TwofaService (TOTP / código), dispositivos de confianza
- **Endpoints:** POST /api/v1/auth/verify-2fa, /profile/me/2fa/*
- **Persistencia:** users (2FA), almacén de desafíos
- **Tests existentes:** twofa tests en auth
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-06: Control de acceso RBAC con vistas por rol

- **Actor:** Administrador
- **Estado:** Parcialmente implementado
- **Frontend:** guards (admin, adminOrVolunteer, teacher, participant), menú por rol (header)
- **Backend:** RequireRole en cada servicio + controles por dueño en servicios
- **Endpoints:** Todos
- **Persistencia:** roles, permissions (auth)
- **Tests existentes:** routes_test RBAC por servicio; sonda de seguridad 38/39; tests nuevos de permisos (auth, labor, campaigns, education)
- **Tests que faltan:** E2E por rol en el navegador
- **Problemas detectados:** (1) No hay panel de configuración de permisos por rol en el front (solo la API /roles y /permissions). (2) No se registran los intentos no autorizados en auditoría. (3) El docente puede crear cursos en borrador por /api/docente/cursos, y la HU dice que no.
- **Correcciones (11/11/2026):** Se corrigieron 6 problemas de permisos (ver Problema 4): labor confiaba en headers del cliente, editar y borrar ofertas sin control, perfiles creados a nombre de otro usuario, baja de perfil por el propio participante, Seguimiento roto para admin y voluntario.
- **Pendiente:** Panel RBAC en el front; auditoría de accesos denegados; decidir si el docente puede proponer cursos.

### HU-07: Crear perfiles de participantes

- **Actor:** Administrador, voluntario
- **Estado:** Parcialmente implementado
- **Frontend:** dashboard/usuarios
- **Backend:** UserService.CreateUser (voluntario solo crea participantes), ParticipantService
- **Endpoints:** POST /api/v1/users, POST /api/v1/participants
- **Persistencia:** users, participants
- **Tests existentes:** user_service_test, participant_service_test (+ TestParticipantService_PermisosDeAltaYBaja)
- **Tests que faltan:** Test de campos obligatorios completos
- **Problemas detectados:** El backend no exige DNI, dirección ni teléfono del participante (la HU los pide). No hay registro en el log de auditoría.
- **Correcciones (11/11/2026):** Un usuario sin rol de gestión ya no puede crear un perfil a nombre de otra cuenta.
- **Pendiente:** Hacer obligatorios los campos de la HU en el DTO; auditoría de altas.

### HU-08: Editar perfiles de participantes

- **Actor:** Administrador, voluntario
- **Estado:** Parcialmente implementado
- **Frontend:** dashboard/usuarios
- **Backend:** UserService.UpdateUser, ParticipantService.UpdateParticipant
- **Endpoints:** PUT /api/v1/users/{id}, PUT /api/v1/participants/{id}
- **Persistencia:** users, participants
- **Tests existentes:** user_service_test
- **Tests que faltan:** Test de auditoría con campos modificados
- **Problemas detectados:** La edición no queda en un log de auditoría con el detalle de campos modificados.
- **Correcciones (11/11/2026):** -
- **Pendiente:** Auditoría de modificaciones en auth.

### HU-09: Eliminar perfiles de participantes (baja lógica)

- **Actor:** Administrador, voluntario
- **Estado:** Parcialmente implementado
- **Frontend:** dashboard/usuarios (confirmación)
- **Backend:** ParticipantService.DeleteParticipant (soft delete GORM)
- **Endpoints:** DELETE /api/v1/participants/{id}, DELETE /api/v1/users/{id}
- **Persistencia:** participants.deleted_at
- **Tests existentes:** participant_service_test
- **Tests que faltan:** -
- **Problemas detectados:** La baja no queda en un log de auditoría.
- **Correcciones (11/11/2026):** El propio participante ya no puede dar de baja su perfil de participante (solo admin y voluntario, como dice la HU).
- **Pendiente:** Auditoría de bajas.

### HU-10: Log de auditoría de acciones sobre perfiles

- **Actor:** Administrador
- **Estado:** No implementado
- **Frontend:** dashboard/auditoria (muestra Educación y Promoción Humana)
- **Backend:** auth no tiene auditoría; Educación y PH sí (colección auditoria)
- **Endpoints:** GET /api/auditoria, GET /api/v1/human-promotion/auditoria
- **Persistencia:** Firestore auditoria (education, PH)
- **Tests existentes:** audit tests de education y PH
- **Tests que faltan:** Tests de auditoría en auth
- **Problemas detectados:** Las altas, bajas y modificaciones de perfiles (auth) no generan registro de auditoría. Sí existe auditoría inmutable y filtrable en Educación y Promoción Humana.
- **Correcciones (11/11/2026):** -
- **Pendiente:** Tabla de auditoría en auth (usuario, acción, recurso, fecha) en la misma transacción, y sumarla a la pantalla de auditoría.

### HU-11: Foto y documentación del participante

- **Actor:** Voluntario
- **Estado:** Parcialmente implementado
- **Frontend:** dashboard/usuarios (documentos)
- **Backend:** ParticipantService.UploadDocument (PDF/JPG/PNG/WEBP, 10 MB, MIME real)
- **Endpoints:** /api/v1/participants/{id}/documents
- **Persistencia:** participant_documents + bucket GCS
- **Tests existentes:** participant_document_test
- **Tests que faltan:** -
- **Problemas detectados:** Los cambios no quedan en auditoría.
- **Correcciones (11/11/2026):** -
- **Pendiente:** Auditoría; verificar límite de 5 MB para la foto de perfil.

### HU-12: Aprobar el registro de voluntarios

- **Actor:** Coordinador / administrador
- **Estado:** Parcialmente implementado
- **Frontend:** pages/pending-approval, dashboard/usuarios
- **Backend:** approval_status + canApprove (solo admin/dev)
- **Endpoints:** PUT /api/v1/volunteers/{id}, PUT /api/v1/users/{id}
- **Persistencia:** volunteers.approval_status
- **Tests existentes:** approval_helper tests
- **Tests que faltan:** -
- **Problemas detectados:** Se notifica la aprobación (email 'cuenta_aprobada'), pero no el rechazo.
- **Correcciones (11/11/2026):** -
- **Pendiente:** Email y aviso de rechazo.

### HU-13: Iniciar sesión con Google

- **Actor:** Usuario
- **Estado:** Implementado
- **Frontend:** login (Firebase Auth)
- **Backend:** AuthService (google_auth_id, vincular/desvincular)
- **Endpoints:** POST /api/v1/auth/google-check-credentials, /profile/me/link-google
- **Persistencia:** users.google_auth_id
- **Tests existentes:** auth tests
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-14: Crear cursos (con materiales opcionales)

- **Actor:** Administrador
- **Estado:** Implementado
- **Frontend:** dashboard/gestion-cursos (admin-cursos)
- **Backend:** CourseService.CreateCourse (Contentful CMA)
- **Endpoints:** POST /api/cursos, POST /api/cursos/{id}/publicar, POST /api/materiales[/upload]
- **Persistencia:** Contentful (course, material)
- **Tests existentes:** course_service_test (+ SinMateriales), admin-cursos.spec (+4 casos)
- **Tests que faltan:** E2E de alta completa contra Contentful de testing
- **Problemas detectados:** Antes de la auditoría no se podía crear un curso sin materiales (formulario, backend y Contentful lo exigían) y no se podía subir material desde el formulario del curso.
- **Correcciones (11/11/2026):** Problema 1: materiales opcionales en las tres capas; botón 'Subir material nuevo' dentro del formulario del curso; aviso claro si el curso queda en borrador.
- **Pendiente:** -

### HU-15: Editar cursos

- **Actor:** Administrador
- **Estado:** Implementado
- **Frontend:** admin-cursos (mensaje 'Se modificó el curso correctamente')
- **Backend:** CourseService.UpdateCourse
- **Endpoints:** PUT /api/cursos/{id}
- **Persistencia:** Contentful
- **Tests existentes:** course_service_test, admin-cursos.spec
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-16: Eliminar cursos (baja lógica)

- **Actor:** Administrador
- **Estado:** Implementado
- **Frontend:** admin-cursos (aviso de inscriptos activos)
- **Backend:** CourseService.DeleteCourse (cursos_baja + auditoría)
- **Endpoints:** DELETE /api/cursos/{id}, GET /api/cursos/{id}/inscriptos-activos
- **Persistencia:** Firestore cursos_baja, auditoria
- **Tests existentes:** baja_y_participantes_test
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-17: Contenido de cursos gestionado en Contentful

- **Actor:** Administrador
- **Estado:** Implementado
- **Frontend:** -
- **Backend:** Webhook de Contentful + caché con contenido stale (ADR 0006)
- **Endpoints:** POST /webhooks/contentful
- **Persistencia:** Contentful + caché en memoria
- **Tests existentes:** content_cache tests, contentful_webhook_controller_test
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-18: Inscribir participantes a cursos

- **Actor:** Voluntario, responsable
- **Estado:** Pendiente de verificación
- **Frontend:** dashboard/inscripciones
- **Backend:** EnrollmentService.Enroll (duplicado, cupo, curso dado de baja)
- **Endpoints:** POST/GET /api/cursos/{id}/inscripciones
- **Persistencia:** Firestore inscripciones + outbox
- **Tests existentes:** enrollment_enroll_test, inscripciones.spec
- **Tests que faltan:** Test de punta a punta del email
- **Problemas detectados:** El email de confirmación depende de Notificaciones vía Pub/Sub; localmente el emulador había quedado sin suscripciones (corregido) y en testing no se pudo verificar.
- **Correcciones (11/11/2026):** Emulador de Pub/Sub autoinicializado (Problema 3).
- **Pendiente:** Verificar en testing que llegue el email (eventos + suscripción push de notifications).

### HU-19: Asistencia por cumplimiento de actividades

- **Actor:** Voluntario, docente
- **Estado:** Implementado
- **Frontend:** asistencia-cursos, asistencia (docente)
- **Backend:** Entrega -> Asistencia en una transacción
- **Endpoints:** /api/actividades/*, /api/participants/{id}/asistencia
- **Persistencia:** Firestore entregas, asistencias
- **Tests existentes:** submission_service_test
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-19bis: Asistencia general diaria del centro

- **Actor:** Voluntario, administrador
- **Estado:** Implementado
- **Frontend:** asistencia-general
- **Backend:** GeneralAttendanceService (upsert por fecha) + auditoría
- **Endpoints:** GET/POST /api/asistencia-general
- **Persistencia:** Firestore asistencia_general, auditoria
- **Tests existentes:** general_attendance_service_test
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-20: Progreso formativo por actividades

- **Actor:** Participante, voluntario
- **Estado:** Implementado
- **Frontend:** seguimiento, curso-detalle (participante)
- **Backend:** ProgressService
- **Endpoints:** GET /api/cursos/{id}/progreso (nuevo), /api/docente/cursos/{id}/progreso
- **Persistencia:** Firestore
- **Tests existentes:** progress_controller_test (nuevo), seguimiento.spec (+3)
- **Tests que faltan:** -
- **Problemas detectados:** La pantalla Seguimiento estaba en el menú de admin y voluntario, pero solo usaba endpoints de docente: 403 al voluntario, lista vacía al admin.
- **Correcciones (11/11/2026):** Endpoint GET /api/cursos/{id}/progreso para admin y voluntario; Seguimiento elige los endpoints según el rol y muestra nombres.
- **Pendiente:** -

### HU-21: Sugerencia de habilidades con IA

- **Actor:** Administrador
- **Estado:** Implementado
- **Frontend:** curso-detalle/activity-skills
- **Backend:** SkillSuggestionService (Gemini por Vertex, OpenAI de respaldo)
- **Endpoints:** POST /api/actividades/sugerir-habilidades
- **Persistencia:** Firestore habilidades, actividad_habilidades
- **Tests existentes:** skill_suggestion_service_test
- **Tests que faltan:** -
- **Problemas detectados:** Diferencia con la HU: se sugiere por actividad y no al crear el curso (decisión del equipo, docs/ia-habilidades.md).
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-22: Docente externo consulta contenidos

- **Actor:** Docente externo
- **Estado:** Implementado
- **Frontend:** mis-contenidos
- **Backend:** TeacherContentService
- **Endpoints:** GET /api/biblioteca, /api/contenidos-docentes/mios
- **Persistencia:** Firestore contenidos_docentes
- **Tests existentes:** teacher_content_service_test
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-23: Docente externo sube contenido con moderación

- **Actor:** Docente externo, administrador
- **Estado:** Pendiente de verificación
- **Frontend:** mis-contenidos, contenido-docente
- **Backend:** TeacherContentService (pendiente/aprobado/rechazado)
- **Endpoints:** /api/contenidos-docentes/*
- **Persistencia:** Firestore + bucket
- **Tests existentes:** teacher_content_routes_test
- **Tests que faltan:** Test de punta a punta de la notificación
- **Problemas detectados:** La notificación del resultado depende de Pub/Sub -> Notificaciones; no se pudo verificar en testing.
- **Correcciones (11/11/2026):** Enlaces de las notificaciones corregidos (/dashboard/mis-contenidos).
- **Pendiente:** Verificar la notificación en testing.

### HU-24: Comentarios en contenido público

- **Actor:** Visitante / usuario
- **Estado:** Implementado
- **Frontend:** campanas-publico, cursos
- **Backend:** CommentService (sanitizado, rate limit)
- **Endpoints:** GET/POST /api/comentarios, DELETE (admin)
- **Persistencia:** Firestore comentarios
- **Tests existentes:** comment_service_test, comment_routes_test
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-25: Registro de datos al descargar de la biblioteca

- **Actor:** Visitante
- **Estado:** Implementado
- **Frontend:** landing/library
- **Backend:** TeacherContentService.PublicDownload (valida teléfono)
- **Endpoints:** POST /api/biblioteca/{id}/descargas
- **Persistencia:** Firestore descargas_biblioteca
- **Tests existentes:** teacher_content_service_test
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-26: Email al inscribirse y recordatorios de sesión

- **Actor:** Participante
- **Estado:** Pendiente de verificación
- **Frontend:** -
- **Backend:** Reglas InscripcionCreada y ActividadRecordatorio (notifications), ReminderService (education)
- **Endpoints:** /internal/reminders (Cloud Scheduler)
- **Persistencia:** notifications (Postgres)
- **Tests existentes:** reminder_service_test, rules_test
- **Tests que faltan:** Test de punta a punta
- **Problemas detectados:** Depende de Pub/Sub, Cloud Scheduler y Resend; no verificado en testing.
- **Correcciones (11/11/2026):** -
- **Pendiente:** Configurar el job de Scheduler y verificar envíos.

### HU-27: Buscar cursos por nombre

- **Actor:** Usuario
- **Estado:** Implementado
- **Frontend:** cursos
- **Backend:** FiltrarCursos (sin tildes)
- **Endpoints:** GET /api/cursos?q=
- **Persistencia:** -
- **Tests existentes:** course_service_test
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-28: Buscar cursos por área formativa

- **Actor:** Usuario
- **Estado:** Implementado
- **Frontend:** cursos
- **Backend:** FiltrarCursosPorArea
- **Endpoints:** GET /api/cursos?area=&q=
- **Persistencia:** -
- **Tests existentes:** course_service_test
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-29: Crear ficha de proceso personal

- **Actor:** Administrador, voluntario
- **Estado:** Implementado
- **Frontend:** promocion-humana/ficha
- **Backend:** FichaService (cifrado de campos, una ficha por participante)
- **Endpoints:** POST /api/v1/human-promotion/fichas
- **Persistencia:** Firestore fichas + auditoria + outbox
- **Tests existentes:** ficha tests, events_test
- **Tests que faltan:** Specs de front de la ficha
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-30: Editar ficha

- **Actor:** Administrador, voluntario
- **Estado:** Implementado
- **Frontend:** ficha
- **Backend:** FichaService.Update + auditoría
- **Endpoints:** PUT /api/v1/human-promotion/fichas/{id}
- **Persistencia:** Firestore
- **Tests existentes:** events_test (+ TestFichaActualizada_LlevaCreadoPor)
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** El evento ficha.actualizada ahora lleva creado_por (Problema 3).
- **Pendiente:** -

### HU-31: Campos estructurados de la ficha

- **Actor:** Administrador, voluntario
- **Estado:** Parcialmente implementado
- **Frontend:** ficha
- **Backend:** Ficha: situacion_actual, avances (cifrados)
- **Endpoints:** /fichas
- **Persistencia:** Firestore
- **Tests existentes:** ficha tests
- **Tests que faltan:** -
- **Problemas detectados:** Nombre, apellido y edad se muestran desde el perfil (auth) y las habilidades desde Educación; no son campos editables de la ficha como pide la HU. Es una decisión para no duplicar datos personales, pero difiere del texto de la HU.
- **Correcciones (11/11/2026):** -
- **Pendiente:** Confirmar con el equipo y, si corresponde, actualizar la HU.

### HU-32: Registrar hitos de avance

- **Actor:** Administrador, voluntario
- **Estado:** Implementado
- **Frontend:** ficha
- **Backend:** HitoService
- **Endpoints:** POST/GET/PUT /fichas/{id}/hitos
- **Persistencia:** Firestore hitos
- **Tests existentes:** hito tests
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-33: Línea de tiempo del participante

- **Actor:** Voluntario, responsable
- **Estado:** Implementado
- **Frontend:** ficha (línea de tiempo mezclada)
- **Backend:** ActividadService
- **Endpoints:** GET /fichas/{id}/actividad
- **Persistencia:** Firestore
- **Tests existentes:** actividad tests
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-34: Línea de tiempo propia del participante

- **Actor:** Participante
- **Estado:** Implementado
- **Frontend:** mi-proceso
- **Backend:** mi-linea-de-tiempo (solo hitos visibles, sin datos confidenciales)
- **Endpoints:** GET /api/v1/human-promotion/mi-linea-de-tiempo
- **Persistencia:** Firestore
- **Tests existentes:** participante tests
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-35: Auditoría de modificaciones de fichas

- **Actor:** Administrador
- **Estado:** Implementado
- **Frontend:** dashboard/auditoria
- **Backend:** audit.RecordTx en la misma transacción
- **Endpoints:** GET /api/v1/human-promotion/auditoria
- **Persistencia:** Firestore auditoria
- **Tests existentes:** events_test
- **Tests que faltan:** -
- **Problemas detectados:** La HU menciona 'eliminar' fichas: no existe eliminación de fichas (no hay nada que auditar).
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-36: Buscar participantes por nombre

- **Actor:** Voluntario, responsable
- **Estado:** Implementado
- **Frontend:** promocion-humana
- **Backend:** ParticipanteController.Buscar
- **Endpoints:** GET /api/v1/human-promotion/participantes
- **Persistencia:** auth (directorio)
- **Tests existentes:** participante tests
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-37: Exportar ficha en PDF

- **Actor:** Voluntario, responsable
- **Estado:** Implementado
- **Frontend:** ficha
- **Backend:** pdf_service (mismos permisos que la ficha)
- **Endpoints:** GET /fichas/{id}/pdf
- **Persistencia:** -
- **Tests existentes:** pdf tests
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-38: Cuenta de empresa con aprobación

- **Actor:** Empresa
- **Estado:** Parcialmente implementado
- **Frontend:** pending-approval, ofertas
- **Backend:** approval_status de company; CreateOffer exige empresa aprobada
- **Endpoints:** /api/v1/companies
- **Persistencia:** companies
- **Tests existentes:** company_service_test
- **Tests que faltan:** -
- **Problemas detectados:** Se notifica la aprobación, no el rechazo.
- **Correcciones (11/11/2026):** -
- **Pendiente:** Notificar el rechazo.

### HU-39: Publicar ofertas laborales

- **Actor:** Empresa
- **Estado:** Parcialmente implementado
- **Frontend:** ofertas
- **Backend:** JobOfferService.CreateOffer
- **Endpoints:** POST /api/v1/labor-integration/offers
- **Persistencia:** Firestore job_offers
- **Tests existentes:** job_offer_permisos_test (nuevo)
- **Tests que faltan:** Test de campos obligatorios
- **Problemas detectados:** Solo se exigen título y descripción; la HU pide todos los campos obligatorios (ubicación, modalidad, tipo, salario, horario, requisitos, fecha límite).
- **Correcciones (11/11/2026):** Solo empresa, admin y voluntario pueden crear ofertas (antes cualquier rol).
- **Pendiente:** Validar todos los campos en el DTO y en el formulario.

### HU-40: Ofertas pendientes hasta aprobación

- **Actor:** Sistema
- **Estado:** Implementado
- **Frontend:** ofertas
- **Backend:** isOfferVisibleForRole, estadoPermitido
- **Endpoints:** GET/PUT /offers
- **Persistencia:** Firestore
- **Tests existentes:** job_offer_permisos_test
- **Tests que faltan:** -
- **Problemas detectados:** Antes: cualquier rol distinto de participante veía las pendientes, y la empresa podía autoaprobarse editando el estado. Excepción vigente: las ofertas creadas por admin o voluntario nacen 'Activa'.
- **Correcciones (11/11/2026):** Pendientes solo visibles para gestión y empresa dueña; solo el admin cambia el estado.
- **Pendiente:** Decidir si las ofertas del voluntario también requieren aprobación.

### HU-41: Aprobar, rechazar o eliminar ofertas

- **Actor:** Administrador
- **Estado:** Parcialmente implementado
- **Frontend:** ofertas
- **Backend:** Approve/Reject (solo admin), DeleteOffer
- **Endpoints:** PUT /offers/{id}/approve|reject, DELETE /offers/{id}
- **Persistencia:** Firestore
- **Tests existentes:** job_offer_controller_test, job_offer_permisos_test
- **Tests que faltan:** -
- **Problemas detectados:** Eliminar no notifica a la empresa. Antes de la auditoría cualquier usuario con sesión podía borrar ofertas.
- **Correcciones (11/11/2026):** Borrar: gestión o empresa dueña; resto 403.
- **Pendiente:** Notificar a la empresa al eliminar.

### HU-42: Eliminar ofertas rechazadas o publicadas

- **Actor:** Administrador
- **Estado:** Parcialmente implementado
- **Frontend:** ofertas
- **Backend:** DeleteOffer (baja lógica si tiene postulaciones)
- **Endpoints:** DELETE /offers/{id}
- **Persistencia:** Firestore
- **Tests existentes:** job_offer_permisos_test
- **Tests que faltan:** -
- **Problemas detectados:** Inserción laboral no tiene log de auditoría.
- **Correcciones (11/11/2026):** -
- **Pendiente:** Auditoría de eliminaciones.

### HU-43: Generar CV desde el perfil

- **Actor:** Participante
- **Estado:** Implementado
- **Frontend:** cv
- **Backend:** rxresume + cv-renderer
- **Endpoints:** /api/v1/rxresume/*
- **Persistencia:** rxresume (Postgres)
- **Tests existentes:** rxresume tests (pocos)
- **Tests que faltan:** Más tests de rxresume (38 % de cobertura)
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-44: Notificación de oferta compatible

- **Actor:** Participante
- **Estado:** No implementado
- **Frontend:** -
- **Backend:** OfertaCreadaRule avisa a admin y voluntario
- **Endpoints:** -
- **Persistencia:** -
- **Tests existentes:** rules_test
- **Tests que faltan:** -
- **Problemas detectados:** No hay cruce de compatibilidad entre ofertas y perfiles; los participantes no reciben aviso.
- **Correcciones (11/11/2026):** -
- **Pendiente:** Definir 'compatible' (habilidades / CV) y la regla de notificación.

### HU-45: Buscar ofertas por título

- **Actor:** Usuario
- **Estado:** Implementado
- **Frontend:** ofertas
- **Backend:** FiltrarOfertas
- **Endpoints:** GET /offers?q=
- **Persistencia:** -
- **Tests existentes:** labor services tests
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-46: Buscar ofertas por empresa

- **Actor:** Usuario
- **Estado:** Implementado
- **Frontend:** ofertas
- **Backend:** FiltrarOfertas
- **Endpoints:** GET /offers?empresa=
- **Persistencia:** -
- **Tests existentes:** labor services tests
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-47: Crear campañas

- **Actor:** Responsable, voluntario
- **Estado:** Implementado
- **Frontend:** dashboard/campanas
- **Backend:** CampaignService.Create
- **Endpoints:** POST /api/v1/campaigns
- **Persistencia:** Firestore campanas
- **Tests existentes:** campaign_service_test, routes_test
- **Tests que faltan:** -
- **Problemas detectados:** La imagen es una URL, no se sube un archivo.
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-48: Editar campañas

- **Actor:** Responsable, voluntario
- **Estado:** Implementado
- **Frontend:** campanas
- **Backend:** CampaignService.Update
- **Endpoints:** PUT /api/v1/campaigns/{id}
- **Persistencia:** Firestore
- **Tests existentes:** campaign_service_test
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** Una campaña dada de baja ya no se puede editar.
- **Pendiente:** -

### HU-49: Publicar campañas

- **Actor:** Responsable, voluntario
- **Estado:** Implementado
- **Frontend:** campanas
- **Backend:** CampaignService.Publish
- **Endpoints:** POST /api/v1/campaigns/{id}/publicar
- **Persistencia:** Firestore
- **Tests existentes:** campaign_service_test
- **Tests que faltan:** -
- **Problemas detectados:** Antes avisaba dos veces (al crear el borrador y al publicar).
- **Correcciones (11/11/2026):** El aviso campana.creada sale una sola vez, al publicar.
- **Pendiente:** -

### HU-50: Ver campañas publicadas sin sesión

- **Actor:** Visitante
- **Estado:** Implementado
- **Frontend:** campanas-publico
- **Backend:** ListPublished
- **Endpoints:** GET /api/v1/campaigns/publicas
- **Persistencia:** Firestore
- **Tests existentes:** routes_test
- **Tests que faltan:** Paginación (ver prueba de carga)
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** Las dadas de baja no se muestran.
- **Pendiente:** -

### HU-51: Buscar campañas por título

- **Actor:** Usuario
- **Estado:** Implementado
- **Frontend:** campanas-publico
- **Backend:** listar (sin tildes)
- **Endpoints:** GET /publicas?q=
- **Persistencia:** -
- **Tests existentes:** campaign_service_test
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-52: Indicadores clave

- **Actor:** Administrador
- **Estado:** Implementado
- **Frontend:** indicadores
- **Backend:** dashboard: services/indicadores.go (7 indicadores)
- **Endpoints:** GET /api/v1/dashboard/resumen
- **Persistencia:** agrega de auth, education, labor, PH
- **Tests existentes:** dashboard tests (85,8 %)
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-53: Gráficas de evolución temporal

- **Actor:** Administrador
- **Estado:** Implementado
- **Frontend:** indicadores
- **Backend:** resumen por período
- **Endpoints:** GET /api/v1/dashboard/resumen
- **Persistencia:** -
- **Tests existentes:** dashboard tests
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-54: Filtros por período y módulo

- **Actor:** Administrador
- **Estado:** Implementado
- **Frontend:** indicadores
- **Backend:** ?desde&hasta&modulo&area
- **Endpoints:** GET /api/v1/dashboard/resumen
- **Persistencia:** -
- **Tests existentes:** dashboard tests
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-55: Exportar reportes en PDF

- **Actor:** Administrador
- **Estado:** Implementado
- **Frontend:** indicadores (reportes.service)
- **Backend:** ReporteExportado (Firestore + PDF en Cloud Storage)
- **Endpoints:** /api/v1/dashboard/reportes
- **Persistencia:** Firestore reportes_exportados, bucket
- **Tests existentes:** dashboard tests, reportes.service.spec
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** Permisos de la SA del dashboard en GCP (run.invoker, storage.objectAdmin).

### HU-56: Módulo de observabilidad

- **Actor:** Desarrollador
- **Estado:** Parcialmente implementado
- **Frontend:** -
- **Backend:** Logs JSON estructurados; métricas nativas de Cloud Run / Cloud Logging
- **Endpoints:** /health por servicio
- **Persistencia:** Cloud Logging
- **Tests existentes:** smoke tests del CD
- **Tests que faltan:** -
- **Problemas detectados:** No hay panel propio en una URL privada ni métricas por servicio exportadas por la app; se depende de la consola de GCP.
- **Correcciones (11/11/2026):** -
- **Pendiente:** Dashboard de Cloud Monitoring (errores, latencia, CPU/memoria por servicio) con alertas, o Grafana.

### HU-57: Mapa con la ubicación del centro

- **Actor:** Usuario
- **Estado:** Implementado
- **Frontend:** landing/location, contacto (iframe de Google Maps)
- **Backend:** -
- **Endpoints:** -
- **Persistencia:** -
- **Tests existentes:** specs de landing
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-58: Integración con Gemini para habilidades

- **Actor:** Administrador
- **Estado:** Implementado
- **Frontend:** activity-skills
- **Backend:** Vertex AI Gemini + OpenAI de respaldo
- **Endpoints:** POST /api/actividades/sugerir-habilidades
- **Persistencia:** Firestore
- **Tests existentes:** skill_suggestion_service_test, skills_ai_test
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-59: Servicio externo de CV

- **Actor:** Participante
- **Estado:** Implementado
- **Frontend:** cv
- **Backend:** rxresume / Reactive Resume
- **Endpoints:** /api/v1/rxresume/*
- **Persistencia:** rxresume
- **Tests existentes:** rxresume tests
- **Tests que faltan:** Test del error claro cuando el servicio no está
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -

### HU-60: Emails ante eventos clave

- **Actor:** Usuario
- **Estado:** Pendiente de verificación
- **Frontend:** notificaciones
- **Backend:** notifications: reglas por evento + delivery worker + Resend
- **Endpoints:** /eventos/*, /internal/delivery/flush
- **Persistencia:** Postgres notifications
- **Tests existentes:** rules_test, worker_test, consumers_test
- **Tests que faltan:** Test de punta a punta con Pub/Sub
- **Problemas detectados:** Localmente el emulador no tenía suscripciones y el contenedor de notifications corre sin AUTH_GRPC_URL (stack viejo). En testing no hay variables PUBSUB_PUSH_* en el CD.
- **Correcciones (11/11/2026):** Emulador autoinicializado; enlaces rotos de las notificaciones; sin notificaciones huérfanas por reintento.
- **Pendiente:** Recrear el stack local; verificar suscripciones push en GCP.

### HU-61: Notificaciones push (PWA)

- **Actor:** Usuario con la PWA
- **Estado:** Parcialmente implementado
- **Frontend:** push-notification.service (FCM)
- **Backend:** FirebasePushService
- **Endpoints:** POST /api/v1/users/me/devices
- **Persistencia:** devices
- **Tests existentes:** push_service_test
- **Tests que faltan:** -
- **Problemas detectados:** El push queda inactivo si no hay credenciales de Firebase (localmente no arranca; FIREBASE_PROJECT_ID no existe en el environment de testing).
- **Correcciones (11/11/2026):** -
- **Pendiente:** Configurar credenciales de FCM en testing y probar en un dispositivo.

### HU-62: Configurar qué notificaciones recibir

- **Actor:** Usuario
- **Estado:** Implementado
- **Frontend:** notificaciones (preferencias por categoría y canal)
- **Backend:** NotificationPreference.AllowsChannel
- **Endpoints:** /api/v1/notifications/preferences
- **Persistencia:** Postgres
- **Tests existentes:** notification_service_test, notificaciones.spec
- **Tests que faltan:** -
- **Problemas detectados:** Ninguno
- **Correcciones (11/11/2026):** -
- **Pendiente:** -
