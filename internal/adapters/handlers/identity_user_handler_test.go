package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	mocks "github.com/lifenetwork-ai/iam-service/internal/mocks"
	"github.com/stretchr/testify/assert"
)

var (
	mockEmail   = "foo@bar.com"
	mockSession = &dto.IdentityUserChallengeDTO{
		SessionID:   uuid.New().String(),
		Receiver:    mockEmail,
		ChallengeAt: time.Now().Unix(),
	}
	mockError = &dto.ErrorDTOResponse{
		Status:  http.StatusInternalServerError,
		Message: "Internal Server Error",
		Code:    "internal_error",
		Details: nil,
	}
)

func setupTestHandler(t *testing.T) (*userHandler, *mocks.MockIdentityUserUseCase) {
	ctrl := gomock.NewController(t)
	mockUseCase := mocks.NewMockIdentityUserUseCase(ctrl)

	handler := &userHandler{
		ucase: mockUseCase,
	}

	return handler, mockUseCase
}

func makeRequest(method string, body interface{}, headers map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		c.Request = httptest.NewRequest(method, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
	} else {
		c.Request = httptest.NewRequest(method, "/", nil)
	}

	for key, value := range headers {
		c.Request.Header.Set(key, value)
	}

	return c, w
}

func TestUserHandler_ChallengeWithEmail(t *testing.T) {
	handler, mockUseCase := setupTestHandler(t)

	t.Run("Success", func(t *testing.T) {
		payload := dto.IdentityChallengeWithEmailDTO{Email: "test@example.com"}
		c, w := makeRequest("POST", payload, map[string]string{"X-Organization-Id": "test-org"})

		mockUseCase.EXPECT().ChallengeWithEmail(c, payload.Email).Return(mockSession, nil)

		handler.ChallengeWithEmail(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Empty email", func(t *testing.T) {
		payload := dto.IdentityChallengeWithEmailDTO{Email: ""}
		c, w := makeRequest("POST", payload, nil)

		handler.ChallengeWithEmail(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UseCase error", func(t *testing.T) {
		payload := dto.IdentityChallengeWithEmailDTO{Email: "test@example.com"}
		c, w := makeRequest("POST", payload, nil)

		mockUseCase.EXPECT().ChallengeWithEmail(c, payload.Email).Return(nil, mockError)

		handler.ChallengeWithEmail(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserHandler_ChallengeVerify(t *testing.T) {
	handler, mockUseCase := setupTestHandler(t)

	t.Run("Success", func(t *testing.T) {
		payload := dto.IdentityChallengeVerifyDTO{
			SessionID: uuid.New().String(),
			Code:      "123456",
		}
		c, w := makeRequest("POST", payload, nil)

		expectedAuth := &dto.IdentityUserAuthDTO{AccessToken: "token"}
		mockUseCase.EXPECT().ChallengeVerify(c, payload.SessionID, payload.Code).Return(expectedAuth, nil)

		handler.ChallengeVerify(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Empty session ID", func(t *testing.T) {
		payload := dto.IdentityChallengeVerifyDTO{SessionID: "", Code: "123456"}
		c, w := makeRequest("POST", payload, nil)

		handler.ChallengeVerify(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Empty code", func(t *testing.T) {
		payload := dto.IdentityChallengeVerifyDTO{SessionID: uuid.New().String(), Code: ""}
		c, w := makeRequest("POST", payload, nil)

		handler.ChallengeVerify(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UseCase error", func(t *testing.T) {
		payload := dto.IdentityChallengeVerifyDTO{
			SessionID: uuid.New().String(),
			Code:      "123456",
		}
		c, w := makeRequest("POST", payload, nil)

		mockUseCase.EXPECT().ChallengeVerify(c, payload.SessionID, payload.Code).Return(nil, mockError)

		handler.ChallengeVerify(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserHandler_RefreshToken(t *testing.T) {
	handler, mockUseCase := setupTestHandler(t)

	t.Run("Success", func(t *testing.T) {
		payload := dto.IdentityRefreshTokenDTO{RefreshToken: "refresh-token"}
		headers := map[string]string{
			"Authorization":     "Bearer access-token",
			"X-Organization-Id": "test-org",
		}
		c, w := makeRequest("POST", payload, headers)

		expectedAuth := &dto.IdentityUserAuthDTO{AccessToken: "new-token"}
		mockUseCase.EXPECT().RefreshToken(c, "access-token", payload.RefreshToken).Return(expectedAuth, nil)

		handler.RefreshToken(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Missing Authorization header", func(t *testing.T) {
		payload := dto.IdentityRefreshTokenDTO{RefreshToken: "refresh-token"}
		c, w := makeRequest("POST", payload, nil)

		handler.RefreshToken(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Authorization format", func(t *testing.T) {
		payload := dto.IdentityRefreshTokenDTO{RefreshToken: "refresh-token"}
		c, w := makeRequest("POST", payload, map[string]string{"Authorization": "InvalidFormat"})

		handler.RefreshToken(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Empty refresh token", func(t *testing.T) {
		payload := dto.IdentityRefreshTokenDTO{RefreshToken: ""}
		c, w := makeRequest("POST", payload, map[string]string{"Authorization": "Bearer token"})

		handler.RefreshToken(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UseCase error", func(t *testing.T) {
		payload := dto.IdentityRefreshTokenDTO{RefreshToken: "refresh-token"}
		headers := map[string]string{
			"Authorization":     "Bearer access-token",
			"X-Organization-Id": "test-org",
		}
		c, w := makeRequest("POST", payload, headers)

		mockUseCase.EXPECT().RefreshToken(c, "access-token", payload.RefreshToken).Return(nil, mockError)

		handler.RefreshToken(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserHandler_Me(t *testing.T) {
	handler, mockUseCase := setupTestHandler(t)

	t.Run("Success", func(t *testing.T) {
		c, w := makeRequest("GET", nil, map[string]string{
			"Authorization":     "Bearer token",
			"X-Organization-Id": "test-org",
		})

		expectedUser := &dto.IdentityUserDTO{ID: uuid.New().String()}
		mockUseCase.EXPECT().Profile(c).Return(expectedUser, nil)

		handler.Me(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("UseCase error", func(t *testing.T) {
		c, w := makeRequest("GET", nil, map[string]string{"Authorization": "Bearer token"})

		mockUseCase.EXPECT().Profile(c).Return(nil, mockError)

		handler.Me(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserHandler_NotImplementedMethods(t *testing.T) {
	handler, _ := setupTestHandler(t)

	testCases := []struct {
		name   string
		method func(*gin.Context)
	}{
		{"Login", handler.Login},
		{"LoginWithGoogle", handler.LoginWithGoogle},
		{"LoginWithFacebook", handler.LoginWithFacebook},
		{"LoginWithApple", handler.LoginWithApple},
		{"Logout", handler.Logout},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := makeRequest("POST", nil, nil)
			tc.method(c)
			assert.Equal(t, http.StatusNotImplemented, w.Code)
		})
	}
}
