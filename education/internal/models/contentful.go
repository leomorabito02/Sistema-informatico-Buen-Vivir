package models

import "encoding/json"

// ContentfulSys representa el bloque "sys" que Contentful adjunta a
// cualquier entry o asset, tanto en la Content Delivery API (CDA) como
// en la Content Management API (CMA).
type ContentfulSys struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Version          int    `json:"version,omitempty"`
	PublishedVersion int    `json:"publishedVersion,omitempty"`
	ContentType      *struct {
		Sys struct {
			ID string `json:"id"`
		} `json:"sys"`
	} `json:"contentType,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// ContentfulLink es la referencia liviana ({sys:{type:"Link",...}}) que
// Contentful usa tanto para leer como para escribir campos de tipo
// Reference o Media.
type ContentfulLink struct {
	Sys struct {
		Type     string `json:"type"`
		LinkType string `json:"linkType"`
		ID       string `json:"id"`
	} `json:"sys"`
}

func NewEntryLink(id string) ContentfulLink {
	l := ContentfulLink{}
	l.Sys.Type = "Link"
	l.Sys.LinkType = "Entry"
	l.Sys.ID = id
	return l
}

func NewAssetLink(id string) ContentfulLink {
	l := ContentfulLink{}
	l.Sys.Type = "Link"
	l.Sys.LinkType = "Asset"
	l.Sys.ID = id
	return l
}

// ContentfulEntry es la forma cruda de una entry tal como la devuelve
// Contentful (CDA o CMA), antes de mapearla al modelo propio de cada
// Content Type.
type ContentfulEntry struct {
	Sys    ContentfulSys              `json:"sys"`
	Fields map[string]json.RawMessage `json:"fields"`
}

// ContentfulAsset es la forma cruda de un asset (Media) de Contentful.
type ContentfulAsset struct {
	Sys    ContentfulSys `json:"sys"`
	Fields struct {
		Title       json.RawMessage `json:"title"`
		Description json.RawMessage `json:"description"`
		File        json.RawMessage `json:"file"`
	} `json:"fields"`
}

// contentfulAssetFile es el contenido del campo "file" de un asset,
// ya desenvuelto de su wrapper de locale.
type contentfulAssetFile struct {
	URL         string `json:"url"`
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
}

// ContentfulCollection es la respuesta de list endpoints de la CDA
// (GET /entries) e incluye los recursos enlazados en "includes" para
// poder resolver referencias (trainingArea, material, coverImage, file)
// sin hacer N+1 requests.
type ContentfulCollection struct {
	Total    int               `json:"total"`
	Skip     int               `json:"skip"`
	Limit    int               `json:"limit"`
	Items    []ContentfulEntry `json:"items"`
	Includes struct {
		Entry []ContentfulEntry `json:"Entry"`
		Asset []ContentfulAsset `json:"Asset"`
	} `json:"includes"`
}

// ContentfulSingleEntry es la respuesta de GET /entries/{id} en la CDA:
// una entry suelta con su propio bloque "includes".
type ContentfulSingleEntry struct {
	ContentfulEntry
	Includes struct {
		Entry []ContentfulEntry `json:"Entry"`
		Asset []ContentfulAsset `json:"Asset"`
	} `json:"includes"`
}

// IncludesIndex permite resolver Links contra los recursos embebidos en
// "includes" sin recorrer los slices repetidamente.
type IncludesIndex struct {
	entries map[string]ContentfulEntry
	assets  map[string]ContentfulAsset
}

func NewIncludesIndex(entries []ContentfulEntry, assets []ContentfulAsset) *IncludesIndex {
	idx := &IncludesIndex{
		entries: make(map[string]ContentfulEntry, len(entries)),
		assets:  make(map[string]ContentfulAsset, len(assets)),
	}
	for _, e := range entries {
		idx.entries[e.Sys.ID] = e
	}
	for _, a := range assets {
		idx.assets[a.Sys.ID] = a
	}
	return idx
}

func (idx *IncludesIndex) Entry(id string) (ContentfulEntry, bool) {
	e, ok := idx.entries[id]
	return e, ok
}

func (idx *IncludesIndex) Asset(id string) (AssetRef, bool) {
	a, ok := idx.assets[id]
	if !ok {
		return AssetRef{}, false
	}
	var file contentfulAssetFile
	_ = json.Unmarshal(a.Fields.File, &file)
	var title string
	_ = json.Unmarshal(a.Fields.Title, &title)
	return AssetRef{
		ID:          a.Sys.ID,
		Title:       title,
		URL:         normalizeAssetURL(file.URL),
		FileName:    file.FileName,
		ContentType: file.ContentType,
	}, true
}

// normalizeAssetURL antepone el protocolo a las URLs de assets de
// Contentful, que vienen como "//images.ctfassets.net/...".
func normalizeAssetURL(url string) string {
	if len(url) >= 2 && url[:2] == "//" {
		return "https:" + url
	}
	return url
}

// AssetRef es la representación resuelta (para el frontend) de una
// referencia Media (coverImage, material.file).
type AssetRef struct {
	ID          string `json:"id"`
	Title       string `json:"title,omitempty"`
	URL         string `json:"url"`
	FileName    string `json:"fileName,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}
