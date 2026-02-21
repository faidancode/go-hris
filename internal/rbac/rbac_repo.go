package rbac

import "gorm.io/gorm"

//go:generate mockgen -source=rbac_repo.go -destination=mock/rbac_repo_mock.go -package=mock
type Repository interface {
	GetEmployeeRoles(companyID string) ([]EmployeeRoleRow, error)
	GetRolePermissions(companyID string) ([]RolePermissionRow, error)

	// Management
	ListRoles(companyID string) ([]RoleRow, error)
	GetRoleByID(id string) (*RoleRow, error)
	GetRoleByName(companyID, name string) (*RoleRow, error)
	CreateRole(role *RoleRow) error
	UpdateRole(role *RoleRow) error
	DeleteRole(id string) error

	ListPermissions() ([]PermissionRow, error)
	GetPermissionsByRoleID(roleID string) ([]PermissionRow, error)
	UpdateRolePermissions(roleID string, permIDs []string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

type RoleRow struct {
	ID          string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CompanyID   string `gorm:"type:uuid"`
	Name        string
	Description string
	CreatedAt   string
	UpdatedAt   string
}

type PermissionRow struct {
	ID       string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Resource string
	Action   string
	Label    string
	Category string
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
func (r *repository) ListRoles(companyID string) ([]RoleRow, error) {
	var result []RoleRow
	err := r.db.Where("company_id = ?", companyID).Find(&result).Error
	return result, err
}

func (r *repository) GetRoleByID(id string) (*RoleRow, error) {
	var result RoleRow
	err := r.db.First(&result, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *repository) GetRoleByName(companyID, name string) (*RoleRow, error) {
	var result RoleRow
	err := r.db.Where("company_id = ? AND name = ?", companyID, name).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *repository) CreateRole(role *RoleRow) error {
	return r.db.Create(role).Error
}

func (r *repository) UpdateRole(role *RoleRow) error {
	return r.db.Save(role).Error
}

func (r *repository) DeleteRole(id string) error {
	return r.db.Delete(&RoleRow{}, "id = ?", id).Error
}

func (r *repository) ListPermissions() ([]PermissionRow, error) {
	var result []PermissionRow
	err := r.db.Order("category, label").Find(&result).Error
	return result, err
}

func (r *repository) GetPermissionsByRoleID(roleID string) ([]PermissionRow, error) {
	var result []PermissionRow
	err := r.db.
		Table("permissions").
		Select("permissions.*").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Scan(&result).Error
	return result, err
}

func (r *repository) UpdateRolePermissions(roleID string, permIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Remove existing
		if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
			return err
		}

		// Add new
		for _, pID := range permIDs {
			if err := tx.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, pID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
