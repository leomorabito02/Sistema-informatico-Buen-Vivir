package models

import "time"

// SkillCategory es la categoría de una habilidad del catálogo. Es una
// lista cerrada (decisión del equipo, 2026-10-06) y se guarda como texto,
// igual que el resto de los estados y enums del repo (ver
// EnrollmentEstado*): todavía está abierta la discusión de si los
// valores fijos van como texto o en una tabla catálogo.
type SkillCategory string

const (
	SkillCategoryTecnologia SkillCategory = "Tecnología"
	SkillCategoryNaturaleza SkillCategory = "Cuidado de la naturaleza"
	SkillCategoryOficio     SkillCategory = "Oficio"
	SkillCategorySocioemoc  SkillCategory = "Socioemocional"
)

// SkillCategories es el orden en que se muestran las categorías y el
// que se le pasa a la IA como lista permitida.
var SkillCategories = []SkillCategory{
	SkillCategoryTecnologia,
	SkillCategoryNaturaleza,
	SkillCategoryOficio,
	SkillCategorySocioemoc,
}

func (c SkillCategory) IsValid() bool {
	for _, v := range SkillCategories {
		if c == v {
			return true
		}
	}
	return false
}

// Origen de una habilidad en el perfil de un participante
// (Participante_Habilidad.origen). "actividad" = la ganó al entregar una
// actividad que la desarrolla; "manual" queda reservado para una futura
// carga directa (hoy ningún endpoint lo escribe).
const (
	SkillOrigenActividad = "actividad"
	SkillOrigenManual    = "manual"
)

// Skill es un documento de la colección Firestore "habilidades"
// (entidad Habilidad del DER). El id del documento es la clave
// normalizada del nombre (ver services.SkillKey): así dos escrituras del
// mismo nombre ("Trabajo en equipo" / "trabajo en  Equipo") caen en el
// mismo documento y Firestore impide el duplicado aunque lleguen a la vez.
type Skill struct {
	ID        string        `json:"id" firestore:"-"`
	Nombre    string        `json:"nombre" firestore:"nombre"`
	Categoria SkillCategory `json:"categoria" firestore:"categoria"`
}

// ActivitySkill es un documento de "actividad_habilidades" (tabla
// pasarela Actividad_Habilidad: qué habilidades desarrolla cada
// actividad). Id del documento: "{actividadId}_{habilidadId}".
type ActivitySkill struct {
	ActividadID string `firestore:"actividad_id"`
	HabilidadID string `firestore:"habilidad_id"`
	// CursoID no es parte de la clave del DER: se guarda para poder
	// listar las habilidades de un curso con una sola consulta, sin
	// recorrer actividad por actividad. Es la misma desnormalización que
	// ya hace "actividades" con curso_id.
	CursoID string `firestore:"curso_id"`
}

// ParticipantSkill es un documento de "participante_habilidades" (tabla
// pasarela Participante_Habilidad). Id del documento:
// "{participanteId}_{habilidadId}": un participante tiene cada habilidad
// una sola vez, aunque la haya ganado en varias actividades. Se conserva
// la primera vez que la obtuvo (origen y fecha no se pisan).
type ParticipantSkill struct {
	ParticipanteID string    `json:"participanteId" firestore:"participante_id"`
	HabilidadID    string    `json:"habilidadId" firestore:"habilidad_id"`
	Origen         string    `json:"origen" firestore:"origen"`
	Fecha          time.Time `json:"fecha" firestore:"fecha"`
	// Actividad y curso en los que la ganó (para el Dashboard: filtrar
	// por área formativa). Vacíos en las ganadas antes de 2026-10-08:
	// ListGrantedSkills los completa al leer, a partir de las entregas.
	ActividadID string `json:"actividadId,omitempty" firestore:"actividad_id,omitempty"`
	CursoID     string `json:"cursoId,omitempty" firestore:"curso_id,omitempty"`
}

// SkillGrant describe de dónde sale una habilidad que se le asigna a
// uno o más participantes.
type SkillGrant struct {
	Origen      string
	Fecha       time.Time
	ActividadID string
	CursoID     string
}

// GrantedSkill es una habilidad ganada por un participante, con el
// curso en el que la ganó: es lo que consulta el Dashboard
// (GET /api/habilidades/otorgadas). CursoID queda vacío si no se pudo
// averiguar (por ejemplo, la entrega ya no existe).
type GrantedSkill struct {
	ParticipanteID string        `json:"participanteId"`
	HabilidadID    string        `json:"habilidadId"`
	Nombre         string        `json:"nombre"`
	Categoria      SkillCategory `json:"categoria"`
	Origen         string        `json:"origen"`
	Fecha          time.Time     `json:"fecha"`
	ActividadID    string        `json:"actividadId,omitempty"`
	CursoID        string        `json:"cursoId,omitempty"`
}

// ParticipantSkillView es lo que devuelven los endpoints de consulta: la
// habilidad del participante con el nombre y la categoría del catálogo
// ya resueltos, para que quien consulta (frontend, Dashboard) no tenga
// que hacer una segunda llamada.
type ParticipantSkillView struct {
	Skill
	Origen string    `json:"origen"`
	Fecha  time.Time `json:"fecha"`
}

// SkillInput es una habilidad tal como la manda el frontend al guardar
// las de una actividad: puede ser una del catálogo (con id) o una nueva
// (solo nombre y categoría, por ejemplo una sugerencia de la IA aceptada
// o una escrita a mano). El backend decide si ya existe por el nombre
// normalizado, no por el id que mande el cliente.
type SkillInput struct {
	ID        string        `json:"id,omitempty"`
	Nombre    string        `json:"nombre"`
	Categoria SkillCategory `json:"categoria"`
}

// ActivitySkillsInput es el body de PUT /api/actividades/{id}/habilidades:
// la lista COMPLETA de habilidades de la actividad (reemplazo, igual que
// el PUT de actividades). Mandar una lista vacía quita todas.
type ActivitySkillsInput struct {
	Habilidades []SkillInput `json:"habilidades"`
}

// CourseSkills es la respuesta de GET /api/cursos/{id}/habilidades: las
// habilidades de todo el curso (unión de las de sus actividades, sin
// repetir) y las de cada actividad, para mostrarlas en el detalle del
// curso con una sola llamada.
type CourseSkills struct {
	Habilidades  []Skill            `json:"habilidades"`
	PorActividad map[string][]Skill `json:"porActividad"`
}
