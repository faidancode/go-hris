package rbac

import "github.com/casbin/casbin/v2"

//go:generate mockgen -source=rbac_service.go -destination=mock/rbac_service_mock.go -package=mock
type Service interface {
	LoadCompanyPolicy(companyID string) error
	Enforce(req EnforceRequest) (bool, error)
}

type service struct {
	repo     Repository
	enforcer *casbin.Enforcer
}

func NewService(repo Repository, enforcer *casbin.Enforcer) Service {
	return &service{
		repo:     repo,
		enforcer: enforcer,
	}
}

func (s *service) LoadCompanyPolicy(companyID string) error {
	s.enforcer.ClearPolicy()

	// Load grouping policy
	employeeRoles, err := s.repo.GetEmployeeRoles(companyID)
	if err != nil {
		return err
	}

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
	return s.enforcer.Enforce(
		req.EmployeeID,
		req.CompanyID,
		req.Resource,
		req.Action,
	)
}
