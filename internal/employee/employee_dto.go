package employee

type CreateEmployeeRequest struct {
	FullName   string `json:"full_name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	PositionID string `json:"position_id" binding:"required,uuid"`
}

type UpdateEmployeeRequest struct {
	FullName   string `json:"full_name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	PositionID string `json:"position_id" binding:"required,uuid"`
}

type EmployeeResponse struct {
	ID           string `json:"id"`
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	CompanyID    string `json:"company_id"`
	DepartmentID string `json:"department_id,omitempty"`
	PositionID   string `json:"position_id,omitempty"`
}
