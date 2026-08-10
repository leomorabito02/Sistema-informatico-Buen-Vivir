# Arquitectura de Servicios — Sistema Buen Vivir

> **Flujo de una request:**
> `Cliente → API Gateway (Nginx) → Microservicio → Controller → Service → Repository → DB`

Todos los microservicios del sistema siguen la misma estructura en capas. Este documento describe las responsabilidades de cada capa y las decisiones técnicas y de seguridad que las rigen de forma transversal.

---

## 0. API Gateway (`api-gateway/nginx.conf`)

**Responsabilidad:** Punto de entrada único al sistema. El cliente nunca habla directamente con un microservicio. El Gateway recibe toda la tráfico, lo enruta al servicio correcto y propaga cabeceras de red.

### Rol en la arquitectura

El API Gateway actúa como **reverse proxy** (Nginx). Sus responsabilidades son estrictamente de red y enrutamiento:

| Responsabilidad | Implementación actual |
|---|---|
| Enrutamiento por prefijo de ruta | `location /api/v1/auth/` → `auth-service:8080` |
| Enrutamiento por regex de entidades | `location ~ ^/api/v1/(users\|roles\|...)` → `auth-service:8080` |
| Propagación de IP real del cliente | `X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto` |
| Terminación de preflight CORS | Responde `204` a `OPTIONS` antes de que lleguen a los servicios |
| Health check propio | `GET /health` → `200 OK` sin tocar ningún servicio |

### Qué NO hace el Gateway (y dónde se maneja)

El Gateway actual **no valida tokens ni autoriza**. Toda la autenticación y autorización ocurre dentro de cada microservicio:

```
Cliente → Gateway (solo enruta) → auth-service (valida JWT, RBAC, ownership)
```

Esto es intencional: cada servicio es responsable de su propia seguridad, y el Gateway puede ser reemplazado o escalado sin afectar la lógica de autorización.

> **Deuda técnica conocida:** El CORS del Gateway está configurado con `Access-Control-Allow-Origin: *`. En producción debe restringirse a los orígenes del frontend, alineándose con la variable `CORS_ALLOWED_ORIGINS` que cada microservicio ya usa internamente.

### Relación con el Rate Limiter de los servicios

El Gateway no implementa rate limiting. Cada microservicio tiene su propio rate limiter (global + estricto) que actúa como **fallback** en caso de que tráfico elevado pase a través del Gateway. En producción el rate limiting primario debería configurarse en el Gateway o en un WAF delante de él.

---

## A. Entry Point & Routes (`cmd/server/main.go`)

**Responsabilidad:** Inicializar dependencias mediante Inyección de Dependencias (DI), levantar el servidor HTTP y registrar endpoints mapeándolos a los controladores correspondientes.

### Composición y arranque

La función `main` es la única que conoce las implementaciones concretas. Instancia repositorios, servicios y controladores en orden de dependencia (repository → service → controller) y los inyecta hacia arriba. Ninguna capa inferior conoce a la superior.

```
Repositories → Services → Controllers → Router
```

El servidor falla con `os.Exit(1)` ante cualquier dependencia crítica no disponible (base de datos, proveedor de identidad externo). No hay arranques en estado degradado.

### Middleware global (aplicado a todas las rutas)

| Middleware | Propósito |
|---|---|
| `middleware.RequestID` | Trazabilidad de requests entre servicios |
| `middleware.RealIP` | Extrae la IP real detrás de proxies/API Gateway |
| `middleware.Recoverer` | Captura panics y evita que el servidor caiga |
| Logging estructurado (`slog`) | Logs JSON compatibles con Google Cloud Logging |
| Security Headers (OWASP) | `X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`, `Strict-Transport-Security` |
| CORS con whitelist | Orígenes permitidos inyectados desde variable de entorno `CORS_ALLOWED_ORIGINS` |
| Rate Limiter global | 100 req/min por IP — fallback de protección DDoS si el API Gateway lo permite pasar |

### Rate Limiting diferenciado por sensibilidad del endpoint

Los endpoints sensibles (login, registro, autenticación con proveedor externo) aplican un `StrictRateLimiter` adicional de **10 intentos por minuto por IP**, independiente del límite global.

```
POST /auth/login    → StrictRateLimiter (10 req/min)
POST /auth/register → StrictRateLimiter (10 req/min)
GET  /users/:id     → RateLimiter global (100 req/min)
```

### Autorización por capas en el router

Las rutas protegidas aplican middlewares en orden encadenado:

1. `AuthMiddleware` — valida el JWT e inyecta `userID`, `roleID`, `isExternal` en el contexto.
2. `RequireRole(...)` — verifica que el rol del token pertenezca a la lista de roles permitidos para ese grupo de rutas. `role_admin` y `role_dev` tienen bypass global.
3. Lógica de ownership — validaciones de propiedad de recurso delegadas al Service (ej: un voluntario solo puede editar su propio perfil).

---

## B. Controller / Presentación (`internal/controllers/`)

**Responsabilidad:** Manejar la entrada/salida HTTP. Extraer parámetros (query string, URL params, JSON body), delegar al Service y formatear la respuesta (JSON, códigos de estado). No contiene lógica de negocio.

### Principios

- **Sin lógica de negocio.** Si hay una condición de negocio, pertenece al Service.
- **Sin acceso a la base de datos.** El Controller no importa repositorios ni `*gorm.DB`.
- **Validación de forma, no de negocio.** Verifica que el body sea parseable y que los campos requeridos estén presentes; las reglas de negocio las rechaza el Service.
- **Errores opacos hacia afuera.** Los mensajes de error expuestos al cliente son genéricos. Los detalles internos se loguean en el servidor, nunca en la respuesta HTTP.

### Extracción del contexto de seguridad

El controlador puede leer del contexto los valores que el `AuthMiddleware` inyectó:

```go
userID := r.Context().Value(middlewares.UserIDKey).(string)
roleID := r.Context().Value(middlewares.RoleIDKey).(string)
```

Estos valores son de solo lectura para el controlador. Nunca se aceptan del body o query string del cliente.

---

## C. Service / Lógica de Negocio (`internal/services/`)

**Responsabilidad:** Núcleo de la aplicación. Reglas de negocio, cálculos, flujos lógicos y orquestación. Totalmente agnóstico del protocolo de entrada (HTTP, gRPC, CLI).

### Principios

- **Interfaces sobre implementaciones.** Cada servicio expone una interfaz (ej: `AuthService`, `JWTService`). Los consumidores dependen de la interfaz, no del struct concreto. Esto permite mocking en tests y desacoplamiento entre paquetes.
- **Sin importación circular.** Los paquetes de capas inferiores nunca importan paquetes de capas superiores. Si un Service necesita tipos del middleware, define su propia interfaz local.

### Seguridad en el flujo de autenticación

| Mecanismo | Implementación |
|---|---|
| Hash de contraseñas | `bcrypt` con cost factor 14 — resistente a ataques de fuerza bruta |
| Tokens de refresh | Generados con `crypto/rand` (CSPRNG) — nunca predecibles |
| Almacenamiento de refresh tokens | Solo se persiste el **hash SHA-256** del token, nunca el valor en texto plano |
| Rotación de sesión | Al hacer login, todos los refresh tokens anteriores del usuario se revocan antes de emitir el nuevo |
| Limpieza de tokens expirados | Se ejecuta en goroutine de segundo plano tras cada login para evitar acumulación |
| Access token (`sub` + `role` + `is_external`) | El JWT embebe el rol para que el middleware autorice sin consultar la BD en cada request |
| Verificación de rol en login | Se valida que el rol del request coincida con el rol registrado — previene escalada de privilegios |
| Token de identidad externo (Google/Firebase) | Se valida con el SDK oficial de Firebase antes de emitir tokens propios. El sistema nunca confía directamente en claims del proveedor externo |

### Transacciones y consistencia

Las operaciones que afectan múltiples tablas (ej: crear usuario + crear perfil de rol) se ejecutan dentro de una transacción GORM. Si cualquier paso falla, se hace rollback automático. Los métodos de repositorio aplican el mismo patrón con `defer recover()` para garantizar rollback incluso ante panics.

### Ownership a nivel de servicio

La validación de que un usuario solo puede operar sobre sus propios recursos se implementa en el Service, no en el Controller ni en el Repository. El Service lee el `roleID` y `userID` del contexto y los compara con el recurso solicitado antes de proceder.

---

## D. Repository / Acceso a Datos (`internal/repositories/`)

**Responsabilidad:** Comunicarse con el motor de persistencia. Traduce entidades de dominio a consultas concretas. Es la única capa que importa el driver de base de datos.

### Principios

- **Interfaz obligatoria.** Cada repositorio expone una interfaz que el Service consume. La implementación concreta se registra en `main.go`. Garantiza testabilidad mediante mocks sin base de datos real.
- **Parámetros en todas las consultas.** Se usan exclusivamente parámetros posicionales (`WHERE email = ?`) o métodos seguros del ORM. Cero concatenación de strings en queries — previene SQL Injection.
- **Propagación de contexto.** Todas las operaciones reciben `context.Context` y lo propagan con `.WithContext(ctx)`. Permite cancelación de queries por timeout o cierre de conexión del cliente.
- **Soft delete.** Los modelos con `DeletedAt` de GORM implementan borrado lógico automáticamente. Las queries de lectura filtran registros eliminados sin código adicional.

### Transacciones con recovery

Todas las mutaciones (Create, Update, Delete) se envuelven en `db.Transaction(...)` con un `defer recover()` que garantiza el rollback incluso ante un panic inesperado:

```go
return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    defer func() {
        if r := recover(); r != nil { tx.Rollback() }
    }()
    return tx.Create(entity).Error
})
```

### Pool de conexiones configurado

| Parámetro | Valor |
|---|---|
| `MaxOpenConns` | 10 |
| `MaxIdleConns` | 5 |
| `ConnMaxLifetime` | 30 minutos |
| `ConnMaxIdleTime` | 5 minutos |

---

## E. Infraestructura Docker

El binario del servicio se construye en una imagen multi-stage:

1. **Stage `builder`** — imagen `golang:alpine` compila con `CGO_ENABLED=0` y strips de debug (`-ldflags="-s -w"`).
2. **Stage final** — imagen `alpine` mínima sin toolchain Go. Solo contiene el binario compilado, certificados CA y datos de timezone.

El proceso corre como usuario no-root (`appuser`) sin capacidades de escritura sobre el binario (`chmod 500`). El binario en producción no puede ser modificado desde dentro del contenedor.

Toda la configuración sensible (DSN, secrets JWT, credenciales Firebase, orígenes CORS) se inyecta exclusivamente mediante **variables de entorno**, nunca hardcodeada ni leída de archivos versionados.
