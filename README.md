# Sistema Informático Buen Vivir

Plataforma basada en **Arquitectura de Microservicios** (Go), **Google Cloud API Gateway** y **Frontend Angular PWA** bajo el patrón de diseño **Modelo-Vista-Controlador (MVC)**

---

## 🏗️ Arquitectura del Sistema

```
Sistema-informatico-Buen-Vivir/
├── api-gateway/                      # Especificación OpenAPI para Google Cloud API Gateway
├── auth/                             # Microservicio de Autenticación y Usuarios (Go MVC)
├── campaigns/                        # Microservicio de Gestión de Campañas (Go MVC)
├── dashboard/                        # Microservicio de Métricas y Tableros (Go MVC)
├── education/                        # Microservicio de Capacitaciones y Cursos (Go MVC)
├── human-promotion/                 # Microservicio de Promoción Humana (Go MVC)
├── labor-intetgration/              # Microservicio de Inserción Laboral (Go MVC)
├── notifications/                    # Microservicio de Notificaciones (Go MVC)
├── rxresume/                         # Microservicio de Generación de CVs (Go MVC)
├── frontend/                         # Aplicación Web Progresiva (Angular PWA)
└── docker-compose.yml                # Orquestación de entorno de desarrollo local
```

---

## 🛠️ Estructura MVC en Microservicios Go

Cada microservicio backend sigue la separación limpia de capas MVC:

* **Model (`internal/models/`)**: Entidades de base de datos, structs Go y DTOs de solicitud/respuesta.
* **Controller (`internal/controllers/`)**: Controladores HTTP (Handlers) encargados del parsing, validación de entrada y formateo de respuesta JSON.
* **Service (`internal/services/`)**: Capa de negocio autónoma.
* **Repository (`internal/repositories/`)**: Capa de persistencia a la base de datos (PostgreSQL/Cloud SQL).
* **Routes (`internal/routes/`)**: Definición de endpoints HTTP.

---

## 🚀 Ejecución en Entorno Local (Docker)

Para iniciar todos los microservicios, la base de datos y la PWA localmente:

```bash
docker compose up --build
```

### Puertos Locales
* **Frontend PWA**: `http://localhost:4200`
* **Auth Service**: `http://localhost:8081`
* **Campaigns Service**: `http://localhost:8082`
* **Dashboard Service**: `http://localhost:8083`
* **Education Service**: `http://localhost:8084`
* **Human Promotion**: `http://localhost:8085`
* **Labor Integration**: `http://localhost:8086`
* **Notifications**: `http://localhost:8087`
* **RxResume**: `http://localhost:8088`
* **PostgreSQL**: `localhost:5432`