# TP Integrador de Calidad de Software: visión general

Tesis: Sistema Buen Vivir (Centro de Ecología Integral, Alta Gracia).
Materia: Calidad de Software, Mag. Pablo Virgolini.
Integrantes: soficuozzo y Leonardo Morabito.
Nivel declarado: **Nivel 3, prototipo avanzado**.

| Fecha | Qué |
|---|---|
| 19/10/2026 | Versión revisable para el profesor (fecha interna) |
| 26/10/2026 | Entrega oficial y defensa |

Las decisiones tomadas en el camino están en [decisiones.md](decisiones.md).

---

## 1. Qué es este trabajo, en pocas palabras

La consigna pide aplicar a la tesis **un proceso de calidad completo, que
acompañe el desarrollo y no se agregue al final**. Al terminar tenemos que
poder demostrar tres cosas (consigna, sección "Contexto"):

1. El sistema **se prueba solo** cada vez que se despliega.
2. Las **integraciones externas** funcionan o están controladas.
3. Las **decisiones de calidad se apoyan en datos**, no en opiniones.

**Alcance.** No analizamos todo el sistema, solo **un microservicio**
(decisión D-19):

| Microservicio | Carpeta | Qué hace | Tecnología |
|---|---|---|---|
| **Educación** | `education/` | Cursos, materiales, inscripciones, asistencia, entregas, habilidades | Go 1.26 |

**Integraciones externas reales** (cosas que no controlamos y pueden fallar):

- Contentful: guarda los cursos. Es **la plataforma externa principal del TP**.
- Firestore.
- Google Cloud Storage.
- auth-service.
- Pub/Sub e IA (Gemini/OpenAI), ambas opcionales.

---

## 2. Cómo nos dividimos el trabajo

La consigna pide **un eje por tema de clase, más un eje transversal de
documentación**. La nota individual puede variar, así que cada aporte tiene
que ser **trazable** (se tiene que ver quién hizo qué).

| # | Eje | Implementa | Documenta |
|---|---|---|---|
| 1 | Introducción a la calidad | **soficuozzo** | soficuozzo |
| 2 | Fundamentos del testing | **Leonardo** | Leonardo (soficuozzo hace la guía conceptual) |
| 3 | Modelos de calidad y niveles de prueba | **soficuozzo** | soficuozzo |
| 4 | Métricas y análisis estático | **soficuozzo** | soficuozzo |
| 5 | Pruebas dinámicas | **Leonardo** | Leonardo (con guía conceptual) |
| 6 | Aseguramiento de la calidad | **Leonardo** | Leonardo (con guía conceptual) |
| 7 | Automatización del testing | **Leonardo** | Leonardo (con guía conceptual) |
| 8 | Herramientas del ecosistema | **soficuozzo** (stress: plan de soficuozzo, scripts acordados con Leonardo) | soficuozzo |
| 9 | Documentación y análisis transversal | **soficuozzo** | soficuozzo |

**Regla de convivencia:** soficuozzo no escribe pruebas, scripts ni
pipelines de los ejes de Leonardo. Solo los explica en
`guia-conceptual-ejes-del-companero.md` y los integra cuando existan (ver
el punto 6).

---

## 3. Los nueve ejes: qué pide la consigna y qué vamos a hacer

Cada eje tiene la misma estructura:

- **Qué pide:** el "resumen guía" de la consigna.
- **En simple:** explicación sin tecnicismos.
- **Qué hay hoy:** lo que encontramos en el diagnóstico del 17/10.
- **Qué vamos a hacer** y **qué evidencia vamos a mostrar**.

### Eje 1. Introducción a la calidad (soficuozzo)

**Qué pide:** objetivos de calidad medibles desde el usuario, el equipo y la
organización, con las prácticas de proceso que los sostienen, y un plan de
contingencia ante la caída de una dependencia crítica.

**En simple:** "calidad" significa cosas distintas según quién mire. Al
usuario le importa que funcione y sea fácil. Al equipo, que el código se
pueda mantener. A la organización, que no se caiga y que proteja los datos.
Esta idea es de la clase del 03/08 ("La calidad cambia según el observador").
Para cada mirada definimos metas con número, por ejemplo "la lista de
cursos carga en menos de X segundos". También preparamos un plan de qué
hacer si se cae algo de lo que dependemos, por ejemplo Contentful o
Firestore.

**Qué hay hoy:**
- No hay objetivos escritos.
- Ya sabemos qué pasa si se cae Contentful: Educación sigue mostrando la
  última versión guardada (caché, ADR 0006).

**Qué vamos a hacer:**
- Una tabla de objetivos (ID, perspectiva, indicador, fórmula, línea base,
  meta, cómo y cada cuánto se mide, evidencia, responsable).
- Un plan de contingencia por dependencia, separando tres cosas: lo que el
  sistema hace hoy, lo que proponemos y lo que realmente implementamos.

**Evidencia:** `01-introduccion-a-la-calidad.md`, con valores medidos.

### Eje 2. Fundamentos del testing (Leonardo)

**Qué pide:** pruebas unitarias en Arrange–Act–Assert, criterio de
aceptación fijado antes de ejecutar, un defecto real con contraste
rojo/verde y los siete principios aplicados al proyecto.

**En simple:**
- Cada prueba tiene tres partes: preparar, ejecutar y verificar (clase del
  10/08).
- El resultado esperado se decide **antes** de correr la prueba.
- Hay que mostrar un error real del sistema: una prueba que da **rojo**
  (encuentra el error) y, después del arreglo, da **verde**.

**Qué hay hoy:** Educación tiene muchas pruebas: 886 al 18/10 (eran 466
antes del commit `42414bc` de Leonardo), que cubren el 92,6 % de las
líneas según SonarQube.

**soficuozzo aporta:** la explicación en la guía conceptual y el control de
que la evidencia rojo/verde sea real.

### Eje 3. Modelos de calidad y niveles de prueba (soficuozzo)

**Qué pide:** matriz de las 8 características de ISO/IEC 25010 con
evidencia, top 3 justificado, mapa de niveles de prueba y despliegue
verificado en cada instancia.

**En simple:** ISO 25010 es una "lista de control" internacional que divide
la calidad en 8 aspectos: que haga lo que debe, que sea rápido, que conviva
con otros sistemas, que sea fácil de usar, que no falle, que sea seguro,
que se pueda mantener y que se pueda llevar a otro entorno (clase del
24/08). Para cada aspecto anotamos cómo se aplica a Educación, con qué lo
medimos y qué evidencia tenemos. Después elegimos los 3 más importantes y
explicamos por qué.

**Qué hay hoy:** nada escrito. Ya hay evidencia para varias filas: el
análisis de seguridad del pipeline, las pruebas, SonarQube y el test que
falla solo en Windows (portabilidad).
**Qué vamos a hacer:**
- La matriz.
- El top 3 justificado con riesgos y datos.
- El mapa de niveles de prueba (unitarias → integración → sistema/API →
  aceptación/E2E): qué existe, quién lo hace y qué falta.
- El procedimiento para comprobar que lo desplegado es lo mismo que se
  analizó. Esto es justamente lo que falló en el caso Knight Capital, visto
  en clase.

**Evidencia:** `02-modelos-de-calidad-y-niveles.md`.

### Eje 4. Métricas y análisis estático (soficuozzo)

**Qué pide:** análisis estático, tabla de métricas interpretadas, una
métrica de proceso y una de proyecto, y una mejora medida antes y después.

**En simple:** el análisis estático revisa el código **sin ejecutarlo**,
como un corrector ortográfico pero para programas (clase del 31/08). Lo
hace SonarQube, que da números como errores probables, código repetido,
complejidad y porcentaje probado. No alcanza con copiar los números: hay
que explicar qué significan. Después elegimos un problema real, lo
mejoramos y mostramos el número de antes y el de después.

**Qué hay hoy (18/10):**
- SonarQube 26.8 configurado para Educación, con el quality gate
  "TP-Calidad-N3" (D-18) y el código nuevo desde el 1/8 (D-17).
- Dos mediciones hechas, las dos con el quality gate en verde:
  - 17/10: 8.088 líneas de código, 0 bugs, 0 vulnerabilidades, 33 problemas
    de mantenibilidad (17 de severidad alta), 80,4 % de cobertura y 0,2 % de
    duplicación.
  - 18/10, con las pruebas nuevas de Leonardo: 92,6 % de cobertura, 886
    pruebas y 36 problemas de mantenibilidad (20 altos); el resto, igual.

**Qué vamos a hacer:**
- La tabla de métricas interpretadas.
- Métrica de proceso: por ejemplo, el tiempo para corregir un defecto o la
  frecuencia de análisis.
- Métrica de proyecto: por ejemplo, la cobertura o la duplicación.
- Una mejora medida.
- Un historial automático por commit (`reportes/historial.csv`).

**Evidencia:** `03-metricas-y-analisis-estatico.md`, capturas de SonarQube
y el historial.

### Eje 5. Pruebas dinámicas (Leonardo)

**Qué pide:** casos de prueba documentados con clases de equivalencia,
valores límite y tablas de decisión; cobertura de sentencia y de decisión.

**En simple:** en vez de probar al azar, se prueba con método (clase del
07/09). Por ejemplo, el cupo de un curso:

- Se agrupan los valores que se comportan igual (clases de equivalencia).
- Se prueban los bordes (valores límite): cupo lleno, cupo lleno menos uno.
- Se combinan condiciones (tablas de decisión).

**soficuozzo aporta:** la guía conceptual y la medición de la cobertura de
decisión (ramas) en SonarQube, que es parte del eje 4.

### Eje 6. Aseguramiento de la calidad (Leonardo)

**Qué pide:** plan de pruebas, riesgos (probabilidad × impacto), prácticas
de QA distinguidas de las de QC y plantillas de caso y de defecto con
trazabilidad.

**En simple:**
- **QA** mejora *cómo* se trabaja, para prevenir.
- **QC** revisa *lo que se hizo*, para detectar (clase del 14/09).
- El plan de pruebas dice qué se prueba, cuándo y con qué criterio.
- Los riesgos se ordenan por probabilidad × impacto.
- Cada requisito tiene que poder seguirse hasta su prueba y su defecto
  (trazabilidad).

**soficuozzo aporta:** la guía conceptual. Los objetivos del eje 1 y la
matriz ISO del eje 3 alimentan los riesgos del plan.

**Evidencia (23/10):** `05-plan-de-pruebas.md` (plan versionado, riesgos de
producto y de proyecto, trazabilidad) y el proyecto BVE de Jira.

### Eje 7. Automatización del testing (Leonardo)

**Qué pide:** pirámide de pruebas automatizadas, pruebas de API e
integración con el MCP o la plataforma externa, smoke en cada despliegue,
pipeline de CI/CD y una prueba de regresión por cada defecto corregido.

**En simple:** que las pruebas corran solas en cada cambio (clase del
21/09). La pirámide son muchas pruebas chicas y rápidas abajo y pocas
grandes arriba. El "smoke" es una prueba rápida de "¿está vivo?" después de
desplegar. Cada error corregido deja una prueba para que no vuelva.

**Qué hay hoy:**
- El pipeline del proyecto ya corre las pruebas de Educación.
- Hay smoke de Educación.
- No hay pruebas automáticas contra el Contentful real.

**soficuozzo aporta:** la guía conceptual y el pipeline propio del TP
(`calidad-tp.yml`), preparado para sumar las pruebas de Leonardo cuando
estén.

### Eje 8. Herramientas del ecosistema (soficuozzo)

**Qué pide:** mapa de herramientas, una herramienta nueva incorporada con
el proceso de cinco pasos, registro en Jira u otra, y pruebas de stress y
de esfuerzo sostenido.

**En simple:**
- **Mapa de herramientas:** qué herramienta se usa para qué (Git, Docker,
  pruebas, seguridad, etc.).
- **Herramienta nueva:** **SonarQube**. Se incorpora con los 5 pasos de la
  clase del 28/09: requisitos → mercado (comparar opciones) → prueba de
  concepto → matriz de decisión con puntajes → despliegue de a poco.
- **Jira:** registrar las tareas y los defectos de calidad.
- **Stress:** "apretar" el sistema con muchos usuarios a la vez para ver
  dónde se rompe.
- **Esfuerzo sostenido:** carga normal durante mucho tiempo, para ver si se
  degrada.

**Qué hay hoy:**
- Hay un script de carga de Educación (k6) muy liviano.
- El reporte de k6 que está en `reports/performance/` **es inventado por un
  script**, así que no sirve como evidencia.
- Jira existe (proyecto SCRUM "Testing"), pero solo tiene ejercicios de
  clase. **Actualización 23/10:** se creó el proyecto BVE con historias,
  casos, defectos y tareas (ver `04-herramientas-del-ecosistema.md`).

**Qué vamos a hacer:**
- El mapa de herramientas.
- Los 5 pasos de SonarQube.
- El plan de stress y esfuerzo sostenido: escenarios, cantidad de usuarios,
  duración, límites y cuándo cortar.
- Resultados **solo si se ejecutan de verdad**.
- Jira al final.

**Evidencia:** `04-herramientas-del-ecosistema.md`.

### Eje 9. Documentación y análisis transversal (soficuozzo)

**Qué pide:** integrar todo lo anterior.

**Qué vamos a hacer:**
- Esta visión general.
- El registro de decisiones.
- La guía de calidad.
- La matriz de cumplimiento de la consigna.
- La guía conceptual de los ejes de Leonardo.
- El informe en PDF para el profesor.

**Evidencia (23/10):** `06-informe-final.md` (nivel, metas, evolución,
análisis crítico y conclusiones).

---

## 4. Metas del Nivel 3 y dónde estamos

Tabla de "Metas de calidad" de la consigna, columna Nivel 3, con el estado
medido el 17/10/2026.

| Aspecto | Nivel 3 exige | Hoy | Responsable |
|---|---|---|---|
| Pipeline | Compilación, pruebas, análisis estático, cobertura, quality gate y reporte automático | Hay compilación y pruebas. El pipeline del TP (`calidad-tp.yml`) corre en GitHub en cada cambio de `education/`; según soficuozzo, la corrida del 18/10 dio verde | soficuozzo (análisis) y Leonardo (pruebas) |
| Quality gate | Bloquea ante bugs críticos, duplicación > 5 % y cobertura de código nuevo < 80 % | Creado ("TP-Calidad-N3"). En verde en las dos mediciones (17/10 y 18/10) | soficuozzo |
| Cobertura de código nuevo | 80 % | 92,6 % (desde el 1/8, SonarQube, 18/10; el 17/10 era 80,4 %) | Medición: soficuozzo. Pruebas: Leonardo |
| Cobertura de código existente | 80 % | 92,6 % (SonarQube, 18/10; el 17/10 era 80,4 %). Un test falla solo en Windows | Medición: soficuozzo. Pruebas: Leonardo |
| Plan de pruebas y riesgos | Versionado, con criterios de entrada y salida, riesgos y trazabilidad | No existe | Leonardo |
| Defectos | Con pasos, esperado y obtenido, vinculados al caso en Jira u otra | No existe | Leonardo (Jira: ambos) |
| Herramienta nueva | Cinco pasos | En curso: SonarQube | soficuozzo |
| Smoke | Ejecutado por el pipeline tras el despliegue | Solo Educación | Leonardo |
| Integración externa | Flujo E2E automatizado en cada despliegue | No existe | Leonardo (entorno: soficuozzo) |
| Stress y esfuerzo sostenido | Ejecutado, con resultados y umbrales | No hay resultados reales | Plan: soficuozzo. Scripts: a acordar con Leonardo |

> Ningún aspecto se va a marcar como cumplido sin evidencia de una
> ejecución real.

---

## 5. Entregables de la consigna

| Entregable (consigna, sección "Entregables") | Quién | Dónde |
|---|---|---|
| Plan de pruebas versionado con matriz ISO 25010 y trazabilidad | Leonardo (plan) y soficuozzo (matriz ISO) | `docs/calidad/` |
| Código de pruebas unitarias, integración, smoke, stress y esfuerzo, ejecutables con un comando | Leonardo (stress: a acordar) | Carpeta de cada servicio |
| Pipeline en el repo con al menos una ejecución en verde y una fallida que detectó un problema real | Ambos (pipeline del TP: soficuozzo) | `.github/workflows/calidad-tp.yml` |
| Reportes automáticos de cobertura, análisis estático y stress | soficuozzo (cobertura y estático), stress a acordar | `docs/calidad/reportes/` y artefactos del pipeline |
| Documentación de la integración externa | Leonardo (pruebas) y soficuozzo (entorno y contingencia) | `docs/calidad/` |
| Registro de defectos completo | Leonardo | Jira |
| Documento de la herramienta nueva con matriz de decisión | soficuozzo | `04-herramientas-del-ecosistema.md` |
| Informe en PDF: nivel, metas cumplidas y no cumplidas, evolución, análisis crítico y conclusiones | soficuozzo (sus ejes y la integración) | `docs/calidad/informe/` |
| Demostración con un despliegue nuevo | Ambos | Entorno de calidad (`docker-compose.calidad.yml`) |

---

## 6. Cómo se conectan las dos partes

```
  Leonardo                                soficuozzo
  ─────────                               ──────────
  pruebas (unit, API, smoke, E2E)  ──►  cobertura importada en SonarQube
  scripts de stress                ──►  plan, umbrales e interpretación
  plan de pruebas y riesgos        ◄──  objetivos (eje 1) y matriz ISO (eje 3)
  defectos en Jira                 ──►  métrica de proceso (tiempo de corrección)
```

**Contrato de integración:** Leonardo deja los resultados en archivos con
formato estándar, y el pipeline del TP los toma sin reescribir nada:

- `education/coverage.out` (cobertura de Go).
- El resumen JSON de k6.

---

## 7. Plan de trabajo hasta el 19/10

| Día | Fase | Qué se hace |
|---|---|---|
| Sáb 17/10 | 1. SonarQube | Recuperar el acceso a SonarQube, línea base de Educación, quality gate "TP-Calidad-N3" (hecho) |
| Dom 18/10 | 2. Entorno y pipeline | `docker-compose.calidad.yml` verificado, `calidad-tp.yml`, mejora antes y después, plan de stress |
| Lun 19/10 | 3. Documentos | Ejes 1, 3, 4 y 8, guía conceptual, matriz de cumplimiento e informe PDF |
| 20 al 26/10 | 4. Integración | Sumar lo de Leonardo, Jira, stress real, ensayo de la defensa |

## 8. Mapa de documentos

| Archivo | Contenido | Estado |
|---|---|---|
| `00-vision-general.md` | Este documento | Hecho |
| `decisiones.md` | Registro de decisiones | Hecho (se actualiza) |
| `01-introduccion-a-la-calidad.md` | Eje 1 | Borrador (18/10) |
| `02-modelos-de-calidad-y-niveles.md` | Eje 3 | Borrador (18/10) |
| `03-metricas-y-analisis-estatico.md` | Eje 4 | Borrador (18/10) |
| `04-herramientas-del-ecosistema.md` | Eje 8 | Borrador (18/10) |
| `guia-de-calidad.md` | Guía general y matriz de cumplimiento | Pendiente |
| `guia-conceptual-ejes-del-companero.md` | Ejes 2, 5, 6 y 7 explicados | Pendiente |
| `guia-scripts-y-resumen-para-leonardo.md` | Qué hace cada script, dónde quedan los resultados y qué contarle a Leonardo | Hecho (18/10) |
| `reportes/` | Historial de métricas por commit | Se genera con `scripts/calidad/analizar.sh` |
| `informe/` | Informe fuente y PDF | Pendiente |

## 9. Fuentes

- Consigna: `TP Integrador - Calidad de Software aplicada al TFG.pdf`.
- Guía de instalación de SonarQube: `Guia_SonarQube.pdf`.
- Clases de la carpeta `Clases/`: 03/08, 10/08, 24/08, 31/08, 07/09,
  14/09, 21/09, 28/09 y 05/10.

Cuando se usa algo que **no** está en el material de clase, se aclara como
"aporte externo".
