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
	ID                 string          `json:"id"`
	Title              string          `json:"title"`
	Description        json.RawMessage `json:"description"`
	ResponsibleTeacher string          `json:"responsibleTeacher,omitempty"`
	CoverImage         *AssetRef       `json:"coverImage,omitempty"`
	Mode               string          `json:"mode"`
	Duration           string          `json:"duration,omitempty"`
	TrainingArea       *TrainingArea   `json:"trainingArea,omitempty"`
	Material           []Material      `json:"material,omitempty"`
	CreatedAt          string          `json:"createdAt,omitempty"`
	UpdatedAt          string          `json:"updatedAt,omitempty"`
}

// CourseInput es el payload que recibe el backend desde Angular para
// crear o editar un curso. La descripción viaja como el documento
// Rich Text de Contentful ya armado por el editor del frontend; el
// backend lo transporta tal cual, sin interpretarlo.
type CourseInput struct {
	Title              string          `json:"title"`
	Description        json.RawMessage `json:"description"`
	ResponsibleTeacher string          `json:"responsibleTeacher,omitempty"`
	CoverImageAssetID  string          `json:"coverImageAssetId,omitempty"`
	Mode               string          `json:"mode"`
	Duration           string          `json:"duration,omitempty"`
	TrainingAreaID     string          `json:"trainingAreaId"`
	MaterialIDs        []string        `json:"materialIds,omitempty"`
}
