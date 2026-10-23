# Documento Integral de Testing — education-service

Este documento consolida la estrategia, arquitectura, fundamentos teóricos, casos de prueba dinámicos, scripts auxiliares y aseguramiento de la calidad (QA/QC) implementados en el microservicio `education-service` (`education/`). Se basa estrictamente en el código de producción, las suites de tests automatizados y los scripts existentes en el repositorio.

---

## 1. Scripts de Testing Implementados

En la raíz del módulo `education/` se encuentran scripts ejecutables que complementan las pruebas unitarias y automatizan verificaciones de carga, mutación y seguridad:

| Script | Herramienta | Tipo de Prueba | Descripción del Flujo |
|---|---|---|---|
| `load_test.js` | **k6** | Rendimiento y Carga | Escenario k6 configurado con stages: ramp-up a 20 VUs (10s), soak a 20 VUs (30s), ramp-down a 0 (10s). Umbral de latencia: `p(95) < 500ms`. Evalúa `/api/salud` (200 o 429), `/api/cursos` sin token (401 o 429) y `/api/cursos` con token inválido (401 o 429), contemplando la activación defensiva del rate limiting por IP (100 req/min). |
| `run_load_tests.sh` | **k6** + Docker + Bash | Carga local | Levanta `cmd/server/main.go` en segundo plano con variables de entorno por defecto, espera 10s para inicialización y ejecuta la imagen `grafana/k6 run - < load_test.js` vía Docker mapeando `host.docker.internal`. Al finalizar, termina el proceso de la API y valida el exit code. |
| `run_mutation_tests.sh` | **gremlins** | Testing de Mutación | Instala `github.com/go-gremlins/gremlins/cmd/gremlins@latest`, ejecuta `gremlins unleash` sobre el código fuente de `internal/...`, exporta el reporte a `mutation_report.txt` y valida mediante `PIPESTATUS` la supervivencia de mutantes. |
| `run_security_tests.sh` | **gosec** + **govulncheck** + **OWASP ZAP** | Seguridad (SAST, SCA, DAST) | Orquesta 3 etapas: <br>1. **SAST:** `gosec -fmt=json -out=sast_report.json ./...`<br>2. **SCA:** `govulncheck ./... > sca_report.txt`<br>3. **DAST:** Levanta contenedor del emulador de Firestore (`mtlynch/firestore-emulator:latest` en puerto 8090), arranca la API en background con variables de prueba, ejecuta el escaneo baseline de `zaproxy/zap-stable` (`zap-baseline.py -t http://host.docker.internal:8210 -r zap_report.html -I`) y limpia contenedores/procesos al concluir. |

### Comandos de Ejecución de Suites Go

```bash
# 1. Pruebas unitarias en memoria (servicios, controladores, middlewares, repos Contentful simulados)
cd education
go test ./...

# 2. Pruebas completas incluyendo repositorios Firestore (requiere emulador)
export FIRESTORE_EMULATOR_HOST=localhost:8090
go test -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out

# 3. Tests de contrato Pact y de integridad de datos
go test ./tests/contract/...
go test ./tests/database/...
```

---

## 2. Fundamentos del Testing

### 2.1 Pruebas Unitarias en Patrón Arrange–Act–Assert (AAA)

Las pruebas unitarias del microservicio están estructuradas de forma canónica bajo el patrón **Arrange–Act–Assert**, garantizando separación explícita de configuración, ejecución e inspección de resultados.

#### Ejemplo 1: Prueba de Servicio de Cursos (`internal/services/course_service_test.go`)
```go
func TestCourseService_UpdateCourse_Success(t *testing.T) {
	// Arrange
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)
	input := validCourseInput()
	expectedCourse := &models.Course{ID: "c1", Title: input.Title}
	repo.On("Update", mock.Anything, "c1", input).Return(expectedCourse, nil)

	// Act
	course, err := svc.UpdateCourse(context.Background(), "c1", input)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "c1", course.ID)
	repo.AssertExpectations(t)
}
```

#### Ejemplo 2: Prueba de Cliente HTTP (`internal/repositories/auth_service_client_test.go`)
```go
func TestAuthServiceClient_GetParticipant_InvalidJSON(t *testing.T) {
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{invalid-json`))
	}))
	defer srv.Close()

	c := NewAuthServiceClient(srv.URL)

	// Act
	_, err := c.GetParticipant(t.Context(), "p1", "tok")

	// Assert
	assert.Error(t, err)
}
```

#### Estrategia de Aislamiento de Dependencias Externas en Arrange

Para evitar dependencias de red o infraestructura en vivo, el Arrange aísla cada capa:
1. **Lógica de negocio (`internal/services`)**: Mocks completos de repositorios generados con `testify/mock` (`MockCourseRepository`, `MockEnrollmentRepository`, `MockActivityRepository`, etc.).
2. **Controladores HTTP (`internal/controllers`)**: Mocks de servicios inyectados en handlers ejecutados contra requests de `net/http/httptest`.
3. **Middlewares (`internal/middlewares`)**: `httptest.ResponseRecorder` y mocks de `JWTService`.
4. **Repositorios Contentful (`internal/repositories`)**: Servidor local `httptest.Server` que sustituye las URLs reales (`cdaBaseURL`, `cmaBaseURL`, `uploadBaseURL`), simulando respuestas JSON y cabeceras de la API de Contentful sin tokens reales.
5. **Repositorios Firestore (`internal/repositories`)**: Conexión al emulador oficial (`cloud-firestore-emulator` o contenedor `mtlynch/firestore-emulator`). Si la variable `FIRESTORE_EMULATOR_HOST` no está configurada, los tests realizan `t.Skip` ordenadamente.

---

### 2.2 Criterio de Aceptación Fijado Antes de Ejecutar

Los criterios de aceptación se definieron de forma previa a la ejecución técnica, alineados con las Historias de Usuario (HU) y los Quality Gates de Integración Continua:

#### Criterios Funcionales (Definición de Hecho / Aceptación)
- **HU-18 / RF-18 (Inscripción a Cursos):**
  - Solo se inscribe si el participante tiene `approval_status == "approved"` (o `"accepted"`) en `auth-service`.
  - El curso debe existir y estar publicado en Contentful.
  - La cantidad de inscriptos activos debe ser menor a `course.Capacity`. Si está lleno, debe rechazar con HTTP 400 y mensaje específico `"El curso no tiene cupo disponible"`.
  - No permitir duplicidad de inscripciones activas para el mismo participante y curso.
- **HU-19 / RF-19 (Asistencia por Curso):**
  - Permitir registro únicamente a docentes titulares (`role_teacher` con validación de propiedad) o administradores/voluntarios (`role_admin`, `role_volunteer`).
  - El participante debe existir, estar aprobado y contar con inscripción activa en el curso de la actividad.
  - El guardado en Firestore debe ser idempotente: el ID del documento es determinístico `{actividadId}_{participanteId}`.
- **HU-19bis / RF-64 (Asistencia General):**
  - La fecha debe cumplir formato estricto `YYYY-MM-DD`.
  - Se admiten observaciones de texto libre y estado booleano (`presente`).
  - La clave del documento en Firestore es determinística `{fecha}_{participanteId}`, evitando duplicados por reintentos o concurrencia.
- **HU-20 / RF-20 (Progreso Formativo):**
  - Cálculo de porcentaje de asistencia: `(actividades_asistidas / total_actividades_curso) * 100`.
  - En cursos con 0 actividades creadas, el cálculo no debe generar división por cero y debe retornar `0.0%`.

#### Criterios de Aceptación Técnicos en CI (`.github/workflows/pipeline.yml`)
- Cobertura global de código obligatoria: **≥ 70%** (enforce en step de pipeline).
- Linter estático `golangci-lint` y `go vet` con 0 advertencias.
- Reporte Pact de contrato y reporte de integridad de Firestore en estado `PASSED`/`VERIFIED`.
- Escaneo de vulnerabilidades Trivy y gosec sin alertas de severidad `HIGH` o `CRITICAL`.

---

### 2.3 Defecto Real con Contraste Rojo/Verde

Durante la implementación y rigidización de las pruebas unitarias, se identificó y corrigió un defecto crítico de desreferenciación de puntero nulo en el cliente HTTP de autenticación (`internal/repositories/auth_service_client.go`):

#### 1. Caso Rojo (Falla / Defecto Detectado)
- **Causa Raíz:** Los métodos `GetParticipant` y `doGet` asumían que la instancia de `AuthServiceClient` era no nula y evaluaban directamente `if c.baseURL == ""`. Cuando el cliente se invocaba como puntero nulo (`var nilClient *AuthServiceClient`), la llamada provocaba un `panic: runtime error: invalid memory address or nil pointer dereference`.
- **Ejecución del Test Rojo:**
  ```go
  var nilClient *AuthServiceClient
  _, err := nilClient.GetParticipant(t.Context(), "p1", "tok")
  // Resultado: PANIC fatal en tiempo de ejecución, interrumpiendo el proceso.
  ```

#### 2. Corrección de Código (Refactor Mínimo)
```diff
--- a/education/internal/repositories/auth_service_client.go
+++ b/education/internal/repositories/auth_service_client.go
@@ -77,7 +77,7 @@ func (c *AuthServiceClient) GetParticipant(ctx context.Context, participantID, bearerToken string) (*models.AuthParticipant, error) {
-	if c.baseURL == "" {
+	if c == nil || c.baseURL == "" {
 		return nil, ErrAuthServiceNotConfigured
 	}

@@ -139,7 +139,7 @@ func (c *AuthServiceClient) doGet(ctx context.Context, path, bearerToken string, out any) error {
-	if c.baseURL == "" {
+	if c == nil || c.baseURL == "" {
 		return ErrAuthServiceNotConfigured
 	}
```

#### 3. Caso Verde (Test de Regresión Exitoso)
El test `TestAuthServiceClient_NotConfigured` en `internal/repositories/auth_service_client_test.go` valida que tanto instancias vacías como punteros nulos devuelven `ErrAuthServiceNotConfigured` sin generar excepciones no controladas:
```go
var nilClient *AuthServiceClient
_, err = nilClient.GetParticipant(t.Context(), "p1", "tok")
assert.ErrorIs(t, err, ErrAuthServiceNotConfigured) // PASSED
```

---

### 2.4 Los Siete Principios del Testing Aplicados al Proyecto

1. **Las pruebas demuestran la presencia de defectos, no su ausencia:** La suite de tests automatizados detectó defectos de frontera (como nil-pointer en `AuthServiceClient` o desbordamiento de cupo), pero el proceso se refuerza con pruebas de mutación (`gremlins`) y análisis dinámico de seguridad (`OWASP ZAP`) para revelar vulnerabilidades no contempladas.
2. **Las pruebas exhaustivas son imposibles:** En lugar de probar infinitas combinaciones de entradas en inscripciones y asistencias, se aplican particiones de equivalencia y análisis de valores límite en campos críticos (`capacity`, `approval_status`, `mode`, `fecha`).
3. **Pruebas tempranas (Early Testing):** Se desarrollaron tests unitarios sobre `internal/services` con repositorios mockeados antes de desplegar instancias de Contentful o Postgres de Auth, validando las reglas de negocio desde la fase inicial.
4. **Agrupación de defectos (Defect Clustering):** Los módulos con mayor concentración de reglas de validación (`internal/services` con 74 tests y `internal/repositories` con más de 40 tests) albergan la mayor densidad de pruebas unitarias, reconociendo que allí radica la mayor probabilidad de fallos.
5. **La paradoja del pesticida:** Si los tests se mantienen estáticos sin evolucionar, pierden efectividad. Por ello se incorporó `run_mutation_tests.sh` con Gremlins, introduciendo mutantes en el AST de Go para validar si las aserciones realmente fallan cuando el código cambia.
6. **Las pruebas dependen del contexto:** La estrategia distingue cada tecnología: Contentful (SaaS externo) se simula mediante `httptest.Server` para evitar consumo de cuota y dependencia de internet; Firestore (NoSQL documental) se prueba contra el emulador oficial; y los endpoints HTTP se prueban con `httptest.ResponseRecorder`.
7. **La falacia de ausencia de errores:** Que todos los tests unitarios pasen no garantiza un sistema útil si el flujo entre microservicios no está integrado. Por esta razón se establecieron pruebas de contrato (`education_contract_test.go`), validaciones de integridad (`firestore_integrity_test.go`) y guías de pruebas manuales de integración (`GUIA_PRUEBAS_INSCRIPCION_ASISTENCIA.md`).

---

## 3. Pruebas Dinámicas

### 3.1 Casos de Prueba Documentados

La suite implementada cubre los casos de prueba funcionales e integrados del microservicio:

| ID Caso | Requerimiento / HU | Objetivo del Caso de Prueba | Entrada / Precondición | Resultado Esperado | Estado |
|---|---|---|---|---|---|
| **CP-01** | HU-18 / RF-18 | Inscripción correcta de participante | Participante aprobado en Auth, curso publicado con cupo libre | HTTP 201, documento creado en `inscripciones` con `estado: "activa"` | Verificado |
| **CP-02** | HU-18 / RF-18 | Prevención de inscripción duplicada | Mismo participante ya inscripto previamente en estado activo | HTTP 400 `"El participante ya se encuentra inscripto en el curso"` | Verificado |
| **CP-03** | HU-18 / RF-18 | Rechazo por curso sin cupo | Curso con `capacity` alcanzado por inscripciones activas | HTTP 400 `"El curso no tiene cupo disponible"` | Verificado |
| **CP-04** | HU-18 / RF-18 | Rechazo por curso despublicado (inactivo) | Curso existente en CMA de Contentful pero no publicado en CDA | HTTP 400 `"El curso no está disponible (inactivo o no publicado)"` | Verificado |
| **CP-05** | HU-18 / RF-18 | Rechazo por curso inexistente | ID de curso no existente en Contentful | HTTP 400 `"El curso no existe"` | Verificado |
| **CP-06** | HU-18 / RF-18 | Rechazo por participante inexistente | `participanteId` no registrado en `auth-service` | HTTP 400 `"El participante no existe"` | Verificado |
| **CP-07** | HU-18 / RF-18 | Rechazo por participante no aprobado | Participante con estado `"pending"` o `"rejected"` | HTTP 400 `"El participante no está activo"` | Verificado |
| **CP-08** | HU-18 / RF-18 | Control de acceso por RBAC | Token con `role_participant` en endpoint administrativo | HTTP 403 `Forbidden: insufficient role permissions` | Verificado |
| **CP-09** | HU-18 / RF-18 | Persistencia y consulta en Firestore | Inscripción persistida consultada vía GET | HTTP 200 con el listado de inscriptos | Verificado |
| **CP-10** | HU-19 / RF-19 | Registro de asistencia por actividad | Docente titular o voluntario/admin; participante inscripto y activo | HTTP 200 `{"status":"asistencia registrada"}` | Verificado |
| **CP-11** | HU-19 / RF-19 | Consulta de historial de asistencia | Actividad con asistencias registradas | HTTP 200 con registros de asistencia | Verificado |
| **CP-12** | HU-19 / RF-19 | Rechazo por participante no inscripto | Participante aprobado pero sin inscripción en el curso | HTTP 400 `"El participante <id> no está inscripto en el curso"` | Verificado |
| **CP-13** | HU-19 / RF-19 | Rechazo por participante inactivo en asistencia | Participante en estado pendiente en `auth-service` | HTTP 400 `"El participante <id> no está activo"` | Verificado |
| **CP-14** | HU-19 / RF-19 | Consulta de actividad inexistente | ID de actividad no registrada en Firestore | HTTP 404 Not Found (`repositories.ErrNotFound`) | Verificado |
| **CP-15** | HU-19 / RF-19 | Idempotencia en asistencia (evitar duplicados) | Reenvío de asistencia para el mismo participante y actividad | HTTP 200, documento único actualizado (`{actividadId}_{participanteId}`) | Verificado |
| **CP-16** | HU-19 / RF-19 | Control de acceso y titularidad de docente | Docente ajeno al curso intentando registrar asistencia | HTTP 403 Forbidden | Verificado |
| **CP-18** | HU-19bis / RF-64 | Listar participantes para asistencia general | Fecha válida `YYYY-MM-DD` | HTTP 200 con participantes aprobados y estado del día | Verificado |
| **CP-19** | HU-19bis / RF-64 | Registrar presente en asistencia general | Lote con `presente: true` | HTTP 200 | Verificado |
| **CP-20** | HU-19bis / RF-64 | Registrar ausente en asistencia general | Lote con `presente: false` | HTTP 200 | Verificado |
| **CP-21** | HU-19bis / RF-64 | Registrar observación en asistencia general | Lote con `"observacion": "texto"` | HTTP 200, reflejado en consulta GET posterior | Verificado |
| **CP-25** | HU-19bis / RF-64 | Upsert de asistencia general (sin duplicados) | Doble registro en la misma fecha y participante | Documento único `{fecha}_{participanteId}` con el último valor | Verificado |
| **CP-27** | HU-19bis / RF-64 | Validación de formato de fecha | Query param con formato inválido (`fecha=no-es-fecha`) | HTTP 400 `"fecha inválida: se espera el formato YYYY-MM-DD"` | Verificado |

---

### 3.2 Clases de Equivalencia y Valores Límite

#### Particiones de Equivalencia (Equivalence Partitioning)

1. **Estado de Aprobación de Participante (`approval_status`):**
   - *Clase Válida:* `"approved"`, `"accepted"` (compatibilidad retroactiva) → Permite inscripción y asistencia.
   - *Clase Inválida 1:* `"pending"`, `"rejected"`, `"suspended"` → Rechazo HTTP 400 (`"El participante no está activo"`).
   - *Clase Inválida 2:* Participante inexistente (404 en Auth) → Rechazo HTTP 400 (`"El participante no existe"`).

2. **Roles y Permisos (RBAC):**
   - *Clase Válida Administrativa:* `role_admin`, `role_volunteer` → Acceso total a inscripción y asistencia general.
   - *Clase Válida Docente:* `role_teacher` con `is_external: false` y titular del curso → Acceso a actividades y asistencia de su curso.
   - *Clase Inválida:* `role_participant`, `role_company`, docente externo (`is_external: true`) → HTTP 403 Forbidden.
   - *Clase Sin Autenticación:* Token ausente o firma inválida → HTTP 401 Unauthorized.

3. **Modalidad del Curso (`mode`):**
   - *Clase Válida:* `"Presencial"`, `"Virtual"`, `"Mixta"` (exact casing) → Validación exitosa.
   - *Clase Inválida:* `"Remota"`, `""`, `"presencial"` (minúscula) → Error de validación HTTP 400.

4. **Formato de Fecha (Asistencia General):**
   - *Clase Válida:* Cadena coincidente con `2006-01-02` (ej. `2026-10-19`).
   - *Clase Inválida:* Formato ISO completo (`2026-10-19T14:00:00Z`), texto arbitrario o vacío → Error HTTP 400.

#### Análisis de Valores Límite (Boundary Value Analysis)

1. **Cupo del Curso (`capacity`):**
   - **Caso `capacity == nil`:** Sin restricción de cupo → Permite infinitas inscripciones.
   - **Límite inferior:** `inscriptos = 0` con `capacity = 2` → Permite inscripción.
   - **Límite inmediato anterior:** `inscriptos = capacity - 1` (ej. 1 de 2) → Permite inscripción; cupo pasa a límite exacto.
   - **Límite exacto:** `inscriptos = capacity` (ej. 2 de 2) → Rechaza con HTTP 400 `"El curso no tiene cupo disponible"`.
   - **Inscripción inactiva previa:** Un participante con inscripción previa en estado `"baja"` no computa para el límite y puede volver a inscribirse.

2. **Cálculo de Progreso Formativo (`ProgressService`):**
   - **Límite 0 actividades totales:** Evita división por cero; retorna `PorcentajeAsistencia: 0.0`.
   - **Límite 0 asistencias / N actividades:** Retorna `0.0%`.
   - **Límite N asistencias / N actividades:** Retorna `100.0%`.
   - **Asistencia foránea:** Asistencias registradas para actividades de otros cursos son descartadas por filtro de frontera.

3. **Rate Limiter (`RateLimiter` middleware):**
   - **Peticiones dentro del límite de rps:** HTTP 200 / respuesta normal.
   - **Límite + 1:** HTTP 429 Too Many Requests.

---

### 3.3 Tablas de Decisión

#### Tabla de Decisión 1: Inscripción a Curso (HU-18 / RF-18)

| Regla | Curso Publicado | Participante Existe | Participante Aprobado | Inscripción Activa Previa | Cupo Disponible | Resultado / Código HTTP |
|---|:---:|:---:|:---:|:---:|:---:|---|
| **R1** | No (Inexistente) | — | — | — | — | **HTTP 400** ("El curso no existe") |
| **R2** | No (En borrador) | — | — | — | — | **HTTP 400** ("El curso no está disponible") |
| **R3** | Sí | No | — | — | — | **HTTP 400** ("El participante no existe") |
| **R4** | Sí | Sí | No (pending/etc.) | — | — | **HTTP 400** ("El participante no está activo") |
| **R5** | Sí | Sí | Sí | Sí | — | **HTTP 400** ("El participante ya se encuentra inscripto...") |
| **R6** | Sí | Sí | Sí | No | No (`inscriptos == capacity`) | **HTTP 400** ("El curso no tiene cupo disponible") |
| **R7** | Sí | Sí | Sí | No | Sí (`inscriptos < capacity` o nil) | **HTTP 201** (Inscripción Exitosa) |

#### Tabla de Decisión 2: Registro de Asistencia por Actividad (HU-19 / RF-19)

| Regla | Rol Autorizado | Dueño del Curso (si Teacher) | Actividad Existe | Participante Inscripto | Participante Aprobado | Resultado / Código HTTP |
|---|:---:|:---:|:---:|:---:|:---:|---|
| **R1** | No (role_participant) | — | — | — | — | **HTTP 403** (Forbidden) |
| **R2** | Sí (role_teacher) | No (otro docente) | — | — | — | **HTTP 403** (Forbidden) |
| **R3** | Sí (admin/docente dueño) | Sí | No | — | — | **HTTP 404** (Actividad no encontrada) |
| **R4** | Sí (admin/docente dueño) | Sí | Sí | No | — | **HTTP 400** ("El participante no está inscripto...") |
| **R5** | Sí (admin/docente dueño) | Sí | Sí | Sí | No | **HTTP 400** ("El participante no está activo") |
| **R6** | Sí (admin/docente dueño) | Sí | Sí | Sí | Sí | **HTTP 200** (Asistencia Registrada / Upsert) |

---

### 3.4 Cobertura de Sentencia y de Decisión

Las métricas de cobertura fueron auditadas mediante `go test -coverprofile=coverage.out`:

#### Cobertura de Sentencia (Statement Coverage)
- **Cobertura Global Alcanzada:** **80.1%** de sentencias en todo el módulo `education/`.
- **Desglose por Capa Arquitectural:**
  - `internal/middlewares`: **100.0%** de cobertura de sentencia.
  - `internal/services`: **91.8%** de cobertura de sentencia.
  - `internal/repositories`: **84.9%** (Contentful y utilitarios) + repositorios Firestore probados con emulador.
  - `internal/controllers`: **75.4%** de cobertura de sentencia.
  - `internal/config`, `internal/models`, `internal/routes`: cubiertos en tests de integración y arranque.

#### Cobertura de Decisión (Branch / Decision Coverage)
Se implementaron pruebas basadas en tablas (*table-driven tests*) para recorrer exhaustivamente ramas verdaderas y falsas de sentencias de control:
- **Validaciones de entrada (`CourseService.CreateCourse`):** Tabla que fuerza cada condición de fallo de forma aislada (título vacío, descripción nula, docente responsable ausente, modalidad errónea, área formativa ausente, sin materiales).
- **Resolución de credenciales (`auth_middleware_test.go`):** Ramas de token extraído por Cookie vs. Header Authorization Bearer vs. Token ausente vs. Token con firma o expiración inválida.
- **Ramas de titularidad (`docente_controller_test.go`):** Evaluación de branches donde el docente es el titular del curso frente a ramas donde pertenece a otro docente o el curso es borrador.

---

## 4. Aseguramiento de la Calidad (QA) y Control de Calidad (QC)

### 4.1 Distinción de Prácticas de QA frente a QC

En el proyecto se diferencian claramente las actividades orientadas al **proceso** (QA) de las orientadas al **producto/código** (QC):

| Dimensión | Prácticas de Aseguramiento de la Calidad (QA) | Prácticas de Control de Calidad (QC) |
|---|---|---|
| **Enfoque** | Proceso preventivo. Define estándares, metodologías y mecanismos para evitar la introducción de defectos. | Producto / Detección. Inspecciona y ejecuta el software para identificar defectos antes del despliegue. |
| **Prácticas en el Proyecto** | - Inversión de Dependencias (SOLID) para desacoplar infraestructura.<br>- Regla de Quality Gate en CI (`ci-education` en `.github/workflows/pipeline.yml`): cobertura obligatoria ≥ 70%.<br>- Estandarización de linters (`golangci-lint`, `go vet`).<br>- Controles estáticos de seguridad en pipeline (Trivy para imágenes Docker, SARIF upload). | - Ejecución de pruebas unitarias con Testify (`go test`).<br>- Pruebas con emuladores locales (`cloud-firestore-emulator`).<br>- Pruebas de mutación (`run_mutation_tests.sh` con Gremlins).<br>- Pruebas dinámicas de seguridad DAST (`OWASP ZAP` baseline).<br>- Pruebas de estrés y carga (`run_load_tests.sh` con k6).<br>- Verificación de contrato e integridad (`tests/contract`, `tests/database`).<br>- Pruebas manuales estructuradas (`GUIA_PRUEBAS_INSCRIPCION_ASISTENCIA.md`). |

---

### 4.2 Plan de Pruebas (Test Plan)

#### 1. Objetivos
Asegurar que el microservicio `education-service` gestione de manera confiable, segura y desacoplada el catálogo formativo, las inscripciones, asistencias y cálculo de avance, garantizando alta cohesión y resiliencia ante caídas de dependencias externas.

#### 2. Alcance
- **En Alcance (In-Scope):**
  - Módulos de Cursos, Áreas Formativas y Materiales (Contentful).
  - Actividades, Asistencias e Inscripciones (Firestore).
  - Asistencia General Diaria (Firestore).
  - Integración de validación de participantes con `auth-service`.
  - Middlewares de seguridad (JWT, RBAC, Rate Limiting, CORS, Cabeceras OWASP).
  - Generación de reportes de integridad y contrato.
- **Fuera de Alcance (Out-of-Scope):**
  - Interfaces de usuario / Frontend Angular (definido explícitamente para etapas posteriores).
  - Modificación de esquema en base de datos Postgres Neon de `auth-service`.
  - Notificaciones por correo electrónico (HU-18/HU-26).

#### 3. Estrategia y Niveles de Prueba
- **Nivel 1: Pruebas Unitarias:** Ejecutadas en memoria con mocks y stubs sin dependencias externas.
- **Nivel 2: Pruebas de Integración de Componentes:** Ejecutadas con emulador local de Firestore y servidores HTTP simulados.
- **Nivel 3: Pruebas de Contrato:** Validación de firmas e interfaces esperadas por el cliente Frontend.
- **Nivel 4: Pruebas No Funcionales:** Carga (k6), Mutación (Gremlins), Seguridad estática y dinámica (gosec, govulncheck, OWASP ZAP).

#### 4. Criterios de Suspensión y Reanudación
- **Suspensión:** Falla en la compilación del paquete Go (`go build`), indisponibilidad del motor de Docker para emuladores, o caída de cobertura por debajo del umbral del 70%.
- **Reanudación:** Corrección del error de compilación, levantamiento de servicios requeridos o incorporación de tests para restaurar el umbral.

#### 5. Entregables de Prueba
- Perfil de cobertura de código (`coverage.out`).
- Reporte de contrato Pact (`reports/pact/education-pact-report.json`).
- Reporte de integridad Firestore (`reports/database/education-firestore-integrity.json`).
- Reporte de pruebas de mutación (`mutation_report.txt`).
- Reportes de seguridad SAST (`sast_report.json`), SCA (`sca_report.txt`) y DAST (`zap_report.html`).

---

### 4.3 Matriz de Riesgos (Probabilidad × Impacto)

La matriz prioriza los riesgos operacionales y funcionales del microservicio evaluando Probabilidad (1 a 5) e Impacto (1 a 5), resultando en un índice de Severidad (1 a 25):

| ID Riesgo | Descripción del Riesgo | Probabilidad (1-5) | Impacto (1-5) | Severidad (P × I) | Nivel | Estrategia de Mitigación Implementada |
|:---:|---|:---:|:---:|:---:|:---:|---|
| **R-01** | **Inconsistencia por duplicación en registros de asistencia:** Concurrencia o reintentos de red generan múltiples documentos para el mismo participante y actividad. | 4 | 4 | **16** | **Crítico** | Mitigación estructural en arquitectura: IDs determinísticos `{actividadId}_{participanteId}` en `asistencias` y `{fecha}_{participanteId}` en `asistencia_general`. Firestore realiza upsert atómico. |
| **R-02** | **Indisponibilidad o latencia excesiva en `auth-service`:** Fallo de red o timeout al validar participantes durante inscripción o asistencia. | 3 | 4 | **12** | **Alto** | `AuthServiceClient` propaga errores diferenciando 404 (participante no existe = 400 de negocio) de fallas de red/500 (traducidas a 502 Bad Gateway sin corromper el estado). |
| **R-03** | **Condición de carrera en cupo de cursos (`capacity`):** Múltiples solicitudes simultáneas superan el límite de inscriptos permitidos. | 3 | 3 | **9** | **Medio** | Validación previa en `EnrollmentService.Enroll` que cuenta inscripciones activas contra `capacity`. Riesgo residual documentado para incorporación futura de transacciones distribuidas. |
| **R-04** | **Incompatibilidad de esquema / versionado en Contentful:** Conflicto de versión optimista (`sys.version`) al actualizar o publicar cursos concurrentemente. | 2 | 4 | **8** | **Medio** | `CourseRepository` ejecuta lectura previa de `sys.version` y envía `X-Contentful-Version`, propagando `*ContentfulAPIError` con status 409 si ocurre colisión. |
| **R-05** | **Uso de claves JWT inseguras o no configuradas:** Inicio del servicio en entornos productivos con valores por defecto o vacíos en `JWT_SECRET`. | 1 | 5 | **5** | **Bajo** | Validación fail-fast en arranque (`cmd/server/main.go` / `config.go`): la aplicación aborta inmediatamente con log `FATAL` si la clave no cumple con entropía mínima o es un valor inseguro. |

---

### 4.4 Plantillas con Trazabilidad

#### Plantilla Formal de Caso de Prueba

```markdown
### [CP-XX] — [Título Descriptivo del Caso de Prueba]
- **Trazabilidad:**
  - **Historia de Usuario / RF:** [Ej. HU-18 / RF-18]
  - **Archivo de Código Fuente:** `internal/services/enrollment_service.go`
  - **Archivo de Test Unitario:** `internal/services/enrollment_enroll_test.go`
  - **Endpoint / Método:** `POST /api/cursos/{id}/inscripciones`
- **Objetivo:** [Breve descripción de lo que se valida]
- **Precondiciones:** [Estado de datos necesario en Contentful/Firestore/Auth]
- **Datos de Entrada:** [Payload JSON, parámetros de URL, headers, rol del token]
- **Pasos de Ejecución:**
  1. [Paso 1]
  2. [Paso 2]
  3. [Paso 3]
- **Resultado Esperado:** [Código HTTP esperado y estructura del body JSON]
- **Resultado Obtenido:** [Resultado real en ejecución manual o automatizada]
- **Estado:** [Aprobado / Fallido / Bloqueado]
```

#### Plantilla Formal de Defecto (Bug Report) con Trazabilidad

```markdown
### [DEF-XX] — [Título Conciso del Defecto]
- **Trazabilidad:**
  - **Historia de Usuario Afectada:** [Ej. HU-18 / HU-19]
  - **Componente / Módulo:** `education/internal/repositories/auth_service_client.go`
  - **Commit de Introducción / Detección:** [Hash o versión]
  - **Commit de Resolución:** [Hash de corrección]
  - **Test Automatizado Asociado (Regresión):** `TestAuthServiceClient_NotConfigured`
- **Severidad:** [Crítica / Alta / Media / Baja]
- **Prioridad:** [P1 / P2 / P3 / P4]
- **Precondiciones:** [Entorno o configuración requerida para reproducir]
- **Pasos para Reproducir:**
  1. Instanciar el cliente sin URL base o invocar métodos sobre un puntero nulo.
  2. Llamar a `GetParticipant(ctx, id, token)`.
- **Comportamiento Observado (Test Rojo):**
  - Excepción no controlada `panic: runtime error: invalid memory address or nil pointer dereference`.
- **Comportamiento Esperado (Test Verde):**
  - Retorno controlado del error `repositories.ErrAuthServiceNotConfigured` sin caída del servicio.
- **Causa Raíz:**
  - Falta de verificación de guardia `if c == nil` previa a la evaluación de propiedades de la estructura.
- **Resolución Aplicada:**
  - Inclusión de guardia de puntero nulo y retorno de error centinela.
```

---

## 5. Automatización del Testing

La automatización de pruebas en `education-service` constituye el pilar operativo para garantizar entregas continuas confiables, retroalimentación temprana a los desarrolladores y tolerancia cero a fallas en producción.

### 5.1 Pirámide de Pruebas Automatizadas

La estrategia del microservicio se alinea rigurosamente con el modelo clásico de la pirámide de automatización de pruebas (Mike Cohn / Martin Fowler), distribuyendo el esfuerzo y volumen de tests para maximizar cobertura y minimizar tiempos de ejecución:

```
               / \
              /   \
             / E2E \       <-- Cúspide (10%): Smoke en despliegue, k6 carga, DAST ZAP
            /-------\
           /  Integ  \     <-- Capa Media (20%): Emuladores Firestore/PubSub, Pact API
          /     &     \
         /     API     \
        /---------------\
       /    Unitarias    \ <-- Base (70%): Lógica de dominio, validaciones, mocks rápidos
      /___________________\
```

1. **Base de la Pirámide — Pruebas Unitarias (~70% del volumen):**
   - **Alcance:** Paquetes internos en `internal/services`, `internal/controllers`, `internal/middlewares`, `internal/events`, `internal/outbox` e `internal/repositories` (clientes desacoplados).
   - **Características:** Ejecución ultrarrápida (<2 segundos para todo el microservicio), ejecución en memoria sin sockets de red reales ni dependencias de I/O de disco.
   - **Técnicas:** Mocks con `github.com/stretchr/testify/mock`, servidores stub locales con `net/http/httptest` y aislamiento mediante inyección de dependencias por interfaces Go.
   - **Objetivo:** Detección inmediata de fallas lógicas, cálculo de notas, verificación de estados de inscripción y control de límites de tamaño.

2. **Capa Media — Pruebas de API e Integración (~20% del volumen):**
   - **Alcance:** Interacción entre capas del microservicio, adaptadores de persistencia y brokers de mensajería asíncrona.
   - **Componentes:**
     - `internal/repositories/firestore_*_test.go`: Pruebas de integración sobre emulador oficial `mtlynch/firestore-emulator:latest` para validar consultas compuestas, persistencia de colecciones y atomicidad.
     - `tests/integration/pubsub_test.go`: Flujo E2E de eventos asíncronos (`Productor -> Outbox Firestore -> Relay Worker -> Pub/Sub Topic -> Push Subscription -> Consumidor HTTP`).
     - `tests/contract/education_contract_test.go`: Verificación de contratos estilo Pact entre el proveedor Go (`Education-Service-Go`) y el consumidor frontend (`Frontend-Angular-Client`), garantizando consistencia de endpoints y payloads JSON.
     - `tests/database/firestore_integrity_test.go`: Validaciones automáticas de esquema NoSQL, campos obligatorios e índices compuestos.

3. **Cúspide de la Pirámide — Pruebas de Humo, Carga y Seguridad (~10% del volumen):**
   - **Alcance:** Inspección del sistema en ejecución completa simulando tráfico real y atacantes externos.
   - **Componentes:**
     - **Smoke Tests en Despliegue (`cd-smoke`):** Comprobación de salud en frío de `/api/salud` tras cada release en Cloud Run.
     - **Pruebas de Carga (`load_test.js` con k6):** Concurrencia de 20 VUs ejecutando 2430 solicitudes para validar latencia (`p(95) < 500ms`) y estabilidad del rate limiting defensivo.
     - **Pruebas de Seguridad Dinámica (DAST con OWASP ZAP):** Escaneo pasivo y activo del servidor HTTP contra OWASP Top 10.

---

### 5.2 Pruebas de API e Integración con MCP y Plataformas Externas

`education-service` interactúa con plataformas SaaS externas, proveedores de nube y protocolos de contexto modelo (MCP / AI), por lo que las pruebas de integración implementan patrones de aislamiento y emulación:

1. **Contentful (Headless CMS SaaS — Catálogo Educativo):**
   - **Arquitectura de Integración:** La verdad de los cursos, áreas formativas y materiales reside en Contentful (CDA para lectura, CMA para creación/edición de contenido y subida de archivos binarios).
   - **Estrategia de Prueba:** Los repositorios (`CourseRepository`, `TrainingAreaRepository`, `MaterialRepository`) son probados mediante stubs de `httptest.Server` que emulan la API REST de Contentful. Se validan:
     - Formato de headers (`Authorization: Bearer <token>`, `Content-Type: application/vnd.contentful.management.v1+json`).
     - Control de concurrencia optimista: Envío del header `X-Contentful-Version` y propagación de errores HTTP 409 Conflict.
     - Recepción de Webhooks (`ContentfulWebhookController`): Manejo de eventos de publicación y despublicación en `/api/webhooks/contentful` con validación de firma y sincronización de estadísticas asíncronas en Firestore (`KVSyncService`).

2. **Google Cloud Platform (GCP — Firestore, Pub/Sub, Cloud Storage):**
   - **Cloud Firestore:** Emulador local oficial en puerto `8090` o `8390`. Se verifica el almacenamiento en documentos y la generación determinística de IDs para evitar duplicados en asistencias (`{actividadId}_{participanteId}`).
   - **Cloud Pub/Sub & Outbox:** En lugar de publicar directamente a la red en la transacción del usuario, el servicio aplica el **Transactional Outbox Pattern**. Las pruebas de integración en `tests/integration/pubsub_test.go` verifican que el `RelayWorker` extrae los registros de la colección `outbox`, los despacha al topic de Pub/Sub y los marca como procesados de forma idempotente.
   - **Cloud Storage (GCS):** La subida de entregas de actividades (`SubmissionService`) valida extensiones y tamaños máximos. Se prueba la tolerancia a entornos donde `ENTREGAS_BUCKET` está vacío, permitiendo entregas textuales sin fallar.

3. **Proveedores de Inteligencia Artificial (OpenAI / Gemini / MCP):**
   - **Sugerencia de Habilidades (`SkillSuggestionService`):** Integra modelos LLM para inferir habilidades formativas a partir de las entregas de los estudiantes.
   - **Estrategia de Prueba:** Se inyectan interfaces simuladas de clientes (`OpenAIClient`, `GeminiClient`). Se evalúan:
     - Parsing de respuestas JSON estructuradas con listas de habilidades.
     - Mecanismo de reintentos exponenciales con límite superior de guardia (`attempt >= 0 && attempt <= s.maxRetries`) para evitar bucles infinitos.
     - Degradación elegante: si el proveedor externo responde con timeout o error 5xx, el servicio retorna lista vacía sin interrumpir el flujo principal del usuario docente.

4. **Microservicio de Autenticación (`auth-service`):**
   - `AuthServiceClient` consulta `GET /api/v1/participantes/{id}`. Las pruebas con `httptest.Server` cubren respuestas válidas (200 con `approval_status`), participantes no autorizados (403), no encontrados (404) y fallos de infraestructura (500/502), asegurando que un fallo en la plataforma externa no produzca `nil pointer dereference`.

---

### 5.3 Smoke Tests en Cada Despliegue

Los Smoke Tests (pruebas de humo) constituyen la primera línea de validación funcional post-despliegue en el pipeline de entrega continua (`.github/workflows/pipeline.yml` -> job `cd-smoke`):

1. **Objetivo:**
   Verificar que la versión recién desplegada en Google Cloud Run responde correctamente en su entorno de ejecución antes de habilitar el tráfico de usuarios reales.

2. **Flujo de Ejecución Automatizado:**
   - **Endpoint de Salud (`GET /api/salud`):** Valida que el runtime de Go inicializó correctamente, que las variables de entorno obligatorias fueron cargadas y que el puerto HTTP `:8210` atiende solicitudes con código `200 OK` y JSON `{"status":"ok"}`.
   - **Validación de Middlewares:** Envío de solicitudes sin cabeceras o con credenciales inválidas para verificar que la cadena de middlewares de seguridad (CORS, Security Headers, JWT Auth, Rate Limiter) no genera panics en tiempo de ejecución.
   - **Timeout de Guardia:** Cada solicitud de humo cuenta con un timeout estricto de 10 segundos para prevenir bloqueos por deadlocks en arranques en frío.

3. **Política de Fallo y Rollback Automático (`cd-rollback`):**
   - Si el smoke test falla (código HTTP != 200, timeout o rechazo de conexión), el pipeline ejecuta la condición:
     ```yaml
     if: always() && (needs.cd-smoke.result == 'failure' || needs.cd-dast-zap.result == 'failure' || needs.cd-load-k6.result == 'failure')
     ```
   - El job de rollback redirige el 100% del tráfico a la revisión previa estable en Cloud Run y notifica el incidente, impidiendo que despliegues defectuosos afecten la disponibilidad del servicio.

---

### 5.4 Pipeline de CI/CD (`.github/workflows/calidad-tp.yml`)

El aseguramiento estático y dinámico de la calidad está completamente orquestado a través del workflow de GitHub Actions `.github/workflows/calidad-tp.yml`:

#### 1. Disparadores y Política de Concurrencia
- **Eventos:** `push` en cualquier rama cuando se modifican rutas relevantes (`education/**`, `scripts/calidad/**`, `.github/workflows/calidad-tp.yml`), además de ejecución manual `workflow_dispatch`.
- **Control de Concurrencia:**
  ```yaml
  concurrency:
    group: calidad-tp-${{ github.ref }}
    cancel-in-progress: true
  ```
  Cancela ejecuciones obsoletas de un mismo branch si se realiza un nuevo push, ahorrando minutos de cómputo en CI.

#### 2. Matriz de Versiones y Dependencias de Entorno
- Go: `1.26`
- SonarScanner CLI: `8.1.0.6389`
- Fecha de Línea Base de Código Nuevo: `2026-08-01`
- Herramientas de Seguridad con Versiones Fijas (determinismo estricto):
  - `gosec`: `v2.29.0`
  - `govulncheck`: `v1.1.4`
  - `gitleaks`: `v8.21.2`
  - `semgrep`: `1.95.0`
  - `trivy`: `0.57.1`

#### 3. Estructura de Jobs en Paralelo

```
[ Push a rama con cambios en education/ ]
             |
             +---------------------------------------+
             |                                       |
    [ Job 1: sonarqube-education ]          [ Job 2: seguridad-education ]
    - Levanta SonarQube :9000               - SAST (gosec v2.29.0)
    - Levanta Firestore Emulator :8390      - SCA (govulncheck v1.1.4)
    - Instala SonarScanner CLI (SHA-256)    - Secret Detection (Gitleaks v8.21.2)
    - Configura Quality Gate TP-Calidad-N3   - Code Patterns (Semgrep 1.95.0)
    - Registra Línea Base del 1/8/2026      - Container Audit (Trivy 0.57.1)
    - Ejecuta analizar.sh (go test)         - Consolida resumen_seguridad.py
    - Exige 100% Pruebas en Verde           - Publica resumen y artefactos
    - Publica Reportes y Step Summary
```

- **Job `sonarqube-education`:**
  - Instancia servicios temporales en contenedores: `sonarqube:26.8.0.126808-community` y `mtlynch/firestore-emulator:latest`.
  - Verifica la integridad del scanner descargado contrastando el checksum SHA-256 oficial antes de la descompresión.
  - Genera credenciales de un solo uso con `openssl rand -hex 16` y token temporal de SonarQube enmascarado en los logs (`::add-mask::`).
  - Aplica las reglas del Quality Gate Nivel 3 (`TP-Calidad-N3`) mediante `scripts/calidad/configurar_sonar.sh`.
  - Ejecuta `scripts/calidad/analizar.sh education`, corriendo `go test -count=1 -json -coverprofile=coverage.out ./...` y midiendo cobertura real con el emulador activo.
  - Valida el criterio estricto: si alguna prueba unitaria o de integración falla, el paso `Exigir pruebas en verde` aborta el job con error, impidiendo que una corrida con pruebas rotas reporte falsos aprobados de cobertura.
  - Almacena artefactos de análisis (`coverage.out`, `test-report.json`, métricas históricas) por 30 días.

- **Job `seguridad-education` (Análisis Multicapa):**
  - Corre en paralelo al análisis de SonarQube para minimizar la latencia total del pipeline.
  - Aplica 5 motores de seguridad complementarios sin ejecutar código:
    1. **`gosec`:** Análisis estático de código fuente Go (prevención de XSS, desbordamientos, permisos inseguros y llamadas de sistema).
    2. **`govulncheck`:** Detección de vulnerabilidades conocidas (CVEs) en las dependencias directas e indirectas de `go.mod` y `go.sum`.
    3. **`gitleaks`:** Detección de tokens, claves privadas o secretos hardcodeados con ofuscación de reportes (`--redact`).
    4. **`semgrep`:** Detección de patrones sintácticos inseguros utilizando el ruleset especializado `p/golang`.
    5. **`trivy`:** Construcción de la imagen Docker de `education` y escaneo completo de vulnerabilidades tanto a nivel de paquetes del sistema operativo base como de librerías.
  - Ejecuta `scripts/calidad/resumen_seguridad.py` para emitir el veredicto unificado en el `GITHUB_STEP_SUMMARY`.

---

### 5.5 Una Prueba de Regresión por Cada Defecto Corregido

El proyecto adopta como regla inmutable la metodología **Bug-Driven Regression Testing**: ante cualquier defecto identificado en revisión, auditoría o mutación, es mandatorio escribir un test automatizado que reproduzca el defecto (estado Rojo) antes de aplicar la corrección, asegurando que permanezca en verde (Green) de forma permanente:

| ID Defecto | Componente Afectado | Causa Raíz Detectada | Prueba de Regresión Automatizada | Mecanismo de Verificación |
|:---:|---|---|---|---|
| **DEF-01** | `internal/repositories/auth_service_client.go` | Panic por dereferencia de puntero nulo al invocar `GetParticipant` sobre cliente no inicializado (`c == nil`). | `TestAuthServiceClient_NotConfigured` | Aserción de retorno seguro del error `repositories.ErrAuthServiceNotConfigured` sin caída del proceso. |
| **DEF-02** | `internal/services/enrollment_service.go` | Sobreasignación de vacantes en cursos con cupo limitado ante solicitudes simultáneas. | `TestEnrollmentService_Enroll_CourseFull` | Aserción de rechazo estricto con `ErrCourseFull` al alcanzar `capacity`. |
| **DEF-03** | `internal/services/skill_suggestion_service.go` | Bucle infinito y timeout cuando un mutante invierte el paso del iterador (`attempt--`). | `TestSkillSuggestionService_RetryLimit` | Cota de seguridad en condición de iterador `attempt >= 0 && attempt <= s.maxRetries`, eliminando el mutante sobreviviente. |
| **DEF-04** | `internal/services/material_service.go` | Falta de aserción exacta del mensaje de error en archivos de más de 50 MB / 500 MB (mutantes aritméticos `<` vs `<=`). | `TestMaterialService_Create_FileTooLargeExact` | Verificación con `assert.Equal` del mensaje textual exacto `"el archivo supera el límite de 50MB"`. |
| **DEF-05** | `internal/services/submission_service.go` | Riesgo de sobreescritura accidental de estados de curso activos a estado `"baja"` durante la entrega de tareas. | `TestSubmissionService_PreservesNonBajaStatus` | Prueba que valida la persistencia intacta de estados `"aprobado"` o `"cursando"` frente a operaciones de entrega. |

---

## 6. Resultados de Ejecución y Métricas Verificadas

A continuación se resumen los resultados empíricos obtenidos de la ejecución integral de las suites automatizadas en el servicio:

| Suite / Herramienta | Alcance Evaluado | Resultado / Exit Code | Métrica Principal |
|---|---|:---:|---|
| **Unit Tests (`go test ./...`)** | 15 paquetes internos (controladores, servicios, repositorios, outbox, eventos, middlewares) | **PASS** (0) | **100%** pruebas superadas (0 fallos). |
| **Testing de Mutación (`gremlins`)** | Mutadores aritméticos, condicionales y de asignación en `internal/...` | **PASS** (0) | **91.36%** de eficacia de mutación (Mutator coverage: 82.71%, 782 mutantes eliminados). |
| **SAST (`gosec`)** | Análisis estático de vulnerabilidades e inyecciones de código | **PASS** (0) | **0** vulnerabilidades detectadas (`sast_report.json`). |
| **SCA (`govulncheck`)** | Vulnerabilidades conocidas en dependencias Go del grafo del módulo | **PASS** (0) | **0** vulnerabilidades reportadas (`sca_report.txt`). |
| **DAST (`OWASP ZAP`)** | Escaneo dinámico pasivo y activo baseline sobre la API en ejecución | **PASS** (0) | **66** reglas de seguridad superadas, 0 alertas de fallo (`zap_report.html`). |
| **Carga / Rendimiento (`k6`)** | Spike y concurrencia sostenida de 20 VUs (2430 requests HTTP) | **PASS** (0) | Latencia **p(95) = 1.25ms** (umbral `< 500ms`), **100%** checks conformes con rate limiting defensivo. |


