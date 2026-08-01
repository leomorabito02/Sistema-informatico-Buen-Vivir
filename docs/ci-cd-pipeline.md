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

## 3. Pipeline Unificado de CI/CD

Fichero: `.github/workflows/ci-cd-pipeline.yml`

### Triggers
- **Pull Request:** Ejecuta únicamente el Job de CI (`build-and-test`).
- **Push a `main` o `testing`:** Ejecuta el Job de CI (`build-and-test`) y posteriormente el Job de CD (`deploy-cloudflare`).

### Estructura de Jobs

#### Job 1: CI (`build-and-test`)
- **Checkout:** Descarga el código del repositorio (`actions/checkout@v4`).
- **Configurar Node:** Inicializa Node.js 22 (`actions/setup-node@v4`).
- **Instalar dependencias:** Ejecuta `npm ci` en la carpeta `frontend`.
- **Calidad de Código:** Ejecuta ESLint (`npx --yes eslint`) y Prettier (`npx --yes prettier --check`).
- **Testing:** Ejecuta `ng test` en modo headless (`npx --yes ng test --watch=false --browsers=ChromeHeadless`).
- **Auditorías de Seguridad:** Auditoría de dependencias (`npm audit`) y escaneo de secretos (`Gitleaks`).
- **Construcción (Build):** Compila el frontend Angular en modo producción (`npx --yes ng build --configuration production`).
- **Artefactos:** Almacena el build compilado (`dist/frontend-buen-vivir/browser`) como artefacto de GitHub Actions.

#### Job 2: CD (`deploy-cloudflare`)
- **Condición:** Requiere el éxito de `build-and-test` (`needs: build-and-test`) y un push a `main` o `testing`.
- **Regla de Ruteo SPA:** Genera el archivo `_redirects` con la regla `/* /index.html 200` para garantizar el soporte de ruteo interno en Angular.
- **Despliegue:** Publica mediante `wrangler pages deploy` directo a Cloudflare Pages hacia el proyecto indicado en `CLOUDFLARE_PROJECT` (ej. `ceibuenvivir`).

---

## 4. Entornos y URLs de Despliegue

| Entorno | Rama | Dominio / URL Oficial | Tipo de Despliegue en Cloudflare |
| :--- | :--- | :--- | :--- |
| **PROD** | `main` | `https://ceibuenvivir.org` (`https://ceibuenvivir.pages.dev`) | Producción |
| **TEST** | `testing` | `https://testing.ceibuenvivir.pages.dev` | Preview / Testing |

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
