# Eje 4. Métricas y análisis estático (Educación)

**Qué pide la consigna:** análisis estático, tabla de métricas interpretadas, una métrica de proceso y una de proyecto, y una mejora medida antes y después.

## 1. Qué es el análisis estático

El examen del código fuente **sin ejecutar el programa**. Una herramienta lee el código, arma un mapa de su estructura y le aplica reglas.

## 2. Cómo está armado

Se analiza todo el codigo Go de education/

| Pieza | Dónde | Qué hace |
|---|---|---|
| Ficha del proyecto | `education/sonar-project.properties` | Le dice a SonarQube qué es código, qué son pruebas y dónde está la cobertura |
| Configuración | `scripts/calidad/sonar/configurar_sonar.sh` | Crea el proyecto y el quality gate |
| Punto de partida | `scripts/calidad/sonar/linea_base.sh` | Registra cómo estaba el código antes de agosto, como referencia para el "código nuevo" |
| Análisis | `scripts/calidad/sonar/analizar.sh` | Pruebas → cobertura → SonarQube → veredicto → historial |
| Automatización | `.github/workflows/calidad-tp.yml` | Lo mismo en cada cambio de `education/`, más el análisis de seguridad |

**"Código nuevo"** = lo cambiado desde el inicio de agosto de 2026.

## 3. Quality gate "TP-Calidad-N3"

Las metas del Nivel 3 de la consigna:

| Condición | Umbral | Resultado | Estado |
|---|---|---|---|
| Confiabilidad (bugs), todo el código | No peor que C (= ningún bug de severidad alta o bloqueante) | A | Cumple |
| Confiabilidad, código nuevo | Ídem | A | Cumple |
| Cobertura, todo el código | ≥ 80 % | **92,5 %** | Cumple |
| Cobertura, código nuevo | ≥ 80 % | **92,5 %** | Cumple |
| Duplicación, todo el código | ≤ 5 % | 0,2 % | Cumple |
| Duplicación, código nuevo | ≤ 5 % | 0,1 % | Cumple |

## 4. Tabla de métricas interpretadas

Cada métrica lleva su tipo, según la clasificación de esa clase: producto, proceso o proyecto.

| Métrica | Tipo | Valor | Qué significa | Dónde se concentra | Consecuencia |
|---|---|---|---|---|---|
| Tamaño (líneas de código sin comentarios ni blancos) | Producto | 8.110 en 92 archivos | Un servicio mediano | Los archivos más grandes son `submission_service.go` (384 líneas) y `contentful_client.go` (252) | La clase advierte que esta métrica "sola es poco confiable": sirve para ubicar dónde mirar, no para juzgar |
| Complejidad ciclomática | Producto | 1.536 en total, unos 3,4 caminos por función | La mayoría de las funciones son chicas | `submission_service.go` (89), `skill_service.go` (64), `contentful_client.go` (61) | "Cuanto mayor la complejidad, más casos de prueba se necesitan". Las más complejas tienen buena cobertura (92 %, 97 %, 88 %): el riesgo está cubierto |
| Complejidad cognitiva | Producto | 14 funciones superan el límite de 15: 11 del servicio y 3 de pruebas | Mide cuánto cuesta *entender* una función: penaliza el anidamiento (if dentro de if). | Las peores: `submission_repository.go:58` (24) y `general_attendance_repository.go:63` (23). En pruebas: `contentful_client_helpers_test.go:52` (28). La que estaba peor, `skill_repository.go:150` (38), se mejoró (sección 7) | Funciones difíciles de leer = más fácil equivocarse al modificarlas. Son 14 de los 19 problemas "altos" de mantenibilidad |
| Duplicación | Producto | 0,2 % (3 bloques) | Casi no hay código copiado | En las pruebas de `course_controller_test.go` hay 4 funciones idénticas entre sí | Hoy no es un problema |
| Cobertura de líneas | Producto | 92,5 %, sobre 4.299 líneas con pruebas posibles | 9 de cada 10 líneas se ejecutan en alguna prueba | Lo que más falta: `cmd/server/main.go` (21,6 %, 105 líneas sin probar), `enrollment_repository.go` (61 %), `file_storage.go` (6,3 %) y `outbox/pubsub.go` (0 %, copia generada) | Lo que queda sin probar es el arranque y las conexiones con servicios externos. |
| Bugs (confiabilidad) | Producto | 0 | SonarQube no encontró errores probables | — | Buena señal |
| Vulnerabilidades y puntos sensibles | Producto | 0 y 0 | Ningún patrón de seguridad riesgoso conocido | — | Se complementa con gosec, govulncheck, Gitleaks, Semgrep y Trivy, que corren en el mismo pipeline de calidad |
| Problemas de mantenibilidad (code smells) | Producto | 35: 19 altos, 4 medios, 12 informativos | Código que funciona pero cuesta mantener | 14 de complejidad cognitiva, 5 textos repetidos (por ejemplo `"/entries/"` 5 veces), 4 funciones idénticas en pruebas y 12 comentarios `TODO` pendientes | Ver la sección 6 |
| Deuda técnica | Producto | 243 minutos (unas 4 horas), 0,1 % del código | Lo que SonarQube estima que lleva arreglar todos los problemas de mantenibilidad | La mayor parte está en las funciones complejas | Calificación A: la deuda es baja para el tamaño |
| Densidad de comentarios | Producto | 18,2 % | Casi 1 de cada 5 líneas es comentario | — | Ayuda a la mantenibilidad: el código explica sus decisiones |
| Densidad de defectos | Producto | Sin dato todavía | Defectos cada 1.000 líneas | — | - |

## 5. Una métrica de proceso y una de proyecto

### Métrica de proceso: frecuencia de integración

La clase da como ejemplo la "frecuencia de integración".

**Qué mide:** cuántos commits por semana tocaron `education/` desde agosto. Sale de `git log --no-merges --since=2026-08-01 -- education`.

| Semana (ISO) | 31 | 32 | 33 | 34 | 35 | 36 | 37 | 38 | 39 | 40 | 41 | 42 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| Commits | 0 | 0 | 0 | 3 | 2 | 1 | 4 | 2 | 1 | 1 | 1 | 4 |

- **Total:** 19 commits en 12 semanas, 1,6 por semana.
- **Interpretación:** el trabajo llega en lotes (semanas de 3 o 4 commits y otras de 0 o 1). Las primeras 3 semanas de agosto no hubo commits. Cuanto más grande el lote, más cosas cambian juntas y más difícil es encontrar qué rompió algo.
- **Meta propuesta:** al menos 2 integraciones por semana, cada una con análisis de SonarQube (lo garantiza `calidad-tp.yml`).

### Métrica de proyecto: desvío del cronograma

Preguntar profe

## 6. Hallazgos y prioridades

| # | Dónde | Problema | Por qué importa | Mejora posible | Prioridad |
|---|---|---|---|---|---|
| 1 | `internal/repositories/skill_repository.go:150` | Complejidad cognitiva 38, más del doble del límite (15) | Es la función que guarda las habilidades de una actividad dentro de una transacción. Si se rompe, se pierden o duplican habilidades | Separarla en funciones chicas | **Resuelto** (sección 7) |
| 2 | `cmd/server/main.go` (21,6 %), `enrollment_repository.go` (61 %), `file_storage.go` (6,3 %) | Las partes con menos cobertura | Un error de arranque o al guardar archivos recién se ve al desplegar | Más pruebas de arranque y de archivos. | **Media** |
| 2b | 3 archivos de pruebas nuevos (`contentful_client_helpers_test.go:52`, `material_repository_test.go:183`, `material_repository_upload_test.go:131`) | Complejidad cognitiva 28, 21 y 22 | Una prueba difícil de leer cuesta mantenerla y, si falla, cuesta entender por qué | Partirlas en casos más chicos (pruebas por tabla) | **Baja** |
| 3 | `contentful_client.go:199`, `docente_controller.go:81` y 3 más | Textos repetidos en lugar de una constante | Si cambia la ruta de Contentful o un mensaje, hay que acordarse de todos los lugares | Una constante por texto | **Baja**: fácil y de poco riesgo |
| 4 | 12 comentarios `TODO` | Tareas anotadas en el código y nunca hechas | Pueden esconder funcionalidad incompleta | Revisarlos uno por uno: hacerlo, pasarlo a Jira o borrarlo | **Media** |

## 7. Mejora medida antes y después

**Qué se hizo:** se redujo la complejidad cognitiva de `skill_repository.go:150` (hallazgo 1) sin cambiar lo que hace.

**La función, en simple:** cuando un docente guarda las habilidades de una actividad, esta función, en una sola operación de la base de datos:

- crea en el catálogo las habilidades que todavía no existen;
- vincula todas a la actividad;
- borra los vínculos de las habilidades que se desmarcaron.

Antes era una sola receta larga, con condiciones dentro de condiciones. Ahora `SetActivitySkills` llama a cuatro pasos con nombre propio:

1. `readActivitySkills`: lee lo que hay;
2. `writeActivitySkills`: guarda y vincula;
3. `ensureInCatalog`: usa la habilidad existente o crea la nueva;
4. `deleteStaleLinks`: borra lo desmarcado.

| Indicador | Antes | Meta | Después | Estado |
|---|---|---|---|---|
| Complejidad cognitiva de la función | 38 | ≤ 15 | Por debajo de 15: SonarQube ya no la marca | Cumple |
| Complejidad cognitiva de todo el archivo | 73 | Bajar | 60 | Cumple |
| Funciones marcadas por complejidad (regla `go:S3776`) | 15 | 14 | 14 | Cumple |
| Problemas de mantenibilidad altos | 20 | 19 | 19 | Cumple |
| Problemas de mantenibilidad, total | 36 | 35 | 35 | Cumple |
| Deuda técnica | 271 min | Bajar | 243 min (28 min menos) | Cumple |
| Pruebas de `skill_repository` | En verde | Siguen en verde, sin cambiarlas | En verde, sin cambiar ninguna prueba | Cumple |
| Todas las pruebas de Educación | 885 de 886 en Windows | Igual | 885 de 886: falla la misma de antes, que solo falla en Windows | Cumple |
| Quality gate | Aprobado | Aprobado | Aprobado | Cumple |
| Cobertura | 92,6 % | ≥ 80 % | 92,5 % | Cumple (ver nota) |

**Nota sobre la cobertura:** bajó una décima porque, al partir la función, el código tiene 11 líneas más que cuentan para la cobertura (las firmas de las funciones nuevas y sus `return`). Los caminos de error que no se pueden provocar en el emulador quedaron sin probar, igual que antes. No se prueba menos que antes.

**Cómo se comprueba que no cambió nada:** las pruebas de `skill_repository_test.go` ya cubrían los cuatro casos importantes, y pasaron antes y después, sin tocarlas:

- respetar una habilidad que ya existe con su nombre original;
- crear una nueva;
- quitar una desmarcada;
- dejar la actividad sin habilidades.

**Por qué esta mejora:** era el peor valor del proyecto, está dentro del alcance y ya tenía pruebas que protegían el cambio. Si las pruebas siguen en verde, el comportamiento no cambió.

## 8. Limitaciones (análisis crítico)

- **El margen del gate depende de las pruebas**
- **SonarQube encuentra lo que sus reglas conocen.** "0 bugs" no quiere decir "sin defectos".
- **La cobertura cuenta líneas ejecutadas, no verificadas.** Una prueba puede pasar por una línea sin comprobar su resultado. El pipeline de la tesis tiene pruebas de mutación (gremlins) que miden eso, y corresponden al eje de pruebas. El 91,36 % de mutación salió de una corrida local; la corrida local del 24/10 (commit `0aa6b29`) dio 97,28 % (500 mutantes eliminados y 14 sobrevivientes; los 56 que excedieron el tiempo no cuentan). El CI del TP ya corre la mutación a pedido (job `mutacion-education`).
- **Un test falla en Windows**.
- **Las métricas de cobertura de Go y de SonarQube difieren:** Go mide la cobertura por sentencias y SonarQube por líneas, así que dan números apenas distintos (por ejemplo, 92,2 % y 92,6 % sobre el mismo código). Las dos son correctas: cuentan cosas distintas. En el TP se usa la de SonarQube, porque es la que evalúa el gate.

## 9. Medición actualizada (código actual de la rama)

La rama incorporó trabajo nuevo en Educación (baja lógica de cursos,
auditoría, historial de asistencia, buscador de participantes). Antes de
medir hubo que actualizar 4 pruebas que quedaron desactualizadas con ese
cambio (DEF-07 en Jira, BVE-41). La medición se hizo con `analizar.sh`,
con el emulador de Firestore y todas las pruebas en verde. Queda en
`historial.csv` como `eb22c0e-modificado` (incluye las pruebas corregidas,
todavía sin commitear).

**Quality gate TP-Calidad-N3: aprobado.**

| Condición | Umbral | Resultado | Estado |
|---|---|---|---|
| Confiabilidad, todo el código | No peor que C | A | Cumple |
| Confiabilidad, código nuevo | No peor que C | A | Cumple |
| Cobertura, todo el código | ≥ 80 % | 90,7 % | Cumple |
| Cobertura, código nuevo | ≥ 80 % | 90,7 % | Cumple |
| Duplicación, todo el código | ≤ 5 % | 0,2 % | Cumple |
| Duplicación, código nuevo | ≤ 5 % | 0,08 % | Cumple |

| Métrica | Antes (`e3357a6`) | Ahora | Qué cambió |
|---|---|---|---|
| Tamaño | 8.119 líneas | 9.905 líneas en 113 archivos, 555 funciones | +1.786 líneas de funcionalidad nueva |
| Pruebas | 891 (1 fallaba en Windows) | 959, todas en verde | +68 pruebas; la de Windows ya pasa |
| Cobertura de líneas | 92,4 % | 90,7 % (486 de 5.218 líneas sin probar) | Bajó 1,7 puntos: el código nuevo tiene menos pruebas que el anterior. Sigue arriba del 80 % |
| Bugs y vulnerabilidades | 0 y 0 | 0 y 0 | — |
| Complejidad ciclomática | 1.538 | 1.898 | Crece con el tamaño |
| Funciones con complejidad cognitiva > 15 | 14 (11 del servicio y 3 de pruebas) | 19 (15 del servicio y 4 de pruebas) | 5 funciones nuevas por encima del límite |
| Problemas de mantenibilidad | 35 (19 altos) | 41 (24 altos, 5 medios, 12 informativos) | 19 de complejidad cognitiva, 12 comentarios TODO, 5 textos repetidos, 5 funciones idénticas |
| Deuda técnica | 243 min | 315 min (0,1 % del código) | Calificación A |
| Densidad de comentarios | 18,2 % | 17,5 % | — |

**Lectura:** el gate sigue aprobado con margen, pero la tendencia de
mantenibilidad empeoró con el código nuevo (más funciones complejas, más
deuda). Es la consecuencia de agregar funcionalidad sin aplicar en el
momento el mismo cuidado que en la mejora medida de la sección 7: el plan
de OBJ-E3 (eje 1) prioriza las funciones con mayor complejidad.
