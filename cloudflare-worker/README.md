# Cloudflare Worker - Contadores de Impacto en Tiempo Real (KV)

Este Cloudflare Worker expone el endpoint en el Edge para la landing page del Centro de Ecología Integral Buen Vivir.

## Endpoints

### 1. Lectura de Estadísticas (Público)
- **URL**: `GET /api/impact`
- **Respuesta**:
  ```json
  {
    "activeParticipants": 248,
    "volunteers": 32,
    "courses": "Pendiente",
    "updatedAt": "2026-09-03T21:00:00.000Z"
  }
  ```

### 2. Sincronización de Estadísticas (Servidor Backend)
- **URL**: `POST /api/impact/update`
- **Headers**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <WORKER_SYNC_TOKEN>`
- **Body**:
  ```json
  {
    "activeParticipants": 250,
    "volunteers": 35
  }
  ```

## Despliegue

```bash
# Entorno Staging
pnpm deploy:staging

# Entorno Producción
pnpm deploy:prod
```
