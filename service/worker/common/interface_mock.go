// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go
//
// Generated by this command:
//
//	mockgen -package common -source interface.go -destination interface_mock.go
//

// Package common is a generated GoMock package.
package common

import (
	reflect "reflect"

	worker "go.temporal.io/sdk/worker"
	namespace "go.temporal.io/server/common/namespace"
	gomock "go.uber.org/mock/gomock"
)

// MockWorkerComponent is a mock of WorkerComponent interface.
type MockWorkerComponent struct {
	ctrl     *gomock.Controller
	recorder *MockWorkerComponentMockRecorder
	isgomock struct{}
}

// MockWorkerComponentMockRecorder is the mock recorder for MockWorkerComponent.
type MockWorkerComponentMockRecorder struct {
	mock *MockWorkerComponent
}

// NewMockWorkerComponent creates a new mock instance.
func NewMockWorkerComponent(ctrl *gomock.Controller) *MockWorkerComponent {
	mock := &MockWorkerComponent{ctrl: ctrl}
	mock.recorder = &MockWorkerComponentMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWorkerComponent) EXPECT() *MockWorkerComponentMockRecorder {
	return m.recorder
}

// DedicatedActivityWorkerOptions mocks base method.
func (m *MockWorkerComponent) DedicatedActivityWorkerOptions() *DedicatedWorkerOptions {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DedicatedActivityWorkerOptions")
	ret0, _ := ret[0].(*DedicatedWorkerOptions)
	return ret0
}

// DedicatedActivityWorkerOptions indicates an expected call of DedicatedActivityWorkerOptions.
func (mr *MockWorkerComponentMockRecorder) DedicatedActivityWorkerOptions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DedicatedActivityWorkerOptions", reflect.TypeOf((*MockWorkerComponent)(nil).DedicatedActivityWorkerOptions))
}

// DedicatedWorkflowWorkerOptions mocks base method.
func (m *MockWorkerComponent) DedicatedWorkflowWorkerOptions() *DedicatedWorkerOptions {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DedicatedWorkflowWorkerOptions")
	ret0, _ := ret[0].(*DedicatedWorkerOptions)
	return ret0
}

// DedicatedWorkflowWorkerOptions indicates an expected call of DedicatedWorkflowWorkerOptions.
func (mr *MockWorkerComponentMockRecorder) DedicatedWorkflowWorkerOptions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DedicatedWorkflowWorkerOptions", reflect.TypeOf((*MockWorkerComponent)(nil).DedicatedWorkflowWorkerOptions))
}

// RegisterActivities mocks base method.
func (m *MockWorkerComponent) RegisterActivities(registry worker.Registry) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterActivities", registry)
}

// RegisterActivities indicates an expected call of RegisterActivities.
func (mr *MockWorkerComponentMockRecorder) RegisterActivities(registry any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterActivities", reflect.TypeOf((*MockWorkerComponent)(nil).RegisterActivities), registry)
}

// RegisterWorkflow mocks base method.
func (m *MockWorkerComponent) RegisterWorkflow(registry worker.Registry) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterWorkflow", registry)
}

// RegisterWorkflow indicates an expected call of RegisterWorkflow.
func (mr *MockWorkerComponentMockRecorder) RegisterWorkflow(registry any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterWorkflow", reflect.TypeOf((*MockWorkerComponent)(nil).RegisterWorkflow), registry)
}

// MockPerNSWorkerComponent is a mock of PerNSWorkerComponent interface.
type MockPerNSWorkerComponent struct {
	ctrl     *gomock.Controller
	recorder *MockPerNSWorkerComponentMockRecorder
	isgomock struct{}
}

// MockPerNSWorkerComponentMockRecorder is the mock recorder for MockPerNSWorkerComponent.
type MockPerNSWorkerComponentMockRecorder struct {
	mock *MockPerNSWorkerComponent
}

// NewMockPerNSWorkerComponent creates a new mock instance.
func NewMockPerNSWorkerComponent(ctrl *gomock.Controller) *MockPerNSWorkerComponent {
	mock := &MockPerNSWorkerComponent{ctrl: ctrl}
	mock.recorder = &MockPerNSWorkerComponentMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPerNSWorkerComponent) EXPECT() *MockPerNSWorkerComponentMockRecorder {
	return m.recorder
}

// DedicatedWorkerOptions mocks base method.
func (m *MockPerNSWorkerComponent) DedicatedWorkerOptions(arg0 *namespace.Namespace) *PerNSDedicatedWorkerOptions {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DedicatedWorkerOptions", arg0)
	ret0, _ := ret[0].(*PerNSDedicatedWorkerOptions)
	return ret0
}

// DedicatedWorkerOptions indicates an expected call of DedicatedWorkerOptions.
func (mr *MockPerNSWorkerComponentMockRecorder) DedicatedWorkerOptions(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DedicatedWorkerOptions", reflect.TypeOf((*MockPerNSWorkerComponent)(nil).DedicatedWorkerOptions), arg0)
}

// Register mocks base method.
func (m *MockPerNSWorkerComponent) Register(arg0 worker.Registry, arg1 *namespace.Namespace, arg2 RegistrationDetails) func() {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", arg0, arg1, arg2)
	ret0, _ := ret[0].(func())
	return ret0
}

// Register indicates an expected call of Register.
func (mr *MockPerNSWorkerComponentMockRecorder) Register(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockPerNSWorkerComponent)(nil).Register), arg0, arg1, arg2)
}
