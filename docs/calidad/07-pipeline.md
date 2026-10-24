# Pipeline de CI de Educación: estructura y cómo correrlo

El TP tiene un solo pipeline, de integración continua:
`.github/workflows/calidad-tp.yml`. No despliega nada. Corre todas las
pruebas de Educación: las del código (unitarias, integración, contrato y
mutación), el análisis estático y de seguridad, y las pruebas sobre el
servicio funcionando (smoke, E2E, carga, DAST, stress y esfuerzo
sostenido). Para estas últimas levanta Educación dentro del propio runner. `pipeline.yml`, el
pipeline de la tesis con despliegue a Cloud Run, queda fuera del TP.

## 1. Estructura

```
push a Calidad  (o Run workflow, que además permite pedir lo largo)
        │
        ▼
ETAPA 1 ─ en paralelo ────────────────────────────── frena si algo falla
  ┌──────────────────────────────┐  ┌──────────────────────────────┐
  │ sonarqube-education          │  │ seguridad-education          │
  │ · emulador de Firestore      │  │ · gosec, govulncheck         │
  │ · unitarias, integración y   │  │ · Gitleaks, Semgrep          │
  │   contrato (go test ./...)   │  │ · Trivy (imagen Docker)      │
  │ · cobertura + SonarQube      │  │                              │
  │ · quality gate TP-Calidad-N3 │  │                              │
  └──────────────────────────────┘  └──────────────────────────────┘
        │ solo si las pruebas y el gate dieron verde
        ▼
ETAPA 2 ─ en paralelo
  ┌──────────────────────────────┐  ┌──────────────────────────────┐
  │ servicio-education           │  │ mutacion-education           │
  │ · levanta el entorno de      │  │ · mutacion.sh                │
  │   calidad en el runner       │  │ · solo a pedido              │
  │ · carga datos por la API     │  └──────────────────────────────┘
  │ · smoke → E2E → carga ST-0 → │
  │   DAST (ZAP) → stress ST-1   │
  │ · esfuerzo sostenido ST-2    │
  │   (solo a pedido)            │
  └──────────────────────────────┘
```

| Etapa | Job | Qué prueba | Frena si… | Evidencia | Estado |
|---|---|---|---|---|---|
| 1 | `sonarqube-education` | Unitarias, integración (emulador de Firestore) y contrato; cobertura y gate | Una prueba falla o el gate da rojo | Artefacto `calidad-education-<commit>` | Hecho |
| 1 | `seguridad-education` | Código, librerías, claves y la imagen Docker | Hallazgo bloqueante | Artefacto `seguridad-education-<commit>` | Hecho |
| 2 | `servicio-education` · smoke | `/api/salud` en UP con Contentful `testing`, 401 sin token, catálogo, actividades y cursos del participante | Algo no responde | `smoke.txt` en el artefacto `servicio-education-<commit>` | Listo |
| 2 | `servicio-education` · E2E | Un curso de punta a punta: alta y publicación en Contentful testing, actividad, inscripción, entrega, asistencia, habilidad, progreso y baja | Falla un paso | `e2e.json` | Listo |
| 2 | `servicio-education` · carga (ST-0) | `scripts/calidad/pruebas/carga.js` (Leonardo): 5 → 10 usuarios en 25 s leyendo el catálogo; un usuario crea unos 4 cursos en Contentful testing | p95 > 800 ms o checks < 95 % | `carga-resumen.json` | Listo |
| 2 | `servicio-education` · DAST | Baseline de OWASP ZAP contra el servicio, el mismo de `seguridad.sh` | Error de ZAP (las advertencias se informan, como en el script) | Artefacto `zap-education-<commit>` | Listo |
| 2 | `servicio-education` · stress (ST-1) | 10 → 20 → 30 → 40 usuarios, 1 min cada escalón, solo lecturas | p95 > 2 s o errores > 5 % en algún escalón | `stress-resumen.json`, con el p95 y los errores de cada escalón | Listo |
| 2 | `servicio-education` · sostenido (ST-2) | 10 usuarios durante 30 min, solo lecturas | p95 de los últimos 5 min > 20 % peor que el de los primeros 5, o errores ≥ 5 % | `sostenido-degradacion.json` y memoria al empezar y al terminar | Listo; solo a pedido |
| 2 | `mutacion-education` | Si las pruebas detectan cambios en el código (`scripts/calidad/pruebas/mutacion.sh`, de Leonardo) | Gremlins da error | Tabla en el resumen y artefacto `mutacion-education-<commit>` | Listo; solo a pedido |

**Qué se levanta en la etapa 2** (`docker-compose.calidad.yml`):
Educación con el mismo Dockerfile de producción, el emulador de Firestore
vacío, un auth-service de prueba con 8 participantes sintéticos y
Contentful real en el environment `testing`. Cloud Storage, Pub/Sub e IA
quedan apagados, y también el límite de pedidos por IP
(`RATE_LIMIT_DISABLED=true`): todos los pedidos de k6 salen de la misma IP
y el límite los cortaría. Los datos se cargan por la API con
`cargar_por_api.py`, así pasan por las mismas validaciones que un dato
real.

Qué hace y qué recibe cada script está en `scripts/calidad/README.md`.

**Cuánto se escribe en Contentful testing** (su plan gratuito limita los
pedidos, riesgo RP-04): el E2E crea un curso y la carga unos 4, y los dos
los borran al terminar. Stress y esfuerzo sostenido solo leen.

**Por qué las cargas son chicas:** parten del pico de la prueba de carga de
Leonardo (10 usuarios). El stress llega a 4 veces ese valor; si ningún
escalón pasa los umbrales, el punto de quiebre del entorno está por encima
de 40 usuarios, y eso es lo que se informa.

**Por qué ST-2 y mutación van solo a pedido:** ST-2 dura 30 minutos y la
mutación, varios más. Correrlos en cada push haría esperar media hora o más
por cada cambio; para el informe alcanza con una corrida pedida antes de la
entrega.

## 2. Cómo correrlo

### En GitHub (la corrida que vale como evidencia)

- **Con cada push a `Calidad`:** corren las etapas 1 y 2, sin ST-2 ni
  mutación.
- **Todo, incluido lo largo:** GitHub → Actions → "TP Calidad — CI
  (Educación)" → Run workflow → rama `Calidad` → tildar
  "Correr el esfuerzo sostenido" y "Correr las pruebas de mutación". Desde
  la terminal:

  ```bash
  gh workflow run calidad-tp.yml --ref Calidad -f esfuerzo_sostenido=true -f mutacion=true
  gh run watch          # sigue la corrida en vivo
  gh run download       # baja los artefactos con los reportes
  ```

- **Qué mirar:** el resumen de la corrida tiene el quality gate, las
  métricas y una tabla con el resultado de smoke, E2E, carga, DAST, stress
  y esfuerzo sostenido. Los reportes completos están en los artefactos.
- **Requisito:** los secrets `CONTENTFUL_SPACE_ID`, `CONTENTFUL_CDA_TOKEN`,
  `CONTENTFUL_CMA_TOKEN` y `CONTENTFUL_LOCALE` del repositorio (ya los usa
  `pipeline.yml`).

### En la PC (para probar antes de subir)

Desde la raíz del repo, en Git Bash:

1. **Etapa 1:** con el emulador de Firestore en el puerto 8390, SonarQube
   26.8 en `http://localhost:9000` y el token en `~/.sonar-token-tp`:

   ```bash
   FIRESTORE_EMULATOR_HOST=127.0.0.1:8390 scripts/calidad/sonar/analizar.sh education
   ```

   Si Go se queja de la versión (`go1.26.0 does not match go1.25.2`),
   anteponer `env -u GOROOT GOTOOLCHAIN=go1.26.9`.

2. **Etapa 2:** levantar el entorno de calidad y correr los scripts como
   indica `scripts/calidad/README.md`. Las pruebas de Leonardo también
   corren solas, sin el entorno: `scripts/calidad/pruebas/carga.sh`,
   `scripts/calidad/pruebas/mutacion.sh` y `scripts/calidad/seguridad/seguridad.sh`.
