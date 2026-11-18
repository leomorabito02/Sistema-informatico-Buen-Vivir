#!/usr/bin/env bash
# Crea la infraestructura de mensajería asíncrona en GCP: topics, DLQs,
# suscripciones push con filtro y OIDC, service accounts, permisos IAM,
# TTL de idempotencia y los barridos del outbox en Cloud Scheduler.
#
# Por qué gcloud y no Terraform: el repo no tiene Terraform ni estado
# remoto, y el pipeline de CD ya despliega con gcloud. Este script es
# IDEMPOTENTE (crea solo lo que falta) y se puede correr de nuevo sin
# romper nada. No toca topics que ya existan, como rxresume-sync-topic.
# Ver docs/adr/0002-push-con-oidc.md y docs/eventos/README.md.
#
# Uso:
#   PROJECT_ID=ceibuenvivir \
#   EDUCATION_URL=https://education-xxxx.run.app \
#   HUMAN_PROMOTION_URL=https://human-promotion-xxxx.run.app \
#   DASHBOARD_URL=https://dashboard-xxxx.run.app \
#   NOTIFICATIONS_URL=https://notifications-xxxx.run.app \
#   EDUCATION_SA=education-run@ceibuenvivir.iam.gserviceaccount.com \
#   HUMAN_PROMOTION_SA=human-promotion-run@ceibuenvivir.iam.gserviceaccount.com \
#   infra/pubsub/setup.sh
#
# Opcionales (cuando Autenticación / Inserción Laboral empiecen a publicar):
#   AUTH_SA=...   LABOR_SA=...
#
# Ojo: el filtro de una suscripción NO se puede cambiar una vez creada.
# Para cambiarlo hay que borrarla y volver a correr el script.
set -euo pipefail

: "${PROJECT_ID:?falta PROJECT_ID}"
: "${EDUCATION_URL:?falta EDUCATION_URL}"
: "${HUMAN_PROMOTION_URL:?falta HUMAN_PROMOTION_URL}"
: "${DASHBOARD_URL:?falta DASHBOARD_URL}"
: "${NOTIFICATIONS_URL:?falta NOTIFICATIONS_URL}"
: "${EDUCATION_SA:?falta EDUCATION_SA (service account de Cloud Run de education)}"
: "${HUMAN_PROMOTION_SA:?falta HUMAN_PROMOTION_SA}"
REGION="${REGION:-us-central1}"

G=(gcloud --project "$PROJECT_ID" --quiet)
PROJECT_NUMBER="$("${G[@]}" projects describe "$PROJECT_ID" --format='value(projectNumber)')"
PUBSUB_AGENT="service-${PROJECT_NUMBER}@gcp-sa-pubsub.iam.gserviceaccount.com"

log() { printf '\n== %s\n' "$*"; }

# ── Topics + dead-letter topics ───────────────────────────────────────
TOPICS=(usuarios-eventos educacion-eventos promocion-humana-eventos contenido-eventos insercion-laboral-eventos campanas-eventos)

ensure_topic() {
  if "${G[@]}" pubsub topics describe "$1" >/dev/null 2>&1; then
    echo "topic $1: ya existe"
  else
    "${G[@]}" pubsub topics create "$1" --message-retention-duration=7d
  fi
}

log "Topics"
for t in "${TOPICS[@]}"; do
  ensure_topic "$t"
  ensure_topic "$t-dlq"
  # Un mensaje que llega a un topic sin suscripciones se pierde: cada DLQ
  # tiene una suscripción pull para revisar lo que falló.
  if ! "${G[@]}" pubsub subscriptions describe "$t-dlq-revision" >/dev/null 2>&1; then
    "${G[@]}" pubsub subscriptions create "$t-dlq-revision" --topic "$t-dlq" --message-retention-duration=7d --expiration-period=never
  fi
  # El agente de Pub/Sub tiene que poder publicar en la DLQ.
  "${G[@]}" pubsub topics add-iam-policy-binding "$t-dlq" --member="serviceAccount:$PUBSUB_AGENT" --role=roles/pubsub.publisher >/dev/null
done

# ── Service accounts dedicadas (una por consumidor) ──────────────────
ensure_sa() { # $1=nombre corto $2=descripción
  local email="$1@${PROJECT_ID}.iam.gserviceaccount.com"
  if ! "${G[@]}" iam service-accounts describe "$email" >/dev/null 2>&1; then
    "${G[@]}" iam service-accounts create "$1" --display-name="$2"
  fi
  # El agente de Pub/Sub firma los tokens OIDC en nombre de esta cuenta.
  "${G[@]}" iam service-accounts add-iam-policy-binding "$email" \
    --member="serviceAccount:$PUBSUB_AGENT" --role=roles/iam.serviceAccountTokenCreator >/dev/null
  echo "$email"
}

log "Service accounts de push"
SA_EDU_USUARIOS="$(ensure_sa sa-push-edu-usuarios 'Push usuarios-eventos -> education')"
SA_EDU_CONTENIDO="$(ensure_sa sa-push-edu-contenido 'Push contenido-eventos -> education')"
SA_PH_EDUCACION="$(ensure_sa sa-push-ph-educacion 'Push educacion-eventos -> human-promotion')"
SA_PH_USUARIOS="$(ensure_sa sa-push-ph-usuarios 'Push usuarios-eventos -> human-promotion')"
SA_DASHBOARD="$(ensure_sa sa-push-dashboard 'Push eventos -> dashboard')"
SA_NOTIF_EDU="$(ensure_sa sa-push-notif-edu 'Push educacion-eventos -> notifications')"
SA_NOTIF_PH="$(ensure_sa sa-push-notif-ph 'Push promocion-humana-eventos -> notifications')"
SA_NOTIF_USUARIOS="$(ensure_sa sa-push-notif-usuarios 'Push usuarios-eventos -> notifications')"
SA_NOTIF_LABOR="$(ensure_sa sa-push-notif-labor 'Push insercion-laboral-eventos -> notifications')"
SA_NOTIF_CAMPANAS="$(ensure_sa sa-push-notif-campanas 'Push campanas-eventos -> notifications')"
SA_SCHEDULER="$(ensure_sa sa-scheduler-outbox 'Cloud Scheduler -> barrido del outbox')"

# ── Suscripciones push ───────────────────────────────────────────────
# Filtros: los mismos que arma events.SubscriptionFilter a partir de los
# tipos que registra cada consumidor.
filter() { local out="" t; for t in "$@"; do out+="${out:+ OR }attributes.event_type = \"$t\""; done; echo "$out"; }

ensure_push_sub() { # $1=nombre $2=topic $3=endpoint $4=sa $5=audience $6=filtro $7=ordering(true|false)
  if "${G[@]}" pubsub subscriptions describe "$1" >/dev/null 2>&1; then
    echo "suscripción $1: ya existe (el filtro no se puede cambiar; para cambiarlo, borrarla)"
  else
    local ordering=()
    [[ "$7" == "true" ]] && ordering=(--enable-message-ordering)
    "${G[@]}" pubsub subscriptions create "$1" \
      --topic "$2" \
      --push-endpoint "$3" \
      --push-auth-service-account "$4" \
      --push-auth-token-audience "$5" \
      --message-filter "$6" \
      --ack-deadline 30 \
      --min-retry-delay 10s --max-retry-delay 600s \
      --dead-letter-topic "$2-dlq" --max-delivery-attempts 5 \
      --expiration-period never \
      "${ordering[@]}"
  fi
  "${G[@]}" pubsub subscriptions add-iam-policy-binding "$1" --member="serviceAccount:$PUBSUB_AGENT" --role=roles/pubsub.subscriber >/dev/null
}

log "Suscripciones"
ensure_push_sub educacion-usuarios usuarios-eventos "$EDUCATION_URL/eventos/usuarios" "$SA_EDU_USUARIOS" "$EDUCATION_URL" \
  "$(filter usuarios.participante.eliminado usuarios.usuario.actualizado)" true
ensure_push_sub educacion-contenido contenido-eventos "$EDUCATION_URL/eventos/contenido" "$SA_EDU_CONTENIDO" "$EDUCATION_URL" \
  "$(filter contenido.despublicado contenido.publicado)" false
ensure_push_sub promocion-humana-educacion educacion-eventos "$HUMAN_PROMOTION_URL/eventos/educacion" "$SA_PH_EDUCACION" "$HUMAN_PROMOTION_URL" \
  "$(filter educacion.asistencia_general.registrada educacion.curso.completado_por_participante educacion.inscripcion.creada)" true
ensure_push_sub promocion-humana-usuarios usuarios-eventos "$HUMAN_PROMOTION_URL/eventos/usuarios" "$SA_PH_USUARIOS" "$HUMAN_PROMOTION_URL" \
  "$(filter usuarios.participante.eliminado)" true
ensure_push_sub dashboard-usuarios usuarios-eventos "$DASHBOARD_URL/eventos" "$SA_DASHBOARD" "$DASHBOARD_URL/eventos" \
  "$(filter usuarios.participante.creado usuarios.participante.eliminado)" false
ensure_push_sub dashboard-educacion educacion-eventos "$DASHBOARD_URL/eventos" "$SA_DASHBOARD" "$DASHBOARD_URL/eventos" \
  "$(filter educacion.curso.completado_por_participante educacion.curso.creado educacion.curso.eliminado educacion.habilidad.otorgada)" false
ensure_push_sub dashboard-insercion-laboral insercion-laboral-eventos "$DASHBOARD_URL/eventos" "$SA_DASHBOARD" "$DASHBOARD_URL/eventos" \
  "$(filter insercion_laboral.postulacion.seleccionada)" false
ensure_push_sub notificaciones-educacion educacion-eventos "$NOTIFICATIONS_URL/eventos/educacion" "$SA_NOTIF_EDU" "$NOTIFICATIONS_URL" \
  "$(filter educacion.actividad.recordatorio educacion.contenido_docente.moderado educacion.contenido_docente.subido educacion.curso.completado_por_participante educacion.curso.creado educacion.curso.eliminado educacion.habilidad.otorgada educacion.inscripcion.creada)" false
ensure_push_sub notificaciones-promocion-humana promocion-humana-eventos "$NOTIFICATIONS_URL/eventos/promocion-humana" "$SA_NOTIF_PH" "$NOTIFICATIONS_URL" \
  "$(filter promocion_humana.ficha.actualizada promocion_humana.ficha.creada promocion_humana.hito.actualizado promocion_humana.hito.registrado)" true
ensure_push_sub notificaciones-usuarios usuarios-eventos "$NOTIFICATIONS_URL/eventos/usuarios" "$SA_NOTIF_USUARIOS" "$NOTIFICATIONS_URL" \
  "$(filter usuarios.email.solicitado usuarios.participante.eliminado usuarios.usuario.creado)" true
ensure_push_sub notificaciones-insercion-laboral insercion-laboral-eventos "$NOTIFICATIONS_URL/eventos/insercion-laboral" "$SA_NOTIF_LABOR" "$NOTIFICATIONS_URL" \
  "$(filter insercion_laboral.oferta.compatible insercion_laboral.oferta.creada insercion_laboral.oferta.eliminada insercion_laboral.oferta.revisada insercion_laboral.postulacion.creada insercion_laboral.postulacion.decidida)" true
ensure_push_sub notificaciones-campanas campanas-eventos "$NOTIFICATIONS_URL/eventos/campanas" "$SA_NOTIF_CAMPANAS" "$NOTIFICATIONS_URL" \
  "$(filter campanas.campana.creada)" false

# ── Permiso para invocar Cloud Run ────────────────────────────────────
# Los servicios son privados (--no-allow-unauthenticated): sin
# roles/run.invoker, Cloud Run rechaza con 403 cada push y cada barrido
# de Cloud Scheduler aunque el token OIDC sea válido (17/11/2026).
SERVICIOS_RUN="$("${G[@]}" run services list --region "$REGION" --format='value(metadata.name,status.url)')"
grant_invoker() { # $1=url del servicio $2..=service accounts
  local url="$1" name sa; shift
  name="$(awk -v u="$url" '$2==u {print $1}' <<<"$SERVICIOS_RUN")"
  if [[ -z "$name" ]]; then
    echo "AVISO: no encontré un servicio de Cloud Run con la URL $url (usar la de 'gcloud run services describe ... --format=value(status.url)')"
    return
  fi
  for sa in "$@"; do
    "${G[@]}" run services add-iam-policy-binding "$name" --region "$REGION"       --member="serviceAccount:$sa" --role=roles/run.invoker >/dev/null
    echo "invoker $sa -> $name"
  done
}

log "Invocadores de Cloud Run"
grant_invoker "$EDUCATION_URL" "$SA_EDU_USUARIOS" "$SA_EDU_CONTENIDO" "$SA_SCHEDULER"
grant_invoker "$HUMAN_PROMOTION_URL" "$SA_PH_EDUCACION" "$SA_PH_USUARIOS" "$SA_SCHEDULER"
grant_invoker "$DASHBOARD_URL" "$SA_DASHBOARD"
grant_invoker "$NOTIFICATIONS_URL" "$SA_NOTIF_EDU" "$SA_NOTIF_PH" "$SA_NOTIF_USUARIOS" "$SA_NOTIF_LABOR" "$SA_NOTIF_CAMPANAS" "$SA_SCHEDULER"

# ── Permisos de publicación (productores) ────────────────────────────
grant_publisher() { # $1=topic $2=service account
  "${G[@]}" pubsub topics add-iam-policy-binding "$1" --member="serviceAccount:$2" --role=roles/pubsub.publisher >/dev/null
  echo "publisher $2 -> $1"
}

log "Productores"
grant_publisher educacion-eventos "$EDUCATION_SA"
grant_publisher contenido-eventos "$EDUCATION_SA" # el receptor del webhook de Contentful vive en education
grant_publisher promocion-humana-eventos "$HUMAN_PROMOTION_SA"
[[ -n "${AUTH_SA:-}" ]] && grant_publisher usuarios-eventos "$AUTH_SA"
[[ -n "${LABOR_SA:-}" ]] && grant_publisher insercion-laboral-eventos "$LABOR_SA"
[[ -n "${CAMPAIGNS_SA:-}" ]] && grant_publisher campanas-eventos "$CAMPAIGNS_SA"

# ── Idempotencia: TTL de eventos_procesados ──────────────────────────
log "TTL de Firestore"
"${G[@]}" firestore fields ttls update expira_en --collection-group=eventos_procesados --enable-ttl --async || \
  echo "(no se pudo configurar el TTL: revisar que la base Firestore sea la (default))"

# ── Barrido del outbox (Cloud Run escala a cero) ─────────────────────
ensure_job() { # $1=nombre $2=url base $3=path (opcional) $4=schedule (opcional)
  local path="${3:-/internal/outbox/flush}"
  local sched="${4:-* * * * *}"
  if "${G[@]}" scheduler jobs describe "$1" --location "$REGION" >/dev/null 2>&1; then
    echo "job $1: ya existe"
  else
    "${G[@]}" scheduler jobs create http "$1" --location "$REGION" \
      --schedule "$sched" --http-method POST --uri "$2$path" \
      --oidc-service-account-email "$SA_SCHEDULER" --oidc-token-audience "$2"
  fi
}

log "Cloud Scheduler"
ensure_job outbox-flush-education "$EDUCATION_URL"
ensure_job outbox-flush-human-promotion "$HUMAN_PROMOTION_URL"
ensure_job delivery-flush-notifications "$NOTIFICATIONS_URL" "/internal/delivery/flush"
ensure_job reminders-education "$EDUCATION_URL" "/internal/reminders/trigger" "0 8 * * *"

cat <<EOF

Listo. Variables a cargar en Cloud Run:

  education:
    PUBSUB_PROJECT_ID=$PROJECT_ID
    PUBSUB_PUSH_AUDIENCE=$EDUCATION_URL
    PUBSUB_PUSH_SERVICE_ACCOUNTS=$SA_EDU_USUARIOS,$SA_EDU_CONTENIDO,$SA_SCHEDULER
    CONTENTFUL_WEBHOOK_SECRET=<desde Secret Manager>

  human-promotion:
    PUBSUB_PROJECT_ID=$PROJECT_ID
    PUBSUB_PUSH_AUDIENCE=$HUMAN_PROMOTION_URL
    PUBSUB_PUSH_SERVICE_ACCOUNTS=$SA_PH_EDUCACION,$SA_PH_USUARIOS,$SA_SCHEDULER

  dashboard:
    PUBSUB_PUSH_AUDIENCE=$DASHBOARD_URL/eventos
    PUBSUB_PUSH_SERVICE_ACCOUNTS=$SA_DASHBOARD

  notifications:
    PUBSUB_PROJECT_ID=$PROJECT_ID
    PUBSUB_PUSH_AUDIENCE=$NOTIFICATIONS_URL
    PUBSUB_PUSH_SERVICE_ACCOUNTS=$SA_NOTIF_EDU,$SA_NOTIF_PH,$SA_NOTIF_USUARIOS,$SA_NOTIF_LABOR,$SA_NOTIF_CAMPANAS,$SA_SCHEDULER

Webhook de Contentful: POST $EDUCATION_URL/webhooks/contentful, con el
header secreto X-Webhook-Secret y los topics Entry.publish, Entry.unpublish
y Entry.delete.
EOF
