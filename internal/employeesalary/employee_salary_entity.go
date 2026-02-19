package employeesalary

import (
	"time"

	"github.com/google/uuid"
)

type EmployeeSalary struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	EmployeeID    uuid.UUID `gorm:"type:uuid;index"`
	EmployeeName  string    `gorm:"column:employee_name;->"`
	BaseSalary    int
	EffectiveDate time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
