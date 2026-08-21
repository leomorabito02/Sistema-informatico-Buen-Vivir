#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "======================================"
echo " 1) SCA: Software Composition Analysis"
echo "======================================"
echo "Running pnpm audit..."
set +e
# pnpm audit requires an active pnpm env, typically we use --audit-level high.
pnpm audit --audit-level high > audit_report.txt
SCA_EXIT=$?
set -e
if [ $SCA_EXIT -eq 0 ]; then
  echo "✅ SCA passed (no high/critical vulnerabilities found in dependencies)."
else
  echo "❌ SCA found high/critical vulnerabilities. Check audit_report.txt"
fi

echo ""
echo "======================================"
echo " 2) SAST: Static Application Security Testing"
echo "======================================"
echo "Running Semgrep..."
set +e
# Run semgrep. We use pnpm dlx to run it without installing it globally.
pnpm dlx semgrep scan --config auto --json -o semgrep_report.json .
SAST_EXIT=$?
set -e
if [ $SAST_EXIT -eq 0 ]; then
  echo "✅ SAST passed."
else
  echo "❌ SAST found vulnerabilities. Check semgrep_report.json"
fi

echo ""
echo "======================================"
echo " 3) DAST: Dynamic Application Security Testing (OWASP ZAP)"
echo "======================================"
echo "Starting Frontend server in background..."
# Start the frontend. Usually pnpm run start serves on 4200.
pnpm run start &
FRONTEND_PID=$!

echo "Waiting for Frontend to initialize (20s)..."
sleep 20

echo "Running OWASP ZAP Baseline Scan via Docker..."
set +e
docker run --rm --network="host" \
  -v $(pwd):/zap/wrk/:rw \
  -t zaproxy/zap-stable \
  zap-baseline.py -t http://localhost:4200 -r zap_report.html -I
DAST_EXIT=$?
set -e

echo "Shutting down Frontend server..."
kill $FRONTEND_PID || true

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
if [ $SCA_EXIT -ne 0 ]; then FINAL_EXIT=$SCA_EXIT; fi
if [ $SAST_EXIT -ne 0 ]; then FINAL_EXIT=$SAST_EXIT; fi
if [ $DAST_EXIT -ne 0 ]; then FINAL_EXIT=$DAST_EXIT; fi

if [ $FINAL_EXIT -eq 0 ]; then
  echo "✅ All frontend security tests passed!"
else
  echo "❌ Some frontend security tests failed."
fi

exit $FINAL_EXIT
