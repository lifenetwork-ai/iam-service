// Code generated by MockGen. DO NOT EDIT.
// Source: ./infrastructures/interfaces/caching.go
//
// Generated by this command:
//
//	mockgen -source=./infrastructures/interfaces/caching.go -package=mock_interfaces -destination=mocks/infrastructures/interfaces/mock_caching.go
//

// Package mock_interfaces is a generated GoMock package.
package mock_interfaces

import (
	context "context"
	fmt "fmt"
	reflect "reflect"
	time "time"

	gomock "go.uber.org/mock/gomock"
)

// MockCacheClient is a mock of CacheClient interface.
type MockCacheClient struct {
	ctrl     *gomock.Controller
	recorder *MockCacheClientMockRecorder
	isgomock struct{}
}

// MockCacheClientMockRecorder is the mock recorder for MockCacheClient.
type MockCacheClientMockRecorder struct {
	mock *MockCacheClient
}

// NewMockCacheClient creates a new mock instance.
func NewMockCacheClient(ctrl *gomock.Controller) *MockCacheClient {
	mock := &MockCacheClient{ctrl: ctrl}
	mock.recorder = &MockCacheClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCacheClient) EXPECT() *MockCacheClientMockRecorder {
	return m.recorder
}

// Del mocks base method.
func (m *MockCacheClient) Del(ctx context.Context, key string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Del", ctx, key)
	ret0, _ := ret[0].(error)
	return ret0
}

// Del indicates an expected call of Del.
func (mr *MockCacheClientMockRecorder) Del(ctx, key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Del", reflect.TypeOf((*MockCacheClient)(nil).Del), ctx, key)
}

// Get mocks base method.
func (m *MockCacheClient) Get(ctx context.Context, key string, dest any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, key, dest)
	ret0, _ := ret[0].(error)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockCacheClientMockRecorder) Get(ctx, key, dest any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCacheClient)(nil).Get), ctx, key, dest)
}

// Set mocks base method.
func (m *MockCacheClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", ctx, key, value, expiration)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockCacheClientMockRecorder) Set(ctx, key, value, expiration any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockCacheClient)(nil).Set), ctx, key, value, expiration)
}

// MockCacheRepository is a mock of CacheRepository interface.
type MockCacheRepository struct {
	ctrl     *gomock.Controller
	recorder *MockCacheRepositoryMockRecorder
	isgomock struct{}
}

// MockCacheRepositoryMockRecorder is the mock recorder for MockCacheRepository.
type MockCacheRepositoryMockRecorder struct {
	mock *MockCacheRepository
}

// NewMockCacheRepository creates a new mock instance.
func NewMockCacheRepository(ctrl *gomock.Controller) *MockCacheRepository {
	mock := &MockCacheRepository{ctrl: ctrl}
	mock.recorder = &MockCacheRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCacheRepository) EXPECT() *MockCacheRepositoryMockRecorder {
	return m.recorder
}

// RemoveItem mocks base method.
func (m *MockCacheRepository) RemoveItem(key fmt.Stringer) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveItem", key)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveItem indicates an expected call of RemoveItem.
func (mr *MockCacheRepositoryMockRecorder) RemoveItem(key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveItem", reflect.TypeOf((*MockCacheRepository)(nil).RemoveItem), key)
}

// RetrieveItem mocks base method.
func (m *MockCacheRepository) RetrieveItem(key fmt.Stringer, val any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RetrieveItem", key, val)
	ret0, _ := ret[0].(error)
	return ret0
}

// RetrieveItem indicates an expected call of RetrieveItem.
func (mr *MockCacheRepositoryMockRecorder) RetrieveItem(key, val any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RetrieveItem", reflect.TypeOf((*MockCacheRepository)(nil).RetrieveItem), key, val)
}

// SaveItem mocks base method.
func (m *MockCacheRepository) SaveItem(key fmt.Stringer, val any, expire time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveItem", key, val, expire)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveItem indicates an expected call of SaveItem.
func (mr *MockCacheRepositoryMockRecorder) SaveItem(key, val, expire any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveItem", reflect.TypeOf((*MockCacheRepository)(nil).SaveItem), key, val, expire)
}
