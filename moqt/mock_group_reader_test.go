package moqt

import (
	"time"

	"github.com/stretchr/testify/mock"
)

var _ GroupReader = (*MockGroupReader)(nil)

type MockGroupReader struct {
	mock.Mock
}

func (m *MockGroupReader) GroupSequence() GroupSequence {
	args := m.Called()
	return args.Get(0).(GroupSequence)
}

func (m *MockGroupReader) ReadFrame() (*Frame, error) {
	args := m.Called()
	return args.Get(0).(*Frame), args.Error(1)
}

func (m *MockGroupReader) CancelRead(code GroupErrorCode) {
	m.Called(code)
}

func (m *MockGroupReader) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}
