package dto

import "time"

type CreateJobOfferRequest struct {
	Title        string    `json:"titulo" validate:"required"`
	Description  string    `json:"descripcion" validate:"required"`
	Location     string    `json:"ubicacion"`
	Modality     string    `json:"modalidad"`
	JobType      string    `json:"tipo_empleo"`
	Salary       string    `json:"salario"`
	Schedule     string    `json:"horario"`
	Requirements []string  `json:"requisitos"`
	Deadline     time.Time `json:"fecha_limite"`
	CreatorName  string    `json:"creador_nombre,omitempty"`
}

type UpdateJobOfferRequest struct {
	Title        string    `json:"titulo"`
	Description  string    `json:"descripcion"`
	Location     string    `json:"ubicacion"`
	Modality     string    `json:"modalidad"`
	JobType      string    `json:"tipo_empleo"`
	Salary       string    `json:"salario"`
	Schedule     string    `json:"horario"`
	Requirements []string  `json:"requisitos"`
	Deadline     time.Time `json:"fecha_limite"`
	Status       string    `json:"estado"`
}
