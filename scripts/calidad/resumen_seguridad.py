"""Resume los reportes de las herramientas de seguridad de Educación.

Lee los reportes que deja el job "seguridad-education" de calidad-tp.yml
(gosec, govulncheck, gitleaks, Semgrep y Trivy), escribe una tabla en
Markdown y termina con error solo si hay un hallazgo bloqueante:

- govulncheck: una vulnerabilidad conocida que el código de Educación
  realmente usa.
- Trivy: una vulnerabilidad CRITICAL o HIGH con arreglo disponible en la
  imagen Docker (mismo criterio que pipeline.yml).

gosec, gitleaks y Semgrep se informan pero no frenan: tienen falsos
positivos (por ejemplo, claves de prueba en scripts de test) que primero
hay que revisar uno por uno.

Uso: python scripts/calidad/resumen_seguridad.py <carpeta-de-reportes>
"""

import json
import sys
from collections import Counter
from pathlib import Path


def leer_json(ruta):
    try:
        return json.loads(Path(ruta).read_text(encoding="utf-8"))
    except (OSError, ValueError):
        return None


def main():
    carpeta = Path(sys.argv[1])
    filas, bloquea = [], []

    gosec = leer_json(carpeta / "gosec.json")
    if gosec is None:
        filas.append(("gosec", "Errores de seguridad en el código Go", "sin reporte", "informa"))
    else:
        sev = Counter(i["severity"] for i in gosec.get("Issues") or [])
        filas.append(("gosec", "Errores de seguridad en el código Go",
                      f"{sum(sev.values())} avisos (altos: {sev['HIGH']}, medios: {sev['MEDIUM']}, bajos: {sev['LOW']})",
                      "informa"))

    codigo = (carpeta / "govulncheck.exit").read_text().strip() if (carpeta / "govulncheck.exit").exists() else ""
    if codigo == "0":
        filas.append(("govulncheck", "Librerías con vulnerabilidades conocidas que el código usa", "0", "bloquea"))
    elif codigo == "3":
        filas.append(("govulncheck", "Librerías con vulnerabilidades conocidas que el código usa", "hay al menos una (ver govulncheck.txt)", "bloquea"))
        bloquea.append("govulncheck encontró vulnerabilidades que el código usa")
    else:
        filas.append(("govulncheck", "Librerías con vulnerabilidades conocidas que el código usa", f"no terminó (código {codigo or '?'})", "bloquea"))
        bloquea.append("govulncheck no pudo terminar")

    gitleaks = leer_json(carpeta / "gitleaks.json")
    if gitleaks is None:
        filas.append(("Gitleaks", "Claves o contraseñas escritas en archivos", "sin reporte", "informa"))
    else:
        archivos = sorted({h["File"].split("education/")[-1] for h in gitleaks})
        detalle = f"{len(gitleaks)} posibles" + (f" en {', '.join(archivos)}" if archivos else "")
        filas.append(("Gitleaks", "Claves o contraseñas escritas en archivos", detalle, "informa"))

    semgrep = leer_json(carpeta / "semgrep.json")
    if semgrep is None:
        filas.append(("Semgrep", "Patrones de código inseguro (reglas de la comunidad)", "sin reporte", "informa"))
    else:
        sev = Counter(r["extra"]["severity"] for r in semgrep.get("results", []))
        filas.append(("Semgrep", "Patrones de código inseguro (reglas de la comunidad)",
                      f"{sum(sev.values())} avisos (ERROR: {sev['ERROR']}, WARNING: {sev['WARNING']}, INFO: {sev['INFO']})",
                      "informa"))

    trivy = leer_json(carpeta / "trivy.json")
    if trivy is None:
        filas.append(("Trivy", "Vulnerabilidades en la imagen Docker", "sin reporte", "bloquea"))
        bloquea.append("Trivy no generó reporte")
    else:
        vulns = [v for r in trivy.get("Results") or [] for v in r.get("Vulnerabilities") or []]
        sev = Counter(v["Severity"] for v in vulns)
        graves = [v for v in vulns if v["Severity"] in ("CRITICAL", "HIGH") and v.get("FixedVersion")]
        filas.append(("Trivy", "Vulnerabilidades en la imagen Docker",
                      f"{len(vulns)} (críticas: {sev['CRITICAL']}, altas: {sev['HIGH']}; con arreglo: {len(graves)})",
                      "bloquea"))
        if graves:
            bloquea.append(f"Trivy encontró {len(graves)} vulnerabilidades críticas o altas con arreglo disponible")

    print("### Educación — análisis de seguridad\n")
    print("| Herramienta | Qué busca | Resultado | ¿Frena el pipeline? |\n|---|---|---|---|")
    for nombre, que, resultado, efecto in filas:
        print(f"| {nombre} | {que} | {resultado} | {'Sí' if efecto == 'bloquea' else 'No, solo informa'} |")
    print()
    if bloquea:
        print("**Resultado: rojo.** " + "; ".join(bloquea) + ".")
    else:
        print("**Resultado: verde.** Ningún hallazgo bloqueante.")
    sys.exit(1 if bloquea else 0)


if __name__ == "__main__":
    main()
