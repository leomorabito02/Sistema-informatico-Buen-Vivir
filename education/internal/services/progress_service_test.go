package services_test

import (
	"context"
	"errors"
	"testing"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Reusa MockActivityRepository, MockEnrollmentRepository y
// MockAttendanceRepository definidos en activity_service_test.go,
// enrollment_service_test.go y attendance_service_test.go.

func TestProgressService_CalculateCourseProgress_ActivityRepoError(t *testing.T) {
	activityRepo := new(MockActivityRepository)
	enrollmentRepo := new(MockEnrollmentRepository)
	attendanceRepo := new(MockAttendanceRepository)
	svc := services.NewProgressService(activityRepo, enrollmentRepo, attendanceRepo)

	activityRepo.On("ListByCourse", mock.Anything, "c1").Return(nil, errors.New("firestore caído"))

	progress, err := svc.CalculateCourseProgress(context.Background(), "c1")
	assert.Error(t, err)
	assert.Nil(t, progress)
}

func TestProgressService_CalculateCourseProgress_EnrollmentRepoError(t *testing.T) {
	activityRepo := new(MockActivityRepository)
	enrollmentRepo := new(MockEnrollmentRepository)
	attendanceRepo := new(MockAttendanceRepository)
	svc := services.NewProgressService(activityRepo, enrollmentRepo, attendanceRepo)

	activityRepo.On("ListByCourse", mock.Anything, "c1").Return([]models.Activity{{ID: "a1"}}, nil)
	enrollmentRepo.On("ListByCourse", mock.Anything, "c1").Return(nil, errors.New("firestore caído"))

	progress, err := svc.CalculateCourseProgress(context.Background(), "c1")
	assert.Error(t, err)
	assert.Nil(t, progress)
}

func TestProgressService_CalculateCourseProgress_NoActivities(t *testing.T) {
	activityRepo := new(MockActivityRepository)
	enrollmentRepo := new(MockEnrollmentRepository)
	attendanceRepo := new(MockAttendanceRepository)
	svc := services.NewProgressService(activityRepo, enrollmentRepo, attendanceRepo)

	activityRepo.On("ListByCourse", mock.Anything, "c1").Return([]models.Activity{}, nil)
	enrollmentRepo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{{ParticipanteID: "p1"}}, nil)
	attendanceRepo.On("ListByParticipant", mock.Anything, "p1").Return([]models.Attendance{}, nil)

	progress, err := svc.CalculateCourseProgress(context.Background(), "c1")
	assert.NoError(t, err)
	assert.Len(t, progress, 1)
	// Sin actividades, el porcentaje no puede calcularse (evita división por cero): queda en 0.
	assert.Equal(t, 0.0, progress[0].PorcentajeAsistencia)
}

func TestProgressService_CalculateCourseProgress_ComputesPercentage(t *testing.T) {
	activityRepo := new(MockActivityRepository)
	enrollmentRepo := new(MockEnrollmentRepository)
	attendanceRepo := new(MockAttendanceRepository)
	svc := services.NewProgressService(activityRepo, enrollmentRepo, attendanceRepo)

	activities := []models.Activity{{ID: "a1"}, {ID: "a2"}}
	activityRepo.On("ListByCourse", mock.Anything, "c1").Return(activities, nil)
	enrollmentRepo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{{ParticipanteID: "p1"}}, nil)

	attendances := []models.Attendance{
		{ActividadID: "a1", Presente: true},
		{ActividadID: "a2", Presente: false},
		// Asistencia de una actividad que no pertenece a este curso: no debe contar.
		{ActividadID: "a-otro-curso", Presente: true},
	}
	attendanceRepo.On("ListByParticipant", mock.Anything, "p1").Return(attendances, nil)

	progress, err := svc.CalculateCourseProgress(context.Background(), "c1")
	assert.NoError(t, err)
	assert.Len(t, progress, 1)
	assert.Equal(t, "p1", progress[0].ParticipanteID)
	assert.Equal(t, 2, progress[0].ActividadesTotal)
	assert.Equal(t, 1, progress[0].ActividadesAsistidas)
	assert.Equal(t, 50.0, progress[0].PorcentajeAsistencia)
}

func TestProgressService_CalculateCourseProgress_AttendanceRepoError(t *testing.T) {
	activityRepo := new(MockActivityRepository)
	enrollmentRepo := new(MockEnrollmentRepository)
	attendanceRepo := new(MockAttendanceRepository)
	svc := services.NewProgressService(activityRepo, enrollmentRepo, attendanceRepo)

	activityRepo.On("ListByCourse", mock.Anything, "c1").Return([]models.Activity{{ID: "a1"}}, nil)
	enrollmentRepo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{{ParticipanteID: "p1"}}, nil)
	attendanceRepo.On("ListByParticipant", mock.Anything, "p1").Return(nil, errors.New("firestore caído"))

	progress, err := svc.CalculateCourseProgress(context.Background(), "c1")
	assert.Error(t, err)
	assert.Nil(t, progress)
}
