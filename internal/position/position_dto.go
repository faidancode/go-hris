package position

type CreatePositionRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	DepartmentID string `json:"department_id" binding:"required"`
}

type UpdatePositionRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	DepartmentID string `json:"department_id" binding:"required"`
}

type PositionResponse struct {
	ID           string `json:"id"`
	CompanyID    string `json:"company_id"`
	DepartmentID string `json:"department_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
