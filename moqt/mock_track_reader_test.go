package moqt

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var _ TrackReader = (*MockTrackReader)(nil)

type MockTrackReader struct {
	mock.Mock
	AcceptGroupFunc    func(context.Context) (GroupReader, error)
	CloseFunc          func() error
	CloseWithErrorFunc func(err error) error
}

func (m *MockTrackReader) AcceptGroup(ctx context.Context) (GroupReader, error) {
	if m.AcceptGroupFunc != nil {
		return m.AcceptGroupFunc(ctx)
	}
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(GroupReader), args.Error(1)
}

func (m *MockTrackReader) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	args := m.Called()
	return args.Error(0)
}

func (m *MockTrackReader) CloseWithError(err error) error {
	if m.CloseWithErrorFunc != nil {
		return m.CloseWithErrorFunc(err)
	}
	args := m.Called(err)
	return args.Error(0)
}
