package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCourseMode_IsValid_EquivalenceClasses tests valid and invalid partitions for CourseMode.
func TestCourseMode_IsValid_EquivalenceClasses(t *testing.T) {
	tests := []struct {
		name     string
		mode     CourseMode
		expected bool
	}{
		// Valid equivalence class
		{"Valid: Presencial", CourseModePresencial, true},
		{"Valid: Virtual", CourseModeVirtual, true},
		{"Valid: Mixta", CourseModeMixta, true},
		// Invalid equivalence class & boundaries
		{"Invalid: Empty string", "", false},
		{"Invalid: Lowercase presencial", "presencial", false},
		{"Invalid: Remota", "Remota", false},
		{"Invalid: Unknown string", "Hibrido", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			result := tt.mode.IsValid()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMaterialType_IsValid_EquivalenceClasses tests valid and invalid partitions for MaterialType.
func TestMaterialType_IsValid_EquivalenceClasses(t *testing.T) {
	tests := []struct {
		name     string
		matType  MaterialType
		expected bool
	}{
		// Valid equivalence class
		{"Valid: PDF", MaterialTypePDF, true},
		{"Valid: Video", MaterialTypeVideo, true},
		{"Valid: Link", MaterialTypeLink, true},
		// Invalid equivalence class & boundaries
		{"Invalid: Empty", "", false},
		{"Invalid: Audio", "Audio", false},
		{"Invalid: Lowercase pdf", "pdf", false},
		{"Invalid: Document", "Document", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			result := tt.matType.IsValid()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestContentfulLinks_NewEntryAndAssetLink tests entry and asset link constructors using AAA.
func TestContentfulLinks_NewEntryAndAssetLink(t *testing.T) {
	t.Run("NewEntryLink creates valid Link struct", func(t *testing.T) {
		// Arrange
		entryID := "entry-123"

		// Act
		link := NewEntryLink(entryID)

		// Assert
		assert.Equal(t, "Link", link.Sys.Type)
		assert.Equal(t, "Entry", link.Sys.LinkType)
		assert.Equal(t, entryID, link.Sys.ID)
	})

	t.Run("NewAssetLink creates valid Asset struct", func(t *testing.T) {
		// Arrange
		assetID := "asset-456"

		// Act
		link := NewAssetLink(assetID)

		// Assert
		assert.Equal(t, "Link", link.Sys.Type)
		assert.Equal(t, "Asset", link.Sys.LinkType)
		assert.Equal(t, assetID, link.Sys.ID)
	})
}

// TestIncludesIndex_EntryAndAssetResolution tests index lookups and asset URL normalization.
func TestIncludesIndex_EntryAndAssetResolution(t *testing.T) {
	// Arrange
	entries := []ContentfulEntry{
		{
			Sys: ContentfulSys{ID: "e1"},
			Fields: map[string]json.RawMessage{
				"title": json.RawMessage(`"Titulo 1"`),
			},
		},
	}

	assets := []ContentfulAsset{
		{
			Sys: ContentfulSys{ID: "a1"},
			Fields: struct {
				Title       json.RawMessage `json:"title"`
				Description json.RawMessage `json:"description"`
				File        json.RawMessage `json:"file"`
			}{
				Title: json.RawMessage(`"Imagen Portada"`),
				File:  json.RawMessage(`{"url":"//images.ctfassets.net/space/asset.png","fileName":"asset.png","contentType":"image/png"}`),
			},
		},
	}

	index := NewIncludesIndex(entries, assets)

	t.Run("Resolve existing entry", func(t *testing.T) {
		// Act
		entry, found := index.Entry("e1")

		// Assert
		require.True(t, found)
		assert.Equal(t, "e1", entry.Sys.ID)
	})

	t.Run("Resolve non-existing entry boundary", func(t *testing.T) {
		// Act
		_, found := index.Entry("non-existent")

		// Assert
		assert.False(t, found)
	})

	t.Run("Resolve existing asset with protocol prefix normalization", func(t *testing.T) {
		// Act
		asset, found := index.Asset("a1")

		// Assert
		require.True(t, found)
		assert.Equal(t, "a1", asset.ID)
		assert.Equal(t, "Imagen Portada", asset.Title)
		assert.Equal(t, "https://images.ctfassets.net/space/asset.png", asset.URL)
		assert.Equal(t, "asset.png", asset.FileName)
		assert.Equal(t, "image/png", asset.ContentType)
	})

	t.Run("Resolve non-existing asset boundary", func(t *testing.T) {
		// Act
		_, found := index.Asset("missing-asset")

		// Assert
		assert.False(t, found)
	})
}

// TestNormalizeAssetURL_Boundaries tests edge cases and boundary values for URL normalization.
func TestNormalizeAssetURL_Boundaries(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Protocol relative URL", "//cdn.contentful.com/asset.jpg", "https://cdn.contentful.com/asset.jpg"},
		{"Standard HTTPS URL", "https://example.com/image.png", "https://example.com/image.png"},
		{"Standard HTTP URL", "http://example.com/image.png", "http://example.com/image.png"},
		{"Boundary: Empty string", "", ""},
		{"Boundary: Single slash", "/", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			result := normalizeAssetURL(tt.input)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSkillCategory_IsValid_DecisionTable tests valid and invalid partitions for SkillCategory.
func TestSkillCategory_IsValid_DecisionTable(t *testing.T) {
	tests := []struct {
		name     string
		category SkillCategory
		expected bool
	}{
		{"Valid: Tecnologia", SkillCategoryTecnologia, true},
		{"Valid: Naturaleza", SkillCategoryNaturaleza, true},
		{"Valid: Oficio", SkillCategoryOficio, true},
		{"Valid: Socioemoc", SkillCategorySocioemoc, true},
		{"Invalid: Empty", "", false},
		{"Invalid: Arte", "Arte", false},
		{"Invalid: Lowercase", "tecnología", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			valid := tt.category.IsValid()

			// Assert
			assert.Equal(t, tt.expected, valid)
		})
	}
}
