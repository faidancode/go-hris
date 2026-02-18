package rbac

type EnforceRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
	CompanyID  string `json:"company_id" binding:"required"`
	Resource   string `json:"resource" binding:"required"`
	Action     string `json:"action" binding:"required"`
}

type EnforceResponse struct {
	Allowed bool `json:"allowed"`
}
