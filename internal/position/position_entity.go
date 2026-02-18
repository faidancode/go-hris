package position

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Position struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Name         string         `gorm:"size:255;not null"`
	CompanyID    uuid.UUID      `gorm:"type:uuid;not null"`
	DepartmentID uuid.UUID      `gorm:"type:uuid;not null"`
	Department   *PositionDepartment `gorm:"foreignKey:DepartmentID;references:ID"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type PositionDepartment struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name string    `gorm:"column:name"`
}

func (PositionDepartment) TableName() string {
	return "departments"
}
