package caching

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/infrastructures/mocks"
	"github.com/stretchr/testify/require"
)

func TestSaveItem(t *testing.T) {
	t.Run("TestCachingRepository_SaveItem", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCacheClient := mocks.NewMockCacheClient(ctrl)
		ctx := context.Background()
		repo := NewCachingRepository(ctx, mockCacheClient)

		// Create the mock Stringer properly
		mockKey := mocks.NewMockStringer(ctrl)
		mockKey.EXPECT().String().Return("testKey").AnyTimes()

		value := "testValue"
		expire := 5 * time.Minute

		// Account for the prefixed key that the repository creates
		expectedKey := fmt.Sprintf("%s_testKey", conf.GetAppName())
		mockCacheClient.EXPECT().Set(ctx, expectedKey, value, expire).Return(nil)

		err := repo.SaveItem(mockKey, value, expire)
		require.NoError(t, err)
	})
}

func TestRetrieveItem(t *testing.T) {
	t.Run("TestCachingRepository_RetrieveItem", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockCacheClient(ctrl)
		ctx := context.Background()
		repo := NewCachingRepository(ctx, mockClient)

		// Create the mock Stringer properly
		mockKey := mocks.NewMockStringer(ctrl)
		mockKey.EXPECT().String().Return("testKey").AnyTimes()

		var value string

		// Account for the prefixed key that the repository creates
		expectedKey := fmt.Sprintf("%s_testKey", conf.GetAppName())
		mockClient.EXPECT().Get(ctx, expectedKey, &value).Return(nil)

		err := repo.RetrieveItem(mockKey, &value)
		require.NoError(t, err)
	})
}

func TestRemoveItem(t *testing.T) {
	t.Run("TestCachingRepository_RemoveItem", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockCacheClient(ctrl)
		ctx := context.Background()
		repo := NewCachingRepository(ctx, mockClient)

		// Create the mock Stringer properly
		mockKey := mocks.NewMockStringer(ctrl)
		mockKey.EXPECT().String().Return("testKey").AnyTimes()

		// Account for the prefixed key that the repository creates
		expectedKey := fmt.Sprintf("%s_testKey", conf.GetAppName())
		mockClient.EXPECT().Del(ctx, expectedKey).Return(nil)

		err := repo.RemoveItem(mockKey)
		require.NoError(t, err)
	})
}
