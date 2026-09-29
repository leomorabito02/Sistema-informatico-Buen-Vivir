import { Injectable, inject } from "@angular/core";
import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { Observable, throwError } from "rxjs";
import { catchError } from "rxjs/operators";
import { environment } from "../../environments/environment";
import {
  Activity,
  ActivityInput,
  Attendance,
  Course,
  CourseInput,
  Enrollment,
  EnrollmentInput,
  GeneralAttendance,
  GeneralAttendanceInput,
  Material,
  MaterialInput,
  ParticipantCourseDetail,
  ParticipantProgress,
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
  listCourses(): Observable<Course[]> {
    return this.http
      .get<Course[]>(`${this.base}/api/cursos`, { withCredentials: true })
      .pipe(catchError(this.handleError));
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

  deleteCourse(id: string): Observable<void> {
    return this.http
      .delete<void>(`${this.base}/api/cursos/${id}`, { withCredentials: true })
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

  deleteMaterial(id: string): Observable<void> {
    return this.http
      .delete<void>(`${this.base}/api/materiales/${id}`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  // ---- Portal docente (role_teacher interno) ---------------------------
  listMyCourses(): Observable<Course[]> {
    return this.http
      .get<Course[]>(`${this.base}/api/docente/cursos`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  listMyDraftCourses(): Observable<Course[]> {
    return this.http
      .get<Course[]>(`${this.base}/api/docente/cursos/borradores`, {
        withCredentials: true,
      })
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

  listEnrollments(cursoId: string): Observable<Enrollment[]> {
    return this.http
      .get<Enrollment[]>(
        `${this.base}/api/docente/cursos/${cursoId}/inscriptos`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  getCourseProgress(cursoId: string): Observable<ParticipantProgress[]> {
    return this.http
      .get<ParticipantProgress[]>(
        `${this.base}/api/docente/cursos/${cursoId}/progreso`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
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

  /** Lista los inscriptos de un curso (sin chequeo de ownership). */
  listCourseEnrollments(cursoId: string): Observable<Enrollment[]> {
    return this.http
      .get<Enrollment[]>(`${this.base}/api/cursos/${cursoId}/inscripciones`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
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
  listMyEnrolledCourses(): Observable<Course[]> {
    return this.http
      .get<Course[]>(`${this.base}/api/participante/cursos`, {
        withCredentials: true,
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
  private readonly handleError = (error: HttpErrorResponse) => {
    let message = "Ocurrió un error al comunicarse con el módulo de Educación.";
    const body = error.error as { error?: string } | string | null;
    if (body && typeof body === "object" && body.error) {
      message = body.error;
    } else if (error.status === 401) {
      message = "Tu sesión expiró o no es válida. Volvé a iniciar sesión.";
    } else if (error.status === 403) {
      message = "No tenés permisos para esta acción en Educación.";
    } else if (error.status === 404) {
      message = "El recurso solicitado no existe o no está publicado.";
    } else if (error.status === 413) {
      message = "Los archivos superan el tamaño permitido (10 MB cada uno).";
    } else if (error.status === 502) {
      message =
        "El módulo de Educación no pudo responder (Contentful/Firestore).";
    }
    return throwError(() => new Error(message));
  };
}
