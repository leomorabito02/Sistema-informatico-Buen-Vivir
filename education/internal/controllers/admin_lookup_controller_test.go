package controllers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/models"
)

func TestEnrollmentController_ListByParticipant(t *testing.T) {
	svc := new(MockEnrollmentService)
	ctrl := controllers.NewEnrollmentController(svc)
	svc.On("ListByParticipant", mock.Anything, "p1").Return([]models.Enrollment{{ID: "e1"}}, nil)
	svc.On("ListByParticipant", mock.Anything, "p2").Return(nil, errors.New("firestore caído"))

	w := httptest.NewRecorder()
	ctrl.ListByParticipant(w, withChiURLParams(httptest.NewRequest("GET", "/api/participants/p1/inscripciones", nil), "id", "p1"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"e1"`)

	w = httptest.NewRecorder()
	ctrl.ListByParticipant(w, withChiURLParams(httptest.NewRequest("GET", "/api/participants/p2/inscripciones", nil), "id", "p2"))
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestEnrollmentController_GetBulkUserCourses(t *testing.T) {
	svc := new(MockEnrollmentService)
	ctrl := controllers.NewEnrollmentController(svc)
	svc.On("GetBulkUserCourses", mock.Anything, []string{"u1", "u2"}).Return([]models.UserCourseSummary{
		{UserID: "u1", Courses: []string{"Huerta"}},
	}, nil)
	svc.On("GetBulkUserCourses", mock.Anything, []string{"boom"}).Return(nil, errors.New("contentful caído"))

	post := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		ctrl.GetBulkUserCourses(w, httptest.NewRequest("POST", "/api/bulk-cursos-usuarios", strings.NewReader(body)))
		return w
	}

	w := post(`{"user_ids":["u1","u2"]}`)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Huerta")

	assert.Equal(t, http.StatusBadRequest, post(`{`).Code)
	assert.Equal(t, http.StatusBadGateway, post(`{"user_ids":["boom"]}`).Code)
}

func TestCourseController_ListByTeacherAdmin(t *testing.T) {
	svc := new(MockCourseService)
	ctrl := controllers.NewCourseController(svc)
	svc.On("ListCoursesByTeacher", mock.Anything, "t1").Return([]models.Course{{ID: "c1"}}, nil)
	svc.On("ListCoursesByTeacher", mock.Anything, "t2").Return(nil, errors.New("contentful caído"))

	w := httptest.NewRecorder()
	ctrl.ListByTeacherAdmin(w, withChiURLParam(httptest.NewRequest("GET", "/api/docentes/t1/cursos", nil), "id", "t1"))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	ctrl.ListByTeacherAdmin(w, withChiURLParam(httptest.NewRequest("GET", "/api/docentes/t2/cursos", nil), "id", "t2"))
	assert.NotEqual(t, http.StatusOK, w.Code)
}
