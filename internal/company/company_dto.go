package company

type CompanyResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Email              string `json:"email"`
	RegistrationNumber string `json:"registration_number"`
	IsActive           bool   `json:"is_active"`
}

type UpdateCompanyRequest struct {
	Name               string `json:"name"`
	RegistrationNumber string `json:"registration_number"`
	IsActive           *bool  `json:"is_active"`
}
