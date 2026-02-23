package department_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"go-hris/internal/department"
	"go-hris/internal/shared/apperror"

	departmentMock "go-hris/internal/department/mock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type serviceDeps struct {
	db        *sql.DB
	sqlMock   sqlmock.Sqlmock
	service   department.Service
	repo      *departmentMock.MockRepository
	redismock redismock.ClientMock
}

func setupServiceTest(t *testing.T) *serviceDeps {
	ctrl := gomock.NewController(t)

	db, sqlMock, _ := sqlmock.New()
	dbRedis, redisMock := redismock.NewClientMock()
	repo := departmentMock.NewMockRepository(ctrl)

	svc := department.NewService(db, repo, dbRedis)

	return &serviceDeps{
		db:        db,
		sqlMock:   sqlMock,
		service:   svc,
		repo:      repo,
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

func TestDepartmentService_GetAll(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()
	companyID := "c56a4180-65aa-42ec-a945-5fd21dec0538"
	cacheKey := fmt.Sprintf("departments:all:%s", companyID)

	t.Run("Hit Cache - Harus ambil data dari Redis", func(t *testing.T) {
		// Data dummy untuk cache
		expectedResp := []department.DepartmentResponse{
			{ID: "pos-1", Name: "HR"},
			{ID: "pos-2", Name: "IT"},
		}
		jsonResp, _ := json.Marshal(expectedResp)

		// Mock Redis Get sukses
		deps.redismock.ExpectGet(cacheKey).SetVal(string(jsonResp))

		resp, err := deps.service.GetAll(ctx, companyID)

		// Verifikasi
		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, "HR", resp[0].Name)

		// Pastikan Repo TIDAK dipanggil jika cache hit
		deps.repo.EXPECT().FindAllByCompany(gomock.Any(), gomock.Any()).Times(0)
	})

	t.Run("Miss Cache - Harus ambil dari DB dan simpan ke Redis", func(t *testing.T) {
		// 1. Mock Redis Get return Nil (Cache Miss)
		deps.redismock.ExpectGet(cacheKey).RedisNil()

		// 2. Data dummy dari DB
		mockDepartments := []department.Department{
			{
				ID:   uuid.New(),
				Name: "Finance",
			},
		}

		// 3. Mock Repo dipanggil tepat satu kali
		deps.repo.EXPECT().
			FindAllByCompany(ctx, companyID).
			Return(mockDepartments, nil).
			Times(1)

		// 4. Mock Redis Set (karena service harus menyimpan hasil DB ke cache)
		// Kita gunakan gomock.Any() untuk argumen value JSON-nya
		deps.redismock.ExpectSet(cacheKey, gomock.Any(), 30*time.Minute).SetVal("OK")

		resp, err := deps.service.GetAll(ctx, companyID)

		// Verifikasi
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "Finance", resp[0].Name)
	})

	t.Run("Database Error - Harus mengembalikan error", func(t *testing.T) {
		deps.redismock.ExpectGet(cacheKey).RedisNil()

		deps.repo.EXPECT().
			FindAllByCompany(ctx, companyID).
			Return(nil, errors.New("db connection error")).
			Times(1)

		resp, err := deps.service.GetAll(ctx, companyID)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestDepartmentService_Create(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()
	companyID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		req := department.CreateDepartmentRequest{Name: "HR"}
		deptID := uuid.New()

		expectTx(t, deps.sqlMock, true)

		deps.repo.EXPECT().
			WithTx(gomock.Any()).
			Return(deps.repo)

		deps.repo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, d *department.Department) (department.Department, error) {
				assert.Equal(t, req.Name, d.Name)
				assert.Equal(t, companyID, d.CompanyID.String())
				d.ID = deptID
				return *d, nil
			})

		resp, err := deps.service.Create(ctx, companyID, req)

		assert.NoError(t, err)
		assert.Equal(t, deptID.String(), resp.ID)
		assert.Equal(t, req.Name, resp.Name)
	})

	t.Run("repo error -> rollback", func(t *testing.T) {
		req := department.CreateDepartmentRequest{Name: "HR"}

		expectTx(t, deps.sqlMock, false) // rollback

		deps.repo.EXPECT().
			WithTx(gomock.Any()).
			Return(deps.repo)

		deps.repo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(errors.New("db error"))

		_, err := deps.service.Create(ctx, companyID, req)

		assert.Error(t, err)
	})
}

func TestDepartmentService_GetByID(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()

	companyID := uuid.New().String()
	targetID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		// 1. Pastikan return mock menggunakan targetID yang sama dengan ekspektasi assert
		expectedDept := &department.Department{
			ID:   uuid.MustParse(targetID),
			Name: "HR",
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

func TestDepartmentService_Update(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()
	targetID := uuid.New()
	companyID := uuid.New()

	t.Run("success", func(t *testing.T) {
		req := department.UpdateDepartmentRequest{Name: "HR Updated"}

		// Mock DB Transaction
		deps.sqlMock.ExpectBegin()

		// Mock Repository calls
		// Pastikan WithTx mengembalikan mock repo yang sama atau mock repo baru
		deps.repo.EXPECT().WithTx(gomock.Any()).Return(deps.repo)

		// Mock FindByIDAndCompany (Harus ada karena dipanggil di service)
		existingDept := &department.Department{
			ID:        targetID,
			CompanyID: companyID,
			Name:      "Old HR",
		}
		deps.repo.EXPECT().
			FindByIDAndCompany(ctx, companyID.String(), targetID.String()).
			Return(existingDept, nil)

		// Mock Update
		deps.repo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, d *department.Department) error {
				assert.Equal(t, req.Name, d.Name)
				assert.Equal(t, targetID, d.ID)
				return nil
			})

		deps.sqlMock.ExpectCommit()

		resp, err := deps.service.Update(ctx, companyID.String(), targetID.String(), req)

		assert.NoError(t, err)
		assert.Equal(t, req.Name, resp.Name)
	})

	t.Run("error - department not found", func(t *testing.T) {
		req := department.UpdateDepartmentRequest{Name: "HR Updated"}

		deps.sqlMock.ExpectBegin()
		deps.repo.EXPECT().WithTx(gomock.Any()).Return(deps.repo)

		// Simulasikan data tidak ditemukan
		deps.repo.EXPECT().
			FindByIDAndCompany(ctx, companyID.String(), targetID.String()).
			Return(nil, errors.New("department not found"))

		deps.sqlMock.ExpectRollback()

		resp, err := deps.service.Update(ctx, companyID.String(), targetID.String(), req)

		assert.Error(t, err)
		assert.Empty(t, resp.ID)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("error - update failed", func(t *testing.T) {
		req := department.UpdateDepartmentRequest{Name: "HR Updated"}

		deps.sqlMock.ExpectBegin()
		deps.repo.EXPECT().WithTx(gomock.Any()).Return(deps.repo)

		existingDept := &department.Department{ID: targetID, CompanyID: companyID}
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

func TestDepartmentService_Delete(t *testing.T) {
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
