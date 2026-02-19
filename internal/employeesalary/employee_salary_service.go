package employeesalary

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

//go:generate mockgen -source=employee_salary_service.go -destination=mock/employee_salary_service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, companyID string, req CreateEmployeeSalaryRequest) (EmployeeSalaryResponse, error)
	GetAll(ctx context.Context, companyID string) ([]EmployeeSalaryResponse, error)
	GetByID(ctx context.Context, companyID, id string) (EmployeeSalaryResponse, error)
	Update(ctx context.Context, companyID, id string, req UpdateEmployeeSalaryRequest) (EmployeeSalaryResponse, error)
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
	req CreateEmployeeSalaryRequest,
) (EmployeeSalaryResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return EmployeeSalaryResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return EmployeeSalaryResponse{}, err
	}

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return EmployeeSalaryResponse{}, err
	}

	salary := &EmployeeSalary{
		ID:            uuid.New(),
		EmployeeID:    employeeID,
		BaseSalary:    req.BaseSalary,
		EffectiveDate: effectiveDate,
	}

	if err := qtx.Create(ctx, salary); err != nil {
		return EmployeeSalaryResponse{}, mapRepositoryError(err)
	}

	created, err := qtx.FindByIDAndCompany(ctx, companyID, salary.ID.String())
	if err != nil {
		return EmployeeSalaryResponse{}, mapRepositoryError(err)
	}

	if err := tx.Commit(); err != nil {
		return EmployeeSalaryResponse{}, err
	}

	return mapToResponse(*created), nil
}

func (s *service) GetAll(
	ctx context.Context,
	companyID string,
) ([]EmployeeSalaryResponse, error) {
	salaries, err := s.repo.FindAllByCompany(ctx, companyID)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	return mapToListResponse(salaries), nil
}

func (s *service) GetByID(
	ctx context.Context,
	companyID, id string,
) (EmployeeSalaryResponse, error) {
	salary, err := s.repo.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return EmployeeSalaryResponse{}, mapRepositoryError(err)
	}

	return mapToResponse(*salary), nil
}

func (s *service) Update(
	ctx context.Context,
	companyID, id string,
	req UpdateEmployeeSalaryRequest,
) (EmployeeSalaryResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return EmployeeSalaryResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	_, err = qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return EmployeeSalaryResponse{}, mapRepositoryError(err)
	}

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return EmployeeSalaryResponse{}, err
	}

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return EmployeeSalaryResponse{}, err
	}

	newSalary := &EmployeeSalary{
		ID:            uuid.New(),
		EmployeeID:    employeeID,
		BaseSalary:    req.BaseSalary,
		EffectiveDate: effectiveDate,
	}

	if err := qtx.Create(ctx, newSalary); err != nil {
		return EmployeeSalaryResponse{}, mapRepositoryError(err)
	}

	if err := tx.Commit(); err != nil {
		return EmployeeSalaryResponse{}, err
	}

	return mapToResponse(*newSalary), nil
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

	if err := qtx.Delete(ctx, companyID, id); err != nil {
		return mapRepositoryError(err)
	}

	return tx.Commit()
}

func mapToResponse(salary EmployeeSalary) EmployeeSalaryResponse {
	return EmployeeSalaryResponse{
		ID:            salary.ID.String(),
		EmployeeID:    salary.EmployeeID.String(),
		EmployeeName:  salary.EmployeeName,
		BaseSalary:    salary.BaseSalary,
		EffectiveDate: salary.EffectiveDate.Format("2006-01-02"),
	}
}

func mapToListResponse(salaries []EmployeeSalary) []EmployeeSalaryResponse {
	res := make([]EmployeeSalaryResponse, len(salaries))
	for i, salary := range salaries {
		res[i] = mapToResponse(salary)
	}
	return res
}
