// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/theparanoids/grafeas-rds/go/v1beta1/storage (interfaces: StorageCreator)

// Package storage is a generated GoMock package.
package storage

import (
	driver "database/sql/driver"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockStorageCreator is a mock of StorageCreator interface.
type MockStorageCreator struct {
	ctrl     *gomock.Controller
	recorder *MockStorageCreatorMockRecorder
}

// MockStorageCreatorMockRecorder is the mock recorder for MockStorageCreator.
type MockStorageCreatorMockRecorder struct {
	mock *MockStorageCreator
}

// NewMockStorageCreator creates a new mock instance.
func NewMockStorageCreator(ctrl *gomock.Controller) *MockStorageCreator {
	mock := &MockStorageCreator{ctrl: ctrl}
	mock.recorder = &MockStorageCreatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorageCreator) EXPECT() *MockStorageCreatorMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockStorageCreator) Create(arg0 driver.Connector, arg1 string) (Storage, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(Storage)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockStorageCreatorMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockStorageCreator)(nil).Create), arg0, arg1)
}
