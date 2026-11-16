# **Buen Vivir App** 

## Informe de auditoría técnica y funcional 

Centro de Ecología Integral · Centro Barrial “La Otra Orilla”, Alta Gracia 

Fecha de realización: 11/11/2026 

Actualización: 15/11/2026 (implementación de los pendientes; estado de cada observación en la sección 16 y paso a paso para testing al final del documento) 

Versión auditada: rama develop, commit base cc8dbfb + cambios sin commitear de esta auditoría 

Documento técnico. No contiene secretos, tokens ni credenciales. 

##### **Índice** 

|**1. Introducción**|**4**|
|---|---|
|**2. Arquitectura general**|**4**|
|**3. Resumen ejecutivo**|**6**|
|**4. Auditoría de historias de usuario**|**7**|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Tabla general|.  .  .  .  .  .  .  .  .  .  .  . 7|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Autenticación y usuarios|.  .  .  .  .  .  .  .  .  .  .  . 9|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Educación|.  .  .  .  .  .  .  .  .  .  .  . 13|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Promoción Humana|.  .  .  .  .  .  .  .  .  .  .  . 19|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Inserción Laboral|.  .  .  .  .  .  .  .  .  .  .  . 22|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Campañas|.  .  .  .  .  .  .  .  .  .  .  . 25|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Dashboard y reportes|.  .  .  .  .  .  .  .  .  .  .  . 26|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Observabilidad|.  .  .  .  .  .  .  .  .  .  .  . 28|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Integraciones externas|.  .  .  .  .  .  .  .  .  .  .  . 28|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Notificaciones|.  .  .  .  .  .  .  .  .  .  .  . 29|
|**5. Auditoría del backend**|**31**|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Estado de los servicios|.  .  .  .  .  .  .  .  .  .  .  . 31|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Arquitectura y organización|.  .  .  .  .  .  .  .  .  .  .  . 31|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Endpoints|.  .  .  .  .  .  .  .  .  .  .  . 31|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Base de datos|.  .  .  .  .  .  .  .  .  .  .  . 31|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Autenticación, autorización y seguridad|.  .  .  .  .  .  .  .  .  .  .  . 31|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Integración entre servicios|.  .  .  .  .  .  .  .  .  .  .  . 32|
|**6. Auditoría del frontend**|**32**|
|**7. Matriz de roles y permisos**|**32**|
|**8. Problemas detectados durante las pruebas manuales**|**33**|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Problema 1. Educación: creación de cursos y materiales|.  .  .  .  .  .  .  .  .  .  .  . 33|
|.  .  .  .  .  .  .  .  .  .  .<br>Problema 2. Pruebas de carga y seguridad en todos los servicios|.  .  .  .  .  .  .  .  .  .  .  . 33|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Problema 3. Promoción Humana: notificaciones|.  .  .  .  .  .  .  .  .  .  .  . 33|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Problema 4. Permisos y acciones según el rol|.  .  .  .  .  .  .  .  .  .  .  . 34|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Problema 5. Campañas: eliminación|.  .  .  .  .  .  .  .  .  .  .  . 34|
|**9. Inventario de tests existentes**|**36**|
|**10. Tests faltantes**|**37**|
|**11. Resultados de testing y seguridad**|**38**|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Carga con k6 (stack Docker local)|.  .  .  .  .  .  .  .  .  .  .  . 38|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Carga en proceso con volumen de datos (Campañas)|.  .  .  .  .  .  .  .  .  .  .  . 38|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Sonda de seguridad|.  .  .  .  .  .  .  .  .  .  .  . 38|
|**12. Pendientes de implementación**|**40**|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Historias no completas|.  .  .  .  .  .  .  .  .  .  .  . 41|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 2 

|**13. Recomendaciones técnicas**|**41**|
|---|---|
|**14. Conclusión**|**42**|
|**15. Anexo técnico**|**43**|
|**16. Actualización del 15/11/2026**|-|
|**Paso a paso para completar la implementación en testing**|-|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Estructura del repositorio|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  . 43|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Comandos de ejecución|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  . 43|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Endpoints principales por servicio|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  . 43|
|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .<br>Archivos importantes|.  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  . 43|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 3 

### **1. Introducción** 

**Descripción del proyecto.** Buen Vivir App es la aplicación web progresiva (PWA) de gestión interna del Centro de Ecología Integral. Reúne en un solo sistema la gestión de usuarios y roles, la formación (cursos, actividades, asistencia y habilidades), el acompañamiento personal de participantes (Promoción Humana), la inserción laboral (ofertas, postulaciones y CV), las campañas de concientización, los indicadores del centro y las notificaciones. 

**Objetivo de la auditoría.** Verificar contra las historias de usuario que cada funcionalidad esté implementada de punta a punta (frontend, API, lógica, base de datos y respuesta), detectar y corregir errores, revisar y completar los tests, investigar cinco problemas reportados en pruebas manuales, ordenar el repositorio y documentar el estado real y lo pendiente. 

**Alcance.** Los nueve servicios Go (auth, education, human-promotion, campaigns, dashboard, notifications, labor-integration, rxresume y la librería pkg), el frontend Angular, el API Gateway (nginx local y OpenAPI de GCP), la infraestructura de Pub/Sub y el pipeline de CI/CD. Se analizaron las 64 historias de usuario del documento “HU DE BUEN VIVIR APP”. 

##### **Metodología.** 

- Lectura completa de las HU y de la documentación existente (CLAUDE.md, README, docs/, ADR). 

- Revisión del código por HU: componente y guard del front, ruta del gateway, middleware de rol, servicio, repositorio y colección o tabla. 

- Ejecución de todas las suites de tests (Go con emuladores de Firestore y Pub/Sub; Karma con ChromeHeadless). 

- Pruebas contra el stack Docker local: carga con k6 y una sonda de seguridad propia, sin escrituras sobre datos reales. 

- Corrección de los problemas con su test, y nueva ejecución de las suites. 

- Clasificación de cada HU en: Implementado, Parcialmente implementado, No implementado, Implementado con errores o Pendiente de verificación. 

**Fuentes consultadas.** HU DE BUEN VIVIR APP.pdf; DER y explicación del DER; CLAUDE.md raíz y de education; docs/eventos, docs/adr, docs/der_educacion.md, docs/ia-habilidades.md; código fuente; .github/workflows/pipeline.yml; docker-compose.yml; api-gateway/nginx.conf y openapi.yaml; resultados de tests de esta auditoría. 

### **2. Arquitectura general** 

Microservicios en Go 1.26 desplegados en Google Cloud Run (escalan a cero), detrás de Google Cloud API Gateway en la nube y de nginx en local. Comunicación REST hacia afuera, HTTP/gRPC interno para consultas síncronas y Pub/Sub para eventos asíncronos con patrón outbox. Frontend Angular 20 con PrimeNG, PWA, publicado en Cloudflare Pages. 

|**Componente**|**Tecnología**|**Responsabilidad**|**Datos**|
|---|---|---|---|
|frontend|Angular 20, PrimeNG, Firebase<br>(Auth y FCM)|PWA para todos los roles y landing pública|-|
|api-gateway|nginx (local) / Google API<br>Gateway (OpenAPI 2.0)|Ruteo, límite de tasa (local), validación de token<br>para auth/labor/rxresume|-|
|auth|Go, chi, GORM, gRPC|Usuarios, roles, perfiles, login, JWT, 2FA, Google,<br>recuperación|PostgreSQL (Neon)|
|education|Go, chi|Cursos, áreas, materiales, inscripciones,<br>actividades, entregas, asistencia, habilidades IA,<br>comentarios, biblioteca, auditoría|Contentful + Firestore +<br>Cloud Storage|
|human-promotion|Go, chi|Fichas de proceso personal (cifradas), hitos, línea de<br>tiempo, PDF, auditoría|Firestore|
|campaigns|Go, chi|Campañas de concientización|Firestore|
|labor-integration|Go, chi|Ofertas laborales y postulaciones|Firestore|
|rxresume / cv-renderer|Go + Reactive Resume (Node)|Generación de CV|PostgreSQL|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 4 

|**Componente**|**Tecnología**|**Responsabilidad**|**Datos**|
|---|---|---|---|
|dashboard|Go|Indicadores, gráficas, filtros, reportes PDF|Agrega de otros servicios;<br>Firestore + bucket para<br>reportes|
|notifications|Go, GORM, gRPC|Notificaciones in-app, email (Resend) y push (FCM),<br>preferencias|PostgreSQL|
|Pub/Sub|Google Pub/Sub (emulador en<br>local)|Eventos de dominio con outbox, DLQ y OIDC|-|



Cada servicio sigue capas controller → service → repository, con interfaces para testear con dobles. Los eventos viajan con un formato único (envelope con event_id, event_type, version, correlation_id, actor_id y payload sin datos personales). 

Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 5 

### **3. Resumen ejecutivo** 

Se analizaron **64 historias de usuario** (HU-01 a HU-62, más HU-02bis y HU-19bis). Los números surgen del análisis de esta auditoría y reflejan el estado después de las correcciones. 

|**Estado**|**Cantidad**|
|---|---|
|Implementado|44|
|Parcialmente implementado|2|
|No implementado|0|
|Implementado con errores|0|
|Pendiente de verificación|18|
|Total|64|

**Actualización 15/11/2026.** Los números de la tabla son los nuevos (al 11/11 eran 41 implementadas, 17 parciales, 2 no implementadas y 4 pendientes de verificación). "Pendiente de verificación" ahora agrupa las HU con código y tests listos que falta probar desplegadas (IMPLEMENTADO - TESTING PENDIENTE en la sección 16). Las dos parciales son HU-31 y HU-61, bloqueadas por una decisión y por credenciales.



##### **Principales problemas encontrados** 

- Crítico (corregido en código): Inserción Laboral aceptaba la identidad que mandaba el cliente en headers; sin sesión se podían aprobar o borrar ofertas. 

- Alto (corregido): editar y borrar ofertas sin control de rol ni dueño; perfiles creados a nombre de otras cuentas; pendientes visibles a cualquier rol. 

- Problema 1 (corregido): no se podía crear un curso sin materiales (formulario, backend y Contentful). 

- Problema 3 (corregido en código y en el emulador): eventos que no llegaban, enlaces rotos, aviso al creador que nunca salía, duplicados de campañas. 

- Problema 5 (implementado): la eliminación de campañas no existía. 

- Pendientes funcionales: bloqueo por intentos de login, auditoría en auth, notificación de ofertas compatibles, panel RBAC, observabilidad propia. 

- Entorno: el stack Docker local corría contenedores del 29/10 (notifications sin AUTH_GRPC_URL y labor sin JWT_SECRET); se reconstruyó el 12/11. 

**Estado general.** El sistema está funcionalmente avanzado: 41 de 64 HU quedan implementadas y no hay errores abiertos conocidos en las implementadas. Lo que falta se concentra en Autenticación (auditoría, bloqueo, sesiones), en la verificación de punta a punta de notificaciones y emails en testing, y en algunas HU de Inserción Laboral. Todas las suites de tests pasan (9 módulos Go y 694 specs de frontend). 

Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 6 

### **4. Auditoría de historias de usuario** 

#### **Tabla general** 

|**HU**|**Módulo**|**Historia**|**Estado**|**Nota principal**|
|---|---|---|---|---|
|HU-01|Autenticación y<br>usuarios|Registrar usuarios con nombre,<br>email, contraseña, teléfono y rol|IMPLEMENTADO|15/11: Gmail obligatorio solo para participantes (decisión de la usuaria).|
|HU-02|Autenticación y<br>usuarios|Iniciar sesión con email y<br>contraseña|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: bloqueo 15 min tras 5 intentos (423). Falta migrar la tabla y probar en testing.|
|HU-02bis|Autenticación y<br>usuarios|Recuperar contraseña por enlace<br>al email|Implementado|Verificar el envío real del email en testing<br>(Resend).|
|HU-03|Autenticación y<br>usuarios|Validar el JWT en cada solicitud<br>protegida|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: tokens rechazados en la auditoría de auth.|
|HU-04|Autenticación y<br>usuarios|Expiración configurable y refresh<br>tokens de uso único|IMPLEMENTADO|15/11: reutilizar un refresh cierra todas las sesiones.|
|HU-05|Autenticación y<br>usuarios|Autenticación de doble factor (2FA)|Implementado|-|
|HU-06|Autenticación y<br>usuarios|Control de acceso RBAC con<br>vistas por rol|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: matriz de permisos de solo lectura, 403 auditados y guards.|
|HU-07|Autenticación y<br>usuarios|Crear perfiles de participantes|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: DNI, dirección y teléfono obligatorios; alta auditada.|
|HU-08|Autenticación y<br>usuarios|Editar perfiles de participantes|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: ediciones auditadas con nombres de campos.|
|HU-09|Autenticación y<br>usuarios|Eliminar perfiles de participantes<br>(baja lógica)|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: bajas auditadas.|
|HU-10|Autenticación y<br>usuarios|Log de auditoría de acciones sobre<br>perfiles|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: tabla auditoria inmutable, endpoint y pantalla. Falta migrar y desplegar.|
|HU-11|Autenticación y<br>usuarios|Foto y documentación del<br>participante|IMPLEMENTADO|15/11: límite de 5 MB en foto y documentos; cambios auditados.|
|HU-12|Autenticación y<br>usuarios|Aprobar el registro de voluntarios|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: rechazo con email 'cuenta_rechazada'.|
|HU-13|Autenticación y<br>usuarios|Iniciar sesión con Google|Implementado|-|
|HU-14|Educación|Crear cursos (con materiales<br>opcionales)|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: aviso de los cursos que usan el material antes de borrarlo.|
|HU-15|Educación|Editar cursos|Implementado|-|
|HU-16|Educación|Eliminar cursos (baja lógica)|Implementado|-|
|HU-17|Educación|Contenido de cursos gestionado<br>en Contentful|Implementado|-|
|HU-18|Educación|Inscribir participantes a cursos|BLOQUEADO|Depende de Pub/Sub y Resend en testing (acceso a GCP).|
|HU-19|Educación|Asistencia por cumplimiento de<br>actividades|Implementado|-|
|HU-19bis|Educación|Asistencia general diaria del centro|Implementado|-|
|HU-20|Educación|Progreso formativo por actividades|Implementado|La pantalla Seguimiento estaba en el menú de<br>admin y voluntario, pero solo usaba endpoints de<br>docente: 403 al voluntario, lista vacía al admin.|
|HU-21|Educación|Sugerencia de habilidades con IA|Implementado|Diferencia con la HU: se sugiere por actividad y<br>no al crear el curso (decisión del equipo,<br>docs/ia-habilidades.md).|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 7 

|**HU**|**Módulo**|**Historia**|**Estado**|**Nota principal**|
|---|---|---|---|---|
|HU-22|Educación|Docente externo consulta<br>contenidos|Implementado|-|
|HU-23<br>|Educación<br>|Docente externo sube contenido<br>con moderación<br>|BLOQUEADO|Depende de Pub/Sub en testing (acceso a GCP).|
|HU-24|Educación|Comentarios en contenido público|Implementado|-|
|HU-25|Educación|Registro de datos al descargar de<br>la biblioteca|Implementado|-|
|HU-26|Educación|Email al inscribirse y recordatorios<br>de sesión|BLOQUEADO|Depende de Pub/Sub, Cloud Scheduler y Resend (acceso a GCP).|
|HU-27|Educación|Buscar cursos por nombre|Implementado|-|
|HU-28|Educación|Buscar cursos por área formativa|Implementado|-|
|HU-29|Promoción Humana|Crear ficha de proceso personal|Implementado|-|
|HU-30|Promoción Humana|Editar ficha|Implementado|-|
|HU-31|Promoción Humana|Campos estructurados de la ficha|BLOQUEADO|Decisión del equipo (mantener los datos leídos de Auth y Educación).|
|HU-32|Promoción Humana|Registrar hitos de avance|Implementado|-|
|HU-33|Promoción Humana|Línea de tiempo del participante|Implementado|-|
|HU-34|Promoción Humana|Línea de tiempo propia del<br>participante|Implementado|-|
|HU-35|Promoción Humana|Auditoría de modificaciones de<br>fichas|Implementado|La HU menciona 'eliminar' fichas: no existe<br>eliminación de fichas (no hay nada que auditar).|
|HU-36|Promoción Humana|Buscar participantes por nombre|Implementado|-|
|HU-37|Promoción Humana|Exportar ficha en PDF|Implementado|-|
|HU-38|Inserción Laboral|Cuenta de empresa con<br>aprobación|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: el rechazo de empresas manda email.|
|HU-39|Inserción Laboral|Publicar ofertas laborales|IMPLEMENTADO|15/11: todos los campos de la HU obligatorios (backend y formulario).|
|HU-40|Inserción Laboral|Ofertas pendientes hasta<br>aprobación|Implementado|Antes: cualquier rol distinto de participante veía<br>las pendientes, y la empresa podía<br>autoaprobarse editando el estado. Excepción<br>vigente: las ofertas creadas por admin o<br>voluntario nacen 'Activa'.|
|HU-41|Inserción Laboral|Aprobar, rechazar o eliminar<br>ofertas|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: borrar avisa a la empresa (evento oferta.eliminada).|
|HU-42|Inserción Laboral|Eliminar ofertas rechazadas o<br>publicadas|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: auditoría de ofertas con endpoint y pantalla.|
|HU-43|Inserción Laboral|Generar CV desde el perfil|Implementado|-|
|HU-44|Inserción Laboral|Notificación de oferta compatible|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: cruce requisito = habilidad y aviso a participantes.|
|HU-45|Inserción Laboral|Buscar ofertas por título|Implementado|-|
|HU-46|Inserción Laboral|Buscar ofertas por empresa|Implementado|-|
|HU-47|Campañas|Crear campañas|Implementado|La imagen es una URL, no se sube un archivo.|
|HU-48|Campañas|Editar campañas|Implementado|-|
|HU-49|Campañas|Publicar campañas|Implementado|Antes avisaba dos veces (al crear el borrador y al<br>publicar).|
|HU-50|Campañas|Ver campañas publicadas sin<br>sesión|IMPLEMENTADO|15/11: paginación del listado público (limit/offset, 'Ver más').|
|HU-51|Campañas|Buscar campañas por título|Implementado|-|
|HU-52|Dashboard y<br>reportes|Indicadores clave|Implementado|-|
|HU-53|Dashboard y<br>reportes|Gráficas de evolución temporal|Implementado|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 8 

|**HU**|**Módulo**|**Historia**|**Estado**|**Nota principal**|
|---|---|---|---|---|
|HU-54|Dashboard y<br>reportes|Filtros por período y módulo|Implementado|-|
|HU-55|Dashboard y<br>reportes|Exportar reportes en PDF|Implementado|Permisos de la SA del dashboard en GCP<br>(run.invoker, storage.objectAdmin).|
|HU-56|Observabilidad|Módulo de observabilidad|IMPLEMENTADO -<br>TESTING PENDIENTE|15/11: script de panel y alertas en Cloud Monitoring; falta correrlo.|
|HU-57|Integraciones<br>externas|Mapa con la ubicación del centro|Implementado|-|
|HU-58|Integraciones<br>externas|Integración con Gemini para<br>habilidades|Implementado|-|
|HU-59|Integraciones<br>externas|Servicio externo de CV|Implementado|-|
|HU-60|Notificaciones|Emails ante eventos clave|BLOQUEADO|Depende de las suscripciones push en GCP.|
|HU-61|Notificaciones|Notificaciones push (PWA)|BLOQUEADO|Faltan credenciales de FCM en testing.|
|HU-62|Notificaciones|Configurar qué notificaciones<br>recibir|Implementado|-|



#### **Autenticación y usuarios** 

##### **HU-01: Registrar usuarios con nombre, email, contraseña, teléfono y rol** 

**Actualización 15/11/2026: IMPLEMENTADO.** Gmail obligatorio solo para participantes (decisión de la usuaria). Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Parcialmente implementado|
|Frontend|pages/register, dashboard/usuarios|
|Backend|auth: UserService.CreateUser, AuthService.Register|
|Endpoints|POST /api/v1/users, POST /api/v1/auth/register|
|Persistencia|PostgreSQL (Neon), tabla users + perfil por rol|
|Tests existentes|auth services/controllers (validador de contraseña, email duplicado, roles permitidos)|
|Tests que faltan|Test de dominio de email|
|Errores detectados|No se valida que el email sea de dominio gmail (criterio de la HU).|
|Correcciones realizadas|-|
|Pendientes|Agregar la validación de dominio gmail en el DTO de registro y en el formulario (o confirmar con el equipo si el<br>criterio sigue vigente).|



##### **HU-02: Iniciar sesión con email y contraseña** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** bloqueo 15 min tras 5 intentos (423). Falta migrar la tabla y probar en testing. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Usuario registrado|
|Estado|Parcialmente implementado|
|Frontend|pages/login|
|Backend|auth: AuthService.Login|
|Endpoints|POST /api/v1/auth/login|
|Persistencia|users, refresh_tokens|
|Tests existentes|auth_service_test, auth_controller_test, k6_auth_load_test.js (CD)|
|Tests que faltan|Test de bloqueo tras 5 intentos|
|Errores detectados|No existe el bloqueo temporal tras 5 intentos fallidos de login (solo hay límite de intentos en el código 2FA y<br>rate limit por IP).|
|Correcciones realizadas|-|
|Pendientes|Contador de intentos fallidos por cuenta con bloqueo temporal (por ejemplo 15 minutos) y su test.|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 9 

##### **HU-02bis: Recuperar contraseña por enlace al email** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Usuario registrado|
|Estado|Implementado|
|Frontend|pages/change-password|
|Backend|auth: PasswordRecoveryService|
|Endpoints|POST /api/v1/auth/forgot-password, POST /api/v1/auth/reset-password|
|Persistencia|password_recoveries (vence a los 30 min, uso único)|
|Tests existentes|password_recovery_service tests|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|Verificar el envío real del email en testing (Resend).|



##### **HU-03: Validar el JWT en cada solicitud protegida** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** tokens rechazados en la auditoría de auth. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Sistema|
|Estado|Parcialmente implementado|
|Frontend|interceptors (sesión expirada)|
|Backend|AuthMiddleware en cada servicio (HS256, type=access, vencimiento)|
|Endpoints|Todos los protegidos|
|Persistencia|-|
|Tests existentes|middlewares_test por servicio; sonda de seguridad: alg=none, otra clave, vencido y refresh rechazados con<br>401|
|Tests que faltan|-|
|Errores detectados|Las validaciones fallidas van al log técnico (slog), no a un log de auditoría consultable.|
|Correcciones realizadas|-|
|Pendientes|Registrar los rechazos de token en una auditoría central (ver HU-10).|



##### **HU-04: Expiración configurable y refresh tokens de uso único** 

**Actualización 15/11/2026: IMPLEMENTADO.** reutilizar un refresh cierra todas las sesiones. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Responsable de seguridad|
|Estado|Parcialmente implementado|
|Frontend|auth.service (refresh)|
|Backend|JWTService (JWT_EXPIRATION_MINUTES, JWT_REFRESH_EXPIRATION_DAYS), AuthService.Refresh|
|Endpoints|POST /api/v1/auth/refresh|
|Persistencia|refresh_tokens (revocado al rotar)|
|Tests existentes|auth_service_refresh_test|
|Tests que faltan|Test de reutilización que invalide la sesión|
|Errores detectados|Reutilizar un refresh token ya usado se rechaza, pero no invalida el resto de la sesión (la HU lo pide).|
|Correcciones realizadas|-|
|Pendientes|Ante reutilización, revocar todos los refresh tokens del usuario (RevokeAllForUser ya existe).|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 10 

##### **HU-05: Autenticación de doble factor (2FA)** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Usuario|
|Estado|Implementado|
|Frontend|login, mi-perfil|
|Backend|TwofaService (TOTP / código), dispositivos de confianza|
|Endpoints|POST /api/v1/auth/verify-2fa, /profile/me/2fa/*|
|Persistencia|users (2FA), almacén de desafíos|
|Tests existentes|twofa tests en auth|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-06: Control de acceso RBAC con vistas por rol** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** matriz de permisos de solo lectura, 403 auditados y guards. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Parcialmente implementado|
|Frontend|guards (admin, adminOrVolunteer, teacher, participant), menú por rol (header)|
|Backend|RequireRole en cada servicio + controles por dueño en servicios|
|Endpoints|Todos|
|Persistencia|roles, permissions (auth)|
|Tests existentes|routes_test RBAC por servicio; sonda de seguridad 39/39; tests nuevos de permisos (auth, labor, campaigns,<br>education)|
|Tests que faltan|E2E por rol en el navegador|
|Errores detectados|(1) No hay panel de configuración de permisos por rol en el front (solo la API /roles y /permissions). (2) No se<br>registran los intentos no autorizados en auditoría. (3) El docente puede crear cursos en borrador por<br>/api/docente/cursos, y la HU dice que no.|
|Correcciones realizadas|Se corrigieron 6 problemas de permisos (ver Problema 4): labor confiaba en headers del cliente, editar y borrar<br>ofertas sin control, perfiles creados a nombre de otro usuario, baja de perfil por el propio participante,<br>Seguimiento roto para admin y voluntario.|
|Pendientes|Panel RBAC en el front; auditoría de accesos denegados; decidir si el docente puede proponer cursos.|



##### **HU-07: Crear perfiles de participantes** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** DNI, dirección y teléfono obligatorios; alta auditada. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador, voluntario|
|Estado|Parcialmente implementado|
|Frontend|dashboard/usuarios|
|Backend|UserService.CreateUser (voluntario solo crea participantes), ParticipantService|
|Endpoints|POST /api/v1/users, POST /api/v1/participants|
|Persistencia|users, participants|
|Tests existentes|user_service_test, participant_service_test (+ TestParticipantService_PermisosDeAltaYBaja)|
|Tests que faltan|Test de campos obligatorios completos|
|Errores detectados|El backend no exige DNI, dirección ni teléfono del participante (la HU los pide). No hay registro en el log de<br>auditoría.|
|Correcciones realizadas|Un usuario sin rol de gestión ya no puede crear un perfil a nombre de otra cuenta.|
|Pendientes|Hacer obligatorios los campos de la HU en el DTO; auditoría de altas.|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 11 

##### **HU-08: Editar perfiles de participantes** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** ediciones auditadas con nombres de campos. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador, voluntario|
|Estado|Parcialmente implementado|
|Frontend|dashboard/usuarios|
|Backend|UserService.UpdateUser, ParticipantService.UpdateParticipant|
|Endpoints|PUT /api/v1/users/{id}, PUT /api/v1/participants/{id}|
|Persistencia|users, participants|
|Tests existentes|user_service_test|
|Tests que faltan|Test de auditoría con campos modificados|
|Errores detectados|La edición no queda en un log de auditoría con el detalle de campos modificados.|
|Correcciones realizadas|-|
|Pendientes|Auditoría de modificaciones en auth.|



##### **HU-09: Eliminar perfiles de participantes (baja lógica)** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** bajas auditadas. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador, voluntario|
|Estado|Parcialmente implementado|
|Frontend|dashboard/usuarios (confirmación)|
|Backend|ParticipantService.DeleteParticipant (soft delete GORM)|
|Endpoints|DELETE /api/v1/participants/{id}, DELETE /api/v1/users/{id}|
|Persistencia|participants.deleted_at|
|Tests existentes|participant_service_test|
|Tests que faltan|-|
|Errores detectados|La baja no queda en un log de auditoría.|
|Correcciones realizadas|El propio participante ya no puede dar de baja su perfil de participante (solo admin y voluntario, como dice la<br>HU).|
|Pendientes|Auditoría de bajas.|



##### **HU-10: Log de auditoría de acciones sobre perfiles** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** tabla auditoria inmutable, endpoint y pantalla. Falta migrar y desplegar. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|No implementado|
|Frontend|dashboard/auditoria (muestra Educación y Promoción Humana)|
|Backend|auth no tiene auditoría; Educación y PH sí (colección auditoria)|
|Endpoints|GET /api/auditoria, GET /api/v1/human-promotion/auditoria|
|Persistencia|Firestore auditoria (education, PH)|
|Tests existentes|audit tests de education y PH|
|Tests que faltan|Tests de auditoría en auth|
|Errores detectados|Las altas, bajas y modificaciones de perfiles (auth) no generan registro de auditoría. Sí existe auditoría<br>inmutable y filtrable en Educación y Promoción Humana.|
|Correcciones realizadas|-|
|Pendientes|Tabla de auditoría en auth (usuario, acción, recurso, fecha) en la misma transacción, y sumarla a la pantalla de<br>auditoría.|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 12 

##### **HU-11: Foto y documentación del participante** 

**Actualización 15/11/2026: IMPLEMENTADO.** límite de 5 MB en foto y documentos; cambios auditados. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Voluntario|
|Estado|Parcialmente implementado|
|Frontend|dashboard/usuarios (documentos)|
|Backend|ParticipantService.UploadDocument (PDF/JPG/PNG/WEBP, 10 MB, MIME real)|
|Endpoints|/api/v1/participants/{id}/documents|
|Persistencia|participant_documents + bucket GCS|
|Tests existentes|participant_document_test|
|Tests que faltan|-|
|Errores detectados|Los cambios no quedan en auditoría.|
|Correcciones realizadas|-|
|Pendientes|Auditoría; verificar límite de 5 MB para la foto de perfil.|



##### **HU-12: Aprobar el registro de voluntarios** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** rechazo con email 'cuenta_rechazada'. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Coordinador / administrador|
|Estado|Parcialmente implementado|
|Frontend|pages/pending-approval, dashboard/usuarios|
|Backend|approval_status + canApprove (solo admin/dev)|
|Endpoints|PUT /api/v1/volunteers/{id}, PUT /api/v1/users/{id}|
|Persistencia|volunteers.approval_status|
|Tests existentes|approval_helper tests|
|Tests que faltan|-|
|Errores detectados|Se notifica la aprobación (email 'cuenta_aprobada'), pero no el rechazo.|
|Correcciones realizadas|-|
|Pendientes|Email y aviso de rechazo.|



##### **HU-13: Iniciar sesión con Google** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Usuario|
|Estado|Implementado|
|Frontend|login (Firebase Auth)|
|Backend|AuthService (google_auth_id, vincular/desvincular)|
|Endpoints|POST /api/v1/auth/google-check-credentials, /profile/me/link-google|
|Persistencia|users.google_auth_id|
|Tests existentes|auth tests|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



#### **Educación** 

Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 13 

##### **HU-14: Crear cursos (con materiales opcionales)** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** aviso de los cursos que usan el material antes de borrarlo. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Implementado|
|Frontend|dashboard/gestion-cursos (admin-cursos)|
|Backend|CourseService.CreateCourse (Contentful CMA)|
|Endpoints|POST /api/cursos, POST /api/cursos/{id}/publicar, POST /api/materiales[/upload]|
|Persistencia|Contentful (course, material)|
|Tests existentes|course_service_test (+ SinMateriales), admin-cursos.spec (+4 casos)|
|Tests que faltan|E2E de alta completa contra Contentful de testing|
|Errores detectados|Antes de la auditoría no se podía crear un curso sin materiales (formulario, backend y Contentful lo exigían) y<br>no se podía subir material desde el formulario del curso.|
|Correcciones realizadas|Problema 1: materiales opcionales en las tres capas; botón 'Subir material nuevo' dentro del formulario del<br>curso; aviso claro si el curso queda en borrador.|
|Pendientes|-|



##### **HU-15: Editar cursos** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Implementado|
|Frontend|admin-cursos (mensaje 'Se modificó el curso correctamente')|
|Backend|CourseService.UpdateCourse|
|Endpoints|PUT /api/cursos/{id}|
|Persistencia|Contentful|
|Tests existentes|course_service_test, admin-cursos.spec|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-16: Eliminar cursos (baja lógica)** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Implementado|
|Frontend|admin-cursos (aviso de inscriptos activos)|
|Backend|CourseService.DeleteCourse (cursos_baja + auditoría)|
|Endpoints|DELETE /api/cursos/{id}, GET /api/cursos/{id}/inscriptos-activos|
|Persistencia|Firestore cursos_baja, auditoria|
|Tests existentes|baja_y_participantes_test|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 14 

##### **HU-17: Contenido de cursos gestionado en Contentful** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Implementado|
|Frontend|-|
|Backend|Webhook de Contentful + caché con contenido stale (ADR 0006)|
|Endpoints|POST /webhooks/contentful|
|Persistencia|Contentful + caché en memoria|
|Tests existentes|content_cache tests, contentful_webhook_controller_test|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-18: Inscribir participantes a cursos** 

**Actualización 15/11/2026: BLOQUEADO.** Depende de Pub/Sub y Resend en testing (acceso a GCP). Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Voluntario, responsable|
|Estado|Pendiente de verificación|
|Frontend|dashboard/inscripciones|
|Backend|EnrollmentService.Enroll (duplicado, cupo, curso dado de baja)|
|Endpoints|POST/GET /api/cursos/{id}/inscripciones|
|Persistencia|Firestore inscripciones + outbox|
|Tests existentes|enrollment_enroll_test, inscripciones.spec|
|Tests que faltan|Test de punta a punta del email|
|Errores detectados|El email de confirmación depende de Notificaciones vía Pub/Sub; localmente el emulador había quedado sin<br>suscripciones (corregido) y en testing no se pudo verificar.|
|Correcciones realizadas|Emulador de Pub/Sub autoinicializado (Problema 3).|
|Pendientes|Verificar en testing que llegue el email (eventos + suscripción push de notifications).|



##### **HU-19: Asistencia por cumplimiento de actividades** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Voluntario, docente|
|Estado|Implementado|
|Frontend|asistencia-cursos, asistencia (docente)|
|Backend|Entrega -> Asistencia en una transacción|
|Endpoints|/api/actividades/*, /api/participants/{id}/asistencia|
|Persistencia|Firestore entregas, asistencias|
|Tests existentes|submission_service_test|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 15 

##### **HU-19bis: Asistencia general diaria del centro** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Voluntario, administrador|
|Estado|Implementado|
|Frontend|asistencia-general|
|Backend|GeneralAttendanceService (upsert por fecha) + auditoría|
|Endpoints|GET/POST /api/asistencia-general|
|Persistencia|Firestore asistencia_general, auditoria|
|Tests existentes|general_attendance_service_test|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-20: Progreso formativo por actividades** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Participante, voluntario|
|Estado|Implementado|
|Frontend|seguimiento, curso-detalle (participante)|
|Backend|ProgressService|
|Endpoints|GET /api/cursos/{id}/progreso (nuevo), /api/docente/cursos/{id}/progreso|
|Persistencia|Firestore|
|Tests existentes|progress_controller_test (nuevo), seguimiento.spec (+3)|
|Tests que faltan|-|
|Errores detectados|La pantalla Seguimiento estaba en el menú de admin y voluntario, pero solo usaba endpoints de docente: 403<br>al voluntario, lista vacía al admin.|
|Correcciones realizadas|Endpoint GET /api/cursos/{id}/progreso para admin y voluntario; Seguimiento elige los endpoints según el rol y<br>muestra nombres.|
|Pendientes|-|



##### **HU-21: Sugerencia de habilidades con IA** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Implementado|
|Frontend|curso-detalle/activity-skills|
|Backend|SkillSuggestionService (Gemini por Vertex, OpenAI de respaldo)|
|Endpoints|POST /api/actividades/sugerir-habilidades|
|Persistencia|Firestore habilidades, actividad_habilidades|
|Tests existentes|skill_suggestion_service_test|
|Tests que faltan|-|
|Errores detectados|Diferencia con la HU: se sugiere por actividad y no al crear el curso (decisión del equipo,<br>docs/ia-habilidades.md).|
|Correcciones realizadas|-|
|Pendientes|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 16 

##### **HU-22: Docente externo consulta contenidos** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Docente externo|
|Estado|Implementado|
|Frontend|mis-contenidos|
|Backend|TeacherContentService|
|Endpoints|GET /api/biblioteca, /api/contenidos-docentes/mios|
|Persistencia|Firestore contenidos_docentes|
|Tests existentes|teacher_content_service_test|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-23: Docente externo sube contenido con moderación** 

**Actualización 15/11/2026: BLOQUEADO.** Depende de Pub/Sub en testing (acceso a GCP). Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Docente externo, administrador|
|Estado|Pendiente de verificación|
|Frontend|mis-contenidos, contenido-docente|
|Backend|TeacherContentService (pendiente/aprobado/rechazado)|
|Endpoints|/api/contenidos-docentes/*|
|Persistencia|Firestore + bucket|
|Tests existentes|teacher_content_routes_test|
|Tests que faltan|Test de punta a punta de la notificación|
|Errores detectados|La notificación del resultado depende de Pub/Sub -> Notificaciones; no se pudo verificar en testing.|
|Correcciones realizadas|Enlaces de las notificaciones corregidos (/dashboard/mis-contenidos).|
|Pendientes|Verificar la notificación en testing.|



##### **HU-24: Comentarios en contenido público** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Visitante / usuario|
|Estado|Implementado|
|Frontend|campanas-publico, cursos|
|Backend|CommentService (sanitizado, rate limit)|
|Endpoints|GET/POST /api/comentarios, DELETE (admin)|
|Persistencia|Firestore comentarios|
|Tests existentes|comment_service_test, comment_routes_test|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 17 

##### **HU-25: Registro de datos al descargar de la biblioteca** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Visitante|
|Estado|Implementado|
|Frontend|landing/library|
|Backend|TeacherContentService.PublicDownload (valida teléfono)|
|Endpoints|POST /api/biblioteca/{id}/descargas|
|Persistencia|Firestore descargas_biblioteca|
|Tests existentes|teacher_content_service_test|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-26: Email al inscribirse y recordatorios de sesión** 

**Actualización 15/11/2026: BLOQUEADO.** Depende de Pub/Sub, Cloud Scheduler y Resend (acceso a GCP). Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Participante|
|Estado|Pendiente de verificación|
|Frontend|-|
|Backend|Reglas InscripcionCreada y ActividadRecordatorio (notifications), ReminderService (education)|
|Endpoints|/internal/reminders (Cloud Scheduler)|
|Persistencia|notifications (Postgres)|
|Tests existentes|reminder_service_test, rules_test|
|Tests que faltan|Test de punta a punta|
|Errores detectados|Depende de Pub/Sub, Cloud Scheduler y Resend; no verificado en testing.|
|Correcciones realizadas|-|
|Pendientes|Configurar el job de Scheduler y verificar envíos.|



##### **HU-27: Buscar cursos por nombre** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Usuario|
|Estado|Implementado|
|Frontend|cursos|
|Backend|FiltrarCursos (sin tildes)|
|Endpoints|GET /api/cursos?q=|
|Persistencia|-|
|Tests existentes|course_service_test|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 18 

##### **HU-28: Buscar cursos por área formativa** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Usuario|
|Estado|Implementado|
|Frontend|cursos|
|Backend|FiltrarCursosPorArea|
|Endpoints|GET /api/cursos?area=&q=|
|Persistencia|-|
|Tests existentes|course_service_test|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



#### **Promoción Humana** 

##### **HU-29: Crear ficha de proceso personal** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador, voluntario|
|Estado|Implementado|
|Frontend|promocion-humana/ficha|
|Backend|FichaService (cifrado de campos, una ficha por participante)|
|Endpoints|POST /api/v1/human-promotion/fichas|
|Persistencia|Firestore fichas + auditoria + outbox|
|Tests existentes|ficha tests, events_test|
|Tests que faltan|Specs de front de la ficha|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-30: Editar ficha** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador, voluntario|
|Estado|Implementado|
|Frontend|ficha|
|Backend|FichaService.Update + auditoría|
|Endpoints|PUT /api/v1/human-promotion/fichas/{id}|
|Persistencia|Firestore|
|Tests existentes|events_test (+ TestFichaActualizada_LlevaCreadoPor)|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|El evento ficha.actualizada ahora lleva creado_por (Problema 3).|
|Pendientes|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 19 

##### **HU-31: Campos estructurados de la ficha** 

**Actualización 15/11/2026: BLOQUEADO.** Decisión del equipo (mantener los datos leídos de Auth y Educación). Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador, voluntario|
|Estado|Parcialmente implementado|
|Frontend|ficha|
|Backend|Ficha: situacion_actual, avances (cifrados)|
|Endpoints|/fichas|
|Persistencia|Firestore|
|Tests existentes|ficha tests|
|Tests que faltan|-|
|Errores detectados|Nombre, apellido y edad se muestran desde el perfil (auth) y las habilidades desde Educación; no son campos<br>editables de la ficha como pide la HU. Es una decisión para no duplicar datos personales, pero difiere del texto<br>de la HU.|
|Correcciones realizadas|-|
|Pendientes|Confirmar con el equipo y, si corresponde, actualizar la HU.|



##### **HU-32: Registrar hitos de avance** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador, voluntario|
|Estado|Implementado|
|Frontend|ficha|
|Backend|HitoService|
|Endpoints|POST/GET/PUT /fichas/{id}/hitos|
|Persistencia|Firestore hitos|
|Tests existentes|hito tests|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-33: Línea de tiempo del participante** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Voluntario, responsable|
|Estado|Implementado|
|Frontend|ficha (línea de tiempo mezclada)|
|Backend|ActividadService|
|Endpoints|GET /fichas/{id}/actividad|
|Persistencia|Firestore|
|Tests existentes|actividad tests|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 20 

##### **HU-34: Línea de tiempo propia del participante** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Participante|
|Estado|Implementado|
|Frontend|mi-proceso|
|Backend|mi-linea-de-tiempo (solo hitos visibles, sin datos confidenciales)|
|Endpoints|GET /api/v1/human-promotion/mi-linea-de-tiempo|
|Persistencia|Firestore|
|Tests existentes|participante tests|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-35: Auditoría de modificaciones de fichas** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Implementado|
|Frontend|dashboard/auditoria|
|Backend|audit.RecordTx en la misma transacción|
|Endpoints|GET /api/v1/human-promotion/auditoria|
|Persistencia|Firestore auditoria|
|Tests existentes|events_test|
|Tests que faltan|-|
|Errores detectados|La HU menciona 'eliminar' fichas: no existe eliminación de fichas (no hay nada que auditar).|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-36: Buscar participantes por nombre** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Voluntario, responsable|
|Estado|Implementado|
|Frontend|promocion-humana|
|Backend|ParticipanteController.Buscar|
|Endpoints|GET /api/v1/human-promotion/participantes|
|Persistencia|auth (directorio)|
|Tests existentes|participante tests|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 21 

##### **HU-37: Exportar ficha en PDF** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Voluntario, responsable|
|Estado|Implementado|
|Frontend|ficha|
|Backend|pdf_service (mismos permisos que la ficha)|
|Endpoints|GET /fichas/{id}/pdf|
|Persistencia|-|
|Tests existentes|pdf tests|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



#### **Inserción Laboral** 

##### **HU-38: Cuenta de empresa con aprobación** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** el rechazo de empresas manda email. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Empresa|
|Estado|Parcialmente implementado|
|Frontend|pending-approval, ofertas|
|Backend|approval_status de company; CreateOffer exige empresa aprobada|
|Endpoints|/api/v1/companies|
|Persistencia|companies|
|Tests existentes|company_service_test|
|Tests que faltan|-|
|Errores detectados|Se notifica la aprobación, no el rechazo.|
|Correcciones realizadas|-|
|Pendientes|Notificar el rechazo.|



##### **HU-39: Publicar ofertas laborales** 

**Actualización 15/11/2026: IMPLEMENTADO.** todos los campos de la HU obligatorios (backend y formulario). Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Empresa|
|Estado|Parcialmente implementado|
|Frontend|ofertas|
|Backend|JobOfferService.CreateOffer|
|Endpoints|POST /api/v1/labor-integration/offers|
|Persistencia|Firestore job_offers|
|Tests existentes|job_offer_permisos_test (nuevo)|
|Tests que faltan|Test de campos obligatorios|
|Errores detectados|Solo se exigen título y descripción; la HU pide todos los campos obligatorios (ubicación, modalidad, tipo,<br>salario, horario, requisitos, fecha límite).|
|Correcciones realizadas|Solo empresa, admin y voluntario pueden crear ofertas (antes cualquier rol).|
|Pendientes|Validar todos los campos en el DTO y en el formulario.|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 22 

##### **HU-40: Ofertas pendientes hasta aprobación** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Sistema|
|Estado|Implementado|
|Frontend|ofertas|
|Backend|isOfferVisibleForRole, estadoPermitido|
|Endpoints|GET/PUT /offers|
|Persistencia|Firestore|
|Tests existentes|job_offer_permisos_test|
|Tests que faltan|-|
|Errores detectados|Antes: cualquier rol distinto de participante veía las pendientes, y la empresa podía autoaprobarse editando el<br>estado. Excepción vigente: las ofertas creadas por admin o voluntario nacen 'Activa'.|
|Correcciones realizadas|Pendientes solo visibles para gestión y empresa dueña; solo el admin cambia el estado.|
|Pendientes|Decidir si las ofertas del voluntario también requieren aprobación.|



##### **HU-41: Aprobar, rechazar o eliminar ofertas** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** borrar avisa a la empresa (evento oferta.eliminada). Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Parcialmente implementado|
|Frontend|ofertas|
|Backend|Approve/Reject (solo admin), DeleteOffer|
|Endpoints|PUT /offers/{id}/approve|reject, DELETE /offers/{id}|
|Persistencia|Firestore|
|Tests existentes|job_offer_controller_test, job_offer_permisos_test|
|Tests que faltan|-|
|Errores detectados|Eliminar no notifica a la empresa. Antes de la auditoría cualquier usuario con sesión podía borrar ofertas.|
|Correcciones realizadas|Borrar: gestión o empresa dueña; resto 403.|
|Pendientes|Notificar a la empresa al eliminar.|



##### **HU-42: Eliminar ofertas rechazadas o publicadas** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** auditoría de ofertas con endpoint y pantalla. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Parcialmente implementado|
|Frontend|ofertas|
|Backend|DeleteOffer (baja lógica si tiene postulaciones)|
|Endpoints|DELETE /offers/{id}|
|Persistencia|Firestore|
|Tests existentes|job_offer_permisos_test|
|Tests que faltan|-|
|Errores detectados|Inserción laboral no tiene log de auditoría.|
|Correcciones realizadas|-|
|Pendientes|Auditoría de eliminaciones.|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 23 

##### **HU-43: Generar CV desde el perfil** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Participante|
|Estado|Implementado|
|Frontend|cv|
|Backend|rxresume + cv-renderer|
|Endpoints|/api/v1/rxresume/*|
|Persistencia|rxresume (Postgres)|
|Tests existentes|rxresume tests (pocos)|
|Tests que faltan|Más tests de rxresume (38 % de cobertura)|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-44: Notificación de oferta compatible** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** cruce requisito = habilidad y aviso a participantes. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Participante|
|Estado|No implementado|
|Frontend|-|
|Backend|OfertaCreadaRule avisa a admin y voluntario|
|Endpoints|-|
|Persistencia|-|
|Tests existentes|rules_test|
|Tests que faltan|-|
|Errores detectados|No hay cruce de compatibilidad entre ofertas y perfiles; los participantes no reciben aviso.|
|Correcciones realizadas|-|
|Pendientes|Definir 'compatible' (habilidades / CV) y la regla de notificación.|



##### **HU-45: Buscar ofertas por título** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Usuario|
|Estado|Implementado|
|Frontend|ofertas|
|Backend|FiltrarOfertas|
|Endpoints|GET /offers?q=|
|Persistencia|-|
|Tests existentes|labor services tests|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 24 

##### **HU-46: Buscar ofertas por empresa** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Usuario|
|Estado|Implementado|
|Frontend|ofertas|
|Backend|FiltrarOfertas|
|Endpoints|GET /offers?empresa=|
|Persistencia|-|
|Tests existentes|labor services tests|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



#### **Campañas** 

##### **HU-47: Crear campañas** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Responsable, voluntario|
|Estado|Implementado|
|Frontend|dashboard/campanas|
|Backend|CampaignService.Create|
|Endpoints|POST /api/v1/campaigns|
|Persistencia|Firestore campanas|
|Tests existentes|campaign_service_test, routes_test|
|Tests que faltan|-|
|Errores detectados|La imagen es una URL, no se sube un archivo.|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-48: Editar campañas** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Responsable, voluntario|
|Estado|Implementado|
|Frontend|campanas|
|Backend|CampaignService.Update|
|Endpoints|PUT /api/v1/campaigns/{id}|
|Persistencia|Firestore|
|Tests existentes|campaign_service_test|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|Una campaña dada de baja ya no se puede editar.|
|Pendientes|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 25 

##### **HU-49: Publicar campañas** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Responsable, voluntario|
|Estado|Implementado|
|Frontend|campanas|
|Backend|CampaignService.Publish|
|Endpoints|POST /api/v1/campaigns/{id}/publicar|
|Persistencia|Firestore|
|Tests existentes|campaign_service_test|
|Tests que faltan|-|
|Errores detectados|Antes avisaba dos veces (al crear el borrador y al publicar).|
|Correcciones realizadas|El aviso campana.creada sale una sola vez, al publicar.|
|Pendientes|-|



##### **HU-50: Ver campañas publicadas sin sesión** 

**Actualización 15/11/2026: IMPLEMENTADO.** paginación del listado público (limit/offset, 'Ver más'). Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Visitante|
|Estado|Implementado|
|Frontend|campanas-publico|
|Backend|ListPublished|
|Endpoints|GET /api/v1/campaigns/publicas|
|Persistencia|Firestore|
|Tests existentes|routes_test|
|Tests que faltan|Paginación (ver prueba de carga)|
|Errores detectados|Ninguno|
|Correcciones realizadas|Las dadas de baja no se muestran.|
|Pendientes|-|



##### **HU-51: Buscar campañas por título** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Usuario|
|Estado|Implementado|
|Frontend|campanas-publico|
|Backend|listar (sin tildes)|
|Endpoints|GET /publicas?q=|
|Persistencia|-|
|Tests existentes|campaign_service_test|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



#### **Dashboard y reportes** 

Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 26 

##### **HU-52: Indicadores clave** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Implementado|
|Frontend|indicadores|
|Backend|dashboard: services/indicadores.go (7 indicadores)|
|Endpoints|GET /api/v1/dashboard/resumen|
|Persistencia|agrega de auth, education, labor, PH|
|Tests existentes|dashboard tests (85,8 %)|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-53: Gráficas de evolución temporal** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Implementado|
|Frontend|indicadores|
|Backend|resumen por período|
|Endpoints|GET /api/v1/dashboard/resumen|
|Persistencia|-|
|Tests existentes|dashboard tests|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-54: Filtros por período y módulo** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Implementado|
|Frontend|indicadores|
|Backend|?desde&hasta&modulo&area|
|Endpoints|GET /api/v1/dashboard/resumen|
|Persistencia|-|
|Tests existentes|dashboard tests|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 27 

##### **HU-55: Exportar reportes en PDF** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Implementado|
|Frontend|indicadores (reportes.service)|
|Backend|ReporteExportado (Firestore + PDF en Cloud Storage)|
|Endpoints|/api/v1/dashboard/reportes|
|Persistencia|Firestore reportes_exportados, bucket|
|Tests existentes|dashboard tests, reportes.service.spec|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|Permisos de la SA del dashboard en GCP (run.invoker, storage.objectAdmin).|



#### **Observabilidad** 

##### **HU-56: Módulo de observabilidad** 

**Actualización 15/11/2026: IMPLEMENTADO - TESTING PENDIENTE.** script de panel y alertas en Cloud Monitoring; falta correrlo. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Desarrollador|
|Estado|Parcialmente implementado|
|Frontend|-|
|Backend|Logs JSON estructurados; métricas nativas de Cloud Run / Cloud Logging|
|Endpoints|/health por servicio|
|Persistencia|Cloud Logging|
|Tests existentes|smoke tests del CD|
|Tests que faltan|-|
|Errores detectados|No hay panel propio en una URL privada ni métricas por servicio exportadas por la app; se depende de la<br>consola de GCP.|
|Correcciones realizadas|-|
|Pendientes|Dashboard de Cloud Monitoring (errores, latencia, CPU/memoria por servicio) con alertas, o Grafana.|



#### **Integraciones externas** 

##### **HU-57: Mapa con la ubicación del centro** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Usuario|
|Estado|Implementado|
|Frontend|landing/location, contacto (iframe de Google Maps)|
|Backend|-|
|Endpoints|-|
|Persistencia|-|
|Tests existentes|specs de landing|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 28 

##### **HU-58: Integración con Gemini para habilidades** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Administrador|
|Estado|Implementado|
|Frontend|activity-skills|
|Backend|Vertex AI Gemini + OpenAI de respaldo|
|Endpoints|POST /api/actividades/sugerir-habilidades|
|Persistencia|Firestore|
|Tests existentes|skill_suggestion_service_test, skills_ai_test|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



##### **HU-59: Servicio externo de CV** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Participante|
|Estado|Implementado|
|Frontend|cv|
|Backend|rxresume / Reactive Resume|
|Endpoints|/api/v1/rxresume/*|
|Persistencia|rxresume|
|Tests existentes|rxresume tests|
|Tests que faltan|Test del error claro cuando el servicio no está|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



#### **Notificaciones** 

##### **HU-60: Emails ante eventos clave** 

**Actualización 15/11/2026: BLOQUEADO.** Depende de las suscripciones push en GCP. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Usuario|
|Estado|Pendiente de verificación|
|Frontend|notificaciones|
|Backend|notifications: reglas por evento + delivery worker + Resend|
|Endpoints|/eventos/*, /internal/delivery/flush|
|Persistencia|Postgres notifications|
|Tests existentes|rules_test, worker_test, consumers_test|
|Tests que faltan|Test de punta a punta con Pub/Sub|
|Errores detectados|Localmente el emulador no tenía suscripciones y el contenedor de notifications corre sin AUTH_GRPC_URL<br>(stack viejo). En testing no hay variables PUBSUB_PUSH_* en el CD.|
|Correcciones realizadas|Emulador autoinicializado; enlaces rotos de las notificaciones; sin notificaciones huérfanas por reintento. Stack<br>reconstruido el 12/11.|
|Pendientes|Verificar suscripciones push en GCP.|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 29 

##### **HU-61: Notificaciones push (PWA)** 

**Actualización 15/11/2026: BLOQUEADO.** Faltan credenciales de FCM en testing. Detalle en la sección 16.

|**Campo**|**Detalle**|
|---|---|
|Actor|Usuario con la PWA|
|Estado|Parcialmente implementado|
|Frontend|push-notification.service (FCM)|
|Backend|FirebasePushService|
|Endpoints|POST /api/v1/users/me/devices|
|Persistencia|devices|
|Tests existentes|push_service_test|
|Tests que faltan|-|
|Errores detectados|El push queda inactivo si no hay credenciales de Firebase (localmente no arranca; FIREBASE_PROJECT_ID<br>no existe en el environment de testing).|
|Correcciones realizadas|-|
|Pendientes|Configurar credenciales de FCM en testing y probar en un dispositivo.|



##### **HU-62: Configurar qué notificaciones recibir** 

|**Campo**|**Detalle**|
|---|---|
|Actor|Usuario|
|Estado|Implementado|
|Frontend|notificaciones (preferencias por categoría y canal)|
|Backend|NotificationPreference.AllowsChannel|
|Endpoints|/api/v1/notifications/preferences|
|Persistencia|Postgres|
|Tests existentes|notification_service_test, notificaciones.spec|
|Tests que faltan|-|
|Errores detectados|Ninguno|
|Correcciones realizadas|-|
|Pendientes|-|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 30 

### **5. Auditoría del backend** 

#### **Estado de los servicios** 

|**Servicio**|**Build**|**Tests**|**Cobertura**|**Observaciones**|
|---|---|---|---|---|
|auth|OK|11/11 paquetes|83,6 %|Sin auditoría; sin bloqueo de login; permisos de alta de perfiles<br>corregidos|
|education|OK|15/15|90,2 %|Problema 1 corregido; endpoint de progreso nuevo|
|human-promotion|OK|11/11|75,8 %|ficha.actualizada con creado_por|
|campaigns|OK|8/8|90,2 %|Baja lógica nueva; aviso solo al publicar|
|dashboard|OK|9/9|85,8 %|-|
|notifications|OK|14/14|63,2 %|Enlaces, huérfanas y verificador en producción corregidos|
|labor-integration|OK|7/7|59,3 %|Vulnerabilidad de identidad y permisos de ofertas corregidos|
|rxresume|OK|2/2|38,4 %|Cobertura baja|
|pkg|OK|2/2|94,0 %|Librería de eventos y outbox|



#### **Arquitectura y organización** 

- Capas controller/service/repository consistentes; interfaces mockeables; errores tipados (ValidationError, ErrNotFound) traducidos a 400/404 sin exponer detalles internos. 

- Variables de entorno validadas al arrancar en education (protege contra mezclar testing y producción de Contentful) y JWT_SECRET obligatorio. 

- Inconsistencias: labor-integration usa http.Error en texto plano mientras el resto responde JSON; labor y auth no tienen auditoría; criterios de rol distintos entre servicios. 

#### **Endpoints** 

- Se revisaron las rutas de los ocho servicios contra las HU (anexo). Se agregaron DELETE /api/v1/campaigns/{id} y GET /api/cursos/{id}/progreso, también en openapi.yaml. 

- El docente puede crear cursos en borrador (POST /api/docente/cursos): la HU-06 no lo permite. Se documenta, no se elimina (funcionalidad existente; decisión del equipo). 

#### **Base de datos** 

- PostgreSQL (Neon) para auth, notifications y rxresume con GORM y soft delete (deleted_at) en perfiles. 

- Firestore para education, PH, campaigns, labor y dashboard; ids determinísticos donde hay unicidad (entregas, asistencias); transacciones para cambio + evento + auditoría. 

- Contentful para cursos, áreas y materiales. En esta auditoría se cambió material a opcional en testing y master. 

- Riesgo: el stack local y testing comparten la misma base Firestore (default) y Neon: cualquier prueba con escritura en local afecta testing. 

#### **Autenticación, autorización y seguridad** 

- JWT HS256 con type=access y vencimiento validado en cada servicio; la sonda rechazó alg=none, otra clave, vencidos y de refresh (401). 

- RBAC por RequireRole más controles por dueño en servicios; 22 casos por rol de la sonda dieron 403 como corresponde. 

- Corregido: identidad por headers en labor; edición y borrado de ofertas; alta de perfiles ajenos; baja de perfil por el participante; verificador de push en producción; puerto de labor expuesto. 

- Abierto: el límite por IP se puede esquivar con X-Forwarded-For (120/120 pedidos pasaron rotando el header); sin bloqueo de login; sin auditoría de accesos denegados. 

- Sin secretos en el repositorio: se revisaron las capturas de despliegue (solo nombres de variables y rutas a Secret Manager). Los .env están en .gitignore. 

Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 31 

- Headers de seguridad presentes en education y PH (CSP, HSTS, X-Frame-Options, nosniff); el gateway nginx local no los agrega y muestra su versión. 

#### **Integración entre servicios** 

- education y dashboard consultan auth por HTTP reenviando el JWT del usuario; notifications y labor usan gRPC o HTTP con ID token en Cloud Run. 

- Pub/Sub: el emulador local quedaba vacío tras un reinicio (corregido). En el CD no hay PUBSUB_PUSH_AUDIENCE ni PUBSUB_PUSH_SERVICE_ACCOUNTS: no se pudo verificar la entrega de eventos en testing. 

- Fallas aisladas: con el resolver de nginx, si un servicio no está, solo fallan sus rutas (502), no todo el gateway. 

### **6. Auditoría del frontend** 

- Pantallas: existen todas las de las HU salvo el panel RBAC (HU-06) y la notificación de ofertas compatibles (HU-44). 

- Datos: las pantallas consumen el backend real; no se encontraron datos ficticios salvo el impacto de la landing (KV del worker con valores por defecto si falla). 

- Navegación y permisos: menú por rol en el header y guards por ruta. Sin guard: usuarios, seguimiento, cv y ofertas (el backend responde igual, pero la pantalla carga). 

- Corregido: formulario de curso (materiales opcionales y subida en el mismo diálogo); botón Eliminar en campañas con confirmación; contador de no leídas en la campana y refresco de notificaciones; Seguimiento por rol con nombres. 

- Manejo de errores: interceptor global, pantallas 404/500, mensajes del backend en toasts. Un 429 del gateway se muestra como error genérico. 

- Tests: 85 archivos de specs, 694 casos, todos OK; 71,45 % de sentencias (umbral 70 %). 

- Cypress: un spec sin ejecutar (no instalado). 

### **7. Matriz de roles y permisos** 

|**Rol**|**Módulos**|**Acciones permitidas**|**Restringidas**|**Estado / problemas**|
|---|---|---|---|---|
|Administrador|Todos|Gestión de usuarios y aprobaciones; cursos,<br>áreas y materiales; moderación; fichas y<br>auditoría; ofertas (aprobar/rechazar/borrar);<br>campañas; dashboard y reportes|-|Implementado (sin panel<br>RBAC; auditoría de auth<br>faltante)|
|Voluntario|Educación, PH,<br>Campañas, Labor|Participantes (alta/edición/baja),<br>inscripciones y asistencia, fichas e hitos,<br>campañas (incluye eliminar), ver ofertas y<br>postular participantes|Dashboard, auditoría,<br>RBAC, crear cursos,<br>aprobar ofertas|Implementado; corregido<br>Seguimiento (antes 403)|
|Participante|Educación, PH,<br>Labor|Sus cursos y entregas, su progreso, su línea<br>de tiempo, ofertas publicadas, su CV,<br>comentarios, preferencias de notificación|Datos de otros, fichas,<br>gestión|Implementado; corregido:<br>no crea perfiles ajenos ni<br>borra su perfil|
|Empresa|Labor|Publicar ofertas (si está aprobada), ver sus<br>ofertas y postulantes, editar/borrar las suyas,<br>campañas públicas|Perfiles de participantes,<br>fichas, aprobar ofertas|Implementado; corregido:<br>no se autoaprueba ni toca<br>ofertas ajenas|
|Docente (interno)|Educación|Sus cursos publicados, actividades,<br>entregas, asistencia, progreso, materiales de<br>sus cursos|Crear/eliminar cursos<br>(pero hoy puede crear<br>borradores), fichas,<br>dashboard|Parcial:<br>/api/docente/cursos POST<br>contradice la HU|
|Docente externo|Educación|Subir contenido (pendiente de moderación),<br>ver estado, editar/borrar mientras está<br>pendiente, biblioteca|Todo lo demás (403 en<br>/api/docente)|Implementado|
|Visitante|Landing|Campañas publicadas, biblioteca (con<br>registro), comentarios, mapa|Todo lo autenticado<br>(401)|Implementado|



La verificación del backend se hizo con la sonda de seguridad (anexo) y con los tests de rutas de cada servicio. 

Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 32 

### **8. Problemas detectados durante las pruebas manuales** 

#### **Problema 1. Educación: creación de cursos y materiales** 

|**Campo**|**Detalle**|
|---|---|
|Descripción|No se podía crear un curso sin materiales y no se podían subir materiales mientras se creaba el curso.|
|Comportamiento esperado|Crear el curso sin materiales; subir materiales durante la creación o después; editar y quitar materiales según<br>permisos.|
|Comportamiento<br>encontrado|El bloqueo estaba en tres capas: (1) el formulario deshabilitaba 'Crear curso' sin materiales, aunque el campo<br>decía '(opcional)'; (2) el backend rechazaba con 400 ('materialIds es requerido'); (3) en Contentful el campo<br>material del content type course era obligatorio, así que publicar fallaba. Los materiales solo se podían subir<br>desde la pestaña Materiales.|
|Causa|Regla de validación heredada del modelo inicial de Contentful ('al menos un material'), replicada en el backend<br>y en el front.|
|Archivos afectados|education/internal/services/course_service.go; frontend/.../admin-cursos.component.{ts,html}; Contentful<br>content type course (testing y master); education/CLAUDE.md|
|Corrección realizada|Se sacó la obligatoriedad en las tres capas (en Contentful con la API, con autorización de la usuaria:<br>required=false, versión 25 en testing y master). Se valida que no lleguen ids vacíos. En el formulario del curso<br>se agregó 'Subir material nuevo': abre el alta de material y lo deja adjuntado sin perder lo escrito. Si el curso se<br>crea pero no se puede publicar, se avisa 'Curso creado en borrador' en lugar de un error. El docente sigue<br>adjuntando materiales a sus cursos desde Mis cursos.|
|Tests ejecutados|Go: TestCourseService_CreateCourse_SinMateriales, ..._CreateCourseAsTeacher_SinMateriales, validación<br>de id vacío. Front: 4 specs nuevos en admin-cursos (crear sin materiales y publicar; borrador si falla publicar;<br>subir material desde el curso; material desde la pestaña no se adjunta). Spec viejo 'sin materiales es inválido'<br>actualizado al requisito.|
|Resultado|education: 15 paquetes OK (90,2 %); front 694/694 OK.|
|Pendientes|Borrar un material que está adjuntado a cursos deja la referencia rota (el curso no se puede republicar hasta<br>quitarlo): falta avisar antes de borrar.|



#### **Problema 2. Pruebas de carga y seguridad en todos los servicios** 

|**Campo**|**Detalle**|
|---|---|
|Descripción|Revisar qué pruebas de rendimiento y seguridad existen y crear y ejecutar las que falten.|
|Comportamiento esperado|Pruebas reales, con herramientas y resultados documentados, sin comprometer datos reales.|
|Comportamiento<br>encontrado|Existían k6 (auth y education) y ZAP solo en el CD de testing; los reportes de tests/security, tests/performance<br>y tests/e2e_a11y se generan con valores fijos (no miden nada). No había pruebas de carga ni de seguridad<br>para campaigns, human-promotion, labor, notifications ni dashboard.|
|Causa|Las pruebas no funcionales se armaron para el pipeline de testing y para el TP de Calidad, no para correr<br>localmente.|
|Archivos afectados|tests/carga/k6_carga_local.js (nuevo); tests/seguridad/sonda/main.go (nuevo);<br>campaigns/internal/routes/carga_test.go (nuevo); api-gateway/nginx.conf; docs/evidencias/*.json|
|Corrección realizada|Se crearon y ejecutaron: (a) carga k6 contra el stack Docker local, solo con rutas sin base o lectura a ritmo<br>bajo; (b) sonda de seguridad de 39 casos (tokens falsificados, roles, CORS, entradas, headers); (c) carga en<br>proceso de Campañas con emulador y 500/2000 registros. nginx ahora responde 429 en lugar de 503 al<br>superar el límite.|
|Tests ejecutados|k6 0.54.0 (Docker), Go net/http, emulador de Firestore.|
|Resultado|k6: 350.639 pedidos a 5.686/s, 0 errores 5xx, p95 12 ms (servicios) y 709 ms (gateway a 8/s). Sonda: 38/39<br>con el stack viejo (la falla era la vulnerabilidad de labor) y 39/39 tras reconstruir el stack y darle JWT_SECRET<br>a labor (12/11). Campañas con 50 clientes: 500 registros p95 485 ms; 2000 registros p95 8,8 s.|
|Pendientes|Paginación del listado de campañas; límite por IP que se puede esquivar con X-Forwarded-For; límite de<br>100/min por IP puede afectar a un centro con una sola IP pública; correr ZAP local.|



#### **Problema 3. Promoción Humana: notificaciones** 

|**Campo**|**Detalle**|
|---|---|
|Descripción|Problemas en las notificaciones del módulo de Promoción Humana.|
|Comportamiento esperado|Avisos al crear y editar fichas y al registrar o editar hitos, a los destinatarios correctos, sin duplicados, con<br>enlace válido y visibles en la interfaz.|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 33 

|**Campo**|**Detalle**|
|---|---|
|Comportamiento<br>encontrado|(1) En local el emulador de Pub/Sub no tenía ni un topic ni una suscripción: ningún evento llegaba. (2) El<br>contenedor de notifications corre con una configuración vieja, sin AUTH_GRPC_URL. (3) Todas las<br>notificaciones de PH llevaban a /dashboard/acompanamiento, ruta inexistente (y había otros enlaces rotos). (4)<br>ficha.actualizada no llevaba creado_por, así que el aviso al creador no salía nunca. (5) Un reintento de<br>Pub/Sub guardaba una notificación huérfana. (6) La interfaz no mostraba no leídas ni se refrescaba sola. (7)<br>Campañas avisaba dos veces. (8) En producción, sin configuración de push, notifications aceptaba eventos sin<br>verificar.|
|Causa|Emulador en memoria sin reinicialización; contenedores creados antes de cambios del compose; enlaces que<br>no coinciden con app.routes.ts; payload incompleto en el productor de PH.|
|Archivos afectados|docker-compose.yml; human-promotion/internal/{events/producers,repositories}/...;<br>notifications/internal/{rules,repositories}, notifications/cmd/server/main.go;<br>campaigns/internal/services/campaign_service.go; frontend: notifications.service, notificaciones.component,<br>header.component|
|Corrección realizada|Emulador que recrea topics y suscripciones al arrancar (y healthcheck que lo espera); se reinicializó el<br>emulador actual. creado_por en ficha.actualizada. Enlaces a /dashboard/promocion-humana/participantes/{id}<br>y demás rutas reales. Sin notificación huérfana si todos los destinatarios ya la tenían. Contador de no leídas en<br>la campana (cada minuto y al navegar) y refresco de la pantalla cada minuto. Campaña avisa solo al publicar.<br>En producción notifications no arranca sin verificar los push.|
|Tests ejecutados|TestFichaActualizada_LlevaCreadoPor (emulador); TestPromocionHumana_EnlacesALaFichaDelParticipante;<br>TestTodasLasReglas_EnlacesExistenEnElFront; TestCreateNotificationAndRecipients_ReintentoNoDuplica;<br>TestCampanaCreada_SoloAlPublicarYUnaVez; specs de notificaciones (contador, refresco, falla del contador<br>no deshace la acción).|
|Resultado|notifications 14 paquetes OK; human-promotion 11 OK; front 694/694 OK.|
|Pendientes|Probar el flujo completo en el navegador (el stack ya se reconstruyó el 12/11); verificar en GCP las<br>suscripciones push y agregar PUBSUB_PUSH_AUDIENCE / PUBSUB_PUSH_SERVICE_ACCOUNTS al CD.|



#### **Problema 4. Permisos y acciones según el rol** 

|**Campo**|**Detalle**|
|---|---|
|Descripción|Auditar qué ve y qué puede hacer cada rol, en front y back.|
|Comportamiento esperado|Que el backend valide siempre, más allá de lo que muestre la interfaz, y que front y back coincidan con HU-06.|
|Comportamiento<br>encontrado|(1) CRÍTICO: labor-integration tomaba el rol de los headers X-User-Role / X-User-ID del cliente si no había<br>token válido; sin sesión se podía aprobar, rechazar o borrar ofertas o ver postulantes (confirmado en vivo<br>contra el puerto 8310). (2) Editar y borrar ofertas no controlaban rol ni dueño, y la empresa podía<br>autoaprobarse. (3) Las ofertas pendientes las veía cualquier rol que no fuera participante. (4) Participante,<br>empresa, docente o voluntario podían crear un perfil a nombre de otra cuenta. (5) El participante podía borrar<br>su perfil de participante (HU-09). (6) Seguimiento roto para admin y voluntario. (7) El puerto de labor estaba<br>abierto a la red local. Además: rutas del dashboard sin guard (usuarios, seguimiento, cv, ofertas: el backend<br>responde igual); el docente puede crear cursos borrador; no hay panel RBAC.|
|Causa|Servicios escritos por distintas personas con criterios distintos; labor dependía de que el gateway nginx pisara<br>los headers, cosa que no pasa con el API Gateway de GCP ni entrando directo al puerto.|
|Archivos afectados|labor-integration/internal/{routes,services,controllers}, cmd/server/main.go;<br>auth/internal/services/{approval_helper,participant,company,teacher,volunteer}_service.go; education<br>(progress_controller, routes); frontend seguimiento; docker-compose.yml; api-gateway/openapi.yaml|
|Corrección realizada|Labor: con JWT_SECRET se descartan siempre los headers del cliente y la identidad sale del JWT; en<br>producción no arranca sin JWT_SECRET; puerto solo en 127.0.0.1. Ofertas: crear (empresa/gestión), editar y<br>borrar (gestión o empresa dueña), estado solo por admin (renovar vencida permitido), pendientes solo para<br>gestión y empresa dueña. Auth: checkCreateProfileOwnership en los cuatro perfiles; el participante no da de<br>baja perfiles. Education: GET /api/cursos/{id}/progreso y Seguimiento por rol.|
|Tests ejecutados|TestIdentidad_NoSeConfiaEnHeadersDelCliente (4 casos); TestUpdateOffer_PermisosYEstado;<br>TestDeleteOffer_Permisos; TestCreateOffer_SoloEmpresaYGestion;<br>TestListOffers_PendientesSoloParaGestion; TestParticipantService_PermisosDeAltaYBaja; progreso<br>(controller + 3 casos RBAC); seguimiento.spec (+3); sonda de seguridad (22 casos por rol, todos OK).|
|Resultado|auth 11, labor 7, education 15 paquetes OK; sonda 39/39 contra el stack reconstruido (12/11); el uso legítimo<br>de labor con token válido sigue respondiendo 200.|
|Pendientes|Panel RBAC; auditoría de accesos denegados; guards en usuarios/cv/ofertas/seguimiento; decisión sobre<br>cursos creados por el docente y ofertas creadas por el voluntario; volunteers y teachers internos pueden listar<br>perfiles de su mismo tipo.|



#### **Problema 5. Campañas: eliminación** 

|**Campo**|**Detalle**|
|---|---|
|Descripción|No se podían eliminar campañas.|
|Comportamiento esperado|Eliminar con permiso, con confirmación, persistiendo en la base, actualizando la lista sin recargar, con<br>mensajes claros y sin borrar datos relacionados.|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 34 

|**Campo**|**Detalle**|
|---|---|
|Comportamiento<br>encontrado|La función no existía: no había ruta DELETE en campaigns-service, ni método en el servicio del front, ni botón.|
|Causa|No estaba en las HU (47 a 51 solo piden crear, editar y publicar), así que no se implementó.|
|Archivos afectados|campaigns/internal/{models,repositories,services,controllers,routes}; frontend campaigns.service.ts,<br>campanas.component.{ts,html}; api-gateway/openapi.yaml|
|Corrección realizada|Baja lógica (decisión de la usuaria): eliminada, eliminada_en y eliminada_por en el documento; no se lista en la<br>gestión ni en la landing, no se puede editar ni publicar, y un segundo DELETE da 404. La pueden usar admin y<br>voluntario (los mismos roles que gestionan campañas). Los comentarios de la campaña (education) se<br>conservan. Front: botón Eliminar con confirmación, quita la tarjeta sin recargar y muestra 'Campaña eliminada<br>correctamente'; ante un error lo muestra y recarga.|
|Tests ejecutados|TestDelete_SacaLaCampanaDeGestionYPublico, TestDelete_Validaciones,<br>TestCampaignRepository_BajaLogica (emulador), TestEliminarCampana_PermisosYEfecto<br>(401/403/200/404), campanas.spec (+4), campaigns.service.spec (+1).|
|Resultado|campaigns 8 paquetes OK (90,2 %); front 694/694 OK. Contra el stack reconstruido: participante -> DELETE<br>da 403 (antes 405 porque la ruta no existía).|
|Pendientes|Desplegar a testing el openapi.yaml con el DELETE (en local ya funciona).|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 35 

### **9. Inventario de tests existentes** 

|**Módulo**<br>|**Tipo**<br>|**Ubicación**<br>|**Herramienta**|**Arch**|**.**<br>**Tests**<br>|**Resultado**<br>|**Cobertura**<br>|**Obs.**|
|---|---|---|---|---|---|---|---|---|
|pkg (librería de<br>eventos)|Unitarios /<br>integración con<br>emulador|pkg/events,<br>pkg/outbox|go test +<br>emuladores|5|32|2/2 paquetes<br>OK|94,0 %||
|auth|Unitarios,<br>controladores,<br>contrato, base de<br>datos|auth/internal/**, auth/t<br>ests/{contract,databa<br>se}|go test,<br>testify,<br>sqlmock|69|527|11/11<br>paquetes OK|83,6 %|+1 test de<br>permisos<br>(auditoría)|
|education|Unitarios,<br>rutas/RBAC,<br>repositorios<br>contra emulador,<br>contrato,<br>integración|education/internal/**,<br>education/tests/*|go test,<br>testify,<br>httptest,<br>emulador<br>Firestore|93|628|15/15<br>paquetes OK|90,2 %|Reescribe report<br>s/database y<br>reports/pact al<br>correr|
|human-promotio<br>n|Unitarios,<br>repositorios con<br>emulador,<br>eventos|human-promotion/int<br>ernal/**|go test +<br>emulador|29|101|11/11<br>paquetes OK|75,8 %|+1 test<br>(auditoría)|
|dashboard|Unitarios,<br>agregación, caché|dashboard/internal/**|go test|14|92|9/9 paquetes<br>OK|85,8 %||
|campaigns|Unitarios,<br>rutas/RBAC,<br>repositorio con<br>emulador, carga<br>en proceso|campaigns/internal/**|go test +<br>emulador|15|70|8/8 paquetes<br>OK|90,2 %|+6 tests<br>(auditoría)|
|notifications|Reglas,<br>consumidores,<br>delivery,<br>repositorio en<br>memoria|notifications/internal/*<br>*|go test|21|98|14/14<br>paquetes OK|63,2 %|+3 tests<br>(auditoría)|
|labor-integration|Unitarios,<br>controladores,<br>identidad|labor-integration/inter<br>nal/**|go test|15|50|7/7 paquetes<br>OK|59,3 %|+8 tests<br>(auditoría)|
|rxresume|Unitarios|rxresume/internal/**|go test|2|7|2/2 paquetes<br>OK|38,4 %|Cobertura baja|
|frontend|Componentes,<br>servicios, guards<br>(Karma/Jasmine)|frontend/src/**/*.spec<br>.ts|Angular CLI +<br>Karma + Chro<br>meHeadless|85|694|694/694 OK|71,45 %<br>sentencias<br>/ 58,32 %<br>ramas|+15 specs<br>(auditoría)|
|frontend E2E|End-to-end|frontend/cypress/e2e/<br>auth.cy.ts|Cypress|1|1|No ejecutable|-|Cypress no está<br>en package.json<br>y falta cypress/s<br>upport/e2e.ts|
|Carga (CD)|Carga|education/load_test.j<br>s, tests/performance/<br>k6_auth_load_test.js|k6|2|-|No ejecutados<br>(solo CD de<br>testing)|-||
|Carga local<br>(nuevo)|Carga|tests/carga/k6_carga<br>_local.js|k6 0.54.0 en<br>Docker|1|3|Ejecutado: 0<br>% de 5xx|-|Ver sección de<br>resultados|
|Carga en<br>proceso (nuevo)|Carga / volumen|campaigns/internal/ro<br>utes/carga_test.go|go test +<br>emulador<br>(CARGA=1)|1|1|Ejecutado: 0<br>errores|-|500 y 2000<br>campañas|
|Seguridad<br>DAST (CD)|Seguridad|zaproxy/action-api-sc<br>an en pipeline.yml|OWASP ZAP|-|-|No ejecutado<br>(solo CD)|-||
|Sonda de<br>seguridad<br>(nuevo)|Seguridad / roles|tests/seguridad/sond<br>a/main.go|Go net/http|1|39|Ejecutado:<br>39/39 OK|-|38/39 antes de<br>reconstruir el<br>stack|
|Reportes con<br>valores fijos|No son pruebas|tests/{security,perfor<br>mance,e2e_a11y}/ge<br>nerate_*.go -><br>reports/|Go|3|-|No miden<br>nada|-|No se tocaron<br>por pedido<br>previo de la<br>usuaria|
|Mutación|Mutación|stryker.conf.json, run<br>_mutation_tests.sh|Stryker /<br>go-mutesting|-|-|No ejecutado<br>(if: false en CI)|-||



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 36 

- Existentes: todos los de la tabla. Ejecutados: las 9 suites Go, Karma, la carga k6 local, la carga en proceso y la sonda (además, los nuevos). 

- Pasan: 79/79 paquetes Go y 694/694 specs. Fallan: ninguno. No pudieron ejecutarse: Cypress (no instalado), k6 y ZAP del CD (requieren GCP), mutación (deshabilitada). 

- Advertencia: correr los tests de education reescribe reports/database y reports/pact; se restauraron para no pisar cambios previos. 

### **10. Tests faltantes** 

|**Grupo**|**Qué probar**|**HU / origen**|**Módulo**|**Prioridad**|
|---|---|---|---|---|
|Backend|Bloqueo tras 5 intentos de login (cuando se implemente)|HU-02|auth|Alta|
|Backend|Auditoría de altas/bajas/modificaciones de perfiles|HU-07 a HU-11|auth|Alta|
|Backend|Reutilización de refresh token invalida toda la sesión|HU-04|auth|Alta|
|Backend|Campos obligatorios de oferta laboral y de participante|HU-07, HU-39|auth, labor|Media|
|Backend|Subir la cobertura de rxresume (38 %) y labor (59 %)|HU-43, HU-59|rxresume, labor|Media|
|Frontend|Specs de la ficha de Promoción Humana y de mi-proceso|HU-29 a HU-34|frontend|Media|
|Frontend|Specs de ofertas por rol (botones Editar/Eliminar/Aprobar)|HU-39 a HU-42|frontend|Media|
|Integración|Pub/Sub de punta a punta: ficha/hito -> notificación en la<br>campana|HU-60, Problema 3|PH, notifications|Alta|
|Integración|Inscripción -> email (Resend en modo prueba)|HU-18, HU-26|education,<br>notifications|Alta|
|End-to-end|Instalar Cypress (o Playwright) y cubrir login por rol, alta de<br>curso sin materiales, eliminar campaña, ficha|HU-02, HU-14,<br>Problema 5|frontend + stack|Alta|
|Seguridad|ZAP local contra el gateway y DAST en cada servicio|Problema 2|todos|Media|
|Seguridad|Rate limit no esquivable por X-Forwarded-For|Problema 2|education, PH,<br>campaigns|Media|
|Rendimiento|Listado paginado de campañas y biblioteca con 2000+<br>registros|HU-50|campaigns,<br>education|Media|
|Carga|Carga con tokens reales en un entorno aislado (no<br>compartido con testing)|Problema 2|todos|Baja|
|Roles y<br>permisos|E2E de acceso directo por URL a rutas sin guard (usuarios,<br>cv, ofertas, seguimiento)|HU-06|frontend|Media|
|Validaciones|Dominio gmail en registro (si se confirma el criterio)|HU-01|auth, frontend|Baja|
|Casos de error|Borrar material adjuntado a cursos (referencia rota)|HU-14|education|Media|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 37 

### **11. Resultados de testing y seguridad** 

#### **Carga con k6 (stack Docker local)** 

Herramienta: k6 0.54.0 en Docker (tests/carga/k6_carga_local.js). Configuración: escenario **gateway** a 8 pedidos/s (debajo del límite de nginx), escenario **servicios** directo a education y human-promotion con rampa a 50 usuarios virtuales (solo rutas sin base: salud y rechazos 401) y escenario **público** a 3/s contra la biblioteca (lectura real de Firestore). 

|**Métrica**|**Valor**|
|---|---|
|Pedidos totales|350639|
|Pedidos por segundo|5685.6|
|Errores 5xx|0 de 350639|
|p95 servicios (ms)|12.3|
|p95 gateway a 8/s (ms)|709.3|
|p95 público biblioteca (ms)|11.3|
|Usuarios virtuales máximos|60|



- Primera corrida (sin respetar el límite): 68 % de respuestas 503 en 0 ms. No era una caída: era el límite de 10/s por IP de nginx. Se cambió a 429. 

- Con 50 usuarios directos, education y PH respondieron 429 al superar 100 pedidos/min por IP, sin errores 5xx ni caídas. 

- El límite por IP se esquivó rotando X-Forwarded-For: 120/120 pedidos pasaron (con la misma IP: 77 OK y 43 rechazados). 

#### **Carga en proceso con volumen de datos (Campañas)** 

|**Campañas cargadas**|**Clientes**|**Pedidos**|**p50**|**p95**|**Máximo**|**Errores**|
|---|---|---|---|---|---|---|
|500|50|1000|134 ms|485 ms|1,02 s|0|
|2000|50|1000|3,3 s|8,8 s|9,6 s|0|



Servicio real con emulador de Firestore (los tiempos absolutos son pesimistas). Conclusión: sin paginación, el listado no escala por encima de unos miles de campañas. 

#### **Sonda de seguridad** 

Herramienta: programa Go propio (tests/seguridad/sonda). Casos: 39; OK: 39; fallas: 0. Solo pedidos que no modifican datos. 

|**Grupo**|**Caso**|**Esperado**|**Obtenido**|**Resultado**|
|---|---|---|---|---|
|autenticacion|education sin token|401|401|OK|
|autenticacion|education token alg=none (admin)|401|401|OK|
|autenticacion|education token firmado con otra clave|401|401|OK|
|autenticacion|education token basura|401|401|OK|
|autenticacion|human-promotion sin token|401|401|OK|
|autenticacion|human-promotion token alg=none|401|401|OK|
|autenticacion|gateway auth /users sin token|401|401|OK|
|autenticacion|gateway auth /users token alg=none|401|401|OK|
|autenticacion|gateway campañas gestión sin token|401|401|OK|
|autenticacion|gateway notificaciones sin token|401|401|OK|
|autenticacion|gateway dashboard sin token|401|401|OK|
|labor-headers|labor directo: aprobar oferta inexistente sin token y X-User-Role admin<br>falsificado|401,403|403|OK|
|labor-headers|gateway: aprobar oferta inexistente sin token y X-User-Role admin<br>falsificado|401,403|401|OK|
|cors|education preflight desde origen no permitido|200,204,403|200|OK|
|entradas|campañas públicas con búsqueda de inyección|200,400|200|OK|
|entradas|campañas públicas con búsqueda de 5000 caracteres|400,414|400|OK|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 38 

|**Grupo**|**Caso**|**Esperado**|**Obtenido**|**Resultado**|
|---|---|---|---|---|
|entradas|comentarios: listar con entidad inválida|400|400|OK|
|roles|education token vencido|401|401|OK|
|roles|education token de refresh usado como acceso|401|401|OK|
|roles|participante -> borradores de cursos (admin)|403|403|OK|
|roles|participante -> crear curso|403|403|OK|
|roles|voluntario -> crear curso (solo admin)|403|403|OK|
|roles|participante -> auditoría de educación|403|403|OK|
|roles|voluntario -> auditoría de educación (solo admin)|403|403|OK|
|roles|participante -> asistencia general|403|403|OK|
|roles|empresa -> buscador de participantes|403|403|OK|
|roles|docente externo -> /api/docente|403|403|OK|
|roles|docente -> rutas de participante|403|403|OK|
|roles|participante -> borrar comentario (moderación admin)|403|403|OK|
|roles|participante -> fichas de promoción humana|403|403|OK|
|roles|docente -> fichas de promoción humana|403|403|OK|
|roles|empresa -> fichas de promoción humana|403|403|OK|
|roles|voluntario -> indicadores (solo admin)|403|403|OK|
|roles|voluntario -> auditoría de PH (solo admin)|403|403|OK|
|roles|empresa -> mi línea de tiempo (solo participante)|403|403|OK|
|roles|participante -> crear campaña (gateway)|403|403|OK|
|roles|empresa -> gestión de campañas (gateway)|403|403|OK|
|roles|participante -> eliminar campaña (gateway)|403,405|403|OK|
|roles|voluntario -> dashboard (solo admin)|401,403|403|OK|



**Limitaciones.** El stack local comparte bases con testing, por lo que no hubo pruebas de carga con escritura ni con tokens de usuarios reales. La carga se midió con el stack del 29/10; la sonda se repitió el 12/11 contra el stack reconstruido (39/39). ZAP no se ejecutó. 

Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 39 

### **12. Pendientes de implementación** 

**Actualización 15/11/2026.** La tabla de abajo es la original del 11/11. El estado actual de cada pendiente está en la sección 16 y lo que falta hacer en testing, en el paso a paso del final. 

|**Descripción**|**HU**|**Módulo**|**Archivos**|**Trabajo necesario**|**Dependenc**<br>**ias**|**Prior.**|**Criterio de aceptación**|
|---|---|---|---|---|---|---|---|
|Probar en el<br>navegador el flujo<br>de notificaciones<br>de PH|Problema<br>3|PH / Notific<br>aciones|-|Con el stack ya<br>reconstruido (12/11):<br>crear una ficha y un<br>hito visible y ver la<br>notificación en la<br>campana|-|Alta|Una ficha nueva genera<br>notificación en la campana<br>del admin y el hito visible le<br>llega al participante.|
|Desplegar a<br>testing los<br>arreglos de<br>seguridad de<br>labor|HU-06,<br>Problema<br>4|Inserción<br>Laboral|labor-integrati<br>on/*|Commit + pase a<br>testing|OK del<br>equipo<br>(código de<br>otro<br>integrante)|Alta|PUT /offers/x/approve sin<br>token con X-User-Role:<br>role_admin responde<br>401/403 en testing.|
|Bloqueo de<br>cuenta tras 5<br>intentos fallidos|HU-02|Autenticació<br>n|auth/internal/s<br>ervices/auth_s<br>ervice.go|Contador por cuenta +<br>bloqueo temporal +<br>mensaje|-|Alta|Al 6.º intento la cuenta<br>queda bloqueada N minutos<br>y se desbloquea sola.|
|Log de auditoría<br>en auth|HU-03,<br>HU-06,<br>HU-07 a<br>HU-11|Autenticació<br>n|auth (nuevo<br>repositorio de<br>auditoría)|Registrar en la misma<br>transacción usuario,<br>acción, recurso y<br>fecha; endpoint de<br>lectura para admin;<br>sumar a<br>/dashboard/auditoria|-|Alta|Cada alta/baja/edición de<br>perfil aparece en la pantalla<br>de auditoría y no se puede<br>editar.|
|Refresh<br>reutilizado<br>invalida la sesión|HU-04|Autenticació<br>n|auth_service_r<br>efresh.go|Si el token está<br>revocado,<br>RevokeAllForUser|-|Alta|Reusar un refresh devuelve<br>401 y el siguiente refresh<br>legítimo también falla.|
|Verificar Pub/Sub<br>y emails en<br>testing|HU-18,<br>HU-23,<br>HU-26,<br>HU-60|Notificacion<br>es|infra/pubsub/s<br>etup.sh,<br>pipeline.yml|Correr setup.sh con N<br>OTIFICATIONS_URL,<br>agregar<br>PUBSUB_PUSH_* al<br>CD, Cloud Scheduler<br>para recordatorios|Acceso a<br>GCP|Alta|Una inscripción en testing<br>genera el email y la<br>notificación in-app.|
|Notificación de<br>oferta compatible|HU-44|Inserción<br>Laboral / No<br>tificaciones|notifications/int<br>ernal/rules/lab<br>or.go|Definir compatibilidad<br>(habilidades) y regla<br>para participantes|Definición<br>funcional|Media|Al aprobarse una oferta, los<br>participantes con<br>habilidades coincidentes<br>reciben aviso.|
|Panel de<br>configuración<br>RBAC|HU-06|Autenticació<br>n / Frontend|frontend<br>(nueva<br>pantalla), auth<br>/roles<br>/permissions|Pantalla de solo admin<br>para ver y editar<br>permisos por rol|-|Media|El admin ve la matriz de<br>permisos y un cambio se<br>refleja en el backend.|
|Campos<br>obligatorios de<br>ofertas y<br>participantes|HU-07,<br>HU-39|Auth, Labor|dto de auth y<br>labor,<br>formularios|validate:required en los<br>campos de la HU|-|Media|Crear sin uno de esos<br>campos responde 400 con<br>el campo faltante.|
|Notificar rechazos<br>de cuentas|HU-12,<br>HU-38|Auth / Notifi<br>caciones|user_service.g<br>o, templates|Email<br>'cuenta_rechazada'|-|Media|Al rechazar, el usuario<br>recibe el email.|
|Push en testing|HU-61|Notificacion<br>es|pipeline.yml, p<br>ush_service.g<br>o|Credenciales de FCM<br>y FIREBASE_PROJEC<br>T_ID|Acceso a<br>Firebase|Media|Con la PWA instalada llega<br>un push al registrar un hito<br>visible.|
|Paginación de<br>listados públicos|HU-50|Campañas,<br>Educación|campaign_rep<br>ository.go,<br>frontend|Limit + cursor;<br>búsqueda con índice|-|Media|Con 2000 campañas el p95<br>del listado público queda<br>por debajo de 500 ms.|
|Rate limit robusto|Problema<br>2|education,<br>PH,<br>campaigns|middlewares|Tomar la IP del último<br>salto confiable o limitar<br>por usuario; revisar el<br>cupo de 100/min para<br>una IP compartida|Conocer<br>cómo arma<br>X-Forwarde<br>d-For el API<br>Gateway|Media|Rotar X-Forwarded-For no<br>esquiva el límite y 15<br>usuarios de una misma IP<br>usan la app sin 429.|
|Guards de<br>frontend faltantes|HU-06|Frontend|app.routes.ts|Guards en usuarios,<br>cv, ofertas y<br>seguimiento|-|Baja|Entrar por URL con un rol<br>no permitido redirige al<br>inicio.|
|Observabilidad|HU-56|Infra|GCP Cloud<br>Monitoring|Dashboard por servicio<br>(errores, latencia,<br>recursos) y alertas|Acceso a<br>GCP|Media|URL privada con el panel y<br>una alerta que se dispara<br>ante caída.|



Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 40 

|**Descripción**|**HU**|**Módulo**|**Archivos**|**Trabajo necesario**|**Dependenc**<br>**ias**|**Prior.**|**Criterio de aceptación**|
|---|---|---|---|---|---|---|---|
|Validación de<br>dominio gmail|HU-01|Autenticació<br>n|dto/auth_dto.g<br>o, register|Confirmar criterio y<br>validar|Definición<br>funcional|Baja|Un email que no es gmail es<br>rechazado con mensaje<br>claro.|
|Aviso al borrar<br>material<br>adjuntado|HU-14|Educación|material_servi<br>ce.go,<br>admin-cursos|Contar cursos que lo<br>usan y avisar o impedir|-|Media|Borrar un material usado<br>muestra en qué cursos está.|
|Instalar y correr<br>E2E|Varias|Frontend|frontend/cypre<br>ss|Agregar Cypress al<br>package.json, support<br>file y specs por rol|-|Alta|npx cypress run pasa en<br>local con el stack Docker.|



#### **Historias no completas** 

|**HU**|**Estado**|**Qué falta**|
|---|---|---|
|HU-01|Parcialmente<br>implementado|Agregar la validación de dominio gmail en el DTO de registro y en el formulario (o confirmar con el<br>equipo si el criterio sigue vigente).|
|HU-02|Parcialmente<br>implementado|Contador de intentos fallidos por cuenta con bloqueo temporal (por ejemplo 15 minutos) y su test.|
|HU-03|Parcialmente<br>implementado|Registrar los rechazos de token en una auditoría central (ver HU-10).|
|HU-04|Parcialmente<br>implementado|Ante reutilización, revocar todos los refresh tokens del usuario (RevokeAllForUser ya existe).|
|HU-06|Parcialmente<br>implementado|Panel RBAC en el front; auditoría de accesos denegados; decidir si el docente puede proponer<br>cursos.|
|HU-07|Parcialmente<br>implementado|Hacer obligatorios los campos de la HU en el DTO; auditoría de altas.|
|HU-08|Parcialmente<br>implementado|Auditoría de modificaciones en auth.|
|HU-09|Parcialmente<br>implementado|Auditoría de bajas.|
|HU-10|No implementado|Tabla de auditoría en auth (usuario, acción, recurso, fecha) en la misma transacción, y sumarla a la<br>pantalla de auditoría.|
|HU-11|Parcialmente<br>implementado|Auditoría; verificar límite de 5 MB para la foto de perfil.|
|HU-12|Parcialmente<br>implementado|Email y aviso de rechazo.|
|HU-18|Pendiente de<br>verificación|Verificar en testing que llegue el email (eventos + suscripción push de notifications).|
|HU-23|Pendiente de<br>verificación|Verificar la notificación en testing.|
|HU-26|Pendiente de<br>verificación|Configurar el job de Scheduler y verificar envíos.|
|HU-31|Parcialmente<br>implementado|Confirmar con el equipo y, si corresponde, actualizar la HU.|
|HU-38|Parcialmente<br>implementado|Notificar el rechazo.|
|HU-39|Parcialmente<br>implementado|Validar todos los campos en el DTO y en el formulario.|
|HU-41|Parcialmente<br>implementado|Notificar a la empresa al eliminar.|
|HU-42|Parcialmente<br>implementado|Auditoría de eliminaciones.|
|HU-44|No implementado|Definir 'compatible' (habilidades / CV) y la regla de notificación.|
|HU-56|Parcialmente<br>implementado|Dashboard de Cloud Monitoring (errores, latencia, CPU/memoria por servicio) con alertas, o<br>Grafana.|
|HU-60|Pendiente de<br>verificación|Verificar suscripciones push en GCP.|
|HU-61|Parcialmente<br>implementado|Configurar credenciales de FCM en testing y probar en un dispositivo.|



### **13. Recomendaciones técnicas** 

Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 41 

##### Son mejoras sugeridas, no requerimientos de las HU (esos están en Pendientes). 

|**Área**|**Recomendación**|
|---|---|
|Seguridad|No confiar nunca en headers de identidad: validar el JWT en cada servicio (labor ya lo hace con JWT_SECRET).<br>Agregar JWT_SECRET al .env local de labor.|
|Seguridad|Agregar PUBSUB_PUSH_AUDIENCE y PUBSUB_PUSH_SERVICE_ACCOUNTS al CD de cada consumidor y<br>hacer que fallen cerrado también en testing.|
|Seguridad|Los .env guardan JWT_SECRET entre comillas simples (Docker Compose las quita): al copiar la clave a otro<br>servicio o a GCP, copiar la línea completa con el mismo formato.|
|Rendimiento|Paginación y límites en listados (campañas, biblioteca, ofertas); hoy se lee toda la colección en cada pedido.|
|Rendimiento|Revisar el límite de 10/s (nginx) y 100/min (servicios) por IP: un centro barrial con una sola IP pública puede<br>quedarse corto en uso normal.|
|Arquitectura|Unificar el modelo de autorización (RequireRole + dueño) en una librería compartida como se hizo con pkg/events,<br>para que labor y auth no diverjan.|
|Arquitectura|Auditoría común (mismo modelo que education y PH) para auth y labor.|
|Mantenibilidad|Reemplazar los reportes con valores fijos de tests/{security,performance,e2e_a11y} por los resultados reales (k6,<br>sonda, ZAP).|
|Testing|Instalar Cypress o Playwright y correr E2E por rol en el CI con el stack Docker.|
|Testing|Subir cobertura de rxresume, labor y notifications; specs de la ficha de PH.|
|Organización|Recrear el stack Docker después de cada cambio de docker-compose.yml o .env (al momento de la auditoría eran<br>del 29/10 y notifications no tenía AUTH_GRPC_URL).|
|Documentación|Mantener docs/HISTORIAS_DE_USUARIO.md y este informe desde scripts/informe-auditoria/datos.py.|
|Experiencia de usuario|Avisar en el front cuando el backend responde 429 (hoy aparece como error genérico).|



### **14. Conclusión** 

Buen Vivir App cubre la mayor parte de lo pedido: Educación, Promoción Humana, Campañas y Dashboard están prácticamente completos y con buena cobertura de tests. La auditoría encontró y corrigió problemas de seguridad serios en Inserción Laboral y en Autenticación, resolvió los cinco problemas reportados y dejó pruebas de carga y de seguridad reutilizables. 

Para considerar el proyecto terminado falta: (1) reconstruir el stack Docker y desplegar los arreglos; (2) completar Autenticación (auditoría, bloqueo por intentos, invalidación de sesión); (3) verificar de punta a punta notificaciones, emails y push en testing; (4) las HU de Inserción Laboral pendientes (validaciones, ofertas compatibles); (5) panel RBAC y observabilidad; (6) E2E por rol. 

**Actualización 15/11/2026.** Los puntos (2), (4) y (5) quedaron implementados con tests (ver sección 16); falta desplegarlos y verificarlos en testing. Siguen abiertos el (1) (despliegue), el (3) (necesita acceso a GCP y Firebase) y el (6) (necesita una base de datos de pruebas aislada). 

Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 42 

### **15. Anexo técnico** 

#### **Estructura del repositorio** 

|**Carpeta**|**Contenido**|
|---|---|
|auth/, education/, human-promotion/,<br>campaigns/, dashboard/, notifications/,<br>labor-integration/, rxresume/|Servicios Go (cmd/server,<br>internal/{controllers,services,repositories,models,routes,middlewares,events,outbox})|
|pkg/|Librería fuente de eventos y outbox (copiada a cada servicio con scripts/sync-eventlib.sh)|
|frontend/|PWA Angular|
|api-gateway/|nginx.conf (local) y openapi.yaml (GCP)|
|infra/pubsub/|setup.sh (GCP) y emulator.sh (local)|
|cloudflare-worker/|Worker de métricas de impacto (KV)|
|docs/|Documentación (este informe, HU, arquitectura, ADR, eventos, evidencias, referencias)|
|tests/|Carga (k6), sonda de seguridad, generadores de reportes del TP de Calidad|
|scripts/|sync-eventlib, informe-auditoria, historico-openapi, calidad|
|.github/workflows/pipeline.yml|CI/CD: tests, imágenes, Cloud Run, smoke, ZAP, k6, rollback|



#### **Comandos de ejecución** 

|**Acción**|**Comando**|
|---|---|
|Levantar todo en local|docker compose up -d --build|
|Tests Go (con emuladores)|FIRESTORE_EMULATOR_HOST=127.0.0.1:8390 PUBSUB_EMULATOR_HOST=localhost:8086 go<br>test ./... (en cada servicio)|
|Tests frontend|cd frontend && npx ng test --watch=false --browsers=ChromeHeadless --code-coverage|
|Carga local|docker run --rm -v "$PWD/tests/carga:/scripts" grafana/k6:0.54.0 run /scripts/k6_carga_local.js|
|Sonda de seguridad|JWT_SECRET=... go run ./tests/seguridad/sonda > docs/evidencias/sonda-seguridad-resultado.json|
|Carga Campañas en proceso|cd campaigns && CARGA=1 FIRESTORE_EMULATOR_HOST=... go test ./internal/routes -run Carga<br>-v|
|Regenerar este informe|python scripts/informe-auditoria/generar.py|



#### **Endpoints principales por servicio** 

|**Servicio**|**Prefijo**|**Rutas**|
|---|---|---|
|auth|/api/v1/auth, /users, /roles, /permissions, /companies,<br>/teachers, /participants, /volunteers, /profile, /directory|login, register, refresh, 2FA, recuperación, CRUD de<br>perfiles, directorio|
|education|/api/cursos, /areas-formativas, /materiales, /docente,<br>/participante, /actividades, /asistencia-general, /habilidades,<br>/comentarios, /biblioteca, /contenidos-docentes, /auditoria|ver education/CLAUDE.md|
|human-promotion|/api/v1/human-promotion|fichas, hitos, actividad, pdf, participantes, indicadores,<br>auditoria, mi-linea-de-tiempo|
|campaigns|/api/v1/campaigns|publicas, CRUD, publicar, DELETE (baja lógica)|
|labor-integration|/api/v1/labor-integration|offers (CRUD, approve, reject, applications), applications<br>(me, CRUD, status)|
|dashboard|/api/v1/dashboard|resumen, reportes|
|notifications|/api/v1/notifications|listado, detalle, read, read-all, unread-count,<br>preferencias; /eventos/* (Pub/Sub)|
|rxresume|/api/v1/rxresume|CV y versión pública|



#### **Archivos importantes** 

- docs/HISTORIAS_DE_USUARIO.md, docs/PENDIENTES.md, docs/ROLES_Y_PERMISOS.md, docs/TESTING.md 

- docs/evidencias/k6-resultado-carga.json y sonda-seguridad-resultado.json 

- scripts/informe-auditoria/datos.py (fuente de este informe) 

Buen Vivir App · Informe de auditoría técnica y funcional · 11/11/2026 

Página 43

### **16. Actualización del 15/11/2026: implementación de los pendientes**

Trabajo hecho sobre `develop` (base `d42cdfd`), **sin commitear todavía**. Se modificó código de `auth/`, `labor-integration/` y `notifications/` con el OK de la usuaria (15/11/2026): **hay que avisarle a Leonardo** antes de pasarlo a testing.

**Decisiones de la usuaria (15/11/2026).**

- auth/ y labor-integration/: se modifican, dejándolo anotado.
- HU-44, "oferta compatible": al publicarse una oferta se avisa a los participantes que tengan al menos una habilidad (ganada en Educación) con el mismo nombre que un requisito de la oferta, sin distinguir mayúsculas ni acentos.
- HU-01, dominio Gmail: obligatorio **solo para participantes** (registro público y alta desde Usuarios). El resto de los roles puede usar cualquier dominio.
- HU-06, panel RBAC: matriz **de solo lectura**. Los permisos siguen en el código de cada servicio; no se hizo una tabla editable.

**Significado de los estados.** IMPLEMENTADO: corregido y verificado con tests automáticos que pasan. IMPLEMENTADO - TESTING PENDIENTE: el código y sus tests unitarios están, pero falta probarlo desplegado (necesita una migración, Pub/Sub, el API Gateway de GCP o un servicio externo). PENDIENTE: no implementado. BLOQUEADO: necesita una decisión, un acceso o una intervención externa.

#### **Estado de cada historia observada**

|**HU**|**Estado**|**Qué se hizo (15/11)**|**Qué falta**|
|---|---|---|---|
|HU-01|IMPLEMENTADO|Gmail obligatorio para participantes en `AuthService.Register`, `UserService.CreateUser`, el formulario de registro y el alta desde Usuarios.|-|
|HU-02|IMPLEMENTADO - TESTING PENDIENTE|Bloqueo por email tras 5 intentos fallidos durante 15 minutos (`LOGIN_MAX_ATTEMPTS`, `LOGIN_LOCK_MINUTES`). Tabla `login_attempts` (guarda el hash del email, no el email). Respuesta 423 con `code: ACCOUNT_LOCKED`; antes del bloqueo avisa cuántos intentos quedan. El login lo muestra y deshabilita el botón. Se desbloquea solo al vencer el plazo; un login correcto reinicia el contador.|Crear la tabla en Neon y probarlo en testing (pasos 4 y 13).|
|HU-03|IMPLEMENTADO - TESTING PENDIENTE|Un token presentado y rechazado (401) en una ruta protegida de auth queda en la auditoría como `token_rechazado`. Los pedidos sin sesión no se registran (para no llenar la tabla).|Migración. Los rechazos en los otros servicios siguen solo en el log técnico.|
|HU-04|IMPLEMENTADO|Reutilizar un refresh token ya usado revoca todos los refresh tokens del usuario y responde 401.|-|
|HU-06|IMPLEMENTADO - TESTING PENDIENTE|Pantalla "Roles y permisos" (`/dashboard/permisos`, solo admin, solo lectura). Los accesos denegados (403) quedan en la auditoría (`acceso_denegado`). Guards por rol en `usuarios`, `cv`, `ofertas` y `seguimiento`.|Migración de la auditoría. El docente sigue pudiendo crear cursos en borrador: se mantiene porque es la moderación de HU-23.|
|HU-07|IMPLEMENTADO - TESTING PENDIENTE|DNI (7 u 8 números), dirección y teléfono obligatorios al crear un participante desde Usuarios y en `POST /participants`; DNI y dirección se guardan en el perfil en la misma transacción que la cuenta. El alta queda en la auditoría.|Probar en testing que el DNI y la dirección quedan guardados.|
|HU-08|IMPLEMENTADO - TESTING PENDIENTE|Cada edición queda en la auditoría con los **nombres** de los campos modificados (nunca los valores).|Migración.|
|HU-09|IMPLEMENTADO - TESTING PENDIENTE|Cada baja queda en la auditoría.|Migración.|
|HU-10|IMPLEMENTADO - TESTING PENDIENTE|Tabla `auditoria` en auth (usuario, rol, acción, recurso, id, campos, resultado, IP, fecha), con un trigger de Postgres que impide editarla o borrarla. La escribe el middleware `AuditTrail`. `GET /api/v1/auth/auditoria` (solo admin) y módulo "Usuarios y perfiles (Autenticación)" en la pantalla Auditoría.|Migración y ruta en el API Gateway. Diferencia con la HU: el registro se escribe al terminar el pedido, no en la misma transacción del cambio (ver observación nueva 7).|
|HU-11|IMPLEMENTADO|Límite de 5 MB para la foto (rxresume, 413) y para los documentos (auth, 400), y el mismo control en el front. Los cambios quedan en la auditoría.|-|
|HU-12|IMPLEMENTADO - TESTING PENDIENTE|Rechazo de cuentas: `PATCH /api/v1/users/{id}/approval` con `rejected` pide el email `cuenta_rechazada`; Notificaciones lo envía. El panel de Usuarios tiene la opción "Rechazado" y aprueba o rechaza por ese endpoint.|Ruta en el API Gateway y verificar el email en testing.|
|HU-14|IMPLEMENTADO - TESTING PENDIENTE|Antes de borrar un material se consulta `GET /api/materiales/{id}/cursos` y la confirmación nombra los cursos que lo usan.|Ruta en el API Gateway.|
|HU-18, HU-23, HU-26, HU-60|BLOQUEADO|Sin cambios de código: dependen de Pub/Sub, Cloud Scheduler y Resend en testing.|Acceso a GCP (pasos 8 a 10).|
|HU-31|BLOQUEADO|Sin cambios.|Decisión del equipo: mantener nombre, edad y habilidades leídos de Auth y Educación (recomendado, no duplica datos personales) o agregarlos como campos de la ficha.|
|HU-38|IMPLEMENTADO - TESTING PENDIENTE|El rechazo de una empresa manda `cuenta_rechazada` (mismo flujo que HU-12).|Verificar el email en testing.|
|HU-39|IMPLEMENTADO|Título, descripción, ubicación, modalidad, tipo de empleo, salario, horario, al menos un requisito y fecha límite son obligatorios al crear y editar (400 con el campo que falta). El formulario los marca y muestra el mensaje del backend.|-|
|HU-41|IMPLEMENTADO - TESTING PENDIENTE|Al borrar una oferta se emite `insercion_laboral.oferta.eliminada`; si la borró la gestión del centro, Notificaciones avisa a la empresa.|Filtros de Pub/Sub (paso 8).|
|HU-42|IMPLEMENTADO - TESTING PENDIENTE|Auditoría de ofertas (crear, editar, aprobar, rechazar, borrar) en Firestore (`auditoria_insercion_laboral`), `GET /api/v1/labor-integration/auditoria` (solo admin) y módulo "Inserción laboral" en la pantalla Auditoría.|Ruta en el API Gateway.|
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
|Problema 4 (permisos de labor)|IMPLEMENTADO - TESTING PENDIENTE|Corregido el 11/11; falta desplegarlo a testing (paso 7).|
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

#### **Testing ejecutado el 15/11/2026**

|**Qué**|**Resultado**|
|---|---|
|`go vet ./...` y `go test ./...` en pkg, auth, labor-integration, notifications, education, human-promotion, campaigns, dashboard y rxresume|Todo pasa. Antes de los cambios también pasaban (auth, labor y notifications se corrieron al empezar).|
|Tests nuevos de backend|auth: bloqueo tras 5 intentos y vencimiento, refresh reutilizado, Gmail, DNI/dirección obligatorios, rechazo de cuentas, 423, auditoría (alta, edición sin valores, aprobar/rechazar, 403, 401 con y sin token, lecturas ignoradas), endpoint de auditoría, límite de 5 MB. labor: campos obligatorios, cruce de compatibilidad, evento al aprobar y al borrar, auditoría solo admin. notifications: reglas de `cuenta_rechazada`, `oferta.eliminada` y `oferta.compatible`. education: cursos que usan un material. campaigns: paginación.|
|`scripts/sync-eventlib.sh --check`|Las copias de la librería de eventos coinciden con `pkg/`.|
|Frontend: `ng test` (Karma, Chrome headless)|725 de 725 specs pasan (incluye los nuevos de login 423, registro Gmail, Usuarios, ofertas, auditoría, guards, material y campañas).|
|Frontend: `ng build`|Compila sin errores (queda una advertencia previa de una clave duplicada `axisIndex`).|
|Validación de `openapi.yaml` y `pipeline.yml`|YAML válido; sin `operationId` repetidos.|

**No ejecutado (pendiente):** los tests que necesitan el emulador de Firestore se saltean localmente (no está corriendo y no hay gcloud); el CI los corre con el emulador. Tampoco se corrieron k6, la sonda de seguridad ni pruebas contra el stack Docker, porque el stack no se reconstruyó (los contenedores no tienen el código nuevo y recrearlo requiere tu pedido) y comparte bases con testing. No hay E2E. Nada se probó en testing.

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

#### **Paso 2. Commit en develop** [Código, con tu OK]

- **Qué:** commitear todo en `develop` (sin ramas feature).
- **Cómo:** pedímelo y lo hago, o `git add -A && git commit -m "Auditoría 15/11: bloqueo de login, auditoría en auth y labor, rechazos, ofertas compatibles, RBAC de solo lectura"`. Revisá antes los archivos sin seguimiento de `docs/` que ya estaban (PDF de referencia, tesis) por si no querés incluirlos.
- **Resultado esperado:** commit nuevo sobre `d42cdfd`.
- **Verificación:** `git log -1 --stat`.

#### **Paso 3. Tests con emuladores (opcional, recomendado)** [Manual]

- **Qué:** correr los tests que hoy se saltean sin emulador.
- **Dónde:** tu máquina (con el emulador de Firestore y el de Pub/Sub levantados) o el CI.
- **Cómo:** `FIRESTORE_EMULATOR_HOST=127.0.0.1:8390 PUBSUB_EMULATOR_HOST=localhost:8086 go test ./...` en `pkg/`, `education/`, `human-promotion/` y `dashboard/`.
- **Resultado esperado:** todo pasa.
- **Verificación:** ningún `FAIL`; en el CI, los jobs `ci-*` en verde.

#### **Paso 4. Migración de la base de auth en Neon** [Credenciales]

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
|9|Dar a un participante de prueba una habilidad (entregar una actividad que la desarrolle) y aprobar una oferta con ese requisito (HU-44)|El participante recibe la notificación "Hay una oferta laboral para vos".|
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
