package models

// FechaLayout es el formato con que se guardan y viajan las fechas de
// los hitos (solo fecha, sin hora). El frontend las muestra DD/MM/AAAA.
const FechaLayout = "2006-01-02"

// Hito es un evento de avance del proceso de un participante (1:N con
// la ficha). Descripcion se guarda cifrada (ver HitoDocument).
type Hito struct {
	ID                  string `json:"id"`
	FichaID             string `json:"ficha_id"`
	Fecha               string `json:"fecha"`
	Descripcion         string `json:"descripcion"`
	ResponsableID       string `json:"responsable_id"`
	VisibleParticipante bool   `json:"visible_participante"`
}

// HitoDocument es la forma persistida en Firestore.
type HitoDocument struct {
	FichaID             string `firestore:"ficha_id"`
	Fecha               string `firestore:"fecha"`
	Descripcion         string `firestore:"descripcion"`
	ResponsableID       string `firestore:"responsable_id"`
	VisibleParticipante bool   `firestore:"visible_participante"`
}

// HitoInput es el body de alta y edición de un hito. ResponsableID es
// opcional: si viene vacío, el responsable es el usuario logueado
// (HU-32). VisibleParticipante en nil equivale a false al crear y a
// "sin cambios" al editar.
type HitoInput struct {
	Fecha               string `json:"fecha"`
	Descripcion         string `json:"descripcion"`
	ResponsableID       string `json:"responsable_id"`
	VisibleParticipante *bool  `json:"visible_participante"`
}
