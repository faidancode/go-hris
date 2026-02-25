package user

import (
	"context"
	"errors"
	"go-hris/internal/shared/contextutil"
	usererrors "go-hris/internal/user/errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=user_service.go -destination=mock/user_service_mock.go -package=mock

type Service interface {
	GetAll(ctx context.Context, companyID string) ([]UserResponse, error)
	GetAllWithRoles(ctx context.Context, companyID string) ([]UserWithRolesResponse, error)
	GetByID(ctx context.Context, companyID, id string) (UserResponse, error)

	Create(ctx context.Context, companyID string, req CreateUserRequest) (UserResponse, error)
	GetCompanyUsers(ctx context.Context, companyID string) ([]UserResponse, error)
	AssignRole(ctx context.Context, companyID string, userID string, roleName string) error
	ToggleStatus(ctx context.Context, companyID string, id string, isActive bool) error

	ChangePassword(ctx context.Context, companyID, userID, currentPassword, newPassword string) error
	ResetPassword(ctx context.Context, companyID, userID, newPassword string) error
	ForceResetPassword(ctx context.Context, companyID, userID, newPassword string) error
}

type service struct {
	repo         Repository
	roleAssigner RoleAssigner
}

type RoleAssigner interface {
	AssignRoleToEmployee(companyID, employeeID, roleName string) error
}

func NewService(repo Repository, roleAssigner ...RoleAssigner) Service {
	var assigner RoleAssigner
	if len(roleAssigner) > 0 {
		assigner = roleAssigner[0]
	}
	return &service{
		repo:         repo,
		roleAssigner: assigner,
	}
}

func (s *service) GetAll(ctx context.Context, companyID string) ([]UserResponse, error) {
	users, err := s.repo.FindAllByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	resp := make([]UserResponse, len(users))
	for i, u := range users {
		resp[i] = mapToResponse(u)
	}

	return resp, nil
}

func (s *service) GetByID(ctx context.Context, companyID, id string) (UserResponse, error) {
	u, err := s.repo.FindByID(ctx, companyID, id)
	if err != nil {
		return UserResponse{}, err
	}

	return UserResponse{
		ID:         u.ID.String(),
		Email:      u.Email,
		EmployeeID: u.EmployeeID.String(),
		IsActive:   u.IsActive,
		CreatedAt:  u.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *service) GetAllWithRoles(ctx context.Context, companyID string) ([]UserWithRolesResponse, error) {
	users, err := s.repo.FindAllByCompanyWithRoles(ctx, companyID)
	if err != nil {
		return nil, err
	}

	resp := make([]UserWithRolesResponse, 0, len(users))
	for _, u := range users {
		roles := []string{}
		if strings.TrimSpace(u.RolesRaw) != "" {
			roles = strings.Split(u.RolesRaw, ",")
		}

		resp = append(resp, UserWithRolesResponse{
			ID:             u.ID,
			EmployeeID:     u.EmployeeID,
			EmployeeNumber: u.EmployeeNumber,
			Email:          u.Email,
			FullName:       u.FullName,
			IsActive:       u.IsActive,
			Roles:          roles,
			CreatedAt:      u.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return resp, nil
}

func (s *service) Create(ctx context.Context, companyID string, req CreateUserRequest) (UserResponse, error) {
	l := contextutil.GetLogger(ctx, nil)

	l.Info("creating user",
		zap.String("employee_id", req.EmployeeID),
		zap.String("email", req.Email),
	)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		l.Error("failed to hash password", zap.Error(err))
		return UserResponse{}, err
	}

	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return UserResponse{}, usererrors.ErrInvalidCompanyID
	}

	u := &User{
		CompanyID:  companyUUID,
		EmployeeID: uuid.MustParse(req.EmployeeID),
		Name:       req.Email, // fallback; profile name mengikuti employee record pada layer lain
		Role:       "EMPLOYEE",
		Email:      req.Email,
		Password:   string(hashedPassword),
		IsActive:   true,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		l.Error("failed to create user", zap.Error(err))
		return UserResponse{}, err
	}

	l.Info("user created successfully", zap.String("email", u.Email))
	return mapToResponse(*u), nil
}

func (s *service) GetCompanyUsers(ctx context.Context, companyID string) ([]UserResponse, error) {
	l := contextutil.GetLogger(ctx, nil)

	users, err := s.repo.FindAllByCompany(ctx, companyID)
	if err != nil {
		l.Error("failed to get company users", zap.Error(err))
		return nil, err
	}

	res := make([]UserResponse, len(users))
	for i, u := range users {
		res[i] = mapToResponse(u)
	}

	return res, nil
}

func (s *service) AssignRole(ctx context.Context, companyID string, userID string, roleName string) error {
	if s.roleAssigner == nil {
		return errors.New("role assigner is not configured")
	}

	u, err := s.repo.FindByID(ctx, companyID, userID)
	if err != nil {
		return err
	}

	return s.roleAssigner.AssignRoleToEmployee(companyID, u.EmployeeID.String(), strings.TrimSpace(roleName))
}

func (s *service) ToggleStatus(ctx context.Context, companyID string, id string, isActive bool) error {
	l := contextutil.GetLogger(ctx, nil)

	u, err := s.repo.FindByID(ctx, companyID, id)
	if err != nil {
		l.Error("failed to find user", zap.Error(err))
		return err
	}

	u.IsActive = isActive

	if err := s.repo.Update(ctx, u); err != nil {
		l.Error("failed to update user status", zap.Error(err))
		return err
	}

	return nil
}

func (s *service) ChangePassword(ctx context.Context, companyID, userID, currentPassword, newPassword string) error {
	l := contextutil.GetLogger(ctx, nil)

	u, err := s.repo.FindByID(ctx, companyID, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(currentPassword)); err != nil {
		return errors.New("invalid current password")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		l.Error("failed to hash new password", zap.Error(err))
		return err
	}

	u.Password = string(hashed)
	return s.repo.Update(ctx, u)
}

func (s *service) ResetPassword(ctx context.Context, companyID, userID, newPassword string) error {
	u, err := s.repo.FindByID(ctx, companyID, userID)
	if err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashed)
	return s.repo.Update(ctx, u)
}

func (s *service) ForceResetPassword(ctx context.Context, companyID, userID, newPassword string) error {
	// Same as reset for now (no extra flag in schema)
	return s.ResetPassword(ctx, companyID, userID, newPassword)
}

func mapToResponse(u User) UserResponse {
	resp := UserResponse{
		ID:         u.ID.String(),
		EmployeeID: u.EmployeeID.String(),
		Email:      u.Email,
		IsActive:   u.IsActive,
		CreatedAt:  u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if u.Employee != nil {
		resp.FullName = u.Employee.FullName
	}
	return resp
}
