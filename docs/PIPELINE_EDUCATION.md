# education-service en el pipeline de CI/CD

Qué se agregó a `.github/workflows/pipeline.yml` para `education-service`,
y **qué falta configurar por fuera del repo** (GitHub Secrets + recursos
de GCP) para que corra de punta a punta.

> El pipeline se armó **calcando los jobs de `auth`** (`ci-auth` /
> `cd-auth`). Nada de `auth` se tocó.

---

## 1. Jobs nuevos

| Job | Etapa | Qué hace | Bloqueante |
|---|---|---|---|
| `ci-education` | CI | golangci-lint · `go test -race` + cobertura (**gate 70%**) con **emulador de Firestore** como service container · contract test · gosec · govulncheck · `docker build` · Trivy · push de imagen a Artifact Registry (solo en `main`/`testing`) | **Sí** — entra al `ci-gate` |
| `mutation-education` | Mutación | gremlins sobre `internal/...` | No (`if: false`, igual que `mutation-auth`) |
| `cd-education` | CD | Captura la revisión anterior (rollback) · deploy a Cloud Run `education-<suffix>` · 100% de tráfico a la nueva revisión | Sí |
| *(en `cd-smoke`)* | CD | Smoke test `GET /api/salud` con ID token, verifica que `contentfulEnvironment` sea el esperado | Sí |
| *(en `cd-rollback`)* | CD | Rollback de `education-<suffix>` a la revisión previa si falla smoke/DAST/k6 | — |

`ci-gate`, `cd-setup` (verificación de imagen), `cd-api-gateway`,
`cd-smoke`, `cd-dast-zap` y `cd-rollback` se actualizaron para incluir
education.

**API Gateway de GCP:** `api-gateway/openapi.yaml` ahora incluye las 18
rutas de education apuntando a `${EDUCATION_SERVICE_URL}` (generadas con
`scripts/historico-openapi/add_education_routes.py`, idempotente). `cd-api-gateway` resuelve esa
URL desde Cloud Run igual que la de auth. Ver §4.

**Dependencia bumpeada:** `google.golang.org/grpc` `v1.82.1 → v1.83.2`
en `education/go.mod` para cerrar `CVE-2026-84445` (HIGH), que el paso
de Trivy de `ci-education` bloquea. `auth` tiene la **misma CVE** en
`v1.83.1` (bloquea `ci-auth`, que es pre-requisito del `ci-gate`) — hay
que bumpear también `auth/go.mod` a `v1.83.2`
(`cd auth && go get google.golang.org/grpc@v1.83.2 && go mod tidy`).

**Dockerfile:** `education/Dockerfile` se alineó con `auth/Dockerfile`
(que ya pasa el scan de Trivy): agrega `apk --no-cache upgrade` en la
etapa de runtime — sin eso, la capa `alpine:latest` quedaba con la
`openssl` del snapshot de la imagen base (3.5.7-r0, `CVE-2026-14456`
HIGH) en vez de la parcheada del repo (3.5.8-r0). De paso trae el resto
del hardening de auth (usuario no-root `appuser`, `tzdata`, binario
stripped con `-ldflags="-s -w"`).

**Lint:** `ci-education` es la primera vez que education pasa por
`golangci-lint`, así que hubo que limpiar ~23 hallazgos pre-existentes
(errores de `Close()`/`Write()` sin chequear → `_ =`; deprecaciones
`SA1019` de `httprate.LimitByIP` / `middleware.RealIP` /
`option.WithCredentialsJSON` / `firestore.Batch` → `//nolint:staticcheck`
con el mismo criterio que auth ya usa; un `switch` a tagged switch; una
cadena de error capitalizada). Todo verificado local con
`golangci-lint v2.13.1` (el mismo del pipeline): **0 issues**.

---

## 2. GitHub Secrets

### 2.1 Ya existen (los usa `auth`, education los reusa) — no hay que hacer nada

`GCP_WORKLOAD_IDENTITY_PROVIDER`, `GCP_SERVICE_ACCOUNT`,
`GCP_PROJECT_ID`, `GCP_REGION`, `ARTIFACT_REGISTRY_HOST`,
`ARTIFACT_REGISTRY_REPO`, `JWT_SECRET`, `GOOGLE_APPLICATION_CREDENTIALS`
(JSON de la service account), `CORS_ALLOWED_ORIGINS`.

> `JWT_SECRET` **tiene que ser el mismo** que usa `auth` — education solo
> valida tokens emitidos por auth. Ya lo es (mismo secret).

### 2.2 Hay que crearlos — Contentful (Settings → Secrets and variables → Actions)

| Secret | Valor | De dónde sale |
|---|---|---|
| `CONTENTFUL_SPACE_ID` | Space ID | Contentful → Settings → General settings |
| `CONTENTFUL_CDA_TOKEN` | Content Delivery API — access token | Contentful → Settings → API keys |
| `CONTENTFUL_CMA_TOKEN` | Content Management API — personal access token | Contentful → Settings → API keys → Content management tokens |
| `CONTENTFUL_LOCALE` | `en-US` (o el locale real del Space) | Contentful → Settings → Locales |

Son **los mismos para TESTING y PROD** (mismo Space; lo único que cambia
es `CONTENTFUL_ENVIRONMENT`, que lo setea el pipeline: `testing` en la
rama `testing`, `master` en `main` — ver `cd-setup`). Por eso pueden ir
como **secrets de repo**, no por Environment.

> El `CMA_TOKEN` da permiso de **escritura** sobre el contenido. Si se
> quiere aislar prod de testing de verdad, crear dos Spaces separados y
> usar Environments de GitHub con distinto `CONTENTFUL_SPACE_ID` — hoy el
> diseño asume un Space único con dos Environments internos.

### 2.3 Environments de GitHub

`PROD` y `TESTING` ya existen (los usa `cd-auth`). `cd-education` usa los
mismos. No hay que crear nada; si los secrets de Contentful se ponen a
nivel repo, tampoco hace falta agregarlos a los Environments.

---

## 3. Recursos de GCP

Todo en el proyecto `${GCP_PROJECT_ID}` (el mismo de auth):

| Recurso | Estado | Acción |
|---|---|---|
| **Artifact Registry** repo `${ARTIFACT_REGISTRY_REPO}` | Ya existe | Ninguna — la imagen se sube como `.../education:<sha>` en el repo existente |
| **Cloud Run** `education-testing` / `education-prod` | No existen | Se crean solos en el primer deploy (`deploy-cloudrun` hace create-or-update). La service account de deploy ya tiene `roles/run.admin` |
| **Firestore** (base `(default)`) | Ya existe (proyecto `ceibuenvivir`, lo usa el módulo) | Verificar que **`${GCP_SERVICE_ACCOUNT}` tenga `roles/datastore.user`** — education corre *como* esa SA (`--service-account`) y necesita leer/escribir `actividades`, `asistencias`, `inscripciones` |
| **Workload Identity** | Ya configurado para auth | Ninguna — mismo provider/SA |
| Invoker del gateway → education | Pendiente (§4) | Cuando se rutee por el gateway, dar `roles/run.invoker` sobre `education-*` a la SA del backend del gateway |

Cómo verificar el permiso de Firestore:

```bash
gcloud projects get-iam-policy "$GCP_PROJECT_ID" \
  --flatten="bindings[].members" \
  --filter="bindings.members:$GCP_SERVICE_ACCOUNT AND bindings.role:roles/datastore.user" \
  --format="value(bindings.role)"
# si no imprime nada:
gcloud projects add-iam-policy-binding "$GCP_PROJECT_ID" \
  --member="serviceAccount:$GCP_SERVICE_ACCOUNT" \
  --role="roles/datastore.user"
```

---

## 4. API Gateway de Google Cloud (hecho)

`api-gateway/openapi.yaml` ahora define las **18 rutas de education**
apuntando a `${EDUCATION_SERVICE_URL}` (mismo patrón que auth:
`x-google-backend` con `APPEND_PATH_TO_ADDRESS` + `jwt_audience`, y un
`options` por path para el preflight CORS; el gateway **no** valida el
JWT, cada servicio valida el suyo).

Se generaron con **`scripts/historico-openapi/add_education_routes.py`** (raíz del repo,
idempotente — volver a correrlo reemplaza las rutas de education, no las
duplica). Rutas cubiertas: `/api/salud`, `/api/cursos` (+`/borradores`,
`/{id}`, `/{id}/publicar`, `/{id}/despublicar`), `/api/areas-formativas`
(+`/{id}`), `/api/materiales` (+`/upload`, `/{id}`), y todo
`/api/docente/*` explícito (el gateway v2 no soporta wildcards).

Cambios de pipeline que lo acompañan:

- **`ci-api-gateway`**: exporta un `EDUCATION_SERVICE_URL` dummy antes del
  `envsubst` para que `swagger-cli validate` pase.
- **`cd-api-gateway`**: ahora `needs: cd-education`, resuelve
  `EDUCATION_SERVICE_URL` con
  `gcloud run services describe "education-$SUFFIX" --format='value(status.url)'`
  y aborta si esa URL o la de auth vienen vacías.
- **`cd-dast-zap`**: `needs: cd-education` — el escaneo ZAP ahora también
  cubre los endpoints de education a través del gateway. Usa el mismo
  `auth/zap.conf` (rules compartidas). Si aparecen alertas nuevas
  específicas de education, se tunean ahí.

**Lo único que queda para quien tenga acceso a GCP:** dar
`roles/run.invoker` sobre `education-testing` / `education-prod` a la
service account del backend del gateway (la misma
`--backend-auth-service-account` que ya usa `cd-api-gateway`, hoy
`${GCP_SERVICE_ACCOUNT}`), si no ya la tiene a nivel proyecto:

```bash
for SUFFIX in testing prod; do
  gcloud run services add-iam-policy-binding "education-$SUFFIX" \
    --region="$GCP_REGION" \
    --member="serviceAccount:$GCP_SERVICE_ACCOUNT" \
    --role="roles/run.invoker"
done
```

> **Nota ESPv2:** `/api/cursos/borradores` y `/api/cursos/{id}` conviven
> (segmento literal vs. variable) — el gateway prioriza el literal. Si
> alguna vez API Gateway rechaza la config por "conflicting paths", la
> salida es mover `borradores` a otro prefijo; hoy no debería pasar.

---

## 5. Checklist para quien tenga los accesos

- [ ] Crear los 4 secrets `CONTENTFUL_*` en GitHub (§2.2)
- [ ] Confirmar que `${GCP_SERVICE_ACCOUNT}` tiene `roles/datastore.user` (§3)
- [ ] Dar `roles/run.invoker` sobre `education-*` a la SA del gateway (§4)
- [ ] Push a `testing` → `ci-education` pasa (cobertura ≥ 70%),
      `cd-education` crea `education-testing`, `cd-api-gateway` re-despliega
      el gateway con las rutas de education
- [ ] `curl` con ID token a `https://education-testing-…run.app/api/salud`
      → `200` con `"contentfulEnvironment":"testing"`
- [ ] `curl https://<gateway-host>/api/salud` → `200` (education vía gateway)
