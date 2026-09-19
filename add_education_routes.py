"""
Agrega al api-gateway/openapi.yaml (Google Cloud API Gateway, Swagger 2.0)
las rutas de education-service, apuntándolas a ${EDUCATION_SERVICE_URL}.

Mismo estilo que el resto del spec (ver generate_openapi.py / add_options.py):
cada operación lleva x-google-backend con APPEND_PATH_TO_ADDRESS + jwt_audience,
y cada path lleva un 'options' para el preflight CORS. Igual que auth, el
gateway NO valida el JWT (no hay bloque 'security'): cada microservicio
valida su propio token.

Idempotente: correrlo de nuevo reemplaza las rutas de education, no las
duplica. Uso:  python add_education_routes.py
"""

import yaml

SPEC_PATH = "api-gateway/openapi.yaml"
BACKEND_VAR = "${EDUCATION_SERVICE_URL}"

# path -> métodos HTTP que expone education (ver education/internal/routes/routes.go)
EDUCATION_PATHS = {
    "/api/salud": ["get"],
    "/api/cursos": ["get", "post"],
    "/api/cursos/borradores": ["get"],
    "/api/cursos/{id}": ["get", "put", "delete"],
    "/api/cursos/{id}/publicar": ["post"],
    "/api/cursos/{id}/despublicar": ["post"],
    "/api/areas-formativas": ["get", "post"],
    "/api/areas-formativas/{id}": ["put", "delete"],
    "/api/materiales": ["get", "post"],
    "/api/materiales/upload": ["post"],
    "/api/materiales/{id}": ["put", "delete"],
    "/api/docente/cursos": ["get", "post"],
    "/api/docente/cursos/borradores": ["get"],
    "/api/docente/cursos/{id}": ["put"],
    "/api/docente/cursos/{cursoId}/actividades": ["get", "post"],
    "/api/docente/actividades/{actividadId}/asistencia": ["get", "post"],
    "/api/docente/cursos/{cursoId}/inscriptos": ["get"],
    "/api/docente/cursos/{cursoId}/progreso": ["get"],
    # role_admin / role_volunteer, sin scope de "mis cursos" (agregado
    # 2026-09-11 — ver education/CLAUDE.md, sección de Endpoints).
    "/api/cursos/{id}/inscripciones": ["get", "post"],
    "/api/actividades/{actividadId}/asistencia": ["get", "post"],
    "/api/asistencia-general": ["get", "post"],
}

CORS_HEADERS = {
    "Access-Control-Allow-Origin": {"type": "string"},
    "Access-Control-Allow-Methods": {"type": "string"},
    "Access-Control-Allow-Headers": {"type": "string"},
    "Access-Control-Expose-Headers": {"type": "string"},
    "Access-Control-Allow-Credentials": {"type": "boolean"},
    "Access-Control-Max-Age": {"type": "integer"},
}


class NoAliasDumper(yaml.SafeDumper):
    def ignore_aliases(self, data):
        return True


def slug(path):
    return path.replace("/", "_").replace("{", "").replace("}", "")


def backend():
    return {
        "address": BACKEND_VAR,
        "path_translation": "APPEND_PATH_TO_ADDRESS",
        "jwt_audience": BACKEND_VAR,
    }


def operation(method, path):
    return {
        "summary": f"{method.upper()} {path}",
        "operationId": f"{method}_{slug(path)}",
        "x-google-backend": backend(),
        "responses": {"200": {"description": "OK"}},
    }


def cors_options(path):
    return {
        "summary": f"CORS preflight for {path}",
        "operationId": f"options_{slug(path)}",
        "x-google-backend": backend(),
        "responses": {"200": {"description": "OK", "headers": dict(CORS_HEADERS)}},
    }


def path_params(path):
    names = [seg[1:-1] for seg in path.split("/") if seg.startswith("{") and seg.endswith("}")]
    return [{"name": n, "in": "path", "required": True, "type": "string"} for n in names]


def build_path_item(path, methods):
    item = {}
    for m in methods:
        item[m] = operation(m, path)
    params = path_params(path)
    if params:
        item["parameters"] = params
    item["options"] = cors_options(path)
    return item


def main():
    with open(SPEC_PATH, "r", encoding="utf-8") as f:
        spec = yaml.safe_load(f)

    paths = spec.setdefault("paths", {})

    # limpiar cualquier ruta de education de una corrida anterior
    for p in list(paths):
        if p in EDUCATION_PATHS:
            del paths[p]

    for path, methods in EDUCATION_PATHS.items():
        paths[path] = build_path_item(path, methods)

    with open(SPEC_PATH, "w", encoding="utf-8") as f:
        yaml.dump(spec, f, sort_keys=False, allow_unicode=True, Dumper=NoAliasDumper)

    print(f"OK: {len(EDUCATION_PATHS)} rutas de education agregadas a {SPEC_PATH}")


if __name__ == "__main__":
    main()
