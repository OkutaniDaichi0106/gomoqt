// Code generated by MockGen. DO NOT EDIT.
// Source: ../stream.go
//
// Generated by this command:
//
//	mockgen -source=../stream.go -destination=./mock_stream.go
//

// Package mock_transport is a generated GoMock package.
package mock_transport

import (
	reflect "reflect"
	time "time"

	transport "github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
	gomock "go.uber.org/mock/gomock"
)

// MockSendStream is a mock of SendStream interface.
type MockSendStream struct {
	ctrl     *gomock.Controller
	recorder *MockSendStreamMockRecorder
	isgomock struct{}
}

// MockSendStreamMockRecorder is the mock recorder for MockSendStream.
type MockSendStreamMockRecorder struct {
	mock *MockSendStream
}

// NewMockSendStream creates a new mock instance.
func NewMockSendStream(ctrl *gomock.Controller) *MockSendStream {
	mock := &MockSendStream{ctrl: ctrl}
	mock.recorder = &MockSendStreamMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSendStream) EXPECT() *MockSendStreamMockRecorder {
	return m.recorder
}

// CancelWrite mocks base method.
func (m *MockSendStream) CancelWrite(arg0 transport.StreamErrorCode) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CancelWrite", arg0)
}

// CancelWrite indicates an expected call of CancelWrite.
func (mr *MockSendStreamMockRecorder) CancelWrite(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CancelWrite", reflect.TypeOf((*MockSendStream)(nil).CancelWrite), arg0)
}

// Close mocks base method.
func (m *MockSendStream) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockSendStreamMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockSendStream)(nil).Close))
}

// SetWriteDeadline mocks base method.
func (m *MockSendStream) SetWriteDeadline(arg0 time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetWriteDeadline", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetWriteDeadline indicates an expected call of SetWriteDeadline.
func (mr *MockSendStreamMockRecorder) SetWriteDeadline(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetWriteDeadline", reflect.TypeOf((*MockSendStream)(nil).SetWriteDeadline), arg0)
}

// StreamID mocks base method.
func (m *MockSendStream) StreamID() transport.StreamID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StreamID")
	ret0, _ := ret[0].(transport.StreamID)
	return ret0
}

// StreamID indicates an expected call of StreamID.
func (mr *MockSendStreamMockRecorder) StreamID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StreamID", reflect.TypeOf((*MockSendStream)(nil).StreamID))
}

// Write mocks base method.
func (m *MockSendStream) Write(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockSendStreamMockRecorder) Write(p any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockSendStream)(nil).Write), p)
}

// MockReceiveStream is a mock of ReceiveStream interface.
type MockReceiveStream struct {
	ctrl     *gomock.Controller
	recorder *MockReceiveStreamMockRecorder
	isgomock struct{}
}

// MockReceiveStreamMockRecorder is the mock recorder for MockReceiveStream.
type MockReceiveStreamMockRecorder struct {
	mock *MockReceiveStream
}

// NewMockReceiveStream creates a new mock instance.
func NewMockReceiveStream(ctrl *gomock.Controller) *MockReceiveStream {
	mock := &MockReceiveStream{ctrl: ctrl}
	mock.recorder = &MockReceiveStreamMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReceiveStream) EXPECT() *MockReceiveStreamMockRecorder {
	return m.recorder
}

// CancelRead mocks base method.
func (m *MockReceiveStream) CancelRead(arg0 transport.StreamErrorCode) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CancelRead", arg0)
}

// CancelRead indicates an expected call of CancelRead.
func (mr *MockReceiveStreamMockRecorder) CancelRead(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CancelRead", reflect.TypeOf((*MockReceiveStream)(nil).CancelRead), arg0)
}

// Read mocks base method.
func (m *MockReceiveStream) Read(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockReceiveStreamMockRecorder) Read(p any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockReceiveStream)(nil).Read), p)
}

// SetReadDeadline mocks base method.
func (m *MockReceiveStream) SetReadDeadline(arg0 time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetReadDeadline", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetReadDeadline indicates an expected call of SetReadDeadline.
func (mr *MockReceiveStreamMockRecorder) SetReadDeadline(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetReadDeadline", reflect.TypeOf((*MockReceiveStream)(nil).SetReadDeadline), arg0)
}

// StreamID mocks base method.
func (m *MockReceiveStream) StreamID() transport.StreamID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StreamID")
	ret0, _ := ret[0].(transport.StreamID)
	return ret0
}

// StreamID indicates an expected call of StreamID.
func (mr *MockReceiveStreamMockRecorder) StreamID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StreamID", reflect.TypeOf((*MockReceiveStream)(nil).StreamID))
}

// MockStream is a mock of Stream interface.
type MockStream struct {
	ctrl     *gomock.Controller
	recorder *MockStreamMockRecorder
	isgomock struct{}
}

// MockStreamMockRecorder is the mock recorder for MockStream.
type MockStreamMockRecorder struct {
	mock *MockStream
}

// NewMockStream creates a new mock instance.
func NewMockStream(ctrl *gomock.Controller) *MockStream {
	mock := &MockStream{ctrl: ctrl}
	mock.recorder = &MockStreamMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStream) EXPECT() *MockStreamMockRecorder {
	return m.recorder
}

// CancelRead mocks base method.
func (m *MockStream) CancelRead(arg0 transport.StreamErrorCode) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CancelRead", arg0)
}

// CancelRead indicates an expected call of CancelRead.
func (mr *MockStreamMockRecorder) CancelRead(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CancelRead", reflect.TypeOf((*MockStream)(nil).CancelRead), arg0)
}

// CancelWrite mocks base method.
func (m *MockStream) CancelWrite(arg0 transport.StreamErrorCode) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CancelWrite", arg0)
}

// CancelWrite indicates an expected call of CancelWrite.
func (mr *MockStreamMockRecorder) CancelWrite(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CancelWrite", reflect.TypeOf((*MockStream)(nil).CancelWrite), arg0)
}

// Close mocks base method.
func (m *MockStream) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStreamMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStream)(nil).Close))
}

// Read mocks base method.
func (m *MockStream) Read(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockStreamMockRecorder) Read(p any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockStream)(nil).Read), p)
}

// SetDeadLine mocks base method.
func (m *MockStream) SetDeadLine(arg0 time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetDeadLine", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetDeadLine indicates an expected call of SetDeadLine.
func (mr *MockStreamMockRecorder) SetDeadLine(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDeadLine", reflect.TypeOf((*MockStream)(nil).SetDeadLine), arg0)
}

// SetReadDeadline mocks base method.
func (m *MockStream) SetReadDeadline(arg0 time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetReadDeadline", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetReadDeadline indicates an expected call of SetReadDeadline.
func (mr *MockStreamMockRecorder) SetReadDeadline(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetReadDeadline", reflect.TypeOf((*MockStream)(nil).SetReadDeadline), arg0)
}

// SetWriteDeadline mocks base method.
func (m *MockStream) SetWriteDeadline(arg0 time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetWriteDeadline", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetWriteDeadline indicates an expected call of SetWriteDeadline.
func (mr *MockStreamMockRecorder) SetWriteDeadline(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetWriteDeadline", reflect.TypeOf((*MockStream)(nil).SetWriteDeadline), arg0)
}

// StreamID mocks base method.
func (m *MockStream) StreamID() transport.StreamID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StreamID")
	ret0, _ := ret[0].(transport.StreamID)
	return ret0
}

// StreamID indicates an expected call of StreamID.
func (mr *MockStreamMockRecorder) StreamID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StreamID", reflect.TypeOf((*MockStream)(nil).StreamID))
}

// Write mocks base method.
func (m *MockStream) Write(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockStreamMockRecorder) Write(p any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockStream)(nil).Write), p)
}

// MockStreamError is a mock of StreamError interface.
type MockStreamError struct {
	ctrl     *gomock.Controller
	recorder *MockStreamErrorMockRecorder
	isgomock struct{}
}

// MockStreamErrorMockRecorder is the mock recorder for MockStreamError.
type MockStreamErrorMockRecorder struct {
	mock *MockStreamError
}

// NewMockStreamError creates a new mock instance.
func NewMockStreamError(ctrl *gomock.Controller) *MockStreamError {
	mock := &MockStreamError{ctrl: ctrl}
	mock.recorder = &MockStreamErrorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStreamError) EXPECT() *MockStreamErrorMockRecorder {
	return m.recorder
}

// Error mocks base method.
func (m *MockStreamError) Error() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Error")
	ret0, _ := ret[0].(string)
	return ret0
}

// Error indicates an expected call of Error.
func (mr *MockStreamErrorMockRecorder) Error() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockStreamError)(nil).Error))
}

// StreamErrorCode mocks base method.
func (m *MockStreamError) StreamErrorCode() transport.StreamErrorCode {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StreamErrorCode")
	ret0, _ := ret[0].(transport.StreamErrorCode)
	return ret0
}

// StreamErrorCode indicates an expected call of StreamErrorCode.
func (mr *MockStreamErrorMockRecorder) StreamErrorCode() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StreamErrorCode", reflect.TypeOf((*MockStreamError)(nil).StreamErrorCode))
}
