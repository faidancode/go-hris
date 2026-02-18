package employee

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CompanyID    uuid.UUID  `gorm:"type:uuid;index"`
	DepartmentID *uuid.UUID `gorm:"type:uuid"`
	PositionID   *uuid.UUID `gorm:"type:uuid"`
	FullName     string
	Email        string `gorm:"uniqueIndex"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
