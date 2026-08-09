# Verificación de Historias de Usuario contra los mocks del frontend

Fecha de verificación inicial: 2026-08-07 (estado del repo antes de aplicar los 7 cambios puntuales del punto 1). Actualizado el mismo día tras resolver los gaps detectados (ver sección siguiente).

Criterio: este es un **prototipo frontend sin backend real**. "Cubierta" significa que el mock simula razonablemente la pantalla/flujo/mensaje descrito en los criterios de aceptación (datos hardcodeados, sin persistencia real). Las HU puramente de backend/seguridad sin ningún correlato visual posible se marcan "No cubierta" y se anota que están fuera del alcance de un mock de frontend puro.

---

## Estado actualizado (post-corrección de gaps)

Tras la verificación inicial se resolvieron la mayoría de los gaps directamente sobre los mocks. Resumen numérico actualizado:

| Estado | Antes | Ahora |
|---|---|---|
| Cubierta | 13 | 49 |
| Parcialmente cubierta | 28 | 10 |
| No cubierta | 22 | 4 |
| **Total** | **63** | **63** |

Las 4 HU que siguen "No cubierta" son intrínsecamente de backend/seguridad o requieren infraestructura fuera del alcance de un mock de frontend: **HU-03** (validación de JWT en cada request), **HU-04** (expiración/rotación de refresh tokens), **HU-22** (rol "docente externo" inexistente como entidad propia — requiere modelar un sexto rol distinto de "docente" en todo el sistema de auth) y **HU-61** (notificaciones push reales de una PWA instalada, requiere manifest + Service Worker + permisos del navegador).

### Detalle de lo implementado por HU (respecto del estado inicial)

| HU | Estado anterior | Estado actual | Qué se implementó |
|---|---|---|---|
| HU-01 | Parcial | **Cubierta** | Campo contraseña + validación (8 car., mayúscula, especial), validación de dominio `@gmail.com`, chequeo de email duplicado y mensajes de error en `UserManagement.tsx`. |
| HU-02 | Parcial | **Cubierta** | Mensaje de error genérico y bloqueo simulado tras 5 intentos fallidos en `Login.tsx`. |
| HU-02bis | No cubierta | **Cubierta** | Nuevo `ForgotPassword.tsx` (ruta `/recuperar-password`): solicitud con mensaje genérico, simulación de enlace de un solo uso y formulario de nueva contraseña con política de seguridad. |
| HU-03 | No cubierta | No cubierta | Sin cambios — fuera de alcance de un mock de frontend puro. |
| HU-04 | No cubierta | No cubierta | Sin cambios — fuera de alcance de un mock de frontend puro. |
| HU-05 | Parcial | Parcial | Sin cambios — sigue faltando una pantalla de perfil para activar/desactivar 2FA. |
| HU-06 | Parcial | Parcial | Se agregaron Log de Auditoría y Observabilidad (antes inexistentes); sigue faltando el panel de configuración RBAC y el rol "docente externo" como entidad separada. |
| HU-07 | Parcial | **Cubierta** | Campos DNI, dirección y fecha de nacimiento agregados en `UserManagement.tsx` y `VolunteerUserManagement.tsx`. |
| HU-08 | Parcial | **Cubierta** | Mensaje de éxito exacto "Cambios guardados correctamente" al editar usuario/participante. |
| HU-09 | Parcial | **Cubierta** | Modal de confirmación antes de eliminar, en ambos componentes de gestión de usuarios. |
| HU-10 | No cubierta | Parcial | Nuevo `AuditLog.tsx` (ruta `/log-auditoria`, solo admin) con tabla filtrable por usuario/fecha/acción; los datos siguen siendo mock estáticos, no se generan dinámicamente por cada acción real de la sesión. |
| HU-11 | Parcial | **Cubierta** | Zona de carga de documentación (PDF/JPG/JPEG/PNG, 10MB) agregada, separada de la foto de perfil. |
| HU-12 | Parcial | Parcial | El modal de cambio de estado ahora se llama "Aprobar/Rechazar" y muestra un toast simulando la notificación al usuario; sigue sin ser un flujo de aprobación dedicado con cola de pendientes. |
| HU-13 | Parcial | Parcial | Sin cambios. |
| HU-14 | Cubierta | Cubierta | Sin cambios. |
| HU-15 | Parcial | **Cubierta** | Formulario de edición conectado, mensaje "Se modificó el curso correctamente". |
| HU-16 | No cubierta | **Cubierta** | Botón eliminar curso (admin), confirmación con aviso de inscriptos activos, baja lógica y mensaje "Se eliminó el curso correctamente". |
| HU-17 | Parcial | Parcial | Sin cambios. |
| HU-18 | Parcial | **Cubierta** | Mensajes exactos "Se inscribe al participante correctamente" y "El participante ya se encuentra inscripto en el curso". |
| HU-19 | Parcial | **Cubierta** | Actividades del curso ahora son interactivas (marcar/desmarcar completadas) para voluntario/docente/admin. |
| HU-20 | Cubierta | Cubierta | Sin cambios. |
| HU-21 | Parcial | **Cubierta** | Botones de aceptar habilidad sugerida (admin) y mención explícita a Gemini con fallback a OpenAI. |
| HU-22 | No cubierta | No cubierta | Sin cambios — requiere el rol "docente externo" como entidad propia. |
| HU-23 | Parcial | Parcial | Sin cambios — sigue mezclado con la vista de docente responsable. |
| HU-24 | No cubierta | **Cubierta** | Sección de comentarios (nombre y apellido, sin cuenta) agregada en `PublicCampaigns.tsx`. |
| HU-25 | No cubierta | **Cubierta** | Formulario de contacto (nombre, apellido, teléfono) obligatorio antes de descargar/ver cada material de la biblioteca. |
| HU-26 | Parcial | Parcial | `NotificationCenter.tsx` ahora está enrutado y accesible; sigue sin simularse el email de confirmación al inscribirse a un curso. |
| HU-27 | Parcial | **Cubierta** | Mensaje "No se encontraron cursos" agregado al catálogo. |
| HU-28 | Cubierta | Cubierta | Sin cambios. |
| HU-29 | No cubierta | **Cubierta** | Botón "Nueva Ficha" con formulario y validación de obligatorios (mensaje "Completar todos los campos obligatorios para continuar"). |
| HU-30 | No cubierta | **Cubierta** | Edición de ficha con mensaje "Ficha personal modificada correctamente". |
| HU-31 | Parcial | **Cubierta** | Campos estructurados nombre, apellido, edad, habilidades, situación personal y avances; nota de formato Día/Mes/Año. |
| HU-32 | Cubierta | Cubierta | Sin cambios. |
| HU-33 | Cubierta | Cubierta | Sin cambios. |
| HU-34 | Cubierta | Cubierta | Sin cambios. |
| HU-35 | No cubierta | Parcial | Ver HU-10: existe la pantalla de log, pero no está conectada a cada edición real de fichas. |
| HU-36 | Cubierta | Cubierta | Sin cambios. |
| HU-37 | Parcial | **Cubierta** | Botón "Exportar PDF" conectado a una acción simulada con confirmación. |
| HU-38 | No cubierta | **Cubierta** | `Register.tsx`: nota explícita de aprobación pendiente para el rol empresa + pantalla de confirmación post-registro diferenciada por rol. |
| HU-39 | Parcial | **Cubierta** | Campo "Fecha límite de postulación" agregado y validado como obligatorio. |
| HU-40 | No cubierta | **Cubierta** | Las ofertas se crean con estado "Pendiente" y no quedan visibles públicamente hasta su aprobación. |
| HU-41 | No cubierta | **Cubierta** | Botones "Aprobar"/"Rechazar" para admin sobre ofertas pendientes, con notificación simulada. |
| HU-42 | No cubierta | **Cubierta** | Botón eliminar oferta con confirmación y mensaje "Oferta laboral eliminada correctamente". |
| HU-43 | Parcial | **Cubierta** | Se nombra explícitamente "RxResume" como servicio de generación de CV. |
| HU-44 | Parcial | **Cubierta** | `NotificationCenter.tsx` enrutado (`/notificaciones`) y accesible desde el ícono de campana del Dashboard. |
| HU-45 | Cubierta | Cubierta | Sin cambios. |
| HU-46 | Cubierta | Cubierta | Sin cambios. |
| HU-47 | Parcial | **Cubierta** | Formulario conectado, mensaje "Campaña creada correctamente". |
| HU-48 | Parcial | **Cubierta** | Formulario conectado, mensaje "Campaña editada correctamente". |
| HU-49 | Cubierta | Cubierta | Sin cambios. |
| HU-50 | Cubierta | Cubierta | Sin cambios. |
| HU-51 | Cubierta | Cubierta | Sin cambios. |
| HU-52 | Parcial | **Cubierta** | Dashboard admin ajustado a los 5 indicadores exactos: participantes activos, cursos en desarrollo, cursos completados, habilidades desarrolladas, inserciones laborales. |
| HU-53 | No cubierta | **Cubierta** | Gráfico de barras de evolución mensual (6 meses) por indicador, seleccionable. |
| HU-54 | No cubierta | **Cubierta** | Filtros de período y módulo agregados (afectan la actividad reciente mostrada). |
| HU-55 | No cubierta | **Cubierta** | Botón "Exportar PDF" del dashboard con confirmación simulada. |
| HU-56 | No cubierta | **Cubierta** | Nuevo `Observability.tsx` (ruta `/observabilidad`, solo admin) con métricas mock de estado/latencia/error por microservicio. |
| HU-57 | Cubierta | Cubierta | Sin cambios. |
| HU-58 | Parcial | **Cubierta** | Mención explícita a Gemini (con fallback a OpenAI) junto a las habilidades sugeridas. |
| HU-59 | Parcial | **Cubierta** | Igual que HU-43 (RxResume) + simulación de error cuando el servicio no está disponible. |
| HU-60 | No cubierta | Parcial | `NotificationCenter.tsx` cubre las notificaciones in-app; sigue sin simularse el envío de email en los flujos de inscripción/oferta/moderación. |
| HU-61 | No cubierta | No cubierta | Sin cambios — requiere manifest PWA y Service Worker real, fuera de alcance de este mock. |
| HU-62 | No cubierta | **Cubierta** | Nuevo `NotificationSettings.tsx` (ruta `/configuracion-notificaciones`) con toggles de email/push por categoría, conectado al ícono de Settings del Dashboard. |

---

## Detalle original (pre-implementación)

Las tablas siguientes documentan el estado **antes** de aplicar las correcciones — se conservan como referencia histórica del diagnóstico inicial.

Hallazgos transversales importantes (aplican a varias HU, se referencian abajo):
- El modelo de roles (`src/app/types/index.ts`, enum `Role`, y `src/app/App.tsx`) sólo contempla **5 roles**: `admin`, `participante`, `voluntario`, `empresa`, `docente`. **No existe el rol "docente externo"** como entidad separada del "docente responsable de curso" que pide HU-06. `src/app/components/TeacherPortal.tsx` mezcla responsabilidades de ambos roles (gestiona "Mis Cursos" con alumnos, propio de docente responsable, y "Subir Material" con moderación pendiente/aprobada, propio de docente externo) bajo una única vista accesible solo con `userRole === 'docente'`.
- No existe ningún **log de auditoría** en ningún componente. Solo hay un botón decorativo sin `onClick` ("Ver log de auditoría") en `src/app/components/UserManagement.tsx`.
- No existe **panel de configuración RBAC**, ni **módulo de observabilidad**, ni pantallas de **filtros/exportación de dashboard**, ni **gráficas de evolución temporal**.
- `src/app/components/NotificationCenter.tsx` y `src/app/components/PersonalPromotion.tsx` están **implementados y exportados** en `src/app/components/index.ts` pero **no están enrutados en `src/app/App.tsx`** — son componentes huérfanos, inaccesibles desde la navegación real de la app.
- Ningún formulario de creación/edición (usuarios, cursos, campañas, fichas) muestra los mensajes de éxito exactos pedidos por las HU; en la mayoría de los casos los formularios ni siquiera tienen `onSubmit`/`onClick` conectado a una acción simulada con feedback visual (toast/alert), solo cierran el modal o no hacen nada.
- No hay validaciones de campos obligatorios con mensajes de error visibles en ningún formulario (solo el atributo HTML `required`, sin mensaje personalizado).

---

## 1. Autenticación y Gestión de Usuarios

| HU | Estado | Si falta algo | Notas |
|---|---|---|---|
| HU-01 | Parcialmente cubierta | `src/app/components/UserManagement.tsx` (modal "Crear Nuevo Usuario"): agregar campo contraseña, validación de dominio gmail, política de contraseña (8 car., 1 mayúscula, 1 especial), validación de email duplicado y mensajes de error por campo faltante. | El modal de creación de usuario **no tiene campo de contraseña**. No hay validación de dominio gmail ni de política de seguridad de contraseña. No hay chequeo de email duplicado. |
| HU-02 | Parcialmente cubierta | `src/app/components/Login.tsx`: agregar mensaje de error genérico ante credenciales inválidas y lógica de bloqueo tras 5 intentos (aunque sea simulada con estado local). | El formulario de login no valida nada: cualquier email/contraseña redirige. No hay mensaje de error "genérico" simulado ni bloqueo tras intentos fallidos. |
| HU-02bis | No cubierta | Falta una pantalla/flujo de "Olvidé mi contraseña" (nuevo componente, p.ej. `src/app/components/ForgotPassword.tsx`) y su ruta en `App.tsx`. | En `Login.tsx` el enlace "¿Olvidaste tu contraseña?" apunta a `href="#"` sin ningún flujo detrás. |
| HU-03 | No cubierta | Sin correlato visual posible (validación de JWT en cada request). | Fuera de alcance de un mock de frontend puro (backend/seguridad). |
| HU-04 | No cubierta | Sin correlato visual posible (expiración de JWT, refresh tokens de uso único). | Fuera de alcance de un mock de frontend puro (backend/seguridad). |
| HU-05 | Parcialmente cubierta | Falta pantalla de perfil/configuración para activar/desactivar 2FA, y manejo de código incorrecto/expirado. | `Login.tsx` simula el paso de 2FA (pantalla de código de 6 dígitos) pero siempre se muestra tras cualquier login y siempre "funciona"; no hay opción de activar/desactivar 2FA desde un perfil de usuario (no existe pantalla de perfil). |
| HU-06 | Parcialmente cubierta | Faltan: panel de configuración RBAC, log de auditoría, módulo de observabilidad (los tres mencionados explícitamente en la vista de administrador). Falta el rol "docente externo" diferenciado del "docente". | `App.tsx` protege rutas por rol razonablemente para lo que existe, pero el rol "docente externo" no existe como tal (ver hallazgo transversal). La empresa no tiene pantalla de "estado de sus propias ofertas" con pendiente/publicada/rechazada (ver HU-40/41). El voluntario si accede a `/gestion-participantes` puede crear/editar/eliminar participantes sin restricción visible de aprobación previa de su propia cuenta. |
| HU-07 | Parcialmente cubierta | `src/app/components/VolunteerUserManagement.tsx` y `UserManagement.tsx`: agregar campos DNI, dirección y fecha de nacimiento al formulario (falta "apellido" como campo separado también). | Los formularios de creación de participante solo piden Nombre, Email, Teléfono (y en UserManagement, Rol). Faltan dni, dirección y fecha de nacimiento exigidos por el criterio de aceptación. No hay registro en log de auditoría (no existe el log). |
| HU-08 | Parcialmente cubierta | Agregar mensaje de éxito "Cambios guardados correctamente" tras guardar en `UserManagement.tsx` / `VolunteerUserManagement.tsx`. | Ningún formulario de edición de usuario muestra un mensaje de éxito. `VolunteerUserManagement.handleSaveUser` solo cierra el modal sin feedback. |
| HU-09 | Parcialmente cubierta | `UserManagement.tsx`: el botón "Eliminar" (ícono `Trash2`) no tiene `onClick`. `VolunteerUserManagement.tsx`: `handleDeleteUser` elimina directamente sin ningún modal de confirmación. | El criterio de aceptación exige confirmación explícita antes de eliminar; en `VolunteerUserManagement` se borra sin confirmar, y en `UserManagement` el botón de eliminar directamente no hace nada. Ninguna acción queda en log de auditoría. |
| HU-10 | No cubierta | Crear un componente de log de auditoría (p.ej. `src/app/components/AuditLog.tsx`) y una ruta protegida solo para admin. | Existe el tipo `AuditLog` en `src/app/types/index.ts` y un botón decorativo "Ver log de auditoría" en `UserManagement.tsx`, pero no hay ninguna pantalla ni dato mock que lo represente. |
| HU-11 | Parcialmente cubierta | `UserManagement.tsx`: agregar una zona de carga separada para "Documentación (DNI, certificados)" con formatos PDF/JPG/JPEG/PNG y 10MB, distinta de la foto de perfil. | Existe zona de carga de "Foto de Perfil" (PNG/JPG hasta 5MB, coincide con el criterio). No existe una segunda zona de carga de documentación con los formatos/tamaño exigidos (solo hay un textarea genérico "Notas / Documentación"). |
| HU-12 | Parcialmente cubierta | Falta un flujo específico de aprobar/rechazar voluntario con notificación simulada, y bloqueo real de funcionalidades mientras está pendiente. | `UserManagement.tsx` tiene un modal genérico "Cambiar Estado" (Activo/Pendiente) aplicable a cualquier usuario, no es un flujo de aprobación dedicado con aprobar/rechazar. En `App.tsx` el estado "pendiente" no bloquea ninguna ruta (el rol se asigna directo al loguear). No hay notificación al voluntario del resultado. |
| HU-13 | Parcialmente cubierta | Simular en `Login.tsx` la distinción entre cuenta nueva vs. vinculada, y el guardado del `googleId` (aunque sea visualmente). | El botón "Google" en `Login.tsx` (`handleGoogleLogin`) loguea directamente con el rol seleccionado en el dropdown "Demo", sin ninguna simulación de creación/vinculación de cuenta. |

## 2. Educación

| HU | Estado | Si falta algo | Notas |
|---|---|---|---|
| HU-14 | Cubierta | — | `src/app/components/CourseCatalog.tsx` modal "Crear Nuevo Curso" (solo admin) tiene título, descripción, área formativa, docente responsable, duración, cupos y carga de materiales. Nivel mock razonable. |
| HU-15 | Parcialmente cubierta | Agregar mensaje de éxito exacto "Se modificó el curso correctamente" y conectar el formulario a una acción (hoy el `<form>` del modal de edición no tiene `onSubmit`). | El modal "Editar Curso" existe con los campos pedidos, pero el botón "Guardar Cambios" no dispara ninguna acción ni mensaje. |
| HU-16 | No cubierta | Agregar botón "Eliminar curso" en `CourseCatalog.tsx` (visible solo para admin) con modal de confirmación, aviso de inscriptos activos y mensaje "Se eliminó el curso correctamente". | No existe ninguna opción de eliminar/dar de baja un curso en el catálogo. |
| HU-17 | Parcialmente cubierta | — | Se menciona textualmente "El contenido del curso se gestiona a través de Contentful (CMS headless)" en los modales de creación/edición de `CourseCatalog.tsx`, pero es solo un párrafo informativo, no hay ninguna simulación de fallo/stale content. |
| HU-18 | Parcialmente cubierta | Agregar mensajes exactos "Se inscribe al participante correctamente" y "El participante ya se encuentra inscripto en el curso" en `CourseDetail.tsx`; y un flujo donde voluntario/responsable elija a qué participante inscribe (hoy solo hay auto-inscripción). | `CourseDetail.tsx` tiene un botón "Inscribirse al Curso" con modal de confirmación, pero es autoinscripción del usuario logueado (participante/voluntario), no "inscribir a un participante" desde una lista; no hay chequeo de duplicados ni los textos de mensaje pedidos. |
| HU-19 | Parcialmente cubierta | Conectar los checkboxes de actividades en `CourseDetail.tsx` a una acción real (hoy son datos hardcodeados `completed: true/false`, no interactivos) y agregar una vista de historial de asistencia por participante/curso. | La estructura visual de módulos/actividades con estado completado existe, pero no es interactiva ni hay pantalla de consulta de historial de asistencia. |
| HU-20 | Cubierta | — | `CourseDetail.tsx` calcula % de progreso sobre actividades completadas; `ProgressTracking.tsx` muestra progreso por curso. |
| HU-21 | Parcialmente cubierta | Agregar botones de aceptar/descartar habilidades sugeridas (rol admin) y mención de Gemini/OpenAI como fallback. | `CourseCatalog.tsx` y `ProgressTracking.tsx` muestran "habilidades sugeridas" como tags estáticos con nota genérica "generadas por IA", sin flujo de aceptar/descartar ni mención de Gemini/OpenAI. |
| HU-22 | No cubierta | Crear una vista de "docente externo" separada (no existe el rol). | No hay rol "docente externo" ni pantalla de consulta de contenidos restringida a ese perfil (ver hallazgo transversal). |
| HU-23 | Parcialmente cubierta | Separar el flujo de "docente externo" del de "docente responsable" (hoy ambos están mezclados en `TeacherPortal.tsx`); agregar estado "Rechazado" (hoy solo hay Aprobado/Pendiente) y notificación de resultado de moderación. | `TeacherPortal.tsx` tiene "Subir Material" con estado Aprobado/Pendiente, que cubre parcialmente el concepto de moderación, pero está disponible solo bajo `userRole === 'docente'` (responsable de curso), mezclado con gestión de "Mis Cursos" y alumnos — viola la separación de accesos de HU-06 entre docente y docente externo. |
| HU-24 | No cubierta | Agregar sección de comentarios (nombre/apellido + texto) en `PublicCampaigns.tsx` y en la biblioteca de recursos de `LandingPage.tsx`, con moderación para admin. | No existe ninguna funcionalidad de comentarios en contenido público. |
| HU-25 | No cubierta | Agregar formulario (nombre, apellido, teléfono) antes de habilitar la descarga en la sección "Biblioteca de Recursos Comunitarios" de `LandingPage.tsx`. | Los botones "Descargar Material" / "Ver Vídeo" en `LandingPage.tsx` no tenían `onClick`; no había formulario de datos de contacto. |
| HU-26 | Parcialmente cubierta | Conectar el flujo de inscripción (`CourseDetail.tsx`) con una confirmación de "email enviado" simulada; enrutar `NotificationCenter.tsx` en `App.tsx`. | `NotificationCenter.tsx` modela notificaciones tipo curso/recordatorio pero es un componente huérfano no enrutado; no hay ninguna simulación de envío de email en la inscripción a curso. |
| HU-27 | Parcialmente cubierta | Agregar mensaje "No se encontraron cursos" cuando la búsqueda no tiene resultados en `CourseCatalog.tsx` (a diferencia de `PublicJobOffers.tsx`/`PublicCampaigns.tsx`, que sí lo muestran). | El buscador por nombre/descripción funciona (filtra `mockCourses`), pero no hay mensaje de "sin resultados". |
| HU-28 | Cubierta | — | El filtro por área formativa (`Dropdown`) se combina correctamente con la búsqueda por texto en `CourseCatalog.tsx`. |

## 3. Promoción Humana

| HU | Estado | Si falta algo | Notas |
|---|---|---|---|
| HU-29 | No cubierta | Agregar botón "Nueva Ficha" y modal de creación con campos obligatorios en `src/app/components/ParticipantTracking.tsx`, con mensaje "Completar todos los campos obligatorios para continuar". | `ParticipantTracking.tsx` solo permite ver timeline y agregar hitos sobre participantes ya existentes (datos hardcodeados); no existe flujo de alta de una ficha nueva. |
| HU-30 | No cubierta | Agregar modal de edición de ficha con mensaje de éxito "Ficha personal modificada correctamente". | No existe el concepto de "ficha" editable como entidad propia; solo existen tarjetas de participante de solo lectura y el modal de hitos. |
| HU-31 | Parcialmente cubierta | Convertir las tarjetas de participante de `ParticipantTracking.tsx` en un formulario editable con todos los campos estructurados (nombre, apellido por separado, edad, habilidades, situación personal, avances) y formato de fecha DD/MM/AAAA. | Las tarjetas muestran nombre, edad, situación actual y habilidades, pero son de solo lectura, no un formulario; no hay campo "apellido" separado ni "avances" como campo propio; el input de fecha del modal de hitos usa `<input type="date">` nativo del navegador, sin garantizar formato DD/MM/AAAA. |
| HU-32 | Cubierta | — | El modal "Registrar Nuevo Hito" / "Editar Hito" en `ParticipantTracking.tsx` tiene exactamente fecha, título, descripción y responsable. |
| HU-33 | Cubierta | — | El modal "Línea de Tiempo" en `ParticipantTracking.tsx` muestra los hitos ordenados cronológicamente con fecha, descripción y responsable. |
| HU-34 | Cubierta | — | `ProgressTracking.tsx` (rol participante) tiene un modal "Cronología de Avances" con su propio historial de actividad (inscripciones, finalizaciones), sin datos confidenciales. |
| HU-35 | No cubierta | Ver HU-10 (mismo log de auditoría, sin implementación). | No existe log de auditoría en ningún lado del sistema. |
| HU-36 | Cubierta | — | El buscador de `ParticipantTracking.tsx` filtra participantes por nombre. |
| HU-37 | Parcialmente cubierta | Conectar el botón "Exportar PDF" del modal de timeline en `ParticipantTracking.tsx` a una acción simulada (ej. `alert()` como hace `CVGenerator.tsx`). | El botón "Exportar PDF" existe visualmente pero no tiene `onClick`, no genera ningún feedback. |

## 4. Inserción Laboral

| HU | Estado | Si falta algo | Notas |
|---|---|---|---|
| HU-38 | No cubierta | Agregar en `Register.tsx` una nota específica para el rol "empresa" indicando que la cuenta queda pendiente de aprobación (hoy solo dice "Colabora con oportunidades laborales", sin mención de aprobación, a diferencia de voluntario/docente que sí tienen nota); simular bloqueo de publicación mientras está pendiente. | No hay ningún estado "pendiente de aprobación" específico para cuentas empresa en ningún componente; `LaborInsertion.tsx` permite a cualquier `userRole === 'empresa'` publicar ofertas sin restricción. |
| HU-39 | Parcialmente cubierta | `src/app/components/LaborInsertion.tsx`: agregar campo "Fecha límite" (deadline) al modal "Publicar Nueva Oferta" — falta por completo. | El modal tiene título, descripción, ubicación, modalidad, tipo, salario, horario y requisitos, pero no tiene fecha límite/cierre de postulación. Tampoco valida todos los campos obligatorios (solo título y ubicación deshabilitan el botón). |
| HU-40 | No cubierta | En `LaborInsertion.tsx`, `handleCreateJob` debería crear la oferta con estado `pendiente` en vez de `'Activa'` directo. | Las ofertas se crean y quedan visibles inmediatamente como "Activa"; no existe el estado "pendiente" ni el paso de aprobación antes de la visibilidad pública. |
| HU-41 | No cubierta | Agregar en `LaborInsertion.tsx` (vista admin) botones de aprobar/rechazar sobre ofertas pendientes. | No existe ninguna acción de aprobar/rechazar ofertas laborales en la interfaz. |
| HU-42 | No cubierta | Agregar botón "Eliminar oferta" en `LaborInsertion.tsx` con mensaje "Oferta laboral eliminada correctamente". | No existe ninguna funcionalidad de eliminación de ofertas laborales. (Nota: el texto de la HU dice "aparece un cartel de error" pero el mensaje citado es de éxito — inconsistencia en la propia HU, a tener en cuenta al implementar). |
| HU-43 | Parcialmente cubierta | Mencionar explícitamente "RxResume" en `src/app/components/CVGenerator.tsx` (hoy el texto es genérico). | `CVGenerator.tsx` compila datos de perfil/cursos/habilidades y genera una vista previa de CV descargable; `handleDownloadPDF` hace `alert('Generando PDF... Esta funcionalidad se integrará con un servicio de generación de CVs')`, sin nombrar RxResume. |
| HU-44 | Parcialmente cubierta | Enrutar `NotificationCenter.tsx` en `App.tsx` para que sea alcanzable. | Existe un ítem mock "Nueva Oferta Laboral... coincide con tu perfil" en `NotificationCenter.tsx`, pero el componente no está enrutado (huérfano). |
| HU-45 | Cubierta | — | `PublicJobOffers.tsx` y `LaborInsertion.tsx` filtran por título de la oferta. |
| HU-46 | Cubierta | — | El mismo buscador de `PublicJobOffers.tsx` matchea también por nombre de empresa (`offer.company`), cumpliendo el criterio aunque comparte el mismo input que la búsqueda por título. |

## 5. Campañas y Concientización

| HU | Estado | Si falta algo | Notas |
|---|---|---|---|
| HU-47 | Parcialmente cubierta | Conectar el `<form>` del modal "Nueva Campaña" en `src/app/components/CampaignManagement.tsx` a una acción y agregar el mensaje "Campaña creada correctamente". | El modal tiene imagen, texto (descripción/contenido) y fecha, pero el formulario no tiene `onSubmit` ni produce ningún mensaje de éxito. |
| HU-48 | Parcialmente cubierta | Igual que HU-47 pero para el modo edición: agregar mensaje "Campaña editada correctamente". | Mismo modal reutilizado para editar, mismo problema: sin acción conectada ni mensaje. |
| HU-49 | Cubierta | — | El campo "Estado" (Borrador/Publicada) del modal de `CampaignManagement.tsx` representa la publicación. |
| HU-50 | Cubierta | — | `PublicCampaigns.tsx` es una ruta pública sin autenticación (`/campanas-publico`) que muestra campañas. Nota: la lista de campañas públicas y la de `CampaignManagement.tsx` son datasets hardcodeados independientes, no están conectados, por lo que no se puede demostrar dinámicamente que una campaña en "Borrador" quede oculta en la vista pública. |
| HU-51 | Cubierta | — | El buscador de `PublicCampaigns.tsx` filtra por título/descripción de campaña. |

## 6. Dashboard y Reportes

| HU | Estado | Si falta algo | Notas |
|---|---|---|---|
| HU-52 | Parcialmente cubierta | En `src/app/components/Dashboard.tsx`, ajustar los indicadores del rol admin a exactamente: participantes activos, cursos en desarrollo, cursos completados, habilidades desarrolladas, inserciones laborales (5 indicadores). | El dashboard admin muestra solo 4 indicadores (Total Participantes, Cursos Activos, Voluntarios, Tasa de Asistencia), que no coinciden con el listado pedido por la HU (faltan cursos completados, habilidades desarrolladas e inserciones laborales; sobran Voluntarios y Tasa de Asistencia). |
| HU-53 | No cubierta | Agregar gráficas (librería de charts) de evolución temporal en `Dashboard.tsx`. | No hay ningún gráfico/chart en el dashboard, solo tarjetas numéricas estáticas. |
| HU-54 | No cubierta | Agregar controles de filtro (fecha, área, módulo) en `Dashboard.tsx`. | No existe ningún filtro en el dashboard. |
| HU-55 | No cubierta | Agregar botón "Exportar PDF" en `Dashboard.tsx`. | No existe ninguna opción de exportación en el dashboard. |

## 7. Observabilidad

| HU | Estado | Si falta algo | Notas |
|---|---|---|---|
| HU-56 | No cubierta | Crear un componente/pantalla de observabilidad (p.ej. `src/app/components/Observability.tsx`) con métricas mock por microservicio, y una ruta protegida solo para admin. | No existe ningún componente ni mención de "observabilidad", métricas, latencia o estado de microservicios en todo el código. Explícitamente mencionado como faltante en la vista de administrador de HU-06. |

## 8. Integraciones Externas

| HU | Estado | Si falta algo | Notas |
|---|---|---|---|
| HU-57 | Cubierta | — | `src/app/components/LandingPage.tsx` tiene una sección "¿Dónde estamos?" con `<iframe>` de Google Maps embed. |
| HU-58 | Parcialmente cubierta | Mencionar explícitamente "Gemini" como proveedor de IA y "OpenAI" como fallback en el texto de `CourseCatalog.tsx`/`ProgressTracking.tsx`. | Las "habilidades sugeridas" solo dicen genéricamente "generadas por IA", sin nombrar Gemini ni mencionar fallback a OpenAI. |
| HU-59 | Parcialmente cubierta | Igual que HU-43: nombrar el servicio (RxResume) y simular un mensaje de error claro si el servicio no está disponible. | `CVGenerator.tsx` genera el CV y tiene un mensaje genérico de "se integrará con un servicio de generación de CVs" vía `alert()`, sin manejo de error de servicio no disponible. |

## 9. Sistema de Notificaciones

| HU | Estado | Si falta algo | Notas |
|---|---|---|---|
| HU-60 | No cubierta | Enrutar `NotificationCenter.tsx` en `App.tsx` y agregar confirmaciones visuales de "email enviado" en los flujos de inscripción/oferta/moderación. | No hay ninguna simulación de envío de email en toda la app (ni siquiera un mensaje "Te enviamos un email de confirmación"). El único componente relacionado (`NotificationCenter.tsx`) es de notificaciones in-app, no de email, y además no está enrutado. |
| HU-61 | No cubierta | No hay evidencia de PWA (no existe `manifest.json` en `public/`, solo `logo.png`); habría que agregar manifest + UI de permiso de notificaciones push. | Sin ningún indicio de configuración PWA ni de push notifications en el proyecto. |
| HU-62 | No cubierta | Crear una pantalla de configuración de notificaciones (existe el tipo `NotificationPreferences` en `src/app/types/index.ts` pero ninguna pantalla lo usa) y enrutarla; conectar el ícono de "Settings" del header en `Dashboard.tsx` (no tiene `onClick`). | El ícono de engranaje (Settings) en el header de `Dashboard.tsx` es puramente decorativo, sin acción ni pantalla asociada. |

---

## Resumen numérico

| Estado | Cantidad |
|---|---|
| Cubierta | 13 |
| Parcialmente cubierta | 28 |
| No cubierta | 22 |
| **Total** | **63** (HU-01 a HU-62, incluyendo HU-02bis) |

Por módulo:

| Módulo | Cubierta | Parcial | No cubierta | Total |
|---|---|---|---|---|
| 1. Autenticación y Gestión de Usuarios | 0 | 10 | 4 | 14 |
| 2. Educación | 3 | 8 | 4 | 15 |
| 3. Promoción Humana | 4 | 2 | 3 | 9 |
| 4. Inserción Laboral | 2 | 3 | 4 | 9 |
| 5. Campañas y Concientización | 3 | 2 | 0 | 5 |
| 6. Dashboard y Reportes | 0 | 1 | 3 | 4 |
| 7. Observabilidad | 0 | 0 | 1 | 1 |
| 8. Integraciones Externas | 1 | 2 | 0 | 3 |
| 9. Sistema de Notificaciones | 0 | 0 | 3 | 3 |
| **Total** | **13** | **28** | **22** | **63** |

> Este resumen y las tablas de este bloque reflejan el diagnóstico **inicial**. El estado real y actualizado del proyecto (49 Cubierta / 10 Parcial / 4 No cubierta) está en la sección "Estado actualizado (post-corrección de gaps)" al comienzo de este documento.
