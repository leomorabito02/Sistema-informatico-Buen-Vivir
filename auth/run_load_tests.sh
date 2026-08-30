#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "======================================"
echo " Load Testing (k6)"
echo "======================================"
echo "Starting Auth API server in background..."
go run cmd/server/main.go &
API_PID=$!

echo "Waiting for API to initialize (10s)..."
sleep 10

echo "Running k6 load tests via Docker..."
set +e
docker run --rm -i --network="host" \
  grafana/k6 run - < load_test.js
K6_EXIT=$?
set -e

echo "Shutting down Auth API server..."
kill $API_PID || true

if [ $K6_EXIT -eq 0 ]; then
  echo "✅ Load tests passed! Latency and error thresholds met."
else
  echo "❌ Load tests failed! Check the metrics above."
fi

exit $K6_EXIT
