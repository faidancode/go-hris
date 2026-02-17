package leave_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"go-hris/internal/leave"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakeLeaveRepository struct {
	withTxFn                 func(tx *sql.Tx) leave.Repository
	createFn                 func(ctx context.Context, l *leave.Leave) error
	findAllByCompanyFn       func(ctx context.Context, companyID string) ([]leave.Leave, error)
	findByIDAndCompanyFn     func(ctx context.Context, companyID, id string) (*leave.Leave, error)
	updateFn                 func(ctx context.Context, l *leave.Leave) error
	deleteFn                 func(ctx context.Context, companyID, id string) error
	employeeBelongsToCompany func(ctx context.Context, companyID, employeeID string) (bool, error)
	hasOverlappingPeriodFn   func(ctx context.Context, companyID, employeeID string, startDate, endDate time.Time, excludeID *string) (bool, error)
}

func (f *fakeLeaveRepository) WithTx(tx *sql.Tx) leave.Repository {
	if f.withTxFn != nil {
		return f.withTxFn(tx)
	}
	return f
}

func (f *fakeLeaveRepository) Create(ctx context.Context, l *leave.Leave) error {
	if f.createFn != nil {
		return f.createFn(ctx, l)
	}
	return nil
}

func (f *fakeLeaveRepository) FindAllByCompany(ctx context.Context, companyID string) ([]leave.Leave, error) {
	if f.findAllByCompanyFn != nil {
		return f.findAllByCompanyFn(ctx, companyID)
	}
	return nil, nil
}

func (f *fakeLeaveRepository) FindByIDAndCompany(ctx context.Context, companyID, id string) (*leave.Leave, error) {
	if f.findByIDAndCompanyFn != nil {
		return f.findByIDAndCompanyFn(ctx, companyID, id)
	}
	return nil, nil
}

func (f *fakeLeaveRepository) Update(ctx context.Context, l *leave.Leave) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, l)
	}
	return nil
}

func (f *fakeLeaveRepository) Delete(ctx context.Context, companyID, id string) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, companyID, id)
	}
	return nil
}

func (f *fakeLeaveRepository) EmployeeBelongsToCompany(ctx context.Context, companyID, employeeID string) (bool, error) {
	if f.employeeBelongsToCompany != nil {
		return f.employeeBelongsToCompany(ctx, companyID, employeeID)
	}
	return true, nil
}

func (f *fakeLeaveRepository) HasOverlappingPeriod(ctx context.Context, companyID, employeeID string, startDate, endDate time.Time, excludeID *string) (bool, error) {
	if f.hasOverlappingPeriodFn != nil {
		return f.hasOverlappingPeriodFn(ctx, companyID, employeeID, startDate, endDate, excludeID)
	}
	return false, nil
}

type leaveServiceDeps struct {
	db      *sql.DB
	sqlMock sqlmock.Sqlmock
	service leave.Service
	repo    *fakeLeaveRepository
}

func setupLeaveServiceTest(t *testing.T) *leaveServiceDeps {
	t.Helper()

	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)

	repo := &fakeLeaveRepository{}
	svc := leave.NewService(db, repo)

	return &leaveServiceDeps{
		db:      db,
		sqlMock: sqlMock,
		service: svc,
		repo:    repo,
	}
}

func expectTx(t *testing.T, mock sqlmock.Sqlmock, commit bool) {
	t.Helper()
	mock.ExpectBegin()
	if commit {
		mock.ExpectCommit()
	} else {
		mock.ExpectRollback()
	}
}

func TestLeaveService_Create(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	actorID := uuid.New().String()
	employeeID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		deps := setupLeaveServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, true)
		req := leave.CreateLeaveRequest{
			EmployeeID: employeeID,
			LeaveType:  "ANNUAL",
			StartDate:  "2026-03-01",
			EndDate:    "2026-03-03",
			Reason:     "Family event",
		}

		deps.repo.employeeBelongsToCompany = func(ctx context.Context, cid, eid string) (bool, error) {
			assert.Equal(t, companyID, cid)
			assert.Equal(t, employeeID, eid)
			return true, nil
		}
		deps.repo.hasOverlappingPeriodFn = func(ctx context.Context, cid, eid string, startDate, endDate time.Time, excludeID *string) (bool, error) {
			assert.Nil(t, excludeID)
			assert.Equal(t, "2026-03-01", startDate.Format("2006-01-02"))
			assert.Equal(t, "2026-03-03", endDate.Format("2006-01-02"))
			return false, nil
		}
		deps.repo.createFn = func(ctx context.Context, l *leave.Leave) error {
			assert.Equal(t, uuid.MustParse(companyID), l.CompanyID)
			assert.Equal(t, uuid.MustParse(employeeID), l.EmployeeID)
			assert.Equal(t, uuid.MustParse(actorID), l.CreatedBy)
			assert.Equal(t, "ANNUAL", l.LeaveType)
			assert.Equal(t, 3, l.TotalDays)
			assert.Equal(t, leave.StatusPending, l.Status)
			return nil
		}

		resp, err := deps.service.Create(ctx, companyID, actorID, req)

		assert.NoError(t, err)
		assert.Equal(t, companyID, resp.CompanyID)
		assert.Equal(t, employeeID, resp.EmployeeID)
		assert.Equal(t, actorID, resp.CreatedBy)
		assert.Equal(t, "ANNUAL", resp.LeaveType)
		assert.Equal(t, 3, resp.TotalDays)
		assert.Equal(t, leave.StatusPending, resp.Status)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("negative overlap period", func(t *testing.T) {
		deps := setupLeaveServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, false)
		req := leave.CreateLeaveRequest{
			EmployeeID: employeeID,
			LeaveType:  "ANNUAL",
			StartDate:  "2026-03-01",
			EndDate:    "2026-03-02",
		}

		deps.repo.employeeBelongsToCompany = func(ctx context.Context, cid, eid string) (bool, error) {
			return true, nil
		}
		deps.repo.hasOverlappingPeriodFn = func(ctx context.Context, cid, eid string, startDate, endDate time.Time, excludeID *string) (bool, error) {
			return true, nil
		}

		_, err := deps.service.Create(ctx, companyID, actorID, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "overlapping period")
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})
}

func TestLeaveService_GetAll(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		deps := setupLeaveServiceTest(t)
		defer deps.db.Close()

		employeeID := uuid.New()
		deps.repo.findAllByCompanyFn = func(ctx context.Context, cid string) ([]leave.Leave, error) {
			assert.Equal(t, companyID, cid)
			return []leave.Leave{
				{
					ID:         uuid.New(),
					CompanyID:  uuid.MustParse(companyID),
					EmployeeID: employeeID,
					LeaveType:  "SICK",
					StartDate:  time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
					EndDate:    time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC),
					TotalDays:  2,
					Status:     leave.StatusPending,
					CreatedBy:  uuid.New(),
				},
			}, nil
		}

		resp, err := deps.service.GetAll(ctx, companyID)

		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, employeeID.String(), resp[0].EmployeeID)
		assert.Equal(t, 2, resp[0].TotalDays)
	})

	t.Run("negative repo error", func(t *testing.T) {
		deps := setupLeaveServiceTest(t)
		defer deps.db.Close()

		deps.repo.findAllByCompanyFn = func(ctx context.Context, cid string) ([]leave.Leave, error) {
			return nil, errors.New("db error")
		}

		resp, err := deps.service.GetAll(ctx, companyID)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestLeaveService_GetByID(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	id := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		deps := setupLeaveServiceTest(t)
		defer deps.db.Close()

		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, cid, targetID string) (*leave.Leave, error) {
			return &leave.Leave{
				ID:         uuid.MustParse(targetID),
				CompanyID:  uuid.MustParse(cid),
				EmployeeID: uuid.New(),
				LeaveType:  "UNPAID",
				StartDate:  time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
				TotalDays:  1,
				Status:     leave.StatusPending,
				CreatedBy:  uuid.New(),
			}, nil
		}

		resp, err := deps.service.GetByID(ctx, companyID, id)

		assert.NoError(t, err)
		assert.Equal(t, id, resp.ID)
		assert.Equal(t, "UNPAID", resp.LeaveType)
	})

	t.Run("negative repo error", func(t *testing.T) {
		deps := setupLeaveServiceTest(t)
		defer deps.db.Close()

		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, cid, targetID string) (*leave.Leave, error) {
			return nil, errors.New("not found")
		}

		resp, err := deps.service.GetByID(ctx, companyID, id)

		assert.Error(t, err)
		assert.Empty(t, resp.ID)
	})
}

func TestLeaveService_Update(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	actorID := uuid.New().String()
	id := uuid.New().String()
	employeeID := uuid.New().String()

	t.Run("success approved flow", func(t *testing.T) {
		deps := setupLeaveServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, true)
		approvedBy := uuid.New().String()
		req := leave.UpdateLeaveRequest{
			EmployeeID: employeeID,
			LeaveType:  "ANNUAL",
			StartDate:  "2026-06-01",
			EndDate:    "2026-06-03",
			Reason:     "Family trip",
			Status:     leave.StatusApproved,
			ApprovedBy: &approvedBy,
		}

		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, cid, targetID string) (*leave.Leave, error) {
			return &leave.Leave{
				ID:         uuid.MustParse(targetID),
				CompanyID:  uuid.MustParse(cid),
				EmployeeID: uuid.New(),
				CreatedBy:  uuid.MustParse(actorID),
			}, nil
		}
		deps.repo.employeeBelongsToCompany = func(ctx context.Context, cid, eid string) (bool, error) {
			return true, nil
		}
		deps.repo.hasOverlappingPeriodFn = func(ctx context.Context, cid, eid string, startDate, endDate time.Time, excludeID *string) (bool, error) {
			assert.NotNil(t, excludeID)
			assert.Equal(t, id, *excludeID)
			return false, nil
		}
		deps.repo.updateFn = func(ctx context.Context, l *leave.Leave) error {
			assert.Equal(t, leave.StatusApproved, l.Status)
			assert.Equal(t, 3, l.TotalDays)
			assert.NotNil(t, l.ApprovedBy)
			assert.Equal(t, approvedBy, l.ApprovedBy.String())
			assert.NotNil(t, l.ApprovedAt)
			return nil
		}

		resp, err := deps.service.Update(ctx, companyID, actorID, id, req)

		assert.NoError(t, err)
		assert.Equal(t, leave.StatusApproved, resp.Status)
		assert.Equal(t, 3, resp.TotalDays)
		assert.NotNil(t, resp.ApprovedBy)
		assert.Equal(t, approvedBy, *resp.ApprovedBy)
		assert.NotNil(t, resp.ApprovedAt)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("negative approved without approved_by", func(t *testing.T) {
		deps := setupLeaveServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, false)
		req := leave.UpdateLeaveRequest{
			EmployeeID: employeeID,
			LeaveType:  "ANNUAL",
			StartDate:  "2026-06-01",
			EndDate:    "2026-06-02",
			Status:     leave.StatusApproved,
		}

		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, cid, targetID string) (*leave.Leave, error) {
			return &leave.Leave{
				ID:        uuid.MustParse(targetID),
				CompanyID: uuid.MustParse(cid),
				CreatedBy: uuid.MustParse(actorID),
			}, nil
		}
		deps.repo.employeeBelongsToCompany = func(ctx context.Context, cid, eid string) (bool, error) {
			return true, nil
		}
		deps.repo.hasOverlappingPeriodFn = func(ctx context.Context, cid, eid string, startDate, endDate time.Time, excludeID *string) (bool, error) {
			return false, nil
		}

		_, err := deps.service.Update(ctx, companyID, actorID, id, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "approved_by is required")
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})
}

func TestLeaveService_Delete(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	id := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		deps := setupLeaveServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, true)
		deps.repo.deleteFn = func(ctx context.Context, cid, targetID string) error {
			assert.Equal(t, companyID, cid)
			assert.Equal(t, id, targetID)
			return nil
		}

		err := deps.service.Delete(ctx, companyID, id)

		assert.NoError(t, err)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("negative repo error", func(t *testing.T) {
		deps := setupLeaveServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, false)
		deps.repo.deleteFn = func(ctx context.Context, cid, targetID string) error {
			return errors.New("delete failed")
		}

		err := deps.service.Delete(ctx, companyID, id)

		assert.Error(t, err)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})
}
