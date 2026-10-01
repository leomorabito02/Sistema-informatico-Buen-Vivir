package models

import "time"

// JobOffer represents a job offer entity in Firestore.
type JobOffer struct {
	ID           string    `firestore:"id,omitempty" json:"id"`
	CompanyID    string    `firestore:"empresa_id" json:"empresa_id"`
	Title        string    `firestore:"titulo" json:"titulo"`
	Description  string    `firestore:"descripcion" json:"descripcion"`
	Location     string    `firestore:"ubicacion" json:"ubicacion"`
	Modality     string    `firestore:"modalidad" json:"modalidad"`
	JobType      string    `firestore:"tipo_empleo" json:"tipo_empleo"`
	Salary       string    `firestore:"salario" json:"salario"`
	Schedule     string    `firestore:"horario" json:"horario"`
	Requirements []string  `firestore:"requisitos" json:"requisitos"`
	Deadline     time.Time `firestore:"fecha_limite" json:"fecha_limite,omitempty"`
	Status       string    `firestore:"estado" json:"estado"`
	ApprovedBy   string    `firestore:"aprobado_por" json:"aprobado_por"`

	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
