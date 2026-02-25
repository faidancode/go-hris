package user

import (
	"context"
	"go-hris/internal/tenant"
	"strings"
	"time"

	"gorm.io/gorm"
)

//go:generate mockgen -source=user_repo.go -destination=mock/user_repo_mock.go -package=mock
type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByID(ctx context.Context, companyID string, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindAllByCompany(ctx context.Context, companyID string) ([]User, error)
	FindAllByCompanyWithRoles(ctx context.Context, companyID string) ([]UserWithRolesRow, error)
	Update(ctx context.Context, u *User) error
}

type UserWithRolesRow struct {
	ID             string
	EmployeeID     string
	EmployeeNumber string
	Email          string
	FullName       string
	IsActive       bool
	CreatedAt      time.Time
	RolesRaw       string `gorm:"column:roles"`
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

func (r *repository) FindByID(ctx context.Context, companyID string, id string) (*User, error) {
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
		Joins("Employee").                        // GORM otomatis join ke tabel employees
		Where("users.company_id = ?", companyID). // Hindari ambiguous column dengan tabel join
		Find(&users).Error

	return users, err
}

func (r *repository) Update(ctx context.Context, u *User) error {
	columns := []string{"IsActive", "UpdatedAt"}

	// Jika password diisi (misal hasil dari hashing di service), sertakan di update
	if u.Password != "" {
		columns = append(columns, "Password")
	}

	return r.db.WithContext(ctx).
		Model(u).
		Select(columns).
		Updates(u).Error
}

func (r *repository) FindAllByCompanyWithRoles(ctx context.Context, companyID string) ([]UserWithRolesRow, error) {
	var rows []UserWithRolesRow

	err := r.db.WithContext(ctx).
		Table("users u").
		Select(`
			u.id,
			u.employee_id,
			u.email,
			e.full_name,
			e.employee_number,
			u.is_active,
			u.created_at,
			COALESCE(string_agg(DISTINCT r.name, ',' ORDER BY r.name), '') AS roles
		`).
		Joins("JOIN employees e ON e.id = u.employee_id").
		Joins("LEFT JOIN employee_roles er ON er.employee_id = u.employee_id").
		Joins("LEFT JOIN roles r ON r.id = er.role_id AND r.company_id = u.company_id").
		Where("u.company_id = ?", companyID).
		Where("u.deleted_at IS NULL").
		Group("u.id, u.employee_id, u.email, e.full_name, e.employee_number, u.is_active, u.created_at").
		Order("u.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for i := range rows {
		rows[i].RolesRaw = strings.TrimSpace(rows[i].RolesRaw)
	}

	return rows, nil
}
