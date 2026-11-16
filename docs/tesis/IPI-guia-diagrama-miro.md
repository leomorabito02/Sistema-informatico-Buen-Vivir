# Guía para actualizar el diagrama de base de datos en Miro (Figura 3)

El texto del IPI (sección "Diseño de Base de Datos") ya describe el modelo correcto.
Esta guía dice qué cambiar en el diagrama de Miro para que coincida con ese texto y
con el código. Los nombres de campo son los reales de Firestore.

Convención para las líneas: `1 ──< N` (uno a muchos), `1 ── 1` (uno a uno),
`1 ── 0..1` (uno a cero o uno). Las relaciones N:M se dibujan con una tabla
intermedia (pasarela) y dos líneas 1:N.

---

## 1. Educación (color de Educación)

### 1.1 Borrar

- **Curso_Habilidad**: ya no existe. Las habilidades se asocian a la **actividad**, no al curso.
- Si en **Asistencia** hay una línea que la conecta con **Curso** o un campo `curso_id`, se borra:
  la asistencia es por actividad.

### 1.2 Agregar o renombrar

| Tabla | Campos | Nota |
|---|---|---|
| **Actividad_Habilidad** (reemplaza a Curso_Habilidad) | `actividad_id` (FK), `habilidad_id` (FK), `curso_id` | `curso_id` es una copia para consultar rápido; no dibujar línea a Curso |
| **Entrega** (nueva) | `id` (PK), `participante_id` (FK), `actividad_id` (FK), `texto`, `archivos` (lista: nombre, tipo, tamaño, ruta), `fecha_entrega` | Una sola entrega por participante y actividad. Los archivos están en Cloud Storage; acá solo la ruta |
| **Asistencia** (revisar) | `id` (PK = `{actividad_id}_{participante_id}`), `participante_id` (FK), `actividad_id` (FK), `presente`, `fecha` | Se genera sola al entregar |
| **Participante_Habilidad** (revisar) | `participante_id` (FK), `habilidad_id` (FK), `origen` ('actividad' o 'manual'), `fecha`, `actividad_id`, `curso_id` | `actividad_id` y `curso_id` dicen de qué entrega salió |
| **Habilidad** (revisar) | `id` (PK), `nombre`, `categoria` | |
| **AsistenciaGeneral** (si no está) | `id` (PK), `participante_id` (FK), `fecha`, `presente`, `observacion`, `updated_at` | Asistencia diaria al centro, sin relación con cursos |
| **CursoBaja** (si no está) | `curso_id` (PK y FK), `titulo`, `fecha_baja`, `usuario_id` (FK), `inscriptos_activos` | Baja lógica del curso |
| **Auditoria** (si no está) | `id` (PK), `usuario_id` (FK), `accion`, `recurso`, `recurso_id`, `fecha`, `ip` | Solo se agrega, nunca se edita |

### 1.3 Relaciones (líneas) de Educación

| Desde | Hacia | Cardinalidad | Qué significa |
|---|---|---|---|
| Docente (Usuario) | Curso | 1 ──< N | un docente tiene a cargo varios cursos |
| Curso | Actividad | 1 ──< N | un curso tiene varias actividades |
| Curso | Inscripcion | 1 ──< N | un curso tiene muchas inscripciones |
| Participante | Inscripcion | 1 ──< N | (Participante N:M Curso resuelta con Inscripcion) |
| Actividad | Entrega | 1 ──< N | una actividad recibe muchas entregas |
| Participante | Entrega | 1 ──< N | (Participante N:M Actividad resuelta con Entrega) |
| Actividad | Asistencia | 1 ──< N | |
| Participante | Asistencia | 1 ──< N | (Participante N:M Actividad resuelta con Asistencia) |
| Entrega | Asistencia | 1 ── 1 | cada entrega genera su asistencia (mismo id); línea punteada con la etiqueta "genera" |
| Actividad | Actividad_Habilidad | 1 ──< N | |
| Habilidad | Actividad_Habilidad | 1 ──< N | (Actividad N:M Habilidad) |
| Participante | Participante_Habilidad | 1 ──< N | |
| Habilidad | Participante_Habilidad | 1 ──< N | (Participante N:M Habilidad) |
| Entrega | Participante_Habilidad | 1 ──< N | línea punteada con la etiqueta "otorga": al entregar se ganan las habilidades de la actividad |
| Participante | AsistenciaGeneral | 1 ──< N | |
| Curso | CursoBaja | 1 ── 0..1 | |
| Usuario | Auditoria | 1 ──< N | |
| Docente (Usuario) | ContenidoDocente | 1 ──< N | |
| ContenidoDocente | DescargaBiblioteca | 1 ──< N | |
| Usuario | Comentario | 1 ──< N (opcional) | `usuario_id` puede venir vacío; sin línea fija a Curso ni a Campaña (usa `entidad_tipo` + `entidad_id`) |

Poner una nota junto a Curso: "Curso, Área formativa y Material viven en
Contentful; las colecciones guardan su id".

---

## 2. Promoción Humana (color de Promoción Humana)

| Desde | Hacia | Cardinalidad |
|---|---|---|
| Participante | FichaProcesoPersonal | 1 ── 1 |
| FichaProcesoPersonal | Hito | 1 ──< N |
| Participante | LineaTiempo | 1 ──< N (campos: `id`, `tipo`, `fecha`, `descripcion`, `curso_id`) |
| Usuario | Auditoria | 1 ──< N |

Nota junto a la ficha: "situación actual, avances y descripción del hito se guardan
cifrados (AES-256-GCM)".

---

## 3. Dashboard

- Si existe una tabla **ReporteExportado** o **Snapshot**, borrarla: el Dashboard
  calcula los indicadores al momento y no guarda copias (ADR 0003). Solo queda
  el registro de reportes descargados, si está en el texto.

---

## 4. Revisión final antes de exportar

1. No queda ninguna tabla ni línea **Curso_Habilidad**.
2. **Entrega** está conectada a Actividad y a Participante.
3. **Asistencia** está conectada a Actividad y a Participante (no a Curso).
4. Las líneas entre servicios distintos (por ejemplo, Participante de
   Autenticación hacia Inscripcion de Educación) son referencias lógicas: si en el
   diagrama hay leyenda, aclarar "referencia por id entre servicios, sin FK real".
5. Exportar como PNG y reemplazar la imagen de la Figura 3 en el documento
   (clic derecho en la imagen → Reemplazar imagen → Subir del ordenador). Así se
   conserva el texto alternativo.
