# Pipeline de CI de Educación: estructura y cómo correrlo

El TP tiene un solo pipeline, de integración continua:
`.github/workflows/calidad-tp.yml`. No despliega nada. Corre todas las
pruebas de Educación: las del código (unitarias, integración, contrato y
mutación), el análisis estático y de seguridad, y las pruebas sobre el
servicio funcionando (smoke, E2E, stress y esfuerzo sostenido). Para estas
últimas levanta Educación dentro del propio runner. `pipeline.yml`, el
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
  │ · levanta el entorno de      │  │ · gremlins sobre internal/   │
  │   calidad en el runner       │  │ · solo a pedido              │
  │ · carga datos por la API     │  └──────────────────────────────┘
  │ · smoke → E2E → stress ST-1  │
  │ · esfuerzo sostenido ST-2    │
  │   (solo a pedido)            │
  └──────────────────────────────┘
```

| Etapa | Job | Qué prueba | Frena si… | Evidencia | Estado |
|---|---|---|---|---|---|
| 1 | `sonarqube-education` | Unitarias, integración (emulador de Firestore) y contrato; cobertura y gate | Una prueba falla o el gate da rojo | Artefacto `calidad-education-<commit>` | Hecho |
| 1 | `seguridad-education` | Código, librerías, claves y la imagen Docker | Hallazgo bloqueante | Artefacto `seguridad-education-<commit>` | Hecho |
| 2 | `servicio-education` · smoke | El servicio arrancó y responde lo básico | El script termina con error | Artefacto `servicio-education-<commit>` | Estructura lista; falta `smoke.sh` (Leonardo) |
| 2 | `servicio-education` · E2E | Un flujo completo con Contentful testing real | Ídem | Ídem | Falta `e2e.sh` (Leonardo) |
| 2 | `servicio-education` · stress | ST-1: 20 → 200 usuarios; frena si p95 > 2 s o errores > 5 % | Se pasa un umbral | `stress-resumen.json` | Falta `stress.js` (Leonardo) |
| 2 | `servicio-education` · sostenido | ST-2: 20 usuarios, 60 min; p95 final ≤ 20 % peor | Ídem | `sostenido-resumen.json` y memoria | Falta `esfuerzo_sostenido.js` (Leonardo); solo a pedido |
| 2 | `mutacion-education` | Si las pruebas detectan cambios en el código | Gremlins da error | Artefacto `mutacion-education-<commit>` | Hecho; solo a pedido |

**Qué se levanta en la etapa 2** (`docker-compose.calidad.yml`):
Educación con el mismo Dockerfile de producción, el emulador de Firestore
vacío, un auth-service de prueba con 8 participantes sintéticos y
Contentful real en el environment `testing`. Cloud Storage, Pub/Sub e IA
quedan apagados. Los datos se cargan por la API con
`cargar_por_api.py`, así pasan por las mismas validaciones que un dato
real.

**Mientras falte un script**, su paso avisa y el resumen de la corrida lo
muestra como `pendiente`: no se da por aprobado. Qué recibe y qué tiene
que hacer cada script está en `scripts/calidad/pruebas/README.md`.

**Por qué ST-2 y mutación van solo a pedido:** ST-2 dura 60 minutos y la
mutación, varios más. Correrlos en cada push haría esperar una hora por
cada cambio; para el informe alcanza con una corrida pedida antes de la
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
  métricas y una tabla con el resultado de smoke, E2E, stress y esfuerzo
  sostenido. Los reportes completos están en los artefactos.
- **Requisito:** los secrets `CONTENTFUL_SPACE_ID`, `CONTENTFUL_CDA_TOKEN`,
  `CONTENTFUL_CMA_TOKEN` y `CONTENTFUL_LOCALE` del repositorio (ya los usa
  `pipeline.yml`).

### En la PC (para probar antes de subir)

Desde la raíz del repo, en Git Bash:

1. **Etapa 1:** con el emulador de Firestore en el puerto 8390, SonarQube
   26.8 en `http://localhost:9000` y el token en `~/.sonar-token-tp`:

   ```bash
   FIRESTORE_EMULATOR_HOST=127.0.0.1:8390 scripts/calidad/analizar.sh education
   ```

   Si Go se queja de la versión (`go1.26.0 does not match go1.25.2`),
   anteponer `env -u GOROOT GOTOOLCHAIN=go1.26.9`.

2. **Etapa 2:** levantar el entorno de calidad y correr los scripts como
   indica `scripts/calidad/pruebas/README.md`.
