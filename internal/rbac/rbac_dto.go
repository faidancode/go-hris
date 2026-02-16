package rbac

type EnforceRequest struct {
	EmployeeID string `json:"employee_id"`
	CompanyID  string `json:"company_id"`
	Resource   string `json:"resource"`
	Action     string `json:"action"`
}

type EnforceResponse struct {
	Allowed bool `json:"allowed"`
}
