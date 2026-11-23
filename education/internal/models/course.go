package models

import "encoding/json"

// CourseMode son los valores del campo "mode" del Content Type "course". Están en
// español y capitalizados como en Contentful: no cambiarlos.
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
	// ResponsibleTeacher es el id de usuario (sub del JWT) del docente responsable, no su nombre.
	ResponsibleTeacher string `json:"responsibleTeacher,omitempty"`
	// ResponsibleTeacherName es el nombre del docente, resuelto contra auth al responder
	// (no se guarda en Contentful).
	ResponsibleTeacherName string        `json:"responsibleTeacherName,omitempty"`
	CoverImage             *AssetRef     `json:"coverImage,omitempty"`
	Mode                   string        `json:"mode"`
	Duration               string        `json:"duration,omitempty"`
	TrainingArea           *TrainingArea `json:"trainingArea,omitempty"`
	Material               []Material    `json:"material,omitempty"`
	// Capacity es el cupo máximo de inscriptos activos. nil = sin límite.
	Capacity  *int   `json:"capacity,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// CourseInput es el payload para crear o editar un curso. La descripción viaja como
// Rich Text de Contentful, sin interpretarla.
type CourseInput struct {
	Title       string          `json:"title"`
	Description json.RawMessage `json:"description"`
	// ResponsibleTeacher se ignora en el PUT de docente: se fuerza al docente autenticado.
	ResponsibleTeacher string   `json:"responsibleTeacher,omitempty"`
	CoverImageAssetID  string   `json:"coverImageAssetId,omitempty"`
	Mode               string   `json:"mode"`
	Duration           string   `json:"duration,omitempty"`
	TrainingAreaID     string   `json:"trainingAreaId"`
	MaterialIDs        []string `json:"materialIds,omitempty"`
	// Capacity: nil = no tocar el cupo.
	Capacity *int `json:"capacity,omitempty"`
}
