# API y endpoints

Todo pasa por el API Gateway: nginx en local (`api-gateway/nginx.conf`,
puerto 8080) y Google Cloud API Gateway en la nube (`api-gateway/openapi.yaml`,
que declara método por método: **una ruta nueva hay que agregarla ahí**).
Cada servicio valida el JWT por su cuenta (cookie `access_token` o
`Authorization: Bearer`). Errores en JSON `{"error": "..."}` salvo
labor-integration (texto plano).

Leyenda de roles: A admin, V voluntario, P participante, E empresa,
D docente interno, DX docente externo, * cualquiera con sesión, — sin sesión.

## auth (`/api/v1/...`)

| Método y ruta | Roles |
|---|---|
| POST `auth/register`, `auth/login`, `auth/refresh`, `auth/logout`, `auth/forgot-password`, `auth/reset-password`, `auth/verify-2fa`, `auth/verify-email` | — (límite estricto por IP) |
| GET `auth/me`, GET/PUT `profile/me`, `profile/me/2fa/*`, `profile/me/link-google` | * |
| POST/GET/PUT/DELETE `users[/{id}]` | servicio decide: A todo; V solo participantes; resto, solo la propia cuenta |
| `roles`, `permissions` | A |
| `participants[/{id}][/documents]` | A, V, P (P solo lo propio; P no crea perfiles ajenos ni borra su perfil) |
| `companies`, `teachers`, `volunteers` | A + el propio rol sobre su perfil |
| `directory/participants/*`, `directory/users/summaries` | A, V (D solo `participants/{id}`) |

## education (puerto 8210)

Resumen; la tabla completa con reglas está en `../education/CLAUDE.md`.

| Método y ruta | Roles |
|---|---|
| GET `/api/salud`, GET/POST `/api/comentarios`, GET `/api/biblioteca`, POST `/api/biblioteca/{id}/descargas` | — |
| GET `/api/cursos[?q=&area=]`, `/api/cursos/{id}`, `/api/areas-formativas`, `/api/materiales` | * |
| POST/PUT/DELETE `/api/cursos`, publicar/despublicar, `/api/cursos/borradores`, áreas, DELETE materiales, `/api/auditoria`, moderación | A |
| POST/PUT `/api/materiales[/upload]` | A, D |
| inscripciones, actividades, entregas, asistencia, asistencia general, `/api/participantes`, **GET `/api/cursos/{id}/progreso`** (nuevo) | A, V |
| `/api/docente/*` | D |
| `/api/participante/*` | P |
| `/api/contenidos-docentes*` | D y DX (subir/editar lo propio), A (moderar) |
| `/api/actividades/sugerir-habilidades`, `/api/actividades/{id}/habilidades` | A, V, D |

## human-promotion (`/api/v1/human-promotion`)

| Ruta | Roles |
|---|---|
| `participantes`, `fichas`, `fichas/{id}`, `fichas/{id}/pdf`, `fichas/{id}/hitos`, `fichas/{id}/actividad` | A, V |
| `indicadores`, `auditoria` | A |
| `mi-linea-de-tiempo` | P |

## campaigns (`/api/v1/campaigns`)

| Ruta | Roles |
|---|---|
| GET `publicas[?q=]` | — |
| GET `/`, POST `/`, GET/PUT `/{id}`, POST `/{id}/publicar`, **DELETE `/{id}`** (baja lógica, nuevo) | A, V |

## labor-integration (`/api/v1/labor-integration`)

| Ruta | Roles |
|---|---|
| GET `offers[?q=&empresa=]` | * (pendientes solo A, V y empresa dueña) |
| POST `offers` | E, A, V |
| PUT/DELETE `offers/{id}` | A, V, empresa dueña (estado solo A) |
| PUT `offers/{id}/approve`, `offers/{id}/reject` | A |
| GET `offers/{id}/applications` | A, V, empresa dueña |
| `applications`, `applications/me`, `applications/{id}[/status]` | según rol en el servicio |

## Otros

| Servicio | Rutas |
|---|---|
| dashboard | GET `/api/v1/dashboard/resumen?desde&hasta&modulo&area`, reportes (A) |
| notifications | `/api/v1/notifications` (listado, detalle, `read`, `read-all`, `unread-count`, preferencias) del propio usuario; `/eventos/*` solo Pub/Sub |
| rxresume | `/api/v1/rxresume/*` (con sesión), `/api/v1/rxresume/public/*` |
| Pub/Sub (internos) | `/eventos/*`, `/internal/outbox/flush`, `/internal/delivery/flush`: no expuestos en el gateway |
