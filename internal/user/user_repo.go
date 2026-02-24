package user

import (
	"context"
	"go-hris/internal/tenant"

	"gorm.io/gorm"
)

//go:generate mockgen -source=user_repo.go -destination=mock/user_repo_mock.go -package=mock
type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByID(ctx context.Context, companyID string, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindAllByCompany(ctx context.Context, companyID string) ([]User, error)
	Update(ctx context.Context, u *User) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, u *User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *repository) FindByID(ctx context.Context, id string, companyID string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		Preload("Employee").
		First(&u, "id = ?", id).Error

	return &u, err
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).First(&u, "email = ?", email).Error
	return &u, err
}

func (r *repository) FindAllByCompany(ctx context.Context, companyID string) ([]User, error) {
	var users []User

	err := r.db.WithContext(ctx).
		Joins("Employee").               // GORM otomatis join ke tabel employees
		Scopes(tenant.Scope(companyID)). // Menggunakan Scope untuk filter company_id
		Find(&users).Error

	return users, err
}

func (r *repository) Update(ctx context.Context, u *User) error {
	return r.db.WithContext(ctx).Save(u).Error
}
