# -*- coding: utf-8 -*-
"""Genera docs/HISTORIAS_DE_USUARIO.md y docs/INFORME_AUDITORIA_BUEN_VIVIR.pdf
desde datos.py. Uso (desde la raíz del repo):

    python scripts/informe-auditoria/generar.py

Requiere reportlab. Usa Arial de Windows si está; si no, Helvetica.
"""
import json
import os
import sys
from collections import Counter, OrderedDict
from xml.sax.saxutils import escape

sys.path.insert(0, os.path.dirname(__file__))
import datos as D  # noqa: E402

from reportlab.lib import colors  # noqa: E402
from reportlab.lib.enums import TA_CENTER  # noqa: E402
from reportlab.lib.pagesizes import A4  # noqa: E402
from reportlab.lib.styles import ParagraphStyle, getSampleStyleSheet  # noqa: E402
from reportlab.lib.units import cm  # noqa: E402
from reportlab.pdfbase import pdfmetrics  # noqa: E402
from reportlab.pdfbase.ttfonts import TTFont  # noqa: E402
from reportlab.platypus import (  # noqa: E402
    BaseDocTemplate, Frame, KeepTogether, LongTable, PageBreak, PageTemplate,
    Paragraph, Spacer, TableStyle,
)
from reportlab.platypus.tableofcontents import TableOfContents  # noqa: E402

RAIZ = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
DOCS = os.path.join(RAIZ, "docs")

# ------------------------------------------------------------------ fuentes
FUENTE, NEGRITA = "Helvetica", "Helvetica-Bold"
try:
    pdfmetrics.registerFont(TTFont("Arial", r"C:\Windows\Fonts\arial.ttf"))
    pdfmetrics.registerFont(TTFont("Arial-Bold", r"C:\Windows\Fonts\arialbd.ttf"))
    pdfmetrics.registerFont(TTFont("Arial-Italic", r"C:\Windows\Fonts\ariali.ttf"))
    pdfmetrics.registerFont(TTFont("Arial-BoldItalic", r"C:\Windows\Fonts\arialbi.ttf"))
    from reportlab.pdfbase.pdfmetrics import registerFontFamily
    registerFontFamily("Arial", normal="Arial", bold="Arial-Bold", italic="Arial-Italic", boldItalic="Arial-BoldItalic")
    FUENTE, NEGRITA = "Arial", "Arial-Bold"
except Exception:  # pragma: no cover
    pass

VERDE = colors.HexColor("#2E6B3F")
VERDE_CLARO = colors.HexColor("#E8F2EA")
GRIS = colors.HexColor("#5F6B66")
COLOR_ESTADO = {
    "IMPLEMENTADO": colors.HexColor("#D9EFD9"),
    "PARCIAL": colors.HexColor("#FFF1C9"),
    "NO_IMPLEMENTADO": colors.HexColor("#F8D3D0"),
    "CON_ERRORES": colors.HexColor("#F6DCC4"),
    "VERIFICAR": colors.HexColor("#DCE6F5"),
}

ss = getSampleStyleSheet()
S = {
    "normal": ParagraphStyle("n", parent=ss["Normal"], fontName=FUENTE, fontSize=9.5, leading=13, spaceAfter=4),
    "celda": ParagraphStyle("c", parent=ss["Normal"], fontName=FUENTE, fontSize=7.6, leading=9.4),
    "celdab": ParagraphStyle("cb", parent=ss["Normal"], fontName=NEGRITA, fontSize=7.6, leading=9.4, textColor=colors.white),
    "h1": ParagraphStyle("h1", parent=ss["Heading1"], fontName=NEGRITA, fontSize=16, leading=20, textColor=VERDE, spaceBefore=6, spaceAfter=8),
    "h2": ParagraphStyle("h2", parent=ss["Heading2"], fontName=NEGRITA, fontSize=12.5, leading=16, textColor=VERDE, spaceBefore=10, spaceAfter=5),
    "h3": ParagraphStyle("h3", parent=ss["Heading3"], fontName=NEGRITA, fontSize=10.5, leading=13, spaceBefore=6, spaceAfter=3),
    "lista": ParagraphStyle("l", parent=ss["Normal"], fontName=FUENTE, fontSize=9.5, leading=13, leftIndent=12, bulletIndent=2, spaceAfter=2),
    "portada1": ParagraphStyle("p1", fontName=NEGRITA, fontSize=30, leading=36, alignment=TA_CENTER, textColor=VERDE),
    "portada2": ParagraphStyle("p2", fontName=FUENTE, fontSize=17, leading=22, alignment=TA_CENTER),
    "portada3": ParagraphStyle("p3", fontName=FUENTE, fontSize=11, leading=15, alignment=TA_CENTER, textColor=GRIS),
    "toc1": ParagraphStyle("t1", fontName=NEGRITA, fontSize=10, leading=14, leftIndent=10),
    "toc2": ParagraphStyle("t2", fontName=FUENTE, fontSize=9, leading=12, leftIndent=26),
}


def P(texto, estilo="normal"):
    return Paragraph(texto, S[estilo])


def E(texto):
    return escape(str(texto))


def lista(items):
    return [Paragraph(E(i), S["lista"], bulletText="•") for i in items]


def tabla(cab, filas, anchos, colores_fila=None):
    data = [[Paragraph(E(c), S["celdab"]) for c in cab]]
    for f in filas:
        data.append([Paragraph(E(v) if not str(v).startswith("<b>") else str(v), S["celda"]) for v in f])
    t = LongTable(data, colWidths=[a * cm for a in anchos], repeatRows=1)
    st = [
        ("BACKGROUND", (0, 0), (-1, 0), VERDE),
        ("VALIGN", (0, 0), (-1, -1), "TOP"),
        ("GRID", (0, 0), (-1, -1), 0.3, colors.HexColor("#B8C4BD")),
        ("LEFTPADDING", (0, 0), (-1, -1), 3), ("RIGHTPADDING", (0, 0), (-1, -1), 3),
        ("TOPPADDING", (0, 0), (-1, -1), 2), ("BOTTOMPADDING", (0, 0), (-1, -1), 2),
    ]
    for i in range(1, len(data)):
        if colores_fila and colores_fila[i - 1] is not None:
            st.append(("BACKGROUND", (0, i), (-1, i), colores_fila[i - 1]))
        elif i % 2 == 0:
            st.append(("BACKGROUND", (0, i), (-1, i), colors.HexColor("#F6F8F7")))
    t.setStyle(TableStyle(st))
    return t


class Doc(BaseDocTemplate):
    def __init__(self, path):
        super().__init__(path, pagesize=A4, leftMargin=2 * cm, rightMargin=2 * cm, topMargin=2 * cm, bottomMargin=2 * cm,
                         title="Buen Vivir App - Informe de auditoría técnica y funcional", author="Equipo Buen Vivir",
                         subject="Auditoría " + D.FECHA)
        marco = Frame(self.leftMargin, self.bottomMargin, self.width, self.height, id="f")
        self.addPageTemplates([
            PageTemplate(id="portada", frames=[marco], onPage=lambda c, d: None),
            PageTemplate(id="cuerpo", frames=[marco], onPage=self.pie),
        ])

    def pie(self, c, d):
        c.saveState()
        c.setFont(FUENTE, 8)
        c.setFillColor(GRIS)
        c.drawString(2 * cm, 1.2 * cm, "Buen Vivir App · Informe de auditoría técnica y funcional · " + D.FECHA)
        c.drawRightString(A4[0] - 2 * cm, 1.2 * cm, "Página %d" % d.page)
        c.setStrokeColor(VERDE)
        c.line(2 * cm, A4[1] - 1.4 * cm, A4[0] - 2 * cm, A4[1] - 1.4 * cm)
        c.restoreState()

    def afterFlowable(self, f):
        if isinstance(f, Paragraph) and f.style.name in ("h1", "h2"):
            nivel = 0 if f.style.name == "h1" else 1
            texto = f.getPlainText()
            clave = "s%d" % id(f)
            self.canv.bookmarkPage(clave)
            self.canv.addOutlineEntry(texto, clave, level=nivel, closed=nivel > 0)
            self.notify("TOCEntry", (nivel, texto, self.page, clave))


# ---------------------------------------------------------------- métricas
def conteos():
    c = Counter(h["estado"] for h in D.HUS)
    return c


def cargar_evidencias():
    out = {}
    for nombre in ("k6-resultado-carga.json", "sonda-seguridad-resultado.json"):
        p = os.path.join(DOCS, "evidencias", nombre)
        if os.path.exists(p):
            with open(p, encoding="utf-8") as f:
                out[nombre] = json.load(f)
    return out


# ------------------------------------------------------------------ markdown
def generar_md():
    c = conteos()
    l = ["# Historias de usuario: estado de implementación", "",
         f"Auditoría del {D.FECHA} ({D.VERSION}). Generado por `scripts/informe-auditoria/generar.py` desde `datos.py`: "
         "no editar a mano. La fuente de las HU es `docs/referencias/HU DE BUEN VIVIR APP.pdf`.", "",
         "## Resumen", "", "| Estado | Cantidad |", "|---|---|"]
    for k, v in D.ESTADOS.items():
        l.append(f"| {v} | {c.get(k, 0)} |")
    l += [f"| **Total** | **{len(D.HUS)}** |", "",
          "El estado es el final, después de las correcciones de la auditoría. Detalle completo en el informe PDF.", "",
          "## Tabla general", "", "| HU | Módulo | Historia | Estado | Problema o pendiente principal |", "|---|---|---|---|---|"]
    for h in D.HUS:
        nota = (h["errores"] or h["pendiente"] or "-").replace("|", "/")
        l.append(f"| {h['id']} | {h['modulo']} | {h['titulo']} | {D.ESTADOS[h['estado']]} | {nota} |")
    l += ["", "## Detalle por historia", ""]
    for h in D.HUS:
        l += [f"### {h['id']}: {h['titulo']}", "",
              f"- **Actor:** {h['actor']}", f"- **Estado:** {D.ESTADOS[h['estado']]}",
              f"- **Frontend:** {h['front']}", f"- **Backend:** {h['back']}",
              f"- **Endpoints:** {h['endpoints']}", f"- **Persistencia:** {h['persistencia']}",
              f"- **Tests existentes:** {h['tests'] or '-'}", f"- **Tests que faltan:** {h['faltan'] or '-'}",
              f"- **Problemas detectados:** {h['errores'] or 'Ninguno'}", f"- **Correcciones (11/11/2026):** {h['correcciones'] or '-'}",
              f"- **Pendiente:** {h['pendiente'] or '-'}", ""]
    with open(os.path.join(DOCS, "HISTORIAS_DE_USUARIO.md"), "w", encoding="utf-8", newline="\n") as f:
        f.write("\n".join(l))


def generar_pendientes():
    l = ["# Pendientes", "",
         f"Estado al {D.FECHA}. Generado por `scripts/informe-auditoria/generar.py` desde `datos.py`: no editar a mano.", "",
         "| Prioridad | Pendiente | HU | Módulo | Trabajo necesario | Dependencias | Criterio de aceptación |", "|---|---|---|---|---|---|---|"]
    orden = {"Alta": 0, "Media": 1, "Baja": 2}
    for d_, hu, mod, _arch, trabajo, dep, prio, crit in sorted(D.PENDIENTES, key=lambda x: orden[x[6]]):
        l.append(f"| {prio} | {d_} | {hu} | {mod} | {trabajo} | {dep} | {crit} |")
    l += ["", "## Historias no completas", "", "| HU | Estado | Qué falta |", "|---|---|---|"]
    for h in D.HUS:
        if h["estado"] != "IMPLEMENTADO":
            l.append(f"| {h['id']} | {D.ESTADOS[h['estado']]} | {(h['pendiente'] or h['errores']).replace('|', '/')} |")
    l += ["", "## Recomendaciones (no son requisitos de las HU)", ""]
    l += [f"- **{a}:** {r}" for a, r in D.RECOMENDACIONES]
    with open(os.path.join(DOCS, "PENDIENTES.md"), "w", encoding="utf-8", newline="\n") as f:
        f.write("\n".join(l) + "\n")


# ---------------------------------------------------------------------- pdf
def generar_pdf():
    ev = cargar_evidencias()
    c = conteos()
    story = []

    # 1. Portada
    story += [Spacer(1, 5 * cm), P("Buen Vivir App", "portada1"), Spacer(1, 0.6 * cm),
              P("Informe de auditoría técnica y funcional", "portada2"), Spacer(1, 1.2 * cm),
              P("Centro de Ecología Integral · Centro Barrial “La Otra Orilla”, Alta Gracia", "portada3"),
              Spacer(1, 0.4 * cm), P("Fecha de realización: " + D.FECHA, "portada3"),
              P("Versión auditada: " + E(D.VERSION), "portada3"), Spacer(1, 6 * cm),
              P("Documento técnico. No contiene secretos, tokens ni credenciales.", "portada3")]
    from reportlab.platypus.doctemplate import NextPageTemplate
    story += [NextPageTemplate("cuerpo"), PageBreak()]

    # Índice
    toc = TableOfContents()
    toc.levelStyles = [S["toc1"], S["toc2"]]
    story += [P("Índice", "h3"), toc, PageBreak()]

    # 2. Introducción
    story += [P("1. Introducción", "h1"),
              P("<b>Descripción del proyecto.</b> Buen Vivir App es la aplicación web progresiva (PWA) de gestión interna del Centro de "
                "Ecología Integral. Reúne en un solo sistema la gestión de usuarios y roles, la formación (cursos, actividades, asistencia "
                "y habilidades), el acompañamiento personal de participantes (Promoción Humana), la inserción laboral (ofertas, postulaciones "
                "y CV), las campañas de concientización, los indicadores del centro y las notificaciones."),
              P("<b>Objetivo de la auditoría.</b> Verificar contra las historias de usuario que cada funcionalidad esté implementada de punta a "
                "punta (frontend, API, lógica, base de datos y respuesta), detectar y corregir errores, revisar y completar los tests, "
                "investigar cinco problemas reportados en pruebas manuales, ordenar el repositorio y documentar el estado real y lo pendiente."),
              P("<b>Alcance.</b> Los nueve servicios Go (auth, education, human-promotion, campaigns, dashboard, notifications, "
                "labor-integration, rxresume y la librería pkg), el frontend Angular, el API Gateway (nginx local y OpenAPI de GCP), la "
                "infraestructura de Pub/Sub y el pipeline de CI/CD. Se analizaron las 64 historias de usuario del documento "
                "“HU DE BUEN VIVIR APP”."),
              P("<b>Metodología.</b>")] + lista([
                "Lectura completa de las HU y de la documentación existente (CLAUDE.md, README, docs/, ADR).",
                "Revisión del código por HU: componente y guard del front, ruta del gateway, middleware de rol, servicio, repositorio y colección o tabla.",
                "Ejecución de todas las suites de tests (Go con emuladores de Firestore y Pub/Sub; Karma con ChromeHeadless).",
                "Pruebas contra el stack Docker local: carga con k6 y una sonda de seguridad propia, sin escrituras sobre datos reales.",
                "Corrección de los problemas con su test, y nueva ejecución de las suites.",
                "Clasificación de cada HU en: Implementado, Parcialmente implementado, No implementado, Implementado con errores o Pendiente de verificación.",
            ]) + [P("<b>Fuentes consultadas.</b> HU DE BUEN VIVIR APP.pdf; DER y explicación del DER; CLAUDE.md raíz y de education; "
                    "docs/eventos, docs/adr, docs/der_educacion.md, docs/ia-habilidades.md; código fuente; .github/workflows/pipeline.yml; "
                    "docker-compose.yml; api-gateway/nginx.conf y openapi.yaml; resultados de tests de esta auditoría.")]

    # 3. Arquitectura
    story += [P("2. Arquitectura general", "h1"),
              P("Microservicios en Go 1.26 desplegados en Google Cloud Run (escalan a cero), detrás de Google Cloud API Gateway en la nube y de "
                "nginx en local. Comunicación REST hacia afuera, HTTP/gRPC interno para consultas síncronas y Pub/Sub para eventos asíncronos con "
                "patrón outbox. Frontend Angular 20 con PrimeNG, PWA, publicado en Cloudflare Pages."),
              tabla(["Componente", "Tecnología", "Responsabilidad", "Datos"], [
                  ("frontend", "Angular 20, PrimeNG, Firebase (Auth y FCM)", "PWA para todos los roles y landing pública", "-"),
                  ("api-gateway", "nginx (local) / Google API Gateway (OpenAPI 2.0)", "Ruteo, límite de tasa (local), validación de token para auth/labor/rxresume", "-"),
                  ("auth", "Go, chi, GORM, gRPC", "Usuarios, roles, perfiles, login, JWT, 2FA, Google, recuperación", "PostgreSQL (Neon)"),
                  ("education", "Go, chi", "Cursos, áreas, materiales, inscripciones, actividades, entregas, asistencia, habilidades IA, comentarios, biblioteca, auditoría", "Contentful + Firestore + Cloud Storage"),
                  ("human-promotion", "Go, chi", "Fichas de proceso personal (cifradas), hitos, línea de tiempo, PDF, auditoría", "Firestore"),
                  ("campaigns", "Go, chi", "Campañas de concientización", "Firestore"),
                  ("labor-integration", "Go, chi", "Ofertas laborales y postulaciones", "Firestore"),
                  ("rxresume / cv-renderer", "Go + Reactive Resume (Node)", "Generación de CV", "PostgreSQL"),
                  ("dashboard", "Go", "Indicadores, gráficas, filtros, reportes PDF", "Agrega de otros servicios; Firestore + bucket para reportes"),
                  ("notifications", "Go, GORM, gRPC", "Notificaciones in-app, email (Resend) y push (FCM), preferencias", "PostgreSQL"),
                  ("Pub/Sub", "Google Pub/Sub (emulador en local)", "Eventos de dominio con outbox, DLQ y OIDC", "-"),
              ], [3, 4, 6.5, 3.5]),
              Spacer(1, 6),
              P("Cada servicio sigue capas controller → service → repository, con interfaces para testear con dobles. Los eventos viajan "
                "con un formato único (envelope con event_id, event_type, version, correlation_id, actor_id y payload sin datos personales).")]

    # 4. Resumen ejecutivo
    filas = [(D.ESTADOS[k], c.get(k, 0)) for k in D.ESTADOS]
    story += [PageBreak(), P("3. Resumen ejecutivo", "h1"),
              P(f"Se analizaron <b>{len(D.HUS)} historias de usuario</b> (HU-01 a HU-62, más HU-02bis y HU-19bis). Los números surgen del análisis "
                "de esta auditoría y reflejan el estado después de las correcciones."),
              tabla(["Estado", "Cantidad"], filas + [("Total", len(D.HUS))], [8, 3],
                    [COLOR_ESTADO[k] for k in D.ESTADOS] + [None]),
              Spacer(1, 8), P("<b>Principales problemas encontrados</b>")] + lista([
                "Crítico (corregido en código): Inserción Laboral aceptaba la identidad que mandaba el cliente en headers; sin sesión se podían aprobar o borrar ofertas.",
                "Alto (corregido): editar y borrar ofertas sin control de rol ni dueño; perfiles creados a nombre de otras cuentas; pendientes visibles a cualquier rol.",
                "Problema 1 (corregido): no se podía crear un curso sin materiales (formulario, backend y Contentful).",
                "Problema 3 (corregido en código y en el emulador): eventos que no llegaban, enlaces rotos, aviso al creador que nunca salía, duplicados de campañas.",
                "Problema 5 (implementado): la eliminación de campañas no existía.",
                "Pendientes funcionales: bloqueo por intentos de login, auditoría en auth, notificación de ofertas compatibles, panel RBAC, observabilidad propia.",
                "Entorno: el stack Docker local corre contenedores del 29/10, con configuración y secretos distintos a los actuales.",
            ]) + [P(f"<b>Estado general.</b> El sistema está funcionalmente avanzado: {c.get('IMPLEMENTADO', 0)} de {len(D.HUS)} HU quedan implementadas y no hay errores abiertos "
                    "conocidos en las implementadas. Lo que falta se concentra en Autenticación (auditoría, bloqueo, sesiones), en la verificación "
                    "de punta a punta de notificaciones y emails en testing, y en algunas HU de Inserción Laboral. Todas las suites de tests pasan "
                    "(9 módulos Go y 694 specs de frontend).")]

    # 5. HU
    story += [PageBreak(), P("4. Auditoría de historias de usuario", "h1"), P("Tabla general", "h2")]
    story.append(tabla(["HU", "Módulo", "Historia", "Estado", "Nota principal"],
                       [(h["id"], h["modulo"], h["titulo"], D.ESTADOS[h["estado"]], h["errores"] or h["pendiente"] or "-") for h in D.HUS],
                       [1.5, 2.6, 4.4, 2.4, 6.1], [COLOR_ESTADO[h["estado"]] for h in D.HUS]))
    modulos = OrderedDict()
    for h in D.HUS:
        modulos.setdefault(h["modulo"], []).append(h)
    for mod, hus in modulos.items():
        story.append(P(mod, "h2"))
        for h in hus:
            filas = [("Actor", h["actor"]), ("Estado", D.ESTADOS[h["estado"]]), ("Frontend", h["front"]), ("Backend", h["back"]),
                     ("Endpoints", h["endpoints"]), ("Persistencia", h["persistencia"]), ("Tests existentes", h["tests"] or "-"),
                     ("Tests que faltan", h["faltan"] or "-"), ("Errores detectados", h["errores"] or "Ninguno"),
                     ("Correcciones realizadas", h["correcciones"] or "-"), ("Pendientes", h["pendiente"] or "-")]
            t = tabla(["Campo", "Detalle"], filas, [3.6, 13.4], [COLOR_ESTADO[h["estado"]] if i == 1 else None for i in range(len(filas))])
            story.append(KeepTogether([P(f"{h['id']}: {E(h['titulo'])}", "h3"), t]))

    # 6. Backend
    story += [PageBreak(), P("5. Auditoría del backend", "h1"),
              P("Estado de los servicios", "h2"),
              tabla(["Servicio", "Build", "Tests", "Cobertura", "Observaciones"], [
                  ("auth", "OK", "11/11 paquetes", "83,6 %", "Sin auditoría; sin bloqueo de login; permisos de alta de perfiles corregidos"),
                  ("education", "OK", "15/15", "90,2 %", "Problema 1 corregido; endpoint de progreso nuevo"),
                  ("human-promotion", "OK", "11/11", "75,8 %", "ficha.actualizada con creado_por"),
                  ("campaigns", "OK", "8/8", "90,2 %", "Baja lógica nueva; aviso solo al publicar"),
                  ("dashboard", "OK", "9/9", "85,8 %", "-"),
                  ("notifications", "OK", "14/14", "63,2 %", "Enlaces, huérfanas y verificador en producción corregidos"),
                  ("labor-integration", "OK", "7/7", "59,3 %", "Vulnerabilidad de identidad y permisos de ofertas corregidos"),
                  ("rxresume", "OK", "2/2", "38,4 %", "Cobertura baja"),
                  ("pkg", "OK", "2/2", "94,0 %", "Librería de eventos y outbox"),
              ], [3.2, 1.3, 2.4, 2, 8.1]),
              P("Arquitectura y organización", "h2")] + lista([
                "Capas controller/service/repository consistentes; interfaces mockeables; errores tipados (ValidationError, ErrNotFound) traducidos a 400/404 sin exponer detalles internos.",
                "Variables de entorno validadas al arrancar en education (protege contra mezclar testing y producción de Contentful) y JWT_SECRET obligatorio.",
                "Inconsistencias: labor-integration usa http.Error en texto plano mientras el resto responde JSON; labor y auth no tienen auditoría; criterios de rol distintos entre servicios.",
            ]) + [P("Endpoints", "h2")] + lista([
                "Se revisaron las rutas de los ocho servicios contra las HU (anexo). Se agregaron DELETE /api/v1/campaigns/{id} y GET /api/cursos/{id}/progreso, también en openapi.yaml.",
                "El docente puede crear cursos en borrador (POST /api/docente/cursos): la HU-06 no lo permite. Se documenta, no se elimina (funcionalidad existente; decisión del equipo).",
            ]) + [P("Base de datos", "h2")] + lista([
                "PostgreSQL (Neon) para auth, notifications y rxresume con GORM y soft delete (deleted_at) en perfiles.",
                "Firestore para education, PH, campaigns, labor y dashboard; ids determinísticos donde hay unicidad (entregas, asistencias); transacciones para cambio + evento + auditoría.",
                "Contentful para cursos, áreas y materiales. En esta auditoría se cambió material a opcional en testing y master.",
                "Riesgo: el stack local y testing comparten la misma base Firestore (default) y Neon: cualquier prueba con escritura en local afecta testing.",
            ]) + [P("Autenticación, autorización y seguridad", "h2")] + lista([
                "JWT HS256 con type=access y vencimiento validado en cada servicio; la sonda rechazó alg=none, otra clave, vencidos y de refresh (401).",
                "RBAC por RequireRole más controles por dueño en servicios; 22 casos por rol de la sonda dieron 403 como corresponde.",
                "Corregido: identidad por headers en labor; edición y borrado de ofertas; alta de perfiles ajenos; baja de perfil por el participante; verificador de push en producción; puerto de labor expuesto.",
                "Abierto: el límite por IP se puede esquivar con X-Forwarded-For (120/120 pedidos pasaron rotando el header); sin bloqueo de login; sin auditoría de accesos denegados.",
                "Sin secretos en el repositorio: se revisaron las capturas de despliegue (solo nombres de variables y rutas a Secret Manager). Los .env están en .gitignore.",
                "Headers de seguridad presentes en education y PH (CSP, HSTS, X-Frame-Options, nosniff); el gateway nginx local no los agrega y muestra su versión.",
            ]) + [P("Integración entre servicios", "h2")] + lista([
                "education y dashboard consultan auth por HTTP reenviando el JWT del usuario; notifications y labor usan gRPC o HTTP con ID token en Cloud Run.",
                "Pub/Sub: el emulador local quedaba vacío tras un reinicio (corregido). En el CD no hay PUBSUB_PUSH_AUDIENCE ni PUBSUB_PUSH_SERVICE_ACCOUNTS: no se pudo verificar la entrega de eventos en testing.",
                "Fallas aisladas: con el resolver de nginx, si un servicio no está, solo fallan sus rutas (502), no todo el gateway.",
            ])

    # 7. Frontend
    story += [P("6. Auditoría del frontend", "h1")] + lista([
        "Pantallas: existen todas las de las HU salvo el panel RBAC (HU-06) y la notificación de ofertas compatibles (HU-44).",
        "Datos: las pantallas consumen el backend real; no se encontraron datos ficticios salvo el impacto de la landing (KV del worker con valores por defecto si falla).",
        "Navegación y permisos: menú por rol en el header y guards por ruta. Sin guard: usuarios, seguimiento, cv y ofertas (el backend responde igual, pero la pantalla carga).",
        "Corregido: formulario de curso (materiales opcionales y subida en el mismo diálogo); botón Eliminar en campañas con confirmación; contador de no leídas en la campana y refresco de notificaciones; Seguimiento por rol con nombres.",
        "Manejo de errores: interceptor global, pantallas 404/500, mensajes del backend en toasts. Un 429 del gateway se muestra como error genérico.",
        "Tests: 85 archivos de specs, 694 casos, todos OK; 71,45 % de sentencias (umbral 70 %).",
        "Cypress: un spec sin ejecutar (no instalado).",
    ])

    # 8. Roles
    story += [P("7. Matriz de roles y permisos", "h1"),
              tabla(["Rol", "Módulos", "Acciones permitidas", "Restringidas", "Estado / problemas"], D.ROLES, [2.3, 2.4, 5.6, 3.2, 3.5]),
              Spacer(1, 6), P("La verificación del backend se hizo con la sonda de seguridad (anexo) y con los tests de rutas de cada servicio.")]

    # 9. Problemas
    story += [PageBreak(), P("8. Problemas detectados durante las pruebas manuales", "h1")]
    for pr in D.PROBLEMAS:
        story.append(P(f"Problema {pr['n']}. {E(pr['titulo'])}", "h2"))
        story.append(tabla(["Campo", "Detalle"], [
            ("Descripción", pr["descripcion"]), ("Comportamiento esperado", pr["esperado"]), ("Comportamiento encontrado", pr["encontrado"]),
            ("Causa", pr["causa"]), ("Archivos afectados", pr["archivos"]), ("Corrección realizada", pr["correccion"]),
            ("Tests ejecutados", pr["tests"]), ("Resultado", pr["resultado"]), ("Pendientes", pr["pendiente"])], [3.6, 13.4]))

    # 10. Inventario de tests
    story += [PageBreak(), P("9. Inventario de tests existentes", "h1"),
              tabla(["Módulo", "Tipo", "Ubicación", "Herramienta", "Arch.", "Tests", "Resultado", "Cobertura", "Obs."],
                    D.TESTS_INVENTARIO, [2.2, 2.4, 2.8, 1.9, 0.9, 1, 2, 1.6, 2.2]),
              Spacer(1, 6)] + lista([
                "Existentes: todos los de la tabla. Ejecutados: las 9 suites Go, Karma, la carga k6 local, la carga en proceso y la sonda (además, los nuevos).",
                "Pasan: 79/79 paquetes Go y 694/694 specs. Fallan: ninguno. No pudieron ejecutarse: Cypress (no instalado), k6 y ZAP del CD (requieren GCP), mutación (deshabilitada).",
                "Advertencia: correr los tests de education reescribe reports/database y reports/pact; se restauraron para no pisar cambios previos.",
            ])

    # 11. Tests faltantes
    story += [P("10. Tests faltantes", "h1"),
              tabla(["Grupo", "Qué probar", "HU / origen", "Módulo", "Prioridad"], D.TESTS_FALTANTES, [2.2, 7.6, 2.8, 2.6, 1.8])]

    # 12. Resultados testing y seguridad
    k6 = ev.get("k6-resultado-carga.json", {}).get("metrics", {})
    sonda = ev.get("sonda-seguridad-resultado.json", {})

    def met(nombre, campo):
        v = k6.get(nombre, {}).get(campo)
        return f"{v:.1f}" if isinstance(v, float) else (str(v) if v is not None else "-")

    story += [PageBreak(), P("11. Resultados de testing y seguridad", "h1"), P("Carga con k6 (stack Docker local)", "h2"),
              P("Herramienta: k6 0.54.0 en Docker (tests/carga/k6_carga_local.js). Configuración: escenario <b>gateway</b> a 8 pedidos/s "
                "(debajo del límite de nginx), escenario <b>servicios</b> directo a education y human-promotion con rampa a 50 usuarios virtuales "
                "(solo rutas sin base: salud y rechazos 401) y escenario <b>público</b> a 3/s contra la biblioteca (lectura real de Firestore)."),
              tabla(["Métrica", "Valor"], [
                  ("Pedidos totales", met("http_reqs", "count")), ("Pedidos por segundo", met("http_reqs", "rate")),
                  ("Errores 5xx", "0 de " + met("http_reqs", "count")),
                  ("p95 servicios (ms)", met("http_req_duration{scenario:servicios}", "p(95)")),
                  ("p95 gateway a 8/s (ms)", met("http_req_duration{scenario:gateway}", "p(95)")),
                  ("p95 público biblioteca (ms)", met("http_req_duration{scenario:publico}", "p(95)")),
                  ("Usuarios virtuales máximos", met("vus_max", "value")),
              ], [8, 6]),
              Spacer(1, 4)] + lista([
                "Primera corrida (sin respetar el límite): 68 % de respuestas 503 en 0 ms. No era una caída: era el límite de 10/s por IP de nginx. Se cambió a 429.",
                "Con 50 usuarios directos, education y PH respondieron 429 al superar 100 pedidos/min por IP, sin errores 5xx ni caídas.",
                "El límite por IP se esquivó rotando X-Forwarded-For: 120/120 pedidos pasaron (con la misma IP: 77 OK y 43 rechazados).",
            ]) + [P("Carga en proceso con volumen de datos (Campañas)", "h2"),
                  tabla(["Campañas cargadas", "Clientes", "Pedidos", "p50", "p95", "Máximo", "Errores"], [
                      ("500", "50", "1000", "134 ms", "485 ms", "1,02 s", "0"),
                      ("2000", "50", "1000", "3,3 s", "8,8 s", "9,6 s", "0"),
                  ], [3, 2, 2, 2, 2, 2, 2]),
                  P("Servicio real con emulador de Firestore (los tiempos absolutos son pesimistas). Conclusión: sin paginación, el listado no escala por encima de unos miles de campañas."),
                  P("Sonda de seguridad", "h2"),
                  P(f"Herramienta: programa Go propio (tests/seguridad/sonda). Casos: {sonda.get('resumen', {}).get('casos', '-')}; "
                    f"OK: {sonda.get('resumen', {}).get('ok', '-')}; fallas: {sonda.get('resumen', {}).get('falla', '-')}. Solo pedidos que no modifican datos.")]
    if sonda.get("casos"):
        story.append(tabla(["Grupo", "Caso", "Esperado", "Obtenido", "Resultado"],
                           [(x["grupo"], x["nombre"], ",".join(map(str, x["espera"])), x["obtuvo"], x["resultado"]) for x in sonda["casos"]],
                           [2.2, 9.3, 1.8, 1.7, 2]))
    story += [Spacer(1, 4), P("<b>Limitaciones.</b> El stack local comparte bases con testing, por lo que no hubo pruebas de carga con escritura ni con tokens de "
                              "usuarios reales. Los contenedores corren código del 29/10: las correcciones se verificaron con tests, no en el stack. ZAP no se ejecutó.")]

    # 13. Pendientes
    story += [PageBreak(), P("12. Pendientes de implementación", "h1"),
              tabla(["Descripción", "HU", "Módulo", "Archivos", "Trabajo necesario", "Dependencias", "Prior.", "Criterio de aceptación"],
                    D.PENDIENTES, [2.4, 1.5, 1.7, 2, 3, 1.7, 1.1, 3.6])]
    hu_pend = [h for h in D.HUS if h["estado"] != "IMPLEMENTADO"]
    story += [P("Historias no completas", "h2"),
              tabla(["HU", "Estado", "Qué falta"], [(h["id"], D.ESTADOS[h["estado"]], h["pendiente"] or h["errores"]) for h in hu_pend],
                    [1.6, 3.2, 12.2], [COLOR_ESTADO[h["estado"]] for h in hu_pend])]

    # 14. Recomendaciones
    story += [P("13. Recomendaciones técnicas", "h1"),
              P("Son mejoras sugeridas, no requerimientos de las HU (esos están en Pendientes)."),
              tabla(["Área", "Recomendación"], D.RECOMENDACIONES, [3, 14])]

    # 15. Conclusión
    story += [P("14. Conclusión", "h1"),
              P("Buen Vivir App cubre la mayor parte de lo pedido: Educación, Promoción Humana, Campañas y Dashboard están prácticamente completos "
                "y con buena cobertura de tests. La auditoría encontró y corrigió problemas de seguridad serios en Inserción Laboral y en Autenticación, "
                "resolvió los cinco problemas reportados y dejó pruebas de carga y de seguridad reutilizables."),
              P("Para considerar el proyecto terminado falta: (1) reconstruir el stack Docker y desplegar los arreglos; (2) completar Autenticación "
                "(auditoría, bloqueo por intentos, invalidación de sesión); (3) verificar de punta a punta notificaciones, emails y push en testing; "
                "(4) las HU de Inserción Laboral pendientes (validaciones, ofertas compatibles); (5) panel RBAC y observabilidad; (6) E2E por rol.")]

    # 16. Anexo
    story += [PageBreak(), P("15. Anexo técnico", "h1"), P("Estructura del repositorio", "h2"),
              tabla(["Carpeta", "Contenido"], [
                  ("auth/, education/, human-promotion/, campaigns/, dashboard/, notifications/, labor-integration/, rxresume/", "Servicios Go (cmd/server, internal/{controllers,services,repositories,models,routes,middlewares,events,outbox})"),
                  ("pkg/", "Librería fuente de eventos y outbox (copiada a cada servicio con scripts/sync-eventlib.sh)"),
                  ("frontend/", "PWA Angular"), ("api-gateway/", "nginx.conf (local) y openapi.yaml (GCP)"),
                  ("infra/pubsub/", "setup.sh (GCP) y emulator.sh (local)"), ("cloudflare-worker/", "Worker de métricas de impacto (KV)"),
                  ("docs/", "Documentación (este informe, HU, arquitectura, ADR, eventos, evidencias, referencias)"),
                  ("tests/", "Carga (k6), sonda de seguridad, generadores de reportes del TP de Calidad"),
                  ("scripts/", "sync-eventlib, informe-auditoria, historico-openapi, calidad"),
                  (".github/workflows/pipeline.yml", "CI/CD: tests, imágenes, Cloud Run, smoke, ZAP, k6, rollback"),
              ], [6, 11]),
              P("Comandos de ejecución", "h2"),
              tabla(["Acción", "Comando"], [
                  ("Levantar todo en local", "docker compose up -d --build"),
                  ("Tests Go (con emuladores)", "FIRESTORE_EMULATOR_HOST=127.0.0.1:8390 PUBSUB_EMULATOR_HOST=localhost:8086 go test ./...  (en cada servicio)"),
                  ("Tests frontend", "cd frontend && npx ng test --watch=false --browsers=ChromeHeadless --code-coverage"),
                  ("Carga local", "docker run --rm -v \"$PWD/tests/carga:/scripts\" grafana/k6:0.54.0 run /scripts/k6_carga_local.js"),
                  ("Sonda de seguridad", "JWT_SECRET=... go run ./tests/seguridad/sonda > docs/evidencias/sonda-seguridad-resultado.json"),
                  ("Carga Campañas en proceso", "cd campaigns && CARGA=1 FIRESTORE_EMULATOR_HOST=... go test ./internal/routes -run Carga -v"),
                  ("Regenerar este informe", "python scripts/informe-auditoria/generar.py"),
              ], [4.5, 12.5]),
              P("Endpoints principales por servicio", "h2"),
              tabla(["Servicio", "Prefijo", "Rutas"], [
                  ("auth", "/api/v1/auth, /users, /roles, /permissions, /companies, /teachers, /participants, /volunteers, /profile, /directory", "login, register, refresh, 2FA, recuperación, CRUD de perfiles, directorio"),
                  ("education", "/api/cursos, /areas-formativas, /materiales, /docente, /participante, /actividades, /asistencia-general, /habilidades, /comentarios, /biblioteca, /contenidos-docentes, /auditoria", "ver education/CLAUDE.md"),
                  ("human-promotion", "/api/v1/human-promotion", "fichas, hitos, actividad, pdf, participantes, indicadores, auditoria, mi-linea-de-tiempo"),
                  ("campaigns", "/api/v1/campaigns", "publicas, CRUD, publicar, DELETE (baja lógica)"),
                  ("labor-integration", "/api/v1/labor-integration", "offers (CRUD, approve, reject, applications), applications (me, CRUD, status)"),
                  ("dashboard", "/api/v1/dashboard", "resumen, reportes"),
                  ("notifications", "/api/v1/notifications", "listado, detalle, read, read-all, unread-count, preferencias; /eventos/* (Pub/Sub)"),
                  ("rxresume", "/api/v1/rxresume", "CV y versión pública"),
              ], [2.6, 7.4, 7]),
              P("Archivos importantes", "h2")] + lista([
                "docs/HISTORIAS_DE_USUARIO.md, docs/PENDIENTES.md, docs/ROLES_Y_PERMISOS.md, docs/TESTING.md",
                "docs/evidencias/k6-resultado-carga.json y sonda-seguridad-resultado.json",
                "scripts/informe-auditoria/datos.py (fuente de este informe)",
            ])

    doc = Doc(os.path.join(DOCS, "INFORME_AUDITORIA_BUEN_VIVIR.pdf"))
    doc.multiBuild(story)


if __name__ == "__main__":
    generar_md()
    generar_pendientes()
    generar_pdf()
    c = conteos()
    print("HU:", len(D.HUS), dict(c))
