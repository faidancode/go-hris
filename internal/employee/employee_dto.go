package employee

type CreateEmployeeRequest struct {
	FullName       string `json:"full_name" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	EmployeeNumber string `json:"employee_number" binding:"required"`
	Phone          string `json:"phone"`
	HireDate       string `json:"hire_date" binding:"required"`
	EmploymentStatus string `json:"employment_status" binding:"required"`
	PositionID     string `json:"position_id" binding:"required,uuid"`
}

type UpdateEmployeeRequest struct {
	FullName       string `json:"full_name" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	EmployeeNumber string `json:"employee_number" binding:"required"`
	Phone          string `json:"phone"`
	HireDate       string `json:"hire_date" binding:"required"`
	EmploymentStatus string `json:"employment_status" binding:"required"`
	PositionID     string `json:"position_id" binding:"required,uuid"`
}

type EmployeeResponse struct {
	ID             string `json:"id"`
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	EmployeeNumber string `json:"employee_number"`
	Phone          string `json:"phone,omitempty"`
	HireDate       string `json:"hire_date"`
	EmploymentStatus string `json:"employment_status"`
	CompanyID      string `json:"company_id"`
	DepartmentID   string `json:"department_id,omitempty"`
	PositionID     string `json:"position_id,omitempty"`
}
