# API Gateway - Google Cloud

Este directorio contiene la especificación OpenAPI utilizada por **Google Cloud API Gateway** para enrutar el tráfico del frontend PWA hacia los microservicios backend desplegados en GCP Cloud Run o Compute Engine.

## Despliegue en GCP

```bash
# 1. Crear la configuración de API
gcloud api-gateway api-configs create buenvivir-config \
  --api=buenvivir-api \
  --openapi-spec=openapi.yaml \
  --project=TU_PROJECT_ID

# 2. Desplegar o actualizar el Gateway
gcloud api-gateway gateways create buenvivir-gateway \
  --api=buenvivir-api \
  --api-config=buenvivir-config \
  --location=us-central1 \
  --project=TU_PROJECT_ID
```
