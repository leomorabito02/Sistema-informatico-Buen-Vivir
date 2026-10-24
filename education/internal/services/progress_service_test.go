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

// Valores límite del progreso (HU-20): ninguna actividad asistida (0 de N)
// y todas asistidas (N de N). Una asistencia con presente=false no suma.
func TestProgressService_CalculateCourseProgress_BoundaryValues(t *testing.T) {
	cases := []struct {
		name         string
		attendances  []models.Attendance
		wantAttended int
		wantPct      float64
	}{
		{
			name:         "0 de N: sin asistencias",
			attendances:  []models.Attendance{},
			wantAttended: 0,
			wantPct:      0.0,
		},
		{
			name: "0 de N: asistencias registradas como ausente",
			attendances: []models.Attendance{
				{ActividadID: "a1", Presente: false},
				{ActividadID: "a2", Presente: false},
				{ActividadID: "a3", Presente: false},
			},
			wantAttended: 0,
			wantPct:      0.0,
		},
		{
			name: "N de N: asistió a todas",
			attendances: []models.Attendance{
				{ActividadID: "a1", Presente: true},
				{ActividadID: "a2", Presente: true},
				{ActividadID: "a3", Presente: true},
			},
			wantAttended: 3,
			wantPct:      100.0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			activityRepo := new(MockActivityRepository)
			enrollmentRepo := new(MockEnrollmentRepository)
			attendanceRepo := new(MockAttendanceRepository)
			svc := services.NewProgressService(activityRepo, enrollmentRepo, attendanceRepo)

			activityRepo.On("ListByCourse", mock.Anything, "c1").Return([]models.Activity{{ID: "a1"}, {ID: "a2"}, {ID: "a3"}}, nil)
			enrollmentRepo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{{ParticipanteID: "p1"}}, nil)
			attendanceRepo.On("ListByParticipant", mock.Anything, "p1").Return(tc.attendances, nil)

			progress, err := svc.CalculateCourseProgress(context.Background(), "c1")

			assert.NoError(t, err)
			assert.Len(t, progress, 1)
			assert.Equal(t, 3, progress[0].ActividadesTotal)
			assert.Equal(t, tc.wantAttended, progress[0].ActividadesAsistidas)
			assert.Equal(t, tc.wantPct, progress[0].PorcentajeAsistencia)
		})
	}
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
