package leave

import (
	"context"
	"database/sql"
	"errors"
	leaveerrors "go-hris/internal/leave/errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	StatusPending   = "PENDING"
	StatusSubmitted = "SUBMITTED"
	StatusApproved  = "APPROVED"
	StatusRejected  = "REJECTED"
	StatusCanceled  = "CANCELLED"
)

//go:generate mockgen -source=leave_service.go -destination=mock/leave_service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, companyID, actorID string, req CreateLeaveRequest) (LeaveResponse, error)
	GetAll(ctx context.Context, companyID string) ([]LeaveResponse, error)
	GetByID(ctx context.Context, companyID, id string) (LeaveResponse, error)
	Update(ctx context.Context, companyID, actorID, id string, req UpdateLeaveRequest) (LeaveResponse, error)
	Submit(ctx context.Context, companyID, actorID, id string) (LeaveResponse, error)
	Approve(ctx context.Context, companyID, actorID, id string) (LeaveResponse, error)
	Reject(ctx context.Context, companyID, actorID, id, rejectionReason string) (LeaveResponse, error)
	Delete(ctx context.Context, companyID, id string) error
}

type service struct {
	db     *sql.DB
	repo   Repository
	logger *zap.Logger
}

func NewService(db *sql.DB, repo Repository, logger ...*zap.Logger) Service {
	l := zap.L().Named("leave.service")
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0].Named("leave.service")
	}
	return &service{db: db, repo: repo, logger: l}
}

func (s *service) Create(ctx context.Context, companyID, actorID string, req CreateLeaveRequest) (LeaveResponse, error) {
	s.logger.Debug("create leave requested",
		zap.String("company_id", companyID),
		zap.String("actor_id", actorID),
		zap.String("employee_id", req.EmployeeID),
		zap.String("start_date", req.StartDate),
		zap.String("end_date", req.EndDate),
	)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("create leave begin tx failed", zap.Error(err))
		return LeaveResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	companyUUID, employeeUUID, createdByUUID, startDate, endDate, err := validateCreateRequest(companyID, actorID, req)
	if err != nil {
		s.logger.Warn("create leave validation failed", zap.Error(err))
		return LeaveResponse{}, err
	}

	belongs, err := qtx.EmployeeBelongsToCompany(ctx, companyID, req.EmployeeID)
	if err != nil {
		s.logger.Error("create leave employee company check failed", zap.Error(err))
		return LeaveResponse{}, err
	}
	if !belongs {
		return LeaveResponse{}, leaveerrors.ErrEmployeeNotInCompany
	}

	overlap, err := qtx.HasOverlappingPeriod(ctx, companyID, req.EmployeeID, startDate, endDate, nil)
	if err != nil {
		s.logger.Error("create leave overlap check failed", zap.Error(err))
		return LeaveResponse{}, err
	}
	if overlap {
		s.logger.Warn("create leave overlap detected",
			zap.String("company_id", companyID),
			zap.String("employee_id", req.EmployeeID),
			zap.String("start_date", req.StartDate),
			zap.String("end_date", req.EndDate),
		)
		return LeaveResponse{}, leaveerrors.ErrLeaveOverlap
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
		s.logger.Error("create leave persist failed", zap.Error(err))
		return LeaveResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("create leave commit failed", zap.Error(err))
		return LeaveResponse{}, err
	}
	s.logger.Info("create leave success",
		zap.String("leave_id", l.ID.String()),
		zap.String("company_id", companyID),
		zap.String("employee_id", req.EmployeeID),
	)

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LeaveResponse{}, leaveerrors.ErrLeaveNotFound
		}
		return LeaveResponse{}, err
	}
	return mapToResponse(*l), nil
}

func (s *service) Update(ctx context.Context, companyID, actorID, id string, req UpdateLeaveRequest) (LeaveResponse, error) {
	s.logger.Debug("update leave requested",
		zap.String("leave_id", id),
		zap.String("company_id", companyID),
		zap.String("actor_id", actorID),
		zap.String("target_status", req.Status),
	)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("update leave begin tx failed", zap.Error(err))
		return LeaveResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	if _, err = uuid.Parse(companyID); err != nil {
		return LeaveResponse{}, leaveerrors.ErrInvalidCompanyID
	}
	if _, err = uuid.Parse(actorID); err != nil {
		return LeaveResponse{}, leaveerrors.ErrInvalidActorID
	}

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return LeaveResponse{}, leaveerrors.ErrInvalidEmployeeID
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
		return LeaveResponse{}, leaveerrors.ErrInvalidDateRange
	}

	l, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LeaveResponse{}, leaveerrors.ErrLeaveNotFound
		}
		return LeaveResponse{}, err
	}
	currentStatus := l.Status
	if !isAllowedStatusTransition(currentStatus, req.Status) {
		return LeaveResponse{}, leaveerrors.ErrInvalidStatusTransition
	}

	belongs, err := qtx.EmployeeBelongsToCompany(ctx, companyID, req.EmployeeID)
	if err != nil {
		return LeaveResponse{}, err
	}
	if !belongs {
		return LeaveResponse{}, leaveerrors.ErrEmployeeNotInCompany
	}

	overlap, err := qtx.HasOverlappingPeriod(ctx, companyID, req.EmployeeID, startDate, endDate, &id)
	if err != nil {
		return LeaveResponse{}, err
	}
	if overlap {
		return LeaveResponse{}, leaveerrors.ErrLeaveOverlap
	}
	if currentStatus == StatusSubmitted && req.Status == StatusApproved {
		if req.EmployeeID != l.EmployeeID.String() ||
			req.LeaveType != l.LeaveType ||
			!startDate.Equal(l.StartDate) ||
			!endDate.Equal(l.EndDate) ||
			req.Reason != l.Reason {
			return LeaveResponse{}, leaveerrors.ErrSubmittedDetailsImmutable
		}
	}

	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
	l.EmployeeID = employeeID
	l.LeaveType = req.LeaveType
	l.StartDate = startDate
	l.EndDate = endDate
	l.TotalDays = totalDays
	l.Reason = req.Reason
	l.Status = req.Status

	if req.Status == StatusApproved {
		if req.ApprovedBy == nil || *req.ApprovedBy == "" {
			return LeaveResponse{}, leaveerrors.ErrApprovedByRequired
		}
		approverID, err := uuid.Parse(*req.ApprovedBy)
		if err != nil {
			return LeaveResponse{}, leaveerrors.ErrInvalidApprovedBy
		}
		l.ApprovedBy = &approverID
		now := time.Now().UTC()
		l.ApprovedAt = &now
	} else {
		l.ApprovedBy = nil
		l.ApprovedAt = nil
	}
	if req.Status == StatusRejected {
		if req.RejectionReason == nil || *req.RejectionReason == "" {
			return LeaveResponse{}, leaveerrors.ErrRejectionReasonRequired
		}
		l.RejectionReason = req.RejectionReason
	} else {
		l.RejectionReason = nil
	}

	if err := qtx.Update(ctx, l); err != nil {
		s.logger.Error("update leave persist failed",
			zap.String("leave_id", id),
			zap.Error(err),
		)
		return LeaveResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("update leave commit failed",
			zap.String("leave_id", id),
			zap.Error(err),
		)
		return LeaveResponse{}, err
	}
	s.logger.Info("update leave success",
		zap.String("leave_id", id),
		zap.String("status", l.Status),
	)

	return mapToResponse(*l), nil
}

func isAllowedStatusTransition(currentStatus, targetStatus string) bool {
	if currentStatus == targetStatus {
		return currentStatus == StatusPending
	}

	switch currentStatus {
	case StatusPending:
		return targetStatus == StatusSubmitted || targetStatus == StatusCanceled
	case StatusSubmitted:
		return targetStatus == StatusApproved || targetStatus == StatusRejected
	default:
		return false
	}
}

func (s *service) Submit(ctx context.Context, companyID, actorID, id string) (LeaveResponse, error) {
	return s.transitionLeaveStatus(ctx, companyID, actorID, id, StatusSubmitted, nil)
}

func (s *service) Approve(ctx context.Context, companyID, actorID, id string) (LeaveResponse, error) {
	return s.transitionLeaveStatus(ctx, companyID, actorID, id, StatusApproved, nil)
}

func (s *service) Reject(ctx context.Context, companyID, actorID, id, rejectionReason string) (LeaveResponse, error) {
	return s.transitionLeaveStatus(ctx, companyID, actorID, id, StatusRejected, &rejectionReason)
}

func (s *service) transitionLeaveStatus(ctx context.Context, companyID, actorID, id, targetStatus string, rejectionReason *string) (LeaveResponse, error) {
	s.logger.Debug("transition leave status requested",
		zap.String("leave_id", id),
		zap.String("company_id", companyID),
		zap.String("actor_id", actorID),
		zap.String("target_status", targetStatus),
	)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("transition leave status begin tx failed", zap.Error(err))
		return LeaveResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	if _, err = uuid.Parse(companyID); err != nil {
		return LeaveResponse{}, leaveerrors.ErrInvalidCompanyID
	}
	actorUUID, err := uuid.Parse(actorID)
	if err != nil {
		return LeaveResponse{}, leaveerrors.ErrInvalidActorID
	}

	l, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LeaveResponse{}, leaveerrors.ErrLeaveNotFound
		}
		return LeaveResponse{}, err
	}
	if !isAllowedStatusTransition(l.Status, targetStatus) {
		s.logger.Warn("transition leave status invalid",
			zap.String("leave_id", id),
			zap.String("from_status", l.Status),
			zap.String("to_status", targetStatus),
		)
		return LeaveResponse{}, leaveerrors.ErrInvalidStatusTransition
	}

	l.Status = targetStatus
	switch targetStatus {
	case StatusApproved:
		l.ApprovedBy = &actorUUID
		now := time.Now().UTC()
		l.ApprovedAt = &now
		l.RejectionReason = nil
	case StatusRejected:
		if rejectionReason == nil || *rejectionReason == "" {
			return LeaveResponse{}, leaveerrors.ErrRejectionReasonRequired
		}
		l.ApprovedBy = nil
		l.ApprovedAt = nil
		l.RejectionReason = rejectionReason
	default:
		l.ApprovedBy = nil
		l.ApprovedAt = nil
		l.RejectionReason = nil
	}

	if err := qtx.Update(ctx, l); err != nil {
		s.logger.Error("transition leave status persist failed",
			zap.String("leave_id", id),
			zap.String("target_status", targetStatus),
			zap.Error(err),
		)
		return LeaveResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		s.logger.Error("transition leave status commit failed",
			zap.String("leave_id", id),
			zap.Error(err),
		)
		return LeaveResponse{}, err
	}
	s.logger.Info("transition leave status success",
		zap.String("leave_id", id),
		zap.String("status", targetStatus),
	)
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
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, leaveerrors.ErrInvalidCompanyID
	}
	employeeUUID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, leaveerrors.ErrInvalidEmployeeID
	}
	createdByUUID, err := uuid.Parse(actorID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, leaveerrors.ErrInvalidActorID
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
		return uuid.Nil, uuid.Nil, uuid.Nil, time.Time{}, time.Time{}, leaveerrors.ErrInvalidDateRange
	}
	return companyUUID, employeeUUID, createdByUUID, startDate, endDate, nil
}

func parseDate(v string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, leaveerrors.ErrInvalidDateFormat
	}
	return t, nil
}

func mapToResponse(l Leave) LeaveResponse {
	resp := LeaveResponse{
		ID:           l.ID.String(),
		CompanyID:    l.CompanyID.String(),
		EmployeeID:   l.EmployeeID.String(),
		EmployeeName: "",
		LeaveType:    l.LeaveType,
		StartDate:    l.StartDate.Format("2006-01-02"),
		EndDate:      l.EndDate.Format("2006-01-02"),
		TotalDays:    l.TotalDays,
		Reason:       l.Reason,
		Status:       l.Status,
		CreatedBy:    l.CreatedBy.String(),
	}
	if l.Employee != nil {
		resp.EmployeeName = l.Employee.FullName
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
