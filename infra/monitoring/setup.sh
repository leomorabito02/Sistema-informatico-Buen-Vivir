#!/usr/bin/env bash
# HU-56: módulo de observabilidad en Cloud Monitoring (panel privado por
# servicio + alertas pasivas compatibles con scale-to-zero). Idempotente.
#
# Uso (Cloud Shell, con permisos de Monitoring Admin y Logging Config Writer):
#   PROJECT_ID=ceibuenvivir ENV_SUFFIX=testing ALERT_EMAIL=equipo@ejemplo.com \
#     bash infra/monitoring/setup.sh
#
# Opcional: definir GATEWAY_HEALTH_URL (ej. https://.../api/v1/campaigns/publicas?limit=1)
# para crear un único uptime check liviano sobre el API Gateway (por defecto
# omitido para preservar scale-to-zero al 100 % y evitar costos).
#
# Crea:
#   1. Un canal de notificación por email (ALERT_EMAIL).
#   2. Una Log-based Metric de logs con severity>=ERROR por servicio Cloud Run.
#   3. Tres alertas pasivas (no despiertan instancias en idle):
#      - Tasa de errores 5xx > 5 %
#      - Latencia p95 > 2 s
#      - Picos de errores en logs (> 10 errores en 5 min)
#   4. El panel "Buen Vivir - <sufijo>" (dashboard.json): pedidos, errores 5xx,
#      latencia p95, cold starts, instancias activas, logs ERROR, CPU y memoria
#      para los 9 servicios Cloud Run (incluyendo cv-renderer).
set -euo pipefail

PROJECT_ID="${PROJECT_ID:?definí PROJECT_ID}"
ENV_SUFFIX="${ENV_SUFFIX:-testing}"
ALERT_EMAIL="${ALERT_EMAIL:?definí ALERT_EMAIL}"
REGION="${REGION:-us-central1}"
GATEWAY_HEALTH_URL="${GATEWAY_HEALTH_URL:-}"
DIR="$(cd "$(dirname "$0")" && pwd)"

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

# 2. Log-based Metric para logs JSON con severity>=ERROR (compatible con scale-to-zero).
LOG_METRIC="buenvivir_${ENV_SUFFIX}_error_logs"
LOG_FILTER="resource.type=\"cloud_run_revision\" AND resource.labels.service_name=~\"^.*-${ENV_SUFFIX}$\" AND severity>=ERROR"
if gcloud logging metrics describe "$LOG_METRIC" >/dev/null 2>&1; then
  gcloud logging metrics update "$LOG_METRIC" \
    --description="Logs con severity>=ERROR en servicios Cloud Run ($ENV_SUFFIX)" \
    --log-filter="$LOG_FILTER" >/dev/null
  echo "log-metric $LOG_METRIC: actualizada"
else
  gcloud logging metrics create "$LOG_METRIC" \
    --description="Logs con severity>=ERROR en servicios Cloud Run ($ENV_SUFFIX)" \
    --log-filter="$LOG_FILTER" >/dev/null
  echo "log-metric $LOG_METRIC: creada"
fi

# 3. Alertas pasivas (se reemplazan si ya existen, por nombre).
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
trap 'rm -rf "$tmp"' EXIT
filtro_svc="resource.type=\"cloud_run_revision\" AND resource.labels.service_name=monitoring.regex.full_match(\".*-$ENV_SUFFIX\")"

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

cat > "$tmp/log_errors.json" <<EOF
{
  "displayName": "Buen Vivir $ENV_SUFFIX - errores en logs",
  "combiner": "OR",
  "conditions": [{
    "displayName": "más de 10 logs ERROR en 5 min",
    "conditionThreshold": {
      "filter": "metric.type=\"logging.googleapis.com/user/$LOG_METRIC\" AND resource.type=\"cloud_run_revision\"",
      "comparison": "COMPARISON_GT", "thresholdValue": 10, "duration": "0s",
      "aggregations": [{"alignmentPeriod": "300s", "perSeriesAligner": "ALIGN_SUM", "crossSeriesReducer": "REDUCE_SUM", "groupByFields": ["resource.label.service_name"]}]
    }
  }]
}
EOF

crear_alerta "Buen Vivir $ENV_SUFFIX - errores 5xx" "$tmp/errores.json"
crear_alerta "Buen Vivir $ENV_SUFFIX - latencia p95" "$tmp/latencia.json"
crear_alerta "Buen Vivir $ENV_SUFFIX - errores en logs" "$tmp/log_errors.json"

# Opcional: único uptime check liviano al API Gateway (15 min) si se indica explícitamente.
if [[ -n "$GATEWAY_HEALTH_URL" ]]; then
  gw_no_proto="${GATEWAY_HEALTH_URL#https://}"
  gw_host="${gw_no_proto%%/*}"
  gw_path="/${gw_no_proto#*/}"
  [[ "$gw_host" == "$gw_no_proto" ]] && gw_path="/"
  if ! gcloud monitoring uptime list-configs --filter="displayName=\"uptime gateway-$ENV_SUFFIX\"" --format='value(name)' | grep -q .; then
    gcloud monitoring uptime create "uptime gateway-$ENV_SUFFIX" \
      --resource-type=uptime-url --resource-labels="host=$gw_host,project_id=$PROJECT_ID" \
      --path="$gw_path" --protocol=https --period=15 --timeout=10 >/dev/null
    echo "uptime gateway-$ENV_SUFFIX: creado ($gw_path)"
  fi
fi

# 4. Panel.
sed "s/__SUFIJO__/$ENV_SUFFIX/g" "$DIR/dashboard.json" > "$tmp/dashboard.json"
existing="$(gcloud monitoring dashboards list --filter="displayName=\"Buen Vivir - $ENV_SUFFIX\"" --format='value(name)' | head -n1)"
if [[ -n "$existing" ]]; then
  gcloud monitoring dashboards delete "$existing" --quiet >/dev/null
fi
DASH="$(gcloud monitoring dashboards create --config-from-file="$tmp/dashboard.json" --format='value(name)')"

echo
echo "Panel privado (Cloud Monitoring): https://console.cloud.google.com/monitoring/dashboards/builder/${DASH##*/}?project=$PROJECT_ID"
echo "Grafana Cloud: importar infra/observability/grafana/dashboards/gcp/buenvivir-cloudrun-overview.json"
