package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"go-hris/internal/auth"
	autherrors "go-hris/internal/auth/errors"
	authMock "go-hris/internal/auth/mock" // sesuaikan path mock Anda
)

func setupAuthRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := authMock.NewMockService(ctrl) // Asumsi nama interface ServiceInterface
	handler := auth.NewHandler(mockService)
	router := setupAuthRouter()
	router.POST("/login", handler.Login)

	t.Run("Success Login - Web Client (Cookie Check)", func(t *testing.T) {
		reqBody := auth.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		expectedResp := auth.AuthResponse{
			ID:        "user-1",
			Email:     "test@example.com",
			CompanyID: "comp-1",
		}

		// Setup Expectation
		mockService.EXPECT().
			Login(gomock.Any(), reqBody.Email, reqBody.Password).
			Return("access-token", "refresh-token", expectedResp, nil)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Client-Type", "WEB") // Trigger cookie logic

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		// Periksa Cookie
		cookies := w.Result().Cookies()
		assert.Len(t, cookies, 2)
		assert.Equal(t, "access_token", cookies[0].Name)
		assert.Equal(t, "access-token", cookies[0].Value)

		// Periksa Body
		var res map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &res)
		assert.Equal(t, "test@example.com", res["data"].(map[string]interface{})["user"].(map[string]interface{})["email"])
	})

	t.Run("Failed Login - Invalid Credentials", func(t *testing.T) {
		mockService.EXPECT().
			Login(gomock.Any(), gomock.Any(), gomock.Any()).
			Return("", "", auth.AuthResponse{}, assert.AnError)

		body, _ := json.Marshal(auth.LoginRequest{Email: "wrong@test.com", Password: "123"})
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := authMock.NewMockService(ctrl)
	handler := auth.NewHandler(mockService)
	router := setupAuthRouter()
	router.POST("/register", handler.Register)

	employeeID := uuid.New()
	companyID := uuid.New()

	t.Run("Success Register", func(t *testing.T) {
		reqData := auth.RegisterRequest{
			Email:      "new@example.com",
			Name:       "New User",
			Password:   "newpassword",
			EmployeeID: employeeID.String(),
			CompanyID:  companyID.String(),
		}
		body, _ := json.Marshal(reqData)

		mockService.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return(auth.AuthResponse{Email: reqData.Email, Name: reqData.Name}, nil)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Failed Register - Validation Error", func(t *testing.T) {
		// Body kosong atau format email salah jika Anda menggunakan tag `binding:"required,email"`
		body := []byte(`{"email": "invalid-email", "name": ""}`)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Harus mengembalikan 400 Bad Request karena ShouldBindJSON gagal
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Memastikan service.Register TIDAK dipanggil jika validasi gagal
		mockService.EXPECT().Register(gomock.Any(), gomock.Any()).Times(0)
	})

	t.Run("Failed Register - Email Already Exists", func(t *testing.T) {
		reqData := auth.RegisterRequest{
			Email:      "exists@example.com",
			Name:       "Existing User",
			Password:   "password123",
			EmployeeID: employeeID.String(),
			CompanyID:  companyID.String(),
		}
		body, _ := json.Marshal(reqData)

		// Simulasi error dari service (misal: Email sudah terdaftar)
		mockService.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return(auth.AuthResponse{}, autherrors.ErrEmailAlreadyRegistered)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		t.Log("Response Body:", w.Body.String()) // Log response body untuk debugging
		// Verifikasi isi response error jika diperlukan
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
