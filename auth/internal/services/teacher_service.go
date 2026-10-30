package services

import (
	"context"
	"errors"
	"time"

	"auth/internal/dto"
	"auth/internal/middlewares"
	"auth/internal/models"
	"auth/internal/repositories"
	"auth/internal/utils"
)

type TeacherService interface {
	CreateTeacher(ctx context.Context, req *dto.CreateTeacherRequest) (*dto.TeacherResponse, error)
	GetTeacherByID(ctx context.Context, id string) (*dto.TeacherResponse, error)
	GetAllTeachers(ctx context.Context) ([]dto.TeacherResponse, error)
	UpdateTeacher(ctx context.Context, id string, req *dto.UpdateTeacherRequest) (*dto.TeacherResponse, error)
	DeleteTeacher(ctx context.Context, id string) error
}

type teacherService struct {
	teacherRepo repositories.TeacherRepository
}

func NewTeacherService(teacherRepo repositories.TeacherRepository) TeacherService {
	return &teacherService{teacherRepo: teacherRepo}
}

func (s *teacherService) CreateTeacher(ctx context.Context, req *dto.CreateTeacherRequest) (*dto.TeacherResponse, error) {
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	status := req.ApprovalStatus
	if status == "" || !canApprove(ctx, "role_teacher") {
		status = "pending"
	}

	teacher := &models.Teacher{
		TeacherID:      "tch-" + utils.GenerateRandomString(10),
		UserID:         req.UserID,
		Specialization: req.Specialization,
		IsExternal:     req.IsExternal,
		ApprovalStatus: status,
	}

	if err := s.teacherRepo.Create(ctx, teacher); err != nil {
		return nil, err
	}

	return dto.ToTeacherResponse(teacher), nil
}

func (s *teacherService) GetTeacherByID(ctx context.Context, id string) (*dto.TeacherResponse, error) {
	teacher, err := s.teacherRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_teacher" {
		reqIsExternal, _ := ctx.Value(middlewares.IsExternalKey).(bool)
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if reqIsExternal && teacher.UserID != reqUserID {
			return nil, errors.New("forbidden: external teachers can only access their own profile")
		}
	}

	return dto.ToTeacherResponse(teacher), nil
}

func (s *teacherService) GetAllTeachers(ctx context.Context) ([]dto.TeacherResponse, error) {
	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_teacher" {
		reqIsExternal, _ := ctx.Value(middlewares.IsExternalKey).(bool)
		if reqIsExternal {
			return nil, errors.New("forbidden: external teachers cannot list all teachers")
		}
	}

	teachers, err := s.teacherRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	var resp []dto.TeacherResponse
	for _, t := range teachers {
		resp = append(resp, *dto.ToTeacherResponse(&t))
	}
	return resp, nil
}

func (s *teacherService) UpdateTeacher(ctx context.Context, id string, req *dto.UpdateTeacherRequest) (*dto.TeacherResponse, error) {
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	teacher, err := s.teacherRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_teacher" {
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if teacher.UserID != reqUserID {
			return nil, errors.New("forbidden: access denied to edit other teacher profiles")
		}
	}

	if req.Specialization != "" {
		teacher.Specialization = req.Specialization
	}
	// Interno o externo lo decide el administrador: un docente externo
	// no puede pasarse a interno para entrar a la sección Docente.
	if req.IsExternal != nil && isAdminRequest(ctx) {
		teacher.IsExternal = *req.IsExternal
	}
	
	if req.ApprovalStatus != "" && canApprove(ctx, "role_teacher") {
		if req.ApprovalStatus == "approved" {
			teacher.ApprovalStatus = approvalStatusApproved
			now := time.Now()
			teacher.ApprovalDate = &now
			teacher.ApprovedBy = resolveApprover(ctx, req.ApprovedBy)
		} else {
			teacher.ApprovalStatus = req.ApprovalStatus
		}
	}

	if err := s.teacherRepo.Update(ctx, teacher); err != nil {
		return nil, err
	}

	return dto.ToTeacherResponse(teacher), nil
}

func (s *teacherService) DeleteTeacher(ctx context.Context, id string) error {
	teacher, err := s.teacherRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_teacher" {
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if teacher.UserID != reqUserID {
			return errors.New("forbidden: access denied to delete other teacher profiles")
		}
	}

	return s.teacherRepo.Delete(ctx, id)
}
