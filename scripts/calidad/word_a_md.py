"""Pasa a .md los cambios hechos a mano en los Word de docs/calidad/word/.

Es el camino inverso de md_a_word.py: si alguien edita un .docx en Word,
este script reescribe el .md con ese contenido, así el .md sigue siendo la
fuente y el próximo md_a_word.py no pisa lo editado.

Uso (desde la raíz del repo, con python-docx instalado):
    python scripts/calidad/word_a_md.py docs/calidad/word/01-introduccion-a-la-calidad.docx

Salida: docs/calidad/<mismo nombre>.md
"""

import re
import sys
from pathlib import Path

from docx import Document
from docx.oxml.ns import qn

RAIZ = Path(__file__).resolve().parents[2]
ORIGEN = RAIZ / "docs" / "calidad"
MONO = {"Consolas", "Courier New", "Cascadia Code", "Cascadia Mono"}


def fuente(run):
    rpr = run.find(qn("w:rPr"))
    if rpr is None:
        return None
    f = rpr.find(qn("w:rFonts"))
    return None if f is None else f.get(qn("w:ascii"))


def activo(run, etiqueta):
    rpr = run.find(qn("w:rPr"))
    if rpr is None:
        return False
    e = rpr.find(qn(etiqueta))
    return e is not None and e.get(qn("w:val")) not in ("0", "false")


def texto_md(p_el, sin_negrita=False):
    """Texto del párrafo con **negrita**, *cursiva* y `código`."""
    tramos = []  # (texto, negrita, cursiva, codigo)
    for run in p_el.iter(qn("w:r")):
        t = ""
        for hijo in run:
            if hijo.tag == qn("w:t"):
                t += hijo.text or ""
            elif hijo.tag == qn("w:tab"):
                t += "\t"
            elif hijo.tag == qn("w:br"):
                t += "\n"
        if not t:
            continue
        codigo = fuente(run) in MONO
        negrita = activo(run, "w:b") and not sin_negrita
        cursiva = activo(run, "w:i")
        if tramos and tramos[-1][1:] == (negrita, cursiva, codigo):
            tramos[-1] = (tramos[-1][0] + t, negrita, cursiva, codigo)
        else:
            tramos.append((t, negrita, cursiva, codigo))
    salida = ""
    for t, negrita, cursiva, codigo in tramos:
        if codigo:
            t = f"`{t}`"
        nucleo = t.strip()
        if not nucleo:
            salida += t
            continue
        ini = t[: len(t) - len(t.lstrip())]
        fin = t[len(t.rstrip()):]
        if cursiva:
            nucleo = f"*{nucleo}*"
        if negrita:
            nucleo = f"**{nucleo}**"
        salida += ini + nucleo + fin
    return salida.replace("****", "")


def sangria_cm(p_el):
    ppr = p_el.find(qn("w:pPr"))
    if ppr is None:
        return 0.0
    ind = ppr.find(qn("w:ind"))
    if ind is None:
        return 0.0
    v = ind.get(qn("w:left")) or ind.get(qn("w:start")) or "0"
    return int(v) / 567


def sombreado(p_el):
    ppr = p_el.find(qn("w:pPr"))
    if ppr is None:
        return None
    shd = ppr.find(qn("w:shd"))
    return None if shd is None else shd.get(qn("w:fill"))


def estilo(p_el, doc):
    ppr = p_el.find(qn("w:pPr"))
    if ppr is None:
        return ""
    s = ppr.find(qn("w:pStyle"))
    if s is None:
        return ""
    sid = s.get(qn("w:val"))
    for st in doc.styles:
        if st.style_id == sid:
            return st.name
    return sid


def es_codigo(p_el):
    runs = [r for r in p_el.iter(qn("w:r")) if "".join(x.text or "" for x in r.iter(qn("w:t"))).strip()]
    return bool(runs) and all(fuente(r) in MONO for r in runs)


def tabla_md(tbl):
    filas = []
    for tr in tbl.iter(qn("w:tr")):
        celdas = []
        for tc in tr.iter(qn("w:tc")):
            partes = [texto_md(p, sin_negrita=not filas).strip() for p in tc.iter(qn("w:p"))]
            celdas.append(" ".join(x for x in partes if x).replace("|", "\\|"))
        filas.append(celdas)
    if not filas:
        return []
    n = max(len(f) for f in filas)
    filas = [f + [""] * (n - len(f)) for f in filas]
    lineas = ["| " + " | ".join(filas[0]) + " |", "|" + "---|" * n]
    lineas += ["| " + " | ".join(f) + " |" for f in filas[1:]]
    return lineas


def convertir(docx: Path) -> str:
    doc = Document(docx)
    out = []
    primero = True
    en_codigo = False

    def separar():
        if out and out[-1] != "":
            out.append("")

    for el in doc.element.body.iterchildren():
        if el.tag == qn("w:tbl"):
            if en_codigo:
                out.append("```"); en_codigo = False
            separar()
            out.extend(tabla_md(el))
            out.append("")
            continue
        if el.tag != qn("w:p"):
            continue
        plano = "".join(x.text or "" for x in el.iter(qn("w:t")))
        if not plano.strip():
            continue

        if es_codigo(el) and sombreado(el):
            if not en_codigo:
                separar(); out.append("```"); en_codigo = True
            out.extend("".join(x.text or "" for x in el.iter(qn("w:t"))).split("\n"))
            continue
        if en_codigo:
            out.append("```"); out.append(""); en_codigo = False

        nombre = estilo(el, doc)
        m = re.match(r"(?:Heading|Título)\s*(\d)", nombre)
        if primero:
            separar(); out.append("# " + plano.strip()); out.append("")
            primero = False
            continue
        if m:
            separar(); out.append("#" * int(m.group(1)) + " " + texto_md(el, sin_negrita=True).strip()); out.append("")
            continue

        texto = texto_md(el)
        cita = "> " if sombreado(el) and not es_codigo(el) else ""
        item = re.match(r"^\s*(\*\*)?([•–\-]|\d+\.)(\*\*)?\t", texto)
        if item or "List" in nombre or "Lista" in nombre:
            marca = item.group(2) if item else "•"
            cuerpo = texto[item.end():] if item else texto
            nivel = 1 if marca == "–" else 0
            if marca[0].isdigit():
                linea = f"{marca} {cuerpo.strip()}"
            else:
                linea = "  " * nivel + "- " + cuerpo.strip()
                if cita:
                    nivel = 0
            if cita and marca[0].isdigit():
                pass
            if not cita and out and out[-1] != "" and not re.match(r"^\s*([-]|\d+\.) ", out[-1]):
                out.append("")
            out.append(cita + linea)
            continue
        if cita:
            if out and out[-1].startswith(">") and not re.match(r"^>\s*- ", out[-1]):
                out.append(">")
            elif out and out[-1] != "" and not out[-1].startswith(">"):
                out.append("")
            out.append(cita + texto.strip())
            continue
        # Párrafo que sigue a un ítem, con la misma sangría del texto del ítem.
        if sangria_cm(el) > 0.9 and out and re.match(r"^\s*([-]|\d+\.) ", out[-1] if out[-1] else (out[-2] if len(out) > 1 else "")):
            out.append("")
            out.append("   " + texto.strip())
            continue
        separar()
        out.append(texto.strip())
        out.append("")

    if en_codigo:
        out.append("```")
    md = "\n".join(out)
    md = re.sub(r"\n{3,}", "\n\n", md).strip() + "\n"
    return md


def main():
    for a in sys.argv[1:]:
        docx = Path(a)
        destino = ORIGEN / (docx.stem + ".md")
        destino.write_text(convertir(docx), encoding="utf-8")
        print(f"{docx.name} -> {destino.relative_to(RAIZ)}")


if __name__ == "__main__":
    main()
