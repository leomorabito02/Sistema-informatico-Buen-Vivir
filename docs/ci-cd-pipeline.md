# Documentación de los Pipelines de CI/CD

Esta documentación describe el funcionamiento, la arquitectura y el flujo de trabajo de los pipelines de Integración Continua (CI) y Despliegue Continuo (CD) para el proyecto **Sistema Informático Buen Vivir**.

---

## 1. Principio Arquitectónico Fundacional

El diseño de estos pipelines sigue el patrón empresarial **Build-Once & Promote-Everywhere**:

1. **El pipeline de CI es el único responsable de validar y construir artefactos.**
2. **El pipeline de CD no recompila ni vuelve a generar ningún binario o asset.** En su lugar, recupera los artefactos exactos validados y publicados por el CI durante el commit SHA correspondiente.
3. **Garantía de consistencia:** El código frontend y backend que se despliega en el entorno de `TEST` es idéntico byte por byte al que se promueve a `PROD`.

---

## 2. Diagrama del Flujo Completo

```
                Pull Request / Push
                          │
                          ▼
                     CI Pipeline
    ──────────────────────────────────────────────
    Stage 1: Checkout & Dependencias (Go + Node)
    Stage 2: Calidad (Golangci-lint, ESLint, Prettier)
    Stage 3: Testing & Cobertura (go test, ng test)
    Stage 4: Seguridad (Semgrep, Gosec, Govulncheck, npm audit, Gitleaks)
    Stage 5: Build (Docker Backend, Angular Production)
    Stage 6: Escaneo Docker (Trivy Scan - Bloqueante si CRITICAL/HIGH)
    Stage 7: Publicación (Artifact Registry + GitHub Artifact)
    ──────────────────────────────────────────────
                          │
             Artefactos versionados por Commit SHA
                          │
              ┌───────────┴───────────┐
              ▼                       ▼
         Push a testing          Push a main
              │                       │
              ▼                       ▼
          CD (TEST)               CD (PROD)
              │                       │
              ▼                       ▼
      Recuperar artefactos     Recuperar artefactos
      (GCP AR + GH Artifact)   (GCP AR + GH Artifact)
              │                       │
              ▼                       ▼
     Deploy Cloud Run TEST    Deploy Cloud Run PROD
     Deploy Cloudflare TEST   Deploy Cloudflare PROD
              │                       │
              ▼                       ▼
     Smoke + Health Checks    Smoke + Health Checks
              │                       │
              └───────────────┬───────┘
                              ▼
                    Rollback automático
                      si el despliegue falla
```

---

## 3. Pipeline de CI (Integración Continua)

Fichero: `.github/workflows/ci.yml`

### Triggers
- **Pull Request** a cualquier rama.
- **Push** a cualquier rama.

### Etapas (Stages)

#### Stage 1 - Preparación
- **Checkout:** Descarga el código del repositorio (`actions/checkout@v4`).
- **Configurar Go:** Inicializa Go 1.22 con caché de módulos activada.
- **Configurar Node:** Inicializa Node.js 20.x con caché de dependencias npm activada.
- **Instalar dependencias:** Ejecuta `npm ci` para el frontend y `go mod download` para los microservicios backend.

#### Stage 2 - Calidad de Código
- **Backend:** Ejecuta `golangci-lint` para auditar sintaxis y calidad Go.
- **Frontend:** Ejecuta ESLint (`npx eslint`) y Prettier (`npx prettier --check`).

#### Stage 3 - Testing y Cobertura
- **Backend:** Ejecuta `go test -v -coverprofile=coverage.out ./...` en los microservicios Go.
- **Frontend:** Ejecuta `ng test --watch=false --browsers=ChromeHeadless --code-coverage`.
- **Cobertura:** Almacena y reporta los indicadores de cobertura de pruebas.

#### Stage 4 - Análisis de Seguridad (SAST, Dependencias y Secretos)
- **SAST Backend/Frontend:**
  - **Semgrep:** Análisis estático multi-lenguaje.
  - **Gosec:** Análisis de seguridad en código Go.
- **Auditoría de Dependencias:**
  - **Govulncheck:** Escaneo de vulnerabilidades conocidas en dependencias Go.
  - **npm audit:** Verificación de vulnerabilidades en paquetes Node.js.
- **Escaneo de Secretos:**
  - **Gitleaks:** Detección de claves API, tokens o credenciales hardcodeadas en el commit.

#### Stage 5 - Construcción (Build)
- **Backend:** Construye la imagen Docker del backend tagged temporalmente con el Commit SHA (`backend:${{ github.sha }}`).
- **Frontend:** Compila la aplicación Angular en modo producción (`ng build --configuration production`).

#### Stage 6 - Escaneo de Imágenes Docker
- **Trivy Image Scan:** Analiza las capas de la imagen Docker generada.
- **Política de bloqueo:** Si se detectan vulnerabilidades clasificadas como `CRITICAL` o `HIGH`, el job falla inmediatamente y se impide la publicación.

#### Stage 7 - Publicación
- Subes el build compilado de Angular como un artefacto persistente de GitHub Actions (`actions/upload-artifact@v4`) identificado por el Commit SHA.

#### Stage 8 - Despliegue Frontend (Cloudflare Pages)
- **Condición:** Se ejecuta únicamente en eventos `push` a las ramas `testing` y `main`.
- **Detección de Entorno:** Asigna el sufijo `test` para la rama `testing` y `prod` para la rama `main`.
- **Despliegue:** Publica automáticamente el directorio `frontend/dist/frontend-buen-vivir/browser` en Cloudflare Pages (`cloudflare/wrangler-action@v3`) hacia el proyecto `${{ secrets.CLOUDFLARE_PROJECT }}-${{ env.SERVICE_SUFFIX }}`.

---

## 4. Pipeline de CD (Integrado)

El pipeline de CD ha sido unificado dentro del archivo `.github/workflows/ci.yml` (Stage 8) para garantizar una ejecución secuencial inmediata y sin fricciones tras la fase de construcción.

### Triggers de Despliegue
- **Push a `testing`:** Publicación automática en el entorno de TEST de Cloudflare Pages (`<CLOUDFLARE_PROJECT>-test`).
- **Push a `main`:** Publicación automática en el entorno de PROD de Cloudflare Pages (`<CLOUDFLARE_PROJECT>-prod`).

---

## 5. Gestión de Secretos y Variables de Entorno

### Configuración en GitHub Repository Secrets

Para habilitar la ejecución correcta de los pipelines, deben configurarse los siguientes secretos en el repositorio:

#### Secretos requeridos por CI:
- `GCP_WORKLOAD_IDENTITY_PROVIDER`: URI del proveedor de Workload Identity en GCP.
- `GCP_SERVICE_ACCOUNT`: Email de la Service Account asignada a GitHub Actions.
- `GCP_PROJECT_ID`: ID del proyecto de Google Cloud Platform.
- `ARTIFACT_REGISTRY`: Dominio/path base de Artifact Registry (ej: `us-central1-docker.pkg.dev/mi-proyecto/repo`).

#### Secretos adicionales requeridos por CD:
- `CLOUDFLARE_API_TOKEN`: Token de API con permisos de edición en Cloudflare Pages.
- `CLOUDFLARE_ACCOUNT_ID`: ID de la cuenta de Cloudflare.
- `CLOUDFLARE_PROJECT`: Nombre del proyecto en Cloudflare Pages.
- `CLOUD_RUN_SERVICE`: Nombre base del servicio en Cloud Run.
- `REGION`: Región geográfica de GCP (ej: `us-central1`).

---

### Buenas Prácticas de Seguridad en Runtime (Google Secret Manager)

> [!IMPORTANT]
> **Ningún secreto de la aplicación (JWT secrets, cadenas de conexión a base de datos, API keys de terceros, etc.) debe living en variables de entorno fijas dentro del repositorio ni en los YAML de GitHub Actions.**

Los secretos de aplicación son inyectados directamente a **Google Cloud Run** en tiempo de ejecución montándolos desde **Google Secret Manager**:

```bash
gcloud run services update mi-servicio \
  --update-secrets=JWT_SECRET=jwt-secret-key:latest,\
DB_PASSWORD=database-password:latest
```

De este modo, el CI/CD solo maneja permisos de infraestructura y despliegue, mientras que la información sensible permanece protegida dentro del enclave de seguridad de GCP.
