package services

import (
	"context"
	"fmt"

	"auth/internal/dto"
	"auth/internal/models"
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

func (s *roleService) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (resp *dto.RoleResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error crítico inesperado: %v", r)
		}
	}()

	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, fmt.Errorf("datos inválidos: %w", err)
	}

	role := &models.Role{
		RoleID:      "rol-" + utils.GenerateRandomString(8),
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("error creando rol: %w", err)
	}

	return dto.ToRoleResponse(role), nil
}

func (s *roleService) GetRoleByID(ctx context.Context, id string) (resp *dto.RoleResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error crítico inesperado: %v", r)
		}
	}()

	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("rol no encontrado: %w", err)
	}
	return dto.ToRoleResponse(role), nil
}

func (s *roleService) GetAllRoles(ctx context.Context) (resp []dto.RoleResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error crítico inesperado: %v", r)
		}
	}()

	roles, err := s.roleRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo roles: %w", err)
	}
	var responseList []dto.RoleResponse
	for _, r := range roles {
		responseList = append(responseList, *dto.ToRoleResponse(&r))
	}
	return responseList, nil
}

func (s *roleService) UpdateRole(ctx context.Context, id string, req *dto.UpdateRoleRequest) (resp *dto.RoleResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error crítico inesperado: %v", r)
		}
	}()

	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, fmt.Errorf("datos inválidos: %w", err)
	}

	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("rol no encontrado: %w", err)
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("error actualizando rol: %w", err)
	}

	return dto.ToRoleResponse(role), nil
}

func (s *roleService) DeleteRole(ctx context.Context, id string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error crítico inesperado: %v", r)
		}
	}()

	_, err = s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("rol no encontrado: %w", err)
	}
	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error eliminando rol: %w", err)
	}
	return nil
}
