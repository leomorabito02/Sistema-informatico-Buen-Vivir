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

type CompanyService interface {
	CreateCompany(ctx context.Context, req *dto.CreateCompanyRequest) (*dto.CompanyResponse, error)
	GetCompanyByID(ctx context.Context, id string) (*dto.CompanyResponse, error)
	GetAllCompanies(ctx context.Context) ([]dto.CompanyResponse, error)
	UpdateCompany(ctx context.Context, id string, req *dto.UpdateCompanyRequest) (*dto.CompanyResponse, error)
	DeleteCompany(ctx context.Context, id string) error
}

type companyService struct {
	companyRepo repositories.CompanyRepository
}

func NewCompanyService(companyRepo repositories.CompanyRepository) CompanyService {
	return &companyService{companyRepo: companyRepo}
}

func (s *companyService) CreateCompany(ctx context.Context, req *dto.CreateCompanyRequest) (*dto.CompanyResponse, error) {
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	status := req.ApprovalStatus
	if status == "" {
		status = "pending"
	}

	company := &models.Company{
		CompanyID:      "comp-" + utils.GenerateRandomString(10),
		UserID:         req.UserID,
		BusinessName:   req.BusinessName,
		Cuit:           req.Cuit,
		ApprovalStatus: status,
	}

	if err := s.companyRepo.Create(ctx, company); err != nil {
		return nil, err
	}

	return dto.ToCompanyResponse(company), nil
}

func (s *companyService) GetCompanyByID(ctx context.Context, id string) (*dto.CompanyResponse, error) {
	company, err := s.companyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_company" {
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if company.UserID != reqUserID {
			return nil, errors.New("forbidden: access denied to other company profiles")
		}
	}

	return dto.ToCompanyResponse(company), nil
}

func (s *companyService) GetAllCompanies(ctx context.Context) ([]dto.CompanyResponse, error) {
	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_company" {
		return nil, errors.New("forbidden: cannot list all companies")
	}

	companies, err := s.companyRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	var resp []dto.CompanyResponse
	for _, c := range companies {
		resp = append(resp, *dto.ToCompanyResponse(&c))
	}
	return resp, nil
}

func (s *companyService) UpdateCompany(ctx context.Context, id string, req *dto.UpdateCompanyRequest) (*dto.CompanyResponse, error) {
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	company, err := s.companyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_company" {
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if company.UserID != reqUserID {
			return nil, errors.New("forbidden: access denied to edit other company profiles")
		}
	}

	if req.BusinessName != "" {
		company.BusinessName = req.BusinessName
	}
	
	if req.ApprovalStatus != "" {
		if req.ApprovalStatus == "approved" {
			company.ApprovalStatus = approvalStatusAccepted
			now := time.Now()
			company.ApprovalDate = &now
			company.ApprovedBy = &req.ApprovedBy
		} else {
			company.ApprovalStatus = req.ApprovalStatus
		}
	}

	if err := s.companyRepo.Update(ctx, company); err != nil {
		return nil, err
	}

	return dto.ToCompanyResponse(company), nil
}

func (s *companyService) DeleteCompany(ctx context.Context, id string) error {
	company, err := s.companyRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_company" {
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if company.UserID != reqUserID {
			return errors.New("forbidden: access denied to delete other company profiles")
		}
	}

	return s.companyRepo.Delete(ctx, id)
}
