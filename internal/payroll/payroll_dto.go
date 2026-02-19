package payroll

type GetPayrollsFilterRequest struct {
	Period       string `form:"period"`
	PeriodStart  string `form:"period_start"`
	PeriodEnd    string `form:"period_end"`
	DepartmentID string `form:"department_id"`
	Status       string `form:"status"`
}

type PayrollQueryFilter struct {
	PeriodStart  *string
	PeriodEnd    *string
	DepartmentID *string
	Status       *string
}

type CreatePayrollRequest struct {
	EmployeeID     string                  `json:"employee_id" binding:"required,uuid"`
	PeriodStart    string                  `json:"period_start" binding:"required"`
	PeriodEnd      string                  `json:"period_end" binding:"required"`
	BaseSalary     int64                   `json:"base_salary" binding:"required"`
	Allowance      int64                   `json:"allowance"`
	OvertimeHours  int64                   `json:"overtime_hours"`
	OvertimeRate   int64                   `json:"overtime_rate"`
	Deduction      int64                   `json:"deduction"`
	AllowanceItems []PayrollComponentInput `json:"allowance_items"`
	DeductionItems []PayrollComponentInput `json:"deduction_items"`
}

type RegeneratePayrollRequest struct {
	BaseSalary     int64                   `json:"base_salary" binding:"required"`
	Allowance      int64                   `json:"allowance"`
	OvertimeHours  int64                   `json:"overtime_hours"`
	OvertimeRate   int64                   `json:"overtime_rate"`
	Deduction      int64                   `json:"deduction"`
	AllowanceItems []PayrollComponentInput `json:"allowance_items"`
	DeductionItems []PayrollComponentInput `json:"deduction_items"`
}

type PayrollComponentInput struct {
	ComponentName string  `json:"component_name" binding:"required"`
	Quantity      int64   `json:"quantity"`
	UnitAmount    int64   `json:"unit_amount" binding:"required"`
	Notes         *string `json:"notes"`
}

type PayrollComponentResponse struct {
	ID            string  `json:"id"`
	ComponentType string  `json:"component_type"`
	ComponentName string  `json:"component_name"`
	Quantity      int64   `json:"quantity"`
	UnitAmount    int64   `json:"unit_amount"`
	TotalAmount   int64   `json:"total_amount"`
	Notes         *string `json:"notes,omitempty"`
}

type PayrollBreakdownLine struct {
	Label      string  `json:"label"`
	Quantity   *int64  `json:"quantity,omitempty"`
	UnitAmount *int64  `json:"unit_amount,omitempty"`
	Amount     int64   `json:"amount"`
	Notes      *string `json:"notes,omitempty"`
}

type PayrollBreakdownResponse struct {
	PayrollID      string                 `json:"payroll_id"`
	EmployeeID     string                 `json:"employee_id"`
	PeriodStart    string                 `json:"period_start"`
	PeriodEnd      string                 `json:"period_end"`
	Status         string                 `json:"status"`
	BaseSalary     PayrollBreakdownLine   `json:"base_salary"`
	Allowances     []PayrollBreakdownLine `json:"allowances"`
	AllowanceTotal int64                  `json:"allowance_total"`
	Overtime       PayrollBreakdownLine   `json:"overtime"`
	Deductions     []PayrollBreakdownLine `json:"deductions"`
	DeductionTotal int64                  `json:"deduction_total"`
	NetSalary      int64                  `json:"net_salary"`
}

type PayrollResponse struct {
	ID                 string                     `json:"id"`
	CompanyID          string                     `json:"company_id"`
	EmployeeID         string                     `json:"employee_id"`
	EmployeeName       string                     `json:"employee_name"`
	PeriodStart        string                     `json:"period_start"`
	PeriodEnd          string                     `json:"period_end"`
	BaseSalary         int64                      `json:"base_salary"`
	TotalAllowance     int64                      `json:"total_allowance"`
	OvertimeHours      int64                      `json:"overtime_hours"`
	OvertimeRate       int64                      `json:"overtime_rate"`
	TotalOvertime      int64                      `json:"total_overtime"`
	TotalDeduction     int64                      `json:"total_deduction"`
	Allowance          int64                      `json:"allowance"`
	Deduction          int64                      `json:"deduction"`
	NetSalary          int64                      `json:"net_salary"`
	Status             string                     `json:"status"`
	CreatedBy          string                     `json:"created_by"`
	PaidAt             *string                    `json:"paid_at,omitempty"`
	ApprovedBy         *string                    `json:"approved_by,omitempty"`
	ApprovedAt         *string                    `json:"approved_at,omitempty"`
	PayslipURL         *string                    `json:"payslip_url,omitempty"`
	PayslipGeneratedAt *string                    `json:"payslip_generated_at,omitempty"`
	Components         []PayrollComponentResponse `json:"components,omitempty"`
}
