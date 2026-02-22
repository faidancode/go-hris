package company

import (
	"time"

	"github.com/google/uuid"
)

type RegistrationType string

const (
	RegistrationTypeNPWP RegistrationType = "NPWP"
	RegistrationTypeNIB  RegistrationType = "NIB"
	RegistrationTypeSIUP RegistrationType = "SIUP"
	RegistrationTypeEIN  RegistrationType = "EIN"
	RegistrationTypeUEN  RegistrationType = "UEN"
)

type CompanyRegistration struct {
	ID        uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID uuid.UUID        `gorm:"type:uuid;not null;index"`
	Type      RegistrationType `gorm:"type:registration_type;not null"`
	Number    string           `gorm:"type:varchar(100);not null"`
	IssuedAt  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
