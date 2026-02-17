package department

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

//go:generate mockgen -source=department_service.go -destination=mock/department_service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, companyID string, req CreateDepartmentRequest) (DepartmentResponse, error)
	GetAll(ctx context.Context, companyID string) ([]DepartmentResponse, error)
	GetByID(ctx context.Context, companyID, id string) (DepartmentResponse, error)
	Update(ctx context.Context, companyID, id string, req UpdateDepartmentRequest) (DepartmentResponse, error)
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
	req CreateDepartmentRequest,
) (DepartmentResponse, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return DepartmentResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	dept := &Department{
		ID:        uuid.New(),
		Name:      req.Name,
		CompanyID: uuid.MustParse(companyID),
	}

	if err := qtx.Create(ctx, dept); err != nil {
		return DepartmentResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return DepartmentResponse{}, err
	}

	return mapToResponse(*dept), nil
}

func (s *service) GetAll(
	ctx context.Context,
	companyID string,
) ([]DepartmentResponse, error) {

	depts, err := s.repo.FindAllByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	return mapToListResponse(depts), nil
}

func (s *service) GetByID(
	ctx context.Context,
	companyID, id string,
) (DepartmentResponse, error) {

	dept, err := s.repo.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return DepartmentResponse{}, err
	}

	return mapToResponse(*dept), nil
}

func (s *service) Update(
	ctx context.Context,
	companyID, id string,
	req UpdateDepartmentRequest,
) (DepartmentResponse, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return DepartmentResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	dept, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return DepartmentResponse{}, err
	}

	dept.Name = req.Name

	if err := qtx.Update(ctx, dept); err != nil {
		return DepartmentResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return DepartmentResponse{}, err
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

func mapToResponse(dept Department) DepartmentResponse {
	return DepartmentResponse{
		ID:        dept.ID.String(),
		Name:      dept.Name,
		CompanyID: dept.CompanyID.String(),
	}
}

func mapToListResponse(depts []Department) []DepartmentResponse {
	res := make([]DepartmentResponse, len(depts))
	for i, d := range depts {
		res[i] = mapToResponse(d)
	}
	return res
}
