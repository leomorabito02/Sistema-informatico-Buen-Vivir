"""Carga datos sintéticos en el Firestore del entorno de calidad.

Los cursos son los reales del environment "testing" de Contentful (se leen
por la API de Educación, sin modificarlos). Todo lo demás es inventado:
participantes "participante-sint-NN", actividades, inscripciones y
habilidades. No hay nombres, DNI, teléfonos ni emails de personas reales.

Se puede correr varias veces: cada documento tiene un id fijo y se
sobrescribe con el mismo contenido.

Uso (desde la raíz del repo, con el entorno levantado):
  python scripts/calidad/entorno/sembrar.py
Solo usa la biblioteca estándar.
"""

import argparse
import json
import os
import sys
import urllib.request
from datetime import datetime, timedelta, timezone

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from token_prueba import token as generar_token  # noqa: E402

PARTICIPANTES = [f"participante-sint-{n:02d}" for n in range(1, 9)]
# id = clave normalizada del nombre, igual que models.SkillKey.
HABILIDADES = [
    ("trabajo-en-equipo", "Trabajo en equipo", "Socioemocional"),
    ("comunicacion-oral", "Comunicación oral", "Socioemocional"),
    ("huerta-agroecologica", "Huerta agroecológica", "Cuidado de la naturaleza"),
    ("compostaje", "Compostaje", "Cuidado de la naturaleza"),
    ("uso-de-planilla-de-calculo", "Uso de planilla de cálculo", "Tecnología"),
    ("carpinteria-basica", "Carpintería básica", "Oficio"),
]
# Mezcla de estados para que los listados tengan casos distintos.
ESTADOS = ["activa", "activa", "activa", "finalizada", "baja"]


def valor(v):
    if isinstance(v, bool):
        return {"booleanValue": v}
    if isinstance(v, datetime):
        return {"timestampValue": v.strftime("%Y-%m-%dT%H:%M:%SZ")}
    return {"stringValue": str(v)}


def guardar(base, coleccion, doc_id, campos):
    # PATCH sobre un id fijo crea o reemplaza el documento (idempotente).
    url = f"{base}/{coleccion}/{doc_id}"
    cuerpo = json.dumps({"fields": {k: valor(v) for k, v in campos.items()}}).encode()
    req = urllib.request.Request(url, data=cuerpo, method="PATCH",
                                 headers={"Content-Type": "application/json",
                                          "Authorization": "Bearer owner"})
    urllib.request.urlopen(req, timeout=10).close()


def cursos(api):
    req = urllib.request.Request(f"{api}/api/cursos",
                                 headers={"Authorization": "Bearer " + generar_token("role_admin", "admin-sint-01")})
    with urllib.request.urlopen(req, timeout=30) as resp:
        return [c["id"] for c in json.load(resp)]


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--api", default="http://127.0.0.1:18210")
    p.add_argument("--firestore", default="http://127.0.0.1:18390")
    p.add_argument("--proyecto", default="buenvivir-calidad")
    p.add_argument("--cursos", type=int, default=3, help="cuántos cursos de testing usar")
    a = p.parse_args()

    base = f"{a.firestore}/v1/projects/{a.proyecto}/databases/(default)/documents"
    ids = cursos(a.api)[: a.cursos]
    if not ids:
        sys.exit("Contentful testing no devolvió cursos publicados.")

    # Fechas fijas: los datos son los mismos en cada corrida.
    inicio = datetime(2026, 9, 1, 12, tzinfo=timezone.utc)
    total = {"habilidades": 0, "actividades": 0, "actividad_habilidades": 0,
             "inscripciones": 0, "participante_habilidades": 0}

    for hid, nombre, categoria in HABILIDADES:
        guardar(base, "habilidades", hid, {"nombre": nombre, "categoria": categoria})
        total["habilidades"] += 1

    for c, curso in enumerate(ids):
        for n in (1, 2):
            act = f"act-sint-{c + 1}-{n}"
            fecha = inicio + timedelta(days=7 * (c * 2 + n))
            guardar(base, "actividades", act, {
                "curso_id": curso, "titulo": f"Actividad sintética {n}",
                "consigna": "Consigna de prueba del entorno de calidad.",
                "fecha": fecha, "created_at": inicio})
            total["actividades"] += 1
            hid = HABILIDADES[(c * 2 + n) % len(HABILIDADES)][0]
            guardar(base, "actividad_habilidades", f"{act}_{hid}",
                    {"actividad_id": act, "habilidad_id": hid, "curso_id": curso})
            total["actividad_habilidades"] += 1

        for i, part in enumerate(PARTICIPANTES):
            estado = ESTADOS[(i + c) % len(ESTADOS)]
            campos = {"participante_id": part, "curso_id": curso,
                      "fecha_inscripcion": inicio, "estado": estado}
            if estado == "baja":
                campos.update({"motivo_baja": "Baja sintética", "fecha_baja": inicio + timedelta(days=20)})
            guardar(base, "inscripciones", f"insc-sint-{c + 1}-{i + 1:02d}", campos)
            total["inscripciones"] += 1

        # Los dos primeros participantes ganan la habilidad de la actividad 1.
        hid = HABILIDADES[(c * 2 + 1) % len(HABILIDADES)][0]
        for part in PARTICIPANTES[:2]:
            guardar(base, "participante_habilidades", f"{part}_{hid}", {
                "participante_id": part, "habilidad_id": hid, "origen": "actividad",
                "fecha": inicio + timedelta(days=10), "actividad_id": f"act-sint-{c + 1}-1",
                "curso_id": curso})
            total["participante_habilidades"] += 1

    print(f"Cursos de Contentful testing usados: {len(ids)}")
    for k, v in total.items():
        print(f"  {k}: {v}")


if __name__ == "__main__":
    main()
