package rbac

import "github.com/casbin/casbin/v2"
import "sync"
import "log"

//go:generate mockgen -source=rbac_service.go -destination=mock/rbac_service_mock.go -package=mock
type Service interface {
	LoadCompanyPolicy(companyID string) error
	Enforce(req EnforceRequest) (bool, error)
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

func (s *service) Enforce(req EnforceRequest) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.loadCompanyPolicyUnlocked(req.CompanyID); err != nil {
		return false, err
	}

	roles := s.enforcer.GetRolesForUserInDomain(req.EmployeeID, req.CompanyID)

	perms, permErr := s.enforcer.GetImplicitPermissionsForUser(req.EmployeeID, req.CompanyID)
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
		log.Printf("rbac enforce result: employee_id=%s company_id=%s resource=%s action=%s err=%v", req.EmployeeID, req.CompanyID, req.Resource, req.Action, err)
		return false, err
	}

	log.Printf("rbac enforce result: employee_id=%s company_id=%s resource=%s action=%s allowed=%t roles=%v permissions=%v",
		req.EmployeeID, req.CompanyID, req.Resource, req.Action, allowed, roles, perms)

	return allowed, nil
}
