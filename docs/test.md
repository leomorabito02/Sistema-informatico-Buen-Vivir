# tests unitarios backend auth
go test -coverprofile=../reports/coverage/auth.out ./internal/services/...
go tool cover -func=../reports/coverage/auth.out

# test unitarios frontend

cd frontend
npx ng test --no-watch --code-coverage
