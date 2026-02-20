package rbac

import (
	"go-hris/internal/domain"
	"log"
	"sync"

	"github.com/casbin/casbin/v2"
)

//go:generate mockgen -source=rbac_service.go -destination=mock/rbac_service_mock.go -package=mock
type Service interface {
	LoadCompanyPolicy(companyID string) error
	Enforce(req domain.EnforceRequest) (bool, error)
	GetEmployeePermissions(employeeID, companyID string) ([]string, error)

	// Management
	ListRoles(companyID string) ([]domain.RoleResponse, error)
	GetRole(id string) (*domain.RoleResponse, error)
	CreateRole(companyID string, req domain.CreateRoleRequest) error
	UpdateRole(id string, req domain.UpdateRoleRequest) error
	DeleteRole(id string) error

	ListPermissions() ([]domain.PermissionResponse, error)
}

type service struct {
	repo     Repository
	enforcer *casbin.Enforcer
	mu       sync.Mutex
}

func NewService(repo Repository, enforcer *casbin.Enforcer) Service {
	return &service{
		repo:     repo,
		enforcer: enforcer,
	}
}

func (s *service) LoadCompanyPolicy(companyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.loadCompanyPolicyUnlocked(companyID)
}

func (s *service) loadCompanyPolicyUnlocked(companyID string) error {
	s.enforcer.ClearPolicy()

	// Load grouping policy
	employeeRoles, err := s.repo.GetEmployeeRoles(companyID)
	if err != nil {
		return err
	}
	log.Printf("rbac load policy: company_id=%s employee_roles=%d", companyID, len(employeeRoles))

	for _, er := range employeeRoles {
		_, err := s.enforcer.AddGroupingPolicy(
			er.EmployeeID,
			er.RoleID,
			companyID,
		)
		if err != nil {
			return err
		}
	}

	// Load permission policy
	rolePerms, err := s.repo.GetRolePermissions(companyID)
	if err != nil {
		return err
	}
	log.Printf("rbac load policy: company_id=%s role_permissions=%d", companyID, len(rolePerms))

	for _, rp := range rolePerms {
		_, err := s.enforcer.AddPolicy(
			rp.RoleID,
			companyID,
			rp.Resource,
			rp.Action,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *service) Enforce(req domain.EnforceRequest) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.loadCompanyPolicyUnlocked(req.CompanyID); err != nil {
		return false, err
	}

	_, permErr := s.enforcer.GetImplicitPermissionsForUser(req.EmployeeID, req.CompanyID)
	if permErr != nil {
		log.Printf("rbac enforce debug: failed_get_permissions employee_id=%s company_id=%s err=%v", req.EmployeeID, req.CompanyID, permErr)
	}

	allowed, err := s.enforcer.Enforce(
		req.EmployeeID,
		req.CompanyID,
		req.Resource,
		req.Action,
	)
	if err != nil {
		return false, err
	}

	return allowed, nil
}
func (s *service) GetEmployeePermissions(employeeID, companyID string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.loadCompanyPolicyUnlocked(companyID); err != nil {
		return nil, err
	}

	perms, err := s.enforcer.GetImplicitPermissionsForUser(employeeID, companyID)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, p := range perms {
		// p format: [sub, dom, res, act]
		// we want res:act
		if len(p) >= 4 {
			result = append(result, p[2]+":"+p[3])
		}
	}

	return result, nil
}

func (s *service) ListRoles(companyID string) ([]domain.RoleResponse, error) {
	roles, err := s.repo.ListRoles(companyID)
	if err != nil {
		return nil, err
	}

	var result []domain.RoleResponse
	for _, r := range roles {
		perms, _ := s.repo.GetPermissionsByRoleID(r.ID)
		pList := []string{}
		for _, p := range perms {
			pList = append(pList, p.Resource+":"+p.Action)
		}

		result = append(result, domain.RoleResponse{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Permissions: pList,
		})
	}
	return result, nil
}

func (s *service) GetRole(id string) (*domain.RoleResponse, error) {
	r, err := s.repo.GetRoleByID(id)
	if err != nil {
		return nil, err
	}

	perms, _ := s.repo.GetPermissionsByRoleID(r.ID)
	pList := []string{}
	for _, p := range perms {
		pList = append(pList, p.Resource+":"+p.Action)
	}

	return &domain.RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Permissions: pList,
	}, nil
}

func (s *service) CreateRole(companyID string, req domain.CreateRoleRequest) error {
	role := &RoleRow{
		CompanyID:   companyID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.CreateRole(role); err != nil {
		return err
	}

	if len(req.Permissions) > 0 {
		allPerms, _ := s.repo.ListPermissions()
		var pIDs []string
		for _, rawP := range req.Permissions {
			for _, p := range allPerms {
				if p.Resource+":"+p.Action == rawP {
					pIDs = append(pIDs, p.ID)
				}
			}
		}
		if err := s.repo.UpdateRolePermissions(role.ID, pIDs); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) UpdateRole(id string, req domain.UpdateRoleRequest) error {
	role, err := s.repo.GetRoleByID(id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	if err := s.repo.UpdateRole(role); err != nil {
		return err
	}

	if req.Permissions != nil {
		allPerms, _ := s.repo.ListPermissions()
		var pIDs []string
		for _, rawP := range req.Permissions {
			for _, p := range allPerms {
				if p.Resource+":"+p.Action == rawP {
					pIDs = append(pIDs, p.ID)
				}
			}
		}
		if err := s.repo.UpdateRolePermissions(id, pIDs); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) DeleteRole(id string) error {
	return s.repo.DeleteRole(id)
}

func (s *service) ListPermissions() ([]domain.PermissionResponse, error) {
	perms, err := s.repo.ListPermissions()
	if err != nil {
		return nil, err
	}

	var result []domain.PermissionResponse
	for _, p := range perms {
		result = append(result, domain.PermissionResponse{
			ID:       p.ID,
			Resource: p.Resource,
			Action:   p.Action,
			Label:    p.Label,
			Category: p.Category,
		})
	}
	return result, nil
}
