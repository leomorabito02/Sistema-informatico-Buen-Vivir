# Configuración y despliegue

## Requisitos

- Docker Desktop (con Docker Compose v2).
- Go 1.26 (los `go.mod` piden 1.26; con un Go anterior, `GOTOOLCHAIN=auto` baja el toolchain).
- Node 22 y pnpm 9 (frontend).
- Java 17+ solo si se quiere correr el emulador de Firestore fuera de Docker.
- Credenciales: los `.env` de cada servicio (no se versionan; ver `.gitignore`).

## Variables de entorno (solo nombres)

Cada servicio lee su `.env` (en Docker vía `env_file`). Plantillas
versionadas: `auth`, `human-promotion`, `labor-integration`,
`notifications`, `rxresume`, `frontend`, `api-gateway` (`.env.example`).
education, campaigns y dashboard no tienen plantilla versionada: sus
variables están en su CLAUDE/README y abajo.

| Servicio | Variables |
|---|---|
| auth | `PORT`, `DATABASE_URL_AUTH`, `JWT_SECRET`, `JWT_REFRESH_SECRET`, `JWT_EXPIRATION_MINUTES`, `JWT_REFRESH_EXPIRATION_DAYS`, `GOOGLE_APPLICATION_CREDENTIALS`, `FIRESTORE_DATABASE`, `CORS_ALLOWED_ORIGINS`, `WORKER_IMPACT_URL`, `WORKER_SYNC_TOKEN`, `NOTIFICATIONS_SERVICE_URL`, `FRONTEND_URL`, `RESET_PASSWORD_PATH` |
| education | `PORT` (8210), `APP_ENV` (testing/production), `CONTENTFUL_SPACE_ID`, `CONTENTFUL_ENVIRONMENT`, `CONTENTFUL_CDA_TOKEN`, `CONTENTFUL_CMA_TOKEN`, `CONTENTFUL_LOCALE`, `JWT_SECRET`, `GOOGLE_APPLICATION_CREDENTIALS`, `FIRESTORE_DATABASE`, `AUTH_SERVICE_URL`, `CORS_ALLOWED_ORIGINS`, `ENTREGAS_BUCKET`, `VERTEX_PROJECT_ID`, `VERTEX_LOCATION`, `GEMINI_MODEL` |
| human-promotion | `PORT`, `JWT_SECRET`, `ENCRYPTION_KEY` / `ENCRYPTION_KEY_SECRET`, `GOOGLE_APPLICATION_CREDENTIALS`, `FIRESTORE_DATABASE`, `AUTH_SERVICE_GRPC_TARGET`, `EDUCATION_SERVICE_URL`, `CORS_ALLOWED_ORIGINS`, `APP_ENV`, `PUBSUB_*` |
| campaigns | `PORT`, `JWT_SECRET`, `GOOGLE_APPLICATION_CREDENTIALS`, `FIRESTORE_DATABASE`, `GOOGLE_CLOUD_PROJECT`, `CORS_ALLOWED_ORIGINS` |
| labor-integration | `PORT`, `GOOGLE_CLOUD_PROJECT`, `GOOGLE_APPLICATION_CREDENTIALS`, `FIRESTORE_DATABASE`, `AUTH_SERVICE_GRPC_TARGET` o `AUTH_SERVICE_URL`, `NOTIFICATIONS_SERVICE_URL`, `INTERNAL_HMAC_SECRET`, **`JWT_SECRET`** (recomendado también en local; obligatorio en producción desde el 11/11/2026) |
| dashboard | `PORT`, `APP_ENV`, `JWT_SECRET`, `FIRESTORE_DATABASE`, `AUTH_SERVICE_URL`, `EDUCATION_SERVICE_URL`, `HUMAN_PROMOTION_SERVICE_URL`, `LABOR_SERVICE_URL`, `REPORTES_BUCKET`, `DASHBOARD_CACHE_TTL`, `UPSTREAM_ID_TOKEN`, `PUBSUB_*` |
| notifications | `PORT`, `GRPC_PORT`, `AUTH_GRPC_URL` o `AUTH_SERVICE_URL`, `RESEND_API_KEY`, `RESEND_FROM_EMAIL_NOTIFICATIONS`, `RESEND_FROM_EMAIL_ALERTS`, `DATABASE_URL_NOTIFICATIONS`, `FIREBASE_PROJECT_ID`, `FRONTEND_URL`, `PUBSUB_*` |
| rxresume | `PORT`, `DATABASE_URL_RXRESUME`, `RENDERER_URL`, `GCP_PROJECT_ID`, `GCP_BUCKET_NAME`, `GOOGLE_APPLICATION_CREDENTIALS` |
| frontend | `API_URL` |
| api-gateway (GCP) | `GCP_PROJECT_ID`, `FIREBASE_PROJECT_ID`, `GATEWAY_HOST`, `*_SERVICE_URL`, `CORS_ALLOWED_ORIGINS`, `RATE_LIMIT_RPS`, `RATE_LIMIT_BURST` |

Pub/Sub (servicios con eventos): `PUBSUB_PROJECT_ID`, `PUBSUB_EMULATOR_HOST`
(solo local), `PUBSUB_PUSH_AUTH` (`disabled` solo en local),
`PUBSUB_PUSH_AUDIENCE`, `PUBSUB_PUSH_SERVICE_ACCOUNTS`. Ver `eventos/README.md`.

**El mismo `JWT_SECRET` en todos los servicios.**

## Entorno local (Docker)

```bash
docker compose up -d --build      # levanta todo
docker compose ps
```

| Servicio | URL local |
|---|---|
| Frontend | http://localhost:4200 |
| API Gateway (nginx) | http://localhost:8080 |
| auth | :8082 (HTTP), :50051 (gRPC) |
| education | :8210 |
| human-promotion | :8085 |
| labor-integration | 127.0.0.1:8310 |
| rxresume | :8087 |
| notifications | :8088 (HTTP), :50052 (gRPC) |
| campaigns, dashboard, cv-renderer | sin puerto publicado (por el gateway) |
| Emulador Pub/Sub | 127.0.0.1:8086 |

Trampas conocidas:
- Después de cambiar `docker-compose.yml` o un `.env`, recrear los
  contenedores (`up -d --build` o `--force-recreate`): un contenedor viejo
  conserva la configuración con la que se creó.
- Tras reconstruir servicios, reiniciar `api-gateway` (nginx guarda IPs viejas → 404/502).
- En Git Bash usar `MSYS_NO_PATHCONV=1` para rutas `/api/...`.
- El emulador de Pub/Sub guarda todo en memoria: desde el 11/11/2026 vuelve
  a crear topics y suscripciones cada vez que arranca.
- El stack local usa Firestore y Neon **reales, compartidos con testing**.

## CI/CD

Un solo workflow: `.github/workflows/pipeline.yml`.

- **CI** por servicio: lint, tests (con emuladores donde hace falta),
  cobertura mínima, `govulncheck`, build de imagen y Trivy; `ci-gate` exige
  que pasen todos.
- **CD** (push a `testing`): imágenes a Artifact Registry, deploy de cada
  servicio a Cloud Run (`<servicio>-<sufijo>`), API Gateway con
  `openapi.yaml` (envsubst), frontend a Cloudflare Pages, worker.
- **Post-deploy:** smoke tests, ZAP (DAST) y k6; si algo falla, rollback a la
  revisión anterior de cada servicio.

Detalle: `ci-cd-pipeline.md` y `PIPELINE_EDUCATION.md`. Variables y secretos
del environment de testing: `referencias/despliegue/` (solo nombres).

Desde el 22/11/2026 el CD configura `FRONTEND_URL=https://testing.ceibuenvivir.pages.dev`
en auth y notifications **solo en TESTING** (paso "Configure FRONTEND_URL").
Es el destino de los enlaces de los correos (confirmar cuenta, recuperar
contraseña) y la base del logo de los correos. En producción no se toca:
si falta, auth usa `https://ceibuenvivir.org` y el logo sale de
`https://ceibuenvivir.pages.dev/assets/logo.png`.

Pendientes de configuración (ver `PENDIENTES.md`): `PUBSUB_PUSH_AUDIENCE` y
`PUBSUB_PUSH_SERVICE_ACCOUNTS` en el CD; credenciales de FCM; suscripciones
push de notifications en GCP (`infra/pubsub/setup.sh`).
