package payroll

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Payroll struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID  uuid.UUID      `gorm:"type:uuid;not null;index:idx_company_status"`
	EmployeeID uuid.UUID      `gorm:"type:uuid;not null;index:idx_employee_period,unique"`
	Employee   *LeaveEmployee `gorm:"foreignKey:EmployeeID;references:ID"`

	// Periode
	PeriodStart time.Time `gorm:"type:date;not null;index:idx_employee_period,unique"`
	PeriodEnd   time.Time `gorm:"type:date;not null;index:idx_employee_period,unique"`

	// Financials disimpan dalam satuan terkecil (mis: sen) untuk hindari floating error.
	BaseSalary     int64 `gorm:"type:bigint;not null;default:0"`
	Allowance      int64 `gorm:"type:bigint;not null;default:0"`
	OvertimeHours  int64 `gorm:"type:bigint;not null;default:0"`
	OvertimeRate   int64 `gorm:"type:bigint;not null;default:0"`
	OvertimeAmount int64 `gorm:"type:bigint;not null;default:0"`
	Deduction      int64 `gorm:"type:bigint;not null;default:0"`
	NetSalary      int64 `gorm:"type:bigint;not null;default:0"`

	// Workflow & Audit
	Status     string     `gorm:"type:varchar(20);not null;default:'DRAFT';index:idx_company_status"`
	CreatedBy  uuid.UUID  `gorm:"type:uuid;not null"`
	ApprovedBy *uuid.UUID `gorm:"type:uuid"` // Pointer karena bisa null

	CreatedAt          time.Time
	UpdatedAt          time.Time
	PaidAt             *time.Time `gorm:"index"`
	ApprovedAt         *time.Time `gorm:"index"`
	PayslipURL         *string
	PayslipGeneratedAt *time.Time     `gorm:"index"`
	DeletedAt          gorm.DeletedAt `gorm:"index"` // Aktifkan Soft Delete jika perlu

	Components []PayrollComponent `gorm:"foreignKey:PayrollID"`

	// Belongs To Relationships (Optional, untuk Eager Loading)
	// Company  Company  `gorm:"foreignKey:CompanyID"`
	// Employee Employee `gorm:"foreignKey:EmployeeID"`
}

type PayrollComponent struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PayrollID     uuid.UUID `gorm:"type:uuid;not null;index"`
	CompanyID     uuid.UUID `gorm:"type:uuid;not null;index"`
	ComponentType string    `gorm:"type:varchar(20);not null;index"`
	ComponentName string    `gorm:"type:varchar(120);not null"`
	Quantity      int64     `gorm:"type:bigint;not null;default:1"`
	UnitAmount    int64     `gorm:"type:bigint;not null;default:0"`
	TotalAmount   int64     `gorm:"type:bigint;not null;default:0"`
	Notes         *string   `gorm:"type:text"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type LeaveEmployee struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	FullName string    `gorm:"column:full_name"`
}

func (LeaveEmployee) TableName() string {
	return "employees"
}
