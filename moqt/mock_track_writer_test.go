package moqt_test

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

var _ moqt.TrackWriter = (*MockTrackWriter)(nil)

// MockTrackWriter is a mock implementation of the TrackWriter interface
// It provides a simple implementation for testing purposes
type MockTrackWriter struct {
	PathValue moqt.TrackPath
	// Add more fields as needed for testing
	OpenGroupFunc func(seq moqt.GroupSequence) (moqt.GroupWriter, error)

	CloseFunc          func() error
	CloseWithErrorFunc func(err error) error

	LatestGroupSequenceValue moqt.GroupSequence
}

func (m *MockTrackWriter) TrackPath() moqt.TrackPath {
	return m.PathValue
}

func (m *MockTrackWriter) LatestGroupSequence() moqt.GroupSequence {
	return m.LatestGroupSequenceValue
}

func (m *MockTrackWriter) Info() moqt.Info {
	return moqt.Info{
		TrackPriority:       1,
		LatestGroupSequence: 0,
		GroupOrder:          0,
	}
}

func (m *MockTrackWriter) OpenGroup(seq moqt.GroupSequence) (moqt.GroupWriter, error) {
	// Update the latest group sequence
	if m.LatestGroupSequenceValue < seq {
		m.LatestGroupSequenceValue = seq
	}

	if m.OpenGroupFunc == nil {
		return nil, nil
	}

	return m.OpenGroupFunc(seq)
}

func (m *MockTrackWriter) Close() error {
	if m.CloseFunc == nil {
		return nil
	}
	return m.CloseFunc()
}

func (m *MockTrackWriter) CloseWithError(err error) error {
	if m.CloseWithErrorFunc == nil {
		return nil
	}
	return m.CloseWithErrorFunc(err)
}
