package company

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate mockgen -destination=mock/company_repo_mock.go -package=mock . Repository
type Repository interface {
	Create(ctx context.Context, company *Company) error
	GetByID(ctx context.Context, id uuid.UUID) (*Company, error)
	GetByEmail(ctx context.Context, email string) (*Company, error)
	Update(ctx context.Context, company *Company) error

	UpsertRegistration(ctx context.Context, reg *CompanyRegistration) error
	GetRegistrationsByCompanyID(ctx context.Context, companyID uuid.UUID) ([]CompanyRegistration, error)
	DeleteRegistration(ctx context.Context, companyID uuid.UUID, regType RegistrationType) error

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

func (r *repository) UpsertRegistration(ctx context.Context, reg *CompanyRegistration) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "company_id"}, {Name: "type"}},
			DoUpdates: clause.AssignmentColumns([]string{"number", "issued_at", "updated_at"}),
		}).
		Create(reg).Error
}

func (r *repository) GetRegistrationsByCompanyID(ctx context.Context, companyID uuid.UUID) ([]CompanyRegistration, error) {
	var regs []CompanyRegistration
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Find(&regs).Error
	return regs, err
}

func (r *repository) DeleteRegistration(ctx context.Context, companyID uuid.UUID, regType RegistrationType) error {
	return r.db.WithContext(ctx).
		Where("company_id = ? AND type = ?", companyID, regType).
		Delete(&CompanyRegistration{}).Error
}
