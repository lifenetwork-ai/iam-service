package ucases

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services"
	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/lifenetwork-ai/iam-service/internal/mocks"
	"github.com/stretchr/testify/assert"
)

const (
	ValidSessionID   = "valid-session-id"
	InvalidSessionID = "invalid-session-id"
	ValidCode        = "123456"
	InvalidCode      = "654321"
	ValidEmail       = "email"
	ValidUserName    = "username"
	ValidUserID      = "valid-user-id"

	ValidPhone       = "+1234567890"
	InvalidPhone     = "invalid-phone"
	ValidPhoneUserID = "valid-phone-user-id"
)

var (
	ValidEmailUser = entities.IdentityUser{
		ID:       ValidUserID,
		Email:    ValidEmail,
		UserName: ValidUserName,
	}

	ValidPhoneUser = entities.IdentityUser{
		ID:    ValidPhoneUserID,
		Phone: ValidPhone,
	}
)

func TestUserUseCase_ChallengeWithEmail_OTP(t *testing.T) {
	tests := []struct {
		name              string
		email             string
		userExists        bool
		emailServiceError error
		sessionSaveError  error
		expectedSuccess   bool
		expectedErrorCode string
		expectedStatus    int
	}{
		{
			name:              "Invalid email format",
			email:             "invalid-email",
			expectedSuccess:   false,
			expectedErrorCode: "INVALID_EMAIL",
			expectedStatus:    http.StatusBadRequest,
		},
		{
			name:              "Valid email - existing user - OTP sent successfully",
			email:             "existing@example.com",
			userExists:        true,
			emailServiceError: nil,
			sessionSaveError:  nil,
			expectedSuccess:   true,
		},
		{
			name:              "Valid email - new user - OTP sent successfully",
			email:             "newuser@example.com",
			userExists:        false,
			emailServiceError: nil,
			sessionSaveError:  nil,
			expectedSuccess:   true,
		},
		{
			name:              "Email service fails to send OTP",
			email:             "test@example.com",
			userExists:        true,
			emailServiceError: errors.New("SMTP server unavailable"),
			expectedSuccess:   false,
			expectedErrorCode: "MSG_SENDING_OTP_FAILED",
			expectedStatus:    http.StatusInternalServerError,
		},
		{
			name:              "OTP session save fails",
			email:             "test@example.com",
			userExists:        true,
			emailServiceError: nil,
			sessionSaveError:  errors.New("Redis connection failed"),
			expectedSuccess:   false,
			expectedErrorCode: "MSG_SAVING_SESSION_FAILED",
			expectedStatus:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mocks.NewMockIdentityUserRepository(ctrl)
			mockSessionRepo := mocks.NewMockAccessSessionRepository(ctrl)
			mockChallengeSessionRepo := mocks.NewMockChallengeSessionRepository(ctrl)
			mockEmailService := mocks.NewMockEmailService(ctrl)
			mockSMSService := mocks.NewMockSMSService(ctrl)
			mockJWTService := mocks.NewMockJWTService(ctrl)

			if tt.email != "invalid-email" {
				if tt.userExists {
					existingUser := &entities.IdentityUser{
						ID:       "user-123",
						Email:    tt.email,
						UserName: tt.email,
					}
					mockUserRepo.EXPECT().
						FindByEmail(gomock.Any(), tt.email).
						Return(existingUser, nil)
				} else {
					mockUserRepo.EXPECT().
						FindByEmail(gomock.Any(), tt.email).
						Return(nil, nil)
					mockUserRepo.EXPECT().
						Create(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, user *entities.IdentityUser) error {
							assert.Equal(t, tt.email, user.Email)
							assert.Equal(t, tt.email, user.UserName)
							return nil
						})
				}

				var capturedOTP string
				mockEmailService.EXPECT().
					SendOTP(gomock.Any(), tt.email, gomock.Any()).
					DoAndReturn(func(ctx context.Context, email, otp string) error {
						capturedOTP = otp
						return tt.emailServiceError
					})

				if tt.emailServiceError == nil {
					mockChallengeSessionRepo.EXPECT().
						SaveChallenge(gomock.Any(), gomock.Any(), gomock.Any(), 5*time.Minute).
						DoAndReturn(func(ctx context.Context, sessionID string, session *entities.ChallengeSession, ttl time.Duration) error {
							assert.Equal(t, "email", session.Type)
							assert.Equal(t, tt.email, session.Email)
							assert.Equal(t, capturedOTP, session.OTP)
							assert.NotEmpty(t, session.OTP)
							return tt.sessionSaveError
						})
				}
			}

			useCase := NewIdentityUserUseCase(
				mockUserRepo,
				mockSessionRepo,
				mockChallengeSessionRepo,
				mockEmailService,
				mockSMSService,
				mockJWTService,
			)

			ctx := context.Background()
			result, err := useCase.ChallengeWithEmail(ctx, tt.email)

			if tt.expectedSuccess {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.email, result.Receiver)
				assert.NotEmpty(t, result.SessionID)
				assert.True(t, result.ChallengeAt > 0)
				assert.Len(t, result.SessionID, 36)
				now := time.Now().Unix()
				assert.True(t, result.ChallengeAt <= now && result.ChallengeAt >= now-5)
			} else {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedErrorCode, err.Code)
				assert.Equal(t, tt.expectedStatus, err.Status)
			}
		})
	}
}

func TestUserUseCase_ChallengeVerify_Email(t *testing.T) {
	conf.GetConfiguration().Env = "PROD"
	t.Run("Challenge session valid - Email type - User not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockIdentityUserRepository(ctrl)
		mockSessionRepo := mocks.NewMockAccessSessionRepository(ctrl)
		mockChallengeSessionRepo := mocks.NewMockChallengeSessionRepository(ctrl)
		mockEmailService := mocks.NewMockEmailService(ctrl)
		mockSMSService := mocks.NewMockSMSService(ctrl)
		mockJWTService := mocks.NewMockJWTService(ctrl)

		useCase := NewIdentityUserUseCase(
			mockUserRepo,
			mockSessionRepo,
			mockChallengeSessionRepo,
			mockEmailService,
			mockSMSService,
			mockJWTService,
		)

		ctx := context.WithValue(context.Background(), "organizationId", "org-123") // nolint:staticcheck

		sessionID := ValidSessionID
		code := ValidCode // FIXED: was ValidSessionID
		expectedSessionValue := entities.ChallengeSession{
			Type:  "email",
			Email: "email",
			OTP:   code,
		}

		mockChallengeSessionRepo.EXPECT().GetChallenge(ctx, sessionID).Return(&expectedSessionValue, nil)
		mockUserRepo.EXPECT().FindByEmail(ctx, expectedSessionValue.Email).Return(nil, nil)
		auth, err := useCase.ChallengeVerify(ctx, sessionID, code)

		assert.NotNil(t, err)
		assert.Nil(t, auth)
		assert.Equal(t, "USER_NOT_FOUND", err.Code)
	})

	t.Run("Challenge session valid - Email type - User found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockIdentityUserRepository(ctrl)
		mockSessionRepo := mocks.NewMockAccessSessionRepository(ctrl)
		mockChallengeSessionRepo := mocks.NewMockChallengeSessionRepository(ctrl)
		mockEmailService := mocks.NewMockEmailService(ctrl)
		mockSMSService := mocks.NewMockSMSService(ctrl)
		mockJWTService := mocks.NewMockJWTService(ctrl)

		useCase := NewIdentityUserUseCase(
			mockUserRepo,
			mockSessionRepo,
			mockChallengeSessionRepo,
			mockEmailService,
			mockSMSService,
			mockJWTService,
		)

		ctx := context.WithValue(context.Background(), "organizationId", "org-123") // nolint:staticcheck

		sessionID := ValidSessionID
		code := ValidCode // FIXED: was ValidSessionID
		expectedSessionValue := entities.ChallengeSession{
			Type:  "email",
			Email: ValidEmail,
			OTP:   code,
		}

		jwtToken := services.JWTToken{
			AccessToken:        "access-token",
			RefreshToken:       "refresh-token",
			AccessTokenExpiry:  int64(30 * time.Minute),
			RefreshTokenExpiry: int64(24 * time.Hour),
			ClaimAt:            time.Now(),
		}

		mockChallengeSessionRepo.EXPECT().GetChallenge(ctx, sessionID).Return(&expectedSessionValue, nil)
		mockUserRepo.EXPECT().FindByEmail(ctx, expectedSessionValue.Email).Return(&ValidEmailUser, nil)
		mockJWTService.EXPECT().GenerateToken(ctx, gomock.Any()).Return(&jwtToken, nil)
		mockSessionRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, nil)
		auth, err := useCase.ChallengeVerify(ctx, sessionID, code)

		assert.Nil(t, err)
		assert.NotNil(t, auth)
		assert.Equal(t, ValidUserID, auth.User.ID)
		assert.Equal(t, expectedSessionValue.Email, auth.User.Email)
		assert.Equal(t, jwtToken.AccessToken, auth.AccessToken)
		assert.Equal(t, jwtToken.RefreshToken, auth.RefreshToken)
		assert.Equal(t, ValidEmailUser.ID, auth.User.ID)
	})

	t.Run("Challenge session not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockIdentityUserRepository(ctrl)
		mockSessionRepo := mocks.NewMockAccessSessionRepository(ctrl)
		mockChallengeSessionRepo := mocks.NewMockChallengeSessionRepository(ctrl)
		mockEmailService := mocks.NewMockEmailService(ctrl)
		mockSMSService := mocks.NewMockSMSService(ctrl)
		mockJWTService := mocks.NewMockJWTService(ctrl)

		useCase := NewIdentityUserUseCase(
			mockUserRepo,
			mockSessionRepo,
			mockChallengeSessionRepo,
			mockEmailService,
			mockSMSService,
			mockJWTService,
		)

		ctx := context.Background()

		sessionID := "non-existent-session-id"
		code := "123456"

		// FIXED: Return error instead of nil, nil
		mockChallengeSessionRepo.EXPECT().GetChallenge(ctx, sessionID).Return(nil, errors.New("session not found"))

		auth, err := useCase.ChallengeVerify(ctx, sessionID, code)
		assert.NotNil(t, err)
		assert.Nil(t, auth)
		assert.Equal(t, "MSG_CHALLENGE_SESSION_NOT_FOUND", err.Code)
		assert.Equal(t, http.StatusNotFound, err.Status)
	})

	t.Run("Valid session, empty otp", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockIdentityUserRepository(ctrl)
		mockSessionRepo := mocks.NewMockAccessSessionRepository(ctrl)
		mockChallengeSessionRepo := mocks.NewMockChallengeSessionRepository(ctrl)
		mockEmailService := mocks.NewMockEmailService(ctrl)
		mockSMSService := mocks.NewMockSMSService(ctrl)
		mockJWTService := mocks.NewMockJWTService(ctrl)

		useCase := NewIdentityUserUseCase(
			mockUserRepo,
			mockSessionRepo,
			mockChallengeSessionRepo,
			mockEmailService,
			mockSMSService,
			mockJWTService,
		)

		ctx := context.Background()

		sessionID := ValidSessionID
		code := ""
		expectedSessionValue := entities.ChallengeSession{
			Type:  "email",
			Email: "email",
			OTP:   ValidCode, // FIXED: was empty code
		}
		mockChallengeSessionRepo.EXPECT().GetChallenge(ctx, sessionID).Return(&expectedSessionValue, nil)
		auth, err := useCase.ChallengeVerify(ctx, sessionID, code)
		assert.NotNil(t, err)
		assert.Nil(t, auth)
		assert.Equal(t, "MSG_INVALID_CODE", err.Code)
	})

	t.Run("Valid session, invalid otp", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockIdentityUserRepository(ctrl)
		mockSessionRepo := mocks.NewMockAccessSessionRepository(ctrl)
		mockChallengeSessionRepo := mocks.NewMockChallengeSessionRepository(ctrl)
		mockEmailService := mocks.NewMockEmailService(ctrl)
		mockSMSService := mocks.NewMockSMSService(ctrl)
		mockJWTService := mocks.NewMockJWTService(ctrl)

		useCase := NewIdentityUserUseCase(
			mockUserRepo,
			mockSessionRepo,
			mockChallengeSessionRepo,
			mockEmailService,
			mockSMSService,
			mockJWTService,
		)

		ctx := context.Background()

		sessionID := ValidSessionID
		code := "invalid-code"
		expectedSessionValue := entities.ChallengeSession{
			Type:  "email",
			Email: "email",
			OTP:   ValidCode, // FIXED: was "123456"
		}
		mockChallengeSessionRepo.EXPECT().GetChallenge(ctx, sessionID).Return(&expectedSessionValue, nil)

		auth, err := useCase.ChallengeVerify(ctx, sessionID, code)
		assert.NotNil(t, err)
		assert.Nil(t, auth)
		assert.Equal(t, "MSG_INVALID_CODE", err.Code)
	})
}
