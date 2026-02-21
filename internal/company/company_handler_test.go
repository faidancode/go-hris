package company_test

import (
	"bytes"
	"encoding/json"
	"go-hris/internal/company"
	companyMock "go-hris/internal/company/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_GetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := companyMock.NewMockService(ctrl)
	handler := company.NewHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		compID := "comp-123"
		mockResp := &company.CompanyResponse{
			ID:   compID,
			Name: "Test Company",
		}

		mockService.EXPECT().GetByID(gomock.Any(), compID).Return(mockResp, nil)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.Use(func(c *gin.Context) {
			c.Set("company_id", compID)
			c.Next()
		})

		r.GET("/me", handler.GetMe)
		req, _ := http.NewRequest(http.MethodGet, "/me", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var res map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &res)
		assert.Equal(t, true, res["ok"])
	})
}

func TestHandler_UpdateMe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := companyMock.NewMockService(ctrl)
	handler := company.NewHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		compID := "comp-123"
		reqBody := company.UpdateCompanyRequest{
			Name: "Updated Name",
		}
		mockResp := &company.CompanyResponse{
			ID:   compID,
			Name: "Updated Name",
		}

		mockService.EXPECT().Update(gomock.Any(), compID, gomock.Any()).Return(mockResp, nil)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.Use(func(c *gin.Context) {
			c.Set("company_id", compID)
			c.Next()
		})

		jsonReq, _ := json.Marshal(reqBody)
		r.PUT("/me", handler.UpdateMe)
		req, _ := http.NewRequest(http.MethodPut, "/me", bytes.NewBuffer(jsonReq))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
