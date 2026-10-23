#!/usr/bin/env bash
# Registra en SonarQube la "foto" de Educación al inicio del código nuevo
# (1/8/2026, decisión D-17), con su fecha real.
#
# Por qué hace falta: con "últimos N días", SonarQube compara contra el
# primer análisis anterior a esa fecha. Si no hay ninguno, toma el primero
# que encuentre (el de hoy) y dice que no hay código nuevo. Este script
# analiza el último commit anterior al 1/8 y lo fecha el día anterior, así
# el análisis normal tiene contra qué comparar.
#
# Solo funciona en un proyecto sin análisis previos: SonarQube no acepta
# análisis con fecha anterior al último que tiene. Se corre una vez, después
# de configurar_sonar.sh y antes del primer analizar.sh.
#
# Uso (desde la raíz del repo): scripts/calidad/linea_base.sh
set -euo pipefail

RAIZ="$(cd "$(dirname "$0")/../.." && pwd)"
SONAR_HOST_URL="${SONAR_HOST_URL:-http://localhost:9000}"
SONAR_SCANNER="${SONAR_SCANNER:-sonar-scanner}"
INICIO_CODIGO_NUEVO="${INICIO_CODIGO_NUEVO:-2026-08-01}"
if [ -z "${SONAR_TOKEN:-}" ] && [ -f "$HOME/.sonar-token-tp" ]; then
  SONAR_TOKEN="$(tr -d '\r\n' < "$HOME/.sonar-token-tp")"
fi
: "${SONAR_TOKEN:?Falta SONAR_TOKEN (o el archivo ~/.sonar-token-tp).}"
export SONAR_TOKEN

PY="$(command -v python3 || command -v python)"
FECHA="$("$PY" -c "from datetime import date,timedelta;print(date.fromisoformat('$INICIO_CODIGO_NUEVO')-timedelta(days=1))")"
COMMIT="$(git -C "$RAIZ" log -1 --before="${INICIO_CODIGO_NUEVO}T00:00:00" --format=%h -- education)"
: "${COMMIT:?No hay commits de education antes de $INICIO_CODIGO_NUEVO}"
echo "== línea base: education en $COMMIT, fechada $FECHA"

TMP="$(mktemp -d)"
trap 'git -C "$RAIZ" worktree remove --force "$TMP/repo" >/dev/null 2>&1 || true; rm -rf "$TMP"' EXIT
git -C "$RAIZ" worktree add --detach "$TMP/repo" "$COMMIT" >/dev/null

# Misma configuración que el análisis normal. Sin pruebas ni cobertura: de
# esta foto solo importa la fecha y el código. Sin SCM porque la carpeta es
# una copia temporal; las fechas de cada línea las aporta el análisis normal.
"$SONAR_SCANNER" \
  -Dsonar.host.url="$SONAR_HOST_URL" \
  -Dproject.settings="$RAIZ/education/sonar-project.properties" \
  -Dsonar.projectBaseDir="$TMP/repo/education" \
  -Dsonar.projectDate="$FECHA" \
  -Dsonar.projectVersion="linea-base-$COMMIT" \
  -Dsonar.scm.disabled=true
