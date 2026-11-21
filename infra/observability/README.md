# Observabilidad — Buen Vivir (HU-56)

Arquitectura híbrida diseñada para **\$0/mes** y **`scale-to-zero` (`--min-instances=0`)** en Cloud Run.

## 1. Desarrollo local (Prometheus + Grafana)

Los servicios exponen `/metrics` (formato Prometheus 0.0.4) únicamente en la red interna `buenvivir-net` cuando `METRICS_ENABLED=true` (nunca expuestos en el API Gateway).

```bash
docker compose --profile observability up -d --build
```

- **Prometheus**: `http://localhost:9090` -> *Status -> Targets* debe listar los 9 servicios en estado `UP` (`auth`, `education`, `labor-integration`, `human-promotion`, `dashboard`, `campaigns`, `notifications`, `rxresume`, `cv-renderer`).
- **Grafana**: `http://localhost:3001` (credenciales por defecto de desarrollo: `GRAFANA_ADMIN_USER` / `GRAFANA_ADMIN_PASSWORD`, por defecto `admin` / `admin`; el dashboard *Buen Vivir - Observabilidad Local (Prometheus)* se provisiona automáticamente).

---

## 2. Nube GCP + Grafana Cloud (sin romper `scale-to-zero`)

En Cloud Run, `METRICS_ENABLED` permanece desactivado y no se realizan scrapes periódicos ni uptime checks directos a servicios privados.

### Paso 1: Habilitar APIs y crear Service Account de solo lectura en GCP (Cloud Shell)

> Requiere permisos de `Project IAM Admin` / `Owner` en el proyecto `ceibuenvivir`.

```bash
gcloud services enable monitoring.googleapis.com logging.googleapis.com --project=ceibuenvivir

gcloud iam service-accounts create grafana-monitoring-reader \
  --display-name="Grafana Cloud Read-Only Monitoring" \
  --project=ceibuenvivir

SA="grafana-monitoring-reader@ceibuenvivir.iam.gserviceaccount.com"
gcloud projects add-iam-policy-binding ceibuenvivir --member="serviceAccount:$SA" --role="roles/monitoring.viewer"
gcloud projects add-iam-policy-binding ceibuenvivir --member="serviceAccount:$SA" --role="roles/logging.viewer"
gcloud iam service-accounts keys create ~/grafana-reader-key.json --iam-account="$SA"

# Descarga el archivo desde Cloud Shell a tu computadora para subirlo a Grafana Cloud:
cloudshell download ~/grafana-reader-key.json
```

### Paso 2: Crear Log-based Metrics, alertas pasivas y panel en Cloud Monitoring

Desde el repositorio actualizado en Cloud Shell (requiere `Monitoring Admin` y `Logging Config Writer`):

```bash
git clone https://github.com/leomorabito02/Sistema-informatico-Buen-Vivir.git
cd Sistema-informatico-Buen-Vivir
# Si ya estaba clonado: git pull

PROJECT_ID=ceibuenvivir ENV_SUFFIX=testing ALERT_EMAIL=tu-email@ejemplo.com \
  bash infra/monitoring/setup.sh
```

- Repetir con `ENV_SUFFIX=prod` cuando se despliegue producción.
- Si anteriormente se había ejecutado una versión vieja de `setup.sh` que creaba uptime checks directos por servicio (rompen `scale-to-zero` y dan 403), eliminarlos en *Monitoring -> Uptime checks* de la consola de GCP o vía CLI:
  ```bash
  gcloud monitoring uptime list-configs --format='value(name)'
  # Borrar los que correspondan a servicios privados con:
  # gcloud monitoring uptime delete <CONFIG_ID> --quiet
  ```
- *(Recomendado, \$0)* En GCP *Billing -> Budgets & alerts*, crear una alerta mensual (ej. USD 5) para detectar consumo anómalo si algo impidiera el `scale-to-zero`.

### Paso 3: Configurar Grafana Cloud (Plan Free Forever — URL privada HU-56)

1. Crear cuenta/organización en [grafana.com](https://grafana.com) y provisionar un stack en región **US Central** (cercana a `us-central1`).
2. **Privacidad (HU-56)**: en *Administration -> Users and access*, desactivar auto-registro público e invitar únicamente a los desarrolladores del equipo.
3. **Conectar Google Cloud Monitoring**:
   - *Connections -> Add new connection -> Google Cloud Monitoring*.
   - Elegir *Upload JWT Key File*, subir `grafana-reader-key.json` y presionar **Save & test** (debe confirmar en verde).
4. **Conectar Google Cloud Logging**:
   - *Connections -> Add new connection -> Google Cloud Logging* con el mismo `grafana-reader-key.json` y presionar **Save & test**.
5. **Limpiar la clave descargada** tanto de Cloud Shell (`rm -f ~/grafana-reader-key.json`) como de tu carpeta local de Descargas.
6. **Importar el dashboard**:
   - *Dashboards -> New -> Import* -> subir `infra/observability/grafana/dashboards/gcp/buenvivir-cloudrun-overview.json` y seleccionar el datasource de Google Cloud Monitoring.
   - Usar el selector superior **Entorno** (`testing` / `prod`).
7. **Configurar alertas en Grafana**:
   - *Alerting -> Contact points*: crear canal Email (y opcionalmente Telegram/Discord).
   - *Alerting -> Alert rules*: crear reglas cada 5 min sobre Google Cloud Monitoring (5xx > 5 %, latencia p95 > 2000 ms, `buenvivir_testing_error_logs` > 10 en 5 min).

### Paso 4: Verificación

1. **Datos visibles**: generar una petición al gateway de `testing` y confirmar en Grafana Cloud que aparecen tráfico, latencia y logs JSON con `correlation_id`.
2. **Scale-to-zero intacto**: dejar `testing` sin tráfico 15–20 minutos y verificar en el panel *Instancias activas* que los 9 servicios bajan a `0`.
3. **Alertas**: enviar una notificación de prueba desde *Contact points -> Test*.
