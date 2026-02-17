package app

import (
	"database/sql"
	"go-hris/internal/auth"
	"go-hris/internal/department"
	"go-hris/internal/employee"
	"go-hris/internal/employeesalary"
	"go-hris/internal/payroll"
	"go-hris/internal/position"
	"go-hris/internal/rbac"
	"go-hris/internal/rbac/infra"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func registerModules(router *gin.Engine, db *sql.DB, gormDB *gorm.DB, rdb *redis.Client) error {
	// --- Repositories ---
	rbacRepo := rbac.NewRepository(gormDB)
	authRepo := auth.NewRepository(gormDB)
	departmentRepo := department.NewRepository(gormDB)
	employeeRepo := employee.NewRepository(gormDB)
	employeeSalaryRepo := employeesalary.NewRepository(gormDB)
	payrollRepo := payroll.NewRepository(gormDB)
	positionRepo := position.NewRepository(gormDB)

	// --- RBAC Core ---
	enforcer, err := infra.NewEnforcer(filepath.Join("internal", "rbac", "infra", "model.conf"))
	if err != nil {
		return err
	}
	rbacService := rbac.NewService(rbacRepo, enforcer)

	// --- Services ---
	authService := auth.NewService(authRepo, rbacService, employeeRepo)
	departmentService := department.NewService(db, departmentRepo)
	employeeService := employee.NewService(db, employeeRepo)
	employeeSalaryService := employeesalary.NewService(db, employeeSalaryRepo)
	payrollService := payroll.NewService(db, payrollRepo)
	positionService := position.NewService(db, positionRepo)

	// --- Handlers ---
	authHandler := auth.NewHandler(authService)
	departmentHandler := department.NewHandler(departmentService)
	employeeHandler := employee.NewHandler(employeeService)
	employeeSalaryHandler := employeesalary.NewHandler(employeeSalaryService)
	payrollHandler := payroll.NewHandlerWithRedis(payrollService, rdb)
	positionHandler := position.NewHandler(positionService)
	rbacHandler := rbac.NewHandler(rbacService)

	// --- Routes Registration ---
	api := router.Group("/api/v1")
	{
		auth.RegisterRoutes(api, authHandler)
		department.RegisterRoutes(api, departmentHandler, rbacService)
		employee.RegisterRoutes(api, employeeHandler, rbacService)
		employeesalary.RegisterRoutes(api, employeeSalaryHandler, rbacService)
		payroll.RegisterRoutes(api, payrollHandler, rbacService, rdb)
		position.RegisterRoutes(api, positionHandler, rbacService)
	}

	rbac.RegisterRoutes(router, rbacHandler)

	return nil
}
