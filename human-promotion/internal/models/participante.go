package models

// ParticipanteInfo son los datos del participante que viven en Auth y
// se consultan por gRPC (no se guardan en este servicio).
type ParticipanteInfo struct {
	ParticipanteID string
	UserID         string
	Nombre         string
	Apellido       string
	// FechaNacimiento en formato AAAA-MM-DD; vacía si no está cargada.
	FechaNacimiento string
}

// DatosParticipante es la parte de solo lectura de la ficha (HU-31):
// se muestra pero se edita en su servicio de origen.
type DatosParticipante struct {
	Nombre          string `json:"nombre"`
	Apellido        string `json:"apellido"`
	FechaNacimiento string `json:"fecha_nacimiento,omitempty"`
	Edad            *int   `json:"edad,omitempty"`
}

// FichaDetalle es la ficha tal como la ve un profesional: el contenido
// confidencial descifrado más los datos de Auth. Participante queda en
// nil si Auth no respondió, para que la ficha se pueda seguir viendo.
type FichaDetalle struct {
	FichaProcesoPersonal
	Participante    *DatosParticipante `json:"participante"`
	CreadoPorNombre string             `json:"creado_por_nombre,omitempty"`
	// Habilidades que ganó el participante en Educación (HU-21, HU-58).
	// Vacío si education-service no respondió o no está configurado.
	Habilidades []string `json:"habilidades"`
}

// HitoDetalle es un hito de la línea de tiempo profesional (HU-33),
// con el nombre del responsable resuelto en Auth.
type HitoDetalle struct {
	Hito
	ResponsableNombre string `json:"responsable_nombre,omitempty"`
}

// HitoParticipante es lo único que ve el participante de un hito
// (HU-34): nunca el responsable, la ficha ni datos confidenciales.
type HitoParticipante struct {
	ID          string `json:"id"`
	Fecha       string `json:"fecha"`
	Descripcion string `json:"descripcion"`
}

// ResultadoBusqueda es un participante encontrado por nombre (HU-36),
// con el id de su ficha si ya tiene una.
type ResultadoBusqueda struct {
	ParticipanteID string  `json:"participante_id"`
	Nombre         string  `json:"nombre"`
	Apellido       string  `json:"apellido"`
	FichaID        *string `json:"ficha_id"`
}
