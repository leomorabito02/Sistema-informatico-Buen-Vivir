# Informe final (transversal: documentación y análisis)

**Qué pide la consigna:** el nivel declarado, las metas cumplidas y no
cumplidas, la evolución de las métricas, el análisis crítico y las
conclusiones. Los números tienen que coincidir con los del pipeline y la
herramienta de análisis.

> **Estado de esta versión:** las cifras de SonarQube son de la última
> medición registrada (commit `e3357a6`). La rama incorporó después trabajo
> nuevo en Educación, así que hay que volver a medir antes de entregar y
> actualizar este informe con esa medición. Ningún número de este documento
> está estimado: todos salen de `docs/calidad/reportes/historial.csv`, de
> SonarQube o de las pruebas.

## 1. Nivel declarado

**Nivel 3: prototipo avanzado.** Educación tiene un pipeline que lo
construye, lo prueba y lo despliega en Cloud Run en dos instancias (testing
y producción), con smoke y vuelta atrás automática si algo falla, y un
pipeline propio del TP con SonarQube y análisis de seguridad. El alcance del
TP es solo el microservicio de Educación.

## 2. Metas cumplidas y no cumplidas

| Aspecto (consigna) | Meta del Nivel 3 | Resultado | Evidencia | Estado |
|---|---|---|---|---|
| Pipeline | Compilación, pruebas, análisis estático, cobertura, quality gate y reporte automático | Todo eso corre en `calidad-tp.yml` en cada cambio de `education/` | `.github/workflows/calidad-tp.yml` | Cumple |
| Quality gate | Bloquea ante bugs críticos, duplicación > 5 % y cobertura del código nuevo < 80 % | Gate TP-Calidad-N3 con esas condiciones, en todo el código y en el nuevo | `scripts/calidad/configurar_sonar.sh` | Cumple |
| Cobertura del código nuevo | ≥ 80 % | 92,4 % | SonarQube, commit `e3357a6` | Cumple (a confirmar con la nueva medición) |
| Cobertura del código existente | ≥ 80 % | 92,4 % | SonarQube, commit `e3357a6` | Cumple (a confirmar con la nueva medición) |
| Plan de pruebas y riesgos | Plan versionado con criterios de entrada y salida, riesgos por probabilidad × impacto y trazabilidad requisito → caso → defecto | `05-plan-de-pruebas.md`: los cuatro criterios, 5 riesgos de producto y 5 de proyecto, matriz de trazabilidad | `docs/calidad/05-plan-de-pruebas.md` y Jira | Cumple cuando el archivo se suba al repositorio. Faltan 3 casos por cargar en Jira |
| Defectos | Registrados con pasos, esperado y obtenido, vinculados a su caso en Jira | DEF-01 y DEF-06, con todos los datos y su prueba de regresión | Jira BVE-28 y BVE-29 | Cumple |
| Herramienta nueva | Cinco pasos: requisitos, mercado, prueba de concepto, matriz de decisión, despliegue incremental | SonarQube Community Build, elegida con matriz ponderada (8,15 contra 7,35 y 7,25) | `04-herramientas-del-ecosistema.md` | Cumple |
| Smoke test | Ejecutado por el pipeline después del despliegue | Existe en `pipeline.yml` (`cd-smoke`, consulta `/api/salud` de Educación) | `pipeline.yml` | Implementado; la evidencia de una corrida queda fuera del alcance de esta etapa |
| Integración externa | Flujo E2E automatizado en cada despliegue | No existe | — | No cumple (fuera del alcance de esta etapa) |
| Stress y esfuerzo sostenido | Stress ejecutado, con resultados y umbrales | Plan con escenarios y umbrales (ST-0 a ST-2), sin ejecutar | `04-herramientas-del-ecosistema.md`, Jira BVE-31 | No cumple |

**Entregables que no se cumplen en esta etapa:** la ejecución fallida del
pipeline que detecte un problema real (DEF-06 tiene su rojo y verde, pero
locales), los reportes de stress generados por el pipeline y la
demostración con un despliegue nuevo.

## 3. Evolución de las métricas

Cada fila es una medición de SonarQube guardada en `historial.csv`. En
todas falla 1 prueba, la misma: solo falla en Windows por los finales de
línea de un archivo (eje 1, OBJ-E5).

| Medición | Commit | Qué cambió | Líneas | Pruebas | Cobertura | Bugs | Problemas de mantenibilidad | Deuda (min) | Duplicación | Gate |
|---|---|---|---|---|---|---|---|---|---|---|
| 17/10 | `eef7523` | Línea base | 8.088 | 466 | 80,4 % | 0 | 33 | 230 | 0,2 % | Aprobado |
| 19/10 | `4ee25d4` | Pruebas nuevas del eje de testing | 8.088 | 886 | 92,6 % | 0 | 36 | 271 | 0,2 % | Aprobado |
| 19/10 | `4ee25d4` + mejora | Mejora medida del eje 4 (`skill_repository.go`) | 8.110 | 886 | 92,5 % | 0 | 35 | 243 | 0,2 % | Aprobado |
| 22/10 | `e3357a6` | Arreglo de DEF-06; SonarQube analizaba también reportes HTML y JSON | 8.589 | 891 | 92,4 % | 1 | 106 | 608 | 0,8 % | Aprobado |
| 22/10 | `e3357a6` + exclusión | Se excluyen los reportes generados del análisis | 8.119 | 891 | 92,4 % | 0 | 35 | 243 | 0,2 % | Aprobado |
| Pendiente | rama actual | Trabajo nuevo incorporado a Educación | — | — | — | — | — | — | — | — |

**Qué muestra la evolución:**

- **La cobertura subió 12 puntos** (80,4 % → 92,4 %) cuando se agregaron las
  pruebas del eje de testing: casi se duplicó la cantidad de pruebas.
- **La mejora del eje 4 bajó la deuda** de 271 a 243 minutos sin tocar
  ninguna prueba.
- **La medición del 22/10 sin exclusión es un falso empeoramiento:**
  SonarQube contaba como código unos reportes generados. Al excluirlos,
  los números volvieron a los anteriores. Es un ejemplo de por qué una
  métrica hay que interpretarla y no solo leerla.

## 4. Análisis crítico

**Lo que funcionó:**

- El quality gate y el historial por commit hicieron que cada número del TP
  se pueda rastrear hasta una corrida.
- Una prueba que ya existía detectó DEF-06, que SonarQube no vio (0 bugs
  antes y después). El análisis estático no reemplaza a las pruebas.
- Jira cerró la trazabilidad requisito → caso → defecto en una sola
  herramienta.

**Lo que no se logró y por qué:**

- **Stress, esfuerzo sostenido y E2E:** quedaron planificados pero no
  ejecutados. Sin ellos no se cumplen dos aspectos del Nivel 3.
- **OBJ-E3 (complejidad):** quedan 14 funciones por encima del límite
  contra una meta de 6. Se priorizó no tocar código que funciona cerca de la
  entrega; hay un plan escrito (eje 1).

**Límites de la evidencia:**

- **Cobertura no es verificación:** cuenta líneas ejecutadas, no
  comprobadas. La mutación (91,36 %) sale de una corrida local, no del
  pipeline.
- **ZAP sin sesión:** el escaneo dinámico solo vio respuestas de rechazo.
  "0 alertas" no prueba que las rutas protegidas sean seguras.
- **Emuladores:** las pruebas de integración corren contra emuladores de
  Firestore y Pub/Sub, que no exigen índices ni permisos como el servicio
  real.
- **Contentful caído (OBJ-U6):** se probó con un servidor que imita a
  Contentful, no cortando el real.
- **Números que envejecen:** al incorporarse trabajo nuevo a la rama, las
  cifras quedaron viejas. Por eso hay que volver a medir antes de entregar.

## 5. Conclusiones

Educación tiene una base de calidad sólida para su tamaño: cobertura de
92,4 %, 0 bugs y 0 vulnerabilidades según SonarQube, deuda baja y un
pipeline que analiza cada cambio con un gate que frena. En el Nivel 3 se
cumplen el pipeline, el quality gate, la cobertura, el plan con riesgos y
trazabilidad, el registro de defectos y la herramienta nueva. No se cumplen
todavía el stress con resultados y el flujo E2E en cada despliegue.

**Recomendaciones para el resto de la tesis:**

1. Ejecutar los escenarios de stress y esfuerzo sostenido y sumarlos al
   pipeline.
2. Agregar el flujo E2E contra Contentful después del smoke.
3. Extender SonarQube y el gate a los demás microservicios (paso 3 del
   despliegue incremental).
4. Seguir el plan de complejidad del eje 1, función por función.
