import { Injectable, inject } from "@angular/core";
import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { Observable, throwError } from "rxjs";
import { catchError } from "rxjs/operators";
import { environment } from "../../environments/environment";
import {
  Activity,
  ActivityInput,
  Attendance,
  AttendanceInput,
  Course,
  CourseInput,
  Enrollment,
  Material,
  MaterialInput,
  ParticipantProgress,
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
  private http = inject(HttpClient);
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

  listAttendance(actividadId: string): Observable<Attendance[]> {
    return this.http
      .get<Attendance[]>(
        `${this.base}/api/docente/actividades/${actividadId}/asistencia`,
        { withCredentials: true },
      )
      .pipe(catchError(this.handleError));
  }

  registerAttendance(
    actividadId: string,
    inputs: AttendanceInput[],
  ): Observable<{ actividadId: string; status: string }> {
    return this.http
      .post<{ actividadId: string; status: string }>(
        `${this.base}/api/docente/actividades/${actividadId}/asistencia`,
        inputs,
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

  // ---------------------------------------------------------------------
  private handleError = (error: HttpErrorResponse) => {
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
    } else if (error.status === 502) {
      message =
        "El módulo de Educación no pudo responder (Contentful/Firestore).";
    }
    return throwError(() => new Error(message));
  };
}
