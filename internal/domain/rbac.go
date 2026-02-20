package domain

type EnforceRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
	CompanyID  string `json:"company_id" binding:"required"`
	Resource   string `json:"resource" binding:"required"`
	Action     string `json:"action" binding:"required"`
}

type EnforceResponse struct {
	Allowed bool `json:"allowed"`
}

type RoleResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type UpdateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type PermissionResponse struct {
	ID       string `json:"id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Label    string `json:"label"`
	Category string `json:"category"`
}
