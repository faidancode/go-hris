package leave

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending  = "PENDING"
	StatusApproved = "APPROVED"
	StatusRejected = "REJECTED"
	StatusCanceled = "CANCELLED"
)

//go:generate mockgen -source=leave_service.go -destination=mock/leave_service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, companyID, actorID string, req CreateLeaveRequest) (LeaveResponse, error)
	GetAll(ctx context.Context, companyID string) ([]LeaveResponse, error)
	GetByID(ctx context.Context, companyID, id string) (LeaveResponse, error)
	Update(ctx context.Context, companyID, actorID, id string, req UpdateLeaveRequest) (LeaveResponse, error)
	Delete(ctx context.Context, companyID, id string) error
}

type service struct {
	db   *sql.DB
	repo Repository
}

func NewService(db *sql.DB, repo Repository) Service {
	return &service{db: db, repo: repo}
}

func (s *service) Create(ctx context.Context, companyID, actorID string, req CreateLeaveRequest) (LeaveResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return LeaveResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	companyUUID, employeeUUID, createdByUUID, startDate, endDate, err := validateCreateRequest(companyID, actorID, req)
	if err != nil {
		return LeaveResponse{}, err
	}

	belongs, err := qtx.EmployeeBelongsToCompany(ctx, companyID, req.EmployeeID)
	if err != nil {
		return LeaveResponse{}, err
	}
	if !belongs {
		return LeaveResponse{}, errors.New("employee does not belong to this company")
	}

	overlap, err := qtx.HasOverlappingPeriod(ctx, companyID, req.EmployeeID, startDate, endDate, nil)
	if err != nil {
		return LeaveResponse{}, err
	}
	if overlap {
		return LeaveResponse{}, errors.New("leave already exists in overlapping period")
	}

	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
	l := &Leave{
		ID:         uuid.New(),
		CompanyID:  companyUUID,
		EmployeeID: employeeUUID,
		LeaveType:  req.LeaveType,
		StartDate:  startDate,
		EndDate:    endDate,
		TotalDays:  totalDays,
		Reason:     req.Reason,
		Status:     StatusPending,
		CreatedBy:  createdByUUID,
	}

	if err := qtx.Create(ctx, l); err != nil {
		return LeaveResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return LeaveResponse{}, err
	}

	return mapToResponse(*l), nil
}

func (s *service) GetAll(ctx context.Context, companyID string) ([]LeaveResponse, error) {
	leaves, err := s.repo.FindAllByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}
	return mapToListResponse(leaves), nil
}

func (s *service) GetByID(ctx context.Context, companyID, id string) (LeaveResponse, error) {
	l, err := s.repo.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return LeaveResponse{}, err
	}
	return mapToResponse(*l), nil
}

func (s *service) Update(ctx context.Context, companyID, actorID, id string, req UpdateLeaveRequest) (LeaveResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return LeaveResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	if _, err = uuid.Parse(companyID); err != nil {
		return LeaveResponse{}, errors.New("invalid company id")
	}
	if _, err = uuid.Parse(actorID); err != nil {
		return LeaveResponse{}, errors.New("invalid actor id")
	}

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return LeaveResponse{}, errors.New("invalid employee id")
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return LeaveResponse{}, err
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		return LeaveResponse{}, err
	}
	if startDate.After(endDate) {
		return LeaveResponse{}, errors.New("start_date must be before or equal end_date")
	}

	l, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return LeaveResponse{}, err
	}

	belongs, err := qtx.EmployeeBelongsToCompany(ctx, companyID, req.EmployeeID)
	if err != nil {
		return LeaveResponse{}, err
	}
	if !belongs {
		return LeaveResponse{}, errors.New("employee does not belong to this company")
	}

	overlap, err := qtx.HasOverlappingPeriod(ctx, companyID, req.EmployeeID, startDate, endDate, &id)
	if err != nil {
		return LeaveResponse{}, err
	}
	if overlap {
		return LeaveResponse{}, errors.New("leave already exists in overlapping period")
	}

	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
	l.EmployeeID = employeeID
	l.LeaveType = req.LeaveType
	l.StartDate = startDate
	l.EndDate = endDate
	l.TotalDays = totalDays
	l.Reason = req.Reason
	l.Status = req.Status

	if req.ApprovedBy != nil && *req.ApprovedBy != "" {
		approverID, err := uuid.Parse(*req.ApprovedBy)
		if err != nil {
			return LeaveResponse{}, errors.New("invalid approved_by")
		}
		l.ApprovedBy = &approverID
		now := time.Now().UTC()
		l.ApprovedAt = &now
	} else if req.Status == StatusApproved {
		return LeaveResponse{}, errors.New("approved_by is required when status is APPROVED")
	}

	if req.Status == StatusRejected && (req.RejectionReason == nil || *req.RejectionReason == "") {
		return LeaveResponse{}, errors.New("rejection_reason is required when status is REJECTED")
	}
	l.RejectionReason = req.RejectionReason

	if err := qtx.Update(ctx, l); err != nil {
		return LeaveResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return LeaveResponse{}, err
	}

	return mapToResponse(*l), nil
}

func (s *service) Delete(ctx context.Context, companyID, id string) error {
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

func validateCreateRequest(companyID, actorID string, req CreateLeaveRequest) (uuid.UUID, uuid.UUID, uuid.UUID, time.Time, time.Time, error) {
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
	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, err
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, err
	}
	if startDate.After(endDate) {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, errors.New("start_date must be before or equal end_date")
	}
	return companyUUID, employeeUUID, createdByUUID, startDate, endDate, nil
}

func parseDate(v string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, errors.New("invalid date format, expected YYYY-MM-DD")
	}
	return t, nil
}

func mapToResponse(l Leave) LeaveResponse {
	resp := LeaveResponse{
		ID:         l.ID.String(),
		CompanyID:  l.CompanyID.String(),
		EmployeeID: l.EmployeeID.String(),
		LeaveType:  l.LeaveType,
		StartDate:  l.StartDate.Format("2006-01-02"),
		EndDate:    l.EndDate.Format("2006-01-02"),
		TotalDays:  l.TotalDays,
		Reason:     l.Reason,
		Status:     l.Status,
		CreatedBy:  l.CreatedBy.String(),
	}
	if l.ApprovedBy != nil {
		v := l.ApprovedBy.String()
		resp.ApprovedBy = &v
	}
	if l.ApprovedAt != nil {
		v := l.ApprovedAt.Format(time.RFC3339)
		resp.ApprovedAt = &v
	}
	resp.RejectionReason = l.RejectionReason
	return resp
}

func mapToListResponse(leaves []Leave) []LeaveResponse {
	resp := make([]LeaveResponse, len(leaves))
	for i, l := range leaves {
		resp[i] = mapToResponse(l)
	}
	return resp
}
