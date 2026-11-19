#!/usr/bin/env bash
# Diagnóstico de SOLO LECTURA de la cadena de notificaciones en un ambiente
# (por defecto testing). No crea, borra ni modifica nada: solo describe y
# lee logs. Se corre en Cloud Shell:
#
#   bash infra/pubsub/diagnostico.sh            # testing
#   SUFIJO=production bash infra/pubsub/diagnostico.sh
#
# La cadena que revisa, en orden:
#   1. el servicio que produce el evento (education, human-promotion, ...)
#      tiene PUBSUB_PROJECT_ID y publica desde su outbox;
#   2. existen los topics y las suscripciones push hacia notifications;
#   3. Pub/Sub y Cloud Scheduler pueden invocar notifications (run.invoker);
#   4. notifications recibe los push (/eventos/*) y responde 2xx;
#   5. el worker de entregas manda el email (Resend) y la notificación in-app.
set -uo pipefail

PROJECT_ID="${PROJECT_ID:-ceibuenvivir}"
REGION="${REGION:-us-central1}"
SUFIJO="${SUFIJO:-testing}"
HORAS="${HORAS:-6}"
G=(gcloud --project "$PROJECT_ID" --quiet)
DESDE="$(date -u -d "-${HORAS} hours" +%Y-%m-%dT%H:%M:%SZ)"

titulo() { printf '\n==================== %s\n' "$*"; }

titulo "1. Variables de eventos en Cloud Run ($SUFIJO)"
for svc in education human-promotion notifications labor-integration campaigns auth dashboard; do
  name="$svc-$SUFIJO"
  if ! "${G[@]}" run services describe "$name" --region "$REGION" >/dev/null 2>&1; then
    echo "$name: NO EXISTE"
    continue
  fi
  echo "--- $name"
  "${G[@]}" run services describe "$name" --region "$REGION" --format=json |
    jq -r '.status.url as $u | "url=\($u)",
      (.spec.template.spec.containers[0].env[]? |
        select(.name | test("^(PUBSUB_|FIREBASE_PROJECT_ID|RESEND_FROM|EDUCATION_API_URL|NOTIFICATIONS_SERVICE_URL)")) |
        "\(.name)=\(.value // "(secreto)")")'
done

titulo "2. Topics y suscripciones push"
"${G[@]}" pubsub topics list --format='value(name.basename())' | grep -E 'eventos' | sort
echo
"${G[@]}" pubsub subscriptions list --format=json |
  jq -r '.[] | select(.pushConfig.pushEndpoint != null) |
    "\(.name | split("/") | last)\n   topic=\(.topic | split("/") | last)\n   endpoint=\(.pushConfig.pushEndpoint)\n   audience=\(.pushConfig.oidcToken.audience // "-")  sa=\(.pushConfig.oidcToken.serviceAccountEmail // "-")\n   filtro=\(.filter // "-")\n   estado=\(.state // "-")"'

titulo "3. Quién puede invocar notifications-$SUFIJO"
"${G[@]}" run services get-iam-policy "notifications-$SUFIJO" --region "$REGION" \
  --format='table(bindings.role, bindings.members)' 2>&1

titulo "4. Cloud Scheduler (barridos del outbox y de entregas)"
"${G[@]}" scheduler jobs list --location "$REGION" \
  --format='table(name.basename(), schedule, state, httpTarget.uri, lastAttemptTime, status.code)' 2>&1

titulo "5. Mensajes en las colas de fallidos (DLQ, sin confirmar)"
for t in educacion-eventos promocion-humana-eventos usuarios-eventos insercion-laboral-eventos campanas-eventos; do
  echo "--- $t-dlq-revision"
  "${G[@]}" pubsub subscriptions pull "$t-dlq-revision" --limit 3 \
    --format='table(message.attributes.event_type, message.publishTime)' 2>&1 | head -6
done

titulo "6. Push recibidos por notifications-$SUFIJO (últimas ${HORAS} h)"
"${G[@]}" logging read \
  "resource.type=cloud_run_revision AND resource.labels.service_name=notifications-$SUFIJO AND httpRequest.requestUrl:\"/eventos/\" AND timestamp>=\"$DESDE\"" \
  --limit 30 --format='table(timestamp, httpRequest.status, httpRequest.requestUrl)' 2>&1

titulo "7. Logs de la app de notifications-$SUFIJO (eventos, entregas, errores)"
"${G[@]}" logging read \
  "resource.type=cloud_run_revision AND resource.labels.service_name=notifications-$SUFIJO AND timestamp>=\"$DESDE\" AND (jsonPayload.level=\"ERROR\" OR jsonPayload.level=\"WARN\" OR textPayload:(\"evento\" OR \"delivery\" OR \"Warning\" OR \"error\" OR \"push\") OR jsonPayload.msg:(\"evento\" OR \"entrega\" OR \"email\"))" \
  --limit 40 --format='table(timestamp, jsonPayload.level, jsonPayload.msg, textPayload)' 2>&1

titulo "8. Outbox de education-$SUFIJO y human-promotion-$SUFIJO (publicación)"
for svc in education human-promotion; do
  echo "--- $svc-$SUFIJO"
  "${G[@]}" logging read \
    "resource.type=cloud_run_revision AND resource.labels.service_name=$svc-$SUFIJO AND timestamp>=\"$DESDE\" AND (jsonPayload.msg:(\"outbox\" OR \"Pub/Sub\" OR \"publica\" OR \"evento\") OR textPayload:(\"outbox\" OR \"pubsub\" OR \"evento\") OR httpRequest.requestUrl:\"/internal/outbox/flush\")" \
    --limit 20 --format='table(timestamp, httpRequest.status, jsonPayload.level, jsonPayload.msg, textPayload)' 2>&1
done

titulo "Listo"
echo "Pegá toda la salida en la conversación."
