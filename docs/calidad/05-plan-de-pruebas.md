# Eje 6. Plan de pruebas, riesgos y trazabilidad (Educación)

**Qué pide la consigna (Nivel 3):** plan versionado con criterios de entrada
y salida, riesgos priorizados por probabilidad × impacto y trazabilidad
requisito → caso → defecto; defectos registrados con pasos, esperado y
obtenido, vinculados a su caso de prueba en Jira u otra herramienta.

**Versionado:** este archivo vive en el repositorio (`docs/calidad/`). Cada
cambio del plan queda en el historial de git.

---

## 1. Objetivo

Asegurar que el microservicio de Educación gestione de manera confiable y
segura el catálogo formativo, las inscripciones, la asistencia y el cálculo
del avance, y que siga respondiendo cuando falla una dependencia externa.

## 2. Alcance

**Dentro del alcance:**

- Cursos, áreas formativas y materiales (Contentful).
- Actividades, entregas, asistencia por actividad e inscripciones (Firestore).
- Asistencia general diaria (Firestore).
- Validación de participantes contra auth-service.
- Seguridad de las rutas: token JWT, roles, límite de pedidos, CORS y cabeceras.
- Historias de usuario del plan: HU-18, HU-19, HU-19bis y HU-20.

**Fuera del alcance:**

- El frontend (Angular).
- El esquema de la base de datos de auth-service.
- Los mails de inscripción y recordatorio (HU-18 y HU-26 en su parte de
  notificaciones), que dependen de un servicio que todavía no existe.
- En esta etapa del TP: la automatización del despliegue. El smoke, el E2E,
  la carga y el stress corren en el CI del TP contra Educación levantada en
  el runner, sin desplegar.

## 3. Estrategia y niveles de prueba

| Nivel | Qué prueba | Cómo |
|---|---|---|
| Unitaria | Una función sola, con lo demás simulado | `go test` con testify/mock y `httptest` |
| Integración | Dos piezas juntas | Repositorios contra el emulador de Firestore; outbox contra el emulador de Pub/Sub; contratos con el frontend |
| Sistema | El servicio entero funcionando | Smoke, E2E, carga, ZAP y stress contra el entorno de calidad, en el job `servicio-education` de `calidad-tp.yml` |
| Aceptación | Que haga lo que pidió el usuario | Casos CP-xx contra los criterios de cada historia |

Contrato, carga, seguridad y mutación no son niveles: son **tipos de
prueba** que se hacen dentro de un nivel. Las pruebas de contrato de
Educación están escritas a mano **al estilo Pact** (`tests/contract`): no
usan la librería Pact, que no está en `go.mod`.

## 4. Criterios de entrada, salida, suspensión y reanudación

| Criterio | Condición |
|---|---|
| **Entrada** (se puede empezar a probar) | El código compila (`go build ./...`), el emulador de Firestore responde y cada historia del alcance tiene sus criterios de aceptación escritos |
| **Salida** (se puede dar por terminada la prueba) | Todas las pruebas en verde, quality gate TP-Calidad-N3 aprobado, todos los casos planificados con su prueba automatizada y ningún defecto de severidad alta abierto en Jira |
| **Suspensión** (se frena) | No compila; no hay Docker o emulador para las pruebas de integración; la cobertura cae por debajo del 80 % |
| **Reanudación** (se retoma) | Se corrige la compilación, se levantan los servicios que faltaban o se agregan pruebas hasta volver al umbral |

**Estado al cierre de esta versión del plan:** se cumplen la entrada y la
salida: 959 pruebas en verde, gate aprobado, todos los casos planificados
con su prueba y ningún defecto abierto. Durante la prueba se aplicó una vez
el criterio de suspensión: al incorporarse trabajo nuevo, 4 pruebas
quedaron en rojo y no se pudo medir hasta corregirlas (DEF-07).

## 5. Trazabilidad requisito → caso → defecto

Cada caso está cargado en Jira (proyecto BVE) como subtarea de su historia.
La columna "Prueba automatizada" es la prueba de `education/` que lo
verifica.

| Historia (Jira) | Caso (Jira) | Qué verifica | Prueba automatizada | Defecto |
|---|---|---|---|---|
| HU-18 (BVE-2) | CP-01 (BVE-6) | Inscripción correcta | `TestEnroll_Success` | — |
| | CP-02 (BVE-7) | Inscripción duplicada rechazada | `TestEnroll_AlreadyEnrolled` | — |
| | CP-03 (BVE-8) | Curso sin cupo (valores límite) | `TestEnroll_CapacityFull`, `TestEnroll_CapacityWithRoomLeft`, `TestEnroll_InactiveEnrollmentAllowsReEnroll` | — |
| | CP-04 (BVE-9) | Curso no publicado | `TestEnroll_CourseNotPublished` | — |
| | CP-05 (BVE-10) | Curso inexistente | `TestEnroll_CourseDoesNotExist` | — |
| | CP-06 (BVE-11) | Participante inexistente | `TestEnroll_ParticipantDoesNotExist` | DEF-01 (BVE-28) |
| | CP-07 (BVE-12) | Participante no aprobado | `TestEnroll_ParticipantNotAccepted`, `TestEnroll_AcceptsApprovedAndLegacyAccepted` | — |
| | CP-08 (BVE-13) | Control de acceso por rol | `TestRegisterRoutes_RBAC_DecisionTable`, `TestRequireRole_InsufficientRole` | — |
| | CP-09 (BVE-14) | Persistencia en Firestore | `TestEnrollmentRepository_ListByCourse_FiltersByCourseID`, `TestEnrollmentController_List` | — |
| HU-19 (BVE-3) | CP-10 (BVE-15) | Entrega con asistencia automática | `TestSubmissionService_Submit_TextOnly_CreatesSubmissionWithAttendance` | — |
| | CP-11 (BVE-16) | Consulta de asistencia de una actividad | `TestAttendanceService_ListByActivity_Success` | — |
| | CP-12 (BVE-17) | Participante no inscripto | `TestSubmissionService_Submit_NotEnrolledIsForbidden` | — |
| | CP-13 (BVE-18) | Inscripción dada de baja | `TestSubmissionService_Submit_NotActiveEnrollmentIsForbidden` | — |
| | CP-14 (BVE-19) | Actividad inexistente | `TestSubmissionService_Submit_ActivityNotFound`, `TestSubmissionController_Submit_ActivityNotFound` | — |
| | CP-15 (BVE-20) | Entrega repetida | `TestSubmissionService_Submit_AlreadySubmitted` | — |
| | CP-16 (BVE-21) | Docente que no es titular | `TestDocenteController_ListAttendance_NotOwnerIsForbidden` | — |
| HU-19bis (BVE-4) | CP-18 (BVE-22) | Listado del día | `TestGeneralAttendance_ListByDate_CombinesParticipantsUsersAndRecords` | — |
| | CP-19 (BVE-23) | Registrar presente | `TestGeneralAttendance_RegisterBatch_Success`, `TestGeneralAttendanceController_ListAndRegister` | — |
| | CP-20 (BVE-24) | Registrar ausente | `TestGeneralAttendance_RegisterBatch_Success`, `TestGeneralAttendanceController_ListAndRegister` | — |
| | CP-21 (BVE-25) | Observación | `TestGeneralAttendanceRepository_RegisterAndList` | — |
| | CP-25 (BVE-26) | Edición sin duplicados y auditoría | `TestGeneralAttendanceRepository_RegisterAndList`, `TestGeneralAttendance_RegistradaYActualizada`, `TestGeneralAttendance_RegistraAuditoria` | — |
| | CP-27 (BVE-27) | Fecha inválida | `TestGeneralAttendance_ListByDate_InvalidDate`, `TestGeneralAttendance_RegisterBatch_Validations` | — |
| HU-20 (BVE-5) | CP-28 (BVE-42) | Curso sin actividades: 0 %, sin división por cero | `TestProgressService_CalculateCourseProgress_NoActivities` | — |
| | CP-29 (BVE-43) | 1 de 2: 50 % (una asistencia de otro curso no cuenta); límites 0 de 3 → 0 % y 3 de 3 → 100 % | `TestProgressService_CalculateCourseProgress_ComputesPercentage`, `TestProgressService_CalculateCourseProgress_BoundaryValues` | — |
| Eventos de cursos (fuera de las 4 HU del plan) | CP-30 (BVE-44) | Al crear o publicar un curso, el evento se guarda en el outbox sin transacción | `TestCourseService_PublicaEventosDeCursos` | DEF-06 (BVE-29) |
| HU-16 y HU-18 (cambio de comportamiento) | CP-01 y CP-09, y las pruebas de baja lógica | La suite tiene que reflejar el comportamiento vigente | `TestCourseService_DeleteCourse_Success`, `TestCourseService_DeleteCourse_RepoError`, `TestCourseService_KVSync_CalledOnMutations`, `TestEnrollmentController_DecisionTable` | DEF-07 (BVE-41) |

**Resumen:** 4 de 4 historias del plan tienen casos y pruebas; 25 casos
cargados en Jira y verificados (CP-01 a CP-30, con los huecos explicados
abajo); 3 defectos, los tres con su prueba de regresión.

**Numeración.** Los números CP-17, CP-22, CP-23, CP-24 y CP-26 no tienen
caso asignado. Se dejan sin usar para no cambiar los identificadores que ya
se citan en el TP y en Jira.

**Hueco detectado en HU-20 (resuelto).** El TP describía los valores límite
"0 de N actividades → 0 %" y "N de N → 100 %", pero ninguna prueba los
verificaba. Se agregó `TestProgressService_CalculateCourseProgress_BoundaryValues`,
con tres subcasos (sin asistencias, todas como ausente y todas presentes),
y pasa.

## 6. Riesgos priorizados (probabilidad × impacto)

Escala de 1 a 5 para probabilidad e impacto, y severidad = P × I (de 1 a
25). Es una escala distinta a la del top 3 de ISO 25010 (eje 3), que usa 1
a 3 porque compara solo 8 características: acá hace falta más resolución
para ordenar riesgos concretos.

### Riesgos de producto

| ID | Riesgo | P | I | P × I | Nivel | Mitigación implementada |
|---|---|---|---|---|---|---|
| R-01 | Registros de asistencia duplicados por reintentos o concurrencia | 4 | 4 | 16 | Crítico | ID determinístico: `{actividadId}_{participanteId}` en asistencia y `{fecha}_{participanteId}` en asistencia general. Probado en CP-25 |
| R-02 | auth-service caído o lento al validar participantes | 3 | 4 | 12 | Alto | El cliente distingue "no existe" (400 de negocio) de una falla de red o 500, que se responde como 502 sin dejar datos a medias (`http_helpers.go`) |
| R-03 | Dos inscripciones simultáneas superan el cupo | 3 | 3 | 9 | Medio | Se cuentan las inscripciones activas antes de inscribir. **Riesgo residual:** el conteo y la escritura no están en la misma transacción |
| R-04 | Dos ediciones simultáneas de un curso en Contentful | 2 | 4 | 8 | Medio | Se envía `X-Contentful-Version` con la versión leída; si hay choque, Contentful responde 409 y el error se propaga (`contentful_client.go`) |
| R-05 | El servicio arranca con una clave JWT insegura | 1 | 5 | 5 | Bajo | Al arrancar, si `JWT_SECRET` está vacío o es uno de cuatro valores por defecto conocidos, el servicio se detiene (`internal/services/jwt_service.go`). **No controla la longitud ni la entropía:** una clave débil que no sea de esa lista se acepta |

### Riesgos de proyecto

| ID | Riesgo | P | I | P × I | Nivel | Mitigación |
|---|---|---|---|---|---|---|
| RP-01 | No llegar a ejecutar stress, esfuerzo sostenido y E2E antes de la entrega | 4 | 4 | 16 | Crítico | Los scripts ya están en el CI del TP (job `servicio-education`); tareas en Jira con responsable. Si no hay corrida con resultados, el informe final lo declara como meta no cumplida |
| RP-02 | Los números del informe quedan viejos cuando cambia el código | 4 | 3 | 12 | Alto | Ya pasó al incorporarse a la rama trabajo nuevo. Cada medición queda en `historial.csv` con su commit, y se vuelve a medir antes de entregar |
| RP-03 | La automatización depende de un solo integrante | 3 | 4 | 12 | Alto | División de tareas escrita y tareas en Jira; los dos integrantes tienen que poder explicar todo el TP |
| RP-04 | Las pruebas contra Contentful real superan el límite de pedidos del plan gratuito | 2 | 3 | 6 | Medio | Las pruebas de código usan un servidor que imita a Contentful. En el CI, el E2E escribe un curso y la carga unos 4 por corrida, y los borran; stress y esfuerzo sostenido solo leen, desde la caché de 5 min |
| RP-05 | El análisis de SonarQube depende de una PC | 2 | 2 | 4 | Bajo | El pipeline del TP levanta su propio SonarQube en cada corrida |

## 7. Entregables de prueba

| Entregable | Dónde | Estado |
|---|---|---|
| Perfil de cobertura | `coverage.out`, generado por `analizar.sh` y por el pipeline del TP | Se genera en cada corrida |
| Historial de métricas | `docs/calidad/reportes/historial.csv` | Una fila por medición |
| Reporte de contratos (al estilo Pact) | `reportes-servicio/pact/education-pact-report.json` | Se genera en cada corrida (no se versiona) |
| Reporte de integridad de Firestore | `reportes-servicio/database/education-firestore-integrity.json` | Igual que el anterior |
| Reporte de mutación | `education/mutation_report.txt` (lo escribe `scripts/calidad/pruebas/mutacion.sh`; no se versiona) | De corridas locales: la corrida local del 24/10 (commit `0aa6b29`) dio 97,28 % (500 mutantes eliminados y 14 sobrevivientes; los 56 que excedieron el tiempo no cuentan). El CI del TP la corre a pedido (job `mutacion-education`) |
| Seguridad estática | Artefactos del job `seguridad-education` | Se generan en cada corrida |
| Seguridad dinámica (ZAP) | `education/zap_report.html` (lo escribe `scripts/calidad/seguridad/seguridad.sh`; no se versiona) y artefacto `zap-education-<commit>` | Escaneo sin iniciar sesión: solo vio respuestas de rechazo |
| Casos y defectos | Jira, proyecto BVE | Ver sección 5 |

## 8. Registro de defectos

Los defectos se registran en Jira como tipo "Error", con la plantilla del
TP: historia afectada, componente, severidad y prioridad, precondición,
pasos para reproducir, resultado esperado y obtenido, causa raíz,
corrección, commit donde apareció y del arreglo, y prueba de regresión.

| Defecto | Jira | Caso que lo relaciona | Prueba de regresión | Estado |
|---|---|---|---|---|
| DEF-01: panic con el cliente de auth-service sin inicializar | BVE-28 | CP-06 | `TestAuthServiceClient_NotConfigured` | Corregido y verificado |
| DEF-06: panic al publicar eventos de cursos | BVE-29 | CP-30 | `TestCourseService_PublicaEventosDeCursos` | Corregido (`e3357a6`) y verificado |
| DEF-07: 4 pruebas desactualizadas al incorporar la baja lógica (HU-16) y las inscripciones con nombre | BVE-41 | CP-01, CP-09 | Las 4 pruebas corregidas | Corregido y verificado (959 pruebas en verde) |

Un defecto pasa a "Listo" solo cuando su prueba de regresión está en verde:
corregido no es lo mismo que verificado.

## 9. Correcciones respecto de la versión anterior del TP

| Dónde | Decía | Corrección |
|---|---|---|
| Matriz de riesgos, R-05 | "aborta si la clave no cumple con entropía mínima" | Solo rechaza la clave vacía y cuatro valores por defecto conocidos |
| Estrategia y entregables | "Pact" | Pruebas de contrato escritas a mano al estilo Pact; no se usa la librería |
| Entregables y resultados | La mutación como parte del pipeline | En `pipeline.yml` el job está apagado; el 91,36 % salió de una corrida local. El CI del TP ya la corre a pedido |
| Valores límite de HU-20 | "0 de N → 0 %" y "N de N → 100 %" como probados | No tenían prueba; se agregó (sección 5) |
| Casos documentados | Numeración CP-01 a CP-27 sin explicar | Se explican los huecos (sección 5) |
