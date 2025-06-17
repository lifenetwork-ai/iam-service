package caching

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestSuite for Redis cache client tests using testcontainers
type RedisCacheTestSuite struct {
	suite.Suite
	client         *redisCacheClient
	ctx            context.Context
	redisContainer testcontainers.Container
	redisHost      string
	redisPort      string
}

func (suite *RedisCacheTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Start Redis container
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.T(), err)

	suite.redisContainer = redisContainer

	// Get the mapped port
	mappedPort, err := redisContainer.MappedPort(suite.ctx, "6379")
	require.NoError(suite.T(), err)

	hostIP, err := redisContainer.Host(suite.ctx)
	require.NoError(suite.T(), err)

	suite.redisHost = hostIP
	suite.redisPort = mappedPort.Port()

	// Override the Redis configuration with container details
	redisConfiguration := conf.GetRedisConfiguration()
	redisConfiguration.RedisAddress = fmt.Sprintf("%s:%s", suite.redisHost, suite.redisPort)
	redisConfiguration.RedisTtl = "5m"

	// Initialize the Redis cache client
	suite.client = NewRedisCacheClient().(*redisCacheClient)
}

func (suite *RedisCacheTestSuite) TearDownSuite() {
	if suite.redisContainer != nil {
		_ = suite.redisContainer.Terminate(suite.ctx)
	}
}

func (suite *RedisCacheTestSuite) SetupTest() {
	// Clean up any existing test keys before each test
	suite.cleanupTestKeys()
}

func (suite *RedisCacheTestSuite) TearDownTest() {
	// Clean up test keys after each test
	suite.cleanupTestKeys()
}

func (suite *RedisCacheTestSuite) cleanupTestKeys() {
	testKeys := []string{
		"test_string_key",
		"test_int_key",
		"test_bool_key",
		"test_struct_key",
		"test_delete_key",
		"test_expiration_key",
		"test_overwrite_key",
		"non_existent_key",
	}

	for _, key := range testKeys {
		_ = suite.client.Del(suite.ctx, key)
	}
}

func (suite *RedisCacheTestSuite) TestSetAndGet() {
	tests := []struct {
		name        string
		key         string
		value       interface{}
		dest        interface{}
		expiration  time.Duration
		expectError bool
	}{
		{
			name:        "String_Value",
			key:         "test_string_key",
			value:       "test_value",
			dest:        new(string),
			expiration:  5 * time.Minute,
			expectError: false,
		},
		{
			name:        "Integer_Value",
			key:         "test_int_key",
			value:       42,
			dest:        new(int),
			expiration:  5 * time.Minute,
			expectError: false,
		},
		{
			name:        "Boolean_Value",
			key:         "test_bool_key",
			value:       true,
			dest:        new(bool),
			expiration:  5 * time.Minute,
			expectError: false,
		},
		{
			name:        "Struct_Value",
			key:         "test_struct_key",
			value:       struct{ Name string }{"Alice"},
			dest:        new(struct{ Name string }),
			expiration:  5 * time.Minute,
			expectError: false,
		},
		{
			name:        "Missing_Key",
			key:         "non_existent_key",
			value:       nil,
			dest:        new(string),
			expiration:  0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.value != nil {
				err := suite.client.Set(suite.ctx, tt.key, tt.value, tt.expiration)
				require.NoError(suite.T(), err)
			}

			err := suite.client.Get(suite.ctx, tt.key, tt.dest)

			if tt.expectError {
				require.Error(suite.T(), err)
				return
			}

			require.NoError(suite.T(), err)

			// Compare values correctly
			switch v := tt.dest.(type) {
			case *int:
				require.Equal(suite.T(), tt.value.(int), *v)
			case *string:
				require.Equal(suite.T(), tt.value.(string), *v)
			case *bool:
				require.Equal(suite.T(), tt.value.(bool), *v)
			case *struct{ Name string }:
				require.Equal(suite.T(), tt.value.(struct{ Name string }).Name, v.Name)
			default:
				suite.T().Fatalf("Unhandled type for test case: %s", tt.name)
			}
		})
	}
}

func (suite *RedisCacheTestSuite) TestDelete() {
	suite.Run("Delete_Existing_Key", func() {
		key := "test_delete_key"
		value := "delete_me"
		expiration := 5 * time.Minute

		// Set a value
		err := suite.client.Set(suite.ctx, key, value, expiration)
		require.NoError(suite.T(), err)

		// Verify it exists
		var retrieved string
		err = suite.client.Get(suite.ctx, key, &retrieved)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), value, retrieved)

		// Delete the value
		err = suite.client.Del(suite.ctx, key)
		require.NoError(suite.T(), err)

		// Ensure it's deleted
		var dest string
		err = suite.client.Get(suite.ctx, key, &dest)
		require.Error(suite.T(), err)
		require.Contains(suite.T(), err.Error(), "cache miss")
	})

	suite.Run("Delete_Non_Existing_Key", func() {
		key := "non_existent_key"
		err := suite.client.Del(suite.ctx, key)
		require.NoError(suite.T(), err) // Should not error on non-existent key
	})
}

func (suite *RedisCacheTestSuite) TestExpiration() {
	suite.Run("Key_Expires_After_TTL", func() {
		key := "test_expiration_key"
		value := "expires_soon"
		expiration := 1 * time.Second

		// Set with short expiration
		err := suite.client.Set(suite.ctx, key, value, expiration)
		require.NoError(suite.T(), err)

		// Verify it exists immediately
		var retrieved string
		err = suite.client.Get(suite.ctx, key, &retrieved)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), value, retrieved)

		// Wait for expiration
		time.Sleep(2 * time.Second)

		// Verify it's expired
		err = suite.client.Get(suite.ctx, key, &retrieved)
		require.Error(suite.T(), err)
		require.Contains(suite.T(), err.Error(), "cache miss")
	})
}

func (suite *RedisCacheTestSuite) TestOverwrite() {
	suite.Run("Overwrite_Existing_Key", func() {
		key := "test_overwrite_key"
		originalValue := "original"
		newValue := "updated"
		expiration := 5 * time.Minute

		// Set original value
		err := suite.client.Set(suite.ctx, key, originalValue, expiration)
		require.NoError(suite.T(), err)

		// Verify original value
		var retrieved string
		err = suite.client.Get(suite.ctx, key, &retrieved)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), originalValue, retrieved)

		// Overwrite with new value
		err = suite.client.Set(suite.ctx, key, newValue, expiration)
		require.NoError(suite.T(), err)

		// Verify new value
		err = suite.client.Get(suite.ctx, key, &retrieved)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), newValue, retrieved)
	})
}

func (suite *RedisCacheTestSuite) TestContextCancellation() {
	suite.Run("Context_Cancellation_Handling", func() {
		key := "test_context_key"
		value := "test_value"
		expiration := 5 * time.Minute

		// Create a context that will be cancelled
		cancelCtx, cancel := context.WithCancel(suite.ctx)
		cancel() // Cancel immediately

		// Operations should handle cancelled context gracefully
		err := suite.client.Set(cancelCtx, key, value, expiration)
		require.Error(suite.T(), err)
		require.Contains(suite.T(), err.Error(), "context")

		var dest string
		err = suite.client.Get(cancelCtx, key, &dest)
		require.Error(suite.T(), err)
		require.Contains(suite.T(), err.Error(), "context")
	})
}

func (suite *RedisCacheTestSuite) TestNilValues() {
	suite.Run("Handle_Nil_Destination", func() {
		key := "test_nil_key"
		value := "test_value"
		expiration := 5 * time.Minute

		// Set value
		err := suite.client.Set(suite.ctx, key, value, expiration)
		require.NoError(suite.T(), err)

		// Try to get with nil destination - should handle gracefully
		err = suite.client.Get(suite.ctx, key, nil)
		require.Error(suite.T(), err) // Should error appropriately for nil destination
	})
}

// Run the test suite
func TestRedisCacheTestSuite(t *testing.T) {
	suite.Run(t, new(RedisCacheTestSuite))
}

// Alternative: Individual test functions if you prefer not to use testify/suite
func TestRedisCacheWithContainer(t *testing.T) {
	ctx := context.Background()

	// Start Redis container
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	// Get connection details
	mappedPort, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	hostIP, err := redisContainer.Host(ctx)
	require.NoError(t, err)

	// Configure Redis client
	redisConfiguration := conf.GetRedisConfiguration()
	redisConfiguration.RedisAddress = fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())
	redisConfiguration.RedisTtl = "5m"

	client := NewRedisCacheClient().(*redisCacheClient)

	// Basic functionality test
	t.Run("Basic_Set_And_Get", func(t *testing.T) {
		key := "basic_test_key"
		value := "hello_world"
		expiration := 5 * time.Minute

		err := client.Set(ctx, key, value, expiration)
		require.NoError(t, err)

		var retrieved string
		err = client.Get(ctx, key, &retrieved)
		require.NoError(t, err)
		require.Equal(t, value, retrieved)

		// Clean up
		_ = client.Del(ctx, key)
	})
}
