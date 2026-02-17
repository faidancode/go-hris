package payroll

type CreatePayrollRequest struct {
	EmployeeID  string `json:"employee_id" binding:"required,uuid"`
	PeriodStart string `json:"period_start" binding:"required"`
	PeriodEnd   string `json:"period_end" binding:"required"`
	BaseSalary  int64  `json:"base_salary" binding:"required"`
	Allowance   int64  `json:"allowance"`
	Deduction   int64  `json:"deduction"`
}

type UpdatePayrollRequest struct {
	EmployeeID  string  `json:"employee_id" binding:"required,uuid"`
	PeriodStart string  `json:"period_start" binding:"required"`
	PeriodEnd   string  `json:"period_end" binding:"required"`
	BaseSalary  int64   `json:"base_salary" binding:"required"`
	Allowance   int64   `json:"allowance"`
	Deduction   int64   `json:"deduction"`
	Status      string  `json:"status" binding:"required,oneof=DRAFT PROCESSED PAID CANCELLED"`
	ApprovedBy  *string `json:"approved_by"`
	PaidAt      *string `json:"paid_at"`
}

type PayrollResponse struct {
	ID          string  `json:"id"`
	CompanyID   string  `json:"company_id"`
	EmployeeID  string  `json:"employee_id"`
	PeriodStart string  `json:"period_start"`
	PeriodEnd   string  `json:"period_end"`
	BaseSalary  int64   `json:"base_salary"`
	Allowance   int64   `json:"allowance"`
	Deduction   int64   `json:"deduction"`
	NetSalary   int64   `json:"net_salary"`
	Status      string  `json:"status"`
	CreatedBy   string  `json:"created_by"`
	PaidAt      *string `json:"paid_at,omitempty"`
	ApprovedBy  *string `json:"approved_by,omitempty"`
	ApprovedAt  *string `json:"approved_at,omitempty"`
}
