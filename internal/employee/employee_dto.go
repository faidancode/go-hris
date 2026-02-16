package employee

type CreateEmployeeRequest struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	CompanyID string `json:"company_id" binding:"required"`
}

type EmployeeResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CompanyID string `json:"company_id"`
}
