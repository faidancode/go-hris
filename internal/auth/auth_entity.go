package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID  uuid.UUID  `gorm:"type:uuid;not null;index"` // Untuk Multi-tenancy
	EmployeeID *uuid.UUID `gorm:"type:uuid;uniqueIndex"`    // Relasi ke data Employee
	Name       string     `gorm:"type:varchar(255);not null"`
	Email      string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password   string     `gorm:"type:varchar(255);not null"`
	Role       string     `gorm:"type:varchar(50);not null;default:'EMPLOYEE'"`
	IsActive   bool       `gorm:"default:true"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}
