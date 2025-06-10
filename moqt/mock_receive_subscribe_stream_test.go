package moqt

import "github.com/stretchr/testify/mock"

// MockReceiveSubscribeStream is a mock implementation of ReceiveSubscribeStream for testing
type MockReceiveSubscribeStream struct {
	mock.Mock
}

func (m *MockReceiveSubscribeStream) SubscribeID() SubscribeID {
	return m.Called().Get(0).(SubscribeID)
}

func (m *MockReceiveSubscribeStream) SubscribeConfig() (*SubscribeConfig, error) {
	args := m.Called()
	return args.Get(0).(*SubscribeConfig), args.Error(1)
}

func (m *MockReceiveSubscribeStream) Updated() <-chan struct{} {
	return m.Called().Get(0).(<-chan struct{})
}
