# tests unitarios backend auth
go test -coverprofile=../reports/coverage/auth.out ./internal/services/...
go tool cover -func=../reports/coverage/auth.out

# test de mutacion backend auth

./run_mutation_tests.sh

# test de seguridad backend auth
./run_security_tests.sh

# test unitarios frontend

npx ng test --no-watch --code-coverage

# test de mutación frontend

pnpm run test:mutate

# test de seguridad frontend
./run_security_tests.sh

SCA: Execute 
pnpm audit --audit-level high 
to catch critical vulnerability dependencies in your package.json.
SAST: Utilize Semgrep via 
pnpm dlx semgrep scan --config auto 

to statically analyze your TypeScript code for frontend-specific vulnerabilities (XSS, bad DOM manipulation, etc.).
DAST: Execute 
pnpm run start 
in the background, run the ZAP baseline scan docker container against http://localhost:4200, and shut down the server afterwards.


# test de carga

docker run --rm -i --network="host" grafana/k6 run - < load_test.js