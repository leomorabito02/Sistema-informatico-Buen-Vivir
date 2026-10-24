"""Resume los reportes de las herramientas de seguridad de Educación.

Lee los reportes que deja el job "seguridad-education" de calidad-tp.yml
(gosec, govulncheck, gitleaks, Semgrep y Trivy), escribe una tabla en
Markdown y termina con error solo si hay un hallazgo bloqueante:

- govulncheck: una vulnerabilidad conocida que el código de Educación
  realmente usa.
- Trivy: una vulnerabilidad CRITICAL o HIGH con arreglo disponible en la
  imagen Docker (mismo criterio que pipeline.yml).
- gosec: vulnerabilidades de severidad alta encontradas en el código Go.
- Gitleaks: secretos, contraseñas o tokens expuestos.
- Semgrep: patrones de código inseguro de severidad ERROR.

Uso: python scripts/calidad/seguridad/resumen_seguridad.py <carpeta-de-reportes>
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
        filas.append(("gosec", "Errores de seguridad en el código Go", "sin reporte", "bloquea"))
        bloquea.append("gosec no generó reporte")
    else:
        sev = Counter(i["severity"] for i in gosec.get("Issues") or [])
        altos = sev["HIGH"]
        filas.append(("gosec", "Errores de seguridad en el código Go",
                      f"{sum(sev.values())} avisos (altos: {altos}, medios: {sev['MEDIUM']}, bajos: {sev['LOW']})",
                      "bloquea"))
        if altos > 0:
            bloquea.append(f"gosec encontró {altos} problemas de severidad alta")

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
        filas.append(("Gitleaks", "Claves o contraseñas escritas en archivos", "sin reporte", "bloquea"))
        bloquea.append("Gitleaks no generó reporte")
    else:
        archivos = sorted({h["File"].split("education/")[-1] for h in gitleaks})
        arch_str = ", ".join(archivos)
        detalle = f"{len(gitleaks)} posibles" + (f" en {arch_str}" if archivos else "")
        filas.append(("Gitleaks", "Claves o contraseñas escritas en archivos", detalle, "bloquea"))
        if gitleaks:
            bloquea.append(f"Gitleaks encontró {len(gitleaks)} posibles secretos en {arch_str}")

    semgrep = leer_json(carpeta / "semgrep.json")
    if semgrep is None:
        filas.append(("Semgrep", "Patrones de código inseguro (reglas de la comunidad)", "sin reporte", "bloquea"))
        bloquea.append("Semgrep no generó reporte")
    else:
        sev = Counter(r["extra"]["severity"] for r in semgrep.get("results", []))
        errores = sev["ERROR"]
        filas.append(("Semgrep", "Patrones de código inseguro (reglas de la comunidad)",
                      f"{sum(sev.values())} avisos (ERROR: {errores}, WARNING: {sev['WARNING']}, INFO: {sev['INFO']})",
                      "bloquea"))
        if errores > 0:
            bloquea.append(f"Semgrep encontró {errores} problemas de severidad ERROR")

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
