package moqt

import (
	"github.com/stretchr/testify/mock"
)

var _ SetupResponseWriter = (*MockSetupResponseWriter)(nil)

// MockSetupResponseWriter is a mock implementation of SetupResponseWriter interface
type MockSetupResponseWriter struct {
	mock.Mock
}

func (m *MockSetupResponseWriter) WriteServerInfo(v Version, extensions *Parameters) error {
	args := m.Called(v, extensions)
	return args.Error(0)
}

func (m *MockSetupResponseWriter) Reject(code SessionErrorCode) error {
	args := m.Called(code)
	return args.Error(0)
}
