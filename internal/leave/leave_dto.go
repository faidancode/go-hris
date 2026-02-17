package leave

type CreateLeaveRequest struct {
	EmployeeID string `json:"employee_id" binding:"required,uuid"`
	LeaveType  string `json:"leave_type" binding:"required,oneof=ANNUAL SICK UNPAID"`
	StartDate  string `json:"start_date" binding:"required"`
	EndDate    string `json:"end_date" binding:"required"`
	Reason     string `json:"reason"`
}

type UpdateLeaveRequest struct {
	EmployeeID      string  `json:"employee_id" binding:"required,uuid"`
	LeaveType       string  `json:"leave_type" binding:"required,oneof=ANNUAL SICK UNPAID"`
	StartDate       string  `json:"start_date" binding:"required"`
	EndDate         string  `json:"end_date" binding:"required"`
	Reason          string  `json:"reason"`
	Status          string  `json:"status" binding:"required,oneof=PENDING APPROVED REJECTED CANCELLED"`
	ApprovedBy      *string `json:"approved_by"`
	RejectionReason *string `json:"rejection_reason"`
}

type LeaveResponse struct {
	ID              string  `json:"id"`
	CompanyID       string  `json:"company_id"`
	EmployeeID      string  `json:"employee_id"`
	LeaveType       string  `json:"leave_type"`
	StartDate       string  `json:"start_date"`
	EndDate         string  `json:"end_date"`
	TotalDays       int     `json:"total_days"`
	Reason          string  `json:"reason"`
	Status          string  `json:"status"`
	CreatedBy       string  `json:"created_by"`
	ApprovedBy      *string `json:"approved_by,omitempty"`
	ApprovedAt      *string `json:"approved_at,omitempty"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
}
