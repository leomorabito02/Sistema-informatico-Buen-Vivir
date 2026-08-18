package models

// TrainingArea representa el Content Type "trainingArea" de Contentful.
// Es un catálogo de referencia (RF-28: permite buscar cursos por área),
// por eso vive como Content Type propio en vez de un campo de texto
// libre dentro de "course".
type TrainingArea struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// TrainingAreaInput es el payload que recibe el backend desde Angular
// para crear o editar un área formativa.
type TrainingAreaInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
