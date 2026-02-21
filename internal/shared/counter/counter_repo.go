package counter

import (
	"context"

	"gorm.io/gorm"
)

//go:generate mockgen -destination=mock/counter_repo_mock.go -package=mock . Repository
type Repository interface {
	GetNextValue(ctx context.Context, companyID string, counterType string) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetNextValue(ctx context.Context, companyID string, counterType string) (int64, error) {
	var nextValue int64

	// Use raw SQL for atomic UPSERT and increment to handle race conditions per company/type
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO company_counters (company_id, counter_type, last_value, updated_at)
		VALUES (?, ?, 1, now())
		ON CONFLICT (company_id, counter_type) DO UPDATE
		SET last_value = company_counters.last_value + 1, updated_at = now()
		RETURNING last_value
	`, companyID, counterType).Scan(&nextValue).Error

	if err != nil {
		return 0, err
	}

	return nextValue, nil
}
