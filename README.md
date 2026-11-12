# Buen Vivir App

Sistema de gestión interna del **Centro de Ecología Integral** (Centro
Barrial "La Otra Orilla", Alta Gracia, Córdoba): una aplicación web
progresiva (PWA) con backend de microservicios en Go sobre Google Cloud.

## Objetivo y problema que resuelve

El centro acompaña a personas de la comunidad con formación, seguimiento
personal e inserción laboral, y difunde campañas ambientales y sociales.
Esa información estaba dispersa (planillas, papeles, mensajes). Buen Vivir
App la reúne en un solo sistema, con acceso por rol, trazabilidad
(auditoría) y protección de datos sensibles (fichas cifradas, eventos sin
datos personales), y da al equipo indicadores para tomar decisiones.

## Funcionalidades principales

- **Usuarios y acceso:** registro con aprobación (voluntarios, empresas,
  docentes), login con email o Google, 2FA, recuperación de contraseña, roles.
- **Educación:** cursos en Contentful (con materiales opcionales),
  inscripciones, actividades y entregas que generan asistencia, asistencia
  general diaria, progreso, habilidades sugeridas con IA (Gemini/OpenAI),
  contenido de docentes externos con moderación, biblioteca pública, comentarios.
- **Promoción Humana:** ficha de proceso personal cifrada, hitos, línea de
  tiempo (también para el participante), exportación PDF, auditoría.
- **Inserción Laboral:** ofertas con aprobación, postulaciones, CV con
  Reactive Resume.
- **Campañas:** crear, editar, publicar y dar de baja; vista pública sin login.
- **Dashboard:** indicadores, gráficas, filtros y reportes PDF.
- **Notificaciones:** in-app, email (Resend) y push (FCM), con preferencias.

Estado de cada historia de usuario: [`docs/HISTORIAS_DE_USUARIO.md`](docs/HISTORIAS_DE_USUARIO.md).

## Roles

Administrador, voluntario, participante, empresa, docente (interno) y
docente externo; los visitantes ven la parte pública. Qué puede hacer cada
uno: [`docs/ROLES_Y_PERMISOS.md`](docs/ROLES_Y_PERMISOS.md).

## Tecnologías

| Capa | Tecnología |
|---|---|
| Frontend | Angular 20, PrimeNG, PWA, Firebase (Auth con Google y FCM), Cloudflare Pages |
| Backend | Go 1.26, chi, GORM, gRPC |
| Datos | PostgreSQL (Neon), Firestore, Contentful (CMS headless), Cloud Storage |
| Mensajería | Google Pub/Sub con outbox (emulador en local) |
| Infraestructura | Cloud Run, Google Cloud API Gateway, nginx (local), Docker Compose, GitHub Actions |
| Integraciones | Contentful, Vertex AI Gemini + OpenAI, Reactive Resume, Resend, Google Maps |

## Arquitectura

```
PWA Angular ──► API Gateway (nginx local / Google API Gateway) ──► microservicios Go
                                                                    │
                     eventos de dominio ◄──── Pub/Sub (outbox) ◄────┘
```

| Servicio | Carpeta | Responsabilidad | Datos |
|---|---|---|---|
| Autenticación | `auth/` | Usuarios, roles, perfiles, JWT, 2FA, Google | PostgreSQL |
| Educación | `education/` | Cursos, inscripciones, actividades, asistencia, habilidades, biblioteca | Contentful + Firestore + Storage |
| Promoción Humana | `human-promotion/` | Fichas, hitos, línea de tiempo, PDF | Firestore |
| Campañas | `campaigns/` | Campañas de concientización | Firestore |
| Inserción Laboral | `labor-integration/` | Ofertas y postulaciones | Firestore |
| CV | `rxresume/` (+ `cv-renderer`) | Generación de CV | PostgreSQL |
| Dashboard | `dashboard/` | Indicadores y reportes | agrega de los demás |
| Notificaciones | `notifications/` | In-app, email, push, preferencias | PostgreSQL |

Más detalle: [`docs/ARQUITECTURA.md`](docs/ARQUITECTURA.md),
[`docs/BASE_DE_DATOS.md`](docs/BASE_DE_DATOS.md),
[`docs/API_Y_ENDPOINTS.md`](docs/API_Y_ENDPOINTS.md),
[`docs/eventos/README.md`](docs/eventos/README.md).

## Estructura del repositorio

```
auth/ education/ human-promotion/ campaigns/ labor-integration/
dashboard/ notifications/ rxresume/   Servicios Go (cmd/server + internal/)
pkg/                    Librería de eventos y outbox (se copia a cada servicio)
frontend/               PWA Angular
api-gateway/            nginx.conf (local) y openapi.yaml (GCP)
infra/pubsub/           Topics y suscripciones (setup.sh en GCP, emulator.sh en local)
cloudflare-worker/      Worker de métricas de impacto (landing)
docs/                   Documentación (índice en docs/README.md)
tests/                  Carga (k6), sonda de seguridad, generadores del TP de Calidad
scripts/                sync-eventlib.sh, informe-auditoria/, historico-openapi/, calidad/
docker-compose.yml      Stack local completo
.github/workflows/      CI/CD
```

## Requisitos previos

Docker Desktop, Go 1.26, Node 22 y pnpm 9. Además, los archivos `.env` de
cada servicio (no se versionan; hay plantillas `.env.example` y la lista de
variables en [`docs/CONFIGURACION_Y_DESPLIEGUE.md`](docs/CONFIGURACION_Y_DESPLIEGUE.md)).
Todos los servicios deben compartir el mismo `JWT_SECRET`.

## Instalación y ejecución

```bash
git clone <repo> && cd Sistema-informatico-Buen-Vivir
# completar los .env de cada servicio (ver docs/CONFIGURACION_Y_DESPLIEGUE.md)
docker compose up -d --build
```

- Frontend: http://localhost:4200
- API Gateway: http://localhost:8080 (por ejemplo `curl http://localhost:8080/api/salud`)

Frontend fuera de Docker:

```bash
cd frontend && pnpm install && pnpm start
```

> El stack local usa Firestore y Neon reales, compartidos con el ambiente de
> testing: cuidado con los datos que se crean.

## Comandos importantes

| Acción | Comando |
|---|---|
| Levantar / recrear todo | `docker compose up -d --build` |
| Ver logs de un servicio | `docker compose logs -f education-service` |
| Tests de un servicio Go | `cd education && FIRESTORE_EMULATOR_HOST=127.0.0.1:8390 PUBSUB_EMULATOR_HOST=localhost:8086 go test ./...` |
| Tests del frontend | `cd frontend && npx ng test --watch=false --browsers=ChromeHeadless --code-coverage` |
| Formato del frontend | `cd frontend && pnpm run format:check` |
| Copiar la librería de eventos | `scripts/sync-eventlib.sh` |
| Prueba de carga local | `docker run --rm -v "$PWD/tests/carga:/scripts" grafana/k6:0.54.0 run /scripts/k6_carga_local.js` |
| Regenerar informe y HU | `python scripts/informe-auditoria/generar.py` |

Cómo correr cada tipo de test: [`docs/TESTING.md`](docs/TESTING.md).

## Estado del proyecto

Auditoría del 11/11/2026 sobre 64 historias de usuario: 41 implementadas,
17 parciales, 2 no implementadas y 4 pendientes de verificación (de punta a
punta en testing). Todas las suites de tests pasan (9 módulos Go y 694
specs del frontend). Lo que falta está en [`docs/PENDIENTES.md`](docs/PENDIENTES.md);
el informe completo es [`docs/INFORME_AUDITORIA_BUEN_VIVIR.pdf`](docs/INFORME_AUDITORIA_BUEN_VIVIR.pdf).

## Documentación complementaria

Índice general: [`docs/README.md`](docs/README.md). Decisiones técnicas:
[`docs/adr/`](docs/adr/README.md). Referencia de Educación:
`education/CLAUDE.md`.
