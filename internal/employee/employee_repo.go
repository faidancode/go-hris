package employee

import (
	"context"
	"database/sql"
	"go-hris/internal/tenant"
	"time"

	"gorm.io/gorm"
)

//go:generate mockgen -source=employee_repo.go -destination=mock/employee_repo_mock.go -package=mock
type Repository interface {
	WithTx(tx *sql.Tx) Repository
	Create(ctx context.Context, emp *Employee) error
	FindAllByCompany(ctx context.Context, companyID string) ([]Employee, error)
	FindOptionsByCompany(ctx context.Context, companyID string) ([]Employee, error)
	FindByIDAndCompany(ctx context.Context, companyID string, id string) (*Employee, error)
	GetDepartmentIDByPosition(ctx context.Context, companyID, positionID string) (string, error)
	Update(ctx context.Context, emp *Employee) error
	Delete(ctx context.Context, companyID string, id string) error
}

type repository struct {
	db *gorm.DB
	tx *sql.Tx
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) WithTx(tx *sql.Tx) Repository {
	return &repository{
		db: r.db,
		tx: tx,
	}
}

func (r *repository) Create(ctx context.Context, emp *Employee) error {
	if r.tx != nil {
		query := `
INSERT INTO employees (
	id,
	company_id,
	department_id,
	position_id,
	employee_number,
	full_name,
	email,
	phone,
	hire_date,
	employment_status,
	created_at,
	updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
`
		now := time.Now().UTC()
		if emp.CreatedAt.IsZero() {
			emp.CreatedAt = now
		}
		emp.UpdatedAt = now
		_, err := r.tx.ExecContext(
			ctx,
			query,
			emp.ID,
			emp.CompanyID,
			emp.DepartmentID,
			emp.PositionID,
			emp.EmployeeNumber,
			emp.FullName,
			emp.Email,
			emp.Phone,
			emp.HireDate,
			emp.EmploymentStatus,
			emp.CreatedAt,
			emp.UpdatedAt,
		)
		return err
	}
	return r.db.WithContext(ctx).Create(emp).Error
}

func (r *repository) FindAllByCompany(ctx context.Context, companyID string) ([]Employee, error) {
	var emps []Employee
	err := r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		Find(&emps).Error
	return emps, err
}

func (r *repository) FindOptionsByCompany(ctx context.Context, companyID string) ([]Employee, error) {
	var emps []Employee
	err := r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		Select("id", "employee_number", "full_name"). // Hanya ambil field yang diperlukan
		Order("full_name ASC").
		Find(&emps).Error
	return emps, err
}

func (r *repository) FindByIDAndCompany(ctx context.Context, companyID string, id string) (*Employee, error) {
	var emp Employee
	err := r.db.WithContext(ctx).
		Preload("Position").
		Preload("Department").
		Scopes(tenant.Scope(companyID)).
		First(&emp, "id = ?", id).Error
	return &emp, err
}

func (r *repository) GetDepartmentIDByPosition(ctx context.Context, companyID, positionID string) (string, error) {
	if r.tx != nil {
		var departmentID string
		query := `
SELECT department_id::text
FROM positions
WHERE id = $1
  AND company_id = $2
  AND deleted_at IS NULL
LIMIT 1
`
		err := r.tx.QueryRowContext(ctx, query, positionID, companyID).Scan(&departmentID)
		if err == sql.ErrNoRows {
			return "", nil
		}
		return departmentID, err
	}

	var departmentID string
	err := r.db.WithContext(ctx).
		Table("positions").
		Select("department_id").
		Where("id = ?", positionID).
		Scopes(tenant.Scope(companyID)).
		Where("deleted_at IS NULL").
		Scan(&departmentID).Error
	return departmentID, err
}

func (r *repository) Update(ctx context.Context, emp *Employee) error {
	return r.db.WithContext(ctx).Save(emp).Error
}

func (r *repository) Delete(ctx context.Context, companyID string, id string) error {
	return r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		Delete(&Employee{}, "id = ?", id).Error
}
