package rbac

import "gorm.io/gorm"

type Repository interface {
	GetEmployeeRoles(companyID string) ([]EmployeeRoleRow, error)
	GetRolePermissions(companyID string) ([]RolePermissionRow, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

type EmployeeRoleRow struct {
	EmployeeID string
	RoleID     string
}

type RolePermissionRow struct {
	RoleID   string
	Resource string
	Action   string
}

func (r *repository) GetEmployeeRoles(companyID string) ([]EmployeeRoleRow, error) {
	var result []EmployeeRoleRow

	err := r.db.
		Table("employee_roles").
		Select("employee_roles.employee_id, employee_roles.role_id").
		Joins("JOIN roles ON roles.id = employee_roles.role_id").
		Where("roles.company_id = ?", companyID).
		Scan(&result).Error

	return result, err
}

func (r *repository) GetRolePermissions(companyID string) ([]RolePermissionRow, error) {
	var result []RolePermissionRow

	err := r.db.
		Table("role_permissions").
		Select("role_permissions.role_id, permissions.resource, permissions.action").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("roles.company_id = ?", companyID).
		Scan(&result).Error

	return result, err
}
