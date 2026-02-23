package employee_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"go-hris/internal/employee"
	employeeerrors "go-hris/internal/employee/errors"
	"go-hris/internal/shared/apperror"

	employeeMock "go-hris/internal/employee/mock"
	counterMock "go-hris/internal/shared/counter/mock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type serviceDeps struct {
	db        *sql.DB
	sqlMock   sqlmock.Sqlmock
	service   employee.Service
	repo      *employeeMock.MockRepository
	counter   *counterMock.MockRepository
	redismock redismock.ClientMock
}

func setupServiceTest(t *testing.T) *serviceDeps {
	ctrl := gomock.NewController(t)

	db, sqlMock, _ := sqlmock.New()
	dbRedis, redisMock := redismock.NewClientMock()
	repo := employeeMock.NewMockRepository(ctrl)
	counterRepo := counterMock.NewMockRepository(ctrl)

	svc := employee.NewService(db, repo, counterRepo, dbRedis)

	return &serviceDeps{
		db:        db,
		sqlMock:   sqlMock,
		service:   svc,
		repo:      repo,
		counter:   counterRepo,
		redismock: redisMock,
	}
}

func expectTx(t *testing.T, mock sqlmock.Sqlmock, commit bool) {
	t.Helper()
	mock.ExpectBegin()
	if commit {
		mock.ExpectCommit()
	} else {
		mock.ExpectRollback()
	}
}
func TestEmployeeService_Create(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()
	companyID := uuid.New().String()

	t.Run("success - auto generate employee number", func(t *testing.T) {
		req := employee.CreateEmployeeRequest{
			FullName:         "HR",
			Email:            "hr@example.com",
			EmployeeNumber:   "", // Empty for auto-generation
			Phone:            "0812",
			HireDate:         "2026-01-01",
			EmploymentStatus: "active",
			PositionID:       uuid.New().String(),
		}
		deptID := uuid.New()
		departmentID := uuid.New().String()

		expectTx(t, deps.sqlMock, true)

		deps.repo.EXPECT().
			WithTx(gomock.Any()).
			Return(deps.repo)

		deps.repo.EXPECT().
			GetDepartmentIDByPosition(ctx, companyID, req.PositionID).
			Return(departmentID, nil)

		deps.counter.EXPECT().
			GetNextValue(ctx, companyID, "employee_number").
			Return(int64(123), nil)

		deps.repo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, d *employee.Employee) error {
				assert.Equal(t, req.FullName, d.FullName)
				assert.Equal(t, "EMP-000123", d.EmployeeNumber)
				assert.Equal(t, companyID, d.CompanyID.String())
				assert.Equal(t, req.Email, d.Email)
				d.ID = deptID
				return nil
			})

		resp, err := deps.service.Create(ctx, companyID, req)

		assert.NoError(t, err)
		assert.Equal(t, deptID.String(), resp.ID)
		assert.Equal(t, "EMP-000123", resp.EmployeeNumber)
	})

	t.Run("repo error -> rollback", func(t *testing.T) {
		req := employee.CreateEmployeeRequest{FullName: "HR", Email: "hr@example.com", EmployeeNumber: "EMP-101", Phone: "0813", HireDate: "2026-01-02", EmploymentStatus: "active", PositionID: uuid.New().String()}
		departmentID := uuid.New().String()

		expectTx(t, deps.sqlMock, false) // rollback

		deps.repo.EXPECT().
			WithTx(gomock.Any()).
			Return(deps.repo)

		deps.repo.EXPECT().
			GetDepartmentIDByPosition(ctx, companyID, req.PositionID).
			Return(departmentID, nil)

		deps.repo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(errors.New("db error"))

		_, err := deps.service.Create(ctx, companyID, req)

		assert.Error(t, err)
	})

	t.Run("duplicate employee number -> conflict error", func(t *testing.T) {
		req := employee.CreateEmployeeRequest{FullName: "HR", Email: "hr@example.com", EmployeeNumber: "EMP-100", Phone: "0812", HireDate: "2026-01-01", EmploymentStatus: "active", PositionID: uuid.New().String()}
		departmentID := uuid.New().String()

		expectTx(t, deps.sqlMock, false)

		deps.repo.EXPECT().
			WithTx(gomock.Any()).
			Return(deps.repo)

		deps.repo.EXPECT().
			GetDepartmentIDByPosition(ctx, companyID, req.PositionID).
			Return(departmentID, nil)

		deps.repo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(&pgconn.PgError{Code: "23505", ConstraintName: "uq_employee_number"})

		_, err := deps.service.Create(ctx, companyID, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, employeeerrors.ErrEmployeeNumberAlreadyExists)
	})
}

func TestEmployeeService_GetAll(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()
	companyID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		mockEmployees := []employee.Employee{
			{ID: uuid.New(), FullName: "Andi", Email: "andi@comp.com"},
			{ID: uuid.New(), FullName: "Budi", Email: "budi@comp.com"},
		}

		deps.repo.EXPECT().
			FindAllByCompany(ctx, companyID).
			Return(mockEmployees, nil).
			Times(1)

		resp, err := deps.service.GetAll(ctx, companyID)

		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, "Andi", resp[0].FullName)
	})

	t.Run("error repository", func(t *testing.T) {
		deps.repo.EXPECT().
			FindAllByCompany(ctx, companyID).
			Return(nil, errors.New("db error"))

		resp, err := deps.service.GetAll(ctx, companyID)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestEmployeeService_GetOptions(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()
	companyID := uuid.New().String()
	// Menggunakan konstanta yang sesuai dengan implementasi service
	cacheKey := employee.EmployeeOptionsKeyPrefix + companyID

	t.Run("Hit Cache - Harus ambil data dari Redis", func(t *testing.T) {
		expectedResp := []employee.EmployeeResponse{
			{ID: uuid.New().String(), FullName: "Caca", EmployeeNumber: "EMP001"},
		}
		jsonResp, _ := json.Marshal(expectedResp)

		deps.redismock.ExpectGet(cacheKey).SetVal(string(jsonResp))

		resp, err := deps.service.GetOptions(ctx, companyID)

		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "Caca", resp[0].FullName)

		// Memastikan DB tidak disentuh
		deps.repo.EXPECT().FindOptionsByCompany(gomock.Any(), gomock.Any()).Times(0)
	})

	t.Run("Miss Cache - Harus ambil dari DB dan simpan ke Redis", func(t *testing.T) {
		// Gunakan ID unik khusus untuk test case ini
		companyID := uuid.New().String()
		cacheKey := employee.EmployeeOptionsKeyPrefix + companyID

		// 1. Mock Redis Get (Miss)
		deps.redismock.ExpectGet(cacheKey).RedisNil()

		// 2. Mock DB Data
		mockEmployees := []employee.Employee{
			{ID: uuid.New(), FullName: "Deni", EmployeeNumber: "EMP002"},
		}

		// 3. Mock Repo - Pastikan companyID MATCH
		deps.repo.EXPECT().
			FindOptionsByCompany(gomock.Any(), companyID).
			Return(mockEmployees, nil).
			Times(1)

		// 4. Mock Redis Set
		deps.redismock.ExpectSet(cacheKey, gomock.Any(), 1*time.Hour).SetVal("OK")

		// Execution
		resp, err := deps.service.GetOptions(ctx, companyID)

		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "Deni", resp[0].FullName)
	})

	t.Run("Database Error - Harus mengembalikan error", func(t *testing.T) {
		// Gunakan ID unik yang berbeda agar tidak bentrok di Singleflight
		companyID := uuid.New().String()
		cacheKey := employee.EmployeeOptionsKeyPrefix + companyID

		// 1. Mock Redis Get (Miss)
		deps.redismock.ExpectGet(cacheKey).RedisNil()

		// 2. Mock Repo Error
		deps.repo.EXPECT().
			FindOptionsByCompany(gomock.Any(), companyID).
			Return(nil, errors.New("database connection lost")).
			Times(1)

		// Execution
		resp, err := deps.service.GetOptions(ctx, companyID)

		// Verifikasi
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "database connection lost")
	})
}

func TestEmployeeService_GetByID(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()
	// Definisikan nilai konstan untuk satu siklus test case agar tidak tertukar
	companyID := uuid.New().String()
	targetID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		// 1. Pastikan return mock menggunakan targetID yang sama dengan ekspektasi assert
		expectedDept := &employee.Employee{
			ID:       uuid.MustParse(targetID),
			FullName: "HR",
		}

		deps.repo.EXPECT().
			FindByIDAndCompany(ctx, companyID, targetID).
			Return(expectedDept, nil).
			Times(1) // Tambahkan Times(1) untuk memastikan dipanggil tepat satu kali

		resp, err := deps.service.GetByID(ctx, companyID, targetID)

		// 2. Verifikasi error dan data
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, targetID, resp.ID, "ID yang dikembalikan harus sama dengan targetID")
	})

	t.Run("not found", func(t *testing.T) {
		// Contoh menggunakan apperror.ErrNotFound yang baru kita bahas
		deps.repo.EXPECT().
			FindByIDAndCompany(ctx, companyID, targetID).
			Return(nil, apperror.ErrNotFound)

		resp, err := deps.service.GetByID(ctx, companyID, targetID)

		assert.Error(t, err)
		assert.Empty(t, resp.ID)
		assert.True(t, errors.Is(err, apperror.ErrNotFound))
	})
}

func TestEmployeeService_Update(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()
	targetID := uuid.New()
	companyID := uuid.New()

	t.Run("success", func(t *testing.T) {
		req := employee.UpdateEmployeeRequest{FullName: "HR Updated", Email: "hr.updated@example.com", EmployeeNumber: "EMP-102", Phone: "0814", HireDate: "2026-01-03", EmploymentStatus: "active", PositionID: uuid.New().String()}
		departmentID := uuid.New().String()

		// Mock DB Transaction
		deps.sqlMock.ExpectBegin()

		// Mock Repository calls
		// Pastikan WithTx mengembalikan mock repo yang sama atau mock repo baru
		deps.repo.EXPECT().WithTx(gomock.Any()).Return(deps.repo)
		deps.repo.EXPECT().
			GetDepartmentIDByPosition(ctx, companyID.String(), req.PositionID).
			Return(departmentID, nil)

		// Mock FindByIDAndCompany (Harus ada karena dipanggil di service)
		existingDept := &employee.Employee{
			ID:        targetID,
			CompanyID: companyID,
			FullName:  "Old HR",
		}
		deps.repo.EXPECT().
			FindByIDAndCompany(ctx, companyID.String(), targetID.String()).
			Return(existingDept, nil)

		// Mock Update
		deps.repo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, d *employee.Employee) error {
				assert.Equal(t, req.FullName, d.FullName)
				assert.Equal(t, req.Email, d.Email)
				assert.Equal(t, targetID, d.ID)
				return nil
			})

		deps.sqlMock.ExpectCommit()

		resp, err := deps.service.Update(ctx, companyID.String(), targetID.String(), req)

		assert.NoError(t, err)
		assert.Equal(t, req.FullName, resp.FullName)
	})

	t.Run("error - employee not found", func(t *testing.T) {
		req := employee.UpdateEmployeeRequest{FullName: "HR Updated", Email: "hr.updated@example.com", EmployeeNumber: "EMP-103", Phone: "0815", HireDate: "2026-01-04", EmploymentStatus: "active", PositionID: uuid.New().String()}
		departmentID := uuid.New().String()

		deps.sqlMock.ExpectBegin()
		deps.repo.EXPECT().WithTx(gomock.Any()).Return(deps.repo)
		deps.repo.EXPECT().
			GetDepartmentIDByPosition(ctx, companyID.String(), req.PositionID).
			Return(departmentID, nil)

		// Simulasikan data tidak ditemukan
		deps.repo.EXPECT().
			FindByIDAndCompany(ctx, companyID.String(), targetID.String()).
			Return(nil, errors.New("employee not found"))

		deps.sqlMock.ExpectRollback()

		resp, err := deps.service.Update(ctx, companyID.String(), targetID.String(), req)

		assert.Error(t, err)
		assert.Empty(t, resp.ID)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("error - update failed", func(t *testing.T) {
		req := employee.UpdateEmployeeRequest{FullName: "HR Updated", Email: "hr.updated@example.com", EmployeeNumber: "EMP-104", Phone: "0816", HireDate: "2026-01-05", EmploymentStatus: "active", PositionID: uuid.New().String()}
		departmentID := uuid.New().String()

		deps.sqlMock.ExpectBegin()
		deps.repo.EXPECT().WithTx(gomock.Any()).Return(deps.repo)
		deps.repo.EXPECT().
			GetDepartmentIDByPosition(ctx, companyID.String(), req.PositionID).
			Return(departmentID, nil)

		existingDept := &employee.Employee{ID: targetID, CompanyID: companyID}
		deps.repo.EXPECT().
			FindByIDAndCompany(ctx, companyID.String(), targetID.String()).
			Return(existingDept, nil)

		// Simulasikan error saat eksekusi update
		deps.repo.EXPECT().
			Update(ctx, gomock.Any()).
			Return(errors.New("db connection error"))

		deps.sqlMock.ExpectRollback()

		_, err := deps.service.Update(ctx, companyID.String(), targetID.String(), req)

		assert.Error(t, err)
	})
}

func TestEmployeeService_Delete(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()
	companyID := uuid.New().String()
	targetID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		// Ekspektasi Transaksi
		expectTx(t, deps.sqlMock, true)

		// Mocking chain repository.WithTx(tx)
		deps.repo.EXPECT().WithTx(gomock.Any()).Return(deps.repo)

		deps.repo.EXPECT().
			Delete(ctx, companyID, targetID).
			Return(nil)

		err := deps.service.Delete(ctx, companyID, targetID)

		assert.NoError(t, err)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("failure - db error", func(t *testing.T) {
		expectTx(t, deps.sqlMock, false) // Rollback

		deps.repo.EXPECT().WithTx(gomock.Any()).Return(deps.repo)

		deps.repo.EXPECT().
			Delete(ctx, companyID, targetID).
			Return(errors.New("db error"))

		err := deps.service.Delete(ctx, companyID, targetID)

		assert.Error(t, err)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})
}
