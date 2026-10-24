"""E2E de Educación: el recorrido completo de un curso, de punta a punta.

Pasa por las tres dependencias reales del entorno de calidad: Contentful
testing (el curso), Firestore (actividades, inscripciones, entregas,
asistencia y habilidades) y el auth-service de prueba (el participante).

  1. Un admin crea un curso [E2E] en Contentful testing (borrador).
  2. Lo publica y el curso aparece en el catálogo.
  3. Crea una actividad y le asigna una habilidad.
  4. Inscribe a un participante de prueba (HU-18).
  5. El participante ve el curso entre los suyos y entrega la actividad
     con un texto (HU-19bis).
  6. La entrega registra la asistencia (HU-19) y otorga la habilidad.
  7. El docente responsable ve el progreso del curso (HU-20).
  8. El admin da de baja el curso (baja lógica, HU-16).
Al final, pase lo que pase, borra de Contentful el curso [E2E]: cada
corrida escribe un solo curso.

Uso (desde la raíz del repo, con el entorno de calidad levantado y los
datos cargados): bash scripts/calidad/pruebas/e2e.sh
Solo usa la biblioteca estándar.
"""

import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request

AQUI = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(AQUI, "..", "entorno"))
from cargar_por_api import datos_de, pedir  # noqa: E402
from token_prueba import ENV, token  # noqa: E402

API = os.environ.get("BASE_URL", "http://127.0.0.1:18210")
REPORTES = os.environ.get("REPORTES", "reportes-servicio")
PREFIJO = "[E2E]"
USUARIO, PARTICIPANTE = "usuario-prueba-08", "participante-prueba-08"
DOCENTE = "docente-prueba-e2e"
HABILIDAD = ("Trabajo en equipo", "Socioemocional")

pasos = []


def paso(nombre, ok, detalle=""):
    pasos.append({"paso": nombre, "ok": ok, "detalle": detalle})
    print(("OK    " if ok else "FALLA ") + nombre + (f" ({detalle})" if detalle else ""))
    if not ok:
        raise AssertionError(nombre)


def esperar(fn, segundos=40):
    """Reintenta fn hasta que devuelva algo verdadero: la CDA de Contentful
    tarda unos segundos en mostrar un curso recién publicado."""
    fin = time.time() + segundos
    while True:
        r = fn()
        if r or time.time() > fin:
            return r
        time.sleep(3)


def credenciales():
    env = {}
    with open(ENV, encoding="utf-8") as f:
        for linea in f:
            if "=" in linea and not linea.startswith("#"):
                k, v = linea.rstrip("\n").split("=", 1)
                env[k] = v
    return env.get("CONTENTFUL_SPACE_ID", ""), env.get("CONTENTFUL_CMA_TOKEN", "")


def cma(metodo, ruta):
    espacio, tok = credenciales()
    if not espacio or not tok:
        return None
    req = urllib.request.Request(
        f"https://api.contentful.com/spaces/{espacio}/environments/testing{ruta}", method=metodo)
    req.add_header("Authorization", f"Bearer {tok}")
    try:
        with urllib.request.urlopen(req, timeout=30) as r:
            texto = r.read().decode()
            return json.loads(texto) if texto else {}
    except urllib.error.HTTPError:
        return None


def limpiar_contentful():
    """Borra los cursos [E2E] de Contentful testing (de esta corrida o de
    una anterior que se cortó)."""
    q = urllib.parse.quote(PREFIJO)
    resp = cma("GET", f"/entries?content_type=course&query={q}&limit=50") or {}
    borrados = 0
    for item in resp.get("items", []):
        # La búsqueda de Contentful es por palabras: se borra solo lo que
        # empieza con el prefijo, nunca otro curso que diga "E2E".
        titulos = (item.get("fields", {}).get("title") or {}).values()
        if not any(str(t).startswith(PREFIJO) for t in titulos):
            continue
        eid = item["sys"]["id"]
        if item["sys"].get("publishedVersion"):
            cma("DELETE", f"/entries/{eid}/published")
        if cma("DELETE", f"/entries/{eid}") is not None:
            borrados += 1
    return borrados


def main():
    os.makedirs(REPORTES, exist_ok=True)
    admin = token("role_admin", "admin-prueba")
    participante = token("role_participant", USUARIO)
    docente = token("role_teacher", DOCENTE)
    limpiar_contentful()
    curso = None
    try:
        codigo, resp = pedir(API, "GET", "/api/areas-formativas", admin)
        areas = datos_de(resp) or []
        codigo2, resp2 = pedir(API, "GET", "/api/materiales", admin)
        materiales = datos_de(resp2) or []
        paso("Hay área formativa y material en Contentful testing",
             codigo == 200 and codigo2 == 200 and bool(areas) and bool(materiales))

        titulo = f"{PREFIJO} Curso de punta a punta {int(time.time())}"
        cuerpo = {
            "title": titulo,
            "description": {"nodeType": "document", "data": {}, "content": [
                {"nodeType": "paragraph", "data": {}, "content": [
                    {"nodeType": "text", "value": "Curso creado por la prueba E2E del TP de Calidad.",
                     "marks": [], "data": {}}]}]},
            "responsibleTeacher": DOCENTE, "mode": "Virtual", "duration": "1 semana",
            "trainingAreaId": areas[0]["id"], "materialIds": [materiales[0]["id"]],
        }
        codigo, resp = pedir(API, "POST", "/api/cursos", admin, cuerpo)
        curso = (datos_de(resp) or {}).get("id") if isinstance(datos_de(resp), dict) else None
        paso("1. El admin crea el curso en Contentful (borrador)", codigo == 201 and bool(curso), f"{codigo}, id {curso}")

        codigo, resp = pedir(API, "POST", f"/api/cursos/{curso}/publicar", admin)
        paso("2. El admin publica el curso", codigo in (200, 204), str(codigo))
        visible = esperar(lambda: pedir(API, "GET", f"/api/cursos/{curso}", admin)[0] == 200)
        paso("2. El curso publicado se lee desde la CDA", bool(visible))

        codigo, resp = pedir(API, "POST", f"/api/cursos/{curso}/actividades", admin,
                             {"titulo": "Actividad E2E", "consigna": "Contá qué aprendiste.",
                              "fecha": "2026-10-01T00:00:00Z"})
        actividad = (datos_de(resp) or {}).get("id")
        paso("3. El admin crea una actividad", codigo in (200, 201) and bool(actividad), str(codigo))
        codigo, _ = pedir(API, "PUT", f"/api/actividades/{actividad}/habilidades", admin,
                          {"habilidades": [{"nombre": HABILIDAD[0], "categoria": HABILIDAD[1]}]})
        paso("3. Le asigna una habilidad", codigo == 200, str(codigo))

        inscripto = esperar(lambda: pedir(API, "POST", f"/api/cursos/{curso}/inscripciones", admin,
                                          {"participanteId": PARTICIPANTE})[0] in (200, 201))
        paso("4. El admin inscribe al participante (HU-18)", bool(inscripto))

        codigo, resp = pedir(API, "GET", "/api/participante/cursos", participante)
        ids = [c.get("id") for c in (datos_de(resp) or [])]
        paso("5. El participante ve el curso entre los suyos", codigo == 200 and curso in ids)
        codigo, resp = pedir(API, "POST", f"/api/participante/actividades/{actividad}/entrega", participante,
                             multipart={"texto": "Entrega de la prueba E2E."})
        paso("5. El participante entrega la actividad (HU-19bis)", codigo in (200, 201), str(codigo))

        codigo, resp = pedir(API, "GET", f"/api/actividades/{actividad}/asistencia", admin)
        presentes = [a["participanteId"] for a in (datos_de(resp) or []) if a.get("presente")]
        paso("6. La entrega registró la asistencia (HU-19)", codigo == 200 and PARTICIPANTE in presentes)
        codigo, resp = pedir(API, "GET", f"/api/participants/{PARTICIPANTE}/habilidades", admin)
        nombres = [h.get("nombre") for h in (datos_de(resp) or [])]
        paso("6. La entrega otorgó la habilidad", codigo == 200 and HABILIDAD[0] in nombres)

        codigo, resp = pedir(API, "GET", f"/api/docente/cursos/{curso}/progreso", docente)
        paso("7. El docente responsable ve el progreso (HU-20)", codigo == 200, str(codigo))

        codigo, resp = pedir(API, "DELETE", f"/api/cursos/{curso}", admin)
        paso("8. El admin da de baja el curso (HU-16)", codigo in (200, 204), str(codigo))
        aprobado = True
    except AssertionError:
        aprobado = False
    finally:
        borrados = limpiar_contentful()
        print(f"Limpieza: {borrados} curso(s) {PREFIJO} borrado(s) de Contentful testing")

    with open(os.path.join(REPORTES, "e2e.json"), "w", encoding="utf-8") as f:
        json.dump({"aprobado": aprobado, "curso": curso, "pasos": pasos,
                   "fecha": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())}, f, ensure_ascii=False, indent=2)
    print("E2E aprobado" if aprobado else "E2E con fallas")
    return 0 if aprobado else 1


if __name__ == "__main__":
    sys.exit(main())
