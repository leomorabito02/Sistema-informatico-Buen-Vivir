"""Pasa los documentos del TP (docs/calidad/*.md) a Word (.docx).

Los .md siguen siendo la fuente: se editan ellos y después se regenera el
Word con este script. Si un documento tiene alguna tabla ancha (6 columnas
o más), el documento entero va apaisado para que las tablas se lean.

Uso (desde la raíz del repo, con python-docx instalado):
    python scripts/calidad/md_a_word.py            # todos los .md de docs/calidad
    python scripts/calidad/md_a_word.py archivo.md # uno solo

Salida: docs/calidad/word/<mismo nombre>.docx
"""

import re
import sys
from pathlib import Path

from docx import Document
from docx.enum.section import WD_ORIENT
from docx.enum.table import WD_TABLE_ALIGNMENT
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.oxml import OxmlElement
from docx.oxml.ns import qn
from docx.shared import Cm, Pt, RGBColor

RAIZ = Path(__file__).resolve().parents[2]
ORIGEN = RAIZ / "docs" / "calidad"
DESTINO = ORIGEN / "word"

# Paleta sobria: negro, grises y nada más.
TITULOS = RGBColor(0x1A, 0x1A, 0x1A)
GRIS_TEXTO = RGBColor(0x40, 0x40, 0x40)
CODIGO = RGBColor(0x33, 0x33, 0x33)
FONDO_ENCABEZADO = "EDEDED"
BORDE_TABLA = "A6A6A6"
FONDO_CITA = "F7F7F7"
FONDO_CODIGO = "F7F7F7"
COLUMNAS_APAISADO = 6
MARGEN = Cm(2)

INLINE = re.compile(r"(\*\*[^*]+\*\*|`[^`]+`|\*[^*\s][^*]*\*|\[[^\]]+\]\([^)]+\))")


def sombrear(elemento, color):
    """Pinta el fondo de una celda o de un párrafo."""
    props = elemento.get_or_add_tcPr() if hasattr(elemento, "get_or_add_tcPr") else elemento.get_or_add_pPr()
    shd = OxmlElement("w:shd")
    shd.set(qn("w:val"), "clear")
    shd.set(qn("w:color"), "auto")
    shd.set(qn("w:fill"), color)
    props.append(shd)


def borde_izquierdo(parrafo, color="A6A6A6"):
    ppr = parrafo._p.get_or_add_pPr()
    bdr = OxmlElement("w:pBdr")
    izq = OxmlElement("w:left")
    for k, v in {"w:val": "single", "w:sz": "18", "w:space": "8", "w:color": color}.items():
        izq.set(qn(k), v)
    bdr.append(izq)
    ppr.append(bdr)


def texto_con_formato(parrafo, texto, tam=None, negrita=False, color=None):
    """Agrega texto interpretando **negrita**, *cursiva*, `código` y [enlaces](url)."""
    for parte in INLINE.split(texto):
        if not parte:
            continue
        run_negrita, cursiva, codigo = negrita, False, False
        if parte.startswith("**") and parte.endswith("**"):
            if "`" in parte[2:-2]:
                texto_con_formato(parrafo, parte[2:-2], tam, True, color)
                continue
            parte, run_negrita = parte[2:-2], True
        elif parte.startswith("`") and parte.endswith("`"):
            parte, codigo = parte[1:-1], True
        elif parte.startswith("[") and "](" in parte:
            parte = parte[1:parte.index("](")]
        elif parte.startswith("*") and parte.endswith("*") and len(parte) > 2:
            parte, cursiva = parte[1:-1], True
        run = parrafo.add_run(parte)
        run.bold = run_negrita
        run.italic = cursiva
        if codigo:
            run.font.name = "Consolas"
            run.font.color.rgb = CODIGO
        if tam:
            run.font.size = Pt(tam - 0.5 if codigo else tam)
        elif codigo:
            run.font.size = Pt(10)
        if color:
            run.font.color.rgb = color


def bordes_finos(tabla):
    """Bordes de 1/2 punto en gris, en lugar del negro por defecto."""
    tblpr = tabla._tbl.tblPr
    bordes = OxmlElement("w:tblBorders")
    for lado in ("top", "left", "bottom", "right", "insideH", "insideV"):
        b = OxmlElement(f"w:{lado}")
        for k, v in {"w:val": "single", "w:sz": "4", "w:space": "0", "w:color": BORDE_TABLA}.items():
            b.set(qn(k), v)
        bordes.append(b)
    tblpr.append(bordes)


def numero_de_pagina(seccion):
    p = seccion.footer.paragraphs[0]
    p.alignment = WD_ALIGN_PARAGRAPH.CENTER
    run = p.add_run()
    for tipo, texto in (("begin", None), (None, "PAGE"), ("end", None)):
        if tipo:
            fld = OxmlElement("w:fldChar")
            fld.set(qn("w:fldCharType"), tipo)
            run._r.append(fld)
        else:
            instr = OxmlElement("w:instrText")
            instr.set(qn("xml:space"), "preserve")
            instr.text = texto
            run._r.append(instr)
    run.font.size = Pt(9)
    run.font.color.rgb = GRIS_TEXTO


def preparar_documento(apaisado=False):
    doc = Document()
    normal = doc.styles["Normal"]
    normal.font.name = "Calibri"
    normal.font.size = Pt(11)
    normal.paragraph_format.space_after = Pt(6)
    normal.paragraph_format.line_spacing = 1.15
    for nivel, tam in ((1, 20), (2, 15), (3, 12.5), (4, 11.5)):
        estilo = doc.styles[f"Heading {nivel}"]
        estilo.font.name = "Calibri"
        estilo.font.size = Pt(tam)
        estilo.font.color.rgb = TITULOS
        estilo.paragraph_format.space_before = Pt(16 if nivel <= 2 else 10)
        estilo.paragraph_format.space_after = Pt(6)
        estilo.paragraph_format.keep_with_next = True
    sec = doc.sections[0]
    sec.orientation = WD_ORIENT.LANDSCAPE if apaisado else WD_ORIENT.PORTRAIT
    sec.page_width, sec.page_height = (Cm(29.7), Cm(21)) if apaisado else (Cm(21), Cm(29.7))
    for lado in ("left_margin", "right_margin", "top_margin", "bottom_margin"):
        setattr(sec, lado, MARGEN)
    numero_de_pagina(sec)
    modo_actual(doc)
    return doc


def modo_actual(doc):
    """Marca el archivo como Word actual, para que no abra en "Modo de compatibilidad"."""
    settings = doc.settings.element
    compat = settings.find(qn("w:compat"))
    if compat is None:
        compat = OxmlElement("w:compat")
        settings.append(compat)
    for s in compat.findall(qn("w:compatSetting")):
        if s.get(qn("w:name")) == "compatibilityMode":
            compat.remove(s)
    ajuste = OxmlElement("w:compatSetting")
    for k, v in {"w:name": "compatibilityMode", "w:uri": "http://schemas.microsoft.com/office/word", "w:val": "15"}.items():
        ajuste.set(qn(k), v)
    compat.append(ajuste)


def ancho_util(doc):
    sec = doc.sections[-1]
    return sec.page_width - sec.left_margin - sec.right_margin


def celdas(linea):
    linea = linea.strip()
    if linea.startswith("|"):
        linea = linea[1:]
    if linea.endswith("|"):
        linea = linea[:-1]
    # Los "|" dentro de `código` no separan columnas.
    partes, actual, en_codigo = [], "", False
    for ch in linea:
        if ch == "`":
            en_codigo = not en_codigo
        if ch == "|" and not en_codigo:
            partes.append(actual.strip())
            actual = ""
        else:
            actual += ch
    partes.append(actual.strip())
    return partes


def agregar_tabla(doc, filas):
    encabezado, cuerpo = filas[0], filas[1:]
    n = len(encabezado)
    tam = 8.5 if n >= 7 else 9 if n >= 5 else 10

    tabla = doc.add_table(rows=1, cols=n)
    tabla.style = "Table Grid"
    bordes_finos(tabla)
    tabla.alignment = WD_TABLE_ALIGNMENT.CENTER
    tabla.autofit = False

    # Ancho de cada columna: primero lo justo para que la palabra más larga no
    # se corte, y el resto se reparte según cuánto texto tiene la columna.
    cm_por_letra = tam * 0.021
    minimos, pesos = [], []
    for i in range(n):
        textos = [re.sub(r"[*`]", "", f[i]) if i < len(f) else "" for f in filas]
        palabra = max((len(p) for t in textos for p in t.split()), default=4)
        minimos.append(Cm(min(palabra, 22) * cm_por_letra + 0.4))
        pesos.append(sum(len(t) for t in textos) + 1)
    sobrante = max(ancho_util(doc) - sum(minimos), 0)
    anchos = [int(m + sobrante * p / sum(pesos)) for m, p in zip(minimos, pesos)]

    def llenar(fila, valores, es_encabezado, par):
        # Una fila no se parte entre dos hojas.
        no_partir = OxmlElement("w:cantSplit")
        no_partir.set(qn("w:val"), "true")
        fila._tr.get_or_add_trPr().append(no_partir)
        for i, celda in enumerate(fila.cells):
            celda.width = anchos[i]
            p = celda.paragraphs[0]
            p.paragraph_format.space_before = Pt(1.5)
            p.paragraph_format.space_after = Pt(1.5)
            p.paragraph_format.line_spacing = 1.0
            # El encabezado no queda solo al pie de una hoja.
            p.paragraph_format.keep_with_next = es_encabezado
            valor = valores[i] if i < len(valores) else ""
            if es_encabezado:
                texto_con_formato(p, valor, tam, negrita=True)
                sombrear(celda._tc, FONDO_ENCABEZADO)
            else:
                texto_con_formato(p, valor, tam)

    llenar(tabla.rows[0], encabezado, True, False)
    trpr = tabla.rows[0]._tr.get_or_add_trPr()
    repetir = OxmlElement("w:tblHeader")
    repetir.set(qn("w:val"), "true")
    trpr.append(repetir)
    for k, valores in enumerate(cuerpo):
        llenar(tabla.add_row(), valores, False, k % 2 == 1)

    doc.add_paragraph()


def agregar_item(doc, marca, texto, nivel, en_cita=False):
    p = doc.add_paragraph()
    sangria = 0.6 + 0.6 * nivel + (0.4 if en_cita else 0)
    p.paragraph_format.left_indent = Cm(sangria + 0.5)
    p.paragraph_format.first_line_indent = Cm(-0.5)
    p.paragraph_format.space_after = Pt(3)
    p.add_run(marca + "\t")
    p.paragraph_format.tab_stops.add_tab_stop(Cm(sangria + 0.5))
    texto_con_formato(p, texto, color=GRIS_TEXTO if en_cita else None)
    if en_cita:
        sombrear(p._p, FONDO_CITA)
        borde_izquierdo(p)
    return p


def convertir(md: Path, salida: Path):
    lineas = md.read_text(encoding="utf-8").splitlines()
    apaisado = any(l.strip().startswith("|") and len(celdas(l)) >= COLUMNAS_APAISADO for l in lineas)
    doc = preparar_documento(apaisado)
    i = 0
    primer_titulo = True

    while i < len(lineas):
        linea = lineas[i]
        quieta = linea.strip()

        if not quieta or quieta == "---":
            i += 1
            continue

        # Bloque de código.
        if quieta.startswith("```"):
            i += 1
            codigo = []
            while i < len(lineas) and not lineas[i].strip().startswith("```"):
                codigo.append(lineas[i])
                i += 1
            i += 1
            p = doc.add_paragraph()
            sombrear(p._p, FONDO_CODIGO)
            p.paragraph_format.left_indent = Cm(0.4)
            run = p.add_run("\n".join(codigo))
            run.font.name = "Consolas"
            run.font.size = Pt(9)
            continue

        # Títulos.
        m = re.match(r"^(#{1,4})\s+(.*)", quieta)
        if m:
            nivel = len(m.group(1))
            if primer_titulo and nivel == 1:
                titulo = doc.add_paragraph()
                titulo.paragraph_format.space_after = Pt(12)
                texto_con_formato(titulo, m.group(2), 20, negrita=True, color=TITULOS)
                primer_titulo = False
            else:
                h = doc.add_heading(level=nivel)
                texto_con_formato(h, m.group(2))
            i += 1
            continue

        # Tablas.
        if quieta.startswith("|"):
            filas = []
            while i < len(lineas) and lineas[i].strip().startswith("|"):
                fila = celdas(lineas[i])
                if not all(re.fullmatch(r":?-{2,}:?", c) for c in fila if c):
                    filas.append(fila)
                i += 1
            agregar_tabla(doc, filas)
            continue

        # Citas (recuadros), que pueden tener listas adentro.
        if quieta.startswith(">"):
            bloque = []
            while i < len(lineas) and lineas[i].strip().startswith(">"):
                bloque.append(re.sub(r"^\s*>\s?", "", lineas[i]))
                i += 1
            procesar_bloque(doc, bloque, en_cita=True)
            continue

        # Listas y párrafos comunes.
        bloque = []
        while i < len(lineas):
            l = lineas[i]
            if not l.strip() and sigue_la_lista(bloque, lineas, i):
                bloque.append(l)
                i += 1
                continue
            if not l.strip() or l.strip().startswith(("|", ">", "```", "#")) or l.strip() == "---":
                break
            bloque.append(l)
            i += 1
        procesar_bloque(doc, bloque)

    salida.parent.mkdir(parents=True, exist_ok=True)
    doc.save(salida)


def sigue_la_lista(bloque, lineas, i):
    """Una línea en blanco no corta la lista si lo que sigue está sangrado."""
    if not any(re.match(r"^\s*([-*]|\d+\.)\s", l) for l in bloque):
        return False
    j = i
    while j < len(lineas) and not lineas[j].strip():
        j += 1
    return j < len(lineas) and lineas[j].startswith(" ")


def procesar_bloque(doc, lineas, en_cita=False):
    """Une líneas cortadas en párrafos e ítems de lista."""
    items = []  # (tipo, marca, nivel, texto)
    corte = False
    for l in lineas:
        if not l.strip():
            corte = True
            continue
        sangria = len(l) - len(l.lstrip())
        m_v = re.match(r"^\s*[-*]\s+(.*)", l)
        m_n = re.match(r"^\s*(\d+)\.\s+(.*)", l)
        if m_v:
            items.append(["item", "•" if sangria < 2 else "–", sangria // 2, m_v.group(1)])
        elif m_n:
            items.append(["item", m_n.group(1) + ".", sangria // 3, m_n.group(2)])
        elif items and corte and sangria > 0 and items[-1][0] != "parrafo":
            items.append(["sigue", "", items[-1][2], l.strip()])
        elif items and not (corte and sangria == 0):
            items[-1][3] += " " + l.strip()
        else:
            items.append(["parrafo", "", 0, l.strip()])
        corte = False
    for tipo, marca, nivel, texto in items:
        if tipo == "item":
            agregar_item(doc, marca, texto, nivel, en_cita)
        elif tipo == "sigue":
            p = doc.add_paragraph()
            p.paragraph_format.left_indent = Cm(1.1 + 0.6 * nivel)
            texto_con_formato(p, texto)
        else:
            p = doc.add_paragraph()
            texto_con_formato(p, texto, color=GRIS_TEXTO if en_cita else None)
            if en_cita:
                p.paragraph_format.left_indent = Cm(0.4)
                sombrear(p._p, FONDO_CITA)
                borde_izquierdo(p)


def main():
    archivos = [Path(a) for a in sys.argv[1:]] or sorted(ORIGEN.glob("*.md"))
    for md in archivos:
        salida = DESTINO / (md.stem + ".docx")
        try:
            convertir(md, salida)
        except PermissionError:
            print(f"{md.name}: no se pudo escribir {salida.name}. ¿Está abierto en Word? Cerralo y volvé a correr el script.")
            continue
        print(f"{md.name} -> {salida.relative_to(RAIZ)}")


if __name__ == "__main__":
    main()
