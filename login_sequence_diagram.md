# Diagrama de Secuencia — Flujo de Login
> Sistema Informático Buen Vivir · Firebase Auth · Firestore · Go MVC · Google Cloud API Gateway

```mermaid
sequenceDiagram
    autonumber

    actor Usuario as 👤 Usuario

    box rgb(13,71,161) Frontend — Angular PWA
        participant LoginComp as LoginComponent<br/>(app/features/auth)
        participant AuthSvc as AuthService<br/>(Angular · HttpClient)
        participant Interceptor as HTTP Interceptor<br/>(JWT Bearer)
        participant Guard as AuthGuard<br/>(Route Guard)
        participant Store as State Store<br/>(NgRx / Signals)
    end

    box rgb(230,81,0) Firebase / Google Identity
        participant FirebaseAuth as Firebase Auth SDK<br/>(signInWithEmailAndPassword)
        participant FirebaseIdP as Firebase Identity<br/>Platform (GCP)
    end

    box rgb(0,96,100) Google Cloud
        participant APIGateway as API Gateway<br/>(openapi.yaml · JWT verify)
        participant CloudLogging as Cloud Logging<br/>(structured JSON)
    end

    box rgb(30,77,43) Auth Microservice — Go (Cloud Run :8081)
        participant Router as Routes<br/>(http.ServeMux)
        participant Controller as AuthController<br/>(JSON decode/encode)
        participant Service as AuthService<br/>(business logic · bcrypt)
        participant Repository as UserRepository<br/>(GORM interface)
    end

    box rgb(74,20,140) Bases de Datos
        participant Firestore as Firestore<br/>(NoSQL · users collection)
        participant CloudSQL as Cloud SQL<br/>(PostgreSQL · GORM)
    end

    %% ══════════════════════════════════════════
    %% PASO 1 — Usuario ingresa credenciales
    %% ══════════════════════════════════════════
    Usuario->>LoginComp: Completa form<br/>{ email, password }
    LoginComp->>LoginComp: Reactive Form validation<br/>(Validators.required · email format)

    alt Formulario inválido
        LoginComp-->>Usuario: ❌ Muestra errores de validación<br/>(inline form errors)
    end

    %% ══════════════════════════════════════════
    %% PASO 2 — Firebase Auth (Identity Provider)
    %% ══════════════════════════════════════════
    LoginComp->>FirebaseAuth: signInWithEmailAndPassword<br/>(email, password)
    FirebaseAuth->>FirebaseIdP: POST /v1/accounts:signInWithPassword<br/>{ email, password, returnSecureToken: true }

    alt Credenciales inválidas en Firebase
        FirebaseIdP-->>FirebaseAuth: 400 INVALID_PASSWORD / EMAIL_NOT_FOUND
        FirebaseAuth-->>LoginComp: FirebaseAuthError<br/>(auth/wrong-password · auth/user-not-found)
        LoginComp-->>Usuario: ❌ "Credenciales inválidas"
    end

    FirebaseIdP-->>FirebaseAuth: 200 OK<br/>{ idToken (JWT), refreshToken,<br/>expiresIn: 3600, localId }
    FirebaseAuth-->>LoginComp: UserCredential<br/>{ user.uid, user.email,<br/>user.getIdToken() → Firebase JWT }

    Note over LoginComp,FirebaseAuth: Firebase JWT firmado con RS256<br/>Contiene: uid · email · roles (custom claims)

    %% ══════════════════════════════════════════
    %% PASO 3 — Llamada al Backend vía API Gateway
    %% ══════════════════════════════════════════
    LoginComp->>AuthSvc: login(email, password, firebaseIdToken)
    AuthSvc->>Interceptor: HttpClient.post(<br/>POST /api/v1/auth/login,<br/>{ email, password })<br/>+ idToken en header
    Interceptor->>Interceptor: Agrega header:<br/>Authorization: Bearer {Firebase JWT}
    Interceptor->>APIGateway: POST /api/v1/auth/login<br/>Content-Type: application/json<br/>Authorization: Bearer {Firebase JWT}

    %% ══════════════════════════════════════════
    %% PASO 4 — API Gateway valida el JWT
    %% ══════════════════════════════════════════
    APIGateway->>APIGateway: Verifica Firebase JWT<br/>• Firma RS256 con JWKS de Google<br/>• Expiry (exp claim)<br/>• Audience (aud = project-id)<br/>• Issuer (iss = accounts.google.com)

    alt JWT inválido / expirado
        APIGateway-->>AuthSvc: 401 Unauthorized<br/>{ "message": "JWT validation failed" }
        AuthSvc-->>LoginComp: Error 401
        LoginComp-->>Usuario: ❌ "Sesión expirada, reintentá"
    end

    APIGateway->>Router: Forward request<br/>POST /api/v1/auth/login<br/>{ email, password } JSON<br/>x-forwarded-user: uid

    %% ══════════════════════════════════════════
    %% PASO 5 — Auth Service Go (MVC)
    %% ══════════════════════════════════════════
    Router->>Controller: HandleFunc("/api/v1/auth/login")<br/>*http.Request

    Controller->>Controller: json.NewDecoder(r.Body).Decode<br/>→ LoginRequest{ email, password }

    alt Body inválido / campos vacíos
        Controller-->>APIGateway: 400 Bad Request<br/>{ "error": "email y contraseña son obligatorios" }
        APIGateway-->>AuthSvc: 400
        AuthSvc-->>LoginComp: Error 400
        LoginComp-->>Usuario: ❌ Error de request
    end

    Controller->>Service: authService.Login(ctx, &LoginRequest)

    Service->>Service: Valida email != "" && password != ""
    Service->>Repository: userRepo.FindByEmail(ctx, email)

    %% ══════════════════════════════════════════
    %% PASO 6 — Lectura en Firestore
    %% ══════════════════════════════════════════
    Repository->>Firestore: GET /databases/(default)/documents<br/>collection: users<br/>where: email == req.Email

    alt Usuario no existe en Firestore
        Firestore-->>Repository: 404 / empty result
        Repository-->>Service: error("usuario no encontrado")
        Service-->>Controller: error("credenciales inválidas")
        Controller-->>APIGateway: 401 Unauthorized<br/>{ "error": "credenciales inválidas" }
        APIGateway-->>AuthSvc: 401
        AuthSvc-->>LoginComp: Error 401
        LoginComp-->>Usuario: ❌ "Credenciales inválidas"
    end

    Firestore-->>Repository: Document snapshot<br/>{ id, email, name, role,<br/>passwordHash, createdAt }
    Repository-->>Service: *User{ ID, Email, Name,<br/>Role, Password(hash) }

    %% ══════════════════════════════════════════
    %% PASO 6b — Fallback Cloud SQL (perfil extendido)
    %% ══════════════════════════════════════════
    Service->>CloudSQL: SELECT * FROM users WHERE email = $1<br/>(GORM · parameterized query)
    CloudSQL-->>Service: UserRecord (rol, permisos,<br/>datos adicionales)

    %% ══════════════════════════════════════════
    %% PASO 7 — Verificación de password y emisión de token
    %% ══════════════════════════════════════════
    Service->>Service: bcrypt.CompareHashAndPassword<br/>(storedHash, inputPassword)

    alt Password incorrecto
        Service-->>Controller: error("credenciales inválidas")
        Controller-->>APIGateway: 401 Unauthorized
        APIGateway-->>AuthSvc: 401
        LoginComp-->>Usuario: ❌ "Credenciales inválidas"
    end

    Service->>Service: jwt.NewWithClaims(RS256)<br/>Payload: { sub: uid, email,<br/>role, exp: now+3600, iss: buenvivir }
    Service->>Service: token.SignedString(jwtSecret)

    Service-->>Controller: &LoginResponse{<br/>Token: "jwt-token",<br/>User: *User }

    %% ══════════════════════════════════════════
    %% PASO 8 — Respuesta HTTP al Gateway
    %% ══════════════════════════════════════════
    Controller->>Controller: w.Header().Set("Content-Type","application/json")<br/>w.WriteHeader(200)<br/>json.NewEncoder(w).Encode(resp)

    Controller-->>APIGateway: 200 OK<br/>Content-Type: application/json<br/>{ "token": "...", "user": { id, email, name, role } }

    %% ══════════════════════════════════════════
    %% PASO 9 — Audit log
    %% ══════════════════════════════════════════
    Service--)CloudLogging: Publish structured log<br/>{ event: "user.login",<br/>uid, email, role,<br/>timestamp, ip, status: "success" }

    Note over CloudLogging: Asíncrono · no bloquea respuesta

    %% ══════════════════════════════════════════
    %% PASO 10 — Frontend procesa respuesta
    %% ══════════════════════════════════════════
    APIGateway-->>Interceptor: 200 OK · JSON response
    Interceptor-->>AuthSvc: LoginResponse{ token, user }
    AuthSvc->>Store: dispatch(AuthActions.loginSuccess({ token, user }))
    Store->>Store: Persiste token en localStorage<br/>Actualiza estado: isAuthenticated = true
    AuthSvc-->>LoginComp: Observable<LoginResponse> resolved

    %% ══════════════════════════════════════════
    %% PASO 11 — Navegación post-login
    %% ══════════════════════════════════════════
    LoginComp->>Guard: Router.navigate(['/dashboard'])
    Guard->>Store: Consulta isAuthenticated del Store
    Store-->>Guard: true
    Guard-->>LoginComp: canActivate() → true
    LoginComp->>Usuario: ✅ Redirige a /dashboard
```

---

## Resumen del flujo por fases

| # | Fase | Actor principal | Protocolo / Formato |
|---|------|----------------|---------------------|
| 1 | Validación de form | `LoginComponent` | Reactive Forms (Angular) |
| 2 | Autenticación Firebase | `Firebase Auth SDK → Firebase IdP` | HTTPS · REST · JSON |
| 3 | Llamada al backend | `AuthService → HTTP Interceptor` | HttpClient · Bearer JWT |
| 4 | Validación JWT | `API Gateway` | RS256 · JWKS Google |
| 5 | Routing HTTP | `http.ServeMux → AuthController` | Go net/http |
| 6 | Lectura de usuario | `UserRepository → Firestore` | Firestore SDK · Document |
| 6b | Perfil extendido | `UserRepository → Cloud SQL` | GORM · SQL parameterizado |
| 7 | Verificación password | `AuthService` | bcrypt + JWT signing (RS256) |
| 8 | Respuesta JSON | `AuthController → API Gateway` | `200 OK` · `encoding/json` |
| 9 | Audit log | `AuthService → Cloud Logging` | Pub/Sub async · JSON structured |
| 10 | Persistencia de estado | `Store (NgRx / Signals)` | localStorage · in-memory |
| 11 | Navegación | `Router → AuthGuard` | Angular Router |

---

## Flujos alternos cubiertos

> [!WARNING]
> **Credenciales inválidas en Firebase** → `FirebaseAuthError` → UI error inline. No llega al backend.

> [!WARNING]
> **JWT inválido/expirado en API Gateway** → `401` antes de tocar el microservicio Go.

> [!WARNING]
> **Usuario no existe en Firestore** → `UserRepository` devuelve error → `401` desde `AuthController`.

> [!WARNING]
> **Password incorrecto (bcrypt)** → `AuthService` devuelve error → `401` desde `AuthController`.

> [!NOTE]
> El audit log a **Cloud Logging** es **asíncrono** — no bloquea el path crítico de respuesta al usuario.
