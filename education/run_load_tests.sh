#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "======================================"
echo " Load Testing (k6) — Education Service"
echo "======================================"
echo "Asumiendo que el entorno ya está configurado o proveyendo defaults de testing..."
export FIRESTORE_EMULATOR_HOST="${FIRESTORE_EMULATOR_HOST:-localhost:8090}"
export APP_ENV="${APP_ENV:-testing}"
export CONTENTFUL_ENVIRONMENT="${CONTENTFUL_ENVIRONMENT:-testing}"
export CONTENTFUL_SPACE_ID="${CONTENTFUL_SPACE_ID:-dummy-space-id}"
export CONTENTFUL_CDA_TOKEN="${CONTENTFUL_CDA_TOKEN:-dummy-cda-token}"
export CONTENTFUL_CMA_TOKEN="${CONTENTFUL_CMA_TOKEN:-dummy-cma-token}"
export JWT_SECRET="${JWT_SECRET:-Kj8mN2pQr5vXz9LbWc3tYs7uFhAe1DgHiJmMpNq4Ovs=}"
export PORT="${PORT:-8210}"

echo "Starting Education API server in background..."
go run cmd/server/main.go &
API_PID=$!

echo "Waiting for API to initialize (10s)..."
sleep 10

echo "Running k6 load tests via Docker..."
set +e
docker run --rm -i \
  --add-host host.docker.internal:host-gateway \
  -e BASE_URL="http://host.docker.internal:${PORT}" \
  grafana/k6 run - < load_test.js
K6_EXIT=$?
set -e

echo "Shutting down Education API server..."
kill $API_PID || true

if [ $K6_EXIT -eq 0 ]; then
  echo "✅ Load tests passed! Latency and error thresholds met."
else
  echo "❌ Load tests failed! Check the metrics above."
fi

exit $K6_EXIT
