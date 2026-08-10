package services

import (
	"context"

	"auth/internal/models"
	"auth/internal/dto"
	"auth/internal/repositories"
	"auth/internal/utils"
)

type RoleService interface {
	CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error)
	GetRoleByID(ctx context.Context, id string) (*dto.RoleResponse, error)
	GetAllRoles(ctx context.Context) ([]dto.RoleResponse, error)
	UpdateRole(ctx context.Context, id string, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	DeleteRole(ctx context.Context, id string) error
}

type roleService struct {
	roleRepo repositories.RoleRepository
}

func NewRoleService(roleRepo repositories.RoleRepository) RoleService {
	return &roleService{roleRepo: roleRepo}
}

func (s *roleService) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	role := &models.Role{
		RoleID:      "rol-" + utils.GenerateRandomString(8),
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	return dto.ToRoleResponse(role), nil
}

func (s *roleService) GetRoleByID(ctx context.Context, id string) (*dto.RoleResponse, error) {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return dto.ToRoleResponse(role), nil
}

func (s *roleService) GetAllRoles(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := s.roleRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	var resp []dto.RoleResponse
	for _, r := range roles {
		resp = append(resp, *dto.ToRoleResponse(&r))
	}
	return resp, nil
}

func (s *roleService) UpdateRole(ctx context.Context, id string, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}

	return dto.ToRoleResponse(role), nil
}

func (s *roleService) DeleteRole(ctx context.Context, id string) error {
	_, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	return s.roleRepo.Delete(ctx, id)
}
