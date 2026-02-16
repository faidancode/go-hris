package employee

import (
	"context"
	"database/sql"
	autherrors "go-hris/internal/auth/errors"
)

//go:generate mockgen -source=employee_service.go -destination=mock/employee_service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, req CreateEmployeeRequest) (EmployeeResponse, error)

	GetAll(
		ctx context.Context,
		companyID string,
		requesterID string,
		hasReadAll bool,
	) ([]EmployeeResponse, error)

	GetByID(
		ctx context.Context,
		requesterID string,
		targetID string,
		hasReadAll bool,
	) (EmployeeResponse, error)
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
	req CreateEmployeeRequest,
) (EmployeeResponse, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return EmployeeResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	emp, err := qtx.Create(ctx, CreateEmployeeRequest{
		Name:      req.Name,
		Email:     req.Email,
		CompanyID: req.CompanyID,
	})
	if err != nil {
		return EmployeeResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return EmployeeResponse{}, err
	}

	return mapToResponse(emp), nil
}

func (s *service) GetAll(
	ctx context.Context,
	companyID string,
	requesterID string,
	hasReadAll bool,
) ([]EmployeeResponse, error) {

	if !hasReadAll {
		// Jika tidak punya read_all,
		// hanya return dirinya sendiri
		emp, err := s.repo.FindByID(ctx, requesterID)
		if err != nil {
			return nil, err
		}

		return []EmployeeResponse{mapToResponse(*emp)}, nil
	}

	emps, err := s.repo.FindAllByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	return mapToListResponse(emps), nil
}

func (s *service) GetByID(
	ctx context.Context,
	requesterID string,
	targetID string,
	hasReadAll bool,
) (EmployeeResponse, error) {

	// Jika bukan admin/HR (hasReadAll=false) dan mencoba akses orang lain → forbidden
	if !hasReadAll && requesterID != targetID {
		return EmployeeResponse{}, autherrors.ErrForbidden
	}

	// Ambil employee dari repo
	emp, err := s.repo.FindByID(ctx, targetID)
	if err != nil {
		return EmployeeResponse{}, err
	}

	return mapToResponse(*emp), nil
}

func mapToResponse(emp Employee) EmployeeResponse {
	return EmployeeResponse{
		ID:        emp.ID.String(),
		Name:      emp.Name,
		Email:     emp.Email,
		CompanyID: emp.CompanyID.String(),
	}
}

func mapToListResponse(emps []Employee) []EmployeeResponse {
	res := make([]EmployeeResponse, len(emps))
	for i, emp := range emps {
		res[i] = mapToResponse(emp)
	}
	return res
}
