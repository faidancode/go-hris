package employee

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

//go:generate mockgen -source=employee_repo.go -destination=mock/employee_repo_mock.go -package=mock
type Repository interface {
	WithTx(tx *sql.Tx) Repository
	Create(ctx context.Context, emp CreateEmployeeRequest) (Employee, error)
	FindAllByCompany(ctx context.Context, companyID string) ([]Employee, error)
	FindByID(ctx context.Context, id string) (*Employee, error)
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

func (r *repository) Create(ctx context.Context, emp CreateEmployeeRequest) (Employee, error) {
	var employee Employee
	employee.Name = emp.Name
	employee.Email = emp.Email
	employee.CompanyID = uuid.MustParse(emp.CompanyID)

	err := r.db.Create(&employee).Error
	return employee, err
}

func (r *repository) FindAllByCompany(ctx context.Context, companyID string) ([]Employee, error) {
	var employees []Employee
	err := r.db.Where("company_id = ?", companyID).Find(&employees).Error
	return employees, err
}

func (r *repository) FindByID(ctx context.Context, id string) (*Employee, error) {
	var emp Employee
	err := r.db.First(&emp, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &emp, nil
}
