package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	mocks "github.com/lifenetwork-ai/iam-service/internal/mocks"
)

var (
	mockSessionID  = uuid.New().String()
	mockSessionDTO = &dto.AccessSessionDTO{
		ID: mockSessionID,
	}
	mockPaginationResponse = &dto.PaginationDTOResponse{
		NextPage: 1,
		Page:     1,
		Size:     10,
		Total:    1,
		Data: []any{
			*mockSessionDTO,
		},
	}
	mockSessionError = &dto.ErrorDTOResponse{
		Status:  http.StatusInternalServerError,
		Message: "Internal Server Error",
		Code:    "internal_error",
		Details: nil,
	}
)

func setupSessionTestHandler(t *testing.T) (*sessionHandler, *mocks.MockAccessSessionUseCase) {
	ctrl := gomock.NewController(t)
	mockUseCase := mocks.NewMockAccessSessionUseCase(ctrl)

	handler := &sessionHandler{
		ucase: mockUseCase,
	}

	return handler, mockUseCase
}

func makeSessionRequest(method, url string, headers map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(method, url, nil)

	for key, value := range headers {
		c.Request.Header.Set(key, value)
	}

	return c, w
}

func TestSessionHandler_GetSessions(t *testing.T) {
	handler, mockUseCase := setupSessionTestHandler(t)

	t.Run("Success with default parameters", func(t *testing.T) {
		c, w := makeSessionRequest("GET", "/api/v1/sessions", map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		mockUseCase.EXPECT().List(c, 1, 10, "").Return(mockPaginationResponse, nil)

		handler.GetSessions(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.PaginationDTOResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(response.Data))
	})

	t.Run("Success with custom parameters", func(t *testing.T) {
		c, w := makeSessionRequest("GET", "/api/v1/sessions?page=2&size=20&keyword=test", map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		mockUseCase.EXPECT().List(c, 2, 20, "test").Return(mockPaginationResponse, nil)

		handler.GetSessions(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid page number - non-numeric", func(t *testing.T) {
		c, w := makeSessionRequest("GET", "/api/v1/sessions?page=invalid", map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		handler.GetSessions(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid page number - zero", func(t *testing.T) {
		c, w := makeSessionRequest("GET", "/api/v1/sessions?page=0", map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		handler.GetSessions(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid page number - negative", func(t *testing.T) {
		c, w := makeSessionRequest("GET", "/api/v1/sessions?page=-1", map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		handler.GetSessions(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid size - non-numeric", func(t *testing.T) {
		c, w := makeSessionRequest("GET", "/api/v1/sessions?size=invalid", map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		handler.GetSessions(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid size - zero", func(t *testing.T) {
		c, w := makeSessionRequest("GET", "/api/v1/sessions?size=0", map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		handler.GetSessions(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid size - negative", func(t *testing.T) {
		c, w := makeSessionRequest("GET", "/api/v1/sessions?size=-1", map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		handler.GetSessions(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UseCase error", func(t *testing.T) {
		c, w := makeSessionRequest("GET", "/api/v1/sessions", map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		mockUseCase.EXPECT().List(c, 1, 10, "").Return(nil, mockSessionError)

		handler.GetSessions(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSessionHandler_GetDetail(t *testing.T) {
	handler, mockUseCase := setupSessionTestHandler(t)

	t.Run("Success", func(t *testing.T) {
		url := "/api/v1/sessions/" + mockSessionID + "?session_id=" + mockSessionID
		c, w := makeSessionRequest("GET", url, map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		mockUseCase.EXPECT().GetByID(c, mockSessionID).Return(mockSessionDTO, nil)

		handler.GetDetail(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Empty session ID", func(t *testing.T) {
		c, w := makeSessionRequest("GET", "/api/v1/sessions/", map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		handler.GetDetail(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing session_id query parameter", func(t *testing.T) {
		c, w := makeSessionRequest("GET", "/api/v1/sessions/"+mockSessionID, map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		handler.GetDetail(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UseCase error", func(t *testing.T) {
		url := "/api/v1/sessions/" + mockSessionID + "?session_id=" + mockSessionID
		c, w := makeSessionRequest("GET", url, map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		mockUseCase.EXPECT().GetByID(c, mockSessionID).Return(nil, mockSessionError)

		handler.GetDetail(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSessionHandler_DeleteSession(t *testing.T) {
	handler, _ := setupSessionTestHandler(t)

	t.Run("Not implemented", func(t *testing.T) {
		c, w := makeSessionRequest("DELETE", "/api/v1/sessions/"+mockSessionID, map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		handler.DeleteSession(c)

		assert.Equal(t, http.StatusNotImplemented, w.Code)
	})
}
