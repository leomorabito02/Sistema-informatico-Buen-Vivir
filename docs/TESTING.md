# Testing

Estado al 11/11/2026 (auditoría). Detalle completo, con tests faltantes,
en `INFORME_AUDITORIA_BUEN_VIVIR.pdf` (secciones 9 a 11).

## Resultado de la última ejecución

| Módulo | Tests | Resultado | Cobertura |
|---|---|---|---|
| pkg | 32 funciones | 2/2 paquetes OK | 94,0 % |
| auth | 527 | 11/11 OK | 83,6 % |
| education | 628 | 15/15 OK | 90,2 % |
| human-promotion | 101 | 11/11 OK | 75,8 % |
| dashboard | 92 | 9/9 OK | 85,8 % |
| campaigns | 70 | 8/8 OK | 90,2 % |
| notifications | 98 | 14/14 OK | 63,2 % |
| labor-integration | 50 | 7/7 OK | 59,3 % |
| rxresume | 7 | 2/2 OK | 38,4 % |
| frontend (Karma) | 694 specs | 694/694 OK | 71,45 % sentencias |

## Cómo correrlos

**Go, con emuladores** (sin emulador, los tests de repositorios se saltean):

```bash
# emulador de Firestore: el del stack de Calidad (127.0.0.1:18390) o
#   gcloud emulators firestore start --host-port=127.0.0.1:8390
# emulador de Pub/Sub: el del docker-compose (localhost:8086)
cd education
FIRESTORE_EMULATOR_HOST=127.0.0.1:8390 PUBSUB_EMULATOR_HOST=localhost:8086 go test ./...
```

Igual en `pkg`, `auth`, `human-promotion`, `campaigns`, `dashboard`,
`notifications`, `labor-integration` y `rxresume`. En Windows no anda
`-race` (no hay gcc).

Ojo: los tests de education reescriben `reports/database/*` y
`reports/pact/*`; si no se quieren esos cambios, `git checkout` de esas rutas.

**Frontend:**

```bash
cd frontend
npx ng test --watch=false --browsers=ChromeHeadless --code-coverage
```

Karma exige 70 % de sentencias. `frontend/coverage/` está versionado:
restaurarlo con `git checkout -- frontend/coverage` si no se quiere el cambio.

## Pruebas no funcionales

| Prueba | Archivo | Cómo | Último resultado |
|---|---|---|---|
| Carga local (k6) | `tests/carga/k6_carga_local.js` | `docker run --rm -v "$PWD/tests/carga:/scripts" grafana/k6:0.54.0 run /scripts/k6_carga_local.js` | 350.639 pedidos, 0 % de 5xx, p95 12 ms (servicios) |
| Carga con volumen (Campañas) | `campaigns/internal/routes/carga_test.go` | `CARGA=1 FIRESTORE_EMULATOR_HOST=... go test ./internal/routes -run Carga -v` | 500 campañas p95 485 ms; 2000 p95 8,8 s |
| Sonda de seguridad | `tests/seguridad/sonda/main.go` | `JWT_SECRET=... go run ./tests/seguridad/sonda > docs/evidencias/sonda-seguridad-resultado.json` | 38/39 OK (falla: labor sin redesplegar) |
| k6 y ZAP del CD | `education/load_test.js`, `tests/performance/k6_auth_load_test.js`, job `cd-dast-zap` | Solo en el CD de testing | No ejecutados en la auditoría |

La carga y la sonda solo hacen pedidos que no modifican datos: el stack
local comparte bases con testing. Evidencias en `evidencias/`.

## Lo que no corre

- `frontend/cypress/`: un spec, sin Cypress instalado ni archivo de soporte.
- Mutación (Stryker / go-mutesting): deshabilitada en el CI (`if: false`).
- `tests/{security,performance,e2e_a11y}/generate_*.go`: generan reportes
  con valores fijos para el TP de Calidad; **no son pruebas**.

## Tests agregados en la auditoría

Cursos sin materiales (Go y front), baja lógica de campañas (servicio,
repositorio con emulador, rutas con permisos, front), notificaciones de PH
(creado_por, enlaces, deduplicación), aviso de campaña único, identidad y
permisos de ofertas en labor, alta de perfiles ajenos y baja de perfil en
auth, progreso para admin/voluntario, contador de notificaciones y
Seguimiento por rol en el front.
