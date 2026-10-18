# Entorno de calidad (TP de Calidad de Software)

Una copia de Educación para probar sin tocar nada real (decisión D-22).
Se define en `docker-compose.calidad.yml`, en la raíz del repo.

| Pieza | Qué usa | Puerto en la PC |
|---|---|---|
| Educación | El mismo `education/Dockerfile` | `127.0.0.1:18210` |
| Firestore | Emulador, vacío en cada arranque | `127.0.0.1:18390` |
| Contentful | El real, environment `testing` (solo lectura desde acá) | — |
| Storage, Pub/Sub, IA, auth-service | Apagados | — |

No comparte red, puertos ni volúmenes con `docker-compose.yml`: los dos
stacks pueden estar prendidos a la vez.

## Cómo se usa

Todo desde la raíz del repo, en Git Bash.

```bash
# 1. Una sola vez: arma .env.calidad (Contentful testing + JWT propio)
bash scripts/calidad/entorno/preparar_env.sh

# 2. Prender el entorno (la primera vez compila la imagen)
docker compose -f docker-compose.calidad.yml up -d --build

# 3. Cargar los datos sintéticos (cuando /api/salud responda 200)
curl http://127.0.0.1:18210/api/salud
python scripts/calidad/entorno/sembrar.py

# 4. Pedir algo con un usuario de prueba
TOKEN=$(python scripts/calidad/entorno/token_prueba.py role_admin)
curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1:18210/api/cursos

# 5. Apagar y borrar (los datos se pierden: es a propósito)
docker compose -f docker-compose.calidad.yml down
```

## Datos sintéticos

`sembrar.py` toma 3 cursos publicados de Contentful `testing` y les crea:
2 actividades por curso, 8 participantes inventados
(`participante-sint-01` a `08`) con estados mezclados (activa,
finalizada, baja), 6 habilidades del catálogo y algunas habilidades
ganadas. Ningún dato es de una persona real. Se puede correr varias veces:
siempre deja lo mismo.

## Usuarios de prueba

`token_prueba.py <rol>` genera un token válido por 60 minutos, firmado con
el `JWT_SECRET` de este entorno (no sirve en ningún otro lado). Roles:
`role_admin`, `role_teacher`, `role_volunteer`, `role_participant`.

## Límites (lo que este entorno no prueba)

- Adjuntos de entregas (sin Cloud Storage): se aceptan solo con texto.
- Lo que consulta a auth-service (por ejemplo, datos de participantes)
  responde que no está configurado.
- Eventos de Pub/Sub y sugerencias de IA.
