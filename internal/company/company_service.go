package company

import (
	"context"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=mock/company_service_mock.go -package=mock . Service
type Service interface {
	GetByID(ctx context.Context, id string) (*CompanyResponse, error)
	GetByEmail(ctx context.Context, email string) (*CompanyResponse, error)
	Update(ctx context.Context, id string, req UpdateCompanyRequest) (*CompanyResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(ctx context.Context, id string) (*CompanyResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	comp, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(comp), nil
}

func (s *service) GetByEmail(ctx context.Context, email string) (*CompanyResponse, error) {
	comp, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(comp), nil
}

func (s *service) Update(ctx context.Context, id string, req UpdateCompanyRequest) (*CompanyResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	comp, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		comp.Name = req.Name
	}
	if req.RegistrationNumber != "" {
		comp.RegistrationNumber = req.RegistrationNumber
	}
	if req.IsActive != nil {
		comp.IsActive = *req.IsActive
	}

	err = s.repo.Update(ctx, comp)
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(comp), nil
}

func (s *service) mapToResponse(c *Company) *CompanyResponse {
	return &CompanyResponse{
		ID:                 c.ID.String(),
		Name:               c.Name,
		Email:              c.Email,
		RegistrationNumber: c.RegistrationNumber,
		IsActive:           c.IsActive,
	}
}
