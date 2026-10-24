# Eje 8. Herramientas del ecosistema (Educación)

**Qué pide la consigna:** mapa de herramientas, una herramienta nueva
incorporada con el proceso de cinco pasos, registro en Jira u otra, y pruebas
de stress y de esfuerzo sostenido.

---

## 1. Mapa de herramientas

La clase de herramientas del ecosistema dice que "ninguna herramienta lo hace todo: cada una
resuelve una parte del ciclo de pruebas". Este mapa usa sus categorías y
suma las de infraestructura que usa la tesis. Lo que figura como "en uso"
está verificado en el repo: `pipeline.yml`, `go.mod` y los scripts.

| Categoría | Qué resuelve | Herramienta en Educación | Dónde | Estado |
|---|---|---|---|---|
| Gestión de pruebas | Planificar, organizar y trazar | **Jira** (proyecto BVE "BuenVivir-CalidadEducacion") | ucc-team-z0eeow4u.atlassian.net/browse/BVE-1 | En uso |
| Análisis estático | Revisar el código sin ejecutarlo | **SonarQube** (nueva) | `calidad-tp.yml`, `scripts/calidad/` | En uso |
| | | golangci-lint (estilo y errores comunes de Go) | `pipeline.yml` | En uso |
| | | gosec y Semgrep (patrones de código inseguro) | `calidad-tp.yml` (gosec también en `pipeline.yml`) | En uso |
| Diseño / BDD | Especificar el comportamiento | Contratos con Pact (no es BDD, pero especifica qué espera cada consumidor) | `education/tests/contract` | En uso |
| Ejecución / registro | Automatizar y correr pruebas | `go test` + testify | `education/**/*_test.go` | En uso |
| | | Emulador de Firestore | `pipeline.yml`, `docker-compose.calidad.yml` | En uso |
| | | gremlins y go-mutesting (pruebas de mutación: ¿las pruebas detectan un cambio en el código?) | `pipeline.yml` | En uso |
| Rendimiento | Medir carga, stress y escala | k6 | `education/load_test.js` (Educación); en `pipeline.yml` k6 corre contra auth-service, no contra Educación | Script liviano, sin resultados reales para Educación |
| Seguridad (extra) | Vulnerabilidades y secretos | govulncheck (librerías con fallas conocidas), Gitleaks (claves en el código), Trivy (imagen Docker) | `calidad-tp.yml` y `pipeline.yml` | En uso |
| | | OWASP ZAP (ataques sobre la app desplegada) | `pipeline.yml` | Solo contra auth-service, no contra Educación |
| CI/CD (extra) | Que todo corra solo en cada cambio | GitHub Actions | `.github/workflows/` | En uso |
| Contenedores y nube (extra) | Empaquetar y desplegar | Docker, Docker Compose, Cloud Run | `Dockerfile`, compose, `pipeline.yml` | En uso |
| Control de versiones (extra) | Historial y trabajo en equipo | Git y GitHub | — | En uso |

**Qué muestra el mapa:** había herramientas de sobra en ejecución y en
seguridad. Faltaban dos cosas:
- una vista **única y medible** de la calidad del código, con un umbral que
  frene: eso lo resuelve SonarQube;
- la **gestión** de pruebas y defectos: eso lo resuelve Jira (sección 3).

## 2. Herramienta nueva: SonarQube, con los cinco pasos

La clase define 5 pasos y aclara que "el proceso es iterativo, no
una línea recta obligatoria".

### Paso 1. Requisitos

Las preguntas de la clase, respondidas para el TP:

| Pregunta de la clase | Respuesta |
|---|---|
| ¿Qué problema concreto queremos resolver? | No había forma de medir la calidad interna del código ni de frenar un cambio que empeore las metas del Nivel 3. Por ejemplo, la cobertura del código nuevo no se medía |
| ¿Quién la va a usar y con qué frecuencia? | El equipo de Educación, en cada cambio de `education/` |
| ¿Con qué otras herramientas tiene que integrarse? | Go (la cobertura de `go test`), Git (para saber qué código es nuevo) y GitHub Actions |
| ¿Qué presupuesto y plazo son razonables? | $0. Funcionando antes de la revisión del TP |
| ¿Qué pasa si no la implementamos? | No se cumple el aspecto "Quality gate" del Nivel 3 y las métricas del informe serían opiniones |

**Requisitos mínimos:**
- R1: analiza Go;
- R2: importa la cobertura;
- R3: tiene un quality gate configurable con las 6 condiciones del Nivel 3;
- R4: distingue el código nuevo;
- R5: es gratis para un repo privado;
- R6: se puede automatizar en GitHub Actions.

### Paso 2. Investigación de mercado

Lista corta de 3 candidatas, como pide la clase:

| Candidata | Qué es | ¿Cumple los requisitos? |
|---|---|---|
| **SonarQube Community Build** | La versión gratuita para instalar uno mismo | R1 a R6. Limitación: no analiza ramas ni pull requests por separado |
| **SonarQube Cloud** (antes SonarCloud) | La misma herramienta como servicio en internet | R1 a R4 y R6. **R5 no**: *aporte externo, según la página del proveedor; es gratis solo para repos públicos, y el repo de la tesis es privado* |
| **golangci-lint** | El linter que ya usa la tesis | R1, R5 y R6. **No** cumple R2, R3 ni R4: no mide cobertura ni duplicación, no tiene historial ni quality gate |

### Paso 3. Prueba de concepto

La clase pide un alcance acotado, criterios de éxito fijados antes, tiempo
limitado y el equipo que la va a usar.

- **Alcance:** solo Educación, en una PC del equipo.
- **Tiempo:** 1 día.
- **Criterios de éxito:** los requisitos del paso 1, que salen de las metas del Nivel 3 de la consigna y se fijaron antes del primer análisis:

| Criterio | Resultado | Evidencia |
|---|---|---|
| Analiza los 92 archivos Go de Educación | Cumple | 8.110 líneas analizadas |
| Importa la cobertura y los resultados de `go test` | Cumple | Cobertura 92,5 % con 886 pruebas |
| El gate tiene las 6 condiciones del Nivel 3 | Cumple | `configurar_sonar.sh` |
| Mide el código nuevo desde agosto | Cumple, después de corregirlo | En el primer intento dio "0 líneas nuevas". Se resolvió con `linea_base.sh` |
| Se repite igual y deja historial | Cumple | `analizar.sh` + `historial.csv` |
| Costo $0 | Cumple | Community Build y GitHub Actions gratis |

**Lo que aprendimos en la PoC:**
1. **El código nuevo necesita un análisis anterior para comparar.** No
   estaba en la guía de clase y casi nos lleva a presentar un dato vacío.
2. **La edición gratuita usa un modo de calidad nuevo (MQR)**, con nombres
   de métricas distintos a los de la guía. El gate se armó con los nombres
   nuevos.
3. **Se había perdido la contraseña de administrador.** Se arrancó con una
   base nueva y se guardó la anterior.

### Paso 4. Evaluación objetiva: matriz de decisión

Se usan los criterios y los pesos del ejemplo de la clase.
Puntajes de 1 a 10.

| Criterio | Peso | SonarQube Community | SonarQube Cloud | golangci-lint |
|---|---|---|---|---|
| Funcionalidad | 30 % | 8: gate, cobertura, duplicación, historial; sin análisis de PR | 9: lo mismo, más análisis de PR | 4: solo reglas de código |
| Costo (TCO) | 25 % | 10: $0, solo tiempo de instalación | 3: pago para un repo privado | 10: $0 |
| Facilidad de uso | 20 % | 7: hay que mantener un servidor (Java) | 9: no hay nada que instalar | 8: un comando |
| Soporte | 15 % | 7: documentación y foro de la comunidad | 8: lo mismo, más soporte del proveedor | 7: comunidad |
| Integración | 10 % | 8: scanner + script en GitHub Actions | 9: integración directa con GitHub | 9: ya está en el pipeline |
| **Total ponderado** | 100 % | **8,15** | 7,35 | 7,25 |

**Cuenta:** cada puntaje se multiplica por su peso y se suman los
resultados. Por ejemplo, SonarQube Community da 0,30×8 + 0,25×10 + 0,20×7 +
0,15×7 + 0,10×8 = **8,15**.

**Decisión:** SonarQube Community Build. golangci-lint queda
como complemento, no como reemplazo.

### Paso 5. Despliegue incremental

La clase propone 3 etapas: piloto, expansión y organización.

| Etapa | Qué es en el TP | Estado |
|---|---|---|
| 1. Piloto | SonarQube local, analizando Educación a mano con `analizar.sh` | Hecho |
| 2. Expansión | Pipeline `calidad-tp.yml`: análisis automático en cada cambio, con todas las pruebas de Educación | Hecho: corre en GitHub y dio verde |
| 3. Organización | Propuesta: sumar el paso de SonarQube a `pipeline.yml` y a los demás microservicios de la tesis | Fuera del alcance del TP. Queda como recomendación |

## 3. Registro en Jira

La clase lo resume así: "testing, desarrollo y gestión de proyecto
comparten una sola fuente de verdad".

**Dónde:** proyecto **BVE** ("BuenVivir-CalidadEducacion") en
`ucc-team-z0eeow4u.atlassian.net`, gestionado por el equipo, con tablero
Scrum. No se reutilizó el proyecto "Testing", que tiene los ejercicios de
clase.

**Cómo está organizado:**

| Tipo de issue | Qué representa | Issues | Estado |
|---|---|---|---|
| Epic | El TP completo: "Calidad de Software - Microservicio Educación" | BVE-1 | Por hacer |
| Historia | Cada historia de usuario del plan de pruebas, con sus criterios de aceptación | BVE-2 (HU-18), BVE-3 (HU-19), BVE-4 (HU-19bis), BVE-5 (HU-20) | Listo, salvo HU-20 (le faltan sus casos) |
| Subtarea con la etiqueta `caso-de-prueba` | Cada caso de prueba, hija de su historia: técnica, precondición, entrada, resultado esperado y obtenido, y la prueba automatizada que lo verifica | BVE-6 a BVE-27 (CP-01 a CP-27) | Listo |
| Error | Cada defecto real: pasos, resultado esperado y obtenido, causa, commit donde apareció y del arreglo, prueba de regresión | BVE-28 (DEF-01), BVE-29 (DEF-06) | Listo |
| Tarea con la etiqueta `pendiente-tp` | Lo que falta para la entrega | BVE-30 a BVE-40 | Por hacer |

No se instaló Zephyr ni Xray: el tipo "caso de prueba" se reemplazó por
subtareas con una etiqueta. Así cada caso cuelga de su historia y el
tablero muestra solo qué casos tiene cada una. Las etiquetas `eje-1` a
`eje-8` y `transversal` permiten filtrar por eje de la consigna.

**Trazabilidad:** la cadena requisito → caso → defecto queda armada con la
jerarquía (historia → subtareas) y con vínculos "se relaciona con":
DEF-01 con CP-06 y HU-18; DEF-06 con la tarea que documenta su caso de
prueba. La matriz completa está en `05-plan-de-pruebas.md`, sección 5.

**Ciclo de vida de un defecto en el tablero:** Por hacer → En curso →
Listo. Como dice la clase de QA y QC, "corregido por el desarrollador" no
es lo mismo que "verificado": un defecto pasa a Listo solo cuando su
prueba de regresión está en verde.

**Consultas útiles** (en Jira la búsqueda usa el nombre interno del tipo
"Error", que es `Bug`):

- Defectos: `project = BVE AND issuetype = Bug`
- Casos de prueba: `project = BVE AND labels = caso-de-prueba`
- Pendientes: `project = BVE AND labels = pendiente-tp AND statusCategory != Done`

## 4. Pruebas de stress y de esfuerzo sostenido: plan

> **Estado: plan escrito, sin ejecutar.** No hay resultados todavía. La
> ejecución corresponde al eje de automatización (tarea BVE-31 en Jira). El
> reporte de k6 que está en `reports/performance/` no es una ejecución real:
> lo genera un script con valores fijos. Hasta que haya una corrida, el
> aspecto "Stress y esfuerzo sostenido" del Nivel 3 no se cumple.

La clase distingue tres pruebas:
- **Carga:** cómo se comporta con el tráfico esperado.
- **Stress:** el punto de quiebre, cuánto soporta hasta fallar.
- **Escalabilidad:** cómo responde al crecer.

La consigna agrega el **esfuerzo sostenido**: una carga normal durante mucho
tiempo, para ver si el servicio se degrada.

### Dónde y con qué

- **Contra el entorno de calidad** (`docker-compose.calidad.yml`, puerto
  18210). **Nunca contra producción** (regla del TP).
- **Herramienta: k6.** Ya está en el repo y en el pipeline. *Aporte externo:*
  la clase muestra JMeter, Gatling y BlazeMeter. Se eligió k6 porque el
  equipo ya lo usa: introducir otra herramienta sería repetir los 5 pasos
  sin necesidad.
- **Usuarios:** tokens de prueba de `token_prueba.py`, con los datos que
  carga `cargar_por_api.py` usando la propia API de Educación.
- **Cuidado con Contentful:** es un servicio real de un tercero. Los
  escenarios calientan la caché antes de empezar (5 min de vida), para no
  descargar la carga sobre Contentful. *Aporte externo:* Contentful limita la
  cantidad de pedidos por segundo, así que un stress sin caché mediría a
  Contentful y no a Educación.
- **Recursos:** el contenedor de Educación tiene un límite de 1 GB de memoria
  y 1 CPU, igual que en Cloud Run (`--memory=1Gi --cpu=1`). Así los
  resultados se parecen a los de producción.

### Escenarios

| ID | Tipo | Qué pide cada usuario virtual | Carga | Duración | Umbral de aprobación | Cuándo cortar |
|---|---|---|---|---|---|---|
| ST-0 | Carga (referencia) | `GET /api/cursos`, `GET /api/cursos/{id}/actividades`, `GET /api/cursos/{id}/inscripciones` | 20 usuarios | 5 min | p95 < 800 ms y errores < 1 % | — |
| ST-1 | Stress | Lo mismo | Escalera: 20 → 50 → 100 → 200 usuarios, 2 min cada escalón | ~8 min | Se registra el escalón donde p95 > 2 s o errores > 5 %: ese es el **punto de quiebre** | Errores > 20 %, el contenedor se reinicia, o la PC se queda sin memoria |
| ST-2 | Esfuerzo sostenido | Lo mismo | 20 usuarios constantes | 60 min | El p95 de los últimos 10 min no supera en más de 20 % al de los primeros 10 min. La memoria del contenedor no crece sin parar | Errores > 5 % sostenidos durante 2 min |

**Qué se guarda de cada corrida:** el resumen JSON de k6
(`--summary-export`), la memoria y CPU del contenedor (`docker stats`), y el
commit y la fecha. Va a `docs/calidad/reportes/stress/`, con el mismo
criterio de evidencia que el historial de SonarQube.

**Qué se espera ver (hipótesis, a confirmar):**
- Las lecturas de cursos salen de la caché y deberían aguantar mucho.
- El primer cuello de botella debería ser Firestore: en el entorno es un
  emulador, más lento que el real.

Si eso pasa, hay que decirlo en el informe: **el punto de quiebre medido es
el del entorno, no el de producción.**

## Fuentes

- Clase de herramientas del ecosistema de testing: los cinco pasos (requisitos, mercado, PoC, matriz de
  decisión con pesos, despliegue incremental), JIRA y Zephyr/Xray, panorama
  y mapa del ecosistema, carga, estrés y escalabilidad, SonarQube y el
  quality gate.
- Consigna del TP: aspectos "Herramienta nueva" y "Stress y esfuerzo
  sostenido" del Nivel 3.
- Repo: `.github/workflows/pipeline.yml`, `education/go.mod`,
  `education/load_test.js`.
- **Aportes externos:** precios de SonarQube Cloud, límite de pedidos de
  Contentful y la elección de k6. Están marcados en el texto donde aparecen.
