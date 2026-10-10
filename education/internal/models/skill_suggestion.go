package models

// SkillSuggestionRequest es el body de
// POST /api/actividades/sugerir-habilidades. Lo manda el frontend cuando
// alguien toca "Sugerir habilidades con IA" (nunca en cada tecla, para
// controlar el costo).
type SkillSuggestionRequest struct {
	CursoID string `json:"cursoId"`
	// ActividadID viene vacío si la actividad todavía no se creó.
	ActividadID string `json:"actividadId,omitempty"`
	Titulo      string `json:"titulo"`
	Consigna    string `json:"consigna"`
	// Excluir son los nombres de las habilidades que ya están en el
	// formulario (aceptadas o cargadas a mano y todavía sin guardar),
	// para que la IA no las vuelva a ofrecer.
	Excluir []string `json:"excluir,omitempty"`
}

// SkillPrompt es lo ÚNICO que se le manda a la IA: datos de la
// actividad y de su curso, más nombres del catálogo. Nunca datos de
// participantes. Tenerlo como tipo propio hace explícito qué sale del
// sistema hacia un proveedor externo.
type SkillPrompt struct {
	CursoTitulo     string
	AreaFormativa   string
	ActividadTitulo string
	Consigna        string
	// Catalogo son nombres de habilidades que ya existen, para que la IA
	// los reuse tal cual en vez de inventar variantes.
	Catalogo []string
	// YaAsignadas son las que la actividad ya tiene: no hay que repetirlas.
	YaAsignadas []string
	// MaxSugerencias es cuántas habilidades puede devolver como máximo.
	MaxSugerencias int
}

// SkillSuggestion es una habilidad propuesta por la IA.
type SkillSuggestion struct {
	// ID viene cargado cuando la sugerencia coincide con una habilidad
	// que ya existe en el catálogo (Existente = true): al aceptarla se
	// reusa esa, no se crea otra.
	ID            string        `json:"id,omitempty"`
	Nombre        string        `json:"nombre"`
	Categoria     SkillCategory `json:"categoria"`
	Justificacion string        `json:"justificacion"`
	Existente     bool          `json:"existente"`
}

// SkillSuggestionResponse es la respuesta del endpoint de sugerencias.
// Proveedor dice qué IA respondió ("gemini" u "openai"), para mostrarlo
// y para el registro.
type SkillSuggestionResponse struct {
	Proveedor   string            `json:"proveedor"`
	Sugerencias []SkillSuggestion `json:"sugerencias"`
}
