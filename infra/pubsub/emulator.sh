#!/usr/bin/env bash
# Crea en el EMULADOR de Pub/Sub los mismos topics y suscripciones push
# (con filtro) que infra/pubsub/setup.sh, apuntando a los servicios que
# corren en local. No necesita gcloud: usa la API REST del emulador.
#
# Uso (con el emulador del docker-compose en localhost:8086):
#   PUBSUB_EMULATOR_HOST=localhost:8086 infra/pubsub/emulator.sh
#
# Los servicios tienen que correr con PUBSUB_PROJECT_ID=$PROJECT,
# PUBSUB_EMULATOR_HOST y PUBSUB_PUSH_AUTH=disabled (el emulador no firma
# los push con OIDC).
set -euo pipefail

HOST="${PUBSUB_EMULATOR_HOST:-localhost:8086}"
PROJECT="${PROJECT:-ceibuenvivir}"
EDUCATION_URL="${EDUCATION_URL:-http://host.docker.internal:8210}"
HUMAN_PROMOTION_URL="${HUMAN_PROMOTION_URL:-http://host.docker.internal:8085}"
DASHBOARD_URL="${DASHBOARD_URL:-http://host.docker.internal:8090}"
NOTIFICATIONS_URL="${NOTIFICATIONS_URL:-http://host.docker.internal:8088}"
API="http://$HOST/v1/projects/$PROJECT"

filter() { local out="" t; for t in "$@"; do out+="${out:+ OR }attributes.event_type = \\\"$t\\\""; done; echo "$out"; }

for t in usuarios-eventos educacion-eventos promocion-humana-eventos contenido-eventos insercion-laboral-eventos campanas-eventos; do
  curl -s -o /dev/null -X PUT "$API/topics/$t" && echo "topic $t"
done

sub() { # $1=nombre $2=topic $3=endpoint $4=filtro
  curl -s -o /dev/null -X PUT "$API/subscriptions/$1" -H 'Content-Type: application/json' -d "{
    \"topic\": \"projects/$PROJECT/topics/$2\",
    \"pushConfig\": {\"pushEndpoint\": \"$3\"},
    \"filter\": \"$4\",
    \"ackDeadlineSeconds\": 30
  }" && echo "suscripción $1 -> $3"
}

sub educacion-usuarios usuarios-eventos "$EDUCATION_URL/eventos/usuarios" "$(filter usuarios.participante.eliminado usuarios.usuario.actualizado)"
sub educacion-contenido contenido-eventos "$EDUCATION_URL/eventos/contenido" "$(filter contenido.despublicado contenido.publicado)"
sub promocion-humana-educacion educacion-eventos "$HUMAN_PROMOTION_URL/eventos/educacion" "$(filter educacion.asistencia_general.registrada educacion.curso.completado_por_participante educacion.inscripcion.creada)"
sub promocion-humana-usuarios usuarios-eventos "$HUMAN_PROMOTION_URL/eventos/usuarios" "$(filter usuarios.participante.eliminado)"
sub dashboard-usuarios usuarios-eventos "$DASHBOARD_URL/eventos" "$(filter usuarios.participante.creado usuarios.participante.eliminado)"
sub dashboard-educacion educacion-eventos "$DASHBOARD_URL/eventos" "$(filter educacion.curso.completado_por_participante educacion.curso.creado educacion.curso.eliminado educacion.habilidad.otorgada)"
sub dashboard-insercion-laboral insercion-laboral-eventos "$DASHBOARD_URL/eventos" "$(filter insercion_laboral.postulacion.seleccionada)"
sub notificaciones-educacion educacion-eventos "$NOTIFICATIONS_URL/eventos/educacion" "$(filter educacion.actividad.recordatorio educacion.contenido_docente.moderado educacion.contenido_docente.subido educacion.curso.completado_por_participante educacion.curso.creado educacion.curso.eliminado educacion.habilidad.otorgada educacion.inscripcion.creada)"
sub notificaciones-promocion-humana promocion-humana-eventos "$NOTIFICATIONS_URL/eventos/promocion-humana" "$(filter promocion_humana.ficha.actualizada promocion_humana.ficha.creada promocion_humana.hito.actualizado promocion_humana.hito.registrado)"
sub notificaciones-usuarios usuarios-eventos "$NOTIFICATIONS_URL/eventos/usuarios" "$(filter usuarios.email.solicitado usuarios.participante.eliminado usuarios.usuario.creado)"
sub notificaciones-insercion-laboral insercion-laboral-eventos "$NOTIFICATIONS_URL/eventos/insercion-laboral" "$(filter insercion_laboral.oferta.creada insercion_laboral.oferta.revisada insercion_laboral.postulacion.creada insercion_laboral.postulacion.decidida)"
sub notificaciones-campanas campanas-eventos "$NOTIFICATIONS_URL/eventos/campanas" "$(filter campanas.campana.creada)"
