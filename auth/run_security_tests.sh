#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Export Go bin to PATH
export PATH="$PATH:$(go env GOPATH)/bin"

echo "======================================"
echo " 1) SAST: Static Application Security Testing"
echo "======================================"
echo "Installing gosec..."
go install github.com/securego/gosec/v2/cmd/gosec@latest
echo "Running gosec..."
# Run gosec. We don't fail immediately if it finds issues, so we can run other tests.
# Temporarily disable set -e for the tests to collect all reports.
set +e
gosec -fmt=json -out=sast_report.json ./...
SAST_EXIT=$?
set -e
if [ $SAST_EXIT -eq 0 ]; then
  echo "✅ SAST passed."
else
  echo "❌ SAST found vulnerabilities. Check sast_report.json"
fi

echo ""
echo "======================================"
echo " 2) SCA: Software Composition Analysis"
echo "======================================"
echo "Installing govulncheck..."
go install golang.org/x/vuln/cmd/govulncheck@latest
echo "Running govulncheck..."
set +e
govulncheck ./... > sca_report.txt
SCA_EXIT=$?
set -e
if [ $SCA_EXIT -eq 0 ]; then
  echo "✅ SCA passed."
else
  echo "❌ SCA found vulnerabilities. Check sca_report.txt"
fi

echo ""
echo "======================================"
echo " 3) DAST: Dynamic Application Security Testing (OWASP ZAP)"
echo "======================================"
echo "Starting Auth API server in background..."
go run cmd/server/main.go &
API_PID=$!

echo "Waiting for API to initialize (10s)..."
sleep 10

echo "Running OWASP ZAP Baseline Scan via Docker..."
set +e
# Run ZAP Docker image.
# We mount the current directory to /zap/wrk to save the report.
# --network="host" is used to allow the container to access localhost on Linux.
docker run --rm --network="host" \
  -v $(pwd):/zap/wrk/:rw \
  -t zaproxy/zap-stable \
  zap-baseline.py -t http://localhost:8080 -r zap_report.html -I
DAST_EXIT=$?
set -e

echo "Shutting down Auth API server..."
kill $API_PID || true

if [ $DAST_EXIT -eq 0 ]; then
  echo "✅ DAST passed."
else
  echo "❌ DAST found vulnerabilities. Check zap_report.html"
fi

echo ""
echo "======================================"
echo " Security Testing Summary"
echo "======================================"
FINAL_EXIT=0
if [ $SAST_EXIT -ne 0 ]; then FINAL_EXIT=$SAST_EXIT; fi
if [ $SCA_EXIT -ne 0 ]; then FINAL_EXIT=$SCA_EXIT; fi
if [ $DAST_EXIT -ne 0 ]; then FINAL_EXIT=$DAST_EXIT; fi

if [ $FINAL_EXIT -eq 0 ]; then
  echo "✅ All security tests passed!"
else
  echo "❌ Some security tests failed."
fi

exit $FINAL_EXIT
