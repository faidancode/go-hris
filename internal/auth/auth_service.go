package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	autherrors "go-hris/internal/auth/errors"
	"go-hris/internal/company"
	"go-hris/internal/employee"
	employeeerrors "go-hris/internal/employee/errors"
	"go-hris/internal/rbac"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=auth_service.go -destination=mock/auth_service_mock.go -package=mock
type Service interface {
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, resp AuthResponse, err error)

	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, resp AuthResponse, err error)

	GetMe(ctx context.Context, userID string) (*AuthResponse, error)

	Register(ctx context.Context, req RegisterRequest) (AuthResponse, error)

	RegisterCompany(ctx context.Context, req RegisterCompanyRequest) (AuthResponse, error)
}

type service struct {
	repo         Repository
	rbac         rbac.Service
	employeeRepo employee.Repository
	companyRepo  company.Repository
}

func NewService(repo Repository, rbac rbac.Service, employeeRepo employee.Repository, companyRepo company.Repository) Service {
	return &service{repo: repo, rbac: rbac, employeeRepo: employeeRepo, companyRepo: companyRepo}
}

func (s *service) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, resp AuthResponse, err error) {
	// 1. Ambil user
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", AuthResponse{}, autherrors.ErrInvalidCredentials
	}

	// 2. Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", AuthResponse{}, autherrors.ErrInvalidCredentials
	}

	// 3. Load company policy untuk Casbin
	if err := s.rbac.LoadCompanyPolicy(user.CompanyID.String()); err != nil {
		return "", "", AuthResponse{}, err
	}

	// 4. Generate token (UserID + EmployeeID + CompanyID + Role)
	role := user.Role
	if role == "" {
		role = "EMPLOYEE"
	}
	accessToken, _ = s.generateToken(
		user.ID.String(),
		user.EmployeeID.String(),
		user.CompanyID.String(),
		role,
		time.Minute*15,
	)
	refreshToken, _ = s.generateToken(
		user.ID.String(),
		user.EmployeeID.String(),
		user.CompanyID.String(),
		role,
		time.Hour*24*7,
	)

	// 5. Get permissions
	perms, _ := s.rbac.GetEmployeePermissions(user.EmployeeID.String(), user.CompanyID.String())

	return accessToken, refreshToken, AuthResponse{
		ID:          user.ID.String(),
		CompanyID:   user.CompanyID.String(),
		EmployeeID:  user.EmployeeID.String(),
		Email:       user.Email,
		Name:        user.Name,
		Role:        role,
		Permissions: perms,
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (string, string, AuthResponse, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, autherrors.ErrInvalidToken
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return "", "", AuthResponse{}, autherrors.ErrInvalidRefreshToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", AuthResponse{}, autherrors.ErrInvalidToken
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return "", "", AuthResponse{}, autherrors.ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", "", AuthResponse{}, autherrors.ErrInvalidUserID
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return "", "", AuthResponse{}, autherrors.ErrUserNotFound
	}

	role := user.Role
	if role == "" {
		role = "EMPLOYEE"
	}

	newAccessToken, err := s.generateToken(
		user.ID.String(),
		user.EmployeeID.String(),
		user.CompanyID.String(),
		role,
		time.Minute*15,
	)
	if err != nil {
		return "", "", AuthResponse{}, autherrors.ErrTokenGenerationFailed
	}

	newRefreshToken, err := s.generateToken(
		user.ID.String(),
		user.EmployeeID.String(),
		user.CompanyID.String(),
		role,
		time.Hour*24*7,
	)
	if err != nil {
		return "", "", AuthResponse{}, autherrors.ErrTokenGenerationFailed
	}

	// Get permissions
	perms, _ := s.rbac.GetEmployeePermissions(user.EmployeeID.String(), user.CompanyID.String())

	return newAccessToken, newRefreshToken, AuthResponse{
		ID:          user.ID.String(),
		CompanyID:   user.CompanyID.String(),
		EmployeeID:  user.EmployeeID.String(),
		Email:       user.Email,
		Name:        user.Name,
		Role:        role,
		Permissions: perms,
	}, nil
}

func (s *service) GetMe(ctx context.Context, userID string) (*AuthResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, autherrors.ErrInvalidUserID
	}

	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, autherrors.ErrUserNotFound
	}

	// Get permissions
	perms, _ := s.rbac.GetEmployeePermissions(u.EmployeeID.String(), u.CompanyID.String())

	return &AuthResponse{
		ID:          u.ID.String(),
		Email:       u.Email,
		Name:        u.Name,
		Role:        u.Role,
		EmployeeID:  u.EmployeeID.String(),
		CompanyID:   u.CompanyID.String(),
		Permissions: perms,
	}, nil
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	// 1. Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResponse{}, err
	}

	// 2. Validasi EmployeeID
	eID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return AuthResponse{}, employeeerrors.ErrInvalidEmployeeID
	}

	// 3. Pastikan employee exist & ambil data CompanyID
	employee, err := s.employeeRepo.FindByIDAndCompany(ctx, req.CompanyID, eID.String())
	if err != nil {
		return AuthResponse{}, employeeerrors.ErrEmployeeNotFound
	}

	// 4. Buat objek User menggunakan data dari Employee yang ditemukan
	var roleName string
	var finalRoleID string

	if req.RoleID != "" {
		// Lookup role by ID to get Name for denormalization
		r, err := s.rbac.GetRole(req.RoleID)
		if err != nil {
			return AuthResponse{}, err
		}
		roleName = r.Name
		finalRoleID = req.RoleID
	} else {
		// Default to "Employee" role
		r, err := s.rbac.GetRoleByName(employee.CompanyID.String(), "Employee")
		if err != nil {
			return AuthResponse{}, err
		}
		roleName = r.Name
		finalRoleID = r.ID
	}

	user := &User{
		ID:         uuid.New(),
		EmployeeID: &eID,
		CompanyID:  employee.CompanyID,
		Email:      req.Email,
		Name:       req.Name,
		Password:   string(hashed),
		Role:       roleName,
		IsActive:   true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return AuthResponse{}, autherrors.ErrEmailAlreadyRegistered
	}

	// 5. Assign role in RBAC (employee_roles table)
	if err := s.rbac.AssignRoleIDToEmployee(eID.String(), finalRoleID); err != nil {
		return AuthResponse{}, err
	}

	// 6. Load policy untuk company agar bisa enforce
	if err := s.rbac.LoadCompanyPolicy(employee.CompanyID.String()); err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		ID:        user.ID.String(),
		CompanyID: user.CompanyID.String(),
		Email:     user.Email,
		Name:      user.Name,
		Role:      roleName,
	}, nil
}

// reusable token generator
func (s *service) generateToken(userID, employeeID, companyID, role string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id":     userID,
		"employee_id": employeeID,
		"company_id":  companyID,
		"role":        role,
		"exp":         time.Now().Add(expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// auth/auth.service.go

func (s *service) RegisterCompany(ctx context.Context, req RegisterCompanyRequest) (AuthResponse, error) {
	// 1. Check if admin email already registered
	_, err := s.repo.GetByEmail(ctx, req.AdminEmail)
	if err == nil {
		return AuthResponse{}, autherrors.ErrEmailAlreadyRegistered
	}

	// 2. Start Transaction
	tx := s.repo.(*repository).db.Begin()
	if tx.Error != nil {
		return AuthResponse{}, tx.Error
	}
	defer tx.Rollback()

	// 3. Create Company
	comp := &company.Company{
		Name:     req.CompanyName,
		Email:    req.CompanyEmail,
		IsActive: true,
	}

	// Gunakan WithTx untuk repository terkait
	txRepo := s.companyRepo.WithTx(tx)
	if err := txRepo.Create(ctx, comp); err != nil {
		return AuthResponse{}, err
	}

	// 4. Seed Default Roles
	// Pastikan s.rbac.WithTx(tx) mengembalikan service RBAC yang menggunakan transaksi tersebut
	rbacTx := s.rbac.WithTx(tx)
	if err := rbacTx.SeedDefaultRoles(comp.ID.String()); err != nil {
		return AuthResponse{}, err
	}

	// 5. Create Employee
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	adminEmployee := &employee.Employee{
		ID:             uuid.New(),
		CompanyID:      comp.ID,
		FullName:       req.AdminName,
		Email:          req.AdminEmail,
		EmployeeNumber: fmt.Sprintf("ADM-%06d", 1),
		HireDate:       time.Now(),
	}

	if err := tx.Create(adminEmployee).Error; err != nil {
		return AuthResponse{}, err
	}

	// 6. Create User
	user := &User{
		ID:         uuid.New(),
		CompanyID:  comp.ID,
		EmployeeID: &adminEmployee.ID,
		Name:       req.AdminName,
		Email:      req.AdminEmail,
		Password:   string(hashedPassword),
		Role:       "SUPERADMIN",
		IsActive:   true,
	}

	if err := tx.Create(user).Error; err != nil {
		return AuthResponse{}, err
	}

	// 7. Assign SUPERADMIN role
	// PENTING: Gunakan rbacTx (bukan s.rbac) agar bisa membaca data yang belum di-commit
	saRole, err := rbacTx.GetRoleByName(comp.ID.String(), "SUPERADMIN")
	if err != nil {
		return AuthResponse{}, err // Error "record not found" terjadi di sini sebelumnya
	}

	if err := rbacTx.AssignRoleIDToEmployee(adminEmployee.ID.String(), saRole.ID); err != nil {
		return AuthResponse{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		ID:        user.ID.String(),
		CompanyID: user.CompanyID.String(),
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
	}, nil
}
