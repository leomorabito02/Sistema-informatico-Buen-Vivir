# Eje 1. Introducción a la calidad (Educación)

Objetivos de calidad medibles desde el usuario, el equipo y la organización, con las prácticas de proceso que los sostienen, y un plan de contingencia ante la caída de una dependencia crítica.

## 1. La idea en simple

- **Al usuario** le importa que el sistema sea "sencillo, rápido y confiable".
- **Al desarrollador**, que el código sea "claro, modular y fácil de modificar".
- **A la organización**, "reducir costos operativos, minimizar incidentes" y que el sistema "evolucione durante años".

**Quiénes son en Educación:**

- **Usuario:** participantes, docentes, voluntarios y administración del Centro "La Otra Orilla".
- **Equipo:** quienes desarrollan y mantienen el microservicio de Educación.
- **Organización:** el Centro de Ecología Integral, que no tiene presupuesto para servicios pagos y maneja datos de personas en situación vulnerable.

## 2. Objetivos de calidad medibles

**Estado de cada objetivo:**

- **Cumple:** medido y dentro de la meta.
- **Parcial:** medido con una muestra chica, o cumple con poco margen.
- **No cumple:** medido y fuera de la meta.
- **Sin medir:** todavía no hay dato.

### Mirada del usuario

| ID | Objetivo | Indicador y fórmula | Línea base | Meta | Cómo y cada cuánto se mide | Estado |
|---|---|---|---|---|---|---|
| OBJ-U1 | Educación hace lo que pidió el usuario | Historias de usuario de Educación con al menos una prueba que las verifica ÷ historias de usuario de Educación × 100 | Las 4 historias del plan de pruebas (HU-18, HU-19, HU-19bis y HU-20) tienen pruebas: 4 de 4. Tres tienen sus casos documentados en Jira (proyecto BVE); HU-20 tiene pruebas (`progress_service_test.go`) pero sus casos todavía no están cargados. Las demás historias de Educación no están en la matriz de trazabilidad (`05-plan-de-pruebas.md`, sección 5) | 100 % | Matriz de trazabilidad y Jira, en cada entrega | Parcial (medido solo sobre las historias del plan) |
| OBJ-U2 | Inscribir a un participante funciona y respeta las reglas | Reglas de inscripción verificadas por pruebas ÷ reglas de inscripción. Las reglas son: curso publicado, participante aprobado, sin inscripción activa repetida y con cupo disponible | 4 de 4 reglas con pruebas (`TestEnroll_CourseNotPublished`, `TestEnroll_ParticipantNotAccepted`, `TestEnroll_AlreadyEnrolled`, `TestEnroll_CapacityFull`). Cobertura del servicio de inscripciones: 100 % | 100 % | Pruebas automáticas, en cada cambio | Cumple |
| OBJ-U3 | Los docentes cargan actividades sin errores | Actividades que se guardan bien ÷ actividades válidas cargadas. Una actividad sin título, consigna o fecha tiene que rechazarse | Cobertura del servicio de actividades: 100 %. Falta medirlo con pedidos reales en el entorno de calidad | 100 % | Pruebas automáticas, en cada cambio, y prueba de API en el entorno de calidad | Parcial |
| OBJ-U4 | Los participantes pueden subir su entrega, con archivos | Entregas guardadas ÷ entregas válidas enviadas. Límites: hasta 5 archivos de 10 MB cada uno y una sola entrega por actividad | Cobertura del servicio de entregas: 100 %. El guardado de archivos en Cloud Storage tiene 6,3 % de cobertura: es la parte menos probada | 100 % | Pruebas automáticas, en cada cambio | Parcial |
| OBJ-U5 | La sugerencia de habilidades con IA responde aunque falle un proveedor | Pedidos de sugerencia que reciben respuesta (con sugerencias o con un aviso claro) ÷ pedidos totales | 14 pruebas del servicio, entre ellas "falla Gemini y responde OpenAI" y "fallan los dos". Cobertura: 98,4 % | 100 % de respuestas. Si fallan las dos IA, el docente carga las habilidades a mano | Pruebas automáticas, en cada cambio | Cumple |
| OBJ-U6 | Los cursos se ven aunque Contentful esté caído | Pedidos de cursos respondidos con Contentful caído ÷ pedidos totales × 100 | La prueba `TestContentfulClient_SirveContenidoStaleSiContentfulFalla` simula Contentful caído (responde 503) y comprueba que la consulta que ya estaba guardada se sigue respondiendo: 1 de 1. También comprueba los dos límites: una consulta que nunca se guardó devuelve error, y un 404 no se tapa con contenido viejo | 100 % mientras haya una versión guardada | Prueba automática con un servidor que imita a Contentful, en cada cambio | Cumple (con Contentful simulado, no cortando el real) |
| OBJ-U7 | Las pantallas de cursos cargan rápido | Tiempo de respuesta p95 de `GET /api/cursos` (el 95 % de los pedidos tarda menos que esto) | Sin dato real: el reporte de k6 que había en el repo era inventado y se borró. Sale del escenario ST-0 del eje 8 (`carga.js`), que ya está en el pipeline pero sin corrida guardada | p95 < 800 ms con 10 usuarios a la vez | k6 (`carga.js`) contra el entorno de calidad, en cada push a `Calidad` (`calidad-tp.yml`) | Sin medir |
| OBJ-U8 | Nadie ve datos que no le corresponden | Rutas protegidas que rechazan sin token (401) y con un rol equivocado (403) ÷ rutas protegidas probadas | 2 de 2 en la prueba manual del entorno de calidad | 100 % de las rutas protegidas | Pruebas de API automatizadas, en cada cambio | Parcial (muestra chica) |

### Mirada del equipo

| ID | Objetivo | Indicador y fórmula | Línea base | Meta | Cómo y cada cuánto se mide | Estado |
|---|---|---|---|---|---|---|
| OBJ-E1 | El código está probado | Cobertura de líneas de SonarQube, total y del código nuevo | 90,7 % y 90,7 % (medición sobre el código actual de la rama, `eb22c0e` con las pruebas corregidas). Era 92,4 % antes de incorporarse el trabajo nuevo | ≥ 80 % las dos | `calidad-tp.yml`, en cada cambio de `education/` | Cumple |
| OBJ-E2 | No hay errores graves | Bugs de severidad alta o bloqueante (SonarQube) | 0 | 0 | Quality gate, en cada cambio | Cumple |
| OBJ-E3 | El código se entiende | Funciones con complejidad cognitiva mayor a 15 | 19 (15 del servicio y 4 de pruebas). Eran 14 antes de incorporarse el trabajo nuevo, y 15 antes de la mejora del eje 4 | ≤ 6 | SonarQube, en cada cambio | No cumple |
| OBJ-E4 | No hay código copiado | Duplicación (SonarQube) | 0,2 % | ≤ 5 % | Quality gate, en cada cambio | Cumple |
| OBJ-E5 | Las pruebas pasan en cualquier PC | Pruebas en verde ÷ pruebas totales, en Linux (GitHub) y en Windows | Windows: 959 de 959 (`analizar.sh`). La prueba que fallaba solo en Windows ya pasa: el `.gitattributes` fuerza el salto de línea de Linux en los scripts. Linux: todas en verde en la última corrida de `calidad-tp.yml` | 100 % en los dos | `analizar.sh` en Windows y `calidad-tp.yml` en Linux | Cumple (en Linux, a confirmar en la próxima corrida del pipeline) |

**Por qué OBJ-E3 queda en "No cumple" y qué se hace.** La meta de 6 funciones o menos no se alcanza: quedan 19 (eran 14; el trabajo nuevo incorporado a la rama sumó 5). Se decidió no refactorizar más código del servicio para el TP, porque cada cambio en funciones que andan bien es un riesgo nuevo y la entrega está cerca. Lo que sí se hizo es la mejora medida del eje 4: la peor función (`skill_repository.go`, complejidad 38) bajó a menos de 15 sin cambiar ninguna prueba. El plan para el resto, de mayor a menor complejidad según la última medición de SonarQube, es:

1. Las del servicio con 23 o más: `enrollment_service.go:150` (25), `submission_repository.go:58` (24) y `general_attendance_repository.go:64` (23). Separarlas en pasos con nombre, igual que la mejora del eje 4.
2. Las 4 funciones de pruebas (30, 28, 22 y 21): pasarlas a pruebas por tabla.
3. Las demás del servicio (entre 16 y 22): revisarlas cuando se modifiquen por otro motivo, no antes.

Cada paso se mide con SonarQube antes y después, como la mejora del eje 4.

**La prueba que fallaba en Windows (resuelto).** Se llama `TestFiltrosDeInfraCoincidenConLosConsumidores`. Comprueba que la configuración de mensajería (el archivo `infra/pubsub/setup.sh`) le mande a Educación los mismos tipos de aviso que Educación sabe procesar. Para eso lee ese archivo. En Windows, Git guardaba los archivos con otro tipo de salto de línea (CRLF en lugar de LF), la prueba no reconocía el contenido y fallaba. **No era un error del servicio sino de cómo se guardaba el archivo en Windows.** El `.gitattributes` del repositorio ahora fuerza el salto de línea de Linux en todos los scripts (`*.sh text eol=lf`) y la prueba pasa también en Windows.

### Mirada de la organización

| ID | Objetivo | Indicador y fórmula | Línea base | Meta | Cómo y cada cuánto se mide | Estado |
|---|---|---|---|---|---|---|
| OBJ-O1 | No hay claves escritas en el código ni librerías con fallas de seguridad conocidas | Claves o contraseñas encontradas en los archivos (Gitleaks) + vulnerabilidades conocidas de librerías que el código usa (govulncheck) + vulnerabilidades de SonarQube | Gitleaks: 0 avisos en los archivos versionados de `education/`. Los 2 avisos que había eran valores de prueba (un token inventado en `PRUEBAS.md` y una clave solo para pruebas locales en `seguridad.sh`, que no coincide con ninguna configuración real); quedaron registrados con su motivo en `.gitleaksignore`. govulncheck: 0. SonarQube: 0 vulnerabilidades | 0 claves reales y 0 vulnerabilidades | `calidad-tp.yml`, en cada cambio | Cumple (verificado localmente; se confirma en la próxima corrida del pipeline) |
| OBJ-O2 | Los avisos automáticos entre servicios no llevan datos personales | Tipos de evento con campos de datos personales (nombre, DNI, teléfono, email, texto libre) | 0, según el catálogo de eventos `docs/eventos/asyncapi.yaml` | 0 | Revisión de cada evento nuevo | Cumple (por diseño) |
| OBJ-O3 | Lo desplegado es lo que se probó | Despliegues con imagen `education:<commit>` igual al commit con el gate aprobado ÷ despliegues | Sin medir (procedimiento en el eje 3, sección 5). Necesita un despliegue, que queda fuera del alcance de esta etapa del TP | 100 % | En cada despliegue | Sin medir |
| OBJ-O4 | Costo cero de herramientas de calidad | Pesos gastados en herramientas del TP | $0: SonarQube Community, GitHub Actions, emuladores y herramientas de seguridad, todos gratuitos | $0 | Al incorporar cada herramienta | Cumple |

**Qué es OBJ-O1, en simple.** Son dos riesgos distintos para el Centro:

- **Una clave escrita en el código.** Si alguien deja una contraseña o un token dentro de un archivo y ese archivo se sube a GitHub, cualquiera con acceso al repositorio puede usarla. Gitleaks revisa los archivos buscando cosas con forma de clave.
- **Una librería con una falla conocida.** Educación usa librerías de terceros. Cuando se descubre una falla en alguna, se publica en una base de datos pública. govulncheck compara las librerías de Educación con esa base y avisa solo si el código usa la parte que falla.

**Qué es un evento (OBJ-O2).** Cuando pasa algo importante en Educación, por ejemplo una inscripción, una entrega o una habilidad otorgada, el servicio manda un aviso automático a los otros servicios del sistema, como el Dashboard o Notificaciones. Ese aviso se llama **evento**. El objetivo es que el aviso lleve solo identificadores (por ejemplo, el id del participante) y nunca el nombre, el DNI ni otros datos personales. Si otro servicio necesita el dato, se lo pide al servicio dueño. Así, aunque un aviso se filtre, no expone a nadie.

## 3. Prácticas de proceso que sostienen los objetivos

Un objetivo sin una práctica que lo cuide se cumple de casualidad.

| Práctica | Qué objetivos sostiene | Dónde está |
|---|---|---|
| Quality gate en cada cambio, que frena si no se cumplen las metas | OBJ-E1, E2, E4 | `calidad-tp.yml` + gate "TP-Calidad-N3" |
| Pruebas automáticas antes de desplegar (unitarias, integración, contrato) | OBJ-U2 a U5, U8, E1, E5 | `pipeline.yml` (`ci-education`) |
| Análisis de seguridad en cada cambio: gosec, govulncheck, Gitleaks, Semgrep y Trivy | OBJ-O1 | `calidad-tp.yml` (job `seguridad-education`) |
| Decisiones escritas: ADR y registro de decisiones del TP | OBJ-O2, E3 | `docs/adr/`, `docs/calidad/decisiones.md` |
| Imagen con el commit en el nombre, smoke y vuelta atrás automática | OBJ-O3, U6 | `pipeline.yml` (`cd-education`, `cd-smoke`, `cd-rollback`) |
| Historial de métricas por commit | Todos los medidos | `docs/calidad/reportes/historial.csv` |
| Entorno de prueba aislado, con datos de prueba | OBJ-U3, U6, U7, U8 | `docker-compose.calidad.yml` |

**Qué es un ADR.** Son las siglas de *Architecture Decision Record* (registro de decisión de arquitectura). Es un documento corto que deja escrito qué se decidió sobre el diseño del sistema, por qué y qué otras opciones había. Por ejemplo, el ADR de la caché de contenido explica por qué Educación guarda una copia de los cursos de Contentful. Sirven para que quien llegue después entienda las decisiones sin tener que adivinarlas. Están en `docs/adr/`.

## 4. Plan de contingencia ante la caída de una dependencia crítica

Para cada dependencia separamos tres cosas, como pide la consigna:

- **Hoy:** lo que el sistema hace ahora;
- **Propuesta:** lo que proponemos;
- **Implementado en el TP:** lo que realmente hicimos.

| Dependencia | Qué tan crítica | Si se cae, hoy… | Detección | Propuesta | Implementado en el TP |
|---|---|---|---|---|---|
| **Contentful** (cursos, materiales) | Alta: es la plataforma externa principal | Se sirve la última versión guardada (caché de 5 min, hasta 500 respuestas; ADR 0006). Los cambios de contenido (crear o publicar cursos) fallan. Si la instancia se reinicia sin caché, los listados fallan | Aviso en el log. Smoke de `/api/salud` (no consulta Contentful) | Prueba automática que corte Contentful y verifique que se sigue respondiendo (OBJ-U6) | Sí: `TestContentfulClient_SirveContenidoStaleSiContentfulFalla`, con Contentful simulado |
| **Firestore** (inscripciones, actividades, asistencia, entregas, habilidades) | Muy alta | Si no se puede crear la conexión al arrancar, el servicio no arranca (`log.Fatalf`). Si cae con el servicio andando, los pedidos que leen o escriben Firestore dan error. **No hay respaldo** | Errores en el log y alertas de Cloud Run | Mensaje claro al usuario ("servicio temporalmente no disponible") y monitoreo de la tasa de errores | No. El riesgo se acepta (ver abajo) |
| **Pub/Sub** (eventos) | Baja para el usuario | Los eventos quedan guardados en Firestore y se reintentan sin límite, cada vez más espaciados (5 s a 10 min). No se pierden (ADR 0004) | Eventos acumulados en `outbox_eventos` | Alerta si el outbox pasa cierto tamaño | El reintento sí está probado (`TestRelayReintentaYRespetaOrderingKey`, `TestRelayBackoff`). La alerta no |
| **Cloud Storage** (adjuntos de entregas) | Media | No se pueden subir ni bajar adjuntos. Las entregas con solo texto siguen funcionando | Error al subir | Guardar el adjunto en una caché temporal y subirlo a Cloud Storage cuando vuelva. Mientras tanto, la entrega queda registrada y el participante ve que el archivo está pendiente. **Todavía no está implementado** | No. Sí está probado que una entrega con solo texto funciona sin bucket (`TestSubmissionService_Submit_FilesWithoutBucket`) |
| **IA (Gemini y OpenAI)** | Baja | Se prueba primero con Gemini y después con OpenAI. Si los dos fallan, las habilidades se cargan a mano | Mensaje "no se pudo generar sugerencias" | — | Sí: `TestSuggest_GeminiFailsAndOpenAIResponds` y `TestSuggest_BothFail` |
| **Un despliegue con errores** | Alta | Si el smoke falla, `cd-rollback` vuelve a la versión anterior | Smoke (`cd-smoke`) | Agregar el commit a `/api/salud` | No. Es parte del despliegue, fuera del alcance de esta etapa del TP |

**Dependencia más crítica: Firestore.** Es la única sin plan B: si cae, Educación deja de funcionar. Para un centro barrial, montar un respaldo de base de datos es caro y complejo. La decisión razonable es aceptar ese riesgo, saberlo y avisar bien al usuario.

### Qué hacer si se cae (procedimiento para el equipo)

1. **Confirmar qué se cayó.** Mirar `/api/salud` y el log de Cloud Run de Educación, y la página de estado del proveedor.
2. **Si es Contentful:** no hacer nada urgente. La caché responde. Avisar a los docentes que no publiquen cursos hasta que vuelva.
3. **Si es Firestore:** avisar a la administración del Centro que Educación no está disponible. No reintentar despliegues.
4. **Si fue un despliegue:** comprobar que `cd-rollback` haya vuelto a la versión anterior. Si no, volver a mano a la revisión anterior de Cloud Run.
5. **Después:** registrar el incidente como defecto en Jira, con qué pasó, cuánto duró y qué se cambia. La métrica de proceso es el tiempo hasta recuperar el servicio.
