package moqt

import (
	"time"

	"github.com/stretchr/testify/mock"
)

var _ GroupReader = (*MockGroupReader)(nil)

type MockGroupReader struct {
	mock.Mock
	GroupSequenceValue  GroupSequence
	ReadFrameFunc       func() (*Frame, error)
	CancelReadFunc      func(reason GroupError)
	SetReadDeadlineFunc func(t time.Time) error
}

func (m *MockGroupReader) GroupSequence() GroupSequence {
	if args := m.Called(); args.Get(0) != nil {
		return args.Get(0).(GroupSequence)
	}
	return m.GroupSequenceValue
}

func (m *MockGroupReader) ReadFrame() (*Frame, error) {
	if m.ReadFrameFunc != nil {
		return m.ReadFrameFunc()
	}
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Frame), args.Error(1)
}

func (m *MockGroupReader) CancelRead(reason GroupError) {
	if m.CancelReadFunc != nil {
		m.CancelReadFunc(reason)
		return
	}
	m.Called(reason)
}

func (m *MockGroupReader) SetReadDeadline(t time.Time) error {
	if m.SetReadDeadlineFunc != nil {
		return m.SetReadDeadlineFunc(t)
	}
	args := m.Called(t)
	return args.Error(0)
}
