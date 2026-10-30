"""Carga los datos de prueba del entorno de calidad usando la API de Educación.

Hace lo mismo que haría la gente con la aplicación, así cada dato pasa por
las mismas validaciones que un dato real:

1. Un admin crea 2 actividades en cada curso (POST /api/cursos/{id}/actividades).
2. Les asigna habilidades (PUT /api/actividades/{id}/habilidades).
3. Un admin inscribe a los 8 participantes de prueba en cada curso
   (POST /api/cursos/{id}/inscripciones). Educación verifica contra el
   auth-service de prueba que existan y estén aprobados.
4. Algunos participantes entregan la primera actividad con un texto
   (POST /api/participante/actividades/{id}/entrega). Al entregar, Educación
   registra la asistencia y les otorga las habilidades.

Los cursos son los reales del environment "testing" de Contentful: se leen,
no se modifican. Los participantes son los del auth-service de prueba
(participante-prueba-01 a 08). No hay datos de personas reales.

Se puede correr varias veces: lo que ya existe se saltea.

Uso (desde la raíz del repo, con el entorno levantado):
  python scripts/calidad/entorno/cargar_por_api.py
Solo usa la biblioteca estándar.
"""

import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.request
import uuid

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from token_prueba import token as generar_token  # noqa: E402

PARTICIPANTES = [(f"usuario-prueba-{n:02d}", f"participante-prueba-{n:02d}") for n in range(1, 9)]
ACTIVIDADES = [
    ("Presentación del curso", "Contá en pocas líneas qué esperás aprender en este curso.", "2026-09-01T00:00:00Z",
     [("Comunicación oral", "Socioemocional"), ("Trabajo en equipo", "Socioemocional")]),
    ("Trabajo práctico 1", "Resolvé la consigna en grupo y subí una breve conclusión.", "2026-09-15T00:00:00Z",
     [("Trabajo en equipo", "Socioemocional"), ("Uso de planilla de cálculo", "Tecnología")]),
]
# Cuántos participantes entregan la primera actividad de cada curso.
ENTREGAN = 5


def pedir(api, metodo, ruta, tok, cuerpo=None, multipart=None):
    """Hace el pedido y devuelve (código, respuesta en JSON o texto)."""
    datos, tipo = None, None
    if cuerpo is not None:
        datos, tipo = json.dumps(cuerpo).encode(), "application/json"
    elif multipart is not None:
        borde = uuid.uuid4().hex
        partes = [f'--{borde}\r\nContent-Disposition: form-data; name="{k}"\r\n\r\n{v}\r\n' for k, v in multipart.items()]
        datos = ("".join(partes) + f"--{borde}--\r\n").encode()
        tipo = f"multipart/form-data; boundary={borde}"
    # Educación limita los pedidos por segundo (429): se espera y se reintenta.
    for intento in range(6):
        req = urllib.request.Request(api + ruta, data=datos, method=metodo)
        req.add_header("Authorization", f"Bearer {tok}")
        if tipo:
            req.add_header("Content-Type", tipo)
        espera = None
        try:
            with urllib.request.urlopen(req, timeout=30) as r:
                texto = r.read().decode()
                codigo = r.status
        except urllib.error.HTTPError as e:
            texto, codigo, espera = e.read().decode(), e.code, e.headers.get("Retry-After")
        if codigo != 429:
            break
        time.sleep(int(espera or 2 ** intento))
    try:
        return codigo, json.loads(texto) if texto else None
    except json.JSONDecodeError:
        return codigo, texto


def datos_de(respuesta):
    """La API a veces envuelve la respuesta en {"data": ...}."""
    if isinstance(respuesta, dict) and "data" in respuesta:
        return respuesta["data"]
    return respuesta


def main():
    parser = argparse.ArgumentParser(description="Carga datos de prueba por la API de Educación.")
    parser.add_argument("--api", default="http://127.0.0.1:18210")
    parser.add_argument("--cursos", type=int, default=3, help="cuántos cursos publicados usar")
    args = parser.parse_args()

    admin = generar_token("role_admin", "admin-prueba")
    resumen = {"actividades": 0, "habilidades": 0, "inscripciones": 0, "entregas": 0, "ya_existian": 0}

    codigo, resp = pedir(args.api, "GET", "/api/cursos", admin)
    if codigo != 200:
        raise SystemExit(f"No pude leer los cursos ({codigo}): {resp}")
    cursos = [c["id"] for c in datos_de(resp)][: args.cursos]
    if not cursos:
        raise SystemExit("No hay cursos publicados en Contentful testing.")

    for curso in cursos:
        # 1 y 2. Actividades y sus habilidades.
        _, resp = pedir(args.api, "GET", f"/api/cursos/{curso}/actividades", admin)
        existentes = {a["titulo"]: a["id"] for a in (datos_de(resp) or [])}
        ids = []
        for titulo, consigna, fecha, habilidades in ACTIVIDADES:
            if titulo in existentes:
                ids.append(existentes[titulo])
                resumen["ya_existian"] += 1
                continue
            codigo, resp = pedir(args.api, "POST", f"/api/cursos/{curso}/actividades", admin,
                                 {"titulo": titulo, "consigna": consigna, "fecha": fecha})
            if codigo not in (200, 201):
                raise SystemExit(f"No pude crear la actividad '{titulo}' ({codigo}): {resp}")
            actividad = datos_de(resp)["id"]
            ids.append(actividad)
            resumen["actividades"] += 1
            codigo, resp = pedir(args.api, "PUT", f"/api/actividades/{actividad}/habilidades", admin,
                                 {"habilidades": [{"nombre": n, "categoria": c} for n, c in habilidades]})
            if codigo != 200:
                raise SystemExit(f"No pude asignar habilidades a '{titulo}' ({codigo}): {resp}")
            resumen["habilidades"] += len(habilidades)

        # 3. Inscripciones.
        for _, participante in PARTICIPANTES:
            codigo, resp = pedir(args.api, "POST", f"/api/cursos/{curso}/inscripciones", admin,
                                 {"participanteId": participante})
            if codigo in (200, 201):
                resumen["inscripciones"] += 1
            elif "ya se encuentra inscripto" in json.dumps(resp, ensure_ascii=False):
                resumen["ya_existian"] += 1
            else:
                raise SystemExit(f"No pude inscribir a {participante} ({codigo}): {resp}")

        # 4. Entregas de la primera actividad.
        for usuario, participante in PARTICIPANTES[:ENTREGAN]:
            tok = generar_token("role_participant", usuario)
            codigo, resp = pedir(args.api, "POST", f"/api/participante/actividades/{ids[0]}/entrega", tok,
                                 multipart={"texto": f"Entrega de prueba de {participante}."})
            if codigo in (200, 201):
                resumen["entregas"] += 1
            elif "Ya entregaste" in json.dumps(resp, ensure_ascii=False):
                resumen["ya_existian"] += 1
            else:
                raise SystemExit(f"No pude entregar como {participante} ({codigo}): {resp}")

    print(f"Cursos usados: {len(cursos)}")
    for clave, valor in resumen.items():
        print(f"  {clave.replace('_', ' ')}: {valor}")


if __name__ == "__main__":
    main()
