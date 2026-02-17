package leave

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Leave struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID  uuid.UUID `gorm:"type:uuid;not null;index:idx_leaves_company_status"`
	EmployeeID uuid.UUID `gorm:"type:uuid;not null;index:idx_leaves_employee_dates"`

	LeaveType string    `gorm:"type:varchar(30);not null;default:'ANNUAL'"`
	StartDate time.Time `gorm:"type:date;not null;index:idx_leaves_employee_dates"`
	EndDate   time.Time `gorm:"type:date;not null;index:idx_leaves_employee_dates"`
	TotalDays int       `gorm:"type:int;not null;default:1"`
	Reason    string    `gorm:"type:text"`

	Status          string     `gorm:"type:varchar(20);not null;default:'PENDING';index:idx_leaves_company_status"`
	CreatedBy       uuid.UUID  `gorm:"type:uuid;not null"`
	ApprovedBy      *uuid.UUID `gorm:"type:uuid"`
	RejectionReason *string    `gorm:"type:text"`

	CreatedAt  time.Time
	UpdatedAt  time.Time
	ApprovedAt *time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index:idx_leaves_deleted_at"`
}
