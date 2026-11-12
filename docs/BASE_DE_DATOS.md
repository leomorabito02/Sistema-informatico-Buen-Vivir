# Bases de datos

Cada servicio es dueño de sus datos (no hay base compartida entre
servicios ni claves foráneas físicas entre ellos: los ids viajan como
texto). Los eventos llevan solo ids, nunca datos personales (ADR 0007).

| Servicio | Almacenamiento | Datos principales |
|---|---|---|
| auth | PostgreSQL (Neon), GORM | `users`, `roles`, `permissions`, `participants`, `companies`, `teachers`, `volunteers`, `participant_documents`, `refresh_tokens`, `password_recoveries`, dispositivos. Bajas lógicas con `deleted_at` |
| education | Contentful (CMA/CDA) | `course`, `trainingArea`, `material` (cursos publicados en la CDA; borradores en la CMA) |
| education | Firestore | `inscripciones`, `actividades`, `entregas`, `asistencias`, `asistencia_general`, `cursos_baja`, `habilidades`, `actividad_habilidades`, `participante_habilidades`, `comentarios`, `contenidos_docentes`, `descargas_biblioteca`, `auditoria`, `outbox_eventos`, `eventos_procesados` |
| education | Cloud Storage | Adjuntos de entregas y contenido docente (bucket privado) |
| human-promotion | Firestore | `fichas` (campos cifrados), `hitos`, `linea_tiempo`, `auditoria`, outbox |
| campaigns | Firestore | `campanas` (baja lógica con `eliminada`, `eliminada_en`, `eliminada_por`) |
| labor-integration | Firestore | `job_offers`, `job_applications` |
| dashboard | Firestore + Cloud Storage | `reportes_exportados` y PDF de reportes |
| notifications | PostgreSQL | notificaciones, destinatarios (con `dedupe_key`), preferencias, eventos procesados |
| rxresume | PostgreSQL | CV (Reactive Resume) |

Detalle de Educación: `der_educacion.md` y `../education/CLAUDE.md`.
DER general: `referencias/DER Buen Vivir-App.jpg` y su explicación en PDF.

## Integridad

- Cambio + evento (outbox) + auditoría en **una transacción** de Firestore
  (education, PH). Ids determinísticos donde hay unicidad
  (`{actividad}_{participante}` en entregas y asistencias, `{fecha}_{participante}`
  en asistencia general).
- Bajas lógicas: perfiles (auth), cursos (`cursos_baja`), campañas,
  ofertas con postulaciones (`estado = Eliminada`).

## Cuidado

El stack Docker local y el ambiente de testing usan **la misma** base
Firestore `(default)` y la misma base Neon. Una prueba con escritura en
local modifica testing. Para tests automáticos se usan los emuladores
(ver `TESTING.md`).
