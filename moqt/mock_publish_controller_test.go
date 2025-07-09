package moqt

import "github.com/stretchr/testify/mock"

var _ PublishController = (*MockPublishController)(nil)

// MockPublishController is a mock implementation of ReceiveSubscribeStream for testing
type MockPublishController struct {
	mock.Mock
}

func (m *MockPublishController) WriteInfo(info Info) error {
	args := m.Called(info)
	return args.Error(0)
}

func (m *MockPublishController) SubscribeID() SubscribeID {
	return m.Called().Get(0).(SubscribeID)
}

func (m *MockPublishController) SubscribeConfig() (*SubscribeConfig, error) {
	args := m.Called()
	return args.Get(0).(*SubscribeConfig), args.Error(1)
}

func (m *MockPublishController) Updated() <-chan struct{} {
	return m.Called().Get(0).(<-chan struct{})
}

func (m *MockPublishController) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPublishController) CloseWithError(code SubscribeErrorCode) error {
	args := m.Called(code)
	return args.Error(0)
}
