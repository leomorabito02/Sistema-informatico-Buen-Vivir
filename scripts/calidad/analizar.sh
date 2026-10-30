#!/usr/bin/env bash
# Análisis de calidad del microservicio del TP (Educación):
#   1. corre las pruebas existentes y genera el informe de cobertura,
#   2. lo analiza con SonarQube y espera el resultado del quality gate,
#   3. agrega una fila al historial de métricas (docs/calidad/reportes/).
#
# Uso (desde la raíz del repo, en Git Bash o Linux):
#   scripts/calidad/analizar.sh education
#
# Variables:
#   SONAR_HOST_URL   (default http://localhost:9000)
#   SONAR_TOKEN      token de análisis; si falta, se lee de ~/.sonar-token-tp
#   SONAR_SCANNER    ejecutable del scanner (default: sonar-scanner del PATH)
#   FIRESTORE_EMULATOR_HOST  sin él, los tests de Firestore de education se saltean
#
# No modifica código. Las pruebas de education reescriben el timestamp de
# reports/database y reports/pact (comportamiento previo de esos tests).
set -uo pipefail

SERVICIO="${1:-}"
RAIZ="$(cd "$(dirname "$0")/../.." && pwd)"
SONAR_HOST_URL="${SONAR_HOST_URL:-http://localhost:9000}"
SONAR_SCANNER="${SONAR_SCANNER:-sonar-scanner}"

case "$SERVICIO" in
  education) DIR="$RAIZ/education"; KEY="buenvivir-education" ;;
  *) echo "uso: $0 education" >&2; exit 2 ;;
esac

if [ -z "${SONAR_TOKEN:-}" ] && [ -f "$HOME/.sonar-token-tp" ]; then
  SONAR_TOKEN="$(tr -d '\r\n' < "$HOME/.sonar-token-tp")"
fi
if [ -z "${SONAR_TOKEN:-}" ]; then
  echo "Falta SONAR_TOKEN (o el archivo ~/.sonar-token-tp)." >&2
  exit 2
fi
export SONAR_TOKEN

COMMIT="$(git -C "$RAIZ" rev-parse --short HEAD)"
# Cambios sin commitear: la versión lo dice para no confundirla con el commit limpio.
if [ -n "$(git -C "$RAIZ" status --porcelain -- "$DIR")" ]; then
  COMMIT="$COMMIT-modificado"
fi

echo "== [$SERVICIO] pruebas y cobertura (commit $COMMIT)"
cd "$DIR" || exit 2
TESTS_OK=si
go test -count=1 -json -coverprofile=coverage.out ./... > test-report.json
[ $? -eq 0 ] || TESTS_OK=no
go tool cover -func=coverage.out | tail -1
echo "Pruebas en verde: $TESTS_OK"
if [ "$TESTS_OK" != "si" ]; then
  echo "::error::Hay pruebas fallidas en $SERVICIO. Deteniendo análisis." >&2
  exit 1
fi
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
echo "Cobertura total: ${COVERAGE}%"
if ! python3 -c "import sys; sys.exit(0 if float('$COVERAGE') >= 80.0 else 1)"; then
  echo "::error::La cobertura (${COVERAGE}%) no supera el umbral requerido del 80%." >&2
  exit 1
fi

PY="$(command -v python3 || command -v python)"

# Código nuevo = desde el 1/8/2026 (decisión D-17). La edición gratuita solo
# acepta "últimos N días", así que N se recalcula en cada análisis para que
# el inicio quede fijo.
INICIO_CODIGO_NUEVO="${INICIO_CODIGO_NUEVO:-2026-08-01}"
DIAS="$("$PY" -c "from datetime import date;print((date.today()-date.fromisoformat('$INICIO_CODIGO_NUEVO')).days)")"
curl -sf -u "$SONAR_TOKEN:" -X POST "$SONAR_HOST_URL/api/new_code_periods/set" \
  -d "project=$KEY&type=NUMBER_OF_DAYS&value=$DIAS" >/dev/null \
  || echo "Aviso: no se pudo fijar el código nuevo en $DIAS días." >&2
echo "Código nuevo: desde $INICIO_CODIGO_NUEVO ($DIAS días)"

echo "== [$SERVICIO] análisis SonarQube"
"$SONAR_SCANNER" \
  -Dsonar.host.url="$SONAR_HOST_URL" \
  -Dsonar.projectVersion="$COMMIT" \
  -Dsonar.qualitygate.wait=true \
  -Dsonar.qualitygate.timeout=300
GATE_EXIT=$?

echo "== [$SERVICIO] historial de métricas"
"$PY" "$RAIZ/scripts/calidad/exportar_metricas.py" \
  --host "$SONAR_HOST_URL" --proyecto "$KEY" --servicio "$SERVICIO" \
  --commit "$COMMIT" --tests-ok "$TESTS_OK" \
  --salida "$RAIZ/docs/calidad/reportes"

# El quality gate decide el resultado: rojo = el script falla.
exit $GATE_EXIT
