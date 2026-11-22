import { Injectable, inject } from "@angular/core";
import {
  HttpClient,
  HttpErrorResponse,
  HttpHeaders,
  HttpParams,
} from "@angular/common/http";
import {
  MonoTypeOperatorFunction,
  Observable,
  retry,
  throwError,
  timer,
} from "rxjs";
import { catchError, map } from "rxjs/operators";
import { withSkipGlobalError } from "../core/interceptors/error.interceptor";
import { environment } from "../../environments/environment";
import {
  Activity,
  ActivityInput,
  Attendance,
  Course,
  CourseInput,
  Enrollment,
  EnrollmentInput,
  FiltroAuditoria,
  GeneralAttendance,
  GeneralAttendanceInput,
  HistorialCurso,
  Material,
  MaterialInput,
  ParticipantCourseDetail,
  ParticipantProgress,
  ParticipanteResumen,
  RegistroAuditoria,
  Submission,
  SubmissionStatusRow,
  TrainingArea,
  TrainingAreaInput,
} from "../core/models/education.model";

/**
 * Cliente del microservicio de Educación (education-service).
 *
 * Mismo patrón que AuthService: `withCredentials: true` (la cookie
 * `access_token` que dejó auth-service viaja sola y education la valida),
 * y `catchError` que normaliza el `{ "error": "..." }` del backend.
 *
 * Base URL:
 *  - dev local (`ng serve -c local`): ambos vacíos -> `""` -> el proxy manda `/api/*` a :8210
 *  - Docker detrás del gateway: `educationApiUrl` vacío -> cae a `apiUrl` (el gateway :8080)
 */

/**
 * `silent: true` no dispara el manejo global de errores (pantalla /500 o
 * diálogo): lo usa una pantalla que junta varios pedidos y resuelve sola
 * el que falla, como Seguimiento (22/11/2026).
 */
export interface SilentOpts {
  silent?: boolean;
}

function silentHeaders(opts: SilentOpts): { headers?: HttpHeaders } {
  return opts.silent ? withSkipGlobalError() : {};
}

/**
 * Con `silent`, reintenta hasta 2 veces las fallas pasajeras (sin red o
 * 5xx: Cloud Run arrancando o saturado) antes de dar el error.
 */
function reintentoSilencioso<T>(opts: SilentOpts): MonoTypeOperatorFunction<T> {
  if (!opts.silent) return (source) => source;
  return retry({
    count: 2,
    delay: (err: unknown, intento: number) => {
      const status = (err as HttpErrorResponse)?.status ?? 0;
      // 429 no se reintenta: el límite es por minuto y reintentar lo empeora.
      if (status === 0 || status >= 500) {
        return timer(700 * intento);
      }
      return throwError(() => err);
    },
  });
}

@Injectable({ providedIn: "root" })
export class EducationService {
  private readonly http = inject(HttpClient);
  private readonly base =
    environment.educationApiUrl || environment.apiUrl || "";

  // ---- Salud -------------------------------------------------------------
  health(): Observable<{
    status: string;
    service: string;
    appEnv: string;
    contentfulEnvironment: string;
  }> {
    return this.http
      .get<{
        status: string;
        service: string;
        appEnv: string;
        contentfulEnvironment: string;
      }>(`${this.base}/api/salud`)
      .pipe(catchError(this.handleError));
  }

  // ---- Catálogo (cualquier rol autenticado) -----------------------------
  /**
   * q (opcional) busca en título, área y descripción y area (id, opcional)
   * deja solo los cursos de esa área formativa (HU-28). Los dos los
   * aplica el backend. No trae los cursos dados de baja.
   */
  listCourses(q = "", area = "", opts: SilentOpts = {}): Observable<Course[]> {
    return this.http
      .get<Course[]>(`${this.base}/api/cursos`, {
        withCredentials: true,
        params: busqueda(q, area),
        ...silentHeaders(opts),
      })
      .pipe(reintentoSilencioso(opts), catchError(this.handleError));
  }

  getCourse(id: string): Observable<Course> {
    return this.http
      .get<Course>(`${this.base}/api/cursos/${id}`, { withCredentials: true })
      .pipe(catchError(this.handleError));
  }

  /** Solo admin: todos los cursos en borrador (sin publicar). */
  listDraftCourses(): Observable<Course[]> {
    return this.http
      .get<Course[]>(`${this.base}/api/cursos/borradores`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  listTrainingAreas(): Observable<TrainingArea[]> {
    return this.http
      .get<TrainingArea[]>(`${this.base}/api/areas-formativas`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  listMaterials(): Observable<Material[]> {
    return this.http
      .get<Material[]>(`${this.base}/api/materiales`, { withCredentials: true })
      .pipe(catchError(this.handleError));
  }

  // ---- Escritura de catálogo (role_admin) ------------------------------
  createCourse(input: CourseInput): Observable<Course> {
    return this.http
      .post<Course>(`${this.base}/api/cursos`, input, { withCredentials: true })
      .pipe(catchError(this.handleError));
  }

  updateCourse(id: string, input: CourseInput): Observable<Course> {
    return this.http
      .put<Course>(`${this.base}/api/cursos/${id}`, input, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  /** Baja lógica del curso (HU-16): sale del catálogo y queda en el historial. */
  deleteCourse(id: string): Observable<void> {
    return this.http
      .delete<void>(`${this.base}/api/cursos/${id}`, { withCredentials: true })
      .pipe(catchError(this.handleError));
  }

  /** Cuántos inscriptos activos tiene el curso: el aviso antes de la baja. */
  countActiveEnrollments(cursoId: string): Observable<number> {
    return this.http
      .get<{ inscriptosActivos: number }>(
        `${this.base}/api/cursos/${cursoId}/inscriptos-activos`,
        { withCredentials: true },
      )
      .pipe(
        map((r) => r.inscriptosActivos),
        catchError(this.handleError),
      );
  }

  /** Registro de auditoría de Educación (solo administración). */
  listAuditoria(filtro: FiltroAuditoria = {}): Observable<RegistroAuditoria[]> {
    return this.http
      .get<RegistroAuditoria[]>(`${this.base}/api/auditoria`, {
        withCredentials: true,
        params: filtrosAuditoria(filtro),
      })
      .pipe(catchError(this.handleError));
  }

  publishCourse(id: string): Observable<{ id: string; status: string }> {
    return this.http
      .post<{ id: string; status: string }>(
        `${this.base}/api/cursos/${id}/publicar`,
        {},
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  unpublishCourse(id: string): Observable<{ id: string; status: string }> {
    return this.http
      .post<{ id: string; status: string }>(
        `${this.base}/api/cursos/${id}/despublicar`,
        {},
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  // ---- Áreas formativas (escritura: role_admin) ------------------------
  createTrainingArea(input: TrainingAreaInput): Observable<TrainingArea> {
    return this.http
      .post<TrainingArea>(`${this.base}/api/areas-formativas`, input, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  updateTrainingArea(
    id: string,
    input: TrainingAreaInput,
  ): Observable<TrainingArea> {
    return this.http
      .put<TrainingArea>(`${this.base}/api/areas-formativas/${id}`, input, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  deleteTrainingArea(id: string): Observable<void> {
    return this.http
      .delete<void>(`${this.base}/api/areas-formativas/${id}`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  // ---- Materiales (escritura: role_admin; se auto-publican) ------------
  createMaterial(input: MaterialInput): Observable<Material> {
    return this.http
      .post<Material>(`${this.base}/api/materiales`, input, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  /** Sube un archivo (imagen/PDF) y crea el material con ese Asset. */
  uploadMaterial(
    file: File,
    title: string,
    type: string,
  ): Observable<Material> {
    const form = new FormData();
    form.append("file", file);
    form.append("title", title);
    form.append("type", type);
    return this.http
      .post<Material>(`${this.base}/api/materiales/upload`, form, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  updateMaterial(id: string, input: MaterialInput): Observable<Material> {
    return this.http
      .put<Material>(`${this.base}/api/materiales/${id}`, input, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  /** HU-14: cursos que tienen adjunto el material (aviso antes de borrar). */
  coursesUsingMaterial(
    id: string,
  ): Observable<{ id: string; titulo: string }[]> {
    return this.http
      .get<{ id: string; titulo: string }[]>(
        `${this.base}/api/materiales/${id}/cursos`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  deleteMaterial(id: string): Observable<void> {
    return this.http
      .delete<void>(`${this.base}/api/materiales/${id}`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  // ---- Portal docente (role_teacher interno) ---------------------------
  listMyCourses(opts: SilentOpts = {}): Observable<Course[]> {
    return this.http
      .get<Course[]>(`${this.base}/api/docente/cursos`, {
        withCredentials: true,
        ...silentHeaders(opts),
      })
      .pipe(reintentoSilencioso(opts), catchError(this.handleError));
  }

  listMyDraftCourses(): Observable<Course[]> {
    return this.http
      .get<Course[]>(`${this.base}/api/docente/cursos/borradores`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  /**
   * El docente crea un curso a su nombre: nace en borrador y queda
   * pendiente hasta que el administrador lo aprueba y publica.
   */
  createMyCourse(input: CourseInput): Observable<Course> {
    return this.http
      .post<Course>(`${this.base}/api/docente/cursos`, input, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  /** El docente da de baja un curso suyo que sigue pendiente (borrador). */
  deleteMyCourse(id: string): Observable<void> {
    return this.http
      .delete<void>(
        `${this.base}/api/docente/cursos/${encodeURIComponent(id)}`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  /** El docente responsable edita su curso (responsibleTeacher se ignora en el backend). */
  updateMyCourse(id: string, input: CourseInput): Observable<Course> {
    return this.http
      .put<Course>(`${this.base}/api/docente/cursos/${id}`, input, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  listActivities(cursoId: string): Observable<Activity[]> {
    return this.http
      .get<Activity[]>(
        `${this.base}/api/docente/cursos/${cursoId}/actividades`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  createActivity(cursoId: string, input: ActivityInput): Observable<Activity> {
    return this.http
      .post<Activity>(
        `${this.base}/api/docente/cursos/${cursoId}/actividades`,
        input,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  /** El docente responsable edita una actividad de su curso. */
  updateActivity(
    actividadId: string,
    input: ActivityInput,
  ): Observable<Activity> {
    return this.http
      .put<Activity>(
        `${this.base}/api/docente/actividades/${actividadId}`,
        input,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  /** Borra la actividad junto con sus entregas y asistencias. */
  deleteActivity(actividadId: string): Observable<void> {
    return this.http
      .delete<void>(`${this.base}/api/docente/actividades/${actividadId}`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  /** Inscriptos del curso con el estado de entrega de la actividad. */
  listActivitySubmissions(
    actividadId: string,
  ): Observable<SubmissionStatusRow[]> {
    return this.http
      .get<SubmissionStatusRow[]>(
        `${this.base}/api/docente/actividades/${actividadId}/entregas`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  /** Asistencia (solo lectura: la generan las entregas). */
  listAttendance(actividadId: string): Observable<Attendance[]> {
    return this.http
      .get<Attendance[]>(
        `${this.base}/api/docente/actividades/${actividadId}/asistencia`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  listEnrollments(
    cursoId: string,
    opts: SilentOpts = {},
  ): Observable<Enrollment[]> {
    return this.http
      .get<Enrollment[]>(
        `${this.base}/api/docente/cursos/${cursoId}/inscriptos`,
        { withCredentials: true, ...silentHeaders(opts) },
      )
      .pipe(reintentoSilencioso(opts), catchError(this.handleError));
  }

  getCourseProgress(
    cursoId: string,
    opts: SilentOpts = {},
  ): Observable<ParticipantProgress[]> {
    return this.http
      .get<ParticipantProgress[]>(
        `${this.base}/api/docente/cursos/${cursoId}/progreso`,
        { withCredentials: true, ...silentHeaders(opts) },
      )
      .pipe(reintentoSilencioso(opts), catchError(this.handleError));
  }

  /**
   * Progreso formativo de cualquier curso (HU-20), para admin y
   * voluntario: `GET /api/cursos/{id}/progreso` (auditoría 2026-11-11).
   */
  getCourseProgressGestion(
    cursoId: string,
    opts: SilentOpts = {},
  ): Observable<ParticipantProgress[]> {
    return this.http
      .get<ParticipantProgress[]>(
        `${this.base}/api/cursos/${encodeURIComponent(cursoId)}/progreso`,
        { withCredentials: true, ...silentHeaders(opts) },
      )
      .pipe(reintentoSilencioso(opts), catchError(this.handleError));
  }

  // ---- Inscripción, asistencia por actividad y asistencia general ------
  // (role_admin / role_volunteer, sin scope de "mis cursos" — agregado
  // 2026-09-17 junto con las pantallas nuevas del frontend. Ver
  // education/CONTEXTO_INSCRIPCIONES_ASISTENCIA.md para el detalle de
  // los endpoints del backend, ya implementados y probados.)

  /** Inscribe un participante en un curso publicado. */
  enrollParticipant(
    cursoId: string,
    input: EnrollmentInput,
  ): Observable<Enrollment> {
    return this.http
      .post<Enrollment>(
        `${this.base}/api/cursos/${cursoId}/inscripciones`,
        input,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  getParticipantEnrollments(participantId: string): Observable<Enrollment[]> {
    return this.http
      .get<Enrollment[]>(
        `${this.base}/api/participants/${participantId}/inscripciones`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  getTeacherCourses(teacherId: string): Observable<Course[]> {
    return this.http
      .get<Course[]>(`${this.base}/api/docentes/${teacherId}/cursos`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  getBulkUserCourses(
    userIds: string[],
  ): Observable<{ user_id: string; courses: string[] }[]> {
    return this.http
      .post<{ user_id: string; courses: string[] }[]>(
        `${this.base}/api/bulk-cursos-usuarios`,
        { user_ids: userIds },
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  /**
   * Buscador de participantes de los selectores (admin y voluntario): el
   * backend busca por nombre, apellido o DNI y devuelve hasta 20. Con
   * cursoId deja afuera a los ya inscriptos en ese curso.
   */
  searchParticipants(
    q: string,
    cursoId = "",
  ): Observable<ParticipanteResumen[]> {
    let params = new HttpParams().set("q", q.trim());
    if (cursoId) params = params.set("cursoId", cursoId);
    return this.http
      .get<ParticipanteResumen[]>(`${this.base}/api/participantes`, {
        withCredentials: true,
        params,
      })
      .pipe(catchError(this.handleError));
  }

  /** Historial de asistencia de un participante, por curso (HU-19). */
  getParticipantAttendanceHistory(
    participanteId: string,
  ): Observable<HistorialCurso[]> {
    return this.http
      .get<HistorialCurso[]>(
        `${this.base}/api/participants/${participanteId}/asistencia`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  /** Inscriptos de un curso con nombre y apellido (sin chequeo de ownership). */
  listCourseEnrollments(
    cursoId: string,
    opts: SilentOpts = {},
  ): Observable<Enrollment[]> {
    return this.http
      .get<Enrollment[]>(`${this.base}/api/cursos/${cursoId}/inscripciones`, {
        withCredentials: true,
        ...silentHeaders(opts),
      })
      .pipe(reintentoSilencioso(opts), catchError(this.handleError));
  }

  /**
   * Actividades de un curso, sin chequeo de ownership (agregado
   * 2026-09-21, ver `POST/GET /api/cursos/{id}/actividades` en
   * education/CLAUDE.md). Reemplaza, para esta pantalla, al `listActivities`
   * de más arriba (exclusivo de `role_teacher`).
   */
  listCourseActivities(cursoId: string): Observable<Activity[]> {
    return this.http
      .get<Activity[]>(`${this.base}/api/cursos/${cursoId}/actividades`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  createCourseActivity(
    cursoId: string,
    input: ActivityInput,
  ): Observable<Activity> {
    return this.http
      .post<Activity>(`${this.base}/api/cursos/${cursoId}/actividades`, input, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  updateCourseActivity(
    actividadId: string,
    input: ActivityInput,
  ): Observable<Activity> {
    return this.http
      .put<Activity>(`${this.base}/api/actividades/${actividadId}`, input, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  /** Borra la actividad junto con sus entregas y asistencias. */
  deleteCourseActivity(actividadId: string): Observable<void> {
    return this.http
      .delete<void>(`${this.base}/api/actividades/${actividadId}`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  /** Inscriptos del curso con el estado de entrega de la actividad. */
  listCourseActivitySubmissions(
    actividadId: string,
  ): Observable<SubmissionStatusRow[]> {
    return this.http
      .get<SubmissionStatusRow[]>(
        `${this.base}/api/actividades/${actividadId}/entregas`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  /**
   * Asistencia de una actividad, sin chequeo de ownership del curso.
   * Solo lectura: la generan las entregas de los participantes.
   */
  listActivityAttendance(actividadId: string): Observable<Attendance[]> {
    return this.http
      .get<Attendance[]>(
        `${this.base}/api/actividades/${actividadId}/asistencia`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  // ---- Participante (role_participant) ----------------------------------

  /** Cursos en los que el participante autenticado está inscripto. */
  listMyEnrolledCourses(q = "", area = ""): Observable<Course[]> {
    return this.http
      .get<Course[]>(`${this.base}/api/participante/cursos`, {
        withCredentials: true,
        params: busqueda(q, area),
      })
      .pipe(catchError(this.handleError));
  }

  /** Curso + actividades con el estado de mi entrega + mi progreso. */
  getMyCourse(cursoId: string): Observable<ParticipantCourseDetail> {
    return this.http
      .get<ParticipantCourseDetail>(
        `${this.base}/api/participante/cursos/${cursoId}`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  /** Entrega una actividad (texto opcional + adjuntos). No se puede anular. */
  submitActivity(
    actividadId: string,
    texto: string,
    archivos: File[],
  ): Observable<Submission> {
    const form = new FormData();
    form.append("texto", texto);
    for (const f of archivos) {
      form.append("archivos", f, f.name);
    }
    return this.http
      .post<Submission>(
        `${this.base}/api/participante/actividades/${actividadId}/entrega`,
        form,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  /**
   * Descarga un adjunto de una entrega. Pasa por education (no hay URL
   * pública): el backend valida que quien pide pueda ver esa entrega.
   */
  downloadSubmissionFile(entregaId: string, indice: number): Observable<Blob> {
    return this.http
      .get(`${this.base}/api/entregas/${entregaId}/archivos/${indice}`, {
        withCredentials: true,
        responseType: "blob",
      })
      .pipe(catchError(this.handleError));
  }

  /**
   * Asistencia general del centro para una fecha (`YYYY-MM-DD`): todos
   * los participantes aprobados + su registro de ese día, si existe.
   */
  getGeneralAttendance(fecha: string): Observable<GeneralAttendance[]> {
    return this.http
      .get<GeneralAttendance[]>(`${this.base}/api/asistencia-general`, {
        withCredentials: true,
        params: { fecha },
      })
      .pipe(catchError(this.handleError));
  }

  /** Upsert en lote de asistencia general para una fecha. */
  registerGeneralAttendance(
    fecha: string,
    inputs: GeneralAttendanceInput[],
  ): Observable<{ fecha: string; status: string }> {
    return this.http
      .post<{ fecha: string; status: string }>(
        `${this.base}/api/asistencia-general`,
        inputs,
        { withCredentials: true, params: { fecha } },
      )
      .pipe(catchError(this.handleError));
  }

  // ---------------------------------------------------------------------
  /**
   * Mensaje para la persona: qué pasó y qué puede hacer. El texto del
   * backend se muestra solo cuando es una regla de negocio en castellano
   * (400, 403, 409, 422, 503); los errores técnicos (404 genérico, 5xx,
   * sin conexión) se reemplazan por un mensaje con una sugerencia.
   */
  private readonly handleError = (error: HttpErrorResponse) => {
    const body = error.error as { error?: string } | string | null;
    const backendMsg =
      body && typeof body === "object" && body.error ? body.error : "";
    const legible =
      backendMsg &&
      [400, 403, 409, 422, 503].includes(error.status) &&
      !/^body inválido/i.test(backendMsg);

    let message =
      "No pudimos completar la acción. Volvé a intentar en unos minutos; si sigue pasando, avisá al equipo del centro.";
    if (legible) {
      message = backendMsg.charAt(0).toUpperCase() + backendMsg.slice(1);
    } else if (error.status === 0) {
      message =
        "No hay conexión con el servidor. Revisá tu internet y volvé a intentar.";
    } else if (error.status === 400) {
      message =
        "Algunos datos no son válidos. Revisá el formulario y volvé a intentar.";
    } else if (error.status === 401) {
      message = "Tu sesión expiró o no es válida. Volvé a iniciar sesión.";
    } else if (error.status === 403) {
      message =
        "No tenés permisos para esta acción. Si creés que es un error, consultá con la persona responsable del centro.";
    } else if (error.status === 404) {
      message =
        "No encontramos lo que buscabas: puede que se haya borrado o que todavía no esté publicado. Volvé a la lista para ver lo actualizado.";
    } else if (error.status === 413) {
      message =
        "Los archivos superan el tamaño permitido (10 MB cada uno). Elegí archivos más chicos o subilos de a uno.";
    } else if (error.status >= 500) {
      message =
        "El servidor de cursos no respondió. Esperá unos minutos y volvé a intentar.";
    }
    return throwError(() => new Error(message));
  };
}

/** Parámetros ?q= y ?area= de los buscadores; vacíos = sin parámetro. */
function busqueda(q: string, area = ""): HttpParams {
  let params = new HttpParams();
  const t = q.trim();
  if (t) params = params.set("q", t);
  if (area) params = params.set("area", area);
  return params;
}

/** Filtros de auditoría como query params, sin los vacíos. */
export function filtrosAuditoria(f: FiltroAuditoria): HttpParams {
  let params = new HttpParams();
  for (const [k, v] of Object.entries(f)) {
    if (v) params = params.set(k, v);
  }
  return params;
}
