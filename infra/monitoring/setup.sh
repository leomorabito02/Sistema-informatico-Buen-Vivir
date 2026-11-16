#!/usr/bin/env bash
# HU-56: módulo de observabilidad en Cloud Monitoring (panel privado por
# servicio + alertas). Idempotente: se puede correr varias veces.
#
# Uso (Cloud Shell, con permisos de Monitoring Admin en el proyecto):
#   PROJECT_ID=ceibuenvivir ENV_SUFFIX=testing ALERT_EMAIL=equipo@ejemplo.com \
#     bash infra/monitoring/setup.sh
#
# Crea:
#   1. Un canal de notificación por email (ALERT_EMAIL).
#   2. Un uptime check por servicio (GET a su endpoint de salud).
#   3. Tres alertas: servicio caído (uptime), tasa de errores 5xx > 5 % y
#      latencia p95 > 2 s, sobre todos los servicios Cloud Run del sufijo.
#   4. El panel "Buen Vivir - <sufijo>" (dashboard.json): pedidos, errores
#      5xx, latencia p95, instancias, CPU y memoria por servicio.
# La URL privada del panel la imprime al final (solo la ven cuentas del
# proyecto con permiso de lectura de Monitoring).
set -euo pipefail

PROJECT_ID="${PROJECT_ID:?definí PROJECT_ID}"
ENV_SUFFIX="${ENV_SUFFIX:-testing}"
ALERT_EMAIL="${ALERT_EMAIL:?definí ALERT_EMAIL}"
REGION="${REGION:-us-central1}"
DIR="$(cd "$(dirname "$0")" && pwd)"

# servicio:ruta de salud
SERVICES=(
  "auth:/health"
  "education:/api/salud"
  "human-promotion:/health"
  "labor-integration:/health"
  "campaigns:/health"
  "dashboard:/health"
  "notifications:/health"
  "rxresume:/health"
)

gcloud config set project "$PROJECT_ID" >/dev/null

# 1. Canal de email (se reutiliza si ya existe).
CHANNEL="$(gcloud beta monitoring channels list \
  --filter="type=email AND labels.email_address=$ALERT_EMAIL" \
  --format='value(name)' | head -n1)"
if [[ -z "$CHANNEL" ]]; then
  CHANNEL="$(gcloud beta monitoring channels create \
    --display-name="Buen Vivir alertas ($ENV_SUFFIX)" \
    --type=email --channel-labels="email_address=$ALERT_EMAIL" \
    --format='value(name)')"
fi
echo "canal: $CHANNEL"

# 2. Uptime checks.
for entry in "${SERVICES[@]}"; do
  svc="${entry%%:*}"; path="${entry#*:}"
  name="$svc-$ENV_SUFFIX"
  url="$(gcloud run services describe "$name" --region="$REGION" --format='value(status.url)' 2>/dev/null || true)"
  if [[ -z "$url" ]]; then
    echo "AVISO: no existe el servicio Cloud Run $name, se saltea"; continue
  fi
  host="${url#https://}"
  if gcloud monitoring uptime list-configs --filter="displayName=\"uptime $name\"" --format='value(name)' | grep -q .; then
    echo "uptime $name: ya existe"; continue
  fi
  gcloud monitoring uptime create "uptime $name" \
    --resource-type=uptime-url --resource-labels="host=$host,project_id=$PROJECT_ID" \
    --path="$path" --protocol=https --period=5 --timeout=10 >/dev/null
  echo "uptime $name: creado ($path)"
done

# 3. Alertas (se reemplazan si ya existen, por nombre).
crear_alerta() { # $1=nombre $2=archivo de política
  local existing
  existing="$(gcloud alpha monitoring policies list --filter="displayName=\"$1\"" --format='value(name)' | head -n1)"
  if [[ -n "$existing" ]]; then
    gcloud alpha monitoring policies delete "$existing" --quiet >/dev/null
  fi
  gcloud alpha monitoring policies create --policy-from-file="$2" \
    --notification-channels="$CHANNEL" >/dev/null
  echo "alerta: $1"
}

tmp="$(mktemp -d)"
filtro_svc="resource.type=\"cloud_run_revision\" AND resource.labels.service_name=monitoring.regex.full_match(\".*-$ENV_SUFFIX\")"

cat > "$tmp/caido.json" <<EOF
{
  "displayName": "Buen Vivir $ENV_SUFFIX - servicio caído",
  "combiner": "OR",
  "conditions": [{
    "displayName": "uptime check falla",
    "conditionThreshold": {
      "filter": "metric.type=\"monitoring.googleapis.com/uptime_check/check_passed\" AND resource.type=\"uptime_url\"",
      "comparison": "COMPARISON_LT", "thresholdValue": 1, "duration": "300s",
      "aggregations": [{"alignmentPeriod": "300s", "perSeriesAligner": "ALIGN_FRACTION_TRUE", "crossSeriesReducer": "REDUCE_MEAN", "groupByFields": ["resource.label.host"]}]
    }
  }]
}
EOF

cat > "$tmp/errores.json" <<EOF
{
  "displayName": "Buen Vivir $ENV_SUFFIX - errores 5xx",
  "combiner": "OR",
  "conditions": [{
    "displayName": "más de 5 % de respuestas 5xx",
    "conditionMonitoringQueryLanguage": {
      "query": "fetch cloud_run_revision | metric 'run.googleapis.com/request_count' | filter resource.service_name =~ '.*-$ENV_SUFFIX' | align rate(5m) | group_by [resource.service_name], [total: sum(val()), err: sum(if(metric.response_code_class = '5xx', val(), 0))] | value [ratio: err / total] | condition ratio > 0.05",
      "duration": "300s"
    }
  }]
}
EOF

cat > "$tmp/latencia.json" <<EOF
{
  "displayName": "Buen Vivir $ENV_SUFFIX - latencia p95",
  "combiner": "OR",
  "conditions": [{
    "displayName": "p95 mayor a 2 s",
    "conditionThreshold": {
      "filter": "metric.type=\"run.googleapis.com/request_latencies\" AND $filtro_svc",
      "comparison": "COMPARISON_GT", "thresholdValue": 2000, "duration": "300s",
      "aggregations": [{"alignmentPeriod": "300s", "perSeriesAligner": "ALIGN_PERCENTILE_95", "crossSeriesReducer": "REDUCE_MAX", "groupByFields": ["resource.label.service_name"]}]
    }
  }]
}
EOF

crear_alerta "Buen Vivir $ENV_SUFFIX - servicio caído" "$tmp/caido.json"
crear_alerta "Buen Vivir $ENV_SUFFIX - errores 5xx" "$tmp/errores.json"
crear_alerta "Buen Vivir $ENV_SUFFIX - latencia p95" "$tmp/latencia.json"

# 4. Panel.
sed "s/__SUFIJO__/$ENV_SUFFIX/g" "$DIR/dashboard.json" > "$tmp/dashboard.json"
existing="$(gcloud monitoring dashboards list --filter="displayName=\"Buen Vivir - $ENV_SUFFIX\"" --format='value(name)' | head -n1)"
if [[ -n "$existing" ]]; then
  gcloud monitoring dashboards delete "$existing" --quiet >/dev/null
fi
DASH="$(gcloud monitoring dashboards create --config-from-file="$tmp/dashboard.json" --format='value(name)')"
rm -rf "$tmp"

echo
echo "Panel privado: https://console.cloud.google.com/monitoring/dashboards/builder/${DASH##*/}?project=$PROJECT_ID"
