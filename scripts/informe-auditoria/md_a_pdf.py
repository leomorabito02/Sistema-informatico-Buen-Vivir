"""Convierte el Markdown simple de la actualización (actualizacion.md) en
elementos de reportlab, con los mismos estilos del informe. Soporta:
títulos (#, ###, ####), párrafos, listas con "-" y "1.", tablas con "|",
**negrita** y `código`. Así el PDF y el .md dicen exactamente lo mismo."""
import re
from xml.sax.saxutils import escape

from reportlab.platypus import Paragraph


def _inline(texto):
    t = escape(texto.strip())
    t = re.sub(r"\*\*(.+?)\*\*", r"<b>\1</b>", t)
    t = re.sub(r"`([^`]+)`", r'<font face="Courier" size="8">\1</font>', t)
    return t


def convertir(md, S, tabla_raw):
    """md: texto. S: estilos del informe. tabla_raw(data, anchos) arma una
    tabla con celdas ya formateadas."""
    def P(texto, estilo, bullet=None):
        return Paragraph(texto, S[estilo], bulletText=bullet)

    out = []
    lineas = md.replace("\r", "").split("\n")
    i = 0
    parrafo = []

    def cerrar_parrafo():
        if parrafo:
            out.append(P(_inline(" ".join(parrafo)), "normal"))
            parrafo.clear()

    while i < len(lineas):
        ln = lineas[i].rstrip()
        s = ln.strip()
        if not s:
            cerrar_parrafo()
            i += 1
            continue
        m = re.match(r"^(#{1,4})\s+(.*)$", s)
        if m:
            cerrar_parrafo()
            nivel = len(m.group(1))
            texto = re.sub(r"\*\*", "", m.group(2)).strip()
            estilo = {1: "h1", 2: "h1", 3: "h1", 4: "h2"}[nivel]
            out.append(P(escape(texto), estilo))
            i += 1
            continue
        if s.startswith("|"):
            cerrar_parrafo()
            filas = []
            while i < len(lineas) and lineas[i].strip().startswith("|"):
                celdas = [c.strip() for c in lineas[i].strip().strip("|").split("|")]
                if not all(re.fullmatch(r":?-{2,}:?", c) for c in celdas):
                    filas.append(celdas)
                i += 1
            if filas:
                cab = [re.sub(r"\*\*", "", c) for c in filas[0]]
                n = len(cab)
                pesos = []
                for col in range(n):
                    largo = max(len(f[col]) if col < len(f) else 0 for f in filas)
                    pesos.append(min(max(largo, 4), 120) ** 0.6)
                tot = sum(pesos)
                anchos = [17.0 * p / tot for p in pesos]
                data = [[P(escape(c), "celdab") for c in cab]]
                for f in filas[1:]:
                    f = (f + [""] * n)[:n]
                    data.append([P(_inline(c.replace("<br>", " ")), "celda") for c in f])
                out.append(tabla_raw(data, anchos))
            continue
        m = re.match(r"^(-|\d+\.)\s+(.*)$", s)
        if m:
            cerrar_parrafo()
            bullet = "•" if m.group(1) == "-" else m.group(1)
            texto = m.group(2)
            i += 1
            # continuación con sangría
            while i < len(lineas) and lineas[i].startswith("  ") and lineas[i].strip() and not re.match(r"^\s*(-|\d+\.)\s", lineas[i]):
                texto += " " + lineas[i].strip()
                i += 1
            out.append(P(_inline(texto), "lista", bullet))
            continue
        parrafo.append(s)
        i += 1
    cerrar_parrafo()
    return out
