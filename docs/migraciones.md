
# Poblar la base de datos con datos iniciales
set -a && source .env && set +a && go run cmd/migrate/main.go

# Actualizar la base de datos con las modificaciones de los model
set -a && source .env && set +a && go run cmd/seed/main.go