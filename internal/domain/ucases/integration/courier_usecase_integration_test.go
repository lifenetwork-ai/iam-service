//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/infrastructures/caching"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases"
	mock_services "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/services"
	mock_otpqueue "github.com/lifenetwork-ai/iam-service/mocks/infrastructures/otp_queue/types"
	"github.com/patrickmn/go-cache"
)

func TestCourierUseCase_ChooseChannel_PhoneNumberValidation_Integration(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup test dependencies
	mockQueue := mock_otpqueue.NewMockOTPQueueRepository(ctrl)
	mockSMSProvider := mock_services.NewMockSMSProvider(ctrl)

	// Use real cache for integration testing
	inMemCache := caching.NewCachingRepository(
		context.Background(),
		caching.NewGoCacheClient(cache.New(5*time.Minute, 10*time.Minute)),
	)

	courierUseCase := ucases.NewCourierUseCase(mockQueue, mockSMSProvider, inMemCache)

	testCases := []struct {
		name            string
		environment     string
		tenantName      string
		receiver        string
		inputChannel    string
		expectedChannel string
		expectError     bool
		expectedErrMsg  string
		expectedErrCode string
	}{
		// Valid phone number cases
		{
			name:            "Valid E164 phone number with SMS channel - Nightly",
			environment:     constants.NightlyEnvironment,
			tenantName:      constants.TenantLifeAI,
			receiver:        "+84344381024",
			inputChannel:    constants.ChannelSMS,
			expectedChannel: constants.ChannelWebhook,
			expectError:     false,
		},
		{
			name:            "Valid Vietnamese phone number with SMS channel - Staging",
			environment:     constants.StagingEnvironment,
			tenantName:      constants.TenantLifeAI,
			receiver:        "+84344381024",
			inputChannel:    constants.ChannelSMS,
			expectedChannel: constants.ChannelSpeedSMS,
			expectError:     false,
		},
		{
			name:            "Valid Thailand phone number with SMS channel - Production",
			environment:     constants.ProductionEnvironment,
			tenantName:      constants.TenantLifeAI,
			receiver:        "+66812345678",
			inputChannel:    constants.ChannelSMS,
			expectedChannel: constants.ChannelSMS,
			expectError:     false,
		},
		{
			name:            "Valid E164 phone number with WhatsApp channel",
			tenantName:      constants.TenantLifeAI,
			receiver:        "+84344381024",
			inputChannel:    constants.ChannelWhatsApp,
			expectedChannel: constants.ChannelWhatsApp,
			expectError:     false,
		},
		{
			name:            "Valid E164 phone number with Zalo channel for Genetica",
			tenantName:      constants.TenantGenetica,
			receiver:        "+84344381024",
			inputChannel:    constants.ChannelZalo,
			expectedChannel: constants.ChannelZalo,
			expectError:     false,
		},
		{
			name:            "Valid Vietnam phone number with SMS channel - Nightly",
			environment:     constants.NightlyEnvironment,
			tenantName:      constants.TenantLifeAI,
			receiver:        "+84987654321",
			inputChannel:    constants.ChannelSMS,
			expectedChannel: constants.ChannelWebhook,
			expectError:     false,
		},

		// Invalid receiver cases - should only accept phone numbers
		{
			name:            "Email address should be rejected",
			tenantName:      constants.TenantLifeAI,
			receiver:        "user@example.com",
			inputChannel:    constants.ChannelWebhook,
			expectError:     true,
			expectedErrMsg:  "Invalid phone number",
			expectedErrCode: "MSG_INVALID_RECEIVER",
		},
		{
			name:            "Username should be rejected",
			tenantName:      constants.TenantLifeAI,
			receiver:        "username123",
			inputChannel:    constants.ChannelWebhook,
			expectError:     true,
			expectedErrMsg:  "Invalid phone number",
			expectedErrCode: "MSG_INVALID_RECEIVER",
		},
		{
			name:            "Empty receiver should be rejected",
			tenantName:      constants.TenantLifeAI,
			receiver:        "",
			inputChannel:    constants.ChannelWebhook,
			expectError:     true,
			expectedErrMsg:  "Invalid phone number",
			expectedErrCode: "MSG_INVALID_RECEIVER",
		},
		{
			name:            "Invalid phone number format should be rejected",
			tenantName:      constants.TenantLifeAI,
			receiver:        "123456",
			inputChannel:    constants.ChannelWebhook,
			expectError:     true,
			expectedErrMsg:  "Invalid phone number",
			expectedErrCode: "MSG_INVALID_RECEIVER",
		},
		{
			name:            "Phone number without country code should be rejected",
			tenantName:      constants.TenantLifeAI,
			receiver:        "0344381024",
			inputChannel:    constants.ChannelWebhook,
			expectError:     true,
			expectedErrMsg:  "Invalid phone number",
			expectedErrCode: "MSG_INVALID_RECEIVER",
		},
		{
			name:            "Invalid E164 format should be rejected",
			tenantName:      constants.TenantLifeAI,
			receiver:        "+84-344-381-024",
			inputChannel:    constants.ChannelWebhook,
			expectError:     true,
			expectedErrMsg:  "Invalid phone number",
			expectedErrCode: "MSG_INVALID_RECEIVER",
		},

		// Invalid channel cases
		{
			name:            "Empty channel should be rejected",
			tenantName:      constants.TenantLifeAI,
			receiver:        "+84344381024",
			inputChannel:    "",
			expectError:     true,
			expectedErrMsg:  "Channel is required",
			expectedErrCode: "MSG_INVALID_CHANNEL",
		},
		{
			name:            "Invalid channel should be rejected",
			tenantName:      constants.TenantLifeAI,
			receiver:        "+84344381024",
			inputChannel:    "invalid_channel",
			expectError:     true,
			expectedErrMsg:  "Channel not supported",
			expectedErrCode: "MSG_CHANNEL_NOT_SUPPORTED",
		},
		{
			name:            "Zalo channel not supported for LifeAI tenant",
			tenantName:      constants.TenantLifeAI,
			receiver:        "+84344381024",
			inputChannel:    constants.ChannelZalo,
			expectError:     true,
			expectedErrMsg:  "Channel not supported",
			expectedErrCode: "MSG_CHANNEL_NOT_SUPPORTED",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				conf.SetEnvironmentForTesting(constants.NightlyEnvironment)
			}()
			conf.SetEnvironmentForTesting(tc.environment)
			err := courierUseCase.ChooseChannel(ctx, tc.tenantName, tc.receiver, tc.inputChannel)

			if tc.expectError {
				require.NotNil(t, err, "Expected error but got nil")
				require.Equal(t, tc.expectedErrCode, err.Code, "Error code mismatch")
				require.Contains(t, err.Message, tc.expectedErrMsg, "Error message mismatch")
			} else {
				require.Nil(t, err, "Expected no error but got: %v", err)

				// Verify that the channel was saved correctly
				channelResponse, getErr := courierUseCase.GetChannel(ctx, tc.tenantName, tc.receiver)
				require.Nil(t, getErr, "Failed to get saved channel")
				require.Equal(t, tc.expectedChannel, channelResponse.Channel, "Saved channel mismatch")
			}
		})
	}
}

func TestCourierUseCase_GetAvailableChannels_Integration(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup test dependencies
	mockQueue := mock_otpqueue.NewMockOTPQueueRepository(ctrl)
	mockSMSProvider := mock_services.NewMockSMSProvider(ctrl)

	// Use real cache for integration testing
	inMemCache := caching.NewCachingRepository(
		context.Background(),
		caching.NewGoCacheClient(cache.New(5*time.Minute, 10*time.Minute)),
	)

	courierUseCase := ucases.NewCourierUseCase(mockQueue, mockSMSProvider, inMemCache)

	testCases := []struct {
		name             string
		tenantName       string
		receiver         string
		expectedChannels []string
	}{
		{
			name:             "LifeAI tenant should support SMS and WhatsApp",
			tenantName:       constants.TenantLifeAI,
			receiver:         "+84344381024",
			expectedChannels: []string{constants.ChannelWebhook, constants.ChannelWhatsApp},
		},
		{
			name:             "Genetica tenant should support SMS and Zalo",
			tenantName:       constants.TenantGenetica,
			receiver:         "+84344381024",
			expectedChannels: []string{constants.ChannelWebhook, constants.ChannelZalo},
		},
		{
			name:             "Unknown tenant should support all channels",
			tenantName:       "unknown_tenant",
			receiver:         "+84344381024",
			expectedChannels: []string{constants.ChannelWebhook, constants.ChannelWhatsApp, constants.ChannelZalo},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			channels := courierUseCase.GetAvailableChannels(ctx, tc.tenantName, tc.receiver)
			require.ElementsMatch(t, tc.expectedChannels, channels, "Available channels mismatch")
		})
	}
}

func TestCourierUseCase_ChooseChannel_CacheIntegration(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup test dependencies
	mockQueue := mock_otpqueue.NewMockOTPQueueRepository(ctrl)
	mockSMSProvider := mock_services.NewMockSMSProvider(ctrl)

	// Use real cache for integration testing
	inMemCache := caching.NewCachingRepository(
		context.Background(),
		caching.NewGoCacheClient(cache.New(5*time.Minute, 10*time.Minute)),
	)

	courierUseCase := ucases.NewCourierUseCase(mockQueue, mockSMSProvider, inMemCache)

	tenantName := constants.TenantLifeAI
	receiver := "+84344381024"
	channel := constants.ChannelWebhook

	// Test choosing a channel
	err := courierUseCase.ChooseChannel(ctx, tenantName, receiver, channel)
	require.Nil(t, err, "Failed to choose channel")

	// Test retrieving the chosen channel
	channelResponse, getErr := courierUseCase.GetChannel(ctx, tenantName, receiver)
	require.Nil(t, getErr, "Failed to get channel")
	require.Equal(t, channel, channelResponse.Channel, "Retrieved channel mismatch")

	// Test overwriting with a different channel
	newChannel := constants.ChannelWhatsApp
	err = courierUseCase.ChooseChannel(ctx, tenantName, receiver, newChannel)
	require.Nil(t, err, "Failed to choose new channel")

	// Verify the channel was updated
	channelResponse, getErr = courierUseCase.GetChannel(ctx, tenantName, receiver)
	require.Nil(t, getErr, "Failed to get updated channel")
	require.Equal(t, newChannel, channelResponse.Channel, "Updated channel mismatch")

	// Test with different receiver (should be independent)
	differentReceiver := "+84987654321"
	differentChannel := constants.ChannelWebhook
	err = courierUseCase.ChooseChannel(ctx, tenantName, differentReceiver, differentChannel)
	require.Nil(t, err, "Failed to choose channel for different receiver")

	// Verify both receivers have their own channels
	channelResponse1, getErr1 := courierUseCase.GetChannel(ctx, tenantName, receiver)
	require.Nil(t, getErr1, "Failed to get channel for first receiver")
	require.Equal(t, newChannel, channelResponse1.Channel, "First receiver channel mismatch")

	channelResponse2, getErr2 := courierUseCase.GetChannel(ctx, tenantName, differentReceiver)
	require.Nil(t, getErr2, "Failed to get channel for second receiver")
	require.Equal(t, differentChannel, channelResponse2.Channel, "Second receiver channel mismatch")
}

func TestCourierUseCase_GetChannel_CacheMiss_Integration(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup test dependencies
	mockQueue := mock_otpqueue.NewMockOTPQueueRepository(ctrl)
	mockSMSProvider := mock_services.NewMockSMSProvider(ctrl)

	// Use real cache for integration testing
	inMemCache := caching.NewCachingRepository(
		context.Background(),
		caching.NewGoCacheClient(cache.New(5*time.Minute, 10*time.Minute)),
	)

	courierUseCase := ucases.NewCourierUseCase(mockQueue, mockSMSProvider, inMemCache)

	tenantName := constants.TenantLifeAI
	receiver := "+84344381024"

	// Test getting channel when no channel has been chosen (cache miss)
	channelResponse, getErr := courierUseCase.GetChannel(ctx, tenantName, receiver)
	require.Nil(t, getErr, "Expected no error on cache miss")
	require.Equal(t, constants.ChannelWebhook, channelResponse.Channel, "Should fallback to SMS on cache miss")
}
