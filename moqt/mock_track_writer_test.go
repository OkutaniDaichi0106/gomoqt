package moqt

var _ TrackWriter = (*MockTrackWriter)(nil)

// MockTrackWriter is a mock implementation of the TrackWriter interface
// It provides a simple implementation for testing purposes
type MockTrackWriter struct {
	// Add more fields as needed for testing
	OpenGroupFunc      func(seq GroupSequence) (GroupWriter, error)
	CloseFunc          func() error
	CloseWithErrorFunc func(err error) error
}

func (m *MockTrackWriter) OpenGroup(seq GroupSequence) (GroupWriter, error) {
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
