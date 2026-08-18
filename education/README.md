# Microservicio de Educación

Integra el módulo de Educación (cursos, áreas formativas, materiales)
con Contentful como headless CMS. Ver `../CLAUDE.md` para el diseño
completo (Content Types, endpoints, decisiones de arquitectura).

Corre en el puerto **8210** (distinto del 8080 de Autenticación).

## Levantar el servicio con Docker

1. Elegí el ambiente contra el que querés correr y copiá la plantilla
   correspondiente a `.env`:

   ```bash
   # Para probar contra el Environment "testing" de Contentful
   cp .env.testing.example .env

   # Para correr contra el Environment "master" (producción) de Contentful
   cp .env.production.example .env
   ```

2. Completá en `.env` los valores reales de Contentful:
   `CONTENTFUL_SPACE_ID`, `CONTENTFUL_CDA_TOKEN`, `CONTENTFUL_CMA_TOKEN`.
   `APP_ENV` y `CONTENTFUL_ENVIRONMENT` ya vienen coherentes según la
   plantilla que copiaste — no los mezcles (el servicio rechaza
   arrancar si detecta una combinación peligrosa).

3. Levantá el contenedor:

   ```bash
   docker compose up --build
   ```

4. Verificá que respondió en el ambiente esperado:

   ```bash
   curl http://localhost:8210/api/salud
   # {"status":"UP","service":"education-service","appEnv":"testing","contentfulEnvironment":"testing"}
   ```

## Endpoints

Ver la lista completa y su justificación (RF-14 a RF-17, RF-28) en
`../CLAUDE.md`.
