package employee_test

import (
	"context"
	"errors"
	"go-hris/internal/employee"
	employeeerrors "go-hris/internal/employee/errors"
	"go-hris/internal/shared/apperror"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type fakeEmployeeService struct {
	CreateFn     func(ctx context.Context, companyID string, req employee.CreateEmployeeRequest) (employee.EmployeeResponse, error)
	GetAllFn     func(ctx context.Context, companyID string) ([]employee.EmployeeResponse, error)
	GetOptionsFn func(ctx context.Context, companyID string) ([]employee.EmployeeResponse, error)
	GetByIDFn    func(ctx context.Context, companyID, id string) (employee.EmployeeResponse, error)
	UpdateFn     func(ctx context.Context, companyID, id string, req employee.UpdateEmployeeRequest) (employee.EmployeeResponse, error)
	DeleteFn     func(ctx context.Context, companyID, id string) error
}

func (f *fakeEmployeeService) Create(ctx context.Context, companyID string, req employee.CreateEmployeeRequest) (employee.EmployeeResponse, error) {
	return f.CreateFn(ctx, companyID, req)
}
func (f *fakeEmployeeService) GetAll(ctx context.Context, companyID string) ([]employee.EmployeeResponse, error) {
	return f.GetAllFn(ctx, companyID)
}
func (f *fakeEmployeeService) GetOptions(ctx context.Context, companyID string) ([]employee.EmployeeResponse, error) {
	return f.GetOptionsFn(ctx, companyID)
}
func (f *fakeEmployeeService) GetByID(ctx context.Context, companyID, id string) (employee.EmployeeResponse, error) {
	return f.GetByIDFn(ctx, companyID, id)
}
func (f *fakeEmployeeService) Update(ctx context.Context, companyID, id string, req employee.UpdateEmployeeRequest) (employee.EmployeeResponse, error) {
	return f.UpdateFn(ctx, companyID, id, req)
}
func (f *fakeEmployeeService) Delete(ctx context.Context, companyID, id string) error {
	return f.DeleteFn(ctx, companyID, id)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func withCompany(companyID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Next()
	}
}

func withUser(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

func TestEmployeeHandler_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		employeeID := uuid.New().String()
		companyID := uuid.New().String()

		// 1. Setup Service Mock
		svc := &fakeEmployeeService{
			CreateFn: func(ctx context.Context, cid string, req employee.CreateEmployeeRequest) (employee.EmployeeResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, "John Doe", req.FullName)
				return employee.EmployeeResponse{
					ID:        uuid.New().String(),
					FullName:  req.FullName,
					Email:     req.Email,
					CompanyID: cid,
				}, nil
			},
		}

		// 2. Setup Handler & Recorder
		h := employee.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 3. Setup Request Body
		body := `{"full_name":"John Doe","email":"john@example.com","employee_number":"EMP-900","phone":"0812","hire_date":"2026-01-01","employment_status":"active","position_id":"` + uuid.New().String() + `"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/employees", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		// 4. Manual Context Injection (Simulasi Middleware)
		// Lewati middleware asli, langsung set value yang dibutuhkan handler
		c.Set("employee_id", employeeID)
		c.Set("company_id", companyID)

		// 5. Panggil Handler Langsung
		h.Create(c)

		// 6. Assertions
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "John Doe")
	})

	t.Run("validation error", func(t *testing.T) {
		// 1. Setup dengan service kosong (tidak akan terpanggil jika validasi gagal)
		svc := &fakeEmployeeService{}
		h := employee.NewHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 2. Kirim JSON kosong atau tidak valid untuk memicu status 400
		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		// 3. Tetap set context value agar tidak error di level middleware (jika ada)
		c.Set("company_id", uuid.New().String())

		// 4. Eksekusi Handler
		h.Create(c)

		// 5. Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		// 1. Setup Mock Service dengan return error
		svc := &fakeEmployeeService{
			CreateFn: func(ctx context.Context, cid string, req employee.CreateEmployeeRequest) (employee.EmployeeResponse, error) {
				return employee.EmployeeResponse{}, errors.New("database connection failed")
			},
		}

		// 2. Setup Handler dan Context Gin secara manual
		h := employee.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 3. Setup Request Body & Context Values
		body := `{"full_name":"HR","email":"hr@company.com","employee_number":"EMP-901","phone":"0813","hire_date":"2026-01-02","employment_status":"active","position_id":"` + uuid.New().String() + `"}`
		req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		// Simulasi data dari middleware (penting agar handler tidak panik/error saat ambil CID)
		c.Set("company_id", uuid.New().String())
		c.Set("employee_id", uuid.New().String())

		// 4. Eksekusi Handler langsung ke fungsinya
		h.Create(c)

		// 5. Assertions
		// Pastikan status code 500 karena service return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// Opsional: Pastikan body mengandung pesan error yang sesuai
		assert.Contains(t, w.Body.String(), "Internal server error")
	})

	t.Run("duplicate employee number returns conflict", func(t *testing.T) {
		companyID := uuid.New().String()
		svc := &fakeEmployeeService{
			CreateFn: func(ctx context.Context, cid string, req employee.CreateEmployeeRequest) (employee.EmployeeResponse, error) {
				return employee.EmployeeResponse{}, employeeerrors.ErrEmployeeNumberAlreadyExists
			},
		}

		h := employee.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"full_name":"John Doe","email":"john2@example.com","employee_number":"EMP-900","phone":"0812","hire_date":"2026-01-01","employment_status":"active","position_id":"` + uuid.New().String() + `"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/employees", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Set("company_id", companyID)

		h.Create(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), apperror.CodeConflict)
		assert.Contains(t, w.Body.String(), "Employee number already exists")
	})
}

func TestEmployeeHandler_GetAll(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()

		svc := &fakeEmployeeService{
			GetAllFn: func(ctx context.Context, cid string) ([]employee.EmployeeResponse, error) {
				assert.Equal(t, companyID, cid)
				return []employee.EmployeeResponse{
					{ID: uuid.New().String(), FullName: "John Doe", Email: "john@example.com"},
					{ID: uuid.New().String(), FullName: "Jane Doe", Email: "jane@example.com"},
				}, nil
			},
		}

		h := employee.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Setup Request & Context
		req := httptest.NewRequest(http.MethodGet, "/employees", nil)
		c.Request = req

		// Simulasi data yang biasanya diset oleh middleware
		c.Set("company_id", companyID)

		// Eksekusi Handler Langsung
		h.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "John Doe")
		assert.Contains(t, w.Body.String(), "Jane Doe")
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeEmployeeService{
			GetAllFn: func(ctx context.Context, cid string) ([]employee.EmployeeResponse, error) {
				return nil, errors.New("database error")
			},
		}

		h := employee.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/employees", nil)
		c.Request = req
		c.Set("company_id", uuid.New().String())

		h.GetAll(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestEmployeeHandler_GetOptions(t *testing.T) {
	t.Run("success - return all options", func(t *testing.T) {
		companyID := uuid.New().String()
		expectedData := []employee.EmployeeResponse{
			{ID: uuid.New().String(), FullName: "Alice Smith", EmployeeNumber: "EMP001"},
			{ID: uuid.New().String(), FullName: "Bob Wilson", EmployeeNumber: "EMP002"},
		}

		svc := &fakeEmployeeService{
			GetOptionsFn: func(ctx context.Context, cid string) ([]employee.EmployeeResponse, error) {
				assert.Equal(t, companyID, cid)
				return expectedData, nil
			},
		}

		h := employee.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/employees/options", nil)
		c.Set("company_id", companyID)

		h.GetOptions(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Alice Smith")
		assert.Contains(t, w.Body.String(), "EMP002")
	})

	t.Run("success - filter by query q", func(t *testing.T) {
		companyID := uuid.New().String()
		svc := &fakeEmployeeService{
			GetOptionsFn: func(ctx context.Context, cid string) ([]employee.EmployeeResponse, error) {
				return []employee.EmployeeResponse{
					{ID: "1", FullName: "Alice Smith", EmployeeNumber: "EMP001"},
					{ID: "2", FullName: "Bob Wilson", EmployeeNumber: "EMP002"},
				}, nil
			},
		}

		h := employee.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Simulasi query pencarian ?q=alice
		c.Request = httptest.NewRequest(http.MethodGet, "/employees/options?q=alice", nil)
		c.Set("company_id", companyID)

		h.GetOptions(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Alice Smith")
		assert.NotContains(t, w.Body.String(), "Bob Wilson")
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeEmployeeService{
			GetOptionsFn: func(ctx context.Context, cid string) ([]employee.EmployeeResponse, error) {
				return nil, errors.New("redis connection failed")
			},
		}

		h := employee.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/employees/options", nil)
		c.Set("company_id", uuid.New().String())

		h.GetOptions(c)

		// Pastikan writeServiceError bekerja (mengembalikan 500 jika error umum)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestEmployeeHandler_GetByID(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		deptID := uuid.New().String()

		svc := &fakeEmployeeService{
			GetByIDFn: func(ctx context.Context, cid, id string) (employee.EmployeeResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, deptID, id)

				return employee.EmployeeResponse{
					ID:        id,
					FullName:  "HR",
					CompanyID: cid,
				}, nil
			},
		}

		r := setupRouter()
		r.Use(withCompany(companyID))

		h := employee.NewHandler(svc)
		r.GET("/employees/:id", h.GetById)

		req := httptest.NewRequest(http.MethodGet, "/employees/"+deptID, nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("cross company error", func(t *testing.T) {
		requestCompanyID := uuid.New().String()
		deptCompanyID := uuid.New().String()
		deptID := uuid.New().String()

		svc := &fakeEmployeeService{
			GetByIDFn: func(ctx context.Context, cid, id string) (employee.EmployeeResponse, error) {
				// simulate cross-company forbidden
				if cid != deptCompanyID {
					return employee.EmployeeResponse{}, errors.New("forbidden")
				}
				return employee.EmployeeResponse{}, nil
			},
		}

		r := setupRouter()
		r.Use(withCompany(requestCompanyID))

		h := employee.NewHandler(svc)
		r.GET("/employees/:id", h.GetById)

		req := httptest.NewRequest(http.MethodGet, "/employees/"+deptID, nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("not found error", func(t *testing.T) {
		svc := &fakeEmployeeService{
			GetByIDFn: func(ctx context.Context, cid, id string) (employee.EmployeeResponse, error) {
				return employee.EmployeeResponse{}, errors.New("not found")
			},
		}

		r := setupRouter()
		r.Use(withCompany(uuid.New().String()))

		h := employee.NewHandler(svc)
		r.GET("/employees/:id", h.GetById)

		req := httptest.NewRequest(http.MethodGet, "/employees/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestEmployeeHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		employeeID := uuid.New().String()

		svc := &fakeEmployeeService{
			UpdateFn: func(ctx context.Context, cid, id string, req employee.UpdateEmployeeRequest) (employee.EmployeeResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, employeeID, id)
				return employee.EmployeeResponse{
					ID:        id,
					FullName:  req.FullName,
					Email:     req.Email, // Pastikan email ikut dikembalikan
					CompanyID: cid,
				}, nil
			},
		}

		h := employee.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// PERBAIKAN: Tambahkan field 'email' agar lolos validasi binding:"required"
		body := `{"full_name":"Finance Update","email":"finance@company.com","employee_number":"EMP-902","phone":"0814","hire_date":"2026-01-03","employment_status":"active","position_id":"` + uuid.New().String() + `"}`
		req := httptest.NewRequest(http.MethodPut, "/employees/"+employeeID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		c.Params = []gin.Param{{Key: "id", Value: employeeID}}
		c.Set("company_id", companyID)

		h.Update(c)

		// Sekarang statusnya harus 200 OK dan body berisi data yang diupdate
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Finance Update")
		assert.Contains(t, w.Body.String(), "finance@company.com")
	})

	t.Run("validation error", func(t *testing.T) {
		h := employee.NewHandler(&fakeEmployeeService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Body kosong memicu validation error
		req := httptest.NewRequest(http.MethodPut, "/employees/123", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "123"}}

		h.Update(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("cross company error", func(t *testing.T) {
		svc := &fakeEmployeeService{
			UpdateFn: func(ctx context.Context, cid, id string, req employee.UpdateEmployeeRequest) (employee.EmployeeResponse, error) {
				// Simulasi service menemukan bahwa ID tersebut bukan milik CompanyID di context
				return employee.EmployeeResponse{}, errors.New("forbidden: cross company access")
			},
		}

		h := employee.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// PERBAIKAN: Payload harus VALID (lengkap) agar lolos validasi Gin (400)
		// dan bisa masuk ke logika Service (500)
		body := `{"full_name":"Finance","email":"finance@company.com","employee_number":"EMP-903","phone":"0815","hire_date":"2026-01-04","employment_status":"active","position_id":"` + uuid.New().String() + `"}`
		req := httptest.NewRequest(http.MethodPut, "/employees/123", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "123"}}
		c.Set("company_id", uuid.New().String())

		h.Update(c)

		// Sekarang, karena binding sukses, handler akan memanggil service.
		// Service mengembalikan error, sehingga handler mengembalikan 500.
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestDeleteEmployeeHandler(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		deptID := uuid.New().String()

		svc := &fakeEmployeeService{
			DeleteFn: func(ctx context.Context, cid, id string) error {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, deptID, id)
				return nil
			},
		}

		r := setupRouter()
		r.Use(withCompany(companyID))

		h := employee.NewHandler(svc)
		r.DELETE("/employees/:id", h.Delete)

		req := httptest.NewRequest(http.MethodDelete, "/employees/"+deptID, nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeEmployeeService{
			DeleteFn: func(ctx context.Context, cid, id string) error {
				return errors.New("failed")
			},
		}

		r := setupRouter()
		r.Use(withCompany(uuid.New().String()))

		h := employee.NewHandler(svc)
		r.DELETE("/employees/:id", h.Delete)

		req := httptest.NewRequest(http.MethodDelete, "/employees/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
