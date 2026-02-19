package payroll

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go-hris/internal/events"
	"go-hris/internal/messaging/kafka"
	payrollerrors "go-hris/internal/payroll/errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	StatusDraft    = "DRAFT"
	StatusApproved = "APPROVED"
	StatusPaid     = "PAID"

	ComponentTypeAllowance = "ALLOWANCE"
	ComponentTypeDeduction = "DEDUCTION"
)

//go:generate mockgen -source=payroll_service.go -destination=mock/payroll_service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, companyID, actorID string, req CreatePayrollRequest) (PayrollResponse, error)
	GetAll(ctx context.Context, companyID string, filterReq GetPayrollsFilterRequest) ([]PayrollResponse, error)
	GetByID(ctx context.Context, companyID, id string) (PayrollResponse, error)
	GetBreakdown(ctx context.Context, companyID, id string) (PayrollBreakdownResponse, error)
	Regenerate(ctx context.Context, companyID, actorID, id string, req RegeneratePayrollRequest) (PayrollResponse, error)
	Approve(ctx context.Context, companyID, actorID, id string) (PayrollResponse, error)
	MarkAsPaid(ctx context.Context, companyID, actorID, id string) (PayrollResponse, error)
	GeneratePayslip(ctx context.Context, companyID, id string) (PayrollResponse, error)
	Delete(ctx context.Context, companyID, id string) error
}

type service struct {
	db     *sql.DB
	repo   Repository
	outbox kafka.OutboxRepository
}

func NewService(db *sql.DB, repo Repository) Service {
	return NewServiceWithOutbox(db, repo, nil)
}

func NewServiceWithOutbox(db *sql.DB, repo Repository, outboxRepo kafka.OutboxRepository) Service {
	return &service{db: db, repo: repo, outbox: outboxRepo}
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
		return PayrollResponse{}, payrollerrors.ErrEmployeeNotInCompany
	}

	overlap, err := qtx.HasOverlappingPeriod(ctx, companyID, req.EmployeeID, periodStart, periodEnd, nil)
	if err != nil {
		return PayrollResponse{}, err
	}
	if overlap {
		return PayrollResponse{}, payrollerrors.ErrPayrollOverlap
	}

	allowanceItems, deductionItems, err := buildComponents(companyUUID, nil, req.AllowanceItems, req.DeductionItems)
	if err != nil {
		return PayrollResponse{}, err
	}
	totalAllowanceItems := sumComponents(allowanceItems)
	totalDeductionItems := sumComponents(deductionItems)

	overtimeAmount, err := calculateOvertime(req.OvertimeHours, req.OvertimeRate)
	if err != nil {
		return PayrollResponse{}, err
	}

	totalAllowance := req.Allowance + totalAllowanceItems
	totalDeduction := req.Deduction + totalDeductionItems
	if err := validateMoney(req.BaseSalary, totalAllowance, totalDeduction); err != nil {
		return PayrollResponse{}, err
	}

	payroll := &Payroll{
		ID:             uuid.New(),
		CompanyID:      companyUUID,
		EmployeeID:     employeeUUID,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		BaseSalary:     req.BaseSalary,
		Allowance:      totalAllowance,
		OvertimeHours:  req.OvertimeHours,
		OvertimeRate:   req.OvertimeRate,
		OvertimeAmount: overtimeAmount,
		Deduction:      totalDeduction,
		NetSalary:      req.BaseSalary + totalAllowance + overtimeAmount - totalDeduction,
		Status:         StatusDraft,
		CreatedBy:      createdByUUID,
	}

	if err := qtx.Create(ctx, payroll); err != nil {
		return PayrollResponse{}, err
	}

	allComponents := attachPayrollID(payroll.ID, append(allowanceItems, deductionItems...))
	if err := qtx.ReplaceComponents(ctx, companyID, payroll.ID.String(), allComponents); err != nil {
		return PayrollResponse{}, err
	}

	persisted, err := qtx.FindByIDAndCompany(ctx, companyID, payroll.ID.String())
	if err != nil {
		return PayrollResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return PayrollResponse{}, err
	}

	return mapToResponse(*persisted), nil
}

func (s *service) GetAll(
	ctx context.Context,
	companyID string,
	filterReq GetPayrollsFilterRequest,
) ([]PayrollResponse, error) {
	filter, err := s.buildListFilter(companyID, filterReq)
	if err != nil {
		return nil, err
	}

	payrolls, err := s.repo.FindAllByCompany(ctx, companyID, filter)
	if err != nil {
		return nil, err
	}

	return mapToListResponse(payrolls), nil
}

func (s *service) buildListFilter(companyID string, req GetPayrollsFilterRequest) (PayrollQueryFilter, error) {
	if _, err := uuid.Parse(companyID); err != nil {
		return PayrollQueryFilter{}, payrollerrors.ErrInvalidCompanyID
	}

	filter := PayrollQueryFilter{}

	if req.DepartmentID != "" {
		if _, err := uuid.Parse(req.DepartmentID); err != nil {
			return PayrollQueryFilter{}, payrollerrors.ErrInvalidDepartmentID
		}
		filter.DepartmentID = &req.DepartmentID
	}

	if req.Status != "" {
		status := strings.ToUpper(strings.TrimSpace(req.Status))
		if !isValidPayrollStatus(status) {
			return PayrollQueryFilter{}, payrollerrors.ErrInvalidStatusFilter
		}
		filter.Status = &status
	}

	periodStart := strings.TrimSpace(req.PeriodStart)
	periodEnd := strings.TrimSpace(req.PeriodEnd)
	if req.Period != "" && periodStart == "" && periodEnd == "" {
		monthStart, monthEnd, err := parseMonthPeriod(req.Period)
		if err != nil {
			return PayrollQueryFilter{}, err
		}
		periodStart = monthStart
		periodEnd = monthEnd
	}

	if periodStart != "" {
		if _, err := parseDate(periodStart); err != nil {
			return PayrollQueryFilter{}, err
		}
		filter.PeriodStart = &periodStart
	}
	if periodEnd != "" {
		if _, err := parseDate(periodEnd); err != nil {
			return PayrollQueryFilter{}, err
		}
		filter.PeriodEnd = &periodEnd
	}
	if filter.PeriodStart != nil && filter.PeriodEnd != nil {
		start, _ := parseDate(*filter.PeriodStart)
		end, _ := parseDate(*filter.PeriodEnd)
		if start.After(end) {
			return PayrollQueryFilter{}, payrollerrors.ErrInvalidDateRange
		}
	}

	return filter, nil
}

func (s *service) GetByID(
	ctx context.Context,
	companyID, id string,
) (PayrollResponse, error) {
	payroll, err := s.repo.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return PayrollResponse{}, payrollerrors.ErrPayrollNotFound
		}
		return PayrollResponse{}, err
	}

	return mapToResponse(*payroll), nil
}

func (s *service) GetBreakdown(
	ctx context.Context,
	companyID, id string,
) (PayrollBreakdownResponse, error) {
	payroll, err := s.repo.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return PayrollBreakdownResponse{}, payrollerrors.ErrPayrollNotFound
		}
		return PayrollBreakdownResponse{}, err
	}

	return mapToBreakdownResponse(*payroll), nil
}

func (s *service) Regenerate(
	ctx context.Context,
	companyID, actorID, id string,
	req RegeneratePayrollRequest,
) (PayrollResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PayrollResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	if _, err = uuid.Parse(companyID); err != nil {
		return PayrollResponse{}, payrollerrors.ErrInvalidCompanyID
	}
	if _, err = uuid.Parse(actorID); err != nil {
		return PayrollResponse{}, payrollerrors.ErrInvalidActorID
	}

	payroll, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return PayrollResponse{}, payrollerrors.ErrPayrollNotFound
		}
		return PayrollResponse{}, err
	}
	if payroll.Status != StatusDraft {
		return PayrollResponse{}, payrollerrors.ErrRegenerateOnlyDraft
	}

	allowanceItems, deductionItems, err := buildComponents(payroll.CompanyID, &payroll.ID, req.AllowanceItems, req.DeductionItems)
	if err != nil {
		return PayrollResponse{}, err
	}
	totalAllowanceItems := sumComponents(allowanceItems)
	totalDeductionItems := sumComponents(deductionItems)

	overtimeAmount, err := calculateOvertime(req.OvertimeHours, req.OvertimeRate)
	if err != nil {
		return PayrollResponse{}, err
	}

	totalAllowance := req.Allowance + totalAllowanceItems
	totalDeduction := req.Deduction + totalDeductionItems
	if err := validateMoney(req.BaseSalary, totalAllowance, totalDeduction); err != nil {
		return PayrollResponse{}, err
	}

	payroll.BaseSalary = req.BaseSalary
	payroll.Allowance = totalAllowance
	payroll.OvertimeHours = req.OvertimeHours
	payroll.OvertimeRate = req.OvertimeRate
	payroll.OvertimeAmount = overtimeAmount
	payroll.Deduction = totalDeduction
	payroll.NetSalary = req.BaseSalary + totalAllowance + overtimeAmount - totalDeduction

	if err := qtx.Update(ctx, payroll); err != nil {
		return PayrollResponse{}, err
	}

	allComponents := append(allowanceItems, deductionItems...)
	if err := qtx.ReplaceComponents(ctx, companyID, payroll.ID.String(), allComponents); err != nil {
		return PayrollResponse{}, err
	}

	persisted, err := qtx.FindByIDAndCompany(ctx, companyID, payroll.ID.String())
	if err != nil {
		return PayrollResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return PayrollResponse{}, err
	}

	return mapToResponse(*persisted), nil
}

func (s *service) Approve(
	ctx context.Context,
	companyID, actorID, id string,
) (PayrollResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PayrollResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	if _, err = uuid.Parse(companyID); err != nil {
		return PayrollResponse{}, payrollerrors.ErrInvalidCompanyID
	}
	actorUUID, err := uuid.Parse(actorID)
	if err != nil {
		return PayrollResponse{}, payrollerrors.ErrInvalidActorID
	}

	payroll, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return PayrollResponse{}, payrollerrors.ErrPayrollNotFound
		}
		return PayrollResponse{}, err
	}
	if payroll.Status != StatusDraft {
		return PayrollResponse{}, payrollerrors.ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	payroll.Status = StatusApproved
	payroll.ApprovedBy = &actorUUID
	payroll.ApprovedAt = &now

	if err := qtx.Update(ctx, payroll); err != nil {
		return PayrollResponse{}, err
	}

	if s.outbox != nil {
		event := events.PayrollPayslipRequestedEvent{
			EventType:   "payroll_payslip_requested",
			PayrollID:   payroll.ID.String(),
			CompanyID:   companyID,
			RequestedBy: actorID,
			OccurredAt:  now,
		}
		payload, err := json.Marshal(event)
		if err != nil {
			return PayrollResponse{}, err
		}

		outboxRepo := s.outbox.WithTx(tx)
		if err := outboxRepo.Create(ctx, kafka.OutboxEvent{
			ID:            uuid.NewString(),
			AggregateType: "payroll",
			AggregateID:   payroll.ID.String(),
			EventType:     event.EventType,
			Topic:         events.PayrollPayslipRequestedTopic,
			Payload:       payload,
			Status:        kafka.OutboxStatusPending,
		}); err != nil {
			return PayrollResponse{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return PayrollResponse{}, err
	}

	return mapToResponse(*payroll), nil
}

func (s *service) MarkAsPaid(
	ctx context.Context,
	companyID, actorID, id string,
) (PayrollResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PayrollResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	if _, err = uuid.Parse(companyID); err != nil {
		return PayrollResponse{}, payrollerrors.ErrInvalidCompanyID
	}
	if _, err = uuid.Parse(actorID); err != nil {
		return PayrollResponse{}, payrollerrors.ErrInvalidActorID
	}

	payroll, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return PayrollResponse{}, payrollerrors.ErrPayrollNotFound
		}
		return PayrollResponse{}, err
	}
	if payroll.Status != StatusApproved {
		return PayrollResponse{}, payrollerrors.ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	payroll.Status = StatusPaid
	payroll.PaidAt = &now

	if err := qtx.Update(ctx, payroll); err != nil {
		return PayrollResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return PayrollResponse{}, err
	}

	return mapToResponse(*payroll), nil
}

func (s *service) GeneratePayslip(
	ctx context.Context,
	companyID, id string,
) (PayrollResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PayrollResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)
	payroll, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return PayrollResponse{}, payrollerrors.ErrPayrollNotFound
		}
		return PayrollResponse{}, err
	}

	if payroll.PayslipURL != nil && *payroll.PayslipURL != "" {
		return mapToResponse(*payroll), tx.Commit()
	}

	content, err := buildSimplePayslipPDF(buildPayslipLines(*payroll))
	if err != nil {
		return PayrollResponse{}, err
	}

	baseDir := os.Getenv("PAYSLIP_STORAGE_DIR")
	if strings.TrimSpace(baseDir) == "" {
		baseDir = filepath.Join("storage", "payslips")
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return PayrollResponse{}, err
	}

	filename := fmt.Sprintf("payslip_%s.pdf", payroll.ID.String())
	absPath := filepath.Join(baseDir, filename)
	if err := os.WriteFile(absPath, content, 0o644); err != nil {
		return PayrollResponse{}, err
	}

	urlPrefix := strings.TrimSpace(os.Getenv("PAYSLIP_PUBLIC_BASE_URL"))
	if urlPrefix == "" {
		urlPrefix = "/files/payslips"
	}
	publicURL := strings.TrimRight(urlPrefix, "/") + "/" + filename
	now := time.Now().UTC()
	payroll.PayslipURL = &publicURL
	payroll.PayslipGeneratedAt = &now

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

	payroll, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return payrollerrors.ErrPayrollNotFound
		}
		return err
	}
	if payroll.Status != StatusDraft {
		return payrollerrors.ErrDeleteOnlyDraft
	}

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
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, payrollerrors.ErrInvalidCompanyID
	}

	employeeUUID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, payrollerrors.ErrInvalidEmployeeID
	}

	createdByUUID, err := uuid.Parse(actorID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, payrollerrors.ErrInvalidActorID
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
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, payrollerrors.ErrInvalidDateRange
	}

	if req.OvertimeHours < 0 || req.OvertimeRate < 0 {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, payrollerrors.ErrInvalidMoneyValue
	}

	return companyUUID, employeeUUID, createdByUUID, periodStart, periodEnd, nil
}

func validateMoney(values ...int64) error {
	for _, v := range values {
		if v < 0 {
			return payrollerrors.ErrInvalidMoneyValue
		}
	}
	return nil
}

func parseDate(v string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, payrollerrors.ErrInvalidDateFormat
	}
	return t, nil
}

func parseMonthPeriod(v string) (string, string, error) {
	t, err := time.Parse("2006-01", strings.TrimSpace(v))
	if err != nil {
		return "", "", payrollerrors.ErrInvalidPeriodFormat
	}

	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)
	return start.Format("2006-01-02"), end.Format("2006-01-02"), nil
}

func isValidPayrollStatus(v string) bool {
	switch v {
	case StatusDraft, StatusApproved, StatusPaid:
		return true
	default:
		return false
	}
}

func buildComponents(
	companyID uuid.UUID,
	payrollID *uuid.UUID,
	allowanceItems []PayrollComponentInput,
	deductionItems []PayrollComponentInput,
) ([]PayrollComponent, []PayrollComponent, error) {
	allowances, err := mapComponentInputs(companyID, payrollID, ComponentTypeAllowance, allowanceItems)
	if err != nil {
		return nil, nil, err
	}
	deductions, err := mapComponentInputs(companyID, payrollID, ComponentTypeDeduction, deductionItems)
	if err != nil {
		return nil, nil, err
	}
	return allowances, deductions, nil
}

func mapComponentInputs(
	companyID uuid.UUID,
	payrollID *uuid.UUID,
	componentType string,
	inputs []PayrollComponentInput,
) ([]PayrollComponent, error) {
	components := make([]PayrollComponent, 0, len(inputs))
	for _, item := range inputs {
		if item.ComponentName == "" || item.UnitAmount < 0 {
			return nil, payrollerrors.ErrInvalidComponent
		}
		quantity := item.Quantity
		if quantity == 0 {
			quantity = 1
		}
		if quantity < 0 {
			return nil, payrollerrors.ErrInvalidComponent
		}
		payrollRef := uuid.Nil
		if payrollID != nil {
			payrollRef = *payrollID
		}

		components = append(components, PayrollComponent{
			ID:            uuid.New(),
			PayrollID:     payrollRef,
			CompanyID:     companyID,
			ComponentType: componentType,
			ComponentName: item.ComponentName,
			Quantity:      quantity,
			UnitAmount:    item.UnitAmount,
			TotalAmount:   quantity * item.UnitAmount,
			Notes:         item.Notes,
		})
	}
	return components, nil
}

func attachPayrollID(payrollID uuid.UUID, items []PayrollComponent) []PayrollComponent {
	for i := range items {
		items[i].PayrollID = payrollID
	}
	return items
}

func sumComponents(items []PayrollComponent) int64 {
	var total int64
	for _, item := range items {
		total += item.TotalAmount
	}
	return total
}

func calculateOvertime(hours, rate int64) (int64, error) {
	if hours < 0 || rate < 0 {
		return 0, payrollerrors.ErrInvalidMoneyValue
	}
	return hours * rate, nil
}

func mapToResponse(payroll Payroll) PayrollResponse {
	resp := PayrollResponse{
		ID:             payroll.ID.String(),
		CompanyID:      payroll.CompanyID.String(),
		EmployeeID:     payroll.EmployeeID.String(),
		PeriodStart:    payroll.PeriodStart.Format("2006-01-02"),
		PeriodEnd:      payroll.PeriodEnd.Format("2006-01-02"),
		BaseSalary:     payroll.BaseSalary,
		TotalAllowance: payroll.Allowance,
		OvertimeHours:  payroll.OvertimeHours,
		OvertimeRate:   payroll.OvertimeRate,
		TotalOvertime:  payroll.OvertimeAmount,
		TotalDeduction: payroll.Deduction,
		Allowance:      payroll.Allowance,
		Deduction:      payroll.Deduction,
		NetSalary:      payroll.NetSalary,
		Status:         payroll.Status,
		CreatedBy:      payroll.CreatedBy.String(),
	}

	if payroll.Employee != nil {
		resp.EmployeeName = payroll.Employee.FullName
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
	if payroll.PayslipURL != nil {
		resp.PayslipURL = payroll.PayslipURL
	}
	if payroll.PayslipGeneratedAt != nil {
		v := payroll.PayslipGeneratedAt.Format(time.RFC3339)
		resp.PayslipGeneratedAt = &v
	}

	if len(payroll.Components) > 0 {
		resp.Components = make([]PayrollComponentResponse, 0, len(payroll.Components))
		for _, item := range payroll.Components {
			resp.Components = append(resp.Components, PayrollComponentResponse{
				ID:            item.ID.String(),
				ComponentType: item.ComponentType,
				ComponentName: item.ComponentName,
				Quantity:      item.Quantity,
				UnitAmount:    item.UnitAmount,
				TotalAmount:   item.TotalAmount,
				Notes:         item.Notes,
			})
		}
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

func mapToBreakdownResponse(payroll Payroll) PayrollBreakdownResponse {
	allowances := make([]PayrollBreakdownLine, 0)
	deductions := make([]PayrollBreakdownLine, 0)

	for _, component := range payroll.Components {
		quantity := component.Quantity
		unitAmount := component.UnitAmount
		line := PayrollBreakdownLine{
			Label:      component.ComponentName,
			Quantity:   &quantity,
			UnitAmount: &unitAmount,
			Amount:     component.TotalAmount,
			Notes:      component.Notes,
		}
		if component.ComponentType == ComponentTypeAllowance {
			allowances = append(allowances, line)
		}
		if component.ComponentType == ComponentTypeDeduction {
			deductions = append(deductions, line)
		}
	}

	allowanceFromComponents := sumBreakdown(allowances)
	allowanceCore := payroll.Allowance - allowanceFromComponents
	if allowanceCore > 0 {
		allowances = append([]PayrollBreakdownLine{{
			Label:  "Allowance (core)",
			Amount: allowanceCore,
		}}, allowances...)
	}

	deductionFromComponents := sumBreakdown(deductions)
	deductionCore := payroll.Deduction - deductionFromComponents
	if deductionCore > 0 {
		deductions = append([]PayrollBreakdownLine{{
			Label:  "Deduction (core)",
			Amount: deductionCore,
		}}, deductions...)
	}

	overtimeHours := payroll.OvertimeHours
	overtimeRate := payroll.OvertimeRate

	return PayrollBreakdownResponse{
		PayrollID:   payroll.ID.String(),
		EmployeeID:  payroll.EmployeeID.String(),
		PeriodStart: payroll.PeriodStart.Format("2006-01-02"),
		PeriodEnd:   payroll.PeriodEnd.Format("2006-01-02"),
		Status:      payroll.Status,
		BaseSalary: PayrollBreakdownLine{
			Label:  "Base Salary",
			Amount: payroll.BaseSalary,
		},
		Allowances:     allowances,
		AllowanceTotal: payroll.Allowance,
		Overtime: PayrollBreakdownLine{
			Label:      "Overtime",
			Quantity:   &overtimeHours,
			UnitAmount: &overtimeRate,
			Amount:     payroll.OvertimeAmount,
		},
		Deductions:     deductions,
		DeductionTotal: payroll.Deduction,
		NetSalary:      payroll.NetSalary,
	}
}

func sumBreakdown(lines []PayrollBreakdownLine) int64 {
	var total int64
	for _, line := range lines {
		total += line.Amount
	}
	return total
}

func buildPayslipLines(payroll Payroll) []string {
	lines := []string{
		"Payslip",
		fmt.Sprintf("Payroll ID: %s", payroll.ID.String()),
		fmt.Sprintf("Employee ID: %s", payroll.EmployeeID.String()),
		fmt.Sprintf("Period: %s to %s", payroll.PeriodStart.Format("2006-01-02"), payroll.PeriodEnd.Format("2006-01-02")),
		fmt.Sprintf("Status: %s", payroll.Status),
		"",
		fmt.Sprintf("Base Salary: %d", payroll.BaseSalary),
		fmt.Sprintf("Total Allowance: %d", payroll.Allowance),
		fmt.Sprintf("Overtime: %d x %d = %d", payroll.OvertimeHours, payroll.OvertimeRate, payroll.OvertimeAmount),
		fmt.Sprintf("Total Deduction: %d", payroll.Deduction),
		"------------------------------",
		fmt.Sprintf("Net Salary: %d", payroll.NetSalary),
	}

	if len(payroll.Components) > 0 {
		lines = append(lines, "", "Components:")
		for _, c := range payroll.Components {
			lines = append(lines, fmt.Sprintf("- %s %s: %d x %d = %d", c.ComponentType, c.ComponentName, c.Quantity, c.UnitAmount, c.TotalAmount))
		}
	}

	return lines
}
