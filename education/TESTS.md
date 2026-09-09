# Suite de pruebas automatizadas — education-service

Este documento explica **qué prueba** la suite de tests de
`education-service`, **cómo está organizada** y **cómo correrla**. Es el
complemento automatizado de `PRUEBAS.md` (guía de pruebas manuales de la
migración a chi) y de `docs/PRUEBAS_AUTH_EDUCATION.md` (pruebas de
integración auth ↔ education con Postman).

> **Origen del diseño.** La suite se construyó **calcando el enfoque de
> `auth/`** (mismos frameworks, misma estructura de carpetas, mismo
> estilo de mocks y de reportes Pact/integridad) para que los dos
> microservicios Go del proyecto se testeen y se midan igual. **No se
> tocó ningún test de `auth/`** — solo se tomó como plantilla.

Última actualización: **2026-09-07**.

---

## 1. De un vistazo

| Capa | Paquete | Nº de tests | Cobertura | Cómo aísla la dependencia externa |
|---|---|---:|---:|---|
| Servicios (lógica de negocio) | `internal/services` | 74 | **91.8%** | Mocks de repositorios (`testify/mock`) |
| Controladores (handlers HTTP) | `internal/controllers` | 45 | **75.4%** | Mocks de servicios + `httptest` |
| Middlewares (JWT, RBAC, CORS, logging…) | `internal/middlewares` | 33 | **100%** | `httptest` + mock de `JWTService` |
| Repositorios Contentful | `internal/repositories` | 32 | **84.9%** | `httptest.Server` haciendo de CDA/CMA/Upload API |
| Repositorios Firestore | `internal/repositories` | 12 | *(incluido arriba)* | **Emulador de Firestore** |
| Contrato (Pact-style) | `tests/contract` | 1 | — | — (genera reporte) |
| Integridad de datos | `tests/database` | 1 | — | — (genera reporte) |
| **Total** | | **~198** | **74.0% global** | |

El pipeline (`.github/workflows/pipeline.yml`, job `ci-education`) exige
**≥ 70% de cobertura global** — el mismo umbral que `ci-auth` — y falla
el build si no se alcanza.

---

## 2. Herramientas

Idénticas a las de `auth/`:

- **`testing`** de la stdlib + **`testify`** (`assert`, `require`,
  `mock`) — único dependency de test agregado a `go.mod`
  (`github.com/stretchr/testify`).
- **`net/http/httptest`** para los handlers y para simular las APIs de
  Contentful.
- **Emulador de Firestore** (`cloud-firestore-emulator`, se levanta en
  CI como contenedor) para los repositorios que hablan con Firestore.
- **k6** (carga), **gremlins** (mutación), **gosec/govulncheck/ZAP**
  (seguridad) — vía los scripts `run_*.sh` (ver §6).

---

## 3. Qué prueba cada capa

### 3.1 Servicios — `internal/services/*_service_test.go`

Es la capa con más tests porque concentra las **reglas de negocio**. Cada
test crea el servicio real inyectándole **mocks de los repositorios**
(las interfaces `CourseRepository`, `ActivityRepository`, etc. existen
justamente para esto). No hay red ni I/O.

Ejemplos de lo que se verifica:

- **`course_service_test.go`** (21 tests): validación de `mode`
  (`Presencial`/`Virtual`/`Mixta`, casing exacto), `trainingAreaId` y al
  menos un `materialId` obligatorios, que `responsibleTeacher` del body
  se **ignore y se fuerce** al `sub` del docente en el flujo
  `/api/docente`, `VerifyOwnership` devolviendo 404 vs 403 según el curso
  esté publicado o sea de otro docente.
- **`material_service_test.go`** / **`training_area_service_test.go`**:
  validación de `type` (`PDF`/`Video`/`Link`), auto-publicación.
- **`attendance_service_test.go`** / **`progress_service_test.go`**:
  cálculo de `porcentajeAsistencia` (asistencias `presente=true` /
  total de actividades del curso), manejo de curso sin actividades.
- **`jwt_service_test.go`**: HS256, rechazo de `type != "access"`,
  rechazo de tokens firmados con otro secreto o expirados, fail-fast si
  `JWT_SECRET` es un default inseguro.

### 3.2 Controladores — `internal/controllers/*_controller_test.go`

Prueban el **handler HTTP** end-to-end dentro del proceso: se arma una
request con `httptest.NewRequest`, se la pasa por el handler con un
**mock del servicio**, y se asertan status code y body JSON.

Cubren: códigos de estado (200/201/400/403/404/409/500), mapeo de
errores de repositorio a HTTP (`ErrNotFound` → 404,
`ContentfulAPIError{409}` → 409, vía `writeError` en `http_helpers.go`),
parseo de path params de chi (`{id}`, `{cursoId}`, `{actividadId}`),
y lectura del `sub`/`role` del contexto que dejó el `AuthMiddleware`.

`testhelpers_test.go` centraliza los errores de repo reutilizables.

### 3.3 Middlewares — `internal/middlewares/*_test.go`

Cobertura **100%**. Uno por archivo de middleware:

| Test | Qué asegura |
|---|---|
| `auth_middleware_test.go` | Extrae el JWT de cookie **o** header `Bearer`; 401 si falta o es inválido; puebla `UserIDKey`/`RoleIDKey`/`IsExternalKey` en el contexto |
| `role_middleware_test.go` | `RequireRole` deja pasar el rol correcto, 403 al incorrecto, **bypass** de `role_admin`/`role_dev`, "sin roles configurados = pasa cualquiera" |
| `external_teacher_middleware_test.go` | `RequireInternalTeacher`: 403 si `is_external=true` |
| `rate_limiter_test.go` | `RateLimiter` / `StrictRateLimiter` responden 429 al exceder la cuota |
| `require_permission_test.go` | `RequirePermissionMiddleware` (copiado de auth, aún sin usar): 401 sin user/rol en contexto, 403 si el `PermissionService` niega o falla, `next` si concede |
| `setup_test.go` | `SetupMiddlewares` agrega los headers OWASP (`X-Content-Type-Options`, `X-Frame-Options`, HSTS, CSP, `Referrer-Policy`, `Permissions-Policy`) y el CORS refleja solo los orígenes de `CORS_ALLOWED_ORIGINS` |
| `logging_test.go` / `middlewares_test.go` / `recoverer_test.go` | Helpers de `log/slog`: `Logging` preserva status, `Recoverer` convierte un panic en 500 JSON, `Chain` aplica middlewares en orden, `AuthHeaders` lee los headers `X-User-*` del gateway |

### 3.4 Repositorios que hablan con Contentful

`course_repository_test.go`, `material_repository_test.go`,
`material_repository_upload_test.go`, `training_area_repository_test.go`,
`contentful_client_helpers_test.go`, `contentful_fields_test.go`.

**Cómo se aísla Contentful:** el `ContentfulClient` tiene los hosts
(`cdaBaseURL`, `cmaBaseURL`, `uploadBaseURL`) como **campos** con los
hosts reales como default. Los tests los apuntan a un
`httptest.Server` local que hace de Content Delivery API / Content
Management API / Upload API. **Nunca se habla por red con Contentful ni
hacen falta tokens reales.**

Se verifica, entre otras cosas:

- El **versionado optimista** de la CMA: `Update`/`Publish` hacen primero
  un `GET` para leer `sys.version` y lo mandan en `X-Contentful-Version`.
- **Borrar = despublicar + borrar** (`DeleteEntryCascade`), tolerando
  404 en el despublicar (idempotente).
- Resolución de `includes` (referencias `trainingArea`/`material`/
  `coverImage`) al leer de la CDA.
- El **flujo de 5 pasos de subida de Asset** (`UploadAndCreate`): subir
  binario → crear Asset → procesar → poll → publicar.
- Mapeo de errores: un 422/404/409 de Contentful se propaga como
  `*ContentfulAPIError` con su `StatusCode`.
- Helpers de parsing (`rawString`, `rawLink`, `rawLinks`,
  `unwrapLocale`, `flattenCMAFields`) con sus casos borde.

### 3.5 Repositorios que hablan con Firestore

`activity_repository_test.go`, `attendance_repository_test.go`,
`enrollment_repository_test.go`, `firestore_client_test.go`.

**Cómo se aísla Firestore:** contra el **emulador oficial de Firestore**,
no contra un proyecto real de Google Cloud. `firestore_testhelpers_test.go`
se conecta vía la variable `FIRESTORE_EMULATOR_HOST`; **si la variable no
está o el emulador no responde, estos tests hacen `t.Skip`** (igual que
los tests de `auth/` que necesitan Postgres). En CI el emulador siempre
está levantado, así que ahí sí corren y cuentan para la cobertura.

Se verifica:

- **`AttendanceRepository.RegisterBatch`**: id de documento
  determinístico `{actividadId}_{participanteId}` — volver a registrar al
  mismo participante en la misma actividad **actualiza, no duplica**; el
  lote se escribe en un `WriteBatch` atómico; input vacío = no-op.
- **`ActivityRepository`**: `Create` setea `created_at` y devuelve el id
  del documento; `Get` devuelve `ErrNotFound` si no existe; `ListByCourse`
  filtra por `curso_id`.
- **`EnrollmentRepository.ListByCourse`**: solo lectura, filtra por
  `curso_id` (la escritura de `inscripciones` pertenece a otro módulo).
- **`NewFirestoreClient`**: detecta si `GOOGLE_APPLICATION_CREDENTIALS`
  es un path o el JSON crudo de la service account, y en el segundo caso
  limpia la variable de entorno.

### 3.6 Contrato e integridad — `tests/`

Mismo formato que `auth/tests/` (no corren un verificador real, dejan un
**artefacto JSON para el pipeline**):

- **`tests/contract/education_contract_test.go`**: lista el contrato
  provider ↔ consumer (Frontend Angular) endpoint por endpoint y escribe
  `reports/pact/education-pact-report.json`.
- **`tests/database/firestore_integrity_test.go`**: documenta las reglas
  de integridad que garantiza el código (id determinístico de
  asistencia, consistencia de `curso_id` entre colecciones, frontera de
  solo-lectura de `inscripciones`) y escribe
  `reports/database/education-firestore-integrity.json`.

---

## 4. Cómo correr la suite

### 4.1 Rápido (sin Firestore) — para iterar en services/controllers/middlewares

```bash
cd education
go test ./...
```

Los tests de los repos de Firestore se **saltan** (`SKIP`). Todo lo demás
corre. Útil para desarrollo día a día.

### 4.2 Completo (con el emulador de Firestore) — como en CI

Opción A — con el jar del emulador (necesita Java 11+):

```bash
cd education

# 1. bajar y levantar el emulador (una vez)
curl -sSL -o /tmp/firestore-emulator.jar \
  https://storage.googleapis.com/firebase-preview-drop/emulator/cloud-firestore-emulator-v1.19.8.jar
java -jar /tmp/firestore-emulator.jar --host=127.0.0.1 --port=8390 &

# 2. correr la suite apuntando a él
export FIRESTORE_EMULATOR_HOST=127.0.0.1:8390
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | tail -1        # -> total: 74.0%
```

Opción B — con Docker (igual que `run_security_tests.sh`):

```bash
docker run --rm -d --name education_test_firestore -p 8390:8080 mtlynch/firestore-emulator:latest
export FIRESTORE_EMULATOR_HOST=localhost:8390
go test ./...
docker rm -f education_test_firestore
```

### 4.3 Ver la cobertura en HTML

```bash
go tool cover -html=coverage.out
```

---

## 5. Cómo lo corre el pipeline

`.github/workflows/pipeline.yml` → job **`ci-education`** (espejo de
`ci-auth`):

1. `golangci-lint` sobre `education/`.
2. Levanta el **emulador de Firestore** (contenedor) y exporta
   `FIRESTORE_EMULATOR_HOST`.
3. `go test -race -coverprofile=coverage.out -covermode=atomic ./...`
   + `go tool cover -func`.
4. **Enforce coverage ≥ 70%** — falla el job si baja.
5. Contract tests (`go test ./tests/contract/...`).
6. `gosec` (SARIF) + `govulncheck`.
7. `docker build` de la imagen + **Trivy** (bloquea CRITICAL/HIGH).
8. En push a `main`/`testing`: push de la imagen a Artifact Registry.

El job se suma a `ci-gate` (que ahora también espera a `ci-education`) y
al `cd-education` de la etapa de deploy.

---

## 6. Scripts auxiliares (no van en `go test`)

Espejo de los de `auth/`, en la raíz de `education/`:

| Script | Qué hace |
|---|---|
| `run_load_tests.sh` | Levanta el server y corre `load_test.js` con **k6** (umbrales de latencia y error). |
| `run_mutation_tests.sh` | **gremlins** (`gremlins unleash`) — mutation testing sobre `internal/...`. |
| `run_security_tests.sh` | **gosec** (SAST) + **govulncheck** (SCA) + **OWASP ZAP** baseline (DAST) contra el server levantado con emulador de Firestore + credenciales dummy de Contentful. |

---

## 7. Qué **no** cubre la suite (a propósito)

- **`cmd/server/main.go`, `internal/config`, `internal/routes`,
  `internal/models`**: 0% — igual que en `auth/`. Son wiring y structs;
  el árbol de rutas se valida en las pruebas manuales de `PRUEBAS.md` y
  en el contract test.
- **Contentful real**: ningún test pega contra la API de Contentful de
  verdad. Los errores de configuración del Space (Environment `testing`
  inexistente, token sin acceso a la organización) se documentan en
  `POSTMAN_EXAMPLES.md` → Troubleshooting, no se testean.
- **El poll de "asset no terminó de procesarse"** en `UploadAsset`: son
  25 intentos × 800 ms; testear el timeout tardaría 20 s, así que solo se
  cubre el camino feliz y el fallo de subida del binario.
- **Integración auth ↔ education con JWT real**: eso es
  `docs/PRUEBAS_AUTH_EDUCATION.md` (manual, con usuarios reales de la DB
  de auth).
