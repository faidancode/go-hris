package attendance

import (
	"context"
	"database/sql"
	"errors"
	"go-hris/internal/shared/apperror"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	statusPresent = "PRESENT"
	statusLate    = "LATE"
)

//go:generate mockgen -source=attendance_service.go -destination=mock/attendance_service_mock.go -package=mock
type Service interface {
	ClockIn(ctx context.Context, companyID, employeeID string, req ClockInRequest) (AttendanceResponse, error)
	ClockOut(ctx context.Context, companyID, employeeID string, req ClockOutRequest) (AttendanceResponse, error)
	GetAll(ctx context.Context, companyID, actorID string, canReadAll bool) ([]AttendanceResponse, error)
}

type service struct {
	db   *sql.DB
	repo Repository
}

func NewService(db *sql.DB, repo Repository) Service {
	return &service{db: db, repo: repo}
}

func (s *service) ClockIn(ctx context.Context, companyID, employeeID string, req ClockInRequest) (AttendanceResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AttendanceResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)
	now := time.Now().UTC()
	today := now.Truncate(24 * time.Hour)

	existing, err := qtx.FindByEmployeeAndDate(ctx, companyID, employeeID, today)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return AttendanceResponse{}, err
	}
	if err == nil && existing != nil {
		return AttendanceResponse{}, errors.New("already clocked in for today")
	}

	status := statusPresent
	if now.Hour() > 9 || (now.Hour() == 9 && now.Minute() > 15) {
		status = statusLate
	}

	source := req.Source
	if source == "" {
		source = "MANUAL"
	}

	row := &Attendance{
		ID:             uuid.New(),
		CompanyID:      uuid.MustParse(companyID),
		EmployeeID:     uuid.MustParse(employeeID),
		AttendanceDate: today,
		ClockIn:        now,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		Status:         status,
		Source:         source,
		Notes:          req.Notes,
	}

	if err := qtx.Create(ctx, row); err != nil {
		return AttendanceResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return AttendanceResponse{}, err
	}
	return mapToResponse(*row), nil
}

func (s *service) ClockOut(ctx context.Context, companyID, employeeID string, req ClockOutRequest) (AttendanceResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AttendanceResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)
	now := time.Now().UTC()
	today := now.Truncate(24 * time.Hour)

	row, err := qtx.FindByEmployeeAndDate(ctx, companyID, employeeID, today)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AttendanceResponse{}, errors.New("clock in not found for today")
		}
		return AttendanceResponse{}, err
	}
	if row.ClockOut != nil {
		return AttendanceResponse{}, errors.New("already clocked out for today")
	}

	row.ClockOut = &now
	if req.Latitude != nil {
		row.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		row.Longitude = req.Longitude
	}
	if req.Notes != nil {
		row.Notes = req.Notes
	}

	if err := qtx.Update(ctx, row); err != nil {
		return AttendanceResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return AttendanceResponse{}, err
	}
	return mapToResponse(*row), nil
}

func (s *service) GetAll(ctx context.Context, companyID, actorID string, canReadAll bool) ([]AttendanceResponse, error) {
	var (
		rows []Attendance
		err  error
	)
	if canReadAll {
		rows, err = s.repo.FindAllByCompany(ctx, companyID)
	} else {
		if _, parseErr := uuid.Parse(actorID); parseErr != nil {
			return nil, apperror.New(apperror.CodeInvalidInput, "invalid actor id", 400)
		}
		rows, err = s.repo.FindAllByCompanyAndEmployee(ctx, companyID, actorID)
	}
	if err != nil {
		return nil, err
	}
	res := make([]AttendanceResponse, len(rows))
	for i, r := range rows {
		res[i] = mapToResponse(r)
	}
	return res, nil
}

func mapToResponse(a Attendance) AttendanceResponse {
	resp := AttendanceResponse{
		ID:             a.ID.String(),
		CompanyID:      a.CompanyID.String(),
		EmployeeID:     a.EmployeeID.String(),
		AttendanceDate: a.AttendanceDate.Format("2006-01-02"),
		ClockIn:        a.ClockIn.Format(time.RFC3339),
		Latitude:       a.Latitude,
		Longitude:      a.Longitude,
		Status:         a.Status,
		Source:         a.Source,
		ExternalRef:    a.ExternalRef,
		Notes:          a.Notes,
	}
	if a.ClockOut != nil {
		v := a.ClockOut.Format(time.RFC3339)
		resp.ClockOut = &v
	}
	return resp
}
