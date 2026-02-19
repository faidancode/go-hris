package employee

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Employee struct {
	ID               uuid.UUID           `gorm:"column:id;type:uuid;primaryKey"`
	CompanyID        uuid.UUID           `gorm:"column:company_id;type:uuid;index"`
	DepartmentID     *uuid.UUID          `gorm:"column:department_id;type:uuid"`
	PositionID       *uuid.UUID          `gorm:"column:position_id;type:uuid"`
	EmployeeNumber   string              `gorm:"column:employee_number"`
	FullName         string              `gorm:"column:full_name"`
	Email            string              `gorm:"column:email;uniqueIndex"`
	Phone            string              `gorm:"column:phone"`
	HireDate         time.Time           `gorm:"column:hire_date;type:date"`
	EmploymentStatus string              `gorm:"column:employment_status"`
	CreatedAt        time.Time           `gorm:"column:created_at"`
	UpdatedAt        time.Time           `gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt      `gorm:"column:deleted_at;index"`
	Department       *EmployeeDepartment `gorm:"foreignKey:DepartmentID;references:ID"`
	Position         *EmployeePosition   `gorm:"foreignKey:PositionID;references:ID"`
}

type EmployeeDepartment struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name string    `gorm:"column:name"`
}

func (EmployeeDepartment) TableName() string {
	return "departments"
}

type EmployeePosition struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name string    `gorm:"column:name"`
}

func (EmployeePosition) TableName() string {
	return "positions"
}
