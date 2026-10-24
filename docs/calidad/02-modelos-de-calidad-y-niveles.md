# Eje 3. Modelos de calidad y niveles de prueba (Educación)

**Qué pide la consigna:** matriz de las 8 características de ISO/IEC 25010 con evidencia, top 3 justificado, mapa de niveles de prueba y despliegue verificado en cada instancia.

## 1. ISO/IEC 25010

**Estado de cada fila:**

- **Con evidencia:** hay evidencia de una ejecución real.
- **Parcial:** hay evidencia parcial, o es solo diseño.
- **Falta:** no hay evidencia.

## 2. Matriz de las 8 características aplicada a Educación

| # | Característica | Pregunta | Cómo aplica a Educación | Cómo se mide / evidencia | Estado |
|---|---|---|---|---|---|
| 1 | **Adecuación funcional** | ¿Hace lo que debe hacer? | Cursos, actividades, inscripciones, asistencia, entregas y habilidades según las historias de usuario | 886 pruebas automatizadas: todas en verde en Linux y 885 en Windows (`historial.csv`). Pruebas de contrato con Pact (`education/tests/contract`) | Con evidencia |
| 2 | **Eficiencia de desempeño** | ¿Responde en tiempos adecuados? | Lecturas de cursos (Contentful con caché de 5 min) y consultas a Firestore | Hay un script de carga liviano (`education/load_test.js`, umbral p95 < 500 ms), pero **no hay resultados reales**: el reporte de `reports/performance/` lo genera un script con valores fijos. Plan de stress en el eje 8 | Falta |
| 3 | **Compatibilidad** | ¿Convive con otros sistemas? | Conversa con Contentful, Firestore, Cloud Storage, auth-service, Pub/Sub y el frontend | Contratos con Pact. Catálogo de eventos AsyncAPI (`docs/eventos/asyncapi.yaml`). En el entorno de calidad, contra Contentful `testing` real, respondió y leyó cursos | Con evidencia |
| 4 | **Usabilidad** (Capacidad de interacción) | ¿Es sencillo de usar? | Educación es un servicio interno sin pantallas propias: la usabilidad se juega en el frontend. Del lado del servicio, mensajes de error claros y en español (por ejemplo, "no tenés acceso a esta entrega") | Solo revisión del código. | Parcial |
| 5 | **Fiabilidad** | ¿Se comporta de forma consistente? | Si Contentful cae, se sirve la última versión guardada (ADR 0006). Los eventos no se pierden: patrón outbox con reintentos (ADR 0004). Los eventos repetidos se descartan. Si el smoke falla, la versión anterior vuelve sola (`cd-rollback`) | SonarQube: 0 bugs, confiabilidad A. Pruebas de la caché (`content_cache_test.go`). | Con evidencia |
| 6 | **Seguridad** | ¿Protege la información? | Datos de participantes de un centro barrial. Cada pedido necesita un token válido y un rol permitido. Los eventos no llevan datos personales (ADR 0007). Una instancia de testing no puede escribir en el Contentful de producción (`validateEnvironmentSafety`) | SonarQube: 0 vulnerabilidades, 0 puntos sensibles. Pipeline de calidad (job `seguridad-education`): govulncheck, 0 vulnerabilidades que el código use; Gitleaks, 2 avisos, los dos con valores de prueba; gosec, 10 avisos (3 altos) para revisar; Semgrep y Trivy, en cada corrida. En el entorno de calidad: sin token responde 401, y un participante pidiendo una ruta de admin recibe 403 | Con evidencia |
| 7 | **Mantenibilidad** | ¿Se puede mantener fácilmente? | 8.110 líneas en 92 archivos, organizadas en capas: controladores, servicios y repositorios | SonarQube: 35 problemas (19 altos), deuda de unas 4 horas, duplicación 0,2 %, cobertura 92,5 %. | Con evidencia |
| 8 | **Portabilidad** (Flexibilidad) | ¿Se instala en distintos entornos? | Corre en Docker, en Cloud Run y en la PC de desarrollo. Una sola imagen para todos los entornos | El mismo `Dockerfile` se usa en `docker-compose.yml`, en `docker-compose.calidad.yml` y en Cloud Run. **Defecto encontrado:** una prueba falla solo en Windows por los finales de línea | Parcial |

### Herramientas de seguridad del pipeline de calidad

Además de SonarQube, el pipeline de calidad (`calidad-tp.yml`) corre cinco herramientas de seguridad sobre Educación. Todas son de análisis estático: leen el código, las librerías o la imagen, sin ejecutar el servicio.

| Herramienta | Qué revisa | Ejemplo de lo que encuentra | ¿Frena el pipeline? |
|---|---|---|---|
| gosec | El código Go, buscando errores de seguridad típicos | Un formulario que acepta archivos de cualquier tamaño, o una dirección web armada con datos que vienen de afuera | No: informa. Sus avisos hay que revisarlos uno por uno, porque tiene falsos positivos |
| govulncheck | Las librerías que usa Educación, contra la base pública de vulnerabilidades de Go | Una librería con una falla conocida que el código realmente usa | Sí |
| Gitleaks | Todos los archivos, buscando claves o contraseñas escritas | Un token de Contentful pegado en un archivo de configuración | No: informa. Hoy marca 2 valores de prueba |
| Semgrep | El código, con reglas de la comunidad para Go | Consultas armadas pegando texto, o criptografía débil | No: informa |
| Trivy | La imagen Docker de Educación: el sistema operativo base y las librerías | Un paquete del sistema operativo con una vulnerabilidad crítica que ya tiene arreglo | Sí, con vulnerabilidades críticas o altas que tengan arreglo (mismo criterio que el pipeline de la tesis) |

**ZAP** es distinto: es una prueba dinámica, porque ataca al servicio mientras corre. En el pipeline de la tesis solo se usa contra auth-service, no contra Educación.

## 3. Top 3 justificado

Lo hacemos con el criterio: riesgo = probabilidad × impacto, con una escala de 1 a 3.

| Puesto | Característica | Probabilidad | Impacto | Riesgo | Por qué |
|---|---|---|---|---|---|
| 1 | **Seguridad** | 2 | 3 | **6** | El servicio guarda datos de personas en situación vulnerable. Que se filtre una inscripción o una entrega afecta a una persona real y a la confianza en el Centro. Hay muchas rutas con permisos distintos (admin, docente, voluntario, participante): un error de rol es fácil de cometer y difícil de ver |
| 2 | **Fiabilidad** | 3 | 2 | **6** | Educación depende de varios servicios externos: Contentful, Firestore y auth-service. Alguno va a fallar en algún momento (probabilidad alta). Como la caché de Contentful y el outbox ya reducen el daño, el impacto es medio |
| 3 | **Mantenibilidad** | 2 | 2 | **4** | Es un microservicio que se va a utilizar y va a crecer a futuro. Hoy está bien (deuda baja, calificación A), pero hay 11 funciones del servicio demasiado complejas (y 3 de pruebas). La cobertura es de 92,5 %. |

## 4. Mapa de niveles de prueba

Lo hago despues con lo de leo

## 5. Despliegue verificado en cada instancia

### Instancias de Educación

| Instancia | Cómo se despliega | Qué la identifica |
|---|---|---|
| Producción (`education-prod`) | `pipeline.yml`, en push a `main` | Imagen `education:<commit>`, `APP_ENV=production`, Contentful `master` |
| Testing (`education-testing`) | `pipeline.yml`, en push a `testing` | Imagen `education:<commit>`, `APP_ENV=testing`, Contentful `testing` |
| Entorno de calidad | `docker-compose.calidad.yml`, a mano | Imagen local `buenvivir-calidad/education:local`, Contentful `testing` forzado |

### Lo que ya garantiza el pipeline de la tesis (verificado leyendo `pipeline.yml`)

1. **La imagen lleva el commit en el nombre** (`education:${{ github.sha }}`). Lo que se despliega es exactamente el commit que pasó las pruebas.
2. **Cloud Run corre 1 sola instancia** (`--max-instances=1`) y se fuerza el 100 % del tráfico a la versión nueva (`update-traffic --to-latest`). No puede quedar "un servidor viejo" atendiendo.
3. **El smoke comprueba el environment**: falla si `/api/salud` no dice el Contentful esperado (`testing` o `master`).
4. **Vuelta atrás automática**: si el smoke falla, `cd-rollback` vuelve a la versión anterior.
5. **El servicio no arranca con una combinación peligrosa**: testing nunca apunta a `master` (`validateEnvironmentSafety`).

### Procedimiento para verificarlo en cada despliegue

| Paso | Qué se compara | Cómo |
|---|---|---|
| 1 | Versión analizada | En `historial.csv`: commit con quality gate aprobado |
| 2 | Versión desplegada | `gcloud run services describe education-<entorno> --format='value(spec.template.spec.containers[0].image)'`: termina en `:<commit>` |
| 3 | Coinciden | El commit del paso 1 tiene que ser el mismo del paso 2 |
| 4 | Configuración correcta | `/api/salud` devuelve el `appEnv` y el `contentfulEnvironment` esperados |

**Hueco encontrado:** `/api/salud` no dice qué commit está corriendo. Para saberlo hay que preguntarle a Cloud Run, con acceso a GCP. **Propuesta, no implementada:** agregar el commit a la respuesta de `/api/salud`, así el smoke lo compara solo. Es un cambio en el código del servicio, por eso queda como propuesta.

> **Estado:** el procedimiento está escrito, pero **todavía no se ejecutó** contra un despliegue real. Se hace en la demostración con un despliegue nuevo que pide la consigna.
