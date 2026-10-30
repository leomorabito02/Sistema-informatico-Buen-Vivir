package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
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
	return m.Called(ctx, id).Error(0)
}

func (m *MockCourseRepository) Unpublish(ctx context.Context, id string) error {
	ret := m.Called(ctx, id)
	return ret.Error(0)
}

func (m *MockCourseRepository) ListByResponsibleTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	args := m.Called(ctx, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *MockCourseRepository) ListDraftsByResponsibleTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	ret := m.Called(ctx, teacherID)
	if val := ret.Get(0); val != nil {
		return val.([]models.Course), ret.Error(1)
	}
	return nil, ret.Error(1)
}

func (m *MockCourseRepository) ExistsInManagement(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockCourseRepository) ListAllDrafts(ctx context.Context) ([]models.Course, error) {
	callArgs := m.Called(ctx)
	if v := callArgs.Get(0); v != nil {
		return v.([]models.Course), callArgs.Error(1)
	}
	return nil, callArgs.Error(1)
}

func (m *MockCourseRepository) ListByResponsibleTeachers(ctx context.Context, teacherIDs []string) ([]models.Course, error) {
	args := m.Called(ctx, teacherIDs)
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

func TestCourseService_VerifyExists(t *testing.T) {
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)

	repo.On("ExistsInManagement", mock.Anything, "borrador").Return(true, nil)
	repo.On("ExistsInManagement", mock.Anything, "no-existe").Return(false, nil)
	repo.On("ExistsInManagement", mock.Anything, "roto").Return(false, errors.New("contentful caído"))

	assert.NoError(t, svc.VerifyExists(context.Background(), "borrador"))
	assert.True(t, repositories.IsNotFound(svc.VerifyExists(context.Background(), "no-existe")))
	assert.Error(t, svc.VerifyExists(context.Background(), "roto"))
	assert.True(t, services.IsValidationError(svc.VerifyExists(context.Background(), " ")))
}

func TestCourseService_UpdateCourse_Success(t *testing.T) {
	// Arrange
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)
	input := validCourseInput()
	expectedCourse := &models.Course{ID: "c1", Title: input.Title}
	repo.On("Update", mock.Anything, "c1", input).Return(expectedCourse, nil)

	// Act
	course, err := svc.UpdateCourse(context.Background(), "c1", input)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "c1", course.ID)
	repo.AssertExpectations(t)
}

func TestCourseService_UpdateCourse_RepoError(t *testing.T) {
	// Arrange
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)
	input := validCourseInput()
	repo.On("Update", mock.Anything, "c1", input).Return(nil, errors.New("update error"))

	// Act
	course, err := svc.UpdateCourse(context.Background(), "c1", input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, course)
	repo.AssertExpectations(t)
}

// DeleteCourse es una baja lógica (HU-16): no borra en Contentful.
func TestCourseService_DeleteCourse_Success(t *testing.T) {
	// Arrange
	repo := new(MockCourseRepository)
	enrollments := new(MockEnrollmentRepository)
	svc := services.NewCourseServiceConBaja(repo, nil, &fakeBajas{}, enrollments)
	repo.On("ExistsInManagement", mock.Anything, "c1").Return(true, nil)
	repo.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)
	enrollments.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{}, nil)

	// Act
	err := svc.DeleteCourse(context.Background(), "c1")

	// Assert
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestCourseService_DeleteCourse_RepoError(t *testing.T) {
	// Arrange
	repo := new(MockCourseRepository)
	svc := services.NewCourseServiceConBaja(repo, nil, &fakeBajas{}, new(MockEnrollmentRepository))
	repo.On("ExistsInManagement", mock.Anything, "c1").Return(false, errors.New("contentful error"))

	// Act
	err := svc.DeleteCourse(context.Background(), "c1")

	// Assert
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestCourseService_PublishCourse_Success(t *testing.T) {
	// Arrange
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)
	repo.On("Publish", mock.Anything, "c1").Return(nil)

	// Act
	err := svc.PublishCourse(context.Background(), "c1")

	// Assert
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCourseService_PublishCourse_RepoError(t *testing.T) {
	// Arrange
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)
	repo.On("Publish", mock.Anything, "c1").Return(errors.New("publish error"))

	// Act
	err := svc.PublishCourse(context.Background(), "c1")

	// Assert
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestCourseService_UnpublishCourse_Success(t *testing.T) {
	// Arrange
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)
	repo.On("Unpublish", mock.Anything, "c1").Return(nil)

	// Act
	err := svc.UnpublishCourse(context.Background(), "c1")

	// Assert
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCourseService_UnpublishCourse_RepoError(t *testing.T) {
	// Arrange
	repo := new(MockCourseRepository)
	kvSync := new(MockKVSyncService)
	svc := services.NewCourseServiceWithEvents(repo, nil, kvSync)
	repo.On("Unpublish", mock.Anything, "c1").Return(errors.New("unpublish error"))

	// Act
	err := svc.UnpublishCourse(context.Background(), "c1")

	// Assert
	assert.Error(t, err)
	repo.AssertExpectations(t)
	kvSync.AssertNotCalled(t, "SyncCourseStatsAsync")
}

func TestCourseService_ListCoursesByTeacher_Success(t *testing.T) {
	// Arrange
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)
	expected := []models.Course{{ID: "c1", ResponsibleTeacher: "teacher-1"}}
	repo.On("ListByResponsibleTeacher", mock.Anything, "teacher-1").Return(expected, nil)

	// Act
	courses, err := svc.ListCoursesByTeacher(context.Background(), "teacher-1")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, courses)
	repo.AssertExpectations(t)
}

func TestCourseService_ListDraftCoursesByTeacher_Success(t *testing.T) {
	// Arrange
	repo := new(MockCourseRepository)
	svc := services.NewCourseService(repo)
	expected := []models.Course{{ID: "c1", ResponsibleTeacher: "teacher-1"}}
	repo.On("ListDraftsByResponsibleTeacher", mock.Anything, "teacher-1").Return(expected, nil)

	// Act
	courses, err := svc.ListDraftCoursesByTeacher(context.Background(), "teacher-1")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, courses)
	repo.AssertExpectations(t)
}

type MockKVSyncService struct {
	mock.Mock
}

func (m *MockKVSyncService) SyncCourseStatsAsync() {
	m.Called()
}

func TestCourseService_KVSync_CalledOnMutations(t *testing.T) {
	// Arrange
	repo := new(MockCourseRepository)
	kvSync := new(MockKVSyncService)
	enrollments := new(MockEnrollmentRepository)
	svc := services.NewCourseServiceConBaja(repo, nil, &fakeBajas{}, enrollments, kvSync)
	ctx := context.Background()

	input := validCourseInput()
	course := &models.Course{ID: "c1", Title: input.Title, ResponsibleTeacher: input.ResponsibleTeacher}

	repo.On("Create", mock.Anything, input).Return(course, nil)
	repo.On("Update", mock.Anything, "c1", input).Return(course, nil)
	repo.On("ExistsInManagement", mock.Anything, "c1").Return(true, nil)
	enrollments.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{}, nil)
	repo.On("Publish", mock.Anything, "c1").Return(nil)
	repo.On("Unpublish", mock.Anything, "c1").Return(nil)
	repo.On("Get", mock.Anything, "c1").Return(course, nil)
	kvSync.On("SyncCourseStatsAsync").Times(7)

	// Act
	_, errCreate := svc.CreateCourse(ctx, input)
	_, errUpdate := svc.UpdateCourse(ctx, "c1", input)
	errDelete := svc.DeleteCourse(ctx, "c1")
	errPublish := svc.PublishCourse(ctx, "c1")
	errUnpublish := svc.UnpublishCourse(ctx, "c1")
	_, errCreateTeacher := svc.CreateCourseAsTeacher(ctx, "teacher-1", input)
	_, errUpdateTeacher := svc.UpdateCourseAsTeacher(ctx, "c1", "teacher-1", input)

	// Assert
	assert.NoError(t, errCreate)
	assert.NoError(t, errUpdate)
	assert.NoError(t, errDelete)
	assert.NoError(t, errPublish)
	assert.NoError(t, errUnpublish)
	assert.NoError(t, errCreateTeacher)
	assert.NoError(t, errUpdateTeacher)
	kvSync.AssertExpectations(t)
}

func TestCourseService_CreateCourseAsTeacher_ValidationAndErrors(t *testing.T) {
	t.Run("Empty teacherID returns validation error", func(t *testing.T) {
		// Arrange
		repo := new(MockCourseRepository)
		svc := services.NewCourseService(repo)

		// Act
		_, err := svc.CreateCourseAsTeacher(context.Background(), "  ", validCourseInput())

		// Assert
		assert.Error(t, err)
		assert.True(t, services.IsValidationError(err))
	})

	t.Run("Invalid course input returns validation error", func(t *testing.T) {
		// Arrange
		repo := new(MockCourseRepository)
		svc := services.NewCourseService(repo)
		input := validCourseInput()
		input.Title = " "

		// Act
		_, err := svc.CreateCourseAsTeacher(context.Background(), "teacher-1", input)

		// Assert
		assert.Error(t, err)
		assert.True(t, services.IsValidationError(err))
	})

	t.Run("Repository create error propagated", func(t *testing.T) {
		// Arrange
		repo := new(MockCourseRepository)
		svc := services.NewCourseService(repo)
		input := validCourseInput()
		repo.On("Create", mock.Anything, input).Return(nil, errors.New("db error"))

		// Act
		_, err := svc.CreateCourseAsTeacher(context.Background(), "teacher-1", input)

		// Assert
		assert.Error(t, err)
	})
}

func TestCourseService_UpdateCourseAsTeacher_ValidationAndErrors(t *testing.T) {
	t.Run("Ownership error propagated", func(t *testing.T) {
		// Arrange
		repo := new(MockCourseRepository)
		svc := services.NewCourseService(repo)
		repo.On("Get", mock.Anything, "c1").Return(nil, repositories.ErrNotFound)

		// Act
		_, err := svc.UpdateCourseAsTeacher(context.Background(), "c1", "teacher-1", validCourseInput())

		// Assert
		assert.Error(t, err)
	})

	t.Run("Invalid course input returns validation error", func(t *testing.T) {
		// Arrange
		repo := new(MockCourseRepository)
		svc := services.NewCourseService(repo)
		course := &models.Course{ID: "c1", ResponsibleTeacher: "teacher-1"}
		repo.On("Get", mock.Anything, "c1").Return(course, nil)
		input := validCourseInput()
		input.Title = " "

		// Act
		_, err := svc.UpdateCourseAsTeacher(context.Background(), "c1", "teacher-1", input)

		// Assert
		assert.Error(t, err)
		assert.True(t, services.IsValidationError(err))
	})

	t.Run("Repository update error propagated", func(t *testing.T) {
		// Arrange
		repo := new(MockCourseRepository)
		svc := services.NewCourseService(repo)
		course := &models.Course{ID: "c1", ResponsibleTeacher: "teacher-1"}
		input := validCourseInput()
		repo.On("Get", mock.Anything, "c1").Return(course, nil)
		repo.On("Update", mock.Anything, "c1", input).Return(nil, errors.New("db error"))

		// Act
		_, err := svc.UpdateCourseAsTeacher(context.Background(), "c1", "teacher-1", input)

		// Assert
		assert.Error(t, err)
	})
}

// TestCourseService_EmptyID_Validation_DecisionTable tests validation error when course ID is empty or whitespace.
func TestCourseService_EmptyID_Validation_DecisionTable(t *testing.T) {
	svc := services.NewCourseService(new(MockCourseRepository))

	t.Run("UpdateCourse with empty ID returns validation error", func(t *testing.T) {
		// Arrange & Act
		_, err := svc.UpdateCourse(context.Background(), "   ", validCourseInput())

		// Assert
		assert.Error(t, err)
		assert.True(t, services.IsValidationError(err))
	})

	t.Run("DeleteCourse with empty ID returns validation error", func(t *testing.T) {
		// Arrange & Act
		err := svc.DeleteCourse(context.Background(), "")

		// Assert
		assert.Error(t, err)
		assert.True(t, services.IsValidationError(err))
	})

	t.Run("PublishCourse with empty ID returns validation error", func(t *testing.T) {
		// Arrange & Act
		err := svc.PublishCourse(context.Background(), "  ")

		// Assert
		assert.Error(t, err)
		assert.True(t, services.IsValidationError(err))
	})

	t.Run("UpdateCourse with invalid input returns validation error", func(t *testing.T) {
		// Arrange
		repo := new(MockCourseRepository)
		svcWithRepo := services.NewCourseService(repo)
		badInput := models.CourseInput{Title: ""} // Invalid empty title

		// Act
		_, err := svcWithRepo.UpdateCourse(context.Background(), "c1", badInput)

		// Assert
		assert.Error(t, err)
		assert.True(t, services.IsValidationError(err))
	})
}


