package moqt

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

var _ http.ResponseWriter = (*MockHTTPResponseWriter)(nil)

// MockHTTPResponseWriter is a mock implementation of http.ResponseWriter interface
type MockHTTPResponseWriter struct {
	mock.Mock
}

func (m *MockHTTPResponseWriter) Header() http.Header {
	args := m.Called()
	if header, ok := args.Get(0).(http.Header); ok {
		return header
	}
	return make(http.Header)
}

func (m *MockHTTPResponseWriter) Write(data []byte) (int, error) {
	args := m.Called(data)
	return args.Int(0), args.Error(1)
}

func (m *MockHTTPResponseWriter) WriteHeader(statusCode int) {
	m.Called(statusCode)
}
