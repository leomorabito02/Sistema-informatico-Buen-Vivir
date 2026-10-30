# Entorno de calidad (TP de Calidad de Software)

Una copia de Educación para probar sin tocar nada real (decisión D-22).
Se define en `docker-compose.calidad.yml`, en la raíz del repo.

| Pieza | Qué usa | Puerto en la PC |
|---|---|---|
| Educación | El mismo `education/Dockerfile` | `127.0.0.1:18210` |
| Firestore | Emulador, vacío en cada arranque | `127.0.0.1:18390` |
| Contentful | El real, environment `testing` (solo lectura desde acá) | — |
| auth-service | Uno de prueba (`auth_falso.py`), con 8 participantes de prueba | — (solo dentro de la red del entorno) |
| Storage, Pub/Sub, IA | Apagados | — |

No comparte red, puertos ni volúmenes con `docker-compose.yml`: los dos
stacks pueden estar prendidos a la vez.

## Cómo se usa

Todo desde la raíz del repo, en Git Bash.

```bash
# 1. Una sola vez: arma .env.calidad (Contentful testing + JWT propio)
bash scripts/calidad/entorno/preparar_env.sh

# 2. Prender el entorno (la primera vez compila la imagen)
docker compose -f docker-compose.calidad.yml up -d --build

# 3. Cargar los datos de prueba por la API (cuando /api/salud responda 200)
curl http://127.0.0.1:18210/api/salud
python scripts/calidad/entorno/cargar_por_api.py

# 4. Pedir algo con un usuario de prueba
TOKEN=$(python scripts/calidad/entorno/token_prueba.py role_admin)
curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1:18210/api/cursos

# 5. Apagar y borrar (los datos se pierden: es a propósito)
docker compose -f docker-compose.calidad.yml down
```

## Datos de prueba

`cargar_por_api.py` no escribe en la base: usa la API de Educación igual
que la aplicación, así cada dato pasa por las mismas validaciones que uno
real. Toma 3 cursos publicados de Contentful `testing` y en cada uno:

1. un admin crea 2 actividades y les asigna habilidades;
2. un admin inscribe a los 8 participantes de prueba (Educación verifica
   contra `auth_falso.py` que existan y estén aprobados);
3. 5 participantes entregan la primera actividad con un texto. Educación
   registra la asistencia y otorga las habilidades.

Resultado: 6 actividades, 24 inscripciones, 15 entregas con su asistencia
y habilidades ganadas. Ningún dato es de una persona real. Se puede correr
varias veces: lo que ya existe lo saltea.

Las inscripciones quedan todas activas: la API no tiene forma de dar de
baja ni finalizar una inscripción.

## Usuarios de prueba

`token_prueba.py <rol>` genera un token válido por 60 minutos, firmado con
el `JWT_SECRET` de este entorno (no sirve en ningún otro lado). Roles:
`role_admin`, `role_teacher`, `role_volunteer`, `role_participant`.

## Límites (lo que este entorno no prueba)

- Adjuntos de entregas (sin Cloud Storage): se aceptan solo con texto.
- auth-service es de prueba: no prueba el login ni los permisos de
  auth-service real, solo lo que Educación le consulta.
- Eventos de Pub/Sub y sugerencias de IA.
