package payroll

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	StatusDraft     = "DRAFT"
	StatusProcessed = "PROCESSED"
	StatusPaid      = "PAID"
	StatusCancelled = "CANCELLED"
)

//go:generate mockgen -source=payroll_service.go -destination=mock/payroll_service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, companyID, actorID string, req CreatePayrollRequest) (PayrollResponse, error)
	GetAll(ctx context.Context, companyID string) ([]PayrollResponse, error)
	GetByID(ctx context.Context, companyID, id string) (PayrollResponse, error)
	Update(ctx context.Context, companyID, actorID, id string, req UpdatePayrollRequest) (PayrollResponse, error)
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
	companyID, actorID string,
	req CreatePayrollRequest,
) (PayrollResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PayrollResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	companyUUID, employeeUUID, createdByUUID, periodStart, periodEnd, err := validateCreateRequest(companyID, actorID, req)
	if err != nil {
		return PayrollResponse{}, err
	}

	belongs, err := qtx.EmployeeBelongsToCompany(ctx, companyID, req.EmployeeID)
	if err != nil {
		return PayrollResponse{}, err
	}
	if !belongs {
		return PayrollResponse{}, errors.New("employee does not belong to this company")
	}

	overlap, err := qtx.HasOverlappingPeriod(ctx, companyID, req.EmployeeID, periodStart, periodEnd, nil)
	if err != nil {
		return PayrollResponse{}, err
	}
	if overlap {
		return PayrollResponse{}, errors.New("payroll already exists in overlapping period")
	}

	netSalary := req.BaseSalary + req.Allowance - req.Deduction

	payroll := &Payroll{
		ID:          uuid.New(),
		CompanyID:   companyUUID,
		EmployeeID:  employeeUUID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		BaseSalary:  req.BaseSalary,
		Allowance:   req.Allowance,
		Deduction:   req.Deduction,
		NetSalary:   netSalary,
		Status:      StatusDraft,
		CreatedBy:   createdByUUID,
	}

	if err := qtx.Create(ctx, payroll); err != nil {
		return PayrollResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return PayrollResponse{}, err
	}

	return mapToResponse(*payroll), nil
}

func (s *service) GetAll(
	ctx context.Context,
	companyID string,
) ([]PayrollResponse, error) {
	payrolls, err := s.repo.FindAllByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	return mapToListResponse(payrolls), nil
}

func (s *service) GetByID(
	ctx context.Context,
	companyID, id string,
) (PayrollResponse, error) {
	payroll, err := s.repo.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return PayrollResponse{}, err
	}

	return mapToResponse(*payroll), nil
}

func (s *service) Update(
	ctx context.Context,
	companyID, actorID, id string,
	req UpdatePayrollRequest,
) (PayrollResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PayrollResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	_, err = uuid.Parse(companyID)
	if err != nil {
		return PayrollResponse{}, errors.New("invalid company id")
	}

	_, err = uuid.Parse(actorID)
	if err != nil {
		return PayrollResponse{}, errors.New("invalid actor id")
	}

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return PayrollResponse{}, errors.New("invalid employee id")
	}

	periodStart, err := parseDate(req.PeriodStart)
	if err != nil {
		return PayrollResponse{}, err
	}
	periodEnd, err := parseDate(req.PeriodEnd)
	if err != nil {
		return PayrollResponse{}, err
	}
	if periodStart.After(periodEnd) {
		return PayrollResponse{}, errors.New("period_start must be before or equal period_end")
	}

	payroll, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return PayrollResponse{}, err
	}

	belongs, err := qtx.EmployeeBelongsToCompany(ctx, companyID, req.EmployeeID)
	if err != nil {
		return PayrollResponse{}, err
	}
	if !belongs {
		return PayrollResponse{}, errors.New("employee does not belong to this company")
	}

	overlap, err := qtx.HasOverlappingPeriod(ctx, companyID, req.EmployeeID, periodStart, periodEnd, &id)
	if err != nil {
		return PayrollResponse{}, err
	}
	if overlap {
		return PayrollResponse{}, errors.New("payroll already exists in overlapping period")
	}

	payroll.EmployeeID = employeeID
	payroll.PeriodStart = periodStart
	payroll.PeriodEnd = periodEnd
	payroll.BaseSalary = req.BaseSalary
	payroll.Allowance = req.Allowance
	payroll.Deduction = req.Deduction
	payroll.NetSalary = req.BaseSalary + req.Allowance - req.Deduction
	payroll.Status = req.Status

	if req.ApprovedBy != nil && *req.ApprovedBy != "" {
		approverID, err := uuid.Parse(*req.ApprovedBy)
		if err != nil {
			return PayrollResponse{}, errors.New("invalid approved_by")
		}
		payroll.ApprovedBy = &approverID
	}

	if req.PaidAt != nil && *req.PaidAt != "" {
		paidAt, err := parseDateTime(*req.PaidAt)
		if err != nil {
			return PayrollResponse{}, err
		}
		payroll.PaidAt = &paidAt
	}

	if payroll.Status == StatusPaid && payroll.PaidAt == nil {
		return PayrollResponse{}, errors.New("paid_at is required when status is PAID")
	}

	if err := qtx.Update(ctx, payroll); err != nil {
		return PayrollResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return PayrollResponse{}, err
	}

	return mapToResponse(*payroll), nil
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

func validateCreateRequest(
	companyID, actorID string,
	req CreatePayrollRequest,
) (uuid.UUID, uuid.UUID, uuid.UUID, time.Time, time.Time, error) {
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, errors.New("invalid company id")
	}

	employeeUUID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, errors.New("invalid employee id")
	}

	createdByUUID, err := uuid.Parse(actorID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, errors.New("invalid actor id")
	}

	periodStart, err := parseDate(req.PeriodStart)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, err
	}
	periodEnd, err := parseDate(req.PeriodEnd)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, err
	}

	if periodStart.After(periodEnd) {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, errors.New("period_start must be before or equal period_end")
	}

	return companyUUID, employeeUUID, createdByUUID, periodStart, periodEnd, nil
}

func parseDate(v string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, errors.New("invalid date format, expected YYYY-MM-DD")
	}
	return t, nil
}

func parseDateTime(v string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return time.Time{}, errors.New("invalid datetime format, expected RFC3339")
	}
	return t, nil
}

func mapToResponse(payroll Payroll) PayrollResponse {
	resp := PayrollResponse{
		ID:          payroll.ID.String(),
		CompanyID:   payroll.CompanyID.String(),
		EmployeeID:  payroll.EmployeeID.String(),
		PeriodStart: payroll.PeriodStart.Format("2006-01-02"),
		PeriodEnd:   payroll.PeriodEnd.Format("2006-01-02"),
		BaseSalary:  payroll.BaseSalary,
		Allowance:   payroll.Allowance,
		Deduction:   payroll.Deduction,
		NetSalary:   payroll.NetSalary,
		Status:      payroll.Status,
		CreatedBy:   payroll.CreatedBy.String(),
	}

	if payroll.ApprovedBy != nil {
		v := payroll.ApprovedBy.String()
		resp.ApprovedBy = &v
	}
	if payroll.PaidAt != nil {
		v := payroll.PaidAt.Format(time.RFC3339)
		resp.PaidAt = &v
	}
	if payroll.ApprovedAt != nil {
		v := payroll.ApprovedAt.Format(time.RFC3339)
		resp.ApprovedAt = &v
	}

	return resp
}

func mapToListResponse(payrolls []Payroll) []PayrollResponse {
	resp := make([]PayrollResponse, len(payrolls))
	for i, payroll := range payrolls {
		resp[i] = mapToResponse(payroll)
	}
	return resp
}
