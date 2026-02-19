package auth

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

//go:generate mockgen -source=auth_repo.go -destination=mock/auth_repo_mock.go -package=mock

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return &user, err
	}
	if err := r.resolveEffectiveRole(ctx, &user); err != nil {
		return &user, err
	}
	return &user, err
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return &user, err
	}
	if err := r.resolveEffectiveRole(ctx, &user); err != nil {
		return &user, err
	}
	return &user, err
}

func (r *repository) resolveEffectiveRole(ctx context.Context, user *User) error {
	if user.EmployeeID == nil || *user.EmployeeID == uuid.Nil {
		if user.Role == "" {
			user.Role = "EMPLOYEE"
		}
		user.Role = strings.ToUpper(strings.TrimSpace(user.Role))
		return nil
	}

	var roleName string
	err := r.db.WithContext(ctx).
		Table("employee_roles er").
		Select("roles.name").
		Joins("JOIN roles ON roles.id = er.role_id").
		Where("er.employee_id = ?", *user.EmployeeID).
		Where("roles.company_id = ?", user.CompanyID).
		Order(`
			CASE UPPER(roles.name)
				WHEN 'SUPERADMIN' THEN 1
				WHEN 'OWNER' THEN 2
				WHEN 'ADMIN' THEN 3
				WHEN 'HR' THEN 4
				WHEN 'FINANCE' THEN 5
				WHEN 'MANAGER' THEN 6
				WHEN 'EMPLOYEE' THEN 7
				ELSE 99
			END ASC`).
		Limit(1).
		Scan(&roleName).Error
	if err != nil {
		return err
	}

	if strings.TrimSpace(roleName) == "" {
		roleName = user.Role
	}
	if strings.TrimSpace(roleName) == "" {
		roleName = "EMPLOYEE"
	}
	user.Role = strings.ToUpper(strings.TrimSpace(roleName))
	return nil
}
