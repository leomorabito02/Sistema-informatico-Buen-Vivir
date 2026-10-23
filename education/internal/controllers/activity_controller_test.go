package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

func newActivityControllerFixture() (*controllers.ActivityController, *MockCourseService, *MockActivityService, *MockSubmissionService) {
	courseSvc := new(MockCourseService)
	activitySvc := new(MockActivityService)
	submissionSvc := new(MockSubmissionService)
	ctrl := controllers.NewActivityController(courseSvc, activitySvc, submissionSvc)
	return ctrl, courseSvc, activitySvc, submissionSvc
}

// TestActivityController_List_DecisionTable tests List with valid, not found and error scenarios.
func TestActivityController_List_DecisionTable(t *testing.T) {
	t.Run("Success: returns activity list", func(t *testing.T) {
		// Arrange
		ctrl, courseSvc, activitySvc, _ := newActivityControllerFixture()
		courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
		activitySvc.On("ListByCourse", mock.Anything, "c1").Return([]models.Activity{{ID: "a1"}}, nil)

		req := withChiURLParam(httptest.NewRequest("GET", "/api/cursos/c1/actividades", nil), "id", "c1")
		w := httptest.NewRecorder()

		// Act
		ctrl.List(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"id":"a1"`)
	})

	t.Run("Course does not exist returns 404", func(t *testing.T) {
		// Arrange
		ctrl, _, activitySvc, _ := newActivityControllerFixture()
		activitySvc.On("ListByCourse", mock.Anything, "c1").Return(nil, repoErrNotFound())

		req := withChiURLParam(httptest.NewRequest("GET", "/api/cursos/c1/actividades", nil), "id", "c1")
		w := httptest.NewRecorder()

		// Act
		ctrl.List(w, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Activity service error returns 502", func(t *testing.T) {
		// Arrange
		ctrl, courseSvc, activitySvc, _ := newActivityControllerFixture()
		courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
		activitySvc.On("ListByCourse", mock.Anything, "c1").Return(nil, errors.New("firestore down"))

		req := withChiURLParam(httptest.NewRequest("GET", "/api/cursos/c1/actividades", nil), "id", "c1")
		w := httptest.NewRecorder()

		// Act
		ctrl.List(w, req)

		// Assert
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})
}

// TestActivityController_Create_DecisionTable tests Create with bad JSON, validation error, and success.
func TestActivityController_Create_DecisionTable(t *testing.T) {
	t.Run("Bad JSON returns 400", func(t *testing.T) {
		// Arrange
		ctrl, courseSvc, _, _ := newActivityControllerFixture()
		courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
		req := withChiURLParam(httptest.NewRequest("POST", "/api/cursos/c1/actividades", bytes.NewBufferString("{bad")), "id", "c1")
		w := httptest.NewRecorder()

		// Act
		ctrl.Create(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Success returns 201 Created", func(t *testing.T) {
		// Arrange
		ctrl, courseSvc, activitySvc, _ := newActivityControllerFixture()
		courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
		input := models.ActivityInput{Titulo: "Clase 1"}
		activitySvc.On("CreateActivity", mock.Anything, "c1", input).Return(&models.Activity{ID: "a1", Titulo: "Clase 1"}, nil)

		body, _ := json.Marshal(input)
		req := withChiURLParam(httptest.NewRequest("POST", "/api/cursos/c1/actividades", bytes.NewBuffer(body)), "id", "c1")
		w := httptest.NewRecorder()

		// Act
		ctrl.Create(w, req)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Service validation error returns 400", func(t *testing.T) {
		// Arrange
		ctrl, courseSvc, activitySvc, _ := newActivityControllerFixture()
		courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
		input := models.ActivityInput{Titulo: ""}
		activitySvc.On("CreateActivity", mock.Anything, "c1", input).Return(nil, services.NewValidationError("titulo requerido"))

		body, _ := json.Marshal(input)
		req := withChiURLParam(httptest.NewRequest("POST", "/api/cursos/c1/actividades", bytes.NewBuffer(body)), "id", "c1")
		w := httptest.NewRecorder()

		// Act
		ctrl.Create(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestActivityController_UpdateAndDelete_Extended tests Update and Delete handlers.
func TestActivityController_UpdateAndDelete_Extended(t *testing.T) {
	t.Run("Update bad JSON returns 400", func(t *testing.T) {
		// Arrange
		ctrl, _, _, _ := newActivityControllerFixture()
		req := withChiURLParam(httptest.NewRequest("PUT", "/api/actividades/a1", bytes.NewBufferString("{bad")), "actividadId", "a1")
		w := httptest.NewRecorder()

		// Act
		ctrl.Update(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Update success returns 200", func(t *testing.T) {
		// Arrange
		ctrl, _, activitySvc, _ := newActivityControllerFixture()
		input := models.ActivityInput{Titulo: "Clase 1 Editada"}
		activitySvc.On("UpdateActivity", mock.Anything, "a1", input).Return(&models.Activity{ID: "a1", Titulo: "Clase 1 Editada"}, nil)

		body, _ := json.Marshal(input)
		req := withChiURLParam(httptest.NewRequest("PUT", "/api/actividades/a1", bytes.NewBuffer(body)), "actividadId", "a1")
		w := httptest.NewRecorder()

		// Act
		ctrl.Update(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Delete success returns 204", func(t *testing.T) {
		// Arrange
		ctrl, _, activitySvc, _ := newActivityControllerFixture()
		activitySvc.On("DeleteActivity", mock.Anything, "a1").Return(nil)

		req := withChiURLParam(httptest.NewRequest("DELETE", "/api/actividades/a1", nil), "actividadId", "a1")
		w := httptest.NewRecorder()

		// Act
		ctrl.Delete(w, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Delete error returns 502", func(t *testing.T) {
		// Arrange
		ctrl, _, activitySvc, _ := newActivityControllerFixture()
		activitySvc.On("DeleteActivity", mock.Anything, "a1").Return(errors.New("db error"))

		req := withChiURLParam(httptest.NewRequest("DELETE", "/api/actividades/a1", nil), "actividadId", "a1")
		w := httptest.NewRecorder()

		// Act
		ctrl.Delete(w, req)

		// Assert
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})
}

// TestActivityController_ListSubmissions_Extended tests listing submissions for an activity.
func TestActivityController_ListSubmissions_Extended(t *testing.T) {
	t.Run("GetActivity error returns 404", func(t *testing.T) {
		// Arrange
		ctrl, _, activitySvc, _ := newActivityControllerFixture()
		activitySvc.On("GetActivity", mock.Anything, "a1").Return(nil, repoErrNotFound())

		req := withChiURLParam(httptest.NewRequest("GET", "/api/actividades/a1/entregas", nil), "actividadId", "a1")
		w := httptest.NewRecorder()

		// Act
		ctrl.ListSubmissions(w, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Success returns submissions list", func(t *testing.T) {
		// Arrange
		ctrl, _, activitySvc, submissionSvc := newActivityControllerFixture()
		act := &models.Activity{ID: "a1"}
		activitySvc.On("GetActivity", mock.Anything, "a1").Return(act, nil)
		submissionSvc.On("ListByActivity", mock.Anything, act, "").Return([]models.SubmissionStatusRow{
			{ParticipanteID: "p1", Estado: models.SubmissionEstadoSinEntregar},
		}, nil)

		req := withChiURLParam(httptest.NewRequest("GET", "/api/actividades/a1/entregas", nil), "actividadId", "a1")
		w := httptest.NewRecorder()

		// Act
		ctrl.ListSubmissions(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "sin_entregar")
	})

	t.Run("SubmissionService error returns 502", func(t *testing.T) {
		// Arrange
		ctrl, _, activitySvc, submissionSvc := newActivityControllerFixture()
		act := &models.Activity{ID: "a1"}
		activitySvc.On("GetActivity", mock.Anything, "a1").Return(act, nil)
		submissionSvc.On("ListByActivity", mock.Anything, act, "").Return(nil, errors.New("db error"))

		req := withChiURLParam(httptest.NewRequest("GET", "/api/actividades/a1/entregas", nil), "actividadId", "a1")
		w := httptest.NewRecorder()

		// Act
		ctrl.ListSubmissions(w, req)

		// Assert
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})
}
