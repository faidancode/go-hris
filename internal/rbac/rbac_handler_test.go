package rbac

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type apiEnvelope struct {
	Ok   bool            `json:"ok"`
	Data json.RawMessage `json:"data"`
}

// =========================================
// Mock Service
// =========================================

type mockService struct{}

func (m *mockService) LoadCompanyPolicy(companyID string) error {
	return nil
}

func (m *mockService) Enforce(req EnforceRequest) (bool, error) {
	if req.Resource == "employee" && req.Action == "read" {
		return true, nil
	}
	return false, nil
}

// =========================================
// TEST: Handler Enforce
// =========================================

func TestHandler_Enforce(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &mockService{}
	handler := NewHandler(service)

	router := gin.Default()
	router.POST("/rbac/enforce", handler.Enforce)

	body := EnforceRequest{
		EmployeeID: "emp-1",
		CompanyID:  "company-1",
		Resource:   "employee",
		Action:     "read",
	}

	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest(
		http.MethodPost,
		"/rbac/enforce",
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var env apiEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &env)
	assert.NoError(t, err)
	assert.True(t, env.Ok)

	var resp EnforceResponse
	err = json.Unmarshal(env.Data, &resp)
	assert.NoError(t, err)

	assert.True(t, resp.Allowed)
}
