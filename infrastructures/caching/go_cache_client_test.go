package caching_test

import (
	"context"
	"testing"
	"time"

	"github.com/lifenetwork-ai/iam-service/infrastructures/caching"
	"github.com/lifenetwork-ai/iam-service/internal/wire/instances"
	"github.com/stretchr/testify/require"
)

func TestNewGoCacheClient(t *testing.T) {
	t.Run("NewGoCacheClient", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		require.NotNil(t, client)
	})
}

func TestGoCacheClient_Set(t *testing.T) {
	t.Run("GoCacheClient_Set", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "GoCacheClient_Set_Key"
		value := "GoCacheClient_Set_Value"
		expiration := 5 * time.Minute

		setErr := client.Set(ctx, key, value, expiration)
		require.NoError(t, setErr)

		dest := ""
		getErr := client.Get(ctx, key, &dest)
		require.NoError(t, getErr)
		require.Equal(t, value, dest)
	})
}

func TestGoCacheClient_Get(t *testing.T) {
	t.Run("GoCacheClient_Get", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "GoCacheClient_Get_Key"
		value := "GoCacheClient_Get_Value"
		expiration := 5 * time.Minute

		setErr := client.Set(ctx, key, value, expiration)
		require.NoError(t, setErr)

		dest := ""
		getErr := client.Get(ctx, key, &dest)
		require.NoError(t, getErr)
		require.Equal(t, value, dest)
	})

	t.Run("GoCacheClient_Get_ItemNotFound", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "nonExistentKey"

		dest := ""
		getErr := client.Get(ctx, key, &dest)
		require.Error(t, getErr)
		require.Equal(t, "item not found in cache", getErr.Error())
	})

	t.Run("GoCacheClient_Get_InvalidDestination", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "GoCacheClient_Get_InvalidDestination_Key"
		value := "GoCacheClient_Get_InvalidDestination_Value"
		expiration := 5 * time.Minute

		setErr := client.Set(ctx, key, value, expiration)
		require.NoError(t, setErr)

		dest := 0
		getErr := client.Get(ctx, key, &dest)
		require.Error(t, getErr)
		require.Equal(t, "cached value type (string) does not match destination type (int)", getErr.Error())
	})

	t.Run("GoCacheClient_Get_NilDestination", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "GoCacheClient_Get_NilDestination_Key"
		value := "GoCacheClient_Get_NilDestination_Value"
		expiration := 5 * time.Minute

		setErr := client.Set(ctx, key, value, expiration)
		require.NoError(t, setErr)

		var dest *string
		getErr := client.Get(ctx, key, dest)
		require.Error(t, getErr)
		require.Equal(t, "destination must be a non-nil pointer", getErr.Error())
	})
}

func TestGoCacheClient_Del(t *testing.T) {
	t.Run("GoCacheClient_Del", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "GoCacheClient_Del_Key"
		value := "GoCacheClient_Del_Value"
		expiration := 5 * time.Minute

		setErr := client.Set(ctx, key, value, expiration)
		require.NoError(t, setErr)

		delErr := client.Del(ctx, key)
		require.NoError(t, delErr)

		dest := ""
		getErr := client.Get(ctx, key, &dest)
		require.Equal(t, "item not found in cache", getErr.Error())
		require.Equal(t, "", dest)
	})
}

func TestGoCacheClient_CacheMapValue(t *testing.T) {
	client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
	ctx := context.Background()
	key := "GoCacheClient_Map_Key"

	cacheValue := map[string]string{
		"type":  "email",
		"email": "user@example.com",
		"otp":   "123456",
	}

	// Set
	err := client.Set(ctx, key, cacheValue, 5*time.Minute)
	require.NoError(t, err)

	// Get
	var result map[string]string
	err = client.Get(ctx, key, &result)
	require.NoError(t, err)
	require.Equal(t, cacheValue, result)
}

func TestGoCacheClient_Expiration(t *testing.T) {
	client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
	ctx := context.Background()
	key := "GoCacheClient_Expiration_Key"
	value := "temp value"

	setErr := client.Set(ctx, key, value, 1*time.Second)
	require.NoError(t, setErr)

	time.Sleep(2 * time.Second)

	dest := ""
	getErr := client.Get(ctx, key, &dest)
	require.Error(t, getErr)
	require.Equal(t, "item not found in cache", getErr.Error())
}

func TestGoCacheClient_Overwrite(t *testing.T) {
	client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
	ctx := context.Background()
	key := "GoCacheClient_Overwrite_Key"
	value1 := "first"
	value2 := "second"

	_ = client.Set(ctx, key, value1, 5*time.Minute)
	_ = client.Set(ctx, key, value2, 5*time.Minute)

	dest := ""
	getErr := client.Get(ctx, key, &dest)
	require.NoError(t, getErr)
	require.Equal(t, value2, dest)
}

type TestUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func TestGoCacheClient_TypeConversions(t *testing.T) {
	t.Run("StructValue_To_PointerStruct", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "struct_to_pointer"

		// Store struct value
		originalUser := TestUser{
			ID:    1,
			Name:  "John Doe",
			Email: "john@example.com",
		}

		err := client.Set(ctx, key, originalUser, 5*time.Minute)
		require.NoError(t, err)

		// Retrieve into pointer
		var userPtr *TestUser
		err = client.Get(ctx, key, &userPtr)
		require.NoError(t, err)
		require.NotNil(t, userPtr)
		require.Equal(t, originalUser.ID, userPtr.ID)
		require.Equal(t, originalUser.Name, userPtr.Name)
		require.Equal(t, originalUser.Email, userPtr.Email)
	})

	t.Run("PointerStruct_To_StructValue", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "pointer_to_struct"

		// Store pointer
		originalUser := &TestUser{
			ID:    2,
			Name:  "Jane Doe",
			Email: "jane@example.com",
		}

		err := client.Set(ctx, key, originalUser, 5*time.Minute)
		require.NoError(t, err)

		// Retrieve into struct value
		var user TestUser
		err = client.Get(ctx, key, &user)
		require.NoError(t, err)
		require.Equal(t, originalUser.ID, user.ID)
		require.Equal(t, originalUser.Name, user.Name)
		require.Equal(t, originalUser.Email, user.Email)
	})

	t.Run("PointerStruct_To_PointerStruct", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "pointer_to_pointer"

		// Store pointer
		originalUser := &TestUser{
			ID:    3,
			Name:  "Bob Smith",
			Email: "bob@example.com",
		}

		err := client.Set(ctx, key, originalUser, 5*time.Minute)
		require.NoError(t, err)

		// Retrieve into pointer
		var userPtr *TestUser
		err = client.Get(ctx, key, &userPtr)
		require.NoError(t, err)
		require.NotNil(t, userPtr)
		require.Equal(t, originalUser.ID, userPtr.ID)
		require.Equal(t, originalUser.Name, userPtr.Name)
		require.Equal(t, originalUser.Email, userPtr.Email)
	})

	t.Run("StructValue_To_StructValue", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "struct_to_struct"

		// Store struct value
		originalUser := TestUser{
			ID:    4,
			Name:  "Alice Johnson",
			Email: "alice@example.com",
		}

		err := client.Set(ctx, key, originalUser, 5*time.Minute)
		require.NoError(t, err)

		// Retrieve into struct value
		var user TestUser
		err = client.Get(ctx, key, &user)
		require.NoError(t, err)
		require.Equal(t, originalUser.ID, user.ID)
		require.Equal(t, originalUser.Name, user.Name)
		require.Equal(t, originalUser.Email, user.Email)
	})
}

func TestGoCacheClient_TypeConversion_EdgeCases(t *testing.T) {
	t.Run("NilPointer_Storage", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "nil_pointer"

		// Store nil pointer
		var nilUser *TestUser
		err := client.Set(ctx, key, nilUser, 5*time.Minute)
		require.NoError(t, err)

		// Retrieve into pointer
		var userPtr *TestUser
		err = client.Get(ctx, key, &userPtr)
		require.NoError(t, err)
		require.Nil(t, userPtr)
	})

	t.Run("IncompatibleTypes_ShouldFail", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "incompatible_types"

		// Store string
		err := client.Set(ctx, key, "string_value", 5*time.Minute)
		require.NoError(t, err)

		// Try to retrieve into struct pointer (should fail)
		var userPtr *TestUser
		err = client.Get(ctx, key, &userPtr)
		require.Error(t, err)
		require.Contains(t, err.Error(), "does not match destination type")
	})

	t.Run("DifferentStructTypes_ShouldFail", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "different_struct_types"

		type DifferentStruct struct {
			Field1 string
			Field2 int
		}

		// Store one struct type
		original := TestUser{ID: 1, Name: "Test", Email: "test@example.com"}
		err := client.Set(ctx, key, original, 5*time.Minute)
		require.NoError(t, err)

		// Try to retrieve into different struct type (should fail)
		var different DifferentStruct
		err = client.Get(ctx, key, &different)
		require.Error(t, err)
		require.Contains(t, err.Error(), "does not match destination type")
	})
}

func TestGoCacheClient_PrimitiveTypeConversions(t *testing.T) {
	t.Run("IntValue_To_IntPointer", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "int_to_int_pointer"

		// Store int value
		originalValue := 42
		err := client.Set(ctx, key, originalValue, 5*time.Minute)
		require.NoError(t, err)

		// Retrieve into int pointer
		var intPtr *int
		err = client.Get(ctx, key, &intPtr)
		require.NoError(t, err)
		require.NotNil(t, intPtr)
		require.Equal(t, originalValue, *intPtr)
	})

	t.Run("StringPointer_To_StringValue", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "string_pointer_to_string"

		// Store string pointer
		originalValue := "hello world"
		err := client.Set(ctx, key, &originalValue, 5*time.Minute)
		require.NoError(t, err)

		// Retrieve into string value
		var str string
		err = client.Get(ctx, key, &str)
		require.NoError(t, err)
		require.Equal(t, originalValue, str)
	})
}

func TestGoCacheClient_SliceAndMapConversions(t *testing.T) {
	t.Run("SliceValue_To_SlicePointer", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "slice_to_slice_pointer"

		// Store slice value
		originalSlice := []string{"a", "b", "c"}
		err := client.Set(ctx, key, originalSlice, 5*time.Minute)
		require.NoError(t, err)

		// Retrieve into slice pointer
		var slicePtr *[]string
		err = client.Get(ctx, key, &slicePtr)
		require.NoError(t, err)
		require.NotNil(t, slicePtr)
		require.Equal(t, originalSlice, *slicePtr)
	})

	t.Run("MapPointer_To_MapValue", func(t *testing.T) {
		client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
		ctx := context.Background()
		key := "map_pointer_to_map"

		// Store map pointer
		originalMap := map[string]int{"a": 1, "b": 2, "c": 3}
		err := client.Set(ctx, key, &originalMap, 5*time.Minute)
		require.NoError(t, err)

		// Retrieve into map value
		var mapValue map[string]int
		err = client.Get(ctx, key, &mapValue)
		require.NoError(t, err)
		require.Equal(t, originalMap, mapValue)
	})
}

// Benchmark tests for type conversion performance
func BenchmarkGoCacheClient_TypeConversions(b *testing.B) {
	client := caching.NewGoCacheClient(instances.GoCacheClientInstance())
	ctx := context.Background()

	user := TestUser{
		ID:    1,
		Name:  "Benchmark User",
		Email: "benchmark@example.com",
	}

	b.Run("StructValue_To_PointerStruct", func(b *testing.B) {
		key := "benchmark_struct_to_pointer"
		_ = client.Set(ctx, key, user, 5*time.Minute)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var userPtr *TestUser
			_ = client.Get(ctx, key, &userPtr)
		}
	})

	b.Run("PointerStruct_To_StructValue", func(b *testing.B) {
		key := "benchmark_pointer_to_struct"
		_ = client.Set(ctx, key, &user, 5*time.Minute)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var userValue TestUser
			_ = client.Get(ctx, key, &userValue)
		}
	})
}
