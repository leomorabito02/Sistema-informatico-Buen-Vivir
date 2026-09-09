package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMaterialRepository struct {
	mock.Mock
}

func (m *MockMaterialRepository) List(ctx context.Context) ([]models.Material, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Material), args.Error(1)
}

func (m *MockMaterialRepository) Create(ctx context.Context, input models.MaterialInput) (*models.Material, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Material), args.Error(1)
}

func (m *MockMaterialRepository) Update(ctx context.Context, id string, input models.MaterialInput) (*models.Material, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Material), args.Error(1)
}

func (m *MockMaterialRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMaterialRepository) UploadAndCreate(ctx context.Context, title, materialType, fileName, contentType string, data []byte) (*models.Material, error) {
	args := m.Called(ctx, title, materialType, fileName, contentType, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Material), args.Error(1)
}

func TestMaterialService_ListMaterials(t *testing.T) {
	repo := new(MockMaterialRepository)
	svc := services.NewMaterialService(repo)

	repo.On("List", mock.Anything).Return([]models.Material{{ID: "m1"}}, nil)

	materials, err := svc.ListMaterials(context.Background())
	assert.NoError(t, err)
	assert.Len(t, materials, 1)
}

func TestMaterialService_CreateMaterial_MissingTitle(t *testing.T) {
	svc := services.NewMaterialService(nil)

	material, err := svc.CreateMaterial(context.Background(), models.MaterialInput{Title: "", Type: "PDF"})
	assert.Error(t, err)
	assert.Nil(t, material)
}

func TestMaterialService_CreateMaterial_InvalidType(t *testing.T) {
	svc := services.NewMaterialService(nil)

	material, err := svc.CreateMaterial(context.Background(), models.MaterialInput{Title: "Guía", Type: "Audio"})
	assert.Error(t, err)
	assert.Nil(t, material)
}

func TestMaterialService_CreateMaterial_InvalidURL(t *testing.T) {
	svc := services.NewMaterialService(nil)

	material, err := svc.CreateMaterial(context.Background(), models.MaterialInput{
		Title: "Video", Type: "Video", URL: "ftp://noesvalida.com",
	})
	assert.Error(t, err)
	assert.Nil(t, material)
}

func TestMaterialService_CreateMaterial_Success(t *testing.T) {
	repo := new(MockMaterialRepository)
	svc := services.NewMaterialService(repo)

	input := models.MaterialInput{Title: "Guía", Type: "PDF", FileAssetID: "asset-1"}
	repo.On("Create", mock.Anything, input).Return(&models.Material{ID: "m1"}, nil)

	material, err := svc.CreateMaterial(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, "m1", material.ID)
}

func TestMaterialService_UpdateMaterial_EmptyID(t *testing.T) {
	svc := services.NewMaterialService(nil)

	material, err := svc.UpdateMaterial(context.Background(), "", models.MaterialInput{Title: "Guía", Type: "PDF"})
	assert.Error(t, err)
	assert.Nil(t, material)
}

func TestMaterialService_DeleteMaterial_EmptyID(t *testing.T) {
	svc := services.NewMaterialService(nil)
	err := svc.DeleteMaterial(context.Background(), "")
	assert.Error(t, err)
}

func TestMaterialService_DeleteMaterial_Success(t *testing.T) {
	repo := new(MockMaterialRepository)
	svc := services.NewMaterialService(repo)

	repo.On("Delete", mock.Anything, "m1").Return(nil)
	err := svc.DeleteMaterial(context.Background(), "m1")
	assert.NoError(t, err)
}

func TestMaterialService_DeleteMaterial_RepoError(t *testing.T) {
	repo := new(MockMaterialRepository)
	svc := services.NewMaterialService(repo)

	repo.On("Delete", mock.Anything, "m1").Return(errors.New("contentful caído"))
	err := svc.DeleteMaterial(context.Background(), "m1")
	assert.Error(t, err)
}

func TestMaterialService_UploadMaterial_MissingTitle(t *testing.T) {
	svc := services.NewMaterialService(nil)

	up := services.MaterialUpload{Title: "", ContentType: "application/pdf", Data: []byte("x")}
	material, err := svc.UploadMaterial(context.Background(), up)
	assert.Error(t, err)
	assert.Nil(t, material)
}

func TestMaterialService_UploadMaterial_VideoRejected(t *testing.T) {
	svc := services.NewMaterialService(nil)

	up := services.MaterialUpload{Title: "Clase 1", Type: "Video", ContentType: "video/mp4", Data: []byte("x")}
	material, err := svc.UploadMaterial(context.Background(), up)
	assert.Error(t, err)
	assert.Nil(t, material)
	assert.Contains(t, err.Error(), "URL")
}

func TestMaterialService_UploadMaterial_EmptyFile(t *testing.T) {
	svc := services.NewMaterialService(nil)

	up := services.MaterialUpload{Title: "Guía", Type: "PDF", ContentType: "application/pdf", Data: nil}
	material, err := svc.UploadMaterial(context.Background(), up)
	assert.Error(t, err)
	assert.Nil(t, material)
}

func TestMaterialService_UploadMaterial_DisallowedContentType(t *testing.T) {
	svc := services.NewMaterialService(nil)

	up := services.MaterialUpload{Title: "Guía", Type: "PDF", ContentType: "application/zip", Data: []byte("x")}
	material, err := svc.UploadMaterial(context.Background(), up)
	assert.Error(t, err)
	assert.Nil(t, material)
	assert.Contains(t, err.Error(), "no permitido")
}

func TestMaterialService_UploadMaterial_TooLarge(t *testing.T) {
	svc := services.NewMaterialService(nil)

	// El límite para imágenes es 5 MB.
	data := make([]byte, 6*1024*1024)
	up := services.MaterialUpload{Title: "Foto", Type: "PDF", ContentType: "image/png", Data: data}
	material, err := svc.UploadMaterial(context.Background(), up)
	assert.Error(t, err)
	assert.Nil(t, material)
	assert.True(t, strings.Contains(err.Error(), "máximo"))
}

func TestMaterialService_UploadMaterial_Success(t *testing.T) {
	repo := new(MockMaterialRepository)
	svc := services.NewMaterialService(repo)

	data := []byte("contenido-pdf")
	repo.On("UploadAndCreate", mock.Anything, "Guía", "PDF", "guia.pdf", "application/pdf", data).
		Return(&models.Material{ID: "m1", Title: "Guía"}, nil)

	up := services.MaterialUpload{
		Title: "Guía", Type: "PDF", FileName: "guia.pdf", ContentType: "application/pdf", Data: data,
	}
	material, err := svc.UploadMaterial(context.Background(), up)
	assert.NoError(t, err)
	assert.Equal(t, "m1", material.ID)
}

// Sin type explícito, UploadMaterial asume PDF por default.
func TestMaterialService_UploadMaterial_DefaultsToPDF(t *testing.T) {
	repo := new(MockMaterialRepository)
	svc := services.NewMaterialService(repo)

	data := []byte("contenido-pdf")
	repo.On("UploadAndCreate", mock.Anything, "Guía", "PDF", "guia.pdf", "application/pdf", data).
		Return(&models.Material{ID: "m1"}, nil)

	up := services.MaterialUpload{
		Title: "Guía", FileName: "guia.pdf", ContentType: "application/pdf", Data: data,
	}
	_, err := svc.UploadMaterial(context.Background(), up)
	assert.NoError(t, err)
}
