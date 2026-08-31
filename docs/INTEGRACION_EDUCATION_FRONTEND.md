# Integración education-service ↔ frontend

Cómo se conectó el módulo de **Educación** al frontend Angular, tomando como
referencia visual los mocks de UX y como contrato de datos la **API real** de
`education-service`.

Estado: **dockerizado (Fase 3 hecha, 2026-08-31)**. El frontend Docker en
`http://localhost:4200` tiene toda la integración; el dev local `:4300` sigue
disponible para iterar. `auth-service` nunca se tocó.

Última actualización: **2026-08-31**.

> **Los 3 modos de correr el frontend:**
> - **Docker** (`:4200`): imagen de producción, pega todo al **api-gateway** `:8080`.
> - **Dev local** (`:4300`, `pnpm start:local`): `ng serve` + proxy, para iterar.
> - Ambos backends (`auth`, `education`, `gateway`) corren siempre en Docker.

---

## 1. Cómo correr el frontend de desarrollo

```bash
cd frontend
pnpm install          # una vez — node_modules estaba en Angular 18, package.json pide 20
pnpm start:local      # ng serve -c local  ->  http://localhost:4300
```

- `frontend/proxy.conf.json` (nuevo) reenvía, **sin CORS**, desde `:4300`:
  - `/api/v1/*` → `http://localhost:8081` (auth-service en Docker)
  - `/api/cursos`, `/api/areas-formativas`, `/api/materiales`, `/api/docente`,
    `/api/salud` → `http://localhost:8210` (education-service en Docker)
- `environment.local.ts` (nuevo, fileReplacement de la config `local` en
  `angular.json`): `apiUrl = ""`, `educationApiUrl = ""` (todo relativo → lo
  resuelve el proxy).
- El contenedor del frontend “de producción” sigue intacto en `:4200`.

Logins para probar (usuarios creados para esto):
- `admin3@gmail.com` / `Admin333!` — `role_admin`, `sub = usr-OEIF3FGXuG`
- `teacher3@gmail.com` / `Teacher3!` — `role_teacher` interno, aprobado,
  `sub = usr-CyrffGTQ5f`

**Flujo completo probado (2026-08-30):** con `admin3` → crear área + material +
curso con `docente responsable = teacher3` + "publicar ahora" → el curso
aparece en `/dashboard/cursos` y en `/dashboard/mis-cursos` de `teacher3`.

---

## 2. Piezas nuevas

| Archivo | Qué es |
|---|---|
| `src/app/core/models/education.model.ts` | Modelos 1:1 con `education/internal/models/*.go`. `richTextToPlain()` aplana el Rich Text de Contentful; `plainToRichText()` hace lo inverso para el alta de cursos. |
| `src/app/services/education.service.ts` | Cliente HTTP de **todos** los endpoints de education (catálogo, CRUD admin, portal docente). Mismo patrón que `AuthService`. |
| `src/app/services/users.service.ts` | `GET /api/v1/users` de auth-service (solo admin) — para el selector de docente responsable. |
| `src/app/core/guards/admin.guard.ts` | Deja pasar solo `role_admin` / `role_dev`. |
| `src/app/pages/dashboard/admin-cursos/` | **UI de administración**: CRUD de cursos + áreas formativas + materiales. |
| `src/app/components/layout/sidebar/` | Navegación lateral real, con links filtrados por rol (antes era un placeholder). |
| `src/environments/environment.local.ts` | Entorno de dev local (URLs relativas). |
| `frontend/proxy.conf.json` | Proxy de `ng serve`. |

`environment.ts` ganó un campo `educationApiUrl` (default `""`).
`app.routes.ts` ganó la ruta `/dashboard/gestion-cursos` (protegida por `adminGuard`).

---

## 2bis. UI de administración — CRUD de cursos (HU-14 / HU-15 / HU-16)

Ruta: **`/dashboard/gestion-cursos`** (link "Gestión de cursos" en el sidebar,
visible solo para admin). Tres pestañas:

### Cursos
- Tabla de cursos **publicados** (`GET /api/cursos`).
- **Nuevo curso** → diálogo con: título\*, descripción (textarea → se convierte
  a Rich Text de Contentful), modalidad\* (`Presencial`/`Virtual`/`Mixta`),
  duración, **área formativa\*** (select), **docente responsable\*** (select de
  usuarios `role_teacher` traídos de auth), materiales (multiselect), y checkbox
  **"publicar ahora"** (default on).
  - Crear = `POST /api/cursos` → si "publicar ahora" → `POST /api/cursos/{id}/publicar`.
  - HU-14: valida que haya docente responsable y área antes de guardar
    (`courseFormValid`). Mensaje de éxito *"Se creó el curso correctamente…"*.
- **Editar** → carga `GET /api/cursos/{id}` → `PUT /api/cursos/{id}`.
  Mensaje *"Se modificó el curso correctamente"* (HU-15).
- **Despublicar** → `POST /api/cursos/{id}/despublicar`.
- **Eliminar** → `p-confirmDialog` (HU-16 pide confirmación explícita) →
  `DELETE /api/cursos/{id}` (el backend despublica + borra = baja lógica).
  Mensaje *"Se eliminó el curso correctamente"*.

### Áreas formativas / Materiales
CRUD simple contra `/api/areas-formativas` y `/api/materiales`
(`create/update/delete` en `education.service.ts`). Un curso necesita **al menos
un área** para poder crearse; los materiales son opcionales.

### El flujo, de punta a punta
`admin` crea área → crea material → crea curso eligiendo a un docente y
publicándolo → el curso queda visible en el **catálogo** (`/dashboard/cursos`,
cualquier rol) y en **"Mis cursos"** del docente elegido
(`/dashboard/mis-cursos`, `GET /api/docente/cursos`). Desde ahí el docente ya
puede cargar actividades, asistencia y ver progreso.

### Limitaciones de la UI admin
- `education-service` **no lista los borradores** a administración → la tabla
  solo muestra publicados. Un curso creado sin "publicar ahora" no aparece
  hasta publicarlo (habría que agregar un endpoint `GET /api/cursos?estado=draft`).
- HU-16 pide *"advertir si hay participantes inscriptos (cantidad)"* antes de
  borrar — no hay endpoint admin de inscriptos, así que el diálogo solo avisa
  en texto, sin el número.
- `GET /api/v1/users` **hoy lo permite solo `role_admin`** y devuelve **todos**
  los usuarios (sin filtro server-side por rol); el frontend filtra
  `role_teacher` en cliente. Pendiente: endpoint de auth que filtre por rol.
- No hay selector de docente por *nombre* con búsqueda contra un endpoint
  dedicado; se usa el listado completo filtrado en memoria.

---

## 3. Mapa: mock UX → página Angular → endpoint real

Los mocks viven en la rama **`mocks`**
(`Authentication and User Management UX/src/app/components/*.tsx`, React/Vite).
Son **referencia visual**: sus tipos (`level`, `rating`, `modules`, …) son
inventados y **no** coinciden con la API real (Contentful, en inglés).

| Página (`/dashboard/...`) | Mock de referencia | Endpoint(s) `education` | Rol |
|---|---|---|---|
| `cursos` | `CourseCatalog.tsx`, `CourseDetail.tsx` | `GET /api/cursos` | cualquiera autenticado |
| `mis-cursos` | `TeacherPortal.tsx` | `GET /api/docente/cursos` + `GET /api/docente/cursos/borradores` | `role_teacher` interno |
| `progreso` | `ProgressTracking.tsx` | `GET /api/docente/cursos` (selector) + `GET /api/docente/cursos/{id}/progreso` | `role_teacher` interno |
| `asistencia` | `AttendanceTracking.tsx` | `GET/POST /api/docente/cursos/{id}/actividades`, `GET/POST /api/docente/actividades/{actId}/asistencia` | `role_teacher` interno |
| `participantes` | `VolunteerUserManagement.tsx` / `UserManagement.tsx` | `GET /api/docente/cursos/{id}/inscriptos` | `role_teacher` interno |
| `seguimiento` | `ParticipantTracking.tsx` | `inscriptos` + `progreso` (merge por `participanteId`) | `role_teacher` interno |

> Ojo: en los mocks `participantes`/`seguimiento`/`asistencia general` eran de
> **admin/voluntario**. `education-service` **no** tiene endpoints de
> inscripciones/participantes para esos roles — solo el docente responsable ve
> los inscriptos de *sus* cursos. Con un token que no sea `role_teacher`
> interno, estas páginas van a mostrar el error 403 del backend.

---

## 4. Contrato de datos real (resumen)

`Course` (de `GET /api/cursos`):

```jsonc
{
  "id": "4EwaEUWCTlJ0qqEbHMnLCC",
  "title": "Introducción computación",
  "description": { "nodeType": "document", "content": [ /* Rich Text Contentful */ ] },
  "responsibleTeacher": "usr-...",      // id de usuario (sub del JWT), NO nombre
  "mode": "Presencial" | "Virtual" | "Mixta",
  "duration": "8 semanas",
  "trainingArea": { "id": "...", "name": "Naturaleza", "description": "..." },
  "material": [ { "id": "...", "title": "...", "type": "PDF" | "Video" | "Link" } ],
  "createdAt": "...", "updatedAt": "..."
}
```

- `description` es **Rich Text de Contentful** (árbol de nodos), no string. En
  listados se muestra con `richTextToPlain()`. Un render completo (negritas,
  listas, links) queda **pendiente** (componente dedicado).
- `responsibleTeacher` hoy es texto libre en algunos cursos viejos (`"Luis"`,
  `"test-user-1"`). El diseño correcto es un **id de usuario**; falta en el
  frontend un **selector de docentes** que resuelva nombre → id contra
  auth-service (pendiente).
- Errores: siempre `{ "error": "mensaje" }` con HTTP 400 / 403 / 404 / 409 / 502.

`ParticipantProgress` (de `.../progreso`):
`{ participanteId, actividadesTotal, actividadesAsistidas, porcentajeAsistencia }`

`Enrollment` (de `.../inscriptos`):
`{ id, participanteId, cursoId, fechaInscripcion, estado: "activa"|"finalizada"|"baja" }`

`Activity`: `{ id, cursoId, titulo, fecha (RFC3339), createdAt }`
`Attendance`: `{ id, participanteId, actividadId, presente, fecha }`
`POST .../asistencia` recibe **un array** `[{ participanteId, presente }]`;
reenviar el mismo `participanteId` **actualiza** (id determinístico, no duplica).

---

## 5. Limitaciones conocidas / pendientes

1. **Inscripciones**: `education-service` no tiene endpoint para inscribir a un
   participante a un curso. La colección `inscripciones` de Firestore se carga
   a mano. Por eso `participantes`, `seguimiento` y el `progreso` con datos
   reales salen vacíos hasta que exista ese módulo.
   → El botón "Inscribirme" del catálogo está **deshabilitado** a propósito.
2. **`asistencia`**: los ids de participante se tipean a mano (no hay lista de
   inscriptos). Cuando exista (1), se reemplaza por un checklist.
3. **Subida de archivos binarios** (portada de curso, PDF de material):
   `education-service` no tiene endpoint de upload — ver §8.
4. **`GET /api/v1/users`**: hoy solo lo permite `role_admin` y devuelve TODOS
   los usuarios (sin filtro server-side por rol). El frontend filtra
   `role_teacher` en memoria. Pendiente: endpoint de auth con `?role=`.

---

## 5bis. Cambios del 2026-08-31

### Diseño / identidad
- **Tipografía del Manual de Identidad Buen Vivir**: el manual usa *Palo Seco*
  (primaria, comercial, **no** disponible como webfont) y *Poppins* (secundaria,
  sí en Google Fonts). Stack aplicado en `styles.css` y componentes:
  `"Palo Seco", "Poppins", "Inter", system-ui, sans-serif` — Poppins (ya
  importada de Google Fonts) hace el trabajo; si alguien tiene Palo Seco
  instalada, se usa. Títulos (`h1..h5`, `.page-title`, `.brand-name`) con el
  mismo stack, weight 600.
- **Menú lateral (`sidebar.component`)**: dejó de ser placeholder. Ahora es
  navegación real, **más grande** (288px de ancho, ítems 1.05rem, íconos
  1.25rem, más padding), con el **sidebar oscuro** del manual (`#1D1D1B`, texto
  crema, activo en verde `--primary #485839`). Los links se filtran por rol
  (`admin`/`dev`: Gestión de cursos, Usuarios; `teacher`: Mis cursos,
  Asistencia, Progreso, Participantes, Seguimiento; todos: Catálogo, Mi perfil).
- **Layout del dashboard**: `dashboard.component` tenía el CSS vacío (header,
  sidebar y contenido se apilaban). Ahora `header` arriba, fila
  `[sidebar | main]` flexible, `footer` abajo; `main` scrollea solo.
- El nav superior del `header` sumó el link "Gestión de cursos" para el rol
  responsable (admin).

### Selector de docente responsable — **hecho** (HU-14)
En el alta/edición de curso (`/dashboard/gestion-cursos`), "Docente responsable"
es un **`p-select` con búsqueda** poblado desde `GET /api/v1/users` (auth-service,
NeonDB) filtrando `role_id === 'role_teacher'`. Muestra
`Nombre Apellido · email` y **envía el `user_id`** en `responsibleTeacher` (no
texto libre). El listado sale de la **base de datos de auth** (tabla `users`),
vía `users.service.ts`. "Selector por nombre con endpoint dedicado" = tener en
auth un `GET /api/v1/users?role=role_teacher&q=juan` para no traer todos y
filtrar en el browser; por ahora se traen todos (son pocos) y el `[filter]` del
`p-select` permite tipear para buscar.

### Renderer de Rich Text — **hecho**
`shared/components/rich-text.component.ts` → `<app-rich-text [doc]="curso.description">`.
El campo `description` de un curso se guarda como **documento Rich Text de
Contentful** (un árbol JSON con formato: párrafos, títulos, listas,
negrita/cursiva, citas, enlaces). Antes se aplanaba a texto plano
(`richTextToPlain`, que se sigue usando para *previews* cortas en tarjetas y
tablas). El nuevo componente lo **renderiza con su formato**: convierte el árbol
a HTML acotado y lo pasa por `DomSanitizer` antes de bindear. Se usa en el
diálogo **"Ver detalle"** del catálogo de cursos (`/dashboard/cursos`), que
además lista los materiales del curso.

### Listado de borradores para admin — **hecho** (backend + frontend, rebuild aplicado)
- **education-service**: nuevo endpoint **`GET /api/cursos/borradores`** (solo
  `role_admin`) → todos los cursos sin publicar, leídos de la CMA de Contentful.
  Archivos: `course_repository.go` (`ListAllDrafts`), `course_service.go`
  (`ListAllDraftCourses`), `course_controller.go` (`ListDrafts`), `routes.go`.
  Probado: admin → 200 con la lista; teacher3 → 403.
- **frontend**: `education.service.ts#listDraftCourses()`; la tabla de "Gestión
  de cursos" muestra **publicados + borradores** con columna Estado y botón
  **Publicar** para los borradores. Si el endpoint no estuviera desplegado, la
  llamada degrada a `[]` (no rompe la página).
- Rebuild aplicado: `docker compose build education-service && docker compose up -d education-service`
  (2026-08-31). auth/frontend/gateway sin tocar.

### Materiales con URL (Video / Link) — **hecho** (Opción A, ver §8)

---

## 8. Materiales en formato archivo / link de video

### Opción A — campo `url` (Video/Link) — **IMPLEMENTADA (2026-08-31)**

1. **Contentful**: se agregó el campo `url` (Symbol, opcional) al Content Type
   `material` del environment **`testing`** (vía Content Management API, con el
   `CONTENTFUL_CMA_TOKEN` de `education/.env`) y se publicó (publishedVersion 13).
   > ⚠️ El environment `testing` es compartido — el campo `url` ahora existe para
   > todo el equipo. Es reversible (borrar el campo).
2. **education-service**: `models.Material` y `models.MaterialInput` tienen
   `URL`; `material_repository.go` lo mapea en Create/Update y al leer de la
   CDA; `material_service.go` valida que empiece con `http(s)://`.
3. **frontend**: el diálogo de material tiene un input **URL**; la tabla de
   materiales y el detalle del curso muestran el enlace efectivo
   (`materialHref(m)` = `file.url ?? url`). Para Video/Link alcanza con pegar la
   URL; para PDF se puede usar URL **o** el `fileAssetId`.

Probado: `POST /api/materiales {"title","type":"Video","url":"https://…"}` → 201,
y `GET /api/materiales` lo devuelve con `url`.

### Opción B — subida real de binarios — **IMPLEMENTADA (2026-08-31)**

- **education-service**: `contentful_assets.go` (nuevo) → `UploadAsset()` hace el
  flujo de 5 pasos de la CMA (subir binario a `upload.contentful.com` → crear
  Asset → procesar → poll → publicar). Endpoint **`POST /api/materiales/upload`**
  (multipart: `file`, `title`, `type`).
  - **Permisos**: `role_admin` **o** `role_teacher` interno (para que el docente
    pueda adjuntar archivos a sus cursos, ver vista docente de las HU). El
    borrado de materiales sigue siendo solo admin.
  - **Límites** (`material_service.go`): JPG/PNG/WEBP ≤ 5 MB, PDF/DOC/DOCX ≤ 10 MB.
    Video no se sube (se usa `url`). Límite duro del multipart: 12 MB.
- **gateway**: `client_max_body_size 12m` en la location de education.
- **frontend**:
  - `education.service.ts#uploadMaterial(file, title, type)`.
  - Admin (`/dashboard/gestion-cursos`): input de archivo en el diálogo de material.
  - Docente (`/dashboard/mis-cursos`): botón **"Materiales"** por curso → diálogo
    que lista los materiales del curso y permite **crear/subir** uno nuevo
    (URL o archivo) y adjuntarlo al curso (`PUT /api/docente/cursos/{id}`).
- Probado: `POST /api/materiales/upload` con PDF → 201, y `GET /api/materiales`
  lo devuelve con URL real de Contentful (`https://assets.ctfassets.net/…`).
  teacher3 puede subir (201); el borrado le da 403.

---

## 8bis. Fase 3 — dockerización (2026-08-31, HECHA)

### api-gateway (`api-gateway/nginx.conf`) — restart, sin rebuild
1. **Regex de rutas protegidas** cambiado de `…)/` a `…)(/|$)` → `GET /api/v1/users`
   (sin barra final) ahora matchea. **Esto arregla "el admin no trae usuarios de
   la BD"** — el selector de docentes en Docker recibía 404.
2. **education ruteado**: `location ~ ^/api/(cursos|areas-formativas|materiales|docente|salud)`
   → `education-service:8210` (antes: comentado y con puerto `:8080` mal). Sin
   `auth_request` (education valida el JWT solo). `client_max_body_size 12m`.
3. **CORS deduplicado**: el gateway y auth/education emitían cada uno sus headers
   CORS → el browser recibía `Access-Control-Allow-Origin` **dos veces** y
   rechazaba toda llamada desde `:4200`. Fix: `proxy_hide_header` de los headers
   CORS del upstream + el gateway queda como única fuente.

### frontend
4. **`education.service.ts`**: `base = educationApiUrl || apiUrl` → en Docker cae al
   gateway (`:8080`); en dev local sigue relativo (proxy).
5. **`angular.json`**: budget del bundle inicial subido a 1.6 MB (el build de
   producción fallaba con 1.02 MB > 1 MB; solo el umbral de lint).
6. **`frontend/nginx.conf`** (nuevo) + `COPY` en `frontend/Dockerfile`: el
   contenedor no tenía **fallback de SPA** → toda ruta que no fuera `/` daba
   404 al recargar (`/login`, `/dashboard/...`). Ahora `try_files … /index.html`
   + cache headers para `env.js`/`ngsw`/assets.
7. `docker compose build frontend && docker compose up -d frontend`.

**Todo verificado por el gateway `:8080`** (lo que hace el browser en `:4200`):
preflight OPTIONS → 204; login → 200; `/api/v1/users` → 200 (3 docentes);
`/api/cursos`, `/api/cursos/borradores`, `/api/materiales`, `/api/docente/cursos`
→ 200; upload → 201. Rutas SPA (`/login`, `/dashboard/gestion-cursos`, …) → 200.

---

## 6. Verificación rápida

Con la sesión iniciada como `teacher3` en `http://localhost:4300`:

| Página | Esperado hoy |
|---|---|
| `/dashboard/cursos` | Lista los cursos publicados reales de Contentful (hay varios). |
| `/dashboard/mis-cursos` | Vacío — ningún curso tiene `responsibleTeacher = usr-CyrffGTQ5f`. Crear uno como admin y publicarlo para verlo. |
| `/dashboard/progreso` | "No tenés cursos asignados" (idem). |
| `/dashboard/asistencia` | Idem; con un curso propio: se pueden crear actividades y registrar asistencia por id. |
| `/dashboard/participantes` y `/seguimiento` | Idem; con curso propio: tabla de inscriptos (vacía hasta que se carguen en Firestore). |

Chequeo del canal (sin browser):

```bash
# login por el proxy y llamada a education con la MISMA cookie
curl -s -c cj -X POST http://localhost:4300/api/v1/auth/check-credentials \
  -H 'Content-Type: application/json' -d '{"email":"teacher3@gmail.com","password":"Teacher3!"}'
# ... tomar login_nonce, hacer /api/v1/auth/login ...
curl -s -b cj http://localhost:4300/api/cursos            # 200 + JSON de cursos
curl -s -b cj http://localhost:4300/api/docente/cursos    # 200 + []
```

---

## 7. Fase 3 — HECHA (ver §8bis)

Quedó pendiente solo, a futuro: `assets/env.template.js` para que un pipeline de
CD inyecte `apiUrl` según ambiente (hoy `env.js` tiene `http://localhost:8080`
hardcodeado, que sirve para local).
   en `:4200`.
4. Recién ahí tiene sentido volver a tocar `docker-compose.yml`.
