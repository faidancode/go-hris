package app

import (
	"database/sql"
	"go-hris/internal/attendance"
	"go-hris/internal/auth"
	"go-hris/internal/company"
	"go-hris/internal/department"
	"go-hris/internal/employee"
	"go-hris/internal/employeesalary"
	"go-hris/internal/leave"
	"go-hris/internal/messaging/kafka"
	"go-hris/internal/payroll"
	"go-hris/internal/position"
	"go-hris/internal/rbac"
	"go-hris/internal/rbac/infra"
	"go-hris/internal/rbac/rbac_http"
	"go-hris/internal/shared/counter"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func registerModules(
	router *gin.Engine,
	db *sql.DB,
	gormDB *gorm.DB,
	rdb *redis.Client,
	logger *zap.Logger,
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
	counterRepo := counter.NewRepository(gormDB)
	companyRepo := company.NewRepository(gormDB)

	// --- RBAC Core ---
	enforcer, err := infra.NewEnforcer(filepath.Join("internal", "rbac", "infra", "model.conf"))
	if err != nil {
		return err
	}
	rbacService := rbac.NewService(rbacRepo, enforcer)

	// --- Services ---
	companyService := company.NewService(companyRepo)
	authService := auth.NewService(authRepo, rbacService, employeeRepo, companyRepo)
	attendanceService := attendance.NewService(db, attendanceRepo)
	departmentService := department.NewService(db, departmentRepo, rdb)
	employeeSalaryService := employeesalary.NewService(db, employeeSalaryRepo)
	employeeService := employee.NewServiceWithOutbox(db, employeeRepo, counterRepo, outboxRepo, rdb)
	leaveService := leave.NewService(db, leaveRepo)
	payrollService := payroll.NewServiceWithOutbox(db, payrollRepo, outboxRepo)
	positionService := position.NewService(db, positionRepo, rdb)

	// --- Handlers ---
	companyHandler := company.NewHandler(companyService)
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
		company.RegisterRoutes(api, companyHandler, rbacService)
		auth.RegisterRoutes(api, authHandler)
		attendance.RegisterRoutes(api, attendanceHandler, rbacService)
		department.RegisterRoutes(api, departmentHandler, rbacService)
		employee.RegisterRoutes(api, employeeHandler, rbacService, logger)
		employeesalary.RegisterRoutes(api, employeeSalaryHandler, rbacService)
		leave.RegisterRoutes(api, leaveHandler, rbacService)
		payroll.RegisterRoutes(api, payrollHandler, rbacService, rdb)
		position.RegisterRoutes(api, positionHandler, rbacService)
		rbac_http.RegisterRoutes(api, rbacHandler, rbacService)
	}

	return nil
}
