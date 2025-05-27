package moqt

// MockReceiveSubscribeStream is a mock implementation of ReceiveSubscribeStream for testing
type MockReceiveSubscribeStream struct {
	subscribeID SubscribeID
	config      *SubscribeConfig
	updatedCh   chan struct{}
	doneCh      chan struct{}
}

func NewMockReceiveSubscribeStream(id SubscribeID) *MockReceiveSubscribeStream {
	return &MockReceiveSubscribeStream{
		subscribeID: id,
		config:      &SubscribeConfig{},
		updatedCh:   make(chan struct{}),
		doneCh:      make(chan struct{}),
	}
}

func (m *MockReceiveSubscribeStream) SubscribeID() SubscribeID {
	return m.subscribeID
}

func (m *MockReceiveSubscribeStream) SubuscribeConfig() *SubscribeConfig {
	return m.config
}

func (m *MockReceiveSubscribeStream) Updated() <-chan struct{} {
	return m.updatedCh
}

func (m *MockReceiveSubscribeStream) Done() <-chan struct{} {
	return m.doneCh
}
