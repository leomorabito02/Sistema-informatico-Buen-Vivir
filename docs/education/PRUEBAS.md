# Guía de pruebas — migración a chi + JWT/RBAC + docente + Firestore + gateway

Cómo probar, tarea por tarea, todo lo agregado el 2026-08-26 a
`education-service`. Complementa a `POSTMAN_EXAMPLES.md` (ejemplos de
body por endpoint) y a `CLAUDE.md` (referencia completa del módulo).

## 0. Bloqueos conocidos, antes de arrancar

1. **`auth-service` no levanta tal cual está el `.env` real**: su
   código lee la variable `DATABASE_URL`, pero el `.env` real tiene
   `DATABASE_URL_AUTH` (nombre distinto). Sin resolver esto, `auth-service`
   y por lo tanto `api-gateway` no van a poder probarse — solo queda
   disponible probar `education-service` suelto (tareas 1 a 6, sin la 7).
2. **Contentful, Environment `testing`**: si los `GET` de curso/área
   devuelven 404, lo más probable es que el Environment `testing`
   todavía no exista en el Space (por default Contentful solo trae
   `master`). Se soluciona en Contentful → Settings → Environments →
   Add Environment. Ver también `POSTMAN_EXAMPLES.md` → Troubleshooting.

## 1. Setup común (una sola vez)

```bash
cd education
docker compose up --build -d
curl http://localhost:8210/api/salud
```
Esperado: `200` con JSON `{"status":"UP","service":"education-service","appEnv":"testing","contentfulEnvironment":"testing"}`.

Generador de tokens de prueba (no depende de que `auth-service` esté
levantado). Guardá el `JWT_SECRET` en una variable de entorno primero,
no lo pegues suelto en la terminal:

```bash
export JWT_SECRET='<el mismo valor que pusiste en education/.env>'

# uso: make_jwt <sub> <role> <is_external:true|false>
make_jwt() {
  node -e "
    const crypto = require('crypto');
    function b64url(b){return Buffer.from(b).toString('base64').replace(/\+/g,'-').replace(/\//g,'_').replace(/=+\$/,'');}
    const h = b64url(JSON.stringify({alg:'HS256',typ:'JWT'}));
    const p = b64url(JSON.stringify({
      sub: '$1', role: '$2', is_external: $3,
      type: 'access', iat: Math.floor(Date.now()/1000), exp: Math.floor(Date.now()/1000)+3600
    }));
    console.log(h+'.'+p+'.'+b64url(crypto.createHmac('sha256', process.env.JWT_SECRET).update(h+'.'+p).digest()));
  "
}

ADMIN=$(make_jwt admin-1 role_admin false)
PARTICIPANT=$(make_jwt part-1 role_participant false)
TEACHER=$(make_jwt test-teacher-1 role_teacher false)
TEACHER_EXT=$(make_jwt test-teacher-2 role_teacher true)
```

---

## TAREA 1 — Migración al router chi

```bash
curl -i http://localhost:8210/api/cursos/algun-id-que-no-existe
```
Esperado: la ruta con `{id}` resuelve al controller (respuesta JSON de
tu código, no un 404 genérico de `net/http`). Si el contenedor arrancó
sin errores en el Setup, chi ya está ruteando correctamente.

## TAREA 2 — Middlewares JWT + RBAC calcados de auth

```bash
curl -i http://localhost:8210/api/cursos                                   # sin token
curl -i -H "Authorization: Bearer token-invalido" http://localhost:8210/api/cursos
```
Esperado: ambos `401` (`Unauthorized: missing token` / `Invalid token`).

Para probar el fail-fast de `JWT_SECRET`: en `education/.env` poné
`JWT_SECRET=` (vacío) o `JWT_SECRET=super-secret-access-key`, `docker
compose up --build`, y confirmá que el contenedor loguea `FATAL:
JWT_SECRET is not set or uses an insecure default...` y no arranca.
Restaurá el valor real después.

## TAREA 3 — RBAC en las rutas existentes

```bash
curl -i -H "Authorization: Bearer $PARTICIPANT" http://localhost:8210/api/cursos
# esperado: NO 401/403 (puede dar 404/502 según el estado real de Contentful, eso está OK)

curl -i -X POST -H "Authorization: Bearer $PARTICIPANT" http://localhost:8210/api/cursos
# esperado: 403 "Forbidden: insufficient role permissions"

curl -i -X POST -H "Authorization: Bearer $ADMIN" -H "Content-Type: application/json" -d '{}' http://localhost:8210/api/cursos
# esperado: 400 (validación de body), NO 403 — confirma que el admin pasa el RBAC
```

## TAREA 4 — `responsibleTeacher` como ID de usuario

1. Con `$ADMIN`, creá un `trainingArea` y un `material` (bodies en
   `POSTMAN_EXAMPLES.md`), anotá sus `id`.
2. Creá un `course` con `"responsibleTeacher": "test-teacher-1"` (el
   mismo `sub` de `$TEACHER`) y los ids de arriba.
3. `POST /api/cursos/{id}/publicar` con `$ADMIN`.
4. `curl -H "Authorization: Bearer $TEACHER" http://localhost:8210/api/docente/cursos`
   → tiene que aparecer ese curso.
5. `PUT /api/docente/cursos/{id}` con `$TEACHER`, mandando en el body
   `"responsibleTeacher": "otro-id-cualquiera"` — después hacé un `GET`
   y confirmá que **no cambió** (sigue siendo `test-teacher-1`).
6. `curl -H "Authorization: Bearer $TEACHER_EXT" http://localhost:8210/api/docente/cursos`
   → `403` (docente externo).

## TAREA 5 — Endpoints `/api/docente`

Usando el `cursoId` publicado en la tarea anterior y `$TEACHER`:

```bash
curl -X POST -H "Authorization: Bearer $TEACHER" -H "Content-Type: application/json" \
  -d '{"titulo":"Práctica 1","fecha":"2026-09-01T14:00:00Z"}' \
  http://localhost:8210/api/docente/cursos/{cursoId}/actividades
# copiá el "id" de la actividad devuelta

curl -H "Authorization: Bearer $TEACHER" \
  http://localhost:8210/api/docente/cursos/{cursoId}/actividades
# esperado: la actividad recién creada

curl -X POST -H "Authorization: Bearer $TEACHER" -H "Content-Type: application/json" \
  -d '[{"participanteId":"p1","presente":true},{"participanteId":"p2","presente":false}]' \
  http://localhost:8210/api/docente/actividades/{actividadId}/asistencia

curl -H "Authorization: Bearer $TEACHER" \
  http://localhost:8210/api/docente/actividades/{actividadId}/asistencia
# esperado: los 2 registros de asistencia

# repetir el POST de asistencia con el mismo participanteId p1 y otro valor de "presente":
# el GET tiene que seguir mostrando UN solo registro para p1 (se actualiza, no duplica)
```

Con `$TEACHER_EXT`, repetir cualquiera de estos `curl` → esperado `403`
en todos.

## TAREA 6 — Firestore

Las pruebas de la TAREA 5 ya ejercitan Firestore (si no da `502`,
escribió/leyó bien). Para confirmarlo visualmente: Firebase Console →
Firestore Database → tienen que aparecer:
- `actividades` con el documento creado arriba,
- `asistencias` con 2 documentos cuyos ids son
  `{actividadId}_p1` y `{actividadId}_p2` (no autogenerados).

`GET /api/docente/cursos/{cursoId}/inscriptos` y `.../progreso` van a
devolver vacío `[]` hasta que cargues manualmente un documento en la
colección `inscripciones` (campos `participante_id`, `curso_id`,
`fecha_inscripcion`, `estado`) — no hay endpoint que la pueble todavía
(documentado como fuera de alcance).

## TAREA 7 — nginx + docker-compose de la raíz

Requiere resolver el bloqueo de `DATABASE_URL`/`DATABASE_URL_AUTH` del
punto 0.

```bash
cd ..
docker compose up --build -d
docker compose ps                              # los 3 servicios "Up"
curl -i http://localhost:8080/health           # 200 OK, del propio nginx
curl -i http://localhost:8080/api/cursos       # sin token -> 401, respondido por education-service vía el gateway
curl -i http://localhost:8080/api/v1/auth/login -X POST   # confirma que el bloque de auth-service sigue andando
```

## Documentación

No se "ejecuta", pero conviene revisar que `CLAUDE.md` y
`POSTMAN_EXAMPLES.md` describen exactamente lo que probaste arriba —
son la referencia para no tener que repetir esta guía la próxima vez.

---

## Checklist rápido

- [ ] `go build ./...` / `go vet ./...` sin errores
- [ ] `/api/salud` responde 200 sin token
- [ ] Cualquier otra ruta sin token → 401
- [ ] `role_participant` puede leer curso/área/material, no puede escribir (403)
- [ ] `role_admin` puede crear/publicar un curso completo
- [ ] `responsibleTeacher` no se puede reasignar vía `PUT /api/docente/cursos/{id}`
- [ ] Docente externo (`is_external:true`) → 403 en todo `/api/docente/*`
- [ ] Crear actividad + registrar asistencia + releer sin duplicados
- [ ] Colecciones `actividades`/`asistencias` visibles en Firestore Console
- [ ] (si se resolvió el bloqueo de auth) los 3 servicios levantan juntos y el gateway proxea bien
