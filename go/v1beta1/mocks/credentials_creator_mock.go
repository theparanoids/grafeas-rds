// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/theparanoids/grafeas-rds/go/v1beta1/storage (interfaces: CredentialsCreator)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	credentials "github.com/aws/aws-sdk-go/aws/credentials"
	gomock "github.com/golang/mock/gomock"
	config "github.com/theparanoids/grafeas-rds/go/config"
)

// MockCredentialsCreator is a mock of CredentialsCreator interface.
type MockCredentialsCreator struct {
	ctrl     *gomock.Controller
	recorder *MockCredentialsCreatorMockRecorder
}

// MockCredentialsCreatorMockRecorder is the mock recorder for MockCredentialsCreator.
type MockCredentialsCreatorMockRecorder struct {
	mock *MockCredentialsCreator
}

// NewMockCredentialsCreator creates a new mock instance.
func NewMockCredentialsCreator(ctrl *gomock.Controller) *MockCredentialsCreator {
	mock := &MockCredentialsCreator{ctrl: ctrl}
	mock.recorder = &MockCredentialsCreatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCredentialsCreator) EXPECT() *MockCredentialsCreatorMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockCredentialsCreator) Create(arg0 config.IAMAuthConfig) (*credentials.Credentials, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0)
	ret0, _ := ret[0].(*credentials.Credentials)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockCredentialsCreatorMockRecorder) Create(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockCredentialsCreator)(nil).Create), arg0)
}
