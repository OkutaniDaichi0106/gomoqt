package moqt

import "github.com/stretchr/testify/mock"

var _ TrackWriter = (*MockTrackWriter)(nil)

// MockTrackWriter is a mock implementation of the TrackWriter interface
// It provides a simple implementation for testing purposes
type MockTrackWriter struct {
	mock.Mock
}

func (m *MockTrackWriter) OpenGroup(seq GroupSequence) (GroupWriter, error) {
	args := m.Called(seq)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(GroupWriter), args.Error(1)
}

func (m *MockTrackWriter) Close() error {
	return m.Called().Error(0)
}

func (m *MockTrackWriter) CloseWithError(code SubscribeErrorCode) error {
	return m.Called(code).Error(0)
}
