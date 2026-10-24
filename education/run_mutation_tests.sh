#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "======================================"
echo " Mutation Testing — Education Service "
echo "======================================"

# 1. Configurar PATH para binarios de Go
export PATH="$PATH:$(go env GOPATH)/bin"

# 2. Instalar Gremlins si no está disponible (versión fija idéntica a CI)
if ! command -v gremlins >/dev/null 2>&1; then
  echo "Instalando Gremlins (v0.5.0)..."
  go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.5.0
else
  echo "Gremlins ya se encuentra instalado."
fi

# 3. Definir target de mutación (default ./internal, o argumento modular)
TARGET="${1:-./internal}"
TARGET="${TARGET%/...}"
TARGET="${TARGET%/}"
echo "Objetivo de mutación: $TARGET"

# 4. Asegurar disponibilidad de Firestore Emulator para no saltear tests de repositorios
export FIRESTORE_EMULATOR_HOST="${FIRESTORE_EMULATOR_HOST:-localhost:8090}"
STARTED_FIRESTORE_CONTAINER=0
EMULATOR_HOST_PORT="${FIRESTORE_EMULATOR_HOST#*://}"
EMULATOR_HOST="${EMULATOR_HOST_PORT%:*}"
EMULATOR_PORT="${EMULATOR_HOST_PORT##*:}"

if ! timeout 2 bash -c "</dev/tcp/$EMULATOR_HOST/$EMULATOR_PORT" 2>/dev/null; then
  echo "Firestore emulator no detectado en $FIRESTORE_EMULATOR_HOST. Iniciando contenedor Docker..."
  docker rm -f education_mutation_firestore >/dev/null 2>&1 || true
  docker run --name education_mutation_firestore -d -p 8090:8080 mtlynch/firestore-emulator:latest
  STARTED_FIRESTORE_CONTAINER=1
  echo "Esperando 5s a que el emulador esté listo..."
  sleep 5
fi

cleanup() {
  if [ "$STARTED_FIRESTORE_CONTAINER" -eq 1 ]; then
    echo ""
    echo "Deteniendo contenedor de Firestore..."
    docker rm -f education_mutation_firestore >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT INT TERM

# 5. Ejecutar Gremlins
echo "Iniciando análisis de mutación con Gremlins..."
set -o pipefail
gremlins unleash "$TARGET" | tee mutation_report.txt
EXIT_CODE=$?

echo ""
if [ $EXIT_CODE -eq 0 ]; then
  echo " ¡Pruebas de mutación pasaron exitosamente!"
else
  echo " Se detectaron mutantes supervivientes o errores. Revisa mutation_report.txt para detalles."
fi

exit $EXIT_CODE

