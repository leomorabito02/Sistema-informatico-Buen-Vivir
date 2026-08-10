# Guía de Postman con API Gateway Local

El API Gateway local (NGINX) está escuchando en el puerto `8080`.
Internamente, el `auth-service` está mapeando todas estas rutas usando `chi` (ver `cmd/server/main.go`).

> [!IMPORTANT]
> Debes enviar el Header `Authorization: Bearer <TU_TOKEN_JWT>` (obtenido en login/register) para acceder a cualquier ruta protegida (las que no son login/register).

## Autenticación (Públicas)

### 1. POST - Login
**URL:** `http://localhost:8080/api/v1/auth/login`
```json
{ "email": "usuario@ejemplo.com", "password": "mypassword123", "role": "admin" }
```

### 2. POST - Register
**URL:** `http://localhost:8080/api/v1/auth/register`
```json
{ "email": "nuevo@ejemplo.com", "password": "mypassword123", "role": "volunteer" }
```

### 3. POST - Google Login (Firebase Auth)
**URL:** `http://localhost:8080/api/v1/auth/google`
```json
{ "id_token": "eyJhbGciOiJSUzI1Ni...", "google_id": "10343...", "role": "company" }
```

## Usuarios (Protegidas)

- **GET** `http://localhost:8080/api/v1/auth/me` (Devuelve info del token actual)
- **POST** `http://localhost:8080/api/v1/users/` (Crear usuario)
- **GET** `http://localhost:8080/api/v1/users/` (Listar todos)
- **GET** `http://localhost:8080/api/v1/users/{id}` (Obtener uno)
- **PUT** `http://localhost:8080/api/v1/users/{id}` (Actualizar)
- **DELETE** `http://localhost:8080/api/v1/users/{id}` (Eliminar)

*Nota: Todas las entidades siguientes comparten exactamente la misma estructura CRUD (POST, GET, GET /{id}, PUT /{id}, DELETE /{id}).*

## Roles (Protegidas)
- `http://localhost:8080/api/v1/roles/`
```json
{ "name": "Admin", "description": "Administrador del sistema" }
```

## Compañías (Protegidas)
- `http://localhost:8080/api/v1/companies/`
```json
{ "name": "Empresa S.A.", "cuit": "30-12345678-9", "address": "Calle Falsa 123" }
```

## Profesores (Teachers) (Protegidas)
- `http://localhost:8080/api/v1/teachers/`
```json
{ "first_name": "Juan", "last_name": "Pérez", "specialty": "Matemáticas" }
```

## Participantes (Protegidas)
- `http://localhost:8080/api/v1/participants/`
```json
{ "first_name": "Ana", "last_name": "García", "age": 25 }
```

## Permisos (Protegidas)
- `http://localhost:8080/api/v1/permissions/`
```json
{ "resource": "users", "action": "create" }
```

## Voluntarios (Protegidas)
- `http://localhost:8080/api/v1/volunteers/`
```json
{ "first_name": "Pedro", "last_name": "Gómez", "availability": "Fines de semana" }
```
