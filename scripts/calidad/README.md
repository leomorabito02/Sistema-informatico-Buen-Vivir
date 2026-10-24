# Scripts del TP de Calidad (Educación)

Todo lo que se corre para medir y probar Educación está acá, separado
por para qué sirve. El código del servicio queda solo en `education/`.
Se corren desde la raíz del repo (los que necesitan estar dentro de
`education/` entran solos).

| Carpeta | Para qué | Scripts |
|---|---|---|
| `sonar/` | Análisis estático con SonarQube y el historial de métricas | `configurar_sonar.sh` (proyecto y quality gate), `linea_base.sh` (foto del 1/8), `analizar.sh` (pruebas, cobertura, análisis y veredicto), `exportar_metricas.py` (agrega la fila a `docs/calidad/reportes/historial.csv`) |
| `seguridad/` | Seguridad | `seguridad.sh` (gosec, govulncheck y ZAP en la PC), `resumen_seguridad.py` (veredicto del job `seguridad-education`) |
| `pruebas/` | Pruebas sobre el servicio funcionando y mutación | `smoke.sh`, `e2e.sh` (`e2e.py`), `carga.js` (`carga.sh` la corre en la PC), `stress.js`, `esfuerzo_sostenido.js`, `comun.js`, `mutacion.sh` |
| `entorno/` | La copia de Educación para probar (ver `entorno/README.md`) | `preparar_env.sh`, `auth_falso.py`, `cargar_por_api.py`, `token_prueba.py` |
| `documentos/` | Pasar los `.md` de `docs/calidad` a Word y de vuelta | `md_a_word.py`, `word_a_md.py` |

Dónde quedan los resultados (ninguno se versiona, salvo el historial):

- `reportes-servicio/`: lo que dejan smoke, E2E, carga y stress, y los
  reportes de las pruebas de contrato y de base de datos de
  `education/tests/`. En el pipeline se sube como artefacto.
- `education/`: `coverage.out`, `test-report.json`, `mutation_report.txt`
  y los de `seguridad.sh` (`sast_report.json`, `sca_report.txt`,
  `zap_report.html`), porque las herramientas los escriben donde corren.
- `docs/calidad/reportes/`: el historial de SonarQube.

Las pruebas de Go del servicio no están acá: las unitarias van al lado
del código (`education/internal/**/*_test.go`) y las que prueban varias
piezas juntas (contrato, base de datos, Pub/Sub) en `education/tests/`.

## Pruebas sobre el servicio (smoke, E2E, carga, DAST y stress)

El job `servicio-education` de `.github/workflows/calidad-tp.yml` levanta
Educación dentro del runner con `docker-compose.calidad.yml`, carga los
datos de prueba y corre estas pruebas contra `http://127.0.0.1:18210`.
Cada una termina con error si no pasa, y eso frena el pipeline.

| Prueba | Script | Qué hace | Umbral | Cuándo corre |
|---|---|---|---|---|
| Smoke | `pruebas/smoke.sh` | `/api/salud` en UP con Contentful `testing`; 401 sin token; catálogo, actividades y cursos del participante responden | Que todo responda | Cada push a `Calidad` |
| E2E | `pruebas/e2e.sh` (`e2e.py`) | Crea un curso `[E2E]` en Contentful testing, lo publica, crea una actividad con habilidad, inscribe a un participante, el participante entrega, se verifica asistencia, habilidad y progreso, y se da de baja el curso. Borra el curso de Contentful al final | Los 13 pasos en verde | Cada push a `Calidad` |
| ST-0 Carga | `pruebas/carga.js` (Leonardo) | 5 → 10 usuarios en 25 s leyendo el catálogo; un solo usuario crea unos 4 cursos `[LOAD-TEST]`, que se borran al final | p95 < 800 ms y checks > 95 % | Cada push a `Calidad` |
| DAST | El baseline de ZAP de `seguridad/seguridad.sh` | Escaneo pasivo del servicio en marcha | Que ZAP no dé error (las advertencias se informan) | Cada push a `Calidad` |
| ST-1 Stress | `pruebas/stress.js` | Escalera de 10 → 20 → 30 → 40 usuarios, 1 min cada escalón, solo lecturas | Falla el primer escalón con p95 > 2 s o errores > 5 %; corta si los errores pasan del 20 % | Cada push a `Calidad` |
| ST-2 Esfuerzo sostenido | `pruebas/esfuerzo_sostenido.js` | 10 usuarios constantes durante 30 min, solo lecturas | p95 de los últimos 5 min ≤ 20 % peor que el de los primeros 5; errores < 5 %; memoria al empezar y al terminar | Solo con Run workflow y la opción tildada |
| Mutación | `pruebas/mutacion.sh` (Leonardo) | gremlins sobre `internal/` | Que gremlins no dé error | Solo con Run workflow y la opción tildada |

`pruebas/comun.js` tiene lo que comparten stress y esfuerzo sostenido:
los tokens y las lecturas de cada usuario virtual (catálogo, actividades e
inscriptos de uno de los cursos cargados).

## Cuánto se escribe en Contentful

Contentful testing es real y su plan gratuito limita los pedidos (riesgo
RP-04). Por eso:

- El E2E escribe **un** curso por corrida y lo borra.
- La carga escribe **unos 4** cursos por corrida (solo el primer usuario
  virtual, una de cada 5 vueltas) y los borra.
- Stress y esfuerzo sostenido **no escriben**: solo leen, y el catálogo
  sale de la caché de 5 minutos de Educación.

## Qué recibe cada script

- `BASE_URL`: la URL de Educación, `http://127.0.0.1:18210`.
- `REPORTES`: la carpeta donde queda el reporte (`smoke.txt`, `e2e.json`,
  `*-resumen.json`, `sostenido-degradacion.json`). Todo se sube como
  artefacto `servicio-education-<commit>`.
- `JWT_SECRET` (solo k6): el de `.env.calidad`, para firmar tokens. Los
  scripts de shell y Python usan `scripts/calidad/entorno/token_prueba.py`.
- Datos ya cargados por `cargar_por_api.py`: 3 cursos de Contentful
  testing, 6 actividades, 24 inscripciones y 15 entregas.
- El rate limiter está apagado en este entorno (`RATE_LIMIT_DISABLED=true`),
  porque todos los pedidos de k6 salen de la misma IP.

## Probarlo en la PC antes de subir

Desde la raíz del repo, en Git Bash:

```bash
bash scripts/calidad/entorno/preparar_env.sh
docker compose -f docker-compose.calidad.yml up -d --build
python scripts/calidad/entorno/cargar_por_api.py
BASE_URL=http://127.0.0.1:18210 REPORTES=reportes-servicio bash scripts/calidad/pruebas/smoke.sh
BASE_URL=http://127.0.0.1:18210 REPORTES=reportes-servicio bash scripts/calidad/pruebas/e2e.sh

# k6 dentro de la red del entorno (en Windows, --network host no llega a la PC).
# ESCALON, DURACION y VENTANA acortan stress y sostenido para probar el script.
JWT=$(grep '^JWT_SECRET=' scripts/calidad/entorno/.env.calidad | cut -d= -f2-)
MSYS_NO_PATHCONV=1 docker run --rm --network buenvivir-calidad-net -v "$(pwd -W):/repo" -w /repo \
  -e BASE_URL=http://education-calidad:8210 -e JWT_SECRET="$JWT" -e ESCALON=10s \
  grafana/k6:0.54.0 run scripts/calidad/pruebas/stress.js

docker compose -f docker-compose.calidad.yml down
```

Las pruebas de Leonardo también corren solas, sin este entorno:
`scripts/calidad/pruebas/carga.sh` (necesita `education/.env.testing`),
`scripts/calidad/pruebas/mutacion.sh` y `scripts/calidad/seguridad/seguridad.sh`.
