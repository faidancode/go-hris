package company

import (
	"context"
	companyerrors "go-hris/internal/company/errors"
	"strings"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=mock/company_service_mock.go -package=mock . Service
type Service interface {
	GetByID(ctx context.Context, id string) (*CompanyResponse, error)
	GetByEmail(ctx context.Context, email string) (*CompanyResponse, error)
	Update(ctx context.Context, id string, req UpdateCompanyRequest) (*CompanyResponse, error)

	UpsertRegistration(ctx context.Context, companyID string, req UpsertCompanyRegistrationRequest) error
	ListRegistrations(ctx context.Context, companyID string) ([]CompanyRegistrationResponse, error)
	DeleteRegistration(ctx context.Context, companyID string, regType RegistrationType) error
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

func (s *service) UpsertRegistration(ctx context.Context, companyID string, req UpsertCompanyRegistrationRequest) error {
	id, err := uuid.Parse(companyID)
	if err != nil {
		return companyerrors.ErrInvalidCompanyID
	}

	reg := &CompanyRegistration{
		CompanyID: id,
		Type:      req.Type,
		Number:    req.Number,
		IssuedAt:  req.IssuedAt,
	}

	if req.Type == "" {
		return companyerrors.ErrInvalidRegistrationType
	}

	if strings.TrimSpace(req.Number) == "" {
		return companyerrors.ErrMissingRequiredFields
	}

	return s.repo.UpsertRegistration(ctx, reg)
}

func (s *service) ListRegistrations(ctx context.Context, companyID string) ([]CompanyRegistrationResponse, error) {
	id, err := uuid.Parse(companyID)
	if err != nil {
		return nil, companyerrors.ErrInvalidCompanyID
	}

	regs, err := s.repo.GetRegistrationsByCompanyID(ctx, id)
	if err != nil {
		return nil, err
	}

	var result []CompanyRegistrationResponse
	for _, r := range regs {
		result = append(result, CompanyRegistrationResponse{
			ID:        r.ID.String(),
			Type:      r.Type,
			Number:    r.Number,
			IssuedAt:  r.IssuedAt,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		})
	}

	return result, nil
}

func (s *service) DeleteRegistration(
	ctx context.Context,
	companyID string,
	regType RegistrationType,
) error {

	id, err := uuid.Parse(companyID)
	if err != nil {
		return companyerrors.ErrInvalidCompanyID
	}

	// Optional: validasi type kosong
	if regType == "" {
		return companyerrors.ErrInvalidRegistrationType
	}

	return s.repo.DeleteRegistration(ctx, id, regType)
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
