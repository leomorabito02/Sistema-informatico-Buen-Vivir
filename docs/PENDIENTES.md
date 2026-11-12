# Pendientes

Estado al 11/11/2026. Generado por `scripts/informe-auditoria/generar.py` desde `datos.py`: no editar a mano.

| Prioridad | Pendiente | HU | Módulo | Trabajo necesario | Dependencias | Criterio de aceptación |
|---|---|---|---|---|---|---|
| Alta | Reconstruir el stack Docker local | Problemas 3 y 5 | Infra local | docker compose up -d --build (recrea notifications con AUTH_GRPC_URL, campaigns con DELETE, labor con el arreglo, emulador autoinicializado, gateway con 429) | La usuaria | La sonda de seguridad da 39/39 y una ficha nueva genera notificación en la campana del admin. |
| Alta | Desplegar a testing los arreglos de seguridad de labor | HU-06, Problema 4 | Inserción Laboral | Commit + pase a testing | OK del equipo (código de otro integrante) | PUT /offers/x/approve sin token con X-User-Role: role_admin responde 401/403 en testing. |
| Alta | Bloqueo de cuenta tras 5 intentos fallidos | HU-02 | Autenticación | Contador por cuenta + bloqueo temporal + mensaje | - | Al 6.º intento la cuenta queda bloqueada N minutos y se desbloquea sola. |
| Alta | Log de auditoría en auth | HU-03, HU-06, HU-07 a HU-11 | Autenticación | Registrar en la misma transacción usuario, acción, recurso y fecha; endpoint de lectura para admin; sumar a /dashboard/auditoria | - | Cada alta/baja/edición de perfil aparece en la pantalla de auditoría y no se puede editar. |
| Alta | Refresh reutilizado invalida la sesión | HU-04 | Autenticación | Si el token está revocado, RevokeAllForUser | - | Reusar un refresh devuelve 401 y el siguiente refresh legítimo también falla. |
| Alta | Verificar Pub/Sub y emails en testing | HU-18, HU-23, HU-26, HU-60 | Notificaciones | Correr setup.sh con NOTIFICATIONS_URL, agregar PUBSUB_PUSH_* al CD, Cloud Scheduler para recordatorios | Acceso a GCP | Una inscripción en testing genera el email y la notificación in-app. |
| Alta | Instalar y correr E2E | Varias | Frontend | Agregar Cypress al package.json, support file y specs por rol | - | npx cypress run pasa en local con el stack Docker. |
| Media | Notificación de oferta compatible | HU-44 | Inserción Laboral / Notificaciones | Definir compatibilidad (habilidades) y regla para participantes | Definición funcional | Al aprobarse una oferta, los participantes con habilidades coincidentes reciben aviso. |
| Media | Panel de configuración RBAC | HU-06 | Autenticación / Frontend | Pantalla de solo admin para ver y editar permisos por rol | - | El admin ve la matriz de permisos y un cambio se refleja en el backend. |
| Media | Campos obligatorios de ofertas y participantes | HU-07, HU-39 | Auth, Labor | validate:required en los campos de la HU | - | Crear sin uno de esos campos responde 400 con el campo faltante. |
| Media | Notificar rechazos de cuentas | HU-12, HU-38 | Auth / Notificaciones | Email 'cuenta_rechazada' | - | Al rechazar, el usuario recibe el email. |
| Media | Push en testing | HU-61 | Notificaciones | Credenciales de FCM y FIREBASE_PROJECT_ID | Acceso a Firebase | Con la PWA instalada llega un push al registrar un hito visible. |
| Media | Paginación de listados públicos | HU-50 | Campañas, Educación | Limit + cursor; búsqueda con índice | - | Con 2000 campañas el p95 del listado público queda por debajo de 500 ms. |
| Media | Rate limit robusto | Problema 2 | education, PH, campaigns | Tomar la IP del último salto confiable o limitar por usuario; revisar el cupo de 100/min para una IP compartida | Conocer cómo arma X-Forwarded-For el API Gateway | Rotar X-Forwarded-For no esquiva el límite y 15 usuarios de una misma IP usan la app sin 429. |
| Media | Observabilidad | HU-56 | Infra | Dashboard por servicio (errores, latencia, recursos) y alertas | Acceso a GCP | URL privada con el panel y una alerta que se dispara ante caída. |
| Media | Aviso al borrar material adjuntado | HU-14 | Educación | Contar cursos que lo usan y avisar o impedir | - | Borrar un material usado muestra en qué cursos está. |
| Baja | Guards de frontend faltantes | HU-06 | Frontend | Guards en usuarios, cv, ofertas y seguimiento | - | Entrar por URL con un rol no permitido redirige al inicio. |
| Baja | Validación de dominio gmail | HU-01 | Autenticación | Confirmar criterio y validar | Definición funcional | Un email que no es gmail es rechazado con mensaje claro. |

## Historias no completas

| HU | Estado | Qué falta |
|---|---|---|
| HU-01 | Parcialmente implementado | Agregar la validación de dominio gmail en el DTO de registro y en el formulario (o confirmar con el equipo si el criterio sigue vigente). |
| HU-02 | Parcialmente implementado | Contador de intentos fallidos por cuenta con bloqueo temporal (por ejemplo 15 minutos) y su test. |
| HU-03 | Parcialmente implementado | Registrar los rechazos de token en una auditoría central (ver HU-10). |
| HU-04 | Parcialmente implementado | Ante reutilización, revocar todos los refresh tokens del usuario (RevokeAllForUser ya existe). |
| HU-06 | Parcialmente implementado | Panel RBAC en el front; auditoría de accesos denegados; decidir si el docente puede proponer cursos. |
| HU-07 | Parcialmente implementado | Hacer obligatorios los campos de la HU en el DTO; auditoría de altas. |
| HU-08 | Parcialmente implementado | Auditoría de modificaciones en auth. |
| HU-09 | Parcialmente implementado | Auditoría de bajas. |
| HU-10 | No implementado | Tabla de auditoría en auth (usuario, acción, recurso, fecha) en la misma transacción, y sumarla a la pantalla de auditoría. |
| HU-11 | Parcialmente implementado | Auditoría; verificar límite de 5 MB para la foto de perfil. |
| HU-12 | Parcialmente implementado | Email y aviso de rechazo. |
| HU-18 | Pendiente de verificación | Verificar en testing que llegue el email (eventos + suscripción push de notifications). |
| HU-23 | Pendiente de verificación | Verificar la notificación en testing. |
| HU-26 | Pendiente de verificación | Configurar el job de Scheduler y verificar envíos. |
| HU-31 | Parcialmente implementado | Confirmar con el equipo y, si corresponde, actualizar la HU. |
| HU-38 | Parcialmente implementado | Notificar el rechazo. |
| HU-39 | Parcialmente implementado | Validar todos los campos en el DTO y en el formulario. |
| HU-41 | Parcialmente implementado | Notificar a la empresa al eliminar. |
| HU-42 | Parcialmente implementado | Auditoría de eliminaciones. |
| HU-44 | No implementado | Definir 'compatible' (habilidades / CV) y la regla de notificación. |
| HU-56 | Parcialmente implementado | Dashboard de Cloud Monitoring (errores, latencia, CPU/memoria por servicio) con alertas, o Grafana. |
| HU-60 | Pendiente de verificación | Recrear el stack local; verificar suscripciones push en GCP. |
| HU-61 | Parcialmente implementado | Configurar credenciales de FCM en testing y probar en un dispositivo. |

## Recomendaciones (no son requisitos de las HU)

- **Seguridad:** No confiar nunca en headers de identidad: validar el JWT en cada servicio (labor ya lo hace con JWT_SECRET). Agregar JWT_SECRET al .env local de labor.
- **Seguridad:** Agregar PUBSUB_PUSH_AUDIENCE y PUBSUB_PUSH_SERVICE_ACCOUNTS al CD de cada consumidor y hacer que fallen cerrado también en testing.
- **Seguridad:** Rotar el JWT_SECRET local si alguna vez se compartió: los contenedores corren con un secreto distinto al del .env actual (señal de configuración desincronizada).
- **Rendimiento:** Paginación y límites en listados (campañas, biblioteca, ofertas); hoy se lee toda la colección en cada pedido.
- **Rendimiento:** Revisar el límite de 10/s (nginx) y 100/min (servicios) por IP: un centro barrial con una sola IP pública puede quedarse corto en uso normal.
- **Arquitectura:** Unificar el modelo de autorización (RequireRole + dueño) en una librería compartida como se hizo con pkg/events, para que labor y auth no diverjan.
- **Arquitectura:** Auditoría común (mismo modelo que education y PH) para auth y labor.
- **Mantenibilidad:** Reemplazar los reportes con valores fijos de tests/{security,performance,e2e_a11y} por los resultados reales (k6, sonda, ZAP).
- **Testing:** Instalar Cypress o Playwright y correr E2E por rol en el CI con el stack Docker.
- **Testing:** Subir cobertura de rxresume, labor y notifications; specs de la ficha de PH.
- **Organización:** Recrear el stack Docker después de cada cambio de docker-compose.yml o .env (los contenedores actuales son del 29/10).
- **Documentación:** Mantener docs/HISTORIAS_DE_USUARIO.md y este informe desde scripts/informe-auditoria/datos.py.
- **Experiencia de usuario:** Avisar en el front cuando el backend responde 429 (hoy aparece como error genérico).
