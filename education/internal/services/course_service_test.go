package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCourseRepository struct {
	mock.Mock
}

func (m *MockCourseRepository) List(ctx context.Context) ([]models.Course, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *MockCourseRepository) Get(ctx context.Context, id string) (*models.Course, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *MockCourseRepository) Create(ctx context.Context, input models.CourseInput) (*models.Course, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *MockCourseRepository) Update(ctx context.Context, id string, input models.CourseInput) (*models.Course, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *MockCourseRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCourseRepository) Publish(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCourseRepository) Unpublish(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCourseRepository) ListByResponsibleTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	args := m.Called(ctx, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *MockCourseRepository) ListDraftsByResponsibleTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	args := m.Called(ctx, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *MockCourseRepository) ListAllDrafts(ctx context.Context) ([]models.Course, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Course), args.Error(1)
}

func validCourseInput() models.CourseInput {
	return models.CourseInput{
		Title:              "Curso de prueba",
		Description:        json.RawMessage(`{"nodeType":"document"}`),
		ResponsibleTeacher: "teacher-1",
		Mode:               "Presencial",
		Duration:           "4 semanas",
		TrainingAreaID:     "area-1",
		MaterialIDs:        []string{"mat-1"},
	}
}

func TestCourseService_ListCourses(t *testing.T) {
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)

	repo.On("List", mock.Anything).Return([]models.Course{{ID: "c1"}}, nil)

	courses, err := svc.ListCourses(context.Background())
	assert.NoError(t, err)
	assert.Len(t, courses, 1)
}

func TestCourseService_ListCourses_RepoError(t *testing.T) {
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)

	repo.On("List", mock.Anything).Return(nil, errors.New("contentful caído"))

	courses, err := svc.ListCourses(context.Background())
	assert.Error(t, err)
	assert.Nil(t, courses)
}

func TestCourseService_GetCourse_EmptyID(t *testing.T) {
	svc := services.NewCourseService(nil)

	course, err := svc.GetCourse(context.Background(), "  ")
	assert.Error(t, err)
	assert.Nil(t, course)
	assert.True(t, services.IsValidationError(err))
}

func TestCourseService_GetCourse_Success(t *testing.T) {
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)

	repo.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)

	course, err := svc.GetCourse(context.Background(), "c1")
	assert.NoError(t, err)
	assert.Equal(t, "c1", course.ID)
}

func TestCourseService_CreateCourse_ValidationErrors(t *testing.T) {
	svc := services.NewCourseService(nil)

	tests := []struct {
		name  string
		input models.CourseInput
	}{
		{"missing title", func() models.CourseInput { i := validCourseInput(); i.Title = ""; return i }()},
		{"missing description", func() models.CourseInput { i := validCourseInput(); i.Description = nil; return i }()},
		{"null description", func() models.CourseInput { i := validCourseInput(); i.Description = json.RawMessage("null"); return i }()},
		{"missing responsibleTeacher", func() models.CourseInput { i := validCourseInput(); i.ResponsibleTeacher = ""; return i }()},
		{"missing duration", func() models.CourseInput { i := validCourseInput(); i.Duration = ""; return i }()},
		{"missing trainingAreaId", func() models.CourseInput { i := validCourseInput(); i.TrainingAreaID = ""; return i }()},
		{"missing materialIds", func() models.CourseInput { i := validCourseInput(); i.MaterialIDs = nil; return i }()},
		{"invalid mode", func() models.CourseInput { i := validCourseInput(); i.Mode = "Remota"; return i }()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			course, err := svc.CreateCourse(context.Background(), tt.input)
			assert.Error(t, err)
			assert.Nil(t, course)
			assert.True(t, services.IsValidationError(err))
		})
	}
}

func TestCourseService_CreateCourse_Success(t *testing.T) {
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)

	input := validCourseInput()
	repo.On("Create", mock.Anything, input).Return(&models.Course{ID: "c1"}, nil)

	course, err := svc.CreateCourse(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, "c1", course.ID)
}

func TestCourseService_UpdateCourse_EmptyID(t *testing.T) {
	svc := services.NewCourseService(nil)

	course, err := svc.UpdateCourse(context.Background(), "", validCourseInput())
	assert.Error(t, err)
	assert.Nil(t, course)
}

func TestCourseService_DeleteCourse_EmptyID(t *testing.T) {
	svc := services.NewCourseService(nil)
	err := svc.DeleteCourse(context.Background(), "")
	assert.Error(t, err)
	assert.True(t, services.IsValidationError(err))
}

func TestCourseService_PublishCourse_EmptyID(t *testing.T) {
	svc := services.NewCourseService(nil)
	err := svc.PublishCourse(context.Background(), "")
	assert.Error(t, err)
}

func TestCourseService_UnpublishCourse_EmptyID(t *testing.T) {
	svc := services.NewCourseService(nil)
	err := svc.UnpublishCourse(context.Background(), "")
	assert.Error(t, err)
}

func TestCourseService_ListCoursesByTeacher_MissingTeacherID(t *testing.T) {
	svc := services.NewCourseService(nil)
	courses, err := svc.ListCoursesByTeacher(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, courses)
}

func TestCourseService_ListDraftCoursesByTeacher_MissingTeacherID(t *testing.T) {
	svc := services.NewCourseService(nil)
	courses, err := svc.ListDraftCoursesByTeacher(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, courses)
}

func TestCourseService_ListAllDraftCourses(t *testing.T) {
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)

	repo.On("ListAllDrafts", mock.Anything).Return([]models.Course{{ID: "draft-1"}}, nil)

	courses, err := svc.ListAllDraftCourses(context.Background())
	assert.NoError(t, err)
	assert.Len(t, courses, 1)
}

func TestCourseService_VerifyOwnership_EmptyID(t *testing.T) {
	svc := services.NewCourseService(nil)
	err := svc.VerifyOwnership(context.Background(), "", "teacher-1")
	assert.Error(t, err)
	assert.True(t, services.IsValidationError(err))
}

func TestCourseService_VerifyOwnership_NotFound(t *testing.T) {
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)

	repo.On("Get", mock.Anything, "c1").Return(nil, errors.New("no encontrado"))

	err := svc.VerifyOwnership(context.Background(), "c1", "teacher-1")
	assert.Error(t, err)
	assert.False(t, services.IsForbiddenError(err))
}

func TestCourseService_VerifyOwnership_WrongTeacher(t *testing.T) {
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)

	repo.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1", ResponsibleTeacher: "teacher-2"}, nil)

	err := svc.VerifyOwnership(context.Background(), "c1", "teacher-1")
	assert.Error(t, err)
	assert.True(t, services.IsForbiddenError(err))
}

func TestCourseService_VerifyOwnership_Success(t *testing.T) {
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)

	repo.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1", ResponsibleTeacher: "teacher-1"}, nil)

	err := svc.VerifyOwnership(context.Background(), "c1", "teacher-1")
	assert.NoError(t, err)
}

func TestCourseService_CreateCourseAsTeacher_MissingTeacherID(t *testing.T) {
	svc := services.NewCourseService(nil)
	course, err := svc.CreateCourseAsTeacher(context.Background(), "", validCourseInput())
	assert.Error(t, err)
	assert.Nil(t, course)
}

// CreateCourseAsTeacher fuerza responsibleTeacher al docente autenticado,
// ignorando cualquier valor que haya mandado el cliente en el body.
func TestCourseService_CreateCourseAsTeacher_ForcesResponsibleTeacher(t *testing.T) {
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)

	input := validCourseInput()
	input.ResponsibleTeacher = "otro-docente"

	repo.On("Create", mock.Anything, mock.MatchedBy(func(in models.CourseInput) bool {
		return in.ResponsibleTeacher == "teacher-1"
	})).Return(&models.Course{ID: "c1", ResponsibleTeacher: "teacher-1"}, nil)

	course, err := svc.CreateCourseAsTeacher(context.Background(), "teacher-1", input)
	assert.NoError(t, err)
	assert.Equal(t, "teacher-1", course.ResponsibleTeacher)
}

func TestCourseService_UpdateCourseAsTeacher_NotOwner(t *testing.T) {
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)

	repo.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1", ResponsibleTeacher: "teacher-2"}, nil)

	course, err := svc.UpdateCourseAsTeacher(context.Background(), "c1", "teacher-1", validCourseInput())
	assert.Error(t, err)
	assert.Nil(t, course)
	assert.True(t, services.IsForbiddenError(err))
}

// UpdateCourseAsTeacher no permite reasignar responsibleTeacher a otro id:
// se ignora el valor del body y se fuerza al docente autenticado.
func TestCourseService_UpdateCourseAsTeacher_CannotReassign(t *testing.T) {
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)

	repo.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1", ResponsibleTeacher: "teacher-1"}, nil)

	input := validCourseInput()
	input.ResponsibleTeacher = "otro-docente"

	repo.On("Update", mock.Anything, "c1", mock.MatchedBy(func(in models.CourseInput) bool {
		return in.ResponsibleTeacher == "teacher-1"
	})).Return(&models.Course{ID: "c1", ResponsibleTeacher: "teacher-1"}, nil)

	course, err := svc.UpdateCourseAsTeacher(context.Background(), "c1", "teacher-1", input)
	assert.NoError(t, err)
	assert.Equal(t, "teacher-1", course.ResponsibleTeacher)
}
