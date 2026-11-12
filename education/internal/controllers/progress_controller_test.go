package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// HU-20 / auditoría 2026-11-11: progreso de un curso para admin y voluntario.
func TestProgressController_CourseProgress(t *testing.T) {
	t.Run("devuelve el progreso de cualquier curso existente", func(t *testing.T) {
		courseSvc, progressSvc := new(MockCourseService), new(MockProgressService)
		ctrl := controllers.NewProgressController(courseSvc, progressSvc)
		courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
		progressSvc.On("CalculateCourseProgress", mock.Anything, "c1").
			Return([]models.ParticipantProgress{{ParticipanteID: "p1", ActividadesTotal: 4, ActividadesAsistidas: 3, PorcentajeAsistencia: 75}}, nil)

		w := httptest.NewRecorder()
		ctrl.CourseProgress(w, withChiURLParam(httptest.NewRequest(http.MethodGet, "/api/cursos/c1/progreso", nil), "id", "c1"))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"participanteId":"p1"`)
		assert.Contains(t, w.Body.String(), `75`)
	})

	t.Run("404 si el curso no existe", func(t *testing.T) {
		courseSvc, progressSvc := new(MockCourseService), new(MockProgressService)
		ctrl := controllers.NewProgressController(courseSvc, progressSvc)
		courseSvc.On("VerifyExists", mock.Anything, "nope").Return(repositories.ErrNotFound)

		w := httptest.NewRecorder()
		ctrl.CourseProgress(w, withChiURLParam(httptest.NewRequest(http.MethodGet, "/api/cursos/nope/progreso", nil), "id", "nope"))

		assert.Equal(t, http.StatusNotFound, w.Code)
		progressSvc.AssertNotCalled(t, "CalculateCourseProgress", mock.Anything, mock.Anything)
	})
}
