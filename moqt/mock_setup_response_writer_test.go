package moqt

import (
	"github.com/stretchr/testify/mock"
)

var _ SetupResponseWriter = (*MockSetupResponseWriter)(nil)

// MockSetupResponseWriter is a mock implementation of SetupResponseWriter interface
type MockSetupResponseWriter struct {
	mock.Mock
}

func (m *MockSetupResponseWriter) SelectVersion(v Version) error {
	args := m.Called(v)
	return args.Error(0)
}

func (m *MockSetupResponseWriter) SetExtensions(extensions *Parameters) {
	m.Called(extensions)
}

func (m *MockSetupResponseWriter) Accept(mux *TrackMux) (*Session, error) {
	args := m.Called(mux)
	return args.Get(0).(*Session), args.Error(1)
}

func (m *MockSetupResponseWriter) Reject(code SessionErrorCode) error {
	args := m.Called(code)
	return args.Error(0)
}
