package moqt

var _ TrackWriter = (*MockTrackWriter)(nil)

// MockTrackWriter is a mock implementation of the TrackWriter interface
// It provides a simple implementation for testing purposes
type MockTrackWriter struct {
	PathValue BroadcastPath
	// Add more fields as needed for testing
	OpenGroupFunc      func(seq GroupSequence) (GroupWriter, error)
	SendGapFunc        func(gap Gap) error
	CloseFunc          func() error
	CloseWithErrorFunc func(err error) error

	LatestGroupSequenceValue GroupSequence
}

func (m *MockTrackWriter) TrackPath() BroadcastPath {
	return m.PathValue
}

func (m *MockTrackWriter) LatestGroupSequence() GroupSequence {
	return m.LatestGroupSequenceValue
}

func (m *MockTrackWriter) Info() Info {
	return Info{
		TrackPriority:       1,
		LatestGroupSequence: 0,
		GroupOrder:          0,
	}
}

func (m *MockTrackWriter) OpenGroup(seq GroupSequence) (GroupWriter, error) {
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

func (m *MockTrackWriter) SendGap(gap Gap) error {
	if m.SendGapFunc == nil {
		return nil
	}
	return m.SendGapFunc(gap)
}
