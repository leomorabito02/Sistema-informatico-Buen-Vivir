package models

// Estados de una campaña (DER, tabla Campaña: "Borrador / publicada").
const (
	EstadoBorrador  = "borrador"
	EstadoPublicada = "publicada"
)

// FechaLayout es el formato de la fecha de la campaña: fecha calendario,
// sin hora (AAAA-MM-DD), igual que las fechas de los hitos en
// human-promotion.
const FechaLayout = "2006-01-02"

// Campaign es una campaña de concientización (HU-47 a HU-51). Los campos
// son los de la tabla Campaña del DER; creado_por es el usuario
// (responsable o voluntario) que la creó, sin FK física a la base de
// Autenticación.
type Campaign struct {
	ID        string `json:"id" firestore:"-"`
	Titulo    string `json:"titulo" firestore:"titulo"`
	Texto     string `json:"texto" firestore:"texto"`
	ImagenURL string `json:"imagen_url" firestore:"imagen_url"`
	Fecha     string `json:"fecha" firestore:"fecha"`
	Estado    string `json:"estado" firestore:"estado"`
	CreadoPor string `json:"creado_por,omitempty" firestore:"creado_por"`
}

// CampaignInput es el cuerpo de POST y PUT /api/v1/campaigns. El estado y
// el autor no se reciben: los decide el servicio (se crea en borrador y
// se publica con su propia operación, HU-49).
type CampaignInput struct {
	Titulo    string `json:"titulo"`
	Texto     string `json:"texto"`
	ImagenURL string `json:"imagen_url"`
	Fecha     string `json:"fecha"`
}
