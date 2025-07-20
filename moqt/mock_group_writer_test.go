package moqt

import (
	"time"

	"github.com/stretchr/testify/mock"
)

var _ GroupWriter = (*MockGroupWriter)(nil)

type MockGroupWriter struct {
	mock.Mock
}

func (m *MockGroupWriter) GroupSequence() GroupSequence {
	args := m.Called()
	return args.Get(0).(GroupSequence)
}

func (m *MockGroupWriter) WriteFrame(frame *Frame) error {
	args := m.Called(frame)
	return args.Error(0)
}

func (m *MockGroupWriter) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockGroupWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGroupWriter) CancelWrite(code GroupErrorCode) {
	args := m.Called(code)
	if args.Get(0) != nil {
		panic(args.Error(0))
	}
}
