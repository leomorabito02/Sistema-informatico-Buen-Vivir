# Pendientes

Estado al 15/11/2026. Generado por `scripts/informe-auditoria/generar.py` desde `datos.py`: no editar a mano.

| Prioridad | Pendiente | HU | Módulo | Trabajo necesario | Dependencias | Criterio de aceptación |
|---|---|---|---|---|---|---|
| Alta | Commit y pase a testing de los cambios del 15/11 | Varias | Todos | Commit en develop, PR develop -> testing | OK de la usuaria; avisar a Leonardo (auth, labor, notifications) | El CD de testing termina en verde. |
| Alta | Migrar tablas nuevas de auth en Neon | HU-02, HU-10 | Autenticación | go run ./cmd/migrate (crea login_attempts, auditoria y el trigger) | Acceso a Neon | Las tablas existen y un UPDATE sobre auditoria falla. |
| Alta | Rutas nuevas en el API Gateway de GCP | HU-10, HU-12, HU-14, HU-42 | Gateway | Desplegar la nueva config del gateway | Acceso a GCP (lo hace el CD) | GET /api/v1/auth/auditoria responde 200 a un admin en testing. |
| Alta | Verificar Pub/Sub y emails en testing | HU-18, HU-23, HU-26, HU-41, HU-44, HU-60 | Notificaciones | Correr setup.sh (filtros nuevos), agregar PUBSUB_PUSH_* al CD, Cloud Scheduler para recordatorios | Acceso a GCP | Una inscripción en testing genera el email y la notificación in-app. |
| Alta | EDUCATION_API_URL en labor (testing) | HU-44 | Inserción Laboral | Variable con la URL del gateway de testing | Acceso a GCP | Aprobar una oferta con un requisito igual a una habilidad avisa al participante. |
| Alta | Probar en el navegador el flujo de notificaciones de PH | Problema 3 | PH / Notificaciones | Crear una ficha y un hito visible y ver la notificación en la campana | Testing desplegado | Una ficha nueva genera notificación en la campana del admin y el hito visible le llega al participante. |
| Media | Instalar y correr E2E | Varias | Frontend | Cypress o Playwright, specs por rol | Base de datos de pruebas aislada (hoy local y testing comparten Neon y Firestore) | Los specs E2E pasan contra un entorno con datos propios. |
| Media | Push en testing | HU-61 | Notificaciones | Credenciales de FCM y FIREBASE_PROJECT_ID | Acceso a Firebase | Con la PWA instalada llega un push al registrar un hito visible. |
| Media | Rate limit robusto | Problema 2 | education, PH, campaigns | Ver en los logs de testing cómo llega X-Forwarded-For y limitar por el último salto confiable o por usuario | Logs de testing | Rotar X-Forwarded-For no esquiva el límite y 15 usuarios de una misma IP usan la app sin 429. |
| Media | Observabilidad en GCP | HU-56 | Infra | Correr el script | Acceso a GCP | URL privada con el panel y una alerta que se dispara ante caída. |
| Baja | Campos estructurados de la ficha | HU-31 | Promoción Humana | Decidir si se mantiene leer nombre, edad y habilidades desde Auth/Educación | Decisión del equipo | La HU queda actualizada o se agregan los campos. |

## Historias no completas

| HU | Estado | Qué falta |
|---|---|---|
| HU-02 | Pendiente de verificación | Migrar la tabla login_attempts en Neon (cmd/migrate o arranque) y probar el bloqueo en testing. |
| HU-03 | Pendiente de verificación | Migrar la tabla auditoria. Los rechazos de token en los otros servicios siguen solo en el log técnico. |
| HU-06 | Pendiente de verificación | Migrar la tabla auditoria. Docente creando cursos en borrador: se mantiene (es la moderación de HU-23). |
| HU-07 | Pendiente de verificación | Probar en testing que el DNI y la dirección queden guardados en el perfil. |
| HU-08 | Pendiente de verificación | Migrar la tabla auditoria. |
| HU-09 | Pendiente de verificación | Migrar la tabla auditoria. |
| HU-10 | Pendiente de verificación | Migrar la tabla y el trigger en Neon; agregar la ruta al API Gateway. Se escribe al terminar el pedido, no en la misma transacción del cambio. |
| HU-11 | Pendiente de verificación | Migrar la tabla auditoria. |
| HU-12 | Pendiente de verificación | Agregar la ruta al API Gateway y verificar el email en testing. |
| HU-18 | Pendiente de verificación | Verificar en testing que llegue el email (eventos + suscripción push de notifications). |
| HU-23 | Pendiente de verificación | Verificar la notificación en testing. |
| HU-26 | Pendiente de verificación | Configurar el job de Scheduler y verificar envíos. |
| HU-31 | Parcialmente implementado | Confirmar con el equipo y, si corresponde, actualizar la HU. |
| HU-38 | Pendiente de verificación | Verificar el email en testing. |
| HU-41 | Pendiente de verificación | Actualizar las suscripciones de Pub/Sub (setup.sh) y verificar en testing. |
| HU-42 | Pendiente de verificación | Agregar la ruta al API Gateway y verificar en testing. |
| HU-44 | Pendiente de verificación | Configurar EDUCATION_API_URL en labor (testing), actualizar Pub/Sub y verificar con un participante con habilidades. |
| HU-56 | Pendiente de verificación | Correr el script en Cloud Shell (requiere acceso a GCP) y probar una alerta. |
| HU-60 | Pendiente de verificación | Verificar suscripciones push en GCP. |
| HU-61 | Parcialmente implementado | Configurar credenciales de FCM en testing y probar en un dispositivo. |

## Recomendaciones (no son requisitos de las HU)

- **Seguridad:** No confiar nunca en headers de identidad: validar el JWT en cada servicio (labor ya lo hace con JWT_SECRET). Agregar JWT_SECRET al .env local de labor.
- **Seguridad:** Agregar PUBSUB_PUSH_AUDIENCE y PUBSUB_PUSH_SERVICE_ACCOUNTS al CD de cada consumidor y hacer que fallen cerrado también en testing.
- **Seguridad:** Los .env guardan JWT_SECRET entre comillas simples (Docker Compose las quita): al copiar la clave a otro servicio o a GCP, copiar la línea completa con el mismo formato.
- **Rendimiento:** Paginación y límites en listados (campañas, biblioteca, ofertas); hoy se lee toda la colección en cada pedido.
- **Rendimiento:** Revisar el límite de 10/s (nginx) y 100/min (servicios) por IP: un centro barrial con una sola IP pública puede quedarse corto en uso normal.
- **Arquitectura:** Unificar el modelo de autorización (RequireRole + dueño) en una librería compartida como se hizo con pkg/events, para que labor y auth no diverjan.
- **Arquitectura:** Auditoría común (mismo modelo que education y PH) para auth y labor.
- **Mantenibilidad:** Reemplazar los reportes con valores fijos de tests/{security,performance,e2e_a11y} por los resultados reales (k6, sonda, ZAP).
- **Testing:** Instalar Cypress o Playwright y correr E2E por rol en el CI con el stack Docker.
- **Testing:** Subir cobertura de rxresume, labor y notifications; specs de la ficha de PH.
- **Organización:** Recrear el stack Docker después de cada cambio de docker-compose.yml o .env (al momento de la auditoría eran del 29/10 y notifications no tenía AUTH_GRPC_URL).
- **Documentación:** Mantener docs/HISTORIAS_DE_USUARIO.md y este informe desde scripts/informe-auditoria/datos.py.
- **Auditoría:** La auditoría de auth y labor se escribe al terminar el pedido (no en la misma transacción): si la base falla justo en ese momento puede quedar un cambio sin registro (queda en el log técnico).
