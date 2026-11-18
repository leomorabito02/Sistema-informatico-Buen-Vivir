# Documentación de Buen Vivir App

Carpeta única de documentación permanente del proyecto (ordenada en la
auditoría del 11/11/2026). El README de la raíz explica qué es el sistema y
cómo levantarlo; acá está el detalle. Cada tema vive en un solo lugar.

## Documentos base

| Documento | Para qué sirve |
|---|---|
| [ARQUITECTURA.md](ARQUITECTURA.md) | Componentes, comunicación, capas de cada servicio |
| [HISTORIAS_DE_USUARIO.md](HISTORIAS_DE_USUARIO.md) | Estado de las 64 HU (generado, no editar a mano) |
| [ROLES_Y_PERMISOS.md](ROLES_Y_PERMISOS.md) | Qué puede ver y hacer cada rol, en front y back |
| [BASE_DE_DATOS.md](BASE_DE_DATOS.md) | Dónde vive cada dato y colecciones/tablas principales |
| [API_Y_ENDPOINTS.md](API_Y_ENDPOINTS.md) | Rutas por servicio y gateway |
| [CONFIGURACION_Y_DESPLIEGUE.md](CONFIGURACION_Y_DESPLIEGUE.md) | Variables de entorno, Docker local, CI/CD |
| [TESTING.md](TESTING.md) | Qué tests hay, cómo correrlos, carga y seguridad |
| [PENDIENTES.md](PENDIENTES.md) | Lo que falta para considerar terminado el proyecto |
| [INFORME_AUDITORIA_BUEN_VIVIR.pdf](INFORME_AUDITORIA_BUEN_VIVIR.pdf) | Informe completo de la auditoría (11/11/2026) |

## Detalle técnico (ya existente)

| Documento | Tema |
|---|---|
| [adr/](adr/README.md) | Decisiones técnicas (ADR 0001 a 0009) |
| [eventos/README.md](eventos/README.md), [eventos/asyncapi.yaml](eventos/asyncapi.yaml) | Mensajería Pub/Sub: catálogo, reglas, cómo agregar eventos |
| [der_educacion.md](der_educacion.md) | Modelo de datos de Educación |
| [ia-habilidades.md](ia-habilidades.md) | Sugerencia de habilidades con IA (HU-21/58) |
| [ci-cd-pipeline.md](ci-cd-pipeline.md), [PIPELINE_EDUCATION.md](PIPELINE_EDUCATION.md) | Pipeline de CI/CD |
| [TESTING_EDUCACION.md](TESTING_EDUCACION.md) | Estrategia de testing de education (TP de Calidad) |
| `../education/CLAUDE.md` | Referencia completa del servicio de Educación |

## Otras carpetas

- `referencias/`: documentos fuente (HU en PDF, DER, explicación del DER,
  diagrama de arquitectura) y `referencias/despliegue/` (capturas de las
  variables de testing y `variables-de-entorno.pdf`; solo nombres, sin
  valores secretos).
- `diagramas/`: diagramas en Markdown (patrones de diseño, secuencia de login).
- `evidencias/`: resultados reales de la carga k6 y de la sonda de seguridad.
- `auditoria/`, `implementacion/`, `testing/`: informes en Word anteriores
  (octubre de 2026), se conservan como historial.
- `tesis/`, `Clases/`: material académico (no es documentación del sistema).
- `diseno/version-humand.html`: maqueta del rediseño de interfaz.

## Regenerar el informe y las HU

```bash
python scripts/informe-auditoria/generar.py
```

La fuente es `scripts/informe-auditoria/datos.py`.
