package controllers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

// SkillController expone las habilidades (HU-21 / HU-58): el catálogo,
// las de cada actividad y las de cada participante.
//
// Quién puede gestionar las habilidades de una actividad: admin y
// voluntario en cualquier curso (los mismos que gestionan actividades),
// y el docente interno solo en los cursos de los que es responsable. Las
// rutas son las mismas para los tres roles y el chequeo de "es mi curso"
// lo hace este controller según el rol, igual que la descarga de
// adjuntos (SubmissionController.DownloadFile).
type SkillController struct {
	courseService     services.CourseService
	activityService   services.ActivityService
	skillService      services.SkillService
	suggestionService services.SkillSuggestionService
	participants      participantResolver
}

// participantResolver traduce el token del participante a su
// participant_id (el sub del JWT es el user_id, no el participant_id que
// usan las colecciones). Lo implementa SubmissionService.
type participantResolver interface {
	ResolveParticipantID(ctx context.Context, bearerToken string) (string, error)
}

func NewSkillController(
	courseService services.CourseService,
	activityService services.ActivityService,
	skillService services.SkillService,
	suggestionService services.SkillSuggestionService,
	participants participantResolver,
) *SkillController {
	return &SkillController{
		courseService:     courseService,
		activityService:   activityService,
		skillService:      skillService,
		suggestionService: suggestionService,
		participants:      participants,
	}
}

// SearchCatalog responde GET /api/habilidades?q=texto (autocompletado
// del campo "agregar habilidad a mano").
func (c *SkillController) SearchCatalog(w http.ResponseWriter, r *http.Request) {
	skills, err := c.skillService.SearchCatalog(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

// Categories responde GET /api/habilidades/categorias: la lista cerrada
// de categorías, para que el frontend no la tenga que repetir a mano.
func (c *SkillController) Categories(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, models.SkillCategories)
}

// CourseSkills responde GET /api/cursos/{id}/habilidades. Lo puede ver
// cualquier rol autenticado: son datos del catálogo del curso, no datos
// de personas.
func (c *SkillController) CourseSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := c.skillService.GetCourseSkills(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

// SetActivitySkills responde PUT /api/actividades/{actividadId}/habilidades.
// Body: {"habilidades": [{"id"?, "nombre", "categoria"}]}, la lista
// completa (reemplazo).
func (c *SkillController) SetActivitySkills(w http.ResponseWriter, r *http.Request) {
	activity, err := c.activityService.GetActivity(r.Context(), chi.URLParam(r, "actividadId"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := c.authorizeCourse(r, activity.CursoID); err != nil {
		writeError(w, err)
		return
	}

	var input models.ActivitySkillsInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	saved, err := c.skillService.SetActivitySkills(r.Context(), activity, input.Habilidades)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

// Suggest responde POST /api/actividades/sugerir-habilidades: pide a la
// IA habilidades para una actividad (nueva o existente). No guarda nada:
// las sugerencias se devuelven al formulario y quedan ahí hasta que
// alguien las acepta y guarda (PUT .../habilidades).
func (c *SkillController) Suggest(w http.ResponseWriter, r *http.Request) {
	var req models.SkillSuggestionRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}
	if err := c.authorizeCourse(r, req.CursoID); err != nil {
		writeError(w, err)
		return
	}
	if req.ActividadID != "" {
		activity, err := c.activityService.GetActivity(r.Context(), req.ActividadID)
		if err != nil {
			writeError(w, err)
			return
		}
		// Sin este chequeo, un docente podría pasar un curso suyo y una
		// actividad de otro curso.
		if activity.CursoID != req.CursoID {
			writeErrorJSON(w, http.StatusBadRequest, "la actividad no pertenece a ese curso")
			return
		}
	}

	// Título y área del curso le dan contexto a la IA. Es opcional: un
	// curso en borrador no está en la Content Delivery API y la sugerencia
	// se hace igual, solo con los datos de la actividad.
	course, err := c.courseService.GetCourse(r.Context(), req.CursoID)
	if err != nil {
		course = nil
	}

	resp, err := c.suggestionService.Suggest(r.Context(), req, course)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// ParticipantSkills responde GET /api/participants/{id}/habilidades
// (admin y voluntario): las habilidades que ganó un participante. Es la
// consulta que puede usar el Servicio de Dashboard para "habilidades
// desarrolladas" y la ficha de Promoción humana.
func (c *SkillController) ParticipantSkills(w http.ResponseWriter, r *http.Request) {
	c.writeParticipantSkills(w, r, chi.URLParam(r, "id"))
}

// MySkills responde GET /api/participante/habilidades: las habilidades
// del propio participante. Su id sale de su token, nunca del request,
// para que nadie pueda pedir las de otra persona.
func (c *SkillController) MySkills(w http.ResponseWriter, r *http.Request) {
	token, _ := r.Context().Value(middlewares.TokenKey).(string)
	participanteID, err := c.participants.ResolveParticipantID(r.Context(), token)
	if err != nil {
		writeError(w, err)
		return
	}
	c.writeParticipantSkills(w, r, participanteID)
}

// GrantedSkills responde GET /api/habilidades/otorgadas (solo admin):
// todas las habilidades ganadas por los participantes, con el curso en
// que se ganaron. Lo consume el servicio de Dashboard.
func (c *SkillController) GrantedSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := c.skillService.ListGrantedSkills(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

func (c *SkillController) writeParticipantSkills(w http.ResponseWriter, r *http.Request, participanteID string) {
	skills, err := c.skillService.ListParticipantSkills(r.Context(), participanteID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

// authorizeCourse decide si quien hace el pedido puede gestionar las
// habilidades de las actividades de ese curso. Admin y voluntario:
// cualquier curso que exista (incluso en borrador, porque las
// actividades se pueden armar antes de publicar). Docente interno: solo
// si es el responsable (CourseService.VerifyOwnership, que además
// responde 404 si el curso no está publicado, como el resto de
// /api/docente).
func (c *SkillController) authorizeCourse(r *http.Request, cursoID string) error {
	role, _ := r.Context().Value(middlewares.RoleIDKey).(string)
	if role == "role_teacher" {
		userID, _ := r.Context().Value(middlewares.UserIDKey).(string)
		return c.courseService.VerifyOwnership(r.Context(), cursoID, userID)
	}
	return c.courseService.VerifyExists(r.Context(), cursoID)
}
