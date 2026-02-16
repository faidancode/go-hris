package employee

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID uuid.UUID `gorm:"type:uuid;index"`
	Name      string
	Email     string `gorm:"uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
