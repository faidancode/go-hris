package rbac_test

import (
	"errors"
	"go-hris/internal/domain"
	"go-hris/internal/rbac"
	rbacMock "go-hris/internal/rbac/mock"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

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
	if err != nil {
		t.Fatalf("failed to create model: %v", err)
	}

	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("failed to create enforcer: %v", err)
	}

	return e
}

func TestRBACService_Management(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := rbacMock.NewMockRepository(ctrl)
	enforcer := newTestEnforcer(t)
	service := rbac.NewService(mockRepo, enforcer)

	companyID := "comp-1"
	roleID := "role-1"

	t.Run("ListRoles - Success", func(t *testing.T) {
		mockRepo.EXPECT().ListRoles(companyID).Return([]rbac.RoleRow{
			{ID: roleID, Name: "Admin", Description: "Admin role"},
		}, nil)
		mockRepo.EXPECT().GetPermissionsByRoleID(roleID).Return([]rbac.PermissionRow{
			{ID: "p1", Resource: "user", Action: "read"},
		}, nil)

		res, err := service.ListRoles(companyID)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "Admin", res[0].Name)
		assert.Contains(t, res[0].Permissions, "user:read")
	})

	t.Run("ListRoles - Error", func(t *testing.T) {
		mockRepo.EXPECT().ListRoles(companyID).Return(nil, errors.New("db error"))
		_, err := service.ListRoles(companyID)
		assert.Error(t, err)
	})

	t.Run("GetRole - Success", func(t *testing.T) {
		mockRepo.EXPECT().GetRoleByID(roleID).Return(&rbac.RoleRow{
			ID: roleID, Name: "Admin",
		}, nil)
		mockRepo.EXPECT().GetPermissionsByRoleID(roleID).Return([]rbac.PermissionRow{
			{Resource: "user", Action: "read"},
		}, nil)

		res, err := service.GetRole(roleID)
		assert.NoError(t, err)
		assert.Equal(t, "Admin", res.Name)
		assert.Contains(t, res.Permissions, "user:read")
	})

	t.Run("GetRole - Not Found", func(t *testing.T) {
		mockRepo.EXPECT().GetRoleByID("wrong").Return(nil, errors.New("not found"))
		_, err := service.GetRole("wrong")
		assert.Error(t, err)
	})

	t.Run("CreateRole - Success", func(t *testing.T) {
		req := domain.CreateRoleRequest{
			Name:        "New Role",
			Permissions: []string{"user:read"},
		}
		mockRepo.EXPECT().CreateRole(gomock.Any()).Return(nil)
		mockRepo.EXPECT().ListPermissions().Return([]rbac.PermissionRow{
			{ID: "p1", Resource: "user", Action: "read"},
		}, nil)
		mockRepo.EXPECT().UpdateRolePermissions(gomock.Any(), []string{"p1"}).Return(nil)

		err := service.CreateRole(companyID, req)
		assert.NoError(t, err)
	})

	t.Run("UpdateRole - Success", func(t *testing.T) {
		req := domain.UpdateRoleRequest{
			Name:        "Updated Name",
			Permissions: []string{"user:write"},
		}
		mockRepo.EXPECT().GetRoleByID(roleID).Return(&rbac.RoleRow{ID: roleID}, nil)
		mockRepo.EXPECT().UpdateRole(gomock.Any()).Return(nil)
		mockRepo.EXPECT().ListPermissions().Return([]rbac.PermissionRow{
			{ID: "p2", Resource: "user", Action: "write"},
		}, nil)
		mockRepo.EXPECT().UpdateRolePermissions(roleID, []string{"p2"}).Return(nil)

		err := service.UpdateRole(roleID, req)
		assert.NoError(t, err)
	})

	t.Run("DeleteRole - Success", func(t *testing.T) {
		mockRepo.EXPECT().DeleteRole(roleID).Return(nil)
		err := service.DeleteRole(roleID)
		assert.NoError(t, err)
	})

	t.Run("ListPermissions - Success", func(t *testing.T) {
		mockRepo.EXPECT().ListPermissions().Return([]rbac.PermissionRow{
			{ID: "p1", Resource: "user", Action: "read", Label: "Read User", Category: "User Management"},
		}, nil)

		res, err := service.ListPermissions()
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "Read User", res[0].Label)
	})
}

func TestRBACService_Enforce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := rbacMock.NewMockRepository(ctrl)
	enforcer := newTestEnforcer(t)
	service := rbac.NewService(mockRepo, enforcer)

	companyID := "comp-1"
	empID := "emp-1"

	t.Run("Allowed", func(t *testing.T) {
		mockRepo.EXPECT().GetEmployeeRoles(companyID).Return([]rbac.EmployeeRoleRow{
			{EmployeeID: empID, RoleID: "admin"},
		}, nil)
		mockRepo.EXPECT().GetRolePermissions(companyID).Return([]rbac.RolePermissionRow{
			{RoleID: "admin", Resource: "user", Action: "read"},
		}, nil)

		allowed, err := service.Enforce(domain.EnforceRequest{
			EmployeeID: empID,
			CompanyID:  companyID,
			Resource:   "user",
			Action:     "read",
		})
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("Denied", func(t *testing.T) {
		// MockRepo might be called again because Enforce reloads policy every time
		mockRepo.EXPECT().GetEmployeeRoles(companyID).Return(nil, nil)
		mockRepo.EXPECT().GetRolePermissions(companyID).Return(nil, nil)

		allowed, err := service.Enforce(domain.EnforceRequest{
			EmployeeID: empID,
			CompanyID:  companyID,
			Resource:   "secret",
			Action:     "write",
		})
		assert.NoError(t, err)
		assert.False(t, allowed)
	})
}
