package attendance

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

//go:generate mockgen -source=attendance_repo.go -destination=mock/attendance_repo_mock.go -package=mock
type Repository interface {
	WithTx(tx *sql.Tx) Repository
	Create(ctx context.Context, a *Attendance) error
	FindByEmployeeAndDate(ctx context.Context, companyID, employeeID string, date time.Time) (*Attendance, error)
	FindAllByCompany(ctx context.Context, companyID string) ([]Attendance, error)
	Update(ctx context.Context, a *Attendance) error
}

type repository struct {
	db *gorm.DB
	tx *sql.Tx
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) WithTx(tx *sql.Tx) Repository {
	return &repository{db: r.db, tx: tx}
}

func (r *repository) Create(ctx context.Context, a *Attendance) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *repository) FindByEmployeeAndDate(ctx context.Context, companyID, employeeID string, date time.Time) (*Attendance, error) {
	var a Attendance
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Where("employee_id = ?", employeeID).
		Where("attendance_date = ?", date.Format("2006-01-02")).
		First(&a).Error
	return &a, err
}

func (r *repository) FindAllByCompany(ctx context.Context, companyID string) ([]Attendance, error) {
	var rows []Attendance
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("attendance_date DESC, clock_in DESC").
		Find(&rows).Error
	return rows, err
}

func (r *repository) Update(ctx context.Context, a *Attendance) error {
	return r.db.WithContext(ctx).Save(a).Error
}
