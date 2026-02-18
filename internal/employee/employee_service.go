package employee

import (
	"context"
	"database/sql"
	"errors"
	"time"

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
	departmentID, err := qtx.GetDepartmentIDByPosition(ctx, companyID, req.PositionID)
	if err != nil {
		return EmployeeResponse{}, err
	}
	if departmentID == "" {
		return EmployeeResponse{}, errors.New("position not found for this company")
	}
	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		return EmployeeResponse{}, errors.New("invalid hire_date format, expected YYYY-MM-DD")
	}

	dept := &Employee{
		ID:           uuid.New(),
		FullName:     req.FullName,
		Email:        req.Email,
		CompanyID:    uuid.MustParse(companyID),
		PositionID:   uuidPtr(req.PositionID),
		DepartmentID: uuidPtr(departmentID),
		EmployeeNumber: req.EmployeeNumber,
		Phone: req.Phone,
		HireDate: hireDate,
		EmploymentStatus: req.EmploymentStatus,
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
	departmentID, err := qtx.GetDepartmentIDByPosition(ctx, companyID, req.PositionID)
	if err != nil {
		return EmployeeResponse{}, err
	}
	if departmentID == "" {
		return EmployeeResponse{}, errors.New("position not found for this company")
	}
	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		return EmployeeResponse{}, errors.New("invalid hire_date format, expected YYYY-MM-DD")
	}

	dept, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return EmployeeResponse{}, err
	}

	dept.FullName = req.FullName
	dept.Email = req.Email
	dept.PositionID = uuidPtr(req.PositionID)
	dept.DepartmentID = uuidPtr(departmentID)
	dept.EmployeeNumber = req.EmployeeNumber
	dept.Phone = req.Phone
	dept.HireDate = hireDate
	dept.EmploymentStatus = req.EmploymentStatus

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

	if err := qtx.Delete(ctx, companyID, id); err != nil {
		return err
	}

	return tx.Commit()
}

func mapToResponse(dept Employee) EmployeeResponse {
	return EmployeeResponse{
		ID:           dept.ID.String(),
		FullName:     dept.FullName,
		Email:        dept.Email,
		EmployeeNumber: dept.EmployeeNumber,
		Phone: dept.Phone,
		HireDate: dept.HireDate.Format("2006-01-02"),
		EmploymentStatus: dept.EmploymentStatus,
		CompanyID:    dept.CompanyID.String(),
		DepartmentID: uuidToString(dept.DepartmentID),
		PositionID:   uuidToString(dept.PositionID),
	}
}

func mapToListResponse(depts []Employee) []EmployeeResponse {
	res := make([]EmployeeResponse, len(depts))
	for i, d := range depts {
		res[i] = mapToResponse(d)
	}
	return res
}

func uuidPtr(v string) *uuid.UUID {
	id, err := uuid.Parse(v)
	if err != nil {
		return nil
	}
	return &id
}

func uuidToString(v *uuid.UUID) string {
	if v == nil {
		return ""
	}
	return v.String()
}
