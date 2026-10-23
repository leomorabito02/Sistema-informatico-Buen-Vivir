"""auth-service de prueba para el entorno de calidad.

Educación le consulta a auth-service tres cosas: si un participante existe
y está aprobado (para inscribirlo), el listado de participantes y usuarios
(para la asistencia general) y "quién soy" (para saber qué participante
entrega). Este servidor responde esas mismas rutas, con el mismo formato que
auth-service, para 8 participantes de prueba. No tiene datos de personas
reales ni se conecta a nada.

Igual que auth-service, exige un token firmado con el JWT_SECRET del
entorno. Solo es visible dentro de la red del entorno (no publica puertos).

Uso: lo levanta docker-compose.calidad.yml. Para correrlo suelto:
  JWT_SECRET=... python scripts/calidad/entorno/auth_falso.py --puerto 18080
Solo usa la biblioteca estándar.
"""

import argparse
import base64
import hashlib
import hmac
import json
import os
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

CANTIDAD = 8
PARTICIPANTES = [
    {"participant_id": f"participante-prueba-{n:02d}", "user_id": f"usuario-prueba-{n:02d}", "approval_status": "approved"}
    for n in range(1, CANTIDAD + 1)
]
USUARIOS = [
    {"user_id": f"usuario-prueba-{n:02d}", "first_name": f"Participante {n:02d}", "last_name": "Prueba"}
    for n in range(1, CANTIDAD + 1)
]
POR_USUARIO = {p["user_id"]: p for p in PARTICIPANTES}
POR_ID = {p["participant_id"]: p for p in PARTICIPANTES}


def b64(texto):
    return base64.urlsafe_b64decode(texto + "=" * (-len(texto) % 4))


def validar(token, secreto):
    """Devuelve los claims si el token es HS256, de acceso y no venció."""
    try:
        cabecera, cuerpo, firma = token.split(".")
        esperada = hmac.new(secreto.encode(), f"{cabecera}.{cuerpo}".encode(), hashlib.sha256).digest()
        if not hmac.compare_digest(esperada, b64(firma)):
            return None
        if json.loads(b64(cabecera)).get("alg") != "HS256":
            return None
        claims = json.loads(b64(cuerpo))
        if claims.get("type") != "access" or claims.get("exp", 0) < time.time():
            return None
        return claims
    except (ValueError, json.JSONDecodeError):
        return None


class Manejador(BaseHTTPRequestHandler):
    secreto = ""

    def responder(self, codigo, cuerpo):
        datos = json.dumps(cuerpo).encode()
        self.send_response(codigo)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(datos)))
        self.end_headers()
        self.wfile.write(datos)

    def do_GET(self):  # noqa: N802 (nombre que exige http.server)
        if self.path == "/health":
            return self.responder(200, {"status": "UP"})
        encabezado = self.headers.get("Authorization", "")
        claims = validar(encabezado.removeprefix("Bearer "), self.secreto) if encabezado.startswith("Bearer ") else None
        if claims is None:
            return self.responder(401, {"success": False, "errors": "token inválido"})

        if self.path == "/api/v1/participants":
            return self.responder(200, {"success": True, "data": PARTICIPANTES})
        if self.path.startswith("/api/v1/participants/"):
            p = POR_ID.get(self.path.rsplit("/", 1)[-1])
            if p is None:
                return self.responder(404, {"success": False, "errors": "no existe"})
            return self.responder(200, {"success": True, "data": p})
        if self.path == "/api/v1/users":
            return self.responder(200, {"success": True, "data": USUARIOS})
        if self.path == "/api/v1/profile/me":
            p = POR_USUARIO.get(claims.get("sub", ""))
            return self.responder(200, {"role_data": {"participant_id": p["participant_id"] if p else ""}})
        return self.responder(404, {"success": False, "errors": "ruta desconocida"})

    def log_message(self, formato, *args):
        print("auth-falso:", formato % args, flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--puerto", type=int, default=8080)
    args = parser.parse_args()
    secreto = os.environ.get("JWT_SECRET", "")
    if len(secreto) < 32:
        raise SystemExit("auth-falso: falta JWT_SECRET (corré preparar_env.sh)")
    Manejador.secreto = secreto
    print(f"auth-falso escuchando en :{args.puerto}", flush=True)
    ThreadingHTTPServer(("0.0.0.0", args.puerto), Manejador).serve_forever()


if __name__ == "__main__":
    main()
