package attendance

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type fakeRepo struct {
	withTxFn                 func(tx *sql.Tx) Repository
	createFn                 func(ctx context.Context, a *Attendance) error
	findByEmployeeAndDateFn  func(ctx context.Context, companyID, employeeID string, date time.Time) (*Attendance, error)
	findAllByCompanyFn       func(ctx context.Context, companyID string) ([]Attendance, error)
	updateFn                 func(ctx context.Context, a *Attendance) error
}

func (f *fakeRepo) WithTx(tx *sql.Tx) Repository { return f.withTxFn(tx) }
func (f *fakeRepo) Create(ctx context.Context, a *Attendance) error { return f.createFn(ctx, a) }
func (f *fakeRepo) FindByEmployeeAndDate(ctx context.Context, companyID, employeeID string, date time.Time) (*Attendance, error) {
	return f.findByEmployeeAndDateFn(ctx, companyID, employeeID, date)
}
func (f *fakeRepo) FindAllByCompany(ctx context.Context, companyID string) ([]Attendance, error) {
	return f.findAllByCompanyFn(ctx, companyID)
}
func (f *fakeRepo) Update(ctx context.Context, a *Attendance) error { return f.updateFn(ctx, a) }

func TestService_ClockInAndClockOut(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	companyID := uuid.New().String()
	employeeID := uuid.New().String()
	ctx := context.Background()

	var saved Attendance
	repo := &fakeRepo{}
	repo.withTxFn = func(tx *sql.Tx) Repository { return repo }
	repo.createFn = func(ctx context.Context, a *Attendance) error { saved = *a; return nil }
	repo.updateFn = func(ctx context.Context, a *Attendance) error { saved = *a; return nil }
	repo.findAllByCompanyFn = func(ctx context.Context, companyID string) ([]Attendance, error) { return nil, nil }
	repo.findByEmployeeAndDateFn = func(ctx context.Context, companyID, employeeID string, date time.Time) (*Attendance, error) {
		if saved.ID == uuid.Nil {
			return nil, gorm.ErrRecordNotFound
		}
		return &saved, nil
	}

	svc := NewService(db, repo)

	mock.ExpectBegin()
	mock.ExpectCommit()
	inResp, err := svc.ClockIn(ctx, companyID, employeeID, ClockInRequest{})
	assert.NoError(t, err)
	assert.NotEmpty(t, inResp.ID)

	mock.ExpectBegin()
	mock.ExpectCommit()
	outResp, err := svc.ClockOut(ctx, companyID, employeeID, ClockOutRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, outResp.ClockOut)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ClockIn_Duplicate(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	companyID := uuid.New().String()
	employeeID := uuid.New().String()
	ctx := context.Background()

	repo := &fakeRepo{}
	repo.withTxFn = func(tx *sql.Tx) Repository { return repo }
	repo.createFn = func(ctx context.Context, a *Attendance) error { return nil }
	repo.updateFn = func(ctx context.Context, a *Attendance) error { return nil }
	repo.findAllByCompanyFn = func(ctx context.Context, companyID string) ([]Attendance, error) { return nil, nil }
	repo.findByEmployeeAndDateFn = func(ctx context.Context, companyID, employeeID string, date time.Time) (*Attendance, error) {
		return &Attendance{ID: uuid.New()}, nil
	}

	svc := NewService(db, repo)
	mock.ExpectBegin()
	mock.ExpectRollback()
	_, err := svc.ClockIn(ctx, companyID, employeeID, ClockInRequest{})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.New("already clocked in for today")) || err.Error() == "already clocked in for today")
	assert.NoError(t, mock.ExpectationsWereMet())
}
