"""Lee de SonarQube las métricas de un proyecto y las agrega al historial.

Genera en --salida:
  historial.csv                     una fila por análisis (se acumula)
  <servicio>-<commit>.json          respuesta cruda de la API, como evidencia

Solo usa la biblioteca estándar para correr igual en Windows y en CI.
El token se lee de la variable SONAR_TOKEN.
"""

import argparse
import base64
import csv
import json
import os
import sys
import urllib.parse
import urllib.request
from datetime import datetime, timezone

# Métricas del modo estándar y del modo MQR: según cómo esté configurado el
# servidor, algunas no existen y quedan vacías en vez de inventarse.
METRICAS = [
    "ncloc", "complexity", "cognitive_complexity",
    "bugs", "vulnerabilities", "code_smells", "security_hotspots",
    "reliability_rating", "security_rating", "sqale_rating", "sqale_index",
    "coverage", "line_coverage", "branch_coverage", "lines_to_cover", "uncovered_lines",
    "duplicated_lines_density", "duplicated_blocks",
    "new_coverage", "new_duplicated_lines_density", "new_bugs", "new_lines",
    "tests", "test_failures", "test_errors",
    "software_quality_reliability_issues", "software_quality_security_issues",
    "software_quality_maintainability_issues", "software_quality_blocker_issues",
    "software_quality_high_issues",
]


def api(host, ruta, params):
    url = f"{host.rstrip('/')}{ruta}?{urllib.parse.urlencode(params)}"
    req = urllib.request.Request(url)
    token = os.environ.get("SONAR_TOKEN", "")
    req.add_header("Authorization", "Basic " + base64.b64encode(f"{token}:".encode()).decode())
    with urllib.request.urlopen(req, timeout=30) as resp:
        return json.load(resp)


def valor(medida):
    if "value" in medida:
        return medida["value"]
    # Las métricas de código nuevo vienen en "period".
    return medida.get("period", {}).get("value", "")


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--host", required=True)
    p.add_argument("--proyecto", required=True)
    p.add_argument("--servicio", required=True)
    p.add_argument("--commit", required=True)
    p.add_argument("--tests-ok", default="")
    p.add_argument("--salida", required=True)
    a = p.parse_args()

    medidas = api(a.host, "/api/measures/component",
                  {"component": a.proyecto, "metricKeys": ",".join(METRICAS)})
    gate = api(a.host, "/api/qualitygates/project_status", {"projectKey": a.proyecto})

    valores = {m["metric"]: valor(m) for m in medidas["component"].get("measures", [])}
    fila = {
        "fecha_utc": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "servicio": a.servicio,
        "proyecto_sonar": a.proyecto,
        "commit": a.commit,
        "tests_en_verde": a.tests_ok,
        "quality_gate": gate["projectStatus"]["status"],
    }
    fila.update({m: valores.get(m, "") for m in METRICAS})

    os.makedirs(a.salida, exist_ok=True)
    historial = os.path.join(a.salida, "historial.csv")
    nuevo = not os.path.exists(historial)
    with open(historial, "a", newline="", encoding="utf-8") as f:
        w = csv.DictWriter(f, fieldnames=list(fila.keys()))
        if nuevo:
            w.writeheader()
        w.writerow(fila)

    crudo = os.path.join(a.salida, f"{a.servicio}-{a.commit}.json")
    with open(crudo, "w", encoding="utf-8") as f:
        json.dump({"fila": fila, "medidas": medidas, "quality_gate": gate}, f,
                  ensure_ascii=False, indent=2)

    print(f"quality gate: {fila['quality_gate']}  cobertura: {fila['coverage']}%  "
          f"duplicación: {fila['duplicated_lines_density']}%  bugs: {fila['bugs']}")
    print(f"historial: {historial}")


if __name__ == "__main__":
    sys.exit(main())
