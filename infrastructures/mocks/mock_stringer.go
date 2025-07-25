// Code generated by MockGen. DO NOT EDIT.
// Source: fmt (interfaces: Stringer)
//
// Generated by this command:
//
//	mockgen -destination=infra/mocks/mock_stringer.go -package=mocks fmt Stringer
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockStringer is a mock of Stringer interface.
type MockStringer struct {
	ctrl     *gomock.Controller
	recorder *MockStringerMockRecorder
	isgomock struct{}
}

// MockStringerMockRecorder is the mock recorder for MockStringer.
type MockStringerMockRecorder struct {
	mock *MockStringer
}

// NewMockStringer creates a new mock instance.
func NewMockStringer(ctrl *gomock.Controller) *MockStringer {
	mock := &MockStringer{ctrl: ctrl}
	mock.recorder = &MockStringerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStringer) EXPECT() *MockStringerMockRecorder {
	return m.recorder
}

// String mocks base method.
func (m *MockStringer) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String.
func (mr *MockStringerMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockStringer)(nil).String))
}
