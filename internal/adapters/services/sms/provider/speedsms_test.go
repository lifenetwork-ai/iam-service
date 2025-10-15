package provider

import (
	"testing"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSpeedSMSProvider(t *testing.T) {
	config := conf.SpeedSMSConfiguration{
		GeneticaSpeedSMSAccessToken: "genetica-token-123",
		LifeSpeedSMSAccessToken:     "life-token-456",
		SpeedSMSBaseURL:             "https://api.speedsms.vn/index.php",
	}

	provider := NewSpeedSMSProvider(config)
	require.NotNil(t, provider)

	speedSMSProvider, ok := provider.(*SpeedSMSProvider)
	require.True(t, ok, "provider should be of type *SpeedSMSProvider")

	// Verify both clients are initialized
	assert.NotNil(t, speedSMSProvider.geneticaClient, "genetica client should be initialized")
	assert.NotNil(t, speedSMSProvider.lifeClient, "life client should be initialized")

	// Verify access tokens are set correctly
	assert.Equal(t, "genetica-token-123", speedSMSProvider.geneticaClient.AccessToken)
	assert.Equal(t, "life-token-456", speedSMSProvider.lifeClient.AccessToken)
}

func TestSpeedSMSProvider_GetChannelType(t *testing.T) {
	config := conf.SpeedSMSConfiguration{
		GeneticaSpeedSMSAccessToken: "genetica-token",
		LifeSpeedSMSAccessToken:     "life-token",
	}

	provider := NewSpeedSMSProvider(config)
	assert.Equal(t, constants.ChannelSpeedSMS, provider.GetChannelType())
}
