#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "Installing Gremlins (Mutation Testing Tool)..."
go install github.com/go-gremlins/gremlins/cmd/gremlins@latest

export PATH="$PATH:$(go env GOPATH)/bin"

echo "Running Gremlins Mutation Tests..."
# tee will output to stdout and save to mutation_report.txt
# We capture the pipe exit status of gremlins, not tee, by using PIPESTATUS
gremlins unleash | tee mutation_report.txt
EXIT_CODE=${PIPESTATUS[0]}

echo ""
if [ $EXIT_CODE -eq 0 ]; then
  echo "✅ Mutation tests passed successfully!"
else
  echo "❌ Mutation tests failed or some mutations survived. Check the report for details."
fi

exit $EXIT_CODE
