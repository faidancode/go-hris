package auth

type RegisterRequest struct {
	CompanyID  string `json:"company_id" binding:"required,uuid"`
	EmployeeID string `json:"employee_id" binding:"required,uuid"`
	Email      string `json:"email" binding:"required,email"`
	Name       string `json:"name" binding:"required"`
	Password   string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	ID         string `json:"id"`
	CompanyID  string `json:"company_id"`
	EmployeeID string `json:"employee_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Role       string `json:"role"`
}
