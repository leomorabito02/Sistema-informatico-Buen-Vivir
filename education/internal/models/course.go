package models

import "encoding/json"

// CourseMode enumera los valores predefinidos del campo "mode" del
// Content Type "course" en Contentful. Los valores están en español
// y capitalizados porque así están configurados en el Space (validation
// "in": ["Presencial","Virtual","Mixta"]) — no traducir ni cambiar el casing.
type CourseMode string

const (
	CourseModePresencial CourseMode = "Presencial"
	CourseModeVirtual    CourseMode = "Virtual"
	CourseModeMixta      CourseMode = "Mixta"
)

func (m CourseMode) IsValid() bool {
	switch m {
	case CourseModePresencial, CourseModeVirtual, CourseModeMixta:
		return true
	default:
		return false
	}
}

// Course representa el Content Type "course" de Contentful, con sus
// referencias (trainingArea, material) ya resueltas a partir del
// bloque "includes" de la respuesta de la Content Delivery API.
type Course struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Description json.RawMessage `json:"description"`
	// ResponsibleTeacher es el id de usuario (el "sub" del JWT emitido
	// por auth-service) del docente responsable del curso, NO su
	// nombre en texto libre. El Content Type de Contentful sigue
	// declarándolo como Short text; es el backend el que reinterpreta
	// su contenido como un id. El frontend todavía necesita un
	// selector de docentes (que resuelva nombre -> id contra
	// auth-service) en vez de un campo de texto libre — pendiente.
	ResponsibleTeacher string        `json:"responsibleTeacher,omitempty"`
	CoverImage         *AssetRef     `json:"coverImage,omitempty"`
	Mode               string        `json:"mode"`
	Duration           string        `json:"duration,omitempty"`
	TrainingArea       *TrainingArea `json:"trainingArea,omitempty"`
	Material           []Material    `json:"material,omitempty"`
	// Capacity es el cupo máximo de inscriptos activos (campo "capacity"
	// en Contentful, Number, opcional). nil = sin límite de cupo — ver
	// EnrollmentService.Enroll, que es quien valida contra este valor.
	Capacity  *int   `json:"capacity,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// CourseInput es el payload que recibe el backend desde Angular para
// crear o editar un curso. La descripción viaja como el documento
// Rich Text de Contentful ya armado por el editor del frontend; el
// backend lo transporta tal cual, sin interpretarlo.
type CourseInput struct {
	Title       string          `json:"title"`
	Description json.RawMessage `json:"description"`
	// ResponsibleTeacher: ver comentario en Course. No requiere
	// validación de formato especial más allá de no estar vacío — es
	// responsabilidad del frontend enviar un id válido de
	// auth-service. En el PUT de docente (UpdateCourseAsTeacher) este
	// valor se ignora y se fuerza al id del propio docente autenticado,
	// para que no pueda reasignar el curso a otra persona.
	ResponsibleTeacher string   `json:"responsibleTeacher,omitempty"`
	CoverImageAssetID  string   `json:"coverImageAssetId,omitempty"`
	Mode               string   `json:"mode"`
	Duration           string   `json:"duration,omitempty"`
	TrainingAreaID     string   `json:"trainingAreaId"`
	MaterialIDs        []string `json:"materialIds,omitempty"`
	// Capacity: ver comentario en Course. Requiere que el campo
	// "capacity" (Number) exista en el content type "course" del Space
	// de Contentful — si se manda un valor antes de que ese campo se
	// haya creado ahí, Contentful rechaza el guardado (422, campo
	// desconocido). nil = no fijar/tocar el cupo.
	Capacity *int `json:"capacity,omitempty"`
}
