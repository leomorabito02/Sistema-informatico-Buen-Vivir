# Pruebas sobre el servicio (smoke, E2E y stress)

El job `servicio-education` de `.github/workflows/calidad-tp.yml` levanta
Educación dentro del runner con `docker-compose.calidad.yml`, carga los
datos de prueba y corre los scripts de esta carpeta. Mientras un script no
exista, su paso queda como "pendiente" en el resumen de la corrida.

| Archivo | Prueba | Cuándo corre |
|---|---|---|
| `smoke.sh` | Smoke: el servicio arrancó y responde lo básico | En cada push a `Calidad` |
| `e2e.sh` | E2E: un flujo completo de punta a punta, con Contentful testing real | En cada push a `Calidad` |
| `stress.js` | ST-1 con k6: 20 → 50 → 100 → 200 usuarios, 2 min cada escalón | En cada push a `Calidad` |
| `esfuerzo_sostenido.js` | ST-2 con k6: 20 usuarios durante 60 minutos | Solo con Run workflow y la opción tildada |

## Qué recibe cada script

- `BASE_URL`: la URL de Educación, `http://127.0.0.1:18210`.
- `REPORTES`: la carpeta donde dejar el reporte. Todo lo que quede ahí se
  sube como artefacto `servicio-education-<commit>`. En los scripts de k6,
  `REPORTES` vale `/repo/reportes-servicio` (adentro del contenedor de k6).
- Tokens: `python3 scripts/calidad/entorno/token_prueba.py <rol>`, con
  `role_admin`, `role_teacher`, `role_volunteer` o `role_participant`.
- Datos ya cargados: 3 cursos de Contentful testing, 6 actividades,
  24 inscripciones y 15 entregas (ver `scripts/calidad/entorno/README.md`).

## Qué tiene que hacer

- Terminar con error (exit distinto de 0) si la prueba no pasa: eso frena
  el pipeline.
- En k6, poner los umbrales del plan como `thresholds`. Para ST-1, que
  falle si el p95 supera 2 s o los errores superan el 5 %. Para ST-2, que
  el p95 final no sea más de 20 % peor que el inicial.
- No usar datos de personas reales ni el environment `master` de
  Contentful (el entorno lo impide igual).

## Probarlo en la PC antes de subir

```bash
bash scripts/calidad/entorno/preparar_env.sh
docker compose -f docker-compose.calidad.yml up -d --build
python scripts/calidad/entorno/cargar_por_api.py
BASE_URL=http://127.0.0.1:18210 REPORTES=reportes-servicio bash scripts/calidad/pruebas/smoke.sh
docker compose -f docker-compose.calidad.yml down
```
