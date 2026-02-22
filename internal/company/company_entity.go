package company

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Company struct {
	ID            uuid.UUID             `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name          string                `gorm:"type:varchar(150);not null"`
	Email         string                `gorm:"type:varchar(255);index"`
	IsActive      bool                  `gorm:"not null;default:true"`
	CreatedAt     time.Time             `gorm:"not null;default:now()"`
	UpdatedAt     time.Time             `gorm:"not null;default:now()"`
	DeletedAt     gorm.DeletedAt        `gorm:"index"`
	Registrations []CompanyRegistration `gorm:"foreignKey:CompanyID"`
}

func (Company) TableName() string {
	return "companies"
}
