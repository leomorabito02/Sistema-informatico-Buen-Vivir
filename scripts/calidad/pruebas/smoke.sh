#!/usr/bin/env bash
# Smoke test de Educación: ¿el servicio arrancó y responde lo básico?
#
# Comprueba, en pocos segundos:
#   1. /api/salud responde 200, "UP", con APP_ENV y Contentful "testing"
#      (el mismo control que hace cd-smoke después de desplegar).
#   2. Una ruta protegida rechaza un pedido sin token (401).
#   3. Con un token de admin se lee el catálogo: Contentful responde.
#   4. Con ese token se leen las actividades de un curso: Firestore responde.
#   5. Un participante ve sus cursos.
# Termina con error en el primer control que falla.
#
# Uso (desde la raíz del repo, con el entorno de calidad levantado):
#   BASE_URL=http://127.0.0.1:18210 REPORTES=reportes-servicio bash scripts/calidad/pruebas/smoke.sh

set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:18210}"
REPORTES="${REPORTES:-reportes-servicio}"
ENTORNO="$(dirname "$0")/../entorno"
PY="$(command -v python3 || command -v python)"
mkdir -p "$REPORTES"
REPORTE="$REPORTES/smoke.txt"
: > "$REPORTE"

ok() { echo "OK    $1" | tee -a "$REPORTE"; }
falla() { echo "FALLA $1" | tee -a "$REPORTE"; exit 1; }

# pedir <token o vacío> <ruta>: deja el cuerpo en $CUERPO y el código en $CODIGO.
pedir() {
  local auth=()
  [ -n "$1" ] && auth=(-H "Authorization: Bearer $1")
  CUERPO=$(curl -s --max-time 10 -w '\n%{http_code}' "${auth[@]}" "$BASE_URL$2") || CUERPO=$'\n000'
  CODIGO="${CUERPO##*$'\n'}"
  CUERPO="${CUERPO%$'\n'*}"
}

# campo <expresión de Python sobre d>: lee el JSON de $CUERPO.
campo() {
  printf '%s' "$CUERPO" | "$PY" -c "import json,sys
d=json.load(sys.stdin)
d=d.get('data',d) if isinstance(d,dict) else d
print($1)"
}

ADMIN=$("$PY" "$ENTORNO/token_prueba.py" role_admin --usuario admin-prueba)
PARTICIPANTE=$("$PY" "$ENTORNO/token_prueba.py" role_participant --usuario usuario-prueba-01)

pedir "" /api/salud
[ "$CODIGO" = 200 ] || falla "/api/salud respondió $CODIGO"
[ "$(campo "d['status']")" = UP ] || falla "/api/salud no dice UP: $CUERPO"
[ "$(campo "d['appEnv']")" = testing ] || falla "APP_ENV no es testing: $CUERPO"
[ "$(campo "d['contentfulEnvironment']")" = testing ] || falla "Contentful no es testing: $CUERPO"
ok "/api/salud: UP, APP_ENV y Contentful en testing"

pedir "" /api/cursos
[ "$CODIGO" = 401 ] || falla "GET /api/cursos sin token respondió $CODIGO (se esperaba 401)"
ok "GET /api/cursos sin token: 401"

pedir "$ADMIN" /api/cursos
[ "$CODIGO" = 200 ] || falla "GET /api/cursos como admin respondió $CODIGO"
CURSOS=$(campo "len(d)")
[ "$CURSOS" -gt 0 ] || falla "el catálogo de Contentful testing está vacío"
CURSO=$(campo "d[0]['id']")
ok "GET /api/cursos como admin: $CURSOS cursos (Contentful responde)"

pedir "$ADMIN" "/api/cursos/$CURSO/actividades"
[ "$CODIGO" = 200 ] || falla "GET /api/cursos/$CURSO/actividades respondió $CODIGO"
ok "GET /api/cursos/{id}/actividades: $(campo "len(d or [])") actividades (Firestore responde)"

pedir "$PARTICIPANTE" /api/participante/cursos
[ "$CODIGO" = 200 ] || falla "GET /api/participante/cursos respondió $CODIGO"
ok "GET /api/participante/cursos como participante: 200"

echo "Smoke aprobado ($(date -u +%FT%TZ))" | tee -a "$REPORTE"
