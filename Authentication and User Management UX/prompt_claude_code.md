# Prompt para Claude Code — Revisión y ajustes de mocks (Sistema Buen Vivir)

## Contexto

Este es el proyecto "Sistema Informático Buen Vivir", en la rama `mocks` de este mismo repo.
La arquitectura general está compuesta por los siguientes microservicios: Frontend (Angular + PWA),
API Gateway, Autenticación, Notificaciones, Campañas, Educación, Promoción Humana, Inserción Laboral,
Dashboard, Curriculum (integración RxResume), y un módulo de Observabilidad. Hay integraciones externas
con Google Maps API, Contentful (CMS) y Gemini (IA generativa para sugerencia de habilidades).

Quiero que trabajes en dos frentes: (1) aplicar una serie de cambios puntuales sobre los mocks
existentes, y (2) recorrer todas las historias de usuario del proyecto y verificar qué está cubierto
a nivel mock y qué no.

---

## 1. CAMBIOS ESPECÍFICOS A APLICAR

1. **Descarga de contenido de biblioteca (landing page):** agregar un campo/formulario obligatorio
   para que la persona que descarga un contenido de la biblioteca pública deje asentado su contacto
   (nombre, apellido y teléfono) antes de habilitar la descarga. Corresponde a HU-25 (RF-25).

2. **Vista de ofertas laborales para el participante:** crear la vista donde el participante puede
   ver las ofertas laborales publicadas (catálogo), acorde a los permisos definidos en HU-06 (vista
   participante: "ver ofertas laborales sugeridas y generar su CV", "buscar ofertas laborales").

3. **Logos más minimalistas:** reemplazar los logos/iconos actuales de cada pestaña/sección por
   versión más minimalista, siguiendo el estilo visual de los logos ya usados en la landing page.
   No usar emojis en ningún ícono de navegación ni de sección.

4. **Campañas y Concientización:** sacar el botón de "Registrarse" de esta sección (no debería
   requerir registro). En su lugar, debe mantenerse/agregarse la posibilidad de dejar comentarios
   sobre el contenido público, acorde a HU-24 (nombre y apellido de quien comenta, sin necesidad de
   cuenta).

5. **Identidad visual unificada:** revisar que todas las vistas/mocks usen la misma paleta de
   colores definida para la identidad visual del proyecto (la misma que ya se usa en la landing
   page), en vez de colores por defecto o inconsistentes entre secciones.

6. **Portal docente — carrusel de fotos:** en la pestaña principal del portal docente, agregar un
   carrusel de fotos para hacer más amigable la experiencia de usuario (landing/bienvenida del
   portal, no un módulo funcional).

7. **Sección de cursos — referencia de diseño:** rediseñar la interfaz gráfica de la sección de
   cursos tomando como referencia el estilo de plataformas tipo Coderhouse (tarjetas de curso,
   distribución visual, tipografía, jerarquía de información).

Aplicá estos cambios directamente sobre los mocks/componentes correspondientes del repo.

---

## 2. VERIFICACIÓN CONTRA HISTORIAS DE USUARIO

Te paso todas las historias de usuario del proyecto, agrupadas por módulo (Autenticación y Gestión
de Usuarios, Educación, Promoción Humana, Inserción Laboral, Campañas y Concientización, Dashboard y
Reportes, Observabilidad, Integraciones Externas, Sistema de Notificaciones).

Recorré cada historia de usuario (HU-01 a HU-62) y verificá si está reflejada a nivel mock (pantallas,
componentes, datos de ejemplo, flujos simulados, validaciones visibles en el formulario, mensajes de
éxito/error mencionados en los criterios de aceptación). Para cada una indicá:

- **Estado:** Cubierta / Parcialmente cubierta / No cubierta
- **Si falta algo:** qué archivo, componente o mock habría que crear o modificar
- **Notas:** cualquier inconsistencia entre lo que pide la HU y lo que existe hoy en el mock (por
  ejemplo, mensajes de error/éxito que no coinciden con el texto exacto pedido, roles que no
  respetan los accesos definidos en HU-06, campos obligatorios que faltan, etc.)

Prestá especial atención a:
- Los textos exactos de mensajes de éxito/error que piden varias HU (ej. "Cambios guardados
  correctamente", "Se modificó el curso correctamente", "Oferta laboral eliminada correctamente",
  etc.) — verificá que estén igual o muy similares en los mocks.
- La matriz de accesos por rol de HU-06 (administrador, voluntario, participante, empresa, docente,
  docente externo) — que cada vista mockeada respete lo que ese rol puede/no puede ver.
- Los flujos de aprobación (voluntarios HU-12, empresas HU-38, ofertas laborales HU-40/41, contenido
  de docente externo HU-23).

Al terminar, devolveme un listado/tabla resumen con las 62 historias y su estado, antes de tocar
ningún archivo relacionado a este punto.

**Historias de usuario:** (pegar el archivo completo de HU-01 a HU-62 acá, o referenciar el archivo
si ya está en el repo, ej. `docs/historias_usuario.md`)

---

## Importante

- No implementes lógica de backend real ni conexión a servicios reales; esto es solo a nivel de
  mocks/frontend.
- Para el punto 2 (verificación de HU), mostrame primero el listado de gaps/resumen antes de generar
  o modificar código. Para el punto 1 (cambios específicos), podés aplicarlos directamente.
- Si alguna HU depende de una decisión de diseño no especificada (por ejemplo, exacto contenido del
  carrusel del portal docente), tomá la decisión más simple y razonable y avisame qué asumiste.
