package moqt

import (
	"github.com/stretchr/testify/mock"
)

var _ SetupHandler = (*MockSetupHandler)(nil)

// MockSetupHandler is a mock implementation of SetupHandler interface
type MockSetupHandler struct {
	mock.Mock
}

func (m *MockSetupHandler) ServeMOQ(w SetupResponseWriter, r *SetupRequest) {
	m.Called(w, r)
}
