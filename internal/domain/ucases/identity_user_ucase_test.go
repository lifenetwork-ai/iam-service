package ucases

// import (
// 	"context"
// 	"errors"
// 	"net/http"
// 	"testing"
// 	"time"

// 	client "github.com/ory/kratos-client-go"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"go.uber.org/mock/gomock"

// 	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
// 	mock_types "github.com/lifenetwork-ai/iam-service/mocks/adapters/repositories/types"
// 	mock_services "github.com/lifenetwork-ai/iam-service/mocks/adapters/services"
// )

// func setupTestUseCase(t *testing.T) (*userUseCase, *mock_types.MockChallengeSessionRepository, *mock_services.MockKratosService, *gomock.Controller) {
// 	ctrl := gomock.NewController(t)
// 	mockChallengeRepo := mock_types.NewMockChallengeSessionRepository(ctrl)
// 	mockKratosService := mock_services.NewMockKratosService(ctrl)

// 	useCase := &userUseCase{
// 		challengeSessionRepo: mockChallengeRepo,
// 		kratosService:        mockKratosService,
// 	}

// 	return useCase, mockChallengeRepo, mockKratosService, ctrl
// }

// func TestChallengeWithPhone(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		phone          string
// 		mockSetup      func(*mock_types.MockChallengeSessionRepository, *mock_services.MockKratosService)
// 		expectedStatus int
// 		expectedCode   string
// 	}{
// 		{
// 			name:  "Success",
// 			phone: "+819012345678",
// 			mockSetup: func(repo *mock_types.MockChallengeSessionRepository, svc *mock_services.MockKratosService) {
// 				flow := &client.LoginFlow{Id: "flow-id"}
// 				svc.EXPECT().InitializeLoginFlow(gomock.Any()).Return(flow, nil)
// 				svc.EXPECT().SubmitLoginFlow(gomock.Any(), flow, "code", gomock.Any(), nil, nil).Return(nil, nil)
// 				repo.EXPECT().SaveChallenge(gomock.Any(), "flow-id", gomock.Any(), 5*time.Minute).Return(nil)
// 			},
// 		},
// 		{
// 			name:           "Invalid Phone",
// 			phone:          "invalid",
// 			expectedStatus: http.StatusBadRequest,
// 			expectedCode:   "INVALID_PHONE_NUMBER",
// 		},
// 		{
// 			name:  "Kratos Flow Error",
// 			phone: "+8109012345678",
// 			mockSetup: func(repo *mock_types.MockChallengeSessionRepository, svc *mock_services.MockKratosService) {
// 				svc.EXPECT().InitializeLoginFlow(gomock.Any()).Return(nil, errors.New("kratos error"))
// 			},
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedCode:   "VERIFICATION_FLOW_FAILED",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			useCase, mockRepo, mockSvc, ctrl := setupTestUseCase(t)
// 			defer ctrl.Finish()

// 			if tt.mockSetup != nil {
// 				tt.mockSetup(mockRepo, mockSvc)
// 			}

// 			result, err := useCase.ChallengeWithPhone(context.Background(), tt.phone)

// 			if tt.expectedStatus != 0 {
// 				require.NotNil(t, err)
// 				assert.Equal(t, tt.expectedStatus, err.Status)
// 				assert.Equal(t, tt.expectedCode, err.Code)
// 				assert.Nil(t, result)
// 			} else {
// 				require.Nil(t, err)
// 				require.NotNil(t, result)
// 				assert.Equal(t, "flow-id", result.FlowID)
// 				assert.Equal(t, tt.phone, result.Receiver)
// 			}
// 		})
// 	}
// }

// func TestVerifyLogin(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		mockSetup      func(*mock_types.MockChallengeSessionRepository, *mock_services.MockKratosService)
// 		expectedStatus int
// 		expectedCode   string
// 	}{
// 		{
// 			name: "Success",
// 			mockSetup: func(repo *mock_types.MockChallengeSessionRepository, svc *mock_services.MockKratosService) {
// 				flow := &client.LoginFlow{Id: "flow-id"}
// 				session := &entities.ChallengeSession{Phone: "+1234567890", Flow: "flow-id"}
// 				loginResult := &client.SuccessfulNativeLogin{
// 					Session: client.Session{
// 						Active:    boolPtr(true),
// 						Id:        "session-id",
// 						ExpiresAt: timePtr(time.Now().Add(time.Hour)),
// 						Identity: &client.Identity{
// 							Id:     "user-id",
// 							Traits: map[string]interface{}{"username": "test"},
// 						},
// 					},
// 					SessionToken: stringPtr("session-token"),
// 				}

// 				svc.EXPECT().GetLoginFlow(gomock.Any(), "flow-id").Return(flow, nil)
// 				repo.EXPECT().GetChallenge(gomock.Any(), "flow-id").Return(session, nil)
// 				svc.EXPECT().SubmitLoginFlow(gomock.Any(), flow, "code", &session.Phone, nil, stringPtr("123456")).Return(loginResult, nil)
// 			},
// 		},
// 		{
// 			name: "Session Not Found",
// 			mockSetup: func(repo *mock_types.MockChallengeSessionRepository, svc *mock_services.MockKratosService) {
// 				flow := &client.LoginFlow{Id: "flow-id"}
// 				svc.EXPECT().GetLoginFlow(gomock.Any(), "flow-id").Return(flow, nil)
// 				repo.EXPECT().GetChallenge(gomock.Any(), "flow-id").Return(nil, errors.New("not found"))
// 			},
// 			expectedStatus: http.StatusNotFound,
// 			expectedCode:   "MSG_CHALLENGE_SESSION_NOT_FOUND",
// 		},
// 		{
// 			name: "Nil Session",
// 			mockSetup: func(repo *mock_types.MockChallengeSessionRepository, svc *mock_services.MockKratosService) {
// 				flow := &client.LoginFlow{Id: "flow-id"}
// 				svc.EXPECT().GetLoginFlow(gomock.Any(), "flow-id").Return(flow, nil)
// 				repo.EXPECT().GetChallenge(gomock.Any(), "flow-id").Return(nil, nil)
// 			},
// 			expectedStatus: http.StatusNotFound,
// 			expectedCode:   "MSG_CHALLENGE_SESSION_NOT_FOUND",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			useCase, mockRepo, mockSvc, ctrl := setupTestUseCase(t)
// 			defer ctrl.Finish()

// 			tt.mockSetup(mockRepo, mockSvc)

// 			result, err := useCase.VerifyLogin(context.Background(), "flow-id", "123456")

// 			if tt.expectedStatus != 0 {
// 				require.NotNil(t, err)
// 				assert.Equal(t, tt.expectedStatus, err.Status)
// 				assert.Equal(t, tt.expectedCode, err.Code)
// 			} else {
// 				require.Nil(t, err)
// 				require.NotNil(t, result)
// 				assert.Equal(t, "session-id", result.SessionID)
// 				assert.Equal(t, "session-token", result.SessionToken)
// 				assert.Equal(t, "user-id", result.User.ID)
// 			}
// 		})
// 	}
// }

// func TestProfile(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		ctx            context.Context
// 		mockSetup      func(*mock_services.MockKratosService)
// 		expectedStatus int
// 		expectedCode   string
// 	}{
// 		{
// 			name: "Success",
// 			ctx:  context.WithValue(context.Background(), "sessionToken", "valid-token"),
// 			mockSetup: func(svc *mock_services.MockKratosService) {
// 				session := &client.Session{
// 					Identity: &client.Identity{
// 						Id: "user-id",
// 						Traits: map[string]interface{}{
// 							"username": "testuser",
// 							"email":    "test@example.com",
// 						},
// 					},
// 				}
// 				svc.EXPECT().WhoAmI(gomock.Any(), "valid-token").Return(session, nil)
// 			},
// 		},
// 		{
// 			name:           "Missing Token",
// 			ctx:            context.Background(),
// 			expectedStatus: http.StatusUnauthorized,
// 			expectedCode:   "MSG_SESSION_TOKEN_MISSING",
// 		},
// 		{
// 			name: "Invalid Session",
// 			ctx:  context.WithValue(context.Background(), "sessionToken", "invalid-token"),
// 			mockSetup: func(svc *mock_services.MockKratosService) {
// 				svc.EXPECT().WhoAmI(gomock.Any(), "invalid-token").Return(nil, errors.New("invalid"))
// 			},
// 			expectedStatus: http.StatusUnauthorized,
// 			expectedCode:   "MSG_INVALID_SESSION",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			useCase, _, mockSvc, ctrl := setupTestUseCase(t)
// 			defer ctrl.Finish()

// 			if tt.mockSetup != nil {
// 				tt.mockSetup(mockSvc)
// 			}

// 			result, err := useCase.Profile(tt.ctx)

// 			if tt.expectedStatus != 0 {
// 				require.NotNil(t, err)
// 				assert.Equal(t, tt.expectedStatus, err.Status)
// 				assert.Equal(t, tt.expectedCode, err.Code)
// 			} else {
// 				require.Nil(t, err)
// 				require.NotNil(t, result)
// 				assert.Equal(t, "user-id", result.ID)
// 			}
// 		})
// 	}
// }

// func TestVerifyRegister(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		mockSetup      func(*mock_services.MockKratosService)
// 		expectedStatus int
// 		expectedCode   string
// 	}{
// 		{
// 			name: "Success",
// 			mockSetup: func(svc *mock_services.MockKratosService) {
// 				flow := &client.RegistrationFlow{Id: "flow-id"}
// 				registrationResult := &client.SuccessfulNativeRegistration{
// 					Session: &client.Session{
// 						Id:        "session-id",
// 						Active:    boolPtr(true),
// 						ExpiresAt: timePtr(time.Now().Add(time.Hour)),
// 						Identity: &client.Identity{
// 							Id: "user-id",
// 							Traits: map[string]interface{}{
// 								"username": "test",
// 								"email":    "test@example.com",
// 								"phone":    "+1234567890",
// 							},
// 						},
// 					},
// 					SessionToken: stringPtr("session-token"),
// 				}

// 				svc.EXPECT().GetRegistrationFlow(gomock.Any(), "flow-id").Return(flow, nil)
// 				svc.EXPECT().SubmitRegistrationFlowWithCode(gomock.Any(), flow, "123456").Return(registrationResult, nil)
// 			},
// 		},
// 		{
// 			name: "Flow Not Found",
// 			mockSetup: func(svc *mock_services.MockKratosService) {
// 				svc.EXPECT().GetRegistrationFlow(gomock.Any(), "flow-id").Return(nil, errors.New("not found"))
// 			},
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedCode:   "MSG_GET_FLOW_FAILED",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			useCase, _, mockSvc, ctrl := setupTestUseCase(t)
// 			defer ctrl.Finish()

// 			tt.mockSetup(mockSvc)

// 			result, err := useCase.VerifyRegister(context.Background(), "flow-id", "123456")

// 			if tt.expectedStatus != 0 {
// 				require.NotNil(t, err)
// 				assert.Equal(t, tt.expectedStatus, err.Status)
// 				assert.Equal(t, tt.expectedCode, err.Code)
// 			} else {
// 				require.Nil(t, err)
// 				require.NotNil(t, result)
// 				assert.Equal(t, "session-id", result.SessionID)
// 				assert.Equal(t, "session-token", result.SessionToken)
// 				assert.Equal(t, "user-id", result.User.ID)
// 				assert.Equal(t, "test", result.User.UserName)
// 				assert.Equal(t, "test@example.com", result.User.Email)
// 			}
// 		})
// 	}
// }

// func TestTraitExtraction(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		traits   map[string]interface{}
// 		key      string
// 		expected string
// 	}{
// 		{"String value", map[string]interface{}{"key": "value"}, "key", "value"},
// 		{"String pointer", map[string]interface{}{"key": stringPtr("value")}, "key", "value"},
// 		{"Nil pointer", map[string]interface{}{"key": (*string)(nil)}, "key", "default"},
// 		{"Missing key", map[string]interface{}{}, "key", "default"},
// 		{"Number to string", map[string]interface{}{"key": 123}, "key", "123"},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := extractStringFromTraits(tt.traits, tt.key, "default")
// 			assert.Equal(t, tt.expected, result)
// 		})
// 	}
// }

// // Helper functions
// func stringPtr(s string) *string {
// 	return &s
// }

// func timePtr(t time.Time) *time.Time {
// 	return &t
// }

// func boolPtr(b bool) *bool {
// 	return &b
// }
