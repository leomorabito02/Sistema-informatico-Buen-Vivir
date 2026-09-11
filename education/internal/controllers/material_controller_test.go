package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

type MockMaterialService struct {
	mock.Mock
}

func (m *MockMaterialService) ListMaterials(ctx context.Context) ([]models.Material, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Material), args.Error(1)
}

func (m *MockMaterialService) CreateMaterial(ctx context.Context, input models.MaterialInput) (*models.Material, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Material), args.Error(1)
}

func (m *MockMaterialService) UpdateMaterial(ctx context.Context, id string, input models.MaterialInput) (*models.Material, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Material), args.Error(1)
}

func (m *MockMaterialService) DeleteMaterial(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMaterialService) UploadMaterial(ctx context.Context, up services.MaterialUpload) (*models.Material, error) {
	args := m.Called(ctx, up)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Material), args.Error(1)
}

func TestMaterialController_List_Success(t *testing.T) {
	mockSvc := new(MockMaterialService)
	ctrl := controllers.NewMaterialController(mockSvc)

	mockSvc.On("ListMaterials", mock.Anything).Return([]models.Material{{ID: "m1"}}, nil)

	req := httptest.NewRequest("GET", "/api/materiales", nil)
	w := httptest.NewRecorder()
	ctrl.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMaterialController_Create_BadJSON(t *testing.T) {
	mockSvc := new(MockMaterialService)
	ctrl := controllers.NewMaterialController(mockSvc)

	req := httptest.NewRequest("POST", "/api/materiales", bytes.NewBufferString("{invalid"))
	w := httptest.NewRecorder()
	ctrl.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMaterialController_Create_Success(t *testing.T) {
	mockSvc := new(MockMaterialService)
	ctrl := controllers.NewMaterialController(mockSvc)

	mockSvc.On("CreateMaterial", mock.Anything, mock.AnythingOfType("models.MaterialInput")).
		Return(&models.Material{ID: "m1"}, nil)

	body, _ := json.Marshal(models.MaterialInput{Title: "Guía", Type: "PDF"})
	req := httptest.NewRequest("POST", "/api/materiales", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	ctrl.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestMaterialController_Update_ServiceError(t *testing.T) {
	mockSvc := new(MockMaterialService)
	ctrl := controllers.NewMaterialController(mockSvc)

	mockSvc.On("UpdateMaterial", mock.Anything, "m1", mock.AnythingOfType("models.MaterialInput")).
		Return(nil, services.NewValidationError("el campo 'title' es requerido"))

	body, _ := json.Marshal(models.MaterialInput{})
	req := withChiURLParam(httptest.NewRequest("PUT", "/api/materiales/m1", bytes.NewBuffer(body)), "id", "m1")
	w := httptest.NewRecorder()
	ctrl.Update(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMaterialController_Delete_Success(t *testing.T) {
	mockSvc := new(MockMaterialService)
	ctrl := controllers.NewMaterialController(mockSvc)

	mockSvc.On("DeleteMaterial", mock.Anything, "m1").Return(nil)

	req := withChiURLParam(httptest.NewRequest("DELETE", "/api/materiales/m1", nil), "id", "m1")
	w := httptest.NewRecorder()
	ctrl.Delete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func newMultipartUploadRequest(t *testing.T, title, materialType, fileName, content string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	_ = writer.WriteField("title", title)
	_ = writer.WriteField("type", materialType)

	if fileName != "" {
		part, err := writer.CreateFormFile("file", fileName)
		if err != nil {
			t.Fatalf("no se pudo crear form file: %v", err)
		}
		_, _ = part.Write([]byte(content))
	}
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/api/materiales/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestMaterialController_Upload_MissingFile(t *testing.T) {
	mockSvc := new(MockMaterialService)
	ctrl := controllers.NewMaterialController(mockSvc)

	req := newMultipartUploadRequest(t, "Guía", "PDF", "", "")
	w := httptest.NewRecorder()
	ctrl.Upload(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMaterialController_Upload_Success(t *testing.T) {
	mockSvc := new(MockMaterialService)
	ctrl := controllers.NewMaterialController(mockSvc)

	mockSvc.On("UploadMaterial", mock.Anything, mock.AnythingOfType("services.MaterialUpload")).
		Return(&models.Material{ID: "m1"}, nil)

	req := newMultipartUploadRequest(t, "Guía", "PDF", "guia.pdf", "contenido")
	w := httptest.NewRecorder()
	ctrl.Upload(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestMaterialController_Upload_ServiceError(t *testing.T) {
	mockSvc := new(MockMaterialService)
	ctrl := controllers.NewMaterialController(mockSvc)

	mockSvc.On("UploadMaterial", mock.Anything, mock.AnythingOfType("services.MaterialUpload")).
		Return(nil, errors.New("contentful caído"))

	req := newMultipartUploadRequest(t, "Guía", "PDF", "guia.pdf", "contenido")
	w := httptest.NewRecorder()
	ctrl.Upload(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)
}
