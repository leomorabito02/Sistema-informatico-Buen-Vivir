### **16. Actualización del 15/11/2026: implementación de los pendientes**

Trabajo hecho sobre `develop`: commit `6f1e6f7` (implementación) más las correcciones que salieron de la verificación de punta a punta del mismo día. Se modificó código de `auth/`, `labor-integration/` y `notifications/` con el OK de la usuaria (15/11/2026): **hay que avisarle a Leonardo** antes de pasarlo a testing.

**Decisiones de la usuaria (15/11/2026).**

- auth/ y labor-integration/: se modifican, dejándolo anotado.
- HU-44, "oferta compatible": al publicarse una oferta se avisa a los participantes que tengan al menos una habilidad (ganada en Educación) con el mismo nombre que un requisito de la oferta, sin distinguir mayúsculas ni acentos.
- HU-01, dominio Gmail: obligatorio **solo para participantes** (registro público y alta desde Usuarios). El resto de los roles puede usar cualquier dominio.
- HU-06, panel RBAC: matriz **de solo lectura**. Los permisos siguen en el código de cada servicio; no se hizo una tabla editable.

**Significado de los estados.** IMPLEMENTADO: corregido y verificado (tests automáticos y, para lo de esta actualización, prueba de punta a punta contra el stack Docker con el código nuevo). IMPLEMENTADO - TESTING PENDIENTE: el código y sus tests unitarios están, pero falta probarlo desplegado (necesita una migración, Pub/Sub, el API Gateway de GCP o un servicio externo). PENDIENTE: no implementado. BLOQUEADO: necesita una decisión, un acceso o una intervención externa.

#### **Estado de cada historia observada**

|**HU**|**Estado**|**Qué se hizo (15/11)**|**Qué falta**|
|---|---|---|---|
|HU-01|IMPLEMENTADO|Gmail obligatorio para participantes en `AuthService.Register`, `UserService.CreateUser`, el formulario de registro y el alta desde Usuarios.|-|
|HU-02|IMPLEMENTADO|Bloqueo por email tras 5 intentos fallidos durante 15 minutos (`LOGIN_MAX_ATTEMPTS`, `LOGIN_LOCK_MINUTES`). Tabla `login_attempts` (guarda el hash del email, no el email). Respuesta 423 con `code: ACCOUNT_LOCKED`; antes del bloqueo avisa cuántos intentos quedan. El login lo muestra y deshabilita el botón. Se desbloquea solo al vencer el plazo; un login correcto reinicia el contador.|Verificado de punta a punta en Docker el 15/11 (ver Testing). Falta solo desplegarlo a testing.|
|HU-03|IMPLEMENTADO|Un token presentado y rechazado (401) en una ruta protegida de auth queda en la auditoría como `token_rechazado`. Los pedidos sin sesión no se registran (para no llenar la tabla).|Verificado de punta a punta en Docker el 15/11 (ver Testing). Falta solo desplegarlo a testing.|
|HU-04|IMPLEMENTADO|Reutilizar un refresh token ya usado revoca todos los refresh tokens del usuario y responde 401.|-|
|HU-06|IMPLEMENTADO|Pantalla "Roles y permisos" (`/dashboard/permisos`, solo admin, solo lectura). Los accesos denegados (403) quedan en la auditoría (`acceso_denegado`). Guards por rol en `usuarios`, `cv`, `ofertas` y `seguimiento`.|Verificado de punta a punta en Docker el 15/11 (ver Testing). Falta solo desplegarlo a testing.|
|HU-07|IMPLEMENTADO|DNI (7 u 8 números), dirección y teléfono obligatorios al crear un participante desde Usuarios y en `POST /participants`; DNI y dirección se guardan en el perfil en la misma transacción que la cuenta. El alta queda en la auditoría.|Verificado de punta a punta en Docker el 15/11 (ver Testing). Falta solo desplegarlo a testing.|
|HU-08|IMPLEMENTADO|Cada edición queda en la auditoría con los **nombres** de los campos modificados (nunca los valores).|Verificado de punta a punta en Docker el 15/11 (ver Testing). Falta solo desplegarlo a testing.|
|HU-09|IMPLEMENTADO|Cada baja queda en la auditoría.|Verificado de punta a punta en Docker el 15/11 (ver Testing). Falta solo desplegarlo a testing.|
|HU-10|IMPLEMENTADO|Tabla `auditoria` en auth (usuario, rol, acción, recurso, id, campos, resultado, IP, fecha), con un trigger de Postgres que impide editarla o borrarla. La escribe el middleware `AuditTrail`. `GET /api/v1/auth/auditoria` (solo admin) y módulo "Usuarios y perfiles (Autenticación)" en la pantalla Auditoría.|Verificado de punta a punta en Docker el 15/11 (ver Testing). Falta solo desplegarlo a testing.|
|HU-11|IMPLEMENTADO|Límite de 5 MB para la foto (rxresume, 413) y para los documentos (auth, 400), y el mismo control en el front. Los cambios quedan en la auditoría.|-|
|HU-12|IMPLEMENTADO|Rechazo de cuentas: `PATCH /api/v1/users/{id}/approval` con `rejected` pide el email `cuenta_rechazada`; Notificaciones lo envía. El panel de Usuarios tiene la opción "Rechazado" y aprueba o rechaza por ese endpoint.|Verificado de punta a punta en Docker el 15/11 (ver Testing). Falta solo desplegarlo a testing.|
|HU-14|IMPLEMENTADO|Antes de borrar un material se consulta `GET /api/materiales/{id}/cursos` y la confirmación nombra los cursos que lo usan.|Verificado de punta a punta en Docker el 15/11 (ver Testing). Falta solo desplegarlo a testing.|
|HU-18, HU-23, HU-26, HU-60|BLOQUEADO|Sin cambios de código: dependen de Pub/Sub, Cloud Scheduler y Resend en testing.|Acceso a GCP (pasos 8 a 10).|
|HU-31|BLOQUEADO|Sin cambios.|Decisión del equipo: mantener nombre, edad y habilidades leídos de Auth y Educación (recomendado, no duplica datos personales) o agregarlos como campos de la ficha.|
|HU-38|IMPLEMENTADO|El rechazo de una empresa manda `cuenta_rechazada` (mismo flujo que HU-12).|Verificado de punta a punta en Docker el 15/11 (ver Testing). Falta solo desplegarlo a testing.|
|HU-39|IMPLEMENTADO|Título, descripción, ubicación, modalidad, tipo de empleo, salario, horario, al menos un requisito y fecha límite son obligatorios al crear y editar (400 con el campo que falta). El formulario los marca y muestra el mensaje del backend.|-|
|HU-41|IMPLEMENTADO|Al borrar una oferta se emite `insercion_laboral.oferta.eliminada`; si la borró la gestión del centro, Notificaciones avisa a la empresa.|Verificado de punta a punta en Docker el 15/11 (ver Testing). Falta solo desplegarlo a testing.|
|HU-42|IMPLEMENTADO|Auditoría de ofertas (crear, editar, aprobar, rechazar, borrar) en Firestore (`auditoria_insercion_laboral`), `GET /api/v1/labor-integration/auditoria` (solo admin) y módulo "Inserción laboral" en la pantalla Auditoría.|Verificado de punta a punta en Docker el 15/11 (ver Testing). Falta solo desplegarlo a testing.|
|HU-44|IMPLEMENTADO - TESTING PENDIENTE|Al aprobarse una oferta (o al crearla publicada la gestión), labor pide a Educación las habilidades otorgadas (`GET /api/habilidades/otorgadas`, reenviando el token del admin), cruza nombres con los requisitos y emite `insercion_laboral.oferta.compatible` con los ids de los participantes; Notificaciones avisa a cada uno. Si Educación no responde, la aprobación sigue igual (queda en el log).|`EDUCATION_API_URL` en testing y filtros de Pub/Sub (pasos 5 y 8).|
|HU-50|IMPLEMENTADO|`GET /api/v1/campaigns/publicas` acepta `?limit=&offset=` y devuelve `{items, total}`; sin `limit` responde igual que antes. La página pública trae de a 12 con "Ver más campañas".|-|
|HU-56|IMPLEMENTADO - TESTING PENDIENTE|`infra/monitoring/setup.sh` y `dashboard.json` crean en Cloud Monitoring el panel privado por servicio (pedidos, 5xx, latencia p95, instancias, CPU, memoria), uptime checks y tres alertas por email.|Correrlo en Cloud Shell (paso 11). No se pudo ejecutar: no hay gcloud local.|
|HU-61|BLOQUEADO|Sin cambios.|Credenciales de FCM y `FIREBASE_PROJECT_ID` en testing.|

#### **Estado de los problemas de las pruebas manuales y de los pendientes**

|**Observación**|**Estado**|**Detalle**|
|---|---|---|
|Problema 1 (cursos sin materiales)|IMPLEMENTADO|Corregido el 11/11; sin cambios.|
|Problema 2 (rate limit por IP)|BLOQUEADO|Elegir el salto confiable de `X-Forwarded-For` sin ver cómo lo arma el API Gateway de GCP puede bloquear a usuarios legítimos o dejar el límite esquivable. Paso 12.|
|Problema 3 (notificaciones de PH)|BLOQUEADO|Falta la prueba en el navegador; no se puede hacer contra el stack local con escritura (comparte bases con testing). Paso 13.|
|Problema 4 (permisos de labor)|IMPLEMENTADO|Corregido el 11/11 y verificado otra vez en Docker; falta desplegarlo a testing (paso 7).|
|Problema 5 (eliminar campañas)|IMPLEMENTADO|Sin cambios.|
|Guards de frontend faltantes|IMPLEMENTADO|Ver HU-06.|
|Panel RBAC|IMPLEMENTADO|Solo lectura, por decisión de la usuaria.|
|Validación de dominio Gmail|IMPLEMENTADO|Ver HU-01.|
|Aviso de 429 en el front|IMPLEMENTADO|Ya existía un diálogo global; se agregó el mensaje en las pantallas de login y recuperación, que no pasan por él.|
|E2E por rol|BLOQUEADO|Hace falta una base de datos de pruebas aislada: hoy el stack local y testing comparten Neon y Firestore, y los E2E escriben datos.|
|Recomendación: librería común de autorización|PENDIENTE|Mejora de arquitectura, no es requisito de las HU.|
|Recomendación: reemplazar reportes con valores fijos|PENDIENTE|Pedido del 21/10 de no tocarlos.|
|Recomendación: subir cobertura de rxresume, labor y notifications|PENDIENTE|Se sumaron tests en labor y notifications, pero no se midió la cobertura.|

#### **Observaciones nuevas encontradas durante el trabajo**

1. **Mensajes de error del login perdidos (corregido).** Auth responde `{"error": "..."}` y el front leía solo `message`, así que el usuario veía "Ocurrió un error inesperado" en vez de, por ejemplo, "la contraseña ingresada es incorrecta". `AuthService.handleError` ahora lee los dos y conserva el status.
2. **La aprobación desde Usuarios no mandaba el email (corregido).** El panel aprobaba con `PUT /api/v1/{tabla}/{id}` y el email `cuenta_aprobada` solo salía desde `POST /users/{id}/approve`, que nadie llamaba. Ahora aprobar o rechazar pasa por `PATCH /users/{id}/approval`.
3. **Rutas faltantes en el API Gateway de GCP (corregido en `openapi.yaml`).** No estaban `/users/{id}/approval` ni las rutas nuevas (`/auth/auditoria`, `/labor-integration/auditoria`, `/materiales/{id}/cursos`). Se agregaron; se aplican al desplegar.
4. **Foto de perfil hasta 10 MB (corregido).** La HU-11 pide 5 MB.
5. **Ofertas viejas incompletas.** Con los campos obligatorios de HU-39, una oferta creada antes sin salario, horario o requisitos no se puede editar ni renovar hasta completarlos: el formulario dice qué campo falta.
6. **El bloqueo de HU-02 se puede usar para molestar.** Quien conozca el email de otra persona puede bloquearle la cuenta 15 minutos con 5 intentos. Es propio de lo que pide la HU; se mitiga con el plazo corto y la recuperación de contraseña.
7. **Auditoría fuera de la transacción.** En auth y labor el registro se escribe al terminar el pedido, antes de responder (Cloud Run corta la CPU al responder). Si la base falla justo ahí, el cambio queda sin registro (sí queda en el log técnico). Hacerlo en la misma transacción obligaría a reescribir todos los servicios de auth.
8. **Detección de refresh reutilizado limitada.** La limpieza `DeleteExpired` borra los tokens revocados, así que un refresh reutilizado después de esa limpieza se rechaza como "no encontrado" sin cerrar las demás sesiones.
9. **Errores de validación del alta respondían 500 (corregido en la verificación).** Crear un participante sin Gmail o sin DNI desde Usuarios devolvía 500 porque el controlador trataba cualquier error del servicio como error interno. Ahora responde 400 con el motivo (`ErrValidacion`), con su test.
10. **El email de rechazo casi nunca iba a salir (corregido en la verificación).** Notificaciones no manda emails "no críticos" a direcciones sin verificar, y quien es rechazado casi nunca verificó todavía su email. El aviso de rechazo ahora es crítico, igual que el de verificación, que va a la misma dirección. Verificado: el email se envió (estado `sent`).

#### **Testing ejecutado el 15/11/2026**

|**Qué**|**Resultado**|
|---|---|
|`go vet ./...` y `go test ./...` en pkg, auth, labor-integration, notifications, education, human-promotion, campaigns, dashboard y rxresume|Todo pasa. Antes de los cambios también pasaban (auth, labor y notifications se corrieron al empezar).|
|Tests nuevos de backend|auth: bloqueo tras 5 intentos y vencimiento, refresh reutilizado, Gmail, DNI/dirección obligatorios, rechazo de cuentas, 423, auditoría (alta, edición sin valores, aprobar/rechazar, 403, 401 con y sin token, lecturas ignoradas), endpoint de auditoría, límite de 5 MB. labor: campos obligatorios, cruce de compatibilidad, evento al aprobar y al borrar, auditoría solo admin. notifications: reglas de `cuenta_rechazada`, `oferta.eliminada` y `oferta.compatible`. education: cursos que usan un material. campaigns: paginación.|
|`scripts/sync-eventlib.sh --check`|Las copias de la librería de eventos coinciden con `pkg/`.|
|Frontend: `ng test` (Karma, Chrome headless)|725 de 725 specs pasan (incluye los nuevos de login 423, registro Gmail, Usuarios, ofertas, auditoría, guards, material y campañas).|
|Frontend: `ng build`|Compila sin errores (queda una advertencia previa de una clave duplicada `axisIndex`).|
|Validación de `openapi.yaml` y `pipeline.yml`|YAML válido; sin `operationId` repetidos.|

|Tests Go con emuladores (Firestore en un contenedor aparte y el Pub/Sub del stack)|pkg, education, human-promotion, dashboard, campaigns, labor-integration y notifications: todo pasa, sin tests salteados.|
|Prueba de punta a punta contra el stack Docker reconstruido (`tests/e2e-hu`, resultado en `docs/evidencias/e2e-hu-resultado.json`)|**40 de 40 OK.** Cubre HU-01, 02, 03, 06, 07, 08, 09, 10, 11, 12, 14, 39, 41, 42, 44 (integración con Educación) y 50, más el trigger de inmutabilidad en Neon y la limpieza de los datos [PRUEBA].|
|Envío real del email de rechazo (HU-12)|Cuenta [PRUEBA] registrada y rechazada: Notificaciones guardó "Tu solicitud de cuenta no fue aprobada" con `email=sent`. La cuenta se dio de baja después.|

**Cómo se hizo la prueba de punta a punta.** Con tu autorización se reconstruyeron auth, labor, notifications, education, campaigns, rxresume, frontend y gateway, y se reinició el emulador de Pub/Sub, que recrea las suscripciones con los filtros nuevos. La prueba firma sus propios tokens con `JWT_SECRET`, así que no usa contraseñas. Escribe solo datos marcados [PRUEBA] en la Neon y el Firestore compartidos con testing y los da de baja al final; la auditoría queda, porque es inmutable. Como parte de HU-02, deja bloqueada 15 minutos la cuenta empresa de prueba. Las tablas `login_attempts` y `auditoria`, con su trigger, quedaron creadas en Neon.

**No ejecutado:** k6 y la sonda de seguridad no se volvieron a correr. Tampoco hay E2E de navegador (necesitan una base aislada). Falta ver llegar a un participante el aviso de oferta compatible de HU-44: las únicas cuentas con habilidades son de otras personas y no se cruzaron ofertas con ellas para no mandarles emails reales. Nada se probó todavía en el ambiente de testing de GCP.

**Archivos principales tocados:** `auth/internal/{middlewares/audit_middleware.go, repositories/security_repository.go, services/login_lock.go, services/participant_rules.go, controllers/audit_controller.go}`, `labor-integration/internal/{services/job_offer_hu.go, repositories/audit_repository.go, repositories/education_client.go}`, `notifications/internal/rules/{account.go, labor.go}`, `pkg/events/catalog.go` (+ copias), `infra/pubsub/{setup.sh, emulator.sh}`, `infra/monitoring/`, `api-gateway/{openapi.yaml, nginx.conf}`, `.github/workflows/pipeline.yml` (variable `EDUCATION_API_URL` de labor), y en el front `auditoria.service.ts`, `roles.guard.ts`, `pages/dashboard/permisos/` más los componentes de login, registro, usuarios, ofertas, auditoría, admin-cursos y campañas públicas. `docs/HISTORIAS_DE_USUARIO.md` y `docs/PENDIENTES.md` se regeneraron desde `datos.py` (el PDF no se regeneró porque lo borraste).

# PASO A PASO PARA COMPLETAR LA IMPLEMENTACIÓN EN TESTING

Los pasos están ordenados por dependencias: cada uno supone que se hicieron los anteriores. Cada paso dice quién lo hace:

- **[Código]**: ya está hecho en el repositorio, o lo puedo hacer yo con tu OK.
- **[Manual]**: lo hacés vos (navegador, Cloud Shell, consola).
- **[Credenciales]**: necesita tus accesos o una decisión tuya (GCP, Neon, Firebase, GitHub).

Ningún paso necesita secretos nuevos salvo el 10 (Firebase). Nunca pegues valores de secretos en el repositorio ni en este informe.

#### **Paso 1. Revisar los cambios y avisar a Leonardo** [Manual]

- **Qué:** revisar el diff y avisarle a Leonardo que se tocaron `auth/`, `labor-integration/` y `notifications/`.
- **Dónde:** en el repo, rama `develop`.
- **Cómo:** `git status` y `git diff --stat`. Muchos archivos de `*/internal/events` y `*/internal/outbox` aparecen modificados solo por fin de línea (los reescribió `scripts/sync-eventlib.sh`); git los normaliza al commitear.
- **Resultado esperado:** visto bueno tuyo y aviso hecho.
- **Verificación:** `git diff --stat -- campaigns` muestra solo `campaign_controller.go` y `events/catalog.go` con cambios reales.

#### **Paso 2. Commit en develop** [Hecho: `6f1e6f7`; las correcciones de la verificación van en un commit aparte]

- **Qué:** commitear en `develop` (sin ramas feature) las correcciones que salieron de la verificación del 15/11 (400 en el alta, email de rechazo crítico, prueba `tests/e2e-hu`, informe y PDF).
- **Cómo:** pedímelo y lo hago.
- **Verificación:** `git log -2 --oneline`.

#### **Paso 3. Tests con emuladores** [Hecho el 15/11 en local; el CI los repite]

- **Qué:** correr los tests que hoy se saltean sin emulador.
- **Dónde:** tu máquina (con el emulador de Firestore y el de Pub/Sub levantados) o el CI.
- **Cómo:** `FIRESTORE_EMULATOR_HOST=127.0.0.1:8390 PUBSUB_EMULATOR_HOST=localhost:8086 go test ./...` en `pkg/`, `education/`, `human-promotion/` y `dashboard/`.
- **Resultado esperado:** todo pasa.
- **Verificación:** ningún `FAIL`; en el CI, los jobs `ci-*` en verde.

#### **Paso 4. Migración de la base de auth en Neon** [Hecho el 15/11]

- **Estado:** **ya está hecho.** El auth del stack local las creó al arrancar, y como la Neon es la misma que usa testing, testing ya las tiene. La prueba de punta a punta verificó las dos tablas y el trigger. Solo hay que repetir este paso si testing usara otra base.
- **Qué:** crear las tablas `login_attempts` y `auditoria` y el trigger que impide editar la auditoría. Son cambios **solo de agregado** (no tocan datos existentes). Ojo: el stack local usa la misma base Neon que testing, así que esto vale para los dos.
- **Dónde:** carpeta `auth/`, con la `DATABASE_URL` de testing en `auth/.env`.
- **Cómo:** `cd auth && go run ./cmd/migrate`. (Si no lo corrés, el servicio crea las tablas y el trigger solo al arrancar en Cloud Run; la migración explícita es más segura porque muestra el error si falla.)
- **Resultado esperado:** log "¡Migración completada con éxito!".
- **Verificación:** en la consola SQL de Neon: `\d login_attempts` y `\d auditoria` existen; `UPDATE auditoria SET accion='x';` responde "el registro de auditoria no se puede modificar ni borrar".

#### **Paso 5. Variables de entorno** [Credenciales]

- **Qué y dónde:**
  - `EDUCATION_API_URL` para labor (HU-44): en GitHub → Settings → Environments → `testing` → Variables, con la URL del API Gateway de testing (`https://buenvivir-gateway-testing-5hb9a3vi.uc.gateway.dev`). El CD ya la pasa (`pipeline.yml`, job `cd-labor-integration`). Si queda vacía, el cruce de ofertas compatibles queda apagado y el resto funciona.
  - `LOGIN_MAX_ATTEMPTS` y `LOGIN_LOCK_MINUTES` en auth: opcionales (5 y 15 por defecto). Solo si querés otros valores, agregalos a `env_vars` del job `cd-auth`.
  - Local: `docker-compose.yml` ya tiene `EDUCATION_API_URL=http://education-service:8210` para labor.
- **Resultado esperado:** la variable figura en el environment de GitHub.
- **Verificación:** después del paso 7, `gcloud run services describe labor-integration-testing --region=us-central1 --format='value(spec.template.spec.containers[0].env)'` muestra `EDUCATION_API_URL`.

#### **Paso 6. Secretos y credenciales** [Credenciales]

- **Qué:** no hay secretos nuevos para lo implementado. Revisar que sigan configurados `JWT_SECRET` (auth, labor y el resto con el mismo valor), `GOOGLE_APPLICATION_CREDENTIALS`, el token CMA de Contentful del environment `testing` y la API key de Resend.
- **Dónde:** GitHub → Environments → `testing` → Secrets; Secret Manager en GCP.
- **Verificación:** el job `ci-secrets` del pipeline pasa.

#### **Paso 7. Despliegue de backend, API Gateway y frontend** [Credenciales]

- **Qué:** llevar `develop` a `testing`. El CD despliega los servicios a Cloud Run, la nueva configuración del API Gateway (desde `api-gateway/openapi.yaml`, con las rutas nuevas) y el front a Cloudflare Pages.
- **Cómo:** `git push origin develop`, luego PR `develop → testing` y merge (o pedime el push con tu OK).
- **Resultado esperado:** pipeline en verde (`ci-gate`, `cd-*`, `cd-api-gateway`, `cd-smoke`).
- **Verificación:** `curl -i https://buenvivir-gateway-testing-5hb9a3vi.uc.gateway.dev/api/v1/auth/auditoria` responde 401 (existe y pide sesión), no 404. Con sesión de admin, 200.

#### **Paso 8. Pub/Sub en GCP: filtros y suscripciones** [Credenciales]

- **Qué:** actualizar el filtro de la suscripción `notificaciones-insercion-laboral` para que reciba `insercion_laboral.oferta.eliminada` y `insercion_laboral.oferta.compatible`, y verificar que existan todas las suscripciones push.
- **Dónde:** Cloud Shell, proyecto `ceibuenvivir`.
- **Cómo:** con las URL de los servicios de testing exportadas (`NOTIFICATIONS_URL`, `EDUCATION_URL`, `HUMAN_PROMOTION_URL`, `DASHBOARD_URL`, etc., ver el encabezado de `infra/pubsub/setup.sh`): `bash infra/pubsub/setup.sh`. Es idempotente.
- **Resultado esperado:** el script termina sin errores.
- **Verificación:** `gcloud pubsub subscriptions describe notificaciones-insercion-laboral --format='value(filter)'` incluye `oferta.compatible` y `oferta.eliminada`.

#### **Paso 9. Autenticación de los push y Cloud Scheduler** [Credenciales]

- **Qué:** (a) agregar `PUBSUB_PUSH_AUDIENCE` y `PUBSUB_PUSH_SERVICE_ACCOUNTS` a los `env_vars` de los consumidores en `pipeline.yml` (notifications, education, human-promotion, dashboard), con los valores que usa `setup.sh`; (b) crear los jobs de Cloud Scheduler para `/internal/outbox/flush` de cada productor (cada minuto) y los recordatorios de sesión de HU-26.
- **Dónde:** `pipeline.yml` (lo puedo hacer yo si me pasás las service accounts) y Cloud Scheduler en GCP.
- **Resultado esperado:** los consumidores aceptan solo pushes firmados por su service account.
- **Verificación:** en los logs de notifications aparecen eventos con `resultado=procesado`; un POST sin token a `/eventos/insercion-laboral` responde 401.

#### **Paso 10. Push (FCM) para HU-61** [Credenciales]

- **Qué:** configurar `FIREBASE_PROJECT_ID` y las credenciales de FCM en el environment de testing para notifications.
- **Dónde:** GitHub → Environments → `testing` y Firebase Console → Cloud Messaging.
- **Verificación:** con la PWA instalada en un celular, registrar un hito visible y recibir el push.

#### **Paso 11. Observabilidad (HU-56)** [Credenciales]

- **Qué:** crear el panel y las alertas.
- **Dónde:** Cloud Shell, con permiso de Monitoring Admin.
- **Cómo:** `PROJECT_ID=ceibuenvivir ENV_SUFFIX=testing ALERT_EMAIL=<email del equipo> bash infra/monitoring/setup.sh`
- **Resultado esperado:** imprime la URL privada del panel "Buen Vivir - testing".
- **Verificación:** el panel muestra pedidos y latencia de los 8 servicios; para probar una alerta, bajar un servicio unos minutos (por ejemplo `--max-instances=0` en uno que no se use) y ver que llegue el email de "servicio caído"; después volverlo a su valor.

#### **Paso 12. Destrabar el rate limit (Problema 2)** [Manual, después yo con el código]

- **Qué:** ver cómo llega `X-Forwarded-For` a Cloud Run detrás del API Gateway.
- **Dónde:** Logs Explorer, servicio `education-testing`.
- **Cómo:** hacer un pedido desde tu celular con datos móviles y otro desde la red del centro, y pasarme el campo `httpRequest.remoteIp` y el header `X-Forwarded-For` de esos pedidos.
- **Resultado esperado:** con eso implemento el límite por el último salto confiable (o por usuario) en education, human-promotion y campaigns.
- **Verificación:** rotar `X-Forwarded-For` no esquiva el límite y 15 usuarios desde la misma IP no reciben 429.

#### **Paso 13. Pruebas manuales en testing** [Manual]

Con los usuarios de prueba (`docs/tesis`, datos `[PRUEBA]`) en `https://testing.ceibuenvivir.pages.dev`:

|**#**|**Prueba**|**Resultado esperado**|
|---|---|---|
|1|Login con contraseña mal 5 veces (HU-02)|Del 1.º al 4.º avisa cuántos intentos quedan; el 5.º bloquea y muestra los minutos; con la contraseña correcta sigue bloqueada hasta que vence.|
|2|Registrarse como participante con un email que no es Gmail (HU-01)|El formulario no deja; como empresa con cualquier dominio, sí.|
|3|Admin → Usuarios → nuevo participante sin DNI o con DNI con puntos (HU-07)|Aviso "El DNI es obligatorio…"; completo, se crea y el perfil muestra DNI y dirección.|
|4|Admin → Usuarios → cuenta pendiente → Rechazado (HU-12/38)|El usuario recibe el email de rechazo; con Aprobado recibe el de aprobación.|
|5|Admin → Auditoría → "Usuarios y perfiles" (HU-10)|Aparecen el alta, la edición (con nombres de campos), el rechazo y los accesos denegados de las pruebas anteriores.|
|6|Entrar como participante a `/dashboard/usuarios` por URL (HU-06)|Vuelve al inicio; en Auditoría queda el 403 si llegó a pedir datos.|
|7|Admin → Roles y permisos (HU-06)|Se ve la matriz de solo lectura.|
|8|Empresa → publicar oferta sin salario (HU-39)|"Completá: salario"; completa, queda pendiente.|
|9|**(Pendiente también en local.)** Dar a un participante de prueba una habilidad (entregar una actividad que la desarrolle) y aprobar una oferta con ese requisito (HU-44)|El participante recibe la notificación "Hay una oferta laboral para vos".|
|10|Admin borra una oferta de una empresa (HU-41)|La empresa recibe "Oferta laboral eliminada".|
|11|Admin → Auditoría → "Inserción laboral" (HU-42)|Aparecen la aprobación y el borrado.|
|12|Admin → Gestión de cursos → borrar un material usado (HU-14)|La confirmación nombra los cursos que lo usan.|
|13|Campañas públicas con más de 12 publicadas (HU-50)|Se ven 12 y "Ver más campañas" trae las siguientes.|
|14|Cerrar sesión, reutilizar un refresh viejo (HU-04)|401 y las demás sesiones del usuario se cierran.|
|15|Ficha nueva e hito visible en Promoción Humana (Problema 3)|Notificación en la campana del admin y del participante.|
|16|Inscribir a un participante y moderar un contenido docente (HU-18, HU-23, HU-60)|Llegan el email y la notificación in-app.|

#### **Paso 14. Pruebas automatizadas pendientes** [Manual / Código]

- k6 y sonda de seguridad (`tests/carga/`, `tests/seguridad/sonda/`) contra testing o contra el stack local reconstruido (reconstruirlo requiere tu pedido; la sonda es de solo lectura). Agregar a la sonda: `PATCH /users/x/approval` sin token → 401 y `GET /api/v1/labor-integration/auditoria` con rol no admin → 403.
- E2E por rol (Cypress o Playwright): **bloqueado** hasta tener una base de datos de pruebas aislada (decisión tuya: un proyecto o base Neon y Firestore aparte para E2E).

#### **Paso 15. Verificación final**

1. Pipeline de testing en verde.
2. `GET /api/salud` de education y `/health` del resto responden 200 por el gateway.
3. Las 16 pruebas del paso 13 dan el resultado esperado.
4. El panel de observabilidad muestra datos y una alerta de prueba llegó.
5. Actualizar `scripts/informe-auditoria/datos.py`: pasar a "Implementado" las HU verificadas y regenerar `docs/HISTORIAS_DE_USUARIO.md` y `docs/PENDIENTES.md` (puedo hacerlo yo con tus resultados).

**Qué queda bloqueado y qué necesito de vos para destrabarlo:** acceso a Neon (paso 4), a GitHub Environments y a GCP (pasos 5 a 11), credenciales de Firebase (paso 10), los headers de los logs (paso 12), la decisión sobre HU-31 y una base aislada para E2E.
