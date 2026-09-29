# Pruebas: auth ↔ frontend ↔ education (JWT)

Cómo probar, de punta a punta:

1. que el **frontend** se loguee contra **auth-service** (cookie de sesión), y
2. que el **JWT que emite auth** sea aceptado por **education-service** (para
   probar los endpoints en Postman / curl).

> Complementa a `education/POSTMAN_EXAMPLES.md` (bodies de cada endpoint de
> education) y a `education/PRUEBAS.md` (pruebas internas de education).

Última verificación de esta guía: **2026-08-30** (flujo completo probado OK).

---

## 0. Panorama — qué corre dónde

| Pieza | Cómo corre | URL |
|---|---|---|
| `api-gateway` (nginx) | Docker | `http://localhost:8080` |
| `auth-service` | Docker | `http://localhost:8081` (directo) / vía gateway en `:8080` |
| `education-service` | Docker | `http://localhost:8210` |
| `frontend` (imagen prod) | Docker | `http://localhost:4200` |
| `frontend` (dev, para integrar education) | `pnpm start:local` | `http://localhost:4300` |

**Nada de lo dockerizado se toca.** El frontend de desarrollo (`:4300`) es
un proceso aparte con `ng serve` + proxy (ver
`docs/INTEGRACION_EDUCATION_FRONTEND.md`). Para esta guía de pruebas alcanza
con lo que ya está en Docker + Postman.

Chequeo rápido de que todo está arriba:

```bash
docker compose ps                                  # 4 servicios "Up"
curl -s http://localhost:8080/health               # OK  (nginx)
curl -s http://localhost:8081/health               # {"status":"UP","service":"auth-service"}
curl -s http://localhost:8210/api/salud            # {"status":"UP","service":"education-service",...}
```

---

## 1. Usuarios de prueba (NeonDB de auth)

Ya existen usuarios por rol. **`teacher3@gmail.com` fue creado para estas
pruebas** y su contraseña es conocida:

| Email | Contraseña | Rol | `user_id` (= `sub` del JWT) | Notas |
|---|---|---|---|---|
| `teacher3@gmail.com` | `Teacher3!` | `role_teacher` | `usr-CyrffGTQ5f` | `is_external=false`, approval **approved** → sirve para `/api/docente/*` |
| `admin3@gmail.com` | `Admin333!` | `role_admin` | `usr-OEIF3FGXuG` | **creado para estas pruebas** — CRUD de cursos/áreas/materiales, listar usuarios |
| `admin@gmail.com` | *(preguntar al equipo)* | `role_admin` | `usr-bs4fdRDDBZ` | escribe cursos/áreas/materiales |
| `teacher@gmail.com` | *(preguntar)* | `role_teacher` | `usr-o43XN6f6gh` | interno, approval approved |
| `participant@gmail.com` | *(preguntar)* | `role_participant` | `usr-anGBnlpTJ1` | solo lectura en education |
| `company@gmail.com` | *(preguntar)* | `role_company` | `usr-a1Vrex13CT` | `status=bloqued` → sirve para probar "cuenta bloqueada" |

Si necesitás otro usuario con contraseña conocida, registralo (ver §2.A) —
**no modifiques los usuarios existentes**.

Roles válidos en la tabla `roles`: `role_admin`, `role_dev`, `role_company`,
`role_participant`, `role_teacher`, `role_volunteer`.
`role_admin` / `role_dev` tienen **bypass** de `RequireRole` en education.

---

## 2. Fase 1a — Login del frontend contra auth (browser)

El frontend hace login en **2 pasos** (no manda user+pass y listo):

1. `POST /api/v1/auth/check-credentials` `{email, password}` → valida y
   devuelve la lista de cuentas del email + un `login_nonce` (vale 2 min).
2. `POST /api/v1/auth/login` `{user_id, role_id, login_nonce}` → emite el
   JWT, lo setea como **cookie `access_token`** (`HttpOnly; Secure;
   SameSite=Strict`) y además lo devuelve en el body (`token`).

### Opción A — registrar un usuario nuevo y entrar

En `http://localhost:4200` (frontend Docker) → **Registrarse**:

- Nombre / Apellido (mín. 2 caracteres)
- Email nuevo
- Teléfono (cualquier valor no vacío)
- Contraseña: **mín. 8, con mayúscula, minúscula y carácter especial**
  (ej. `Prueba123!`)
- Rol

`register` deja la sesión iniciada al toque (setea las cookies y devuelve
token). Los perfiles de `teacher` / `participant` / `company` / `volunteer`
quedan en `approval_status = "pending"` → el `authGuard` te manda a
`/pending-approval` hasta que un admin apruebe. Para `role_participant` con
approval pendiente igual podés navegar; para probar el dashboard completo
usá un usuario `approved` (ej. `teacher3@gmail.com`) o aprobalo.

### Opción B — entrar con `teacher3@gmail.com`

1. `http://localhost:4200` → **Iniciar sesión**
2. Email `teacher3@gmail.com`, contraseña `Teacher3!`
3. Si el email tuviera varias cuentas (varios roles), aparece la pantalla
   **elegir rol** (`/select-role`). `teacher3` tiene una sola.
4. Entra al dashboard.

### Qué verificar en el browser (DevTools)

- **Application → Cookies → `http://localhost:4200`**: tiene que aparecer
  `access_token` (y `refresh_token`), `HttpOnly`, `Secure`, `SameSite=Strict`.
  > `Secure` sobre `http://localhost` funciona: los navegadores tratan a
  > `localhost` como contexto seguro.
- **Network**: `check-credentials` → 200, `login` → 200, y después
  `GET /api/v1/auth/me` → 200 con `{user, approval_status, ...}`.
  Si `/auth/me` da 401, la cookie no se está mandando (ver §5).
- **Logout**: `POST /api/v1/auth/logout` → 200 y las cookies se borran
  (`Max-Age=-1`).

### curl equivalente (sin browser)

```bash
# paso 1
CC=$(curl -s -c cookies.txt -X POST http://localhost:8080/api/v1/auth/check-credentials \
  -H "Content-Type: application/json" \
  -d '{"email":"teacher3@gmail.com","password":"Teacher3!"}')
echo "$CC"
NONCE=$(echo "$CC" | grep -o '"login_nonce":"[^"]*"' | cut -d'"' -f4)

# paso 2  (user_id y role_id salen de la respuesta anterior)
curl -s -c cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"usr-CyrffGTQ5f\",\"role_id\":\"role_teacher\",\"login_nonce\":\"$NONCE\"}"

# sesión viva
curl -s -b cookies.txt http://localhost:8080/api/v1/auth/me
```

---

## 3. Fase 1b — El JWT de auth contra education (Postman / curl)

`education-service` **no habla con auth**: valida el JWT por su cuenta, con
el **mismo `JWT_SECRET`** (está igual en `auth/.env` y `education/.env`) y
algoritmo HS256. Acepta el token en:

- header `Authorization: Bearer <token>`, **o**
- cookie `access_token` (si Postman ya la tiene de haber pegado contra auth).

Claims que mira education: `sub` (→ id de usuario), `role`, `is_external`,
`type` (tiene que ser `"access"`). TTL del access token: **15 min**.

### 3.1 Obtener un token real

Del body de `POST /api/v1/auth/login` (o de `register`), campo `token`.
Ejemplo con `teacher3`:

```bash
TK=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"usr-CyrffGTQ5f\",\"role_id\":\"role_teacher\",\"login_nonce\":\"$NONCE\"}" \
  | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "$TK"
```

En **Postman**: request a `POST http://localhost:8080/api/v1/auth/login`,
después en el request a education → pestaña **Authorization** → tipo
**Bearer Token** → pegar `{{token}}` (o guardarlo en una variable de entorno
de Postman desde un test script:
`pm.environment.set("token", pm.response.json().token)`).

### 3.2 Probar education con ese token

```bash
# sin token -> 401
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8210/api/cursos          # 401

# con token de teacher3 -> 200 (cualquier rol autenticado puede leer)
curl -s -w "\n%{http_code}\n" http://localhost:8210/api/cursos \
  -H "Authorization: Bearer $TK"                                                    # 200

# endpoint de docente interno -> 200 (rol role_teacher + is_external=false)
curl -s -w "\n%{http_code}\n" http://localhost:8210/api/docente/cursos \
  -H "Authorization: Bearer $TK"                                                    # 200  (lista vacía si no tiene cursos)

# escritura -> 403 (teacher no es admin)
curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8210/api/cursos \
  -H "Authorization: Bearer $TK" -H "Content-Type: application/json" -d '{}'        # 403
```

Resultado esperado (verificado 2026-08-30): `401 / 200 / 200 / 403`.

### 3.3 Matriz RBAC de education

| Endpoint | Requiere |
|---|---|
| `GET /api/salud` | nada (público) |
| `GET /api/cursos`, `/api/cursos/{id}`, `/api/areas-formativas`, `/api/materiales` | cualquier JWT válido |
| `POST/PUT/DELETE` de curso/área/material, `publicar`/`despublicar` | `role_admin` (o `role_dev`) |
| `/api/docente/*` | `role_teacher` **y** `is_external=false` |

Un docente externo (`is_external=true`) recibe **403** en todo `/api/docente/*`.
Hoy no hay ningún usuario `is_external=true` en la DB — para probar ese caso
hay que registrar uno y marcarlo, o forjar un JWT de prueba (ver
`education/POSTMAN_EXAMPLES.md`, generador con Node).

### 3.4 Flujo completo de education (con token de admin + token de teacher)

Ver `education/POSTMAN_EXAMPLES.md`. Resumen del orden:

1. `POST /api/areas-formativas` (admin) → guardar `id`
2. `POST /api/materiales` (admin) → guardar `id`
3. `POST /api/cursos` (admin) con `trainingAreaId`, `materialIds`, y
   **`responsibleTeacher: "usr-CyrffGTQ5f"`** (el `sub` de `teacher3`)
4. `POST /api/cursos/{id}/publicar` (admin)
5. `GET /api/cursos` (cualquiera) → el curso aparece resuelto
6. `GET /api/docente/cursos` (token de `teacher3`) → aparece ese curso
7. `POST /api/docente/cursos/{id}/actividades`, `.../asistencia`, etc.

---

## 4. Checklist

- [ ] `docker compose ps` → 4 servicios Up
- [ ] `/health` (8080 y 8081) y `/api/salud` (8210) responden
- [ ] Browser: login con `teacher3@gmail.com` / `Teacher3!` deja cookie `access_token`
- [ ] Browser: `GET /api/v1/auth/me` → 200
- [ ] `POST /api/v1/auth/login` devuelve `token` en el body
- [ ] education `GET /api/cursos` sin token → 401
- [ ] education `GET /api/cursos` con el token → 200
- [ ] education `GET /api/docente/cursos` con el token de teacher3 → 200
- [ ] education `POST /api/cursos` con token de teacher3 → 403
- [ ] education `POST /api/cursos` con token de admin → 400 (valida body), no 403

---

## 5. Troubleshooting

### `GET /api/v1/auth/me` → 401 desde el browser
La cookie `access_token` no viaja. Causas típicas:
- El request no lleva `withCredentials: true` (el interceptor
  `auth.interceptor.ts` ya lo fuerza; revisá que la request pase por
  `HttpClient` y no `fetch` crudo).
- Cross-site real (frontend y auth en dominios distintos). En local
  `localhost:4200` ↔ `localhost:8080` es **same-site** (el puerto no cuenta),
  así que `SameSite=Strict` no bloquea.
- La cookie expiró (15 min). Volvé a loguear.

### education responde 401 con un token que parece válido
- El token venció (15 min). Pedí uno nuevo.
- `JWT_SECRET` distinto entre `auth/.env` y `education/.env`. Tienen que ser
  **idénticos**. Verificar:
  ```bash
  grep JWT_SECRET auth/.env education/.env
  ```
- El claim `type` no es `"access"` (pasa si pegaste el `refresh_token` por
  error — ese no es un JWT).

### education responde 403 en `/api/docente/*` con un token de teacher
- El rol del token no es `role_teacher` (mirá el claim `role`).
- `is_external=true` en el token → docente externo, bloqueado a propósito.
- Para `GET`/`PUT` de un curso puntual: el curso está en borrador (no
  publicado) → da **404**, no 403. Publicalo primero con un admin.

### `POST` a education (crear curso/área) → `401 OrganizationAccessGrantRequired`
No es auth: es que el token de **Contentful** no tiene acceso a la
organización del Space. Ver `education/POSTMAN_EXAMPLES.md` § Troubleshooting.

### CORS al pegar desde el browser (no desde Postman)
`auth` y `education` tienen `CORS_ALLOWED_ORIGINS=http://localhost:4200`.
Desde el frontend Docker (`:4200`) no hay problema. Desde el frontend de
desarrollo (`:4300`) **no se pega directo**: se usa el proxy de `ng serve`
(mismo origen, sin CORS) — ver `docs/INTEGRACION_EDUCATION_FRONTEND.md`.
