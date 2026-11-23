package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/config"
	"buenvivir/education/internal/controllers"
)

type mockJWTService struct {
	mock.Mock
}

func (m *mockJWTService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(jwt.MapClaims), args.Error(1)
}

func newTestControllers(cfg *config.Config) Controllers {
	return Controllers{
		Health:            controllers.NewHealthController(cfg),
		Course:            controllers.NewCourseController(nil),
		TrainingArea:      controllers.NewTrainingAreaController(nil),
		Material:          controllers.NewMaterialController(nil),
		Docente:           controllers.NewDocenteController(nil, nil, nil, nil, nil, nil),
		Enrollment:        controllers.NewEnrollmentController(nil),
		Attendance:        controllers.NewAttendanceController(nil, nil),
		GeneralAttendance: controllers.NewGeneralAttendanceController(nil),
		Activity:          controllers.NewActivityController(nil, nil, nil),
		Submission:        controllers.NewSubmissionController(nil, nil, nil),
		Skill:             controllers.NewSkillController(nil, nil, nil, nil, nil),
		Progress:          controllers.NewProgressController(nil, nil),
	}
}

// TestRegisterRoutes_PublicAndProtectedEndpoints tests route mapping and auth requirements using AAA.
func TestRegisterRoutes_PublicAndProtectedEndpoints(t *testing.T) {
	// Arrange
	r := chi.NewRouter()
	jwtSvc := new(mockJWTService)
	cfg := &config.Config{
		AppEnv:                config.AppEnvTesting,
		ContentfulEnvironment: config.ContentfulEnvTesting,
	}

	ctrls := newTestControllers(cfg)
	RegisterRoutes(r, jwtSvc, ctrls)

	t.Run("Public endpoint: GET /api/salud returns 200 without token", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/api/salud", nil)
		rec := httptest.NewRecorder()

		// Act
		r.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"status":"UP"`)
	})

	t.Run("Protected endpoint: GET /api/cursos without token returns 401 Unauthorized", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/api/cursos", nil)
		rec := httptest.NewRecorder()

		// Act
		r.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// TestRegisterRoutes_RBAC_DecisionTable validates route access permissions using a Decision Table.
func TestRegisterRoutes_RBAC_DecisionTable(t *testing.T) {
	// Arrange
	r := chi.NewRouter()
	jwtSvc := new(mockJWTService)
	cfg := &config.Config{
		AppEnv:                config.AppEnvTesting,
		ContentfulEnvironment: config.ContentfulEnvTesting,
	}

	ctrls := newTestControllers(cfg)
	RegisterRoutes(r, jwtSvc, ctrls)

	decisionMatrix := []struct {
		name         string
		method       string
		path         string
		role         string
		isExternal   bool
		expectedCode int
	}{
		{
			name:         "Admin draft courses: role_participant denied with 403",
			method:       http.MethodGet,
			path:         "/api/cursos/borradores",
			role:         "role_participant",
			isExternal:   false,
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "Teacher routes: external teacher rejected with 403",
			method:       http.MethodGet,
			path:         "/api/docente/cursos",
			role:         "role_teacher",
			isExternal:   true,
			expectedCode: http.StatusForbidden,
		},
		// HU-20: progreso de cualquier curso, solo admin y voluntario.
		{
			name:         "Course progress: role_participant denied with 403",
			method:       http.MethodGet,
			path:         "/api/cursos/c1/progreso",
			role:         "role_participant",
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "Course progress: role_teacher denied with 403 (usa /api/docente)",
			method:       http.MethodGet,
			path:         "/api/cursos/c1/progreso",
			role:         "role_teacher",
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "Course progress: role_company denied with 403",
			method:       http.MethodGet,
			path:         "/api/cursos/c1/progreso",
			role:         "role_company",
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "Participant routes: role_teacher rejected with 403",
			method:       http.MethodGet,
			path:         "/api/participante/cursos",
			role:         "role_teacher",
			isExternal:   false,
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tt := range decisionMatrix {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			claims := jwt.MapClaims{
				"sub":         "user-test",
				"role":        tt.role,
				"is_external": tt.isExternal,
				"type":        "access",
			}
			jwtSvc.On("ValidateToken", "valid-token").Return(claims, nil).Once()

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			rec := httptest.NewRecorder()

			// Act
			r.ServeHTTP(rec, req)

			// Assert
			require.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}
