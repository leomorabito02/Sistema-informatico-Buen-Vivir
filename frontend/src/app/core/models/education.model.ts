/**
 * Modelos del microservicio de Educación (education-service).
 *
 * Reflejan 1:1 lo que devuelve la API real (Contentful + Firestore),
 * NO los tipos "aspiracionales" de los mocks de UX. Fuente de verdad:
 * education/internal/models/*.go y education/POSTMAN_EXAMPLES.md.
 */

/** Documento Rich Text de Contentful (el backend lo transporta tal cual). */
export interface RichTextNode {
  nodeType: string;
  data?: Record<string, unknown>;
  value?: string;
  marks?: { type: string }[];
  content?: RichTextNode[];
}
export type RichTextDocument = RichTextNode;

export interface AssetRef {
  url?: string;
  title?: string;
  contentType?: string;
}

export interface TrainingArea {
  id: string;
  name: string;
  description?: string;
}

export interface TrainingAreaInput {
  name: string;
  description?: string;
}

export type MaterialType = "PDF" | "Video" | "Link";

export interface Material {
  id: string;
  title: string;
  type?: MaterialType;
  file?: AssetRef;
  /** Enlace externo (Video/Link). Si hay `file` se prioriza `file`. */
  url?: string;
}

export interface MaterialInput {
  title: string;
  type?: MaterialType;
  fileAssetId?: string;
  /** Enlace externo (debe empezar con http:// o https://). */
  url?: string;
}

/** Devuelve el enlace efectivo de un material: el archivo subido o la URL externa. */
export function materialHref(m: Material): string | null {
  return m.file?.url ?? m.url ?? null;
}

export type CourseMode = "Presencial" | "Virtual" | "Mixta";

export interface Course {
  id: string;
  title: string;
  /** Rich Text de Contentful. Usar richTextToPlain() para mostrarlo como texto. */
  description: RichTextDocument | null;
  /** id de usuario (sub del JWT) del docente responsable, NO su nombre. */
  responsibleTeacher?: string;
  coverImage?: AssetRef;
  mode: string;
  duration?: string;
  trainingArea?: TrainingArea;
  material?: Material[];
  createdAt?: string;
  updatedAt?: string;
}

/** Payload para crear/editar un curso (reemplazo completo). */
export interface CourseInput {
  title: string;
  description: RichTextDocument;
  responsibleTeacher?: string;
  coverImageAssetId?: string;
  mode: string;
  duration?: string;
  trainingAreaId: string;
  materialIds?: string[];
}

export interface Activity {
  id: string;
  cursoId: string;
  titulo: string;
  consigna: string;
  /**
   * Fecha calendario de la actividad, como medianoche UTC en RFC3339
   * (ej. "2026-10-05T00:00:00Z"). Mostrarla con
   * `date: 'dd/MM/yyyy' : 'UTC'` y armarla con `toActivityFecha()`.
   */
  fecha: string;
  createdAt: string;
}

export interface ActivityInput {
  titulo: string;
  consigna: string;
  /** Ver `toActivityFecha()` en shared/utils/date-ymd.ts. */
  fecha: string;
}

/**
 * Registro de asistencia por actividad. No se carga a mano: lo genera
 * la entrega del participante (presente=true, fecha = fecha de la
 * actividad).
 */
export interface Attendance {
  id: string;
  participanteId: string;
  actividadId: string;
  presente: boolean;
  fecha: string;
}

/** Metadatos de un adjunto de una entrega (el binario vive en el bucket). */
export interface SubmissionFile {
  nombre: string;
  tipo: string;
  tamano: number;
}

/** Entrega de un participante para una actividad. Única e inmutable. */
export interface Submission {
  id: string;
  participanteId: string;
  actividadId: string;
  texto: string;
  archivos: SubmissionFile[];
  /** RFC3339, momento en que se entregó. */
  fechaEntrega: string;
}

export type SubmissionEstado = "entregado" | "sin_entregar";

/** Fila de `GET .../actividades/{id}/entregas`: un inscripto y su entrega. */
export interface SubmissionStatusRow {
  participanteId: string;
  nombre?: string;
  apellido?: string;
  estado: SubmissionEstado;
  entrega?: Submission;
}

/** Actividad tal como la ve el participante, con su propia entrega. */
export interface ParticipantActivity extends Activity {
  entrega?: Submission;
}

/** Respuesta de `GET /api/participante/cursos/{id}`. */
export interface ParticipantCourseDetail {
  curso: Course;
  actividades: ParticipantActivity[];
  progreso: ParticipantProgress;
}

export type EnrollmentEstado = "activa" | "finalizada" | "baja";

export interface Enrollment {
  id: string;
  participanteId: string;
  cursoId: string;
  fechaInscripcion: string;
  estado: EnrollmentEstado;
  /** Los completa `GET /api/cursos/{id}/inscripciones` contra auth-service. */
  nombre?: string;
  apellido?: string;
}

/** Payload de `POST /api/cursos/{id}/inscripciones` (role_admin/role_volunteer). */
export interface EnrollmentInput {
  participanteId: string;
}

/** Resultado de `GET /api/participantes?q=` (buscador de los selectores). */
export interface ParticipanteResumen {
  participanteId: string;
  nombre: string;
  apellido: string;
}

/** `GET /api/participants/{id}/asistencia`: historial por curso (HU-19). */
export interface HistorialCurso {
  cursoId: string;
  cursoTitulo: string;
  estadoInscripcion: EnrollmentEstado;
  fechaInscripcion: string;
  actividadesTotal: number;
  actividadesAsistidas: number;
  porcentajeAsistencia: number;
  actividades: {
    actividadId: string;
    titulo: string;
    fecha: string;
    presente: boolean;
  }[];
}

/**
 * Fila del registro de auditoría (Educación y Promoción Humana, ya
 * llevadas a la misma forma por cada service del frontend).
 */
export interface RegistroAuditoria {
  id: string;
  usuarioId: string;
  usuarioNombre?: string;
  accion: string;
  recurso: string;
  recursoId: string;
  fecha: string;
  ip?: string;
}

/** Filtros del registro de auditoría; fechas en `YYYY-MM-DD`. */
export interface FiltroAuditoria {
  desde?: string;
  hasta?: string;
  accion?: string;
  recurso?: string;
}

/**
 * Fila de `GET /api/asistencia-general`: un participante aprobado del
 * centro + su asistencia general para la fecha pedida. `presente` es
 * `null` cuando todavía no se cargó ningún registro ese día (distinto
 * de `false`, que es "se cargó y marcó ausente"). `nombre`/`apellido`
 * pueden venir vacíos si auth-service no pudo resolver el usuario.
 */
export interface GeneralAttendance {
  participanteId: string;
  nombre?: string;
  apellido?: string;
  presente: boolean | null;
  observacion?: string;
}

/** Ítem del body de `POST /api/asistencia-general?fecha=YYYY-MM-DD`. */
export interface GeneralAttendanceInput {
  participanteId: string;
  presente: boolean;
  observacion?: string;
}

/**
 * Progreso formativo (HU-20): actividades entregadas / total de
 * actividades del curso. `actividadesAsistidas` = actividades
 * entregadas (cada entrega genera su asistencia).
 */
export interface ParticipantProgress {
  participanteId: string;
  actividadesTotal: number;
  actividadesAsistidas: number;
  porcentajeAsistencia: number;
}

export interface EducationApiError {
  error: string;
}

/**
 * Aplana un documento Rich Text de Contentful a texto plano.
 * Suficiente para listados y previews; para render completo hace falta
 * un componente dedicado (pendiente, ver INTEGRACION_EDUCATION_FRONTEND.md).
 */
export function richTextToPlain(
  doc: RichTextDocument | null | undefined,
): string {
  if (!doc) return "";
  const walk = (node: RichTextNode): string => {
    if (typeof node.value === "string") return node.value;
    if (Array.isArray(node.content)) {
      const inner = node.content.map(walk).join("");
      return node.nodeType === "paragraph" ? inner + "\n" : inner;
    }
    return "";
  };
  return walk(doc).replace(/\n+$/g, "").trim();
}

/**
 * Construye un documento Rich Text mínimo de Contentful a partir de texto
 * plano (cada línea = un párrafo). Es lo que espera `POST/PUT /api/cursos`
 * en el campo `description`.
 */
export function plainToRichText(text: string): RichTextDocument {
  const paragraphs = (text || "").split(/\r?\n/);
  return {
    nodeType: "document",
    data: {},
    content: paragraphs.map((line) => ({
      nodeType: "paragraph",
      data: {},
      content: [{ nodeType: "text", value: line, marks: [], data: {} }],
    })),
  };
}
