package company

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

//go:generate mockgen -destination=mock/company_repo_mock.go -package=mock . Repository
type Repository interface {
	Create(ctx context.Context, company *Company) error
	GetByID(ctx context.Context, id uuid.UUID) (*Company, error)
	GetByEmail(ctx context.Context, email string) (*Company, error)
	Update(ctx context.Context, company *Company) error
	WithTx(tx *gorm.DB) Repository
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}

func (r *repository) Create(ctx context.Context, company *Company) error {
	return r.db.WithContext(ctx).Create(company).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Company, error) {
	var company Company
	err := r.db.WithContext(ctx).First(&company, "id = ?", id).Error
	return &company, err
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*Company, error) {
	var company Company
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&company).Error
	return &company, err
}

func (r *repository) Update(ctx context.Context, company *Company) error {
	return r.db.WithContext(ctx).Save(company).Error
}
