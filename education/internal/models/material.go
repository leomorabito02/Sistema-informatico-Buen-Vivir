package models

// MaterialType enumera los valores predefinidos del campo "type" del
// Content Type "material" tal como están configurados en Contentful.
type MaterialType string

const (
	MaterialTypePDF   MaterialType = "PDF"
	MaterialTypeVideo MaterialType = "Video"
	MaterialTypeLink  MaterialType = "Link"
)

func (t MaterialType) IsValid() bool {
	switch t {
	case MaterialTypePDF, MaterialTypeVideo, MaterialTypeLink:
		return true
	default:
		return false
	}
}

// Material representa el Content Type "material" de Contentful.
type Material struct {
	ID    string    `json:"id"`
	Title string    `json:"title"`
	Type  string    `json:"type,omitempty"`
	File  *AssetRef `json:"file,omitempty"`
	// URL es un enlace externo (video, Drive, etc.). Si el material tiene File se usa
	// File; si no, URL.
	URL string `json:"url,omitempty"`
}

// MaterialInput es el payload que recibe el backend desde Angular para
// crear o editar un material. FileAssetID referencia un Asset ya
// existente en Contentful; URL es un enlace externo (Video/Link). Se
// puede mandar uno, otro, o ninguno.
type MaterialInput struct {
	Title       string `json:"title"`
	Type        string `json:"type,omitempty"`
	FileAssetID string `json:"fileAssetId,omitempty"`
	URL         string `json:"url,omitempty"`
}
