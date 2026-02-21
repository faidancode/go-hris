package employee

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go-hris/internal/events"
	"go-hris/internal/messaging/kafka"
	"go-hris/internal/shared/counter"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
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
	db      *sql.DB
	repo    Repository
	counter counter.Repository
	outbox  kafka.OutboxRepository
	logger  *zap.Logger
}

func NewService(db *sql.DB, repo Repository, counter counter.Repository, logger ...*zap.Logger) Service {
	return NewServiceWithOutbox(db, repo, counter, nil, logger...)
}

func NewServiceWithOutbox(
	db *sql.DB,
	repo Repository,
	counter counter.Repository,
	outboxRepo kafka.OutboxRepository,
	logger ...*zap.Logger,
) Service {
	l := zap.L().Named("employee.service")
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0].Named("employee.service")
	}
	return &service{db: db, repo: repo, counter: counter, outbox: outboxRepo, logger: l}
}

func (s *service) Create(
	ctx context.Context,
	companyID string,
	req CreateEmployeeRequest,
) (EmployeeResponse, error) {
	s.logger.Debug("create employee requested",
		zap.String("company_id", companyID),
		zap.String("position_id", req.PositionID),
		zap.String("email", req.Email),
	)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("create employee begin tx failed", zap.Error(err))
		return EmployeeResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)
	departmentID, err := qtx.GetDepartmentIDByPosition(ctx, companyID, req.PositionID)
	if err != nil {
		s.logger.Error("create employee get department by position failed", zap.Error(err))
		return EmployeeResponse{}, err
	}
	if departmentID == "" {
		s.logger.Warn("create employee position not found in company",
			zap.String("company_id", companyID),
			zap.String("position_id", req.PositionID),
		)
		return EmployeeResponse{}, errors.New("position not found for this company")
	}
	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		s.logger.Warn("create employee invalid hire_date",
			zap.String("hire_date", req.HireDate),
			zap.Error(err),
		)
		return EmployeeResponse{}, errors.New("invalid hire_date format, expected YYYY-MM-DD")
	}

	if req.EmployeeNumber == "" {
		nextVal, err := s.counter.GetNextValue(ctx, companyID, "employee_number")
		if err != nil {
			s.logger.Error("create employee generate number failed", zap.Error(err))
			return EmployeeResponse{}, err
		}
		emplNumber := fmt.Sprintf("EMP-%06d", nextVal)
		req.EmployeeNumber = emplNumber
	}

	empl := &Employee{
		ID:               uuid.New(),
		FullName:         req.FullName,
		Email:            req.Email,
		CompanyID:        uuid.MustParse(companyID),
		PositionID:       uuidPtr(req.PositionID),
		DepartmentID:     uuidPtr(departmentID),
		EmployeeNumber:   req.EmployeeNumber,
		Phone:            req.Phone,
		HireDate:         hireDate,
		EmploymentStatus: req.EmploymentStatus,
	}

	if err := qtx.Create(ctx, empl); err != nil {
		s.logger.Error("create employee persist failed", zap.Error(err))
		return EmployeeResponse{}, mapRepositoryError(err)
	}

	event := events.EmployeeCreatedEvent{
		EventType:  "employee_created",
		EmployeeID: empl.ID.String(),
		CompanyID:  companyID,
		OccurredAt: time.Now().UTC(),
	}
	if s.outbox != nil {
		payload, err := json.Marshal(event)
		if err != nil {
			s.logger.Error("marshal employee_created event failed", zap.Error(err))
			return EmployeeResponse{}, err
		}

		outboxRepo := s.outbox.WithTx(tx)
		if err := outboxRepo.Create(ctx, kafka.OutboxEvent{
			ID:            uuid.NewString(),
			AggregateType: "employee",
			AggregateID:   empl.ID.String(),
			EventType:     event.EventType,
			Topic:         events.EmployeeCreatedTopic,
			Payload:       payload,
			Status:        kafka.OutboxStatusPending,
		}); err != nil {
			s.logger.Error("create employee outbox persist failed",
				zap.String("employee_id", empl.ID.String()),
				zap.Error(err),
			)
			return EmployeeResponse{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("create employee commit failed", zap.Error(err))
		return EmployeeResponse{}, err
	}
	if s.outbox != nil {
		s.logger.Info("create employee outbox queued",
			zap.String("employee_id", empl.ID.String()),
		)
	}
	s.logger.Info("create employee success", zap.String("employee_id", empl.ID.String()))

	return mapToResponse(*empl), nil
}

func (s *service) GetAll(
	ctx context.Context,
	companyID string,
) ([]EmployeeResponse, error) {
	s.logger.Debug("get all employees requested", zap.String("company_id", companyID))
	depts, err := s.repo.FindAllByCompany(ctx, companyID)
	if err != nil {
		s.logger.Error("get all employees failed", zap.Error(err))
		return nil, mapRepositoryError(err)
	}

	return mapToListResponse(depts), nil
}

func (s *service) GetByID(
	ctx context.Context,
	companyID, id string,
) (EmployeeResponse, error) {
	s.logger.Debug("get employee by id requested",
		zap.String("company_id", companyID),
		zap.String("employee_id", id),
	)
	empl, err := s.repo.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		s.logger.Error("get employee by id failed", zap.Error(err))
		return EmployeeResponse{}, mapRepositoryError(err)
	}

	return mapToResponse(*empl), nil
}

func (s *service) Update(
	ctx context.Context,
	companyID, id string,
	req UpdateEmployeeRequest,
) (EmployeeResponse, error) {
	s.logger.Debug("update employee requested",
		zap.String("company_id", companyID),
		zap.String("employee_id", id),
		zap.String("position_id", req.PositionID),
	)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("update employee begin tx failed", zap.Error(err))
		return EmployeeResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)
	departmentID, err := qtx.GetDepartmentIDByPosition(ctx, companyID, req.PositionID)
	if err != nil {
		s.logger.Error("update employee get department by position failed", zap.Error(err))
		return EmployeeResponse{}, err
	}
	if departmentID == "" {
		s.logger.Warn("update employee position not found in company",
			zap.String("company_id", companyID),
			zap.String("position_id", req.PositionID),
		)
		return EmployeeResponse{}, errors.New("position not found for this company")
	}
	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		s.logger.Warn("update employee invalid hire_date",
			zap.String("hire_date", req.HireDate),
			zap.Error(err),
		)
		return EmployeeResponse{}, errors.New("invalid hire_date format, expected YYYY-MM-DD")
	}

	empl, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		s.logger.Error("update employee fetch existing failed", zap.Error(err))
		return EmployeeResponse{}, mapRepositoryError(err)
	}

	empl.FullName = req.FullName
	empl.Email = req.Email
	empl.PositionID = uuidPtr(req.PositionID)
	empl.DepartmentID = uuidPtr(departmentID)
	empl.EmployeeNumber = req.EmployeeNumber
	empl.Phone = req.Phone
	empl.HireDate = hireDate
	empl.EmploymentStatus = req.EmploymentStatus

	if err := qtx.Update(ctx, empl); err != nil {
		s.logger.Error("update employee persist failed", zap.Error(err))
		return EmployeeResponse{}, mapRepositoryError(err)
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("update employee commit failed", zap.Error(err))
		return EmployeeResponse{}, err
	}
	s.logger.Info("update employee success", zap.String("employee_id", id))

	return mapToResponse(*empl), nil
}

func (s *service) Delete(
	ctx context.Context,
	companyID, id string,
) error {
	s.logger.Debug("delete employee requested",
		zap.String("company_id", companyID),
		zap.String("employee_id", id),
	)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("delete employee begin tx failed", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	if err := qtx.Delete(ctx, companyID, id); err != nil {
		s.logger.Error("delete employee failed", zap.Error(err))
		return mapRepositoryError(err)
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("delete employee commit failed", zap.Error(err))
		return err
	}
	s.logger.Info("delete employee success", zap.String("employee_id", id))
	return nil
}

func mapToResponse(empl Employee) EmployeeResponse {
	resp := EmployeeResponse{
		ID:               empl.ID.String(),
		FullName:         empl.FullName,
		Email:            empl.Email,
		EmployeeNumber:   empl.EmployeeNumber,
		Phone:            empl.Phone,
		HireDate:         empl.HireDate.Format("2006-01-02"),
		EmploymentStatus: empl.EmploymentStatus,
		CompanyID:        empl.CompanyID.String(),
		DepartmentID:     uuidToString(empl.DepartmentID),
		PositionID:       uuidToString(empl.PositionID),
	}
	if empl.Department != nil {
		resp.Department = &EmployeeDepartmentResponse{
			ID:   empl.Department.ID.String(),
			Name: empl.Department.Name,
		}
	}
	if empl.Position != nil {
		resp.Position = &EmployeePositionResponse{
			ID:   empl.Position.ID.String(),
			Name: empl.Position.Name,
		}
	}
	return resp
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
