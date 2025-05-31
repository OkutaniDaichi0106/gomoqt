package moqt

import "github.com/stretchr/testify/mock"

// MockReceiveSubscribeStream is a mock implementation of ReceiveSubscribeStream for testing
type MockReceiveSubscribeStream struct {
	mock.Mock
}

func (m *MockReceiveSubscribeStream) SubscribeID() SubscribeID {
	return m.Called().Get(0).(SubscribeID)
}

func (m *MockReceiveSubscribeStream) SubscribeConfig() *SubscribeConfig {
	return m.Called().Get(0).(*SubscribeConfig)
}

func (m *MockReceiveSubscribeStream) Updated() <-chan struct{} {
	return m.Called().Get(0).(<-chan struct{})
}

func (m *MockReceiveSubscribeStream) Done() <-chan struct{} {
	return m.Called().Get(0).(<-chan struct{})
}
