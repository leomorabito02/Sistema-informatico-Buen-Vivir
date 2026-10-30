#!/usr/bin/env bash
# Arma scripts/calidad/entorno/.env.calidad para el entorno de calidad:
#   - copia de education/.env solo las variables de Contentful, y solo si
#     ese archivo apunta al environment "testing" (D-08);
#   - genera un JWT_SECRET nuevo, propio de este entorno: los tokens de
#     prueba (token_prueba.py) no sirven en ningún otro lado.
# No muestra ningún valor por pantalla. El archivo generado no se versiona.
#
# Uso (desde la raíz del repo): bash scripts/calidad/entorno/preparar_env.sh
set -euo pipefail

RAIZ="$(cd "$(dirname "$0")/../../.." && pwd)"
ORIGEN="$RAIZ/education/.env"
DESTINO="$RAIZ/scripts/calidad/entorno/.env.calidad"

[ -f "$ORIGEN" ] || { echo "No existe education/.env" >&2; exit 1; }
valor() { grep -E "^$1=" "$ORIGEN" | tail -1 | cut -d= -f2- | tr -d '\r'; }

if [ "$(valor CONTENTFUL_ENVIRONMENT)" != "testing" ]; then
  echo "education/.env no apunta a Contentful 'testing': no se copia nada." >&2
  exit 1
fi
for v in CONTENTFUL_SPACE_ID CONTENTFUL_CDA_TOKEN CONTENTFUL_CMA_TOKEN; do
  [ -n "$(valor "$v")" ] || { echo "Falta $v en education/.env" >&2; exit 1; }
done

umask 077
{
  echo "# Generado por preparar_env.sh el $(date +%F). No versionar."
  for v in CONTENTFUL_SPACE_ID CONTENTFUL_CDA_TOKEN CONTENTFUL_CMA_TOKEN CONTENTFUL_LOCALE; do
    echo "$v=$(valor "$v")"
  done
  echo "JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || python -c 'import secrets;print(secrets.token_hex(32))')"
} > "$DESTINO"
echo "Listo: scripts/calidad/entorno/.env.calidad (Contentful testing + JWT_SECRET nuevo)"
