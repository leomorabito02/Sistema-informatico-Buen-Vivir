package controllers

import (
	"context"
	"errors"
	"testing"

	"buenvivir/education/api/proto/educationpb"
	"buenvivir/education/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubSkillService struct {
	granted     []models.GrantedSkill
	partSkills  []models.ParticipantSkillView
	err         error
	lastPartArg string
}

func (s *stubSkillService) SearchCatalog(context.Context, string) ([]models.Skill, error) {
	return nil, nil
}
func (s *stubSkillService) GetCourseSkills(context.Context, string) (*models.CourseSkills, error) {
	return nil, nil
}
func (s *stubSkillService) SetActivitySkills(context.Context, *models.Activity, []models.SkillInput) ([]models.Skill, error) {
	return nil, nil
}
func (s *stubSkillService) ListParticipantSkills(_ context.Context, participanteID string) ([]models.ParticipantSkillView, error) {
	s.lastPartArg = participanteID
	return s.partSkills, s.err
}
func (s *stubSkillService) ListGrantedSkills(context.Context) ([]models.GrantedSkill, error) {
	return s.granted, s.err
}

func TestGrpcEducationController_GetGrantedSkills(t *testing.T) {
	// Arrange
	svc := &stubSkillService{
		granted: []models.GrantedSkill{
			{ParticipanteID: "p-1", Nombre: "Soldadura"},
			{ParticipanteID: "p-2", Nombre: "Carpintería"},
		},
	}
	ctrl := NewGrpcEducationController(svc)

	// Act
	resp, err := ctrl.GetGrantedSkills(context.Background(), &educationpb.GetGrantedSkillsRequest{})

	// Assert
	require.NoError(t, err)
	require.Len(t, resp.GetSkills(), 2)
	assert.Equal(t, "p-1", resp.GetSkills()[0].GetParticipanteId())
	assert.Equal(t, "Soldadura", resp.GetSkills()[0].GetNombre())
}

func TestGrpcEducationController_GetParticipantSkills(t *testing.T) {
	// Arrange
	svc := &stubSkillService{
		partSkills: []models.ParticipantSkillView{
			{Skill: models.Skill{Nombre: "Trabajo en Equipo"}},
			{Skill: models.Skill{Nombre: "Ofimática"}},
		},
	}
	ctrl := NewGrpcEducationController(svc)

	// Act
	resp, err := ctrl.GetParticipantSkills(context.Background(), &educationpb.GetParticipantSkillsRequest{ParticipanteId: "p-10"})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "p-10", svc.lastPartArg)
	assert.Equal(t, []string{"Trabajo en Equipo", "Ofimática"}, resp.GetNombres())
}

func TestGrpcEducationController_ValidationAndErrors(t *testing.T) {
	ctrl := NewGrpcEducationController(&stubSkillService{err: errors.New("firestore down")})

	_, err := ctrl.GetParticipantSkills(context.Background(), &educationpb.GetParticipantSkillsRequest{ParticipanteId: "  "})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = ctrl.GetGrantedSkills(context.Background(), &educationpb.GetGrantedSkillsRequest{})
	assert.Equal(t, codes.Internal, status.Code(err))
}
