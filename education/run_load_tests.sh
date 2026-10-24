#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "======================================"
echo " Load Testing (k6) — Education Service"
echo "======================================"

# 1. Cargar variables desde .env.testing si existe
if [ -f .env.testing ]; then
  echo "Cargando variables desde .env.testing..."
  set -a
  # shellcheck disable=SC1091
  source .env.testing
  set +a
fi

# Defaults de ejecución para testing
export APP_ENV="${APP_ENV:-testing}"
export CONTENTFUL_ENVIRONMENT="${CONTENTFUL_ENVIRONMENT:-testing}"
export PORT="${PORT:-8210}"
export RATE_LIMIT_DISABLED="true"
export FIRESTORE_EMULATOR_HOST="${FIRESTORE_EMULATOR_HOST:-localhost:8090}"

# Si el puerto configurado ya está ocupado en el host, usar puerto aislado
if timeout 1 bash -c "</dev/tcp/localhost/$PORT" 2>/dev/null; then
  echo "Aviso: El puerto $PORT está en uso. Usando puerto 8219 para aislar la prueba..."
  export PORT="8219"
fi

# 2. Validar variables requeridas para prueba real con Contentful
MISSING_VARS=()
[ -z "$CONTENTFUL_SPACE_ID" ] && MISSING_VARS+=("CONTENTFUL_SPACE_ID")
[ -z "$CONTENTFUL_CDA_TOKEN" ] && MISSING_VARS+=("CONTENTFUL_CDA_TOKEN")
[ -z "$CONTENTFUL_CMA_TOKEN" ] && MISSING_VARS+=("CONTENTFUL_CMA_TOKEN")
[ -z "$JWT_SECRET" ] && MISSING_VARS+=("JWT_SECRET")

if [ ${#MISSING_VARS[@]} -ne 0 ]; then
  echo " Error: Faltan variables obligatorias para el test de carga real contra Contentful:"
  for v in "${MISSING_VARS[@]}"; do
    echo "   - $v"
  done
  echo "Por favor configúralas en .env.testing o en el entorno."
  exit 1
fi

# 3. Control de emulador de Firestore
STARTED_FIRESTORE_CONTAINER=0
EMULATOR_HOST_PORT="${FIRESTORE_EMULATOR_HOST#*://}"
EMULATOR_HOST="${EMULATOR_HOST_PORT%:*}"
EMULATOR_PORT="${EMULATOR_HOST_PORT##*:}"

if ! timeout 2 bash -c "</dev/tcp/$EMULATOR_HOST/$EMULATOR_PORT" 2>/dev/null; then
  echo "Firestore emulator no detectado en $FIRESTORE_EMULATOR_HOST. Iniciando contenedor Docker..."
  docker run --name education_load_test_firestore -d -p 8090:8080 mtlynch/firestore-emulator:latest || true
  STARTED_FIRESTORE_CONTAINER=1
  echo "Esperando 5s a que el emulador esté listo..."
  sleep 5
fi

# 4. Funciones de limpieza y trap de salida
cleanup_contentful() {
  if [ -n "$CONTENTFUL_CMA_TOKEN" ] && [ -n "$CONTENTFUL_SPACE_ID" ]; then
    echo "Verificando y purgando entradas [LOAD-TEST] residuales en Contentful CMA..."
    local cma_url="https://api.contentful.com/spaces/${CONTENTFUL_SPACE_ID}/environments/${CONTENTFUL_ENVIRONMENT}/entries?content_type=course&query=%5BLOAD-TEST%5D"
    local response
    response=$(curl -s -H "Authorization: Bearer ${CONTENTFUL_CMA_TOKEN}" "$cma_url" || true)
    local entry_ids
    if command -v jq >/dev/null 2>&1; then
      entry_ids=$(echo "$response" | jq -r '.items[]?.sys.id // empty' 2>/dev/null || true)
    else
      entry_ids=$(echo "$response" | grep -o '"id": *"[^"]*"' | cut -d'"' -f4 || true)
    fi

    for eid in $entry_ids; do
      if [ -n "$eid" ]; then
        echo "Eliminando entrada residual: $eid"
        curl -s -X DELETE -H "Authorization: Bearer ${CONTENTFUL_CMA_TOKEN}" \
          "https://api.contentful.com/spaces/${CONTENTFUL_SPACE_ID}/environments/${CONTENTFUL_ENVIRONMENT}/entries/${eid}/published" >/dev/null 2>&1 || true
        curl -s -X DELETE -H "Authorization: Bearer ${CONTENTFUL_CMA_TOKEN}" \
          "https://api.contentful.com/spaces/${CONTENTFUL_SPACE_ID}/environments/${CONTENTFUL_ENVIRONMENT}/entries/${eid}" >/dev/null 2>&1 || true
      fi
    done
  fi
}

cleanup() {
  echo ""
  echo "Cerrando servidor Education API..."
  if [ -n "$API_PID" ]; then
    kill "$API_PID" 2>/dev/null || true
  fi
  cleanup_contentful
  if [ "$STARTED_FIRESTORE_CONTAINER" -eq 1 ]; then
    echo "Deteniendo contenedor de Firestore..."
    docker rm -f education_load_test_firestore >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT INT TERM

# 5. Iniciar servidor de Education
echo "Iniciando Education API server en background (PORT=${PORT})..."
go run cmd/server/main.go &
API_PID=$!

echo "Esperando a que la API esté lista en http://localhost:${PORT}/api/salud..."
READY=0
for i in {1..25}; do
  if ! kill -0 "$API_PID" 2>/dev/null; then
    echo " El proceso del servidor terminó inesperadamente."
    exit 1
  fi
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${PORT}/api/salud" || true)
  if [ "$STATUS" = "200" ]; then
    READY=1
    echo " Servidor listo (intento $i)."
    break
  fi
  sleep 1
done

if [ "$READY" -ne 1 ]; then
  echo " Timeout esperando inicialización del servidor."
  exit 1
fi

# 6. Ejecutar k6 vía Docker
echo "Ejecutando pruebas de carga k6 vía Docker..."
set +e
docker run --rm -i \
  --add-host host.docker.internal:host-gateway \
  -e BASE_URL="http://host.docker.internal:${PORT}" \
  -e JWT_SECRET="${JWT_SECRET}" \
  -e CONTENTFUL_CMA_TOKEN="${CONTENTFUL_CMA_TOKEN}" \
  -e CONTENTFUL_SPACE_ID="${CONTENTFUL_SPACE_ID}" \
  -e CONTENTFUL_ENVIRONMENT="${CONTENTFUL_ENVIRONMENT}" \
  grafana/k6 run - < load_test.js
K6_EXIT=$?
set -e

if [ $K6_EXIT -eq 0 ]; then
  echo " ¡Pruebas de carga pasaron exitosamente! Umbrales de latencia y Contentful cumplidos."
else
  echo " Las pruebas de carga fallaron. Revisa las métricas anteriores."
fi

exit $K6_EXIT

