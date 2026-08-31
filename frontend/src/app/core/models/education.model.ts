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
  /** RFC3339 */
  fecha: string;
  createdAt: string;
}

export interface ActivityInput {
  titulo: string;
  /** RFC3339 */
  fecha: string;
}

export interface Attendance {
  id: string;
  participanteId: string;
  actividadId: string;
  presente: boolean;
  fecha: string;
}

export interface AttendanceInput {
  participanteId: string;
  presente: boolean;
}

export type EnrollmentEstado = "activa" | "finalizada" | "baja";

export interface Enrollment {
  id: string;
  participanteId: string;
  cursoId: string;
  fechaInscripcion: string;
  estado: EnrollmentEstado;
}

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
export function richTextToPlain(doc: RichTextDocument | null | undefined): string {
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
