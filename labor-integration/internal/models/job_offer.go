package models

import "time"

// JobOffer represents a job offer entity in Firestore.
type JobOffer struct {
	ID           string     `firestore:"id,omitempty" json:"id"`
	CompanyID    string     `firestore:"empresa_id" json:"empresa_id"`
	CreatorName  string     `firestore:"creador_nombre,omitempty" json:"creador_nombre,omitempty"`
	Title        string     `firestore:"titulo" json:"titulo"`
	Description  string     `firestore:"descripcion" json:"descripcion"`
	Location     string     `firestore:"ubicacion" json:"ubicacion"`
	Modality     string     `firestore:"modalidad" json:"modalidad"`
	JobType      string     `firestore:"tipo_empleo" json:"tipo_empleo"`
	Salary       string     `firestore:"salario" json:"salario"`
	Schedule     string     `firestore:"horario" json:"horario"`
	Vacancies    int        `firestore:"vacantes" json:"vacantes"`
	Requirements []string   `firestore:"requisitos" json:"requisitos"`
	Deadline     time.Time  `firestore:"fecha_limite" json:"fecha_limite,omitempty"`
	Status       string     `firestore:"estado" json:"estado"`
	ApprovedBy   string     `firestore:"aprobado_por" json:"aprobado_por"`
	ClosedAt     *time.Time `firestore:"fecha_cierre,omitempty" json:"fecha_cierre,omitempty"`
	ClosedBy     string     `firestore:"cerrado_por,omitempty" json:"cerrado_por,omitempty"`

	// Contact details populated only for staff roles (admin/volunteer).
	CompanyEmail string `firestore:"-" json:"empresa_email,omitempty"`
	CompanyPhone string `firestore:"-" json:"empresa_telefono,omitempty"`

	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
