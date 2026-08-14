package services

import (
	"context"

	"auth/internal/dto"
	"auth/internal/models"
	"auth/internal/repositories"
	"auth/internal/utils"
)

type PermissionService interface {
	CreatePermission(ctx context.Context, req *dto.CreatePermissionRequest) (*dto.PermissionResponse, error)
	GetPermissionByID(ctx context.Context, id string) (*dto.PermissionResponse, error)
	GetAllPermissions(ctx context.Context) ([]dto.PermissionResponse, error)
	UpdatePermission(ctx context.Context, id string, req *dto.UpdatePermissionRequest) (*dto.PermissionResponse, error)
	DeletePermission(ctx context.Context, id string) error
	CheckPermission(ctx context.Context, roleID, action, resource string) (bool, error)
}

type permissionService struct {
	permissionRepo repositories.PermissionRepository
}

func NewPermissionService(permissionRepo repositories.PermissionRepository) PermissionService {
	return &permissionService{permissionRepo: permissionRepo}
}

func (s *permissionService) CreatePermission(ctx context.Context, req *dto.CreatePermissionRequest) (*dto.PermissionResponse, error) {
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	permission := &models.Permission{
		PermissionID: "perm-" + utils.GenerateRandomString(8),
		Action:       req.Action,
		Resource:     req.Resource,
		Description:  req.Description,
	}

	if err := s.permissionRepo.Create(ctx, permission); err != nil {
		return nil, err
	}

	return dto.ToPermissionResponse(permission), nil
}

func (s *permissionService) GetPermissionByID(ctx context.Context, id string) (*dto.PermissionResponse, error) {
	permission, err := s.permissionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return dto.ToPermissionResponse(permission), nil
}

func (s *permissionService) GetAllPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	permissions, err := s.permissionRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	var resp []dto.PermissionResponse
	for _, p := range permissions {
		resp = append(resp, *dto.ToPermissionResponse(&p))
	}
	return resp, nil
}

func (s *permissionService) UpdatePermission(ctx context.Context, id string, req *dto.UpdatePermissionRequest) (*dto.PermissionResponse, error) {
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	permission, err := s.permissionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Action != "" {
		permission.Action = req.Action
	}
	if req.Resource != "" {
		permission.Resource = req.Resource
	}
	if req.Description != "" {
		permission.Description = req.Description
	}

	if err := s.permissionRepo.Update(ctx, permission); err != nil {
		return nil, err
	}

	return dto.ToPermissionResponse(permission), nil
}

func (s *permissionService) DeletePermission(ctx context.Context, id string) error {
	_, err := s.permissionRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	return s.permissionRepo.Delete(ctx, id)
}

func (s *permissionService) CheckPermission(ctx context.Context, roleID, action, resource string) (bool, error) {
	var count int64
	err := s.permissionRepo.CheckRolePermission(ctx, roleID, action, resource, &count)
	return count > 0, err
}
