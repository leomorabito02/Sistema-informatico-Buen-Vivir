"""Genera un token de acceso de prueba para el entorno de calidad.

El token tiene la misma forma que los de auth-service (HS256, claims sub,
role, type=access) y se firma con el JWT_SECRET de .env.calidad, que solo
existe en este entorno. Los usuarios son sintéticos.

Uso (desde la raíz del repo):
  python scripts/calidad/entorno/token_prueba.py role_admin
  python scripts/calidad/entorno/token_prueba.py role_teacher --usuario docente-sint-01
Roles: role_admin, role_teacher, role_volunteer, role_participant.
"""

import argparse
import base64
import hashlib
import hmac
import json
import os
import sys
import time

ENV = os.path.join(os.path.dirname(os.path.abspath(__file__)), ".env.calidad")
ROLES = ["role_admin", "role_teacher", "role_volunteer", "role_participant"]


def secreto():
    with open(ENV, encoding="utf-8") as f:
        for linea in f:
            if linea.startswith("JWT_SECRET="):
                return linea.split("=", 1)[1].strip()
    sys.exit("No hay JWT_SECRET en .env.calidad (correr preparar_env.sh).")


def b64(datos):
    return base64.urlsafe_b64encode(datos).rstrip(b"=").decode()


def token(rol, usuario, minutos=60):
    ahora = int(time.time())
    cabecera = {"alg": "HS256", "typ": "JWT"}
    claims = {"sub": usuario, "role": rol, "type": "access",
              "is_external": False, "iat": ahora, "exp": ahora + minutos * 60}
    cuerpo = b64(json.dumps(cabecera).encode()) + "." + b64(json.dumps(claims).encode())
    firma = hmac.new(secreto().encode(), cuerpo.encode(), hashlib.sha256).digest()
    return cuerpo + "." + b64(firma)


def main():
    p = argparse.ArgumentParser()
    p.add_argument("rol", choices=ROLES)
    p.add_argument("--usuario", help="id del usuario sintético (default según el rol)")
    a = p.parse_args()
    print(token(a.rol, a.usuario or a.rol.replace("role_", "") + "-sint-01"))


if __name__ == "__main__":
    main()
