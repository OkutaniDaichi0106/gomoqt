package moqt

import (
	"time"

	"github.com/stretchr/testify/mock"
)

var _ GroupWriter = (*MockGroupWriter)(nil)

type MockGroupWriter struct {
	mock.Mock
	GroupSequenceValue   GroupSequence
	WriteFrameFunc       func(*Frame) error
	SetWriteDeadlineFunc func(time.Time) error
	CloseFunc            func() error
	CloseWithErrorFunc   func(err error) error
}

func (m *MockGroupWriter) GroupSequence() GroupSequence {
	if args := m.Called(); args.Get(0) != nil {
		return args.Get(0).(GroupSequence)
	}
	return m.GroupSequenceValue
}

func (m *MockGroupWriter) WriteFrame(frame *Frame) error {
	if m.WriteFrameFunc != nil {
		return m.WriteFrameFunc(frame)
	}
	args := m.Called(frame)
	return args.Error(0)
}

func (m *MockGroupWriter) SetWriteDeadline(t time.Time) error {
	if m.SetWriteDeadlineFunc != nil {
		return m.SetWriteDeadlineFunc(t)
	}
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockGroupWriter) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	args := m.Called()
	return args.Error(0)
}

func (m *MockGroupWriter) CloseWithError(err error) error {
	if m.CloseWithErrorFunc != nil {
		return m.CloseWithErrorFunc(err)
	}
	args := m.Called(err)
	return args.Error(0)
}
