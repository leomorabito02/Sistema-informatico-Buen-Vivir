#!/usr/bin/env bash
# E2E de Educación: el recorrido completo de un curso (ver e2e.py).
# Uso (desde la raíz del repo, con el entorno de calidad levantado):
#   BASE_URL=http://127.0.0.1:18210 REPORTES=reportes-servicio bash scripts/calidad/pruebas/e2e.sh
set -euo pipefail
PY="$(command -v python3 || command -v python)"
exec "$PY" "$(dirname "$0")/e2e.py"
