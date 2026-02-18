package employee

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

//go:generate mockgen -source=employee_service.go -destination=mock/employee_service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, companyID string, req CreateEmployeeRequest) (EmployeeResponse, error)
	GetAll(ctx context.Context, companyID string) ([]EmployeeResponse, error)
	GetByID(ctx context.Context, companyID, id string) (EmployeeResponse, error)
	Update(ctx context.Context, companyID, id string, req UpdateEmployeeRequest) (EmployeeResponse, error)
	Delete(ctx context.Context, companyID, id string) error
}

type service struct {
	db   *sql.DB
	repo Repository
}

func NewService(db *sql.DB, repo Repository) Service {
	return &service{db: db, repo: repo}
}

func (s *service) Create(
	ctx context.Context,
	companyID string,
	req CreateEmployeeRequest,
) (EmployeeResponse, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return EmployeeResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	dept := &Employee{
		ID:        uuid.New(),
		Name:      req.Name,
		Email:     req.Email,
		CompanyID: uuid.MustParse(companyID),
	}

	if err := qtx.Create(ctx, dept); err != nil {
		return EmployeeResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return EmployeeResponse{}, err
	}

	return mapToResponse(*dept), nil
}

func (s *service) GetAll(
	ctx context.Context,
	companyID string,
) ([]EmployeeResponse, error) {

	depts, err := s.repo.FindAllByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	return mapToListResponse(depts), nil
}

func (s *service) GetByID(
	ctx context.Context,
	companyID, id string,
) (EmployeeResponse, error) {

	dept, err := s.repo.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return EmployeeResponse{}, err
	}

	return mapToResponse(*dept), nil
}

func (s *service) Update(
	ctx context.Context,
	companyID, id string,
	req UpdateEmployeeRequest,
) (EmployeeResponse, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return EmployeeResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	dept, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return EmployeeResponse{}, err
	}

	dept.Name = req.Name
	dept.Email = req.Email

	if err := qtx.Update(ctx, dept); err != nil {
		return EmployeeResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return EmployeeResponse{}, err
	}

	return mapToResponse(*dept), nil
}

func (s *service) Delete(
	ctx context.Context,
	companyID, id string,
) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	if err := qtx.Delete(ctx, id, companyID); err != nil {
		return err
	}

	return tx.Commit()
}

func mapToResponse(dept Employee) EmployeeResponse {
	return EmployeeResponse{
		ID:        dept.ID.String(),
		Name:      dept.Name,
		Email:     dept.Email,
		CompanyID: dept.CompanyID.String(),
	}
}

func mapToListResponse(depts []Employee) []EmployeeResponse {
	res := make([]EmployeeResponse, len(depts))
	for i, d := range depts {
		res[i] = mapToResponse(d)
	}
	return res
}
