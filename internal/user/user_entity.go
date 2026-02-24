package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID         uuid.UUID      `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID  uuid.UUID      `gorm:"column:company_id;type:uuid;not null;index"`
	EmployeeID uuid.UUID      `gorm:"column:employee_id;type:uuid;not null"`
	Name       string         `gorm:"column:name;type:varchar(255)"`
	Role       string         `gorm:"column:role;type:varchar(50);default:EMPLOYEE"`
	Email      string         `gorm:"column:email;type:text;not null;uniqueIndex"`
	Password   string         `gorm:"column:password;type:text;not null"`
	IsActive   bool           `gorm:"column:is_active;default:true"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relasi ke Employee (untuk mengambil CompanyID atau data Profile)
	Employee *UserEmployee `gorm:"foreignKey:EmployeeID;references:ID"`
}

// UserEmployee adalah sub-struct untuk join data minimal dari employee
type UserEmployee struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	CompanyID uuid.UUID `gorm:"column:company_id"`
	FullName  string    `gorm:"column:full_name"`
}

func (UserEmployee) TableName() string {
	return "employees"
}
