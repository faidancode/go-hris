package user

type CreateUserRequest struct {
	EmployeeID string `json:"employee_id" binding:"required,uuid"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
}

type UpdateUserStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type UserResponse struct {
	ID         string `json:"id"`
	EmployeeID string `json:"employee_id"`
	Email      string `json:"email"`
	IsActive   bool   `json:"is_active"`
	FullName   string `json:"full_name,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type UserWithRolesResponse struct {
	ID             string   `json:"id"`
	EmployeeID     string   `json:"employee_id"`
	EmployeeNumber string   `json:"employee_number"`
	Email          string   `json:"email"`
	FullName       string   `json:"full_name,omitempty"`
	IsActive       bool     `json:"is_active"`
	Roles          []string `json:"roles"`
	CreatedAt      string   `json:"created_at"`
}

type AssignRoleRequest struct {
	RoleName string `json:"role_name" binding:"required"`
}

type ForceResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
