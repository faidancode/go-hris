package attendance

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Attendance struct {
	ID             uuid.UUID      `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID      uuid.UUID      `gorm:"column:company_id;type:uuid;not null;index"`
	EmployeeID     uuid.UUID      `gorm:"column:employee_id;type:uuid;not null;index"`
	AttendanceDate time.Time      `gorm:"column:attendance_date;type:date;not null;index"`
	ClockIn        time.Time      `gorm:"column:clock_in;type:timestamptz;not null"`
	ClockOut       *time.Time     `gorm:"column:clock_out;type:timestamptz"`
	Latitude       *float64       `gorm:"column:latitude"`
	Longitude      *float64       `gorm:"column:longitude"`
	Status         string         `gorm:"column:status;type:varchar(20);not null;default:PRESENT"`
	Source         string         `gorm:"column:source;type:varchar(30);not null;default:MANUAL"`
	ExternalRef    *string        `gorm:"column:external_ref;type:varchar(100)"`
	Notes          *string        `gorm:"column:notes;type:text"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
	Employee       *EmployeeRef   `gorm:"foreignKey:EmployeeID;references:ID"`
}

func (Attendance) TableName() string {
	return "attendances"
}

type EmployeeRef struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	FullName string    `gorm:"column:full_name"`
}

func (EmployeeRef) TableName() string {
	return "employees"
}
