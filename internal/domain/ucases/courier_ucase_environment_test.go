package ucases

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/infrastructures/caching"
	mock_services "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/services"
	mock_otpqueue "github.com/lifenetwork-ai/iam-service/mocks/infrastructures/otp_queue/types"
	"github.com/patrickmn/go-cache"
)

// TestCourierUseCase_ChooseChannel_EnvironmentBasedRouting tests that SMS routing to SpeedSMS
// only happens in Staging or Production environments
func TestCourierUseCase_ChooseChannel_EnvironmentBasedRouting(t *testing.T) {
	testCases := []struct {
		name                 string
		environment          string
		tenantName           string
		receiver             string
		channel              string
		expectedActualChannel string
		shouldRouteToSpeedSMS bool
	}{
		// Production environment cases
		{
			name:                 "PROD environment routes Vietnamese phone to SpeedSMS",
			environment:          "PROD",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+84344381024",
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSpeedSMS,
			shouldRouteToSpeedSMS: true,
		},
		{
			name:                 "PRODUCTION environment routes Vietnamese phone to SpeedSMS",
			environment:          "PRODUCTION",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+84987654321",
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSpeedSMS,
			shouldRouteToSpeedSMS: true,
		},
		{
			name:                 "production (lowercase) environment routes Vietnamese phone to SpeedSMS",
			environment:          "production",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+84912345678",
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSpeedSMS,
			shouldRouteToSpeedSMS: true,
		},

		// Staging environment cases
		{
			name:                 "STAGING environment routes Vietnamese phone to SpeedSMS",
			environment:          "STAGING",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+84344381024",
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSpeedSMS,
			shouldRouteToSpeedSMS: true,
		},
		{
			name:                 "staging (lowercase) environment routes Vietnamese phone to SpeedSMS",
			environment:          "staging",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+84987654321",
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSpeedSMS,
			shouldRouteToSpeedSMS: true,
		},

		// Development environment cases - should NOT route to SpeedSMS
		{
			name:                 "DEV environment does NOT route Vietnamese phone to SpeedSMS",
			environment:          "DEV",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+84344381024",
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSMS,
			shouldRouteToSpeedSMS: false,
		},
		{
			name:                 "DEVELOPMENT environment does NOT route Vietnamese phone to SpeedSMS",
			environment:          "DEVELOPMENT",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+84987654321",
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSMS,
			shouldRouteToSpeedSMS: false,
		},
		{
			name:                 "dev (lowercase) environment does NOT route Vietnamese phone to SpeedSMS",
			environment:          "dev",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+84912345678",
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSMS,
			shouldRouteToSpeedSMS: false,
		},

		// Local/Test environment cases - should NOT route to SpeedSMS
		{
			name:                 "LOCAL environment does NOT route Vietnamese phone to SpeedSMS",
			environment:          "LOCAL",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+84344381024",
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSMS,
			shouldRouteToSpeedSMS: false,
		},
		{
			name:                 "TEST environment does NOT route Vietnamese phone to SpeedSMS",
			environment:          "TEST",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+84987654321",
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSMS,
			shouldRouteToSpeedSMS: false,
		},

		// Non-Vietnamese phone numbers should never route to SpeedSMS regardless of environment
		{
			name:                 "PROD environment does NOT route non-Vietnamese phone to SpeedSMS",
			environment:          "PROD",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+66812345678", // Thailand
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSMS,
			shouldRouteToSpeedSMS: false,
		},
		{
			name:                 "STAGING environment does NOT route US phone to SpeedSMS",
			environment:          "STAGING",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+12025551234", // US
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSMS,
			shouldRouteToSpeedSMS: false,
		},

		// Non-SMS channels should never route to SpeedSMS
		{
			name:                 "PROD environment does NOT route WhatsApp to SpeedSMS",
			environment:          "PROD",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+84344381024",
			channel:              constants.ChannelWhatsApp,
			expectedActualChannel: constants.ChannelWhatsApp,
			shouldRouteToSpeedSMS: false,
		},
		{
			name:                 "STAGING environment does NOT route Zalo to SpeedSMS",
			environment:          "STAGING",
			tenantName:           constants.TenantGenetica, // Genetica supports Zalo
			receiver:             "+84987654321",
			channel:              constants.ChannelZalo,
			expectedActualChannel: constants.ChannelZalo,
			shouldRouteToSpeedSMS: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup: Set environment for testing
			originalEnv := conf.GetEnvironment()
			defer func() {
				// Cleanup: Restore original environment
				conf.SetEnvironmentForTesting(originalEnv)
			}()

			// Set the test environment
			conf.SetEnvironmentForTesting(tc.environment)

			// Create test dependencies
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockQueue := mock_otpqueue.NewMockOTPQueueRepository(ctrl)
			mockSMSProvider := mock_services.NewMockSMSProvider(ctrl)

			inMemCache := caching.NewCachingRepository(
				context.Background(),
				caching.NewGoCacheClient(cache.New(5*time.Minute, 10*time.Minute)),
			)

			courierUseCase := NewCourierUseCase(mockQueue, mockSMSProvider, inMemCache)

			// Use tenant from test case if specified, otherwise default to LifeAI
			tenantName := tc.tenantName
			if tenantName == "" {
				tenantName = constants.TenantLifeAI
			}

			// Execute: Choose channel
			err := courierUseCase.ChooseChannel(ctx, tenantName, tc.receiver, tc.channel)
			require.Nil(t, err, "Failed to choose channel")

			// Verify: Check that the correct channel was saved
			channelResponse, getErr := courierUseCase.GetChannel(ctx, tenantName, tc.receiver)
			require.Nil(t, getErr, "Failed to get channel")
			require.Equal(t, tc.expectedActualChannel, channelResponse.Channel,
				"Environment: %s, Expected channel: %s, Got: %s",
				tc.environment, tc.expectedActualChannel, channelResponse.Channel)
		})
	}
}

// TestShouldRouteToSpeedSMS_EnvironmentCheck tests the shouldRouteToSpeedSMS helper function
func TestShouldRouteToSpeedSMS_EnvironmentCheck(t *testing.T) {
	testCases := []struct {
		name        string
		environment string
		shouldRoute bool
	}{
		// Should route in production environments
		{"PROD should route", "PROD", true},
		{"PRODUCTION should route", "PRODUCTION", true},
		{"production (lowercase) should route", "production", true},
		{"Production (mixed case) should route", "Production", true},

		// Should route in staging environments
		{"STAGING should route", "STAGING", true},
		{"staging (lowercase) should route", "staging", true},
		{"Staging (mixed case) should route", "Staging", true},

		// Should NOT route in other environments
		{"DEV should not route", "DEV", false},
		{"DEVELOPMENT should not route", "DEVELOPMENT", false},
		{"dev (lowercase) should not route", "dev", false},
		{"LOCAL should not route", "LOCAL", false},
		{"local (lowercase) should not route", "local", false},
		{"TEST should not route", "TEST", false},
		{"test (lowercase) should not route", "test", false},
		{"QA should not route", "QA", false},
		{"UAT should not route", "UAT", false},
		{"empty string should not route", "", false},
		{"random environment should not route", "RANDOM", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup: Set environment for testing
			originalEnv := conf.GetEnvironment()
			defer func() {
				// Cleanup: Restore original environment
				conf.SetEnvironmentForTesting(originalEnv)
			}()

			// Set the test environment
			conf.SetEnvironmentForTesting(tc.environment)

			// Execute and verify
			result := shouldRouteToSpeedSMS()
			require.Equal(t, tc.shouldRoute, result,
				"Environment: %s, Expected shouldRoute: %v, Got: %v",
				tc.environment, tc.shouldRoute, result)
		})
	}
}

// TestCourierUseCase_ChooseChannel_EdgeCases tests edge cases for environment-based routing
func TestCourierUseCase_ChooseChannel_EdgeCases(t *testing.T) {
	testCases := []struct {
		name                 string
		environment          string
		tenantName           string
		receiver             string
		channel              string
		expectedActualChannel string
	}{
		{
			name:                 "Genetica tenant in PROD routes Vietnamese to SpeedSMS",
			environment:          "PROD",
			tenantName:           constants.TenantGenetica,
			receiver:             "+84344381024",
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSpeedSMS,
		},
		{
			name:                 "Genetica tenant in DEV does NOT route Vietnamese to SpeedSMS",
			environment:          "DEV",
			tenantName:           constants.TenantGenetica,
			receiver:             "+84344381024",
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSMS,
		},
		{
			name:                 "Environment with whitespace still routes correctly",
			environment:          " STAGING ",
			tenantName:           constants.TenantLifeAI,
			receiver:             "+84344381024",
			channel:              constants.ChannelSMS,
			expectedActualChannel: constants.ChannelSpeedSMS,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			originalEnv := conf.GetEnvironment()
			defer func() {
				conf.SetEnvironmentForTesting(originalEnv)
			}()

			conf.SetEnvironmentForTesting(tc.environment)

			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockQueue := mock_otpqueue.NewMockOTPQueueRepository(ctrl)
			mockSMSProvider := mock_services.NewMockSMSProvider(ctrl)

			inMemCache := caching.NewCachingRepository(
				context.Background(),
				caching.NewGoCacheClient(cache.New(5*time.Minute, 10*time.Minute)),
			)

			courierUseCase := NewCourierUseCase(mockQueue, mockSMSProvider, inMemCache)

			// Execute
			err := courierUseCase.ChooseChannel(ctx, tc.tenantName, tc.receiver, tc.channel)
			require.Nil(t, err, "Failed to choose channel")

			// Verify
			channelResponse, getErr := courierUseCase.GetChannel(ctx, tc.tenantName, tc.receiver)
			require.Nil(t, getErr, "Failed to get channel")
			require.Equal(t, tc.expectedActualChannel, channelResponse.Channel,
				"Environment: %s, Tenant: %s, Expected: %s, Got: %s",
				tc.environment, tc.tenantName, tc.expectedActualChannel, channelResponse.Channel)
		})
	}
}
