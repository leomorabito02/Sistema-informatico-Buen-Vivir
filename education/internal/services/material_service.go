package services

import (
	"context"
	"fmt"
	"strings"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// MaterialUpload es el payload ya parseado del multipart de
// POST /api/materiales/upload.
type MaterialUpload struct {
	Title       string
	Type        string
	FileName    string
	ContentType string
	Data        []byte
}

// MaterialService valida las reglas de negocio de "material".
type MaterialService interface {
	ListMaterials(ctx context.Context) ([]models.Material, error)
	CreateMaterial(ctx context.Context, input models.MaterialInput) (*models.Material, error)
	UpdateMaterial(ctx context.Context, id string, input models.MaterialInput) (*models.Material, error)
	DeleteMaterial(ctx context.Context, id string) error
	// UploadMaterial crea un material subiendo el binario a Contentful
	// como Asset. Valida tipo de archivo y tamaño (imágenes 5 MB, PDF
	// 10 MB; el video se referencia por URL, no se sube).
	UploadMaterial(ctx context.Context, up MaterialUpload) (*models.Material, error)
}

const (
	maxImageBytes = 5 * 1024 * 1024
	maxDocBytes   = 10 * 1024 * 1024
)

var allowedUploadTypes = map[string]int{
	"image/jpeg":      maxImageBytes,
	"image/png":       maxImageBytes,
	"image/webp":      maxImageBytes,
	"application/pdf": maxDocBytes,
	"application/msword": maxDocBytes,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": maxDocBytes,
}

type materialService struct {
	repo repositories.MaterialRepository
}

func NewMaterialService(repo repositories.MaterialRepository) MaterialService {
	return &materialService{repo: repo}
}

func (s *materialService) ListMaterials(ctx context.Context) ([]models.Material, error) {
	return s.repo.List(ctx)
}

func (s *materialService) CreateMaterial(ctx context.Context, input models.MaterialInput) (*models.Material, error) {
	if err := validateMaterialInput(input); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, input)
}

func (s *materialService) UpdateMaterial(ctx context.Context, id string, input models.MaterialInput) (*models.Material, error) {
	if strings.TrimSpace(id) == "" {
		return nil, NewValidationError("el id del material es requerido")
	}
	if err := validateMaterialInput(input); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, input)
}

func (s *materialService) DeleteMaterial(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return NewValidationError("el id del material es requerido")
	}
	return s.repo.Delete(ctx, id)
}

func (s *materialService) UploadMaterial(ctx context.Context, up MaterialUpload) (*models.Material, error) {
	if strings.TrimSpace(up.Title) == "" {
		return nil, NewValidationError("el campo 'title' es requerido")
	}
	if up.Type == "" {
		up.Type = string(models.MaterialTypePDF)
	}
	if !models.MaterialType(up.Type).IsValid() {
		return nil, NewValidationError("el campo 'type' debe ser uno de: PDF, Video, Link")
	}
	if up.Type == string(models.MaterialTypeVideo) {
		return nil, NewValidationError("los videos se cargan por URL, no por archivo")
	}
	if len(up.Data) == 0 {
		return nil, NewValidationError("el archivo está vacío")
	}

	maxBytes, ok := allowedUploadTypes[up.ContentType]
	if !ok {
		return nil, NewValidationError("tipo de archivo no permitido: " + up.ContentType +
			" (se aceptan JPG, PNG, WEBP, PDF, DOC/DOCX)")
	}
	if len(up.Data) > maxBytes {
		return nil, NewValidationError(fmt.Sprintf(
			"el archivo pesa %.1f MB y el máximo para este tipo es %d MB",
			float64(len(up.Data))/(1024*1024), maxBytes/(1024*1024)))
	}

	return s.repo.UploadAndCreate(ctx, up.Title, up.Type, up.FileName, up.ContentType, up.Data)
}

func validateMaterialInput(input models.MaterialInput) error {
	if strings.TrimSpace(input.Title) == "" {
		return NewValidationError("el campo 'title' es requerido")
	}
	materialType := models.MaterialType(input.Type)
	if !materialType.IsValid() {
		return NewValidationError("el campo 'type' debe ser uno de: PDF, Video, Link")
	}
	if u := strings.TrimSpace(input.URL); u != "" &&
		!strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		return NewValidationError("el campo 'url' debe empezar con http:// o https://")
	}
	return nil
}
