package employeesalary

type CreateEmployeeSalaryRequest struct {
	EmployeeID    string `json:"employee_id" binding:"required"`
	BaseSalary    int    `json:"base_salary" binding:"required"`
	EffectiveDate string `json:"effective_date" binding:"required"`
}

type UpdateEmployeeSalaryRequest struct {
	EmployeeID    string `json:"employee_id" binding:"required"`
	BaseSalary    int    `json:"base_salary" binding:"required"`
	EffectiveDate string `json:"effective_date" binding:"required"`
}

type EmployeeSalaryResponse struct {
	ID            string `json:"id"`
	EmployeeID    string `json:"employee_id"`
	EmployeeName  string `json:"employee_name,omitempty"`
	BaseSalary    int    `json:"base_salary"`
	EffectiveDate string `json:"effective_date"`
}
