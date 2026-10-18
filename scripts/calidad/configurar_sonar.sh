#!/usr/bin/env bash
# Deja un servidor SonarQube listo para el TP de Calidad:
#   - proyecto buenvivir-education,
#   - quality gate "TP-Calidad-N3" con las metas del Nivel 3 (decisión D-18),
#   - asignación del gate al proyecto.
# El período de código nuevo lo fija analizar.sh en cada corrida (D-17).
#
# Se puede correr varias veces: lo que ya existe se deja como está, y las
# condiciones del gate se vuelven a dejar exactamente como dice la tabla.
#
# Uso (desde la raíz del repo, en Git Bash o Linux):
#   scripts/calidad/configurar_sonar.sh
# Variables: SONAR_HOST_URL (default http://localhost:9000) y SONAR_TOKEN
# (o el archivo ~/.sonar-token-tp). El token tiene que ser de un administrador.
set -euo pipefail

SONAR_HOST_URL="${SONAR_HOST_URL:-http://localhost:9000}"
if [ -z "${SONAR_TOKEN:-}" ] && [ -f "$HOME/.sonar-token-tp" ]; then
  SONAR_TOKEN="$(tr -d '\r\n' < "$HOME/.sonar-token-tp")"
fi
: "${SONAR_TOKEN:?Falta SONAR_TOKEN (o el archivo ~/.sonar-token-tp).}"

KEY="buenvivir-education"
GATE="TP-Calidad-N3"
PY="$(command -v python3 || command -v python)"

api() { # api GET|POST ruta [args de curl...]
  local metodo="$1" ruta="$2"; shift 2
  curl -s -u "$SONAR_TOKEN:" -X "$metodo" "$SONAR_HOST_URL$ruta" "$@"
}

# Metas del Nivel 3 (consigna, tabla "Metas de calidad"): métrica, operador, umbral.
# Confiabilidad > 3 = peor que C = hay al menos un bug de severidad alta o bloqueante.
CONDICIONES="
software_quality_reliability_rating GT 3
new_software_quality_reliability_rating GT 3
coverage LT 80
new_coverage LT 80
duplicated_lines_density GT 5
new_duplicated_lines_density GT 5
"

echo "== proyecto $KEY"
if api GET "/api/projects/search" --get -d "projects=$KEY" | grep -q "\"key\":\"$KEY\""; then
  echo "ya existe"
else
  api POST "/api/projects/create" -d "project=$KEY" \
    -d "name=Buen%20Vivir%20-%20Educaci%C3%B3n" -d "visibility=private" >/dev/null
  echo "creado"
fi

echo "== quality gate $GATE"
if ! api GET "/api/qualitygates/show" --get -d "name=$GATE" | grep -q '"conditions"'; then
  api POST "/api/qualitygates/create" -d "name=$GATE" >/dev/null
  echo "creado"
fi

# Al crearse, el gate copia las condiciones de "Sonar way". Se borran todas y
# se cargan las de la tabla, para que el gate sea siempre el mismo.
for id in $(api GET "/api/qualitygates/show" --get -d "name=$GATE" \
    | "$PY" -c "import json,sys;[print(c['id']) for c in json.load(sys.stdin)['conditions']]"); do
  api POST "/api/qualitygates/delete_condition" -d "id=$id" >/dev/null
done
echo "$CONDICIONES" | while read -r metrica op umbral; do
  [ -n "$metrica" ] || continue
  api POST "/api/qualitygates/create_condition" \
    -d "gateName=$GATE" -d "metric=$metrica" -d "op=$op" -d "error=$umbral" >/dev/null
  echo "condición: $metrica $op $umbral"
done

api POST "/api/qualitygates/select" -d "projectKey=$KEY" -d "gateName=$GATE" >/dev/null
echo "== gate asignado a $KEY"
