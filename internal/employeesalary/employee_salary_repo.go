package employeesalary

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

//go:generate mockgen -source=employee_salary_repo.go -destination=mock/employee_salary_repo_mock.go -package=mock
type Repository interface {
	WithTx(tx *sql.Tx) Repository
	Create(ctx context.Context, salary *EmployeeSalary) error
	FindAllByCompany(ctx context.Context, companyID string) ([]EmployeeSalary, error)
	FindByIDAndCompany(ctx context.Context, companyID string, id string) (*EmployeeSalary, error)
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

func (r *repository) Create(ctx context.Context, salary *EmployeeSalary) error {
	return r.db.WithContext(ctx).Create(salary).Error
}

func (r *repository) FindAllByCompany(ctx context.Context, companyID string) ([]EmployeeSalary, error) {
	var salaries []EmployeeSalary
	query := `
SELECT
	employee_salaries.*,
	employees.full_name AS employee_name
FROM employee_salaries
JOIN employees ON employees.id = employee_salaries.employee_id
WHERE employees.company_id = ?
ORDER BY
	employees.full_name ASC,
	employee_salaries.effective_date DESC,
	employee_salaries.created_at DESC
`

	err := r.db.WithContext(ctx).Raw(query, companyID).Scan(&salaries).Error
	return salaries, err
}

func (r *repository) FindByIDAndCompany(ctx context.Context, companyID string, id string) (*EmployeeSalary, error) {
	var salary EmployeeSalary
	err := r.db.WithContext(ctx).
		Table("employee_salaries").
		Select("employee_salaries.*, employees.full_name AS employee_name").
		Joins("JOIN employees ON employees.id = employee_salaries.employee_id").
		Where("employee_salaries.id = ?", id).
		Where("employees.company_id = ?", companyID).
		First(&salary).Error
	return &salary, err
}

func (r *repository) Delete(ctx context.Context, companyID string, id string) error {
	return r.db.WithContext(ctx).
		Table("employee_salaries").
		Joins("JOIN employees ON employees.id = employee_salaries.employee_id").
		Where("employee_salaries.id = ?", id).
		Where("employees.company_id = ?", companyID).
		Delete(&EmployeeSalary{}).Error
}
