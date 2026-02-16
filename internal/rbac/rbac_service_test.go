package rbac

import (
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/stretchr/testify/assert"
)

// =========================================
// Mock Repository
// =========================================

type mockRepo struct{}

func (m *mockRepo) GetEmployeeRoles(companyID string) ([]EmployeeRoleRow, error) {
	return []EmployeeRoleRow{
		{
			EmployeeID: "emp-1",
			RoleID:     "role-owner",
		},
	}, nil
}

func (m *mockRepo) GetRolePermissions(companyID string) ([]RolePermissionRow, error) {
	return []RolePermissionRow{
		{
			RoleID:   "role-owner",
			Resource: "employee",
			Action:   "read",
		},
	}, nil
}

// =========================================
// Helper: Test Enforcer
// =========================================

func newTestEnforcer(t *testing.T) *casbin.Enforcer {
	modelText := `[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	assert.NoError(t, err)

	e, err := casbin.NewEnforcer(m)
	assert.NoError(t, err)

	return e
}

// =========================================
// TEST: Load + Enforce
// =========================================

func TestRBACService_Enforce(t *testing.T) {
	repo := &mockRepo{}
	enforcer := newTestEnforcer(t)

	service := NewService(repo, enforcer)

	err := service.LoadCompanyPolicy("company-1")
	assert.NoError(t, err)

	// Should allow
	allowed, err := service.Enforce(EnforceRequest{
		EmployeeID: "emp-1",
		CompanyID:  "company-1",
		Resource:   "employee",
		Action:     "read",
	})

	assert.NoError(t, err)
	assert.True(t, allowed)

	// Should deny
	denied, err := service.Enforce(EnforceRequest{
		EmployeeID: "emp-1",
		CompanyID:  "company-1",
		Resource:   "salary",
		Action:     "delete",
	})

	assert.NoError(t, err)
	assert.False(t, denied)
}
