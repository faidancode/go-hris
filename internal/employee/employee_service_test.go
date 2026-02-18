package employee_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"go-hris/internal/employee"
	"go-hris/internal/shared/apperror"

	employeeMock "go-hris/internal/employee/mock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type serviceDeps struct {
	db      *sql.DB
	sqlMock sqlmock.Sqlmock
	service employee.Service
	repo    *employeeMock.MockRepository
}

func setupServiceTest(t *testing.T) *serviceDeps {
	ctrl := gomock.NewController(t)

	db, sqlMock, _ := sqlmock.New()
	repo := employeeMock.NewMockRepository(ctrl)

	svc := employee.NewService(db, repo)

	return &serviceDeps{
		db:      db,
		sqlMock: sqlMock,
		service: svc,
		repo:    repo,
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

	t.Run("success", func(t *testing.T) {
		req := employee.CreateEmployeeRequest{FullName: "HR", Email: "hr@example.com", PositionID: uuid.New().String()}
		deptID := uuid.New()
		departmentID := uuid.New().String()

		expectTx(t, deps.sqlMock, true)

		deps.repo.EXPECT().
			WithTx(gomock.Any()).
			Return(deps.repo)

		deps.repo.EXPECT().
			GetDepartmentIDByPosition(ctx, companyID, req.PositionID).
			Return(departmentID, nil)

		deps.repo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, d *employee.Employee) error {
				assert.Equal(t, req.FullName, d.FullName)
				assert.Equal(t, companyID, d.CompanyID.String())
				assert.Equal(t, req.Email, d.Email)
				d.ID = deptID
				return nil
			})

		resp, err := deps.service.Create(ctx, companyID, req)

		assert.NoError(t, err)
		assert.Equal(t, deptID.String(), resp.ID)
		assert.Equal(t, req.FullName, resp.FullName)
	})

	t.Run("repo error -> rollback", func(t *testing.T) {
		req := employee.CreateEmployeeRequest{FullName: "HR", Email: "hr@example.com", PositionID: uuid.New().String()}
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
		req := employee.UpdateEmployeeRequest{FullName: "HR Updated", Email: "hr.updated@example.com", PositionID: uuid.New().String()}
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
		req := employee.UpdateEmployeeRequest{FullName: "HR Updated", Email: "hr.updated@example.com", PositionID: uuid.New().String()}
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
		req := employee.UpdateEmployeeRequest{FullName: "HR Updated", Email: "hr.updated@example.com", PositionID: uuid.New().String()}
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
