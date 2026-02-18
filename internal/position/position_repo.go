package position

import (
	"context"
	"database/sql"
	"go-hris/internal/tenant"

	"gorm.io/gorm"
)

//go:generate mockgen -source=position_repo.go -destination=mock/position_repo_mock.go -package=mock
type Repository interface {
	WithTx(tx *sql.Tx) Repository
	Create(ctx context.Context, dept *Position) error
	FindAllByCompany(ctx context.Context, companyID string) ([]Position, error)
	FindByIDAndCompany(ctx context.Context, companyID string, id string) (*Position, error)
	Update(ctx context.Context, dept *Position) error
	Delete(ctx context.Context, companyID string, id string) error
}

type repository struct {
	db *gorm.DB
	tx *sql.Tx
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) WithTx(tx *sql.Tx) Repository {
	return &repository{
		db: r.db,
		tx: tx,
	}
}

func (r *repository) Create(ctx context.Context, dept *Position) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *repository) FindAllByCompany(ctx context.Context, companyID string) ([]Position, error) {
	var depts []Position
	err := r.db.WithContext(ctx).
		Preload("Department").
		Scopes(tenant.Scope(companyID)).
		Find(&depts).Error
	return depts, err
}

func (r *repository) FindByIDAndCompany(ctx context.Context, companyID string, id string) (*Position, error) {
	var dept Position
	err := r.db.WithContext(ctx).
		Preload("Department").
		Scopes(tenant.Scope(companyID)).
		First(&dept, "id = ?", id).Error
	return &dept, err
}

func (r *repository) Update(ctx context.Context, dept *Position) error {
	// Avoid persisting preloaded Department association on update.
	return r.db.WithContext(ctx).Omit("Department").Save(dept).Error
}

func (r *repository) Delete(ctx context.Context, companyID string, id string) error {
	return r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		Delete(&Position{}, "id = ?", id).Error
}
