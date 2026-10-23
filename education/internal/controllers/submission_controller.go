package controllers

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

// maxSubmissionUploadBytes es el límite duro del multipart de una
// entrega: hasta 5 archivos de 10 MB más el texto. Los límites por
// archivo los valida SubmissionService.
const maxSubmissionUploadBytes = 52 << 20

// submissionTypesByExt completa el Content-Type cuando el navegador
// manda un adjunto como application/octet-stream (pasa con algunos
// .docx/.xlsx según el sistema operativo).
var submissionTypesByExt = map[string]string{
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".txt":  "text/plain",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

// SubmissionController agrupa los endpoints del rol participante
// (/api/participante: mis cursos, detalle del curso, entregar) y la
// descarga de adjuntos de una entrega, que usan todos los roles.
type SubmissionController struct {
	courseService     services.CourseService
	activityService   services.ActivityService
	submissionService services.SubmissionService
}

func NewSubmissionController(courseService services.CourseService, activityService services.ActivityService, submissionService services.SubmissionService) *SubmissionController {
	return &SubmissionController{courseService: courseService, activityService: activityService, submissionService: submissionService}
}

func (c *SubmissionController) participantID(r *http.Request) (string, error) {
	token, _ := r.Context().Value(middlewares.TokenKey).(string)
	return c.submissionService.ResolveParticipantID(r.Context(), token)
}

// ListMyCourses responde GET /api/participante/cursos
func (c *SubmissionController) ListMyCourses(w http.ResponseWriter, r *http.Request) {
	q, ok := busquedaCursos(w, r)
	if !ok {
		return
	}
	participanteID, err := c.participantID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	courses, err := c.submissionService.ListParticipantCourses(r.Context(), participanteID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, services.FiltrarCursosPorArea(services.FiltrarCursos(courses, q), r.URL.Query().Get("area")))
}

// GetMyCourse responde GET /api/participante/cursos/{id}
func (c *SubmissionController) GetMyCourse(w http.ResponseWriter, r *http.Request) {
	participanteID, err := c.participantID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	detail, err := c.submissionService.GetParticipantCourse(r.Context(), participanteID, chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// Submit responde POST /api/participante/actividades/{actividadId}/entrega
// (multipart/form-data): campo "texto" opcional y cero o más archivos
// en el campo "archivos".
func (c *SubmissionController) Submit(w http.ResponseWriter, r *http.Request) {
	participanteID, err := c.participantID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxSubmissionUploadBytes)
	// #nosec G120 -- request body is bounded by http.MaxBytesReader above
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "no se pudo leer la entrega (¿los archivos superan el tamaño permitido?)")
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	var uploads []models.SubmissionUpload
	for _, fh := range r.MultipartForm.File["archivos"] {
		f, err := fh.Open()
		if err != nil {
			writeErrorJSON(w, http.StatusBadRequest, "no se pudo leer el archivo "+fh.Filename)
			return
		}
		data, err := io.ReadAll(f)
		_ = f.Close()
		if err != nil {
			writeErrorJSON(w, http.StatusBadRequest, "no se pudo leer el archivo "+fh.Filename)
			return
		}
		uploads = append(uploads, models.SubmissionUpload{
			Nombre: fh.Filename,
			Tipo:   uploadContentType(fh.Filename, fh.Header.Get("Content-Type")),
			Data:   data,
		})
	}

	sub, err := c.submissionService.Submit(r.Context(), chi.URLParam(r, "actividadId"), participanteID, r.FormValue("texto"), uploads)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sub)
}

// DownloadFile responde GET /api/entregas/{entregaId}/archivos/{indice}.
// Pueden descargar: admin/voluntario (cualquier entrega), el docente
// responsable del curso de la actividad, y el propio participante que
// hizo la entrega.
func (c *SubmissionController) DownloadFile(w http.ResponseWriter, r *http.Request) {
	index, err := strconv.Atoi(chi.URLParam(r, "indice"))
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "índice de archivo inválido")
		return
	}
	sub, err := c.submissionService.GetSubmission(r.Context(), chi.URLParam(r, "entregaId"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := c.authorizeSubmissionAccess(r, sub); err != nil {
		writeError(w, err)
		return
	}

	file, rc, err := c.submissionService.OpenFile(r.Context(), sub, index)
	if err != nil {
		writeError(w, err)
		return
	}
	defer func() { _ = rc.Close() }()

	w.Header().Set("Content-Type", file.Tipo)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": file.Nombre}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}

func (c *SubmissionController) authorizeSubmissionAccess(r *http.Request, sub *models.Submission) error {
	role, _ := r.Context().Value(middlewares.RoleIDKey).(string)
	switch role {
	case "role_admin", "role_dev", "role_volunteer":
		return nil
	case "role_teacher":
		if isExternal, _ := r.Context().Value(middlewares.IsExternalKey).(bool); isExternal {
			return services.NewForbiddenError("no tenés acceso a esta entrega")
		}
		activity, err := c.activityService.GetActivity(r.Context(), sub.ActividadID)
		if err != nil {
			return err
		}
		return c.courseService.VerifyOwnership(r.Context(), activity.CursoID, teacherIDFromRequest(r))
	case "role_participant":
		participanteID, err := c.participantID(r)
		if err != nil {
			return err
		}
		if participanteID != sub.ParticipanteID {
			return services.NewForbiddenError("no tenés acceso a esta entrega")
		}
		return nil
	default:
		return services.NewForbiddenError("no tenés acceso a esta entrega")
	}
}

func uploadContentType(fileName, declared string) string {
	declared = strings.TrimSpace(strings.SplitN(declared, ";", 2)[0])
	if declared != "" && declared != "application/octet-stream" {
		return declared
	}
	return submissionTypesByExt[strings.ToLower(filepath.Ext(fileName))]
}
