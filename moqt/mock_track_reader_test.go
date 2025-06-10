package moqt

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var _ TrackReader = (*MockTrackReader)(nil)

type MockTrackReader struct {
	mock.Mock
}

func (m *MockTrackReader) AcceptGroup(ctx context.Context) (GroupReader, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(GroupReader), args.Error(1)
}

func (m *MockTrackReader) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTrackReader) CloseWithError(code SubscribeErrorCode) error {
	args := m.Called(code)
	return args.Error(0)
}
