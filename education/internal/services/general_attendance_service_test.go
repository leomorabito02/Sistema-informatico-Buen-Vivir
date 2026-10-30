package services_test

import (
	"context"
	"errors"
	"testing"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockGeneralAttendanceRepository struct {
	mock.Mock
}

func (m *MockGeneralAttendanceRepository) ListByDate(ctx context.Context, fecha string) ([]models.GeneralAttendance, error) {
	args := m.Called(ctx, fecha)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GeneralAttendance), args.Error(1)
}

func (m *MockGeneralAttendanceRepository) RegisterBatch(ctx context.Context, fecha string, inputs []models.GeneralAttendanceInput) error {
	return m.Called(ctx, fecha, inputs).Error(0)
}

func newGeneralAttendanceFixture(t *testing.T, auth *fakeAuthServer) (*MockGeneralAttendanceRepository, services.GeneralAttendanceService) {
	repo := new(MockGeneralAttendanceRepository)
	return repo, services.NewGeneralAttendanceService(repo, auth.start(t))
}

func centroConParticipantes() *fakeAuthServer {
	return &fakeAuthServer{
		participants: []models.AuthParticipant{
			{ParticipantID: "p1", UserID: "u1", ApprovalStatus: "accepted"},
			{ParticipantID: "p2", UserID: "u2", ApprovalStatus: "accepted"},
			{ParticipantID: "p3", UserID: "u3", ApprovalStatus: "pending"},
			{ParticipantID: "p4", UserID: "u4"}, // approval_status vacío se tolera
		},
		users: []models.AuthUser{
			{UserID: "u1", FirstName: "Ana", LastName: "Pérez"},
			{UserID: "u2", FirstName: "Beto", LastName: "Gómez"},
		},
	}
}

func TestGeneralAttendance_ListByDate_CombinesParticipantsUsersAndRecords(t *testing.T) {
	auth := centroConParticipantes()
	repo, svc := newGeneralAttendanceFixture(t, auth)
	repo.On("ListByDate", mock.Anything, "2026-09-28").Return([]models.GeneralAttendance{
		{ParticipanteID: "p1", Presente: true, Observacion: "llegó tarde"},
		{ParticipanteID: "p2", Presente: false},
	}, nil)

	views, err := svc.ListByDate(context.Background(), " 2026-09-28 ", "tok")

	require.NoError(t, err)
	require.Len(t, views, 3, "el participante pendiente (p3) no aparece")
	assert.Equal(t, "tok", auth.receivedToken)

	assert.Equal(t, "p1", views[0].ParticipanteID)
	assert.Equal(t, "Ana", views[0].Nombre)
	assert.Equal(t, "Pérez", views[0].Apellido)
	require.NotNil(t, views[0].Presente)
	assert.True(t, *views[0].Presente)
	assert.Equal(t, "llegó tarde", views[0].Observacion)

	require.NotNil(t, views[1].Presente, "ausente cargado es false, no nil")
	assert.False(t, *views[1].Presente)

	assert.Equal(t, "p4", views[2].ParticipanteID)
	assert.Nil(t, views[2].Presente, "sin registro para la fecha")
	assert.Empty(t, views[2].Nombre, "sin usuario asociado queda solo el id")
}

func TestGeneralAttendance_ListByDate_UsersFailureIsNotFatal(t *testing.T) {
	auth := centroConParticipantes()
	auth.failUsers = true
	repo, svc := newGeneralAttendanceFixture(t, auth)
	repo.On("ListByDate", mock.Anything, "2026-09-28").Return([]models.GeneralAttendance{}, nil)

	views, err := svc.ListByDate(context.Background(), "2026-09-28", "tok")

	require.NoError(t, err)
	require.Len(t, views, 3)
	assert.Empty(t, views[0].Nombre)
}

func TestGeneralAttendance_ListByDate_InvalidDate(t *testing.T) {
	_, svc := newGeneralAttendanceFixture(t, centroConParticipantes())

	for _, fecha := range []string{"", "28/09/2026", "2026-13-01"} {
		_, err := svc.ListByDate(context.Background(), fecha, "tok")
		assert.True(t, services.IsValidationError(err), fecha)
	}
}

func TestGeneralAttendance_ListByDate_Errors(t *testing.T) {
	auth := centroConParticipantes()
	auth.failList = true
	_, svc := newGeneralAttendanceFixture(t, auth)
	_, err := svc.ListByDate(context.Background(), "2026-09-28", "tok")
	require.Error(t, err)
	assert.False(t, services.IsValidationError(err))

	boom := errors.New("firestore caído")
	repo, svc := newGeneralAttendanceFixture(t, centroConParticipantes())
	repo.On("ListByDate", mock.Anything, "2026-09-28").Return(nil, boom)
	_, err = svc.ListByDate(context.Background(), "2026-09-28", "tok")
	assert.ErrorIs(t, err, boom)
}

func TestGeneralAttendance_RegisterBatch_Success(t *testing.T) {
	repo, svc := newGeneralAttendanceFixture(t, centroConParticipantes())
	inputs := []models.GeneralAttendanceInput{
		{ParticipanteID: "p1", Presente: true},
		{ParticipanteID: "p4", Presente: false, Observacion: "avisó"},
	}
	repo.On("RegisterBatch", mock.Anything, "2026-09-28", inputs).Return(nil)

	err := svc.RegisterBatch(context.Background(), "2026-09-28", inputs, "tok")

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGeneralAttendance_RegisterBatch_Validations(t *testing.T) {
	repo, svc := newGeneralAttendanceFixture(t, centroConParticipantes())
	ctx := context.Background()

	cases := map[string]struct {
		fecha  string
		inputs []models.GeneralAttendanceInput
		msg    string
	}{
		"fecha inválida":           {"2026/09/28", []models.GeneralAttendanceInput{{ParticipanteID: "p1"}}, "fecha inválida"},
		"lote vacío":               {"2026-09-28", nil, "al menos un registro"},
		"sin participanteId":       {"2026-09-28", []models.GeneralAttendanceInput{{ParticipanteID: " "}}, "participanteId"},
		"participante inexistente": {"2026-09-28", []models.GeneralAttendanceInput{{ParticipanteID: "p1"}, {ParticipanteID: "zzz"}}, "zzz no existe"},
		"participante pendiente":   {"2026-09-28", []models.GeneralAttendanceInput{{ParticipanteID: "p3"}}, "p3 no está activo"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := svc.RegisterBatch(ctx, tc.fecha, tc.inputs, "tok")
			require.True(t, services.IsValidationError(err))
			assert.Contains(t, err.Error(), tc.msg)
		})
	}
	repo.AssertNotCalled(t, "RegisterBatch", mock.Anything, mock.Anything, mock.Anything)
}

func TestGeneralAttendance_RegisterBatch_Errors(t *testing.T) {
	auth := centroConParticipantes()
	auth.failGet = true
	repo, svc := newGeneralAttendanceFixture(t, auth)
	err := svc.RegisterBatch(context.Background(), "2026-09-28", []models.GeneralAttendanceInput{{ParticipanteID: "p1"}}, "tok")
	require.Error(t, err)
	assert.False(t, services.IsValidationError(err))
	repo.AssertNotCalled(t, "RegisterBatch", mock.Anything, mock.Anything, mock.Anything)

	boom := errors.New("firestore caído")
	repo, svc = newGeneralAttendanceFixture(t, centroConParticipantes())
	repo.On("RegisterBatch", mock.Anything, "2026-09-28", mock.Anything).Return(boom)
	err = svc.RegisterBatch(context.Background(), "2026-09-28", []models.GeneralAttendanceInput{{ParticipanteID: "p1"}}, "tok")
	assert.ErrorIs(t, err, boom)
}
