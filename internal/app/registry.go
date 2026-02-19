package app

import (
	"database/sql"
	"go-hris/internal/attendance"
	"go-hris/internal/auth"
	"go-hris/internal/department"
	"go-hris/internal/employee"
	"go-hris/internal/employeesalary"
	"go-hris/internal/leave"
	"go-hris/internal/messaging/kafka"
	"go-hris/internal/payroll"
	"go-hris/internal/position"
	"go-hris/internal/rbac"
	"go-hris/internal/rbac/infra"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func registerModules(
	router *gin.Engine,
	db *sql.DB,
	gormDB *gorm.DB,
	rdb *redis.Client,
) error {
	// --- Repositories ---
	rbacRepo := rbac.NewRepository(gormDB)
	attendanceRepo := attendance.NewRepository(gormDB)
	authRepo := auth.NewRepository(gormDB)
	departmentRepo := department.NewRepository(gormDB)
	employeeRepo := employee.NewRepository(gormDB)
	employeeSalaryRepo := employeesalary.NewRepository(gormDB)
	leaveRepo := leave.NewRepository(gormDB)
	outboxRepo := kafka.NewOutboxRepository(db)
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
	attendanceService := attendance.NewService(db, attendanceRepo)
	departmentService := department.NewService(db, departmentRepo)
	employeeSalaryService := employeesalary.NewService(db, employeeSalaryRepo)
	employeeService := employee.NewServiceWithOutbox(db, employeeRepo, outboxRepo)
	leaveService := leave.NewService(db, leaveRepo)
	payrollService := payroll.NewServiceWithOutbox(db, payrollRepo, outboxRepo)
	positionService := position.NewService(db, positionRepo)

	// --- Handlers ---
	authHandler := auth.NewHandler(authService)
	attendanceHandler := attendance.NewHandler(attendanceService)
	departmentHandler := department.NewHandler(departmentService)
	employeeHandler := employee.NewHandler(employeeService)
	employeeSalaryHandler := employeesalary.NewHandler(employeeSalaryService)
	leaveHandler := leave.NewHandler(leaveService)
	payrollHandler := payroll.NewHandlerWithRedis(payrollService, rdb)
	positionHandler := position.NewHandler(positionService)
	rbacHandler := rbac.NewHandler(rbacService)

	// --- Routes Registration ---
	api := router.Group("/api/v1")
	{
		auth.RegisterRoutes(api, authHandler)
		attendance.RegisterRoutes(api, attendanceHandler, rbacService)
		department.RegisterRoutes(api, departmentHandler, rbacService)
		employee.RegisterRoutes(api, employeeHandler, rbacService)
		employeesalary.RegisterRoutes(api, employeeSalaryHandler, rbacService)
		leave.RegisterRoutes(api, leaveHandler, rbacService)
		payroll.RegisterRoutes(api, payrollHandler, rbacService, rdb)
		position.RegisterRoutes(api, positionHandler, rbacService)
	}

	rbac.RegisterRoutes(router, rbacHandler)

	return nil
}
