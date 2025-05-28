package quic

import "time"

var _ ReceiveStream = (*MockReceiveStream)(nil)

type MockReceiveStream struct {
	StreamIDValue       StreamID
	ReadFunc            func(p []byte) (n int, err error)
	CancelReadFunc      func(StreamErrorCode)
	SetReadDeadlineFunc func(t time.Time) error
}

func (m *MockReceiveStream) StreamID() StreamID {
	return m.StreamIDValue
}

func (m *MockReceiveStream) Read(p []byte) (n int, err error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(p)
	}
	return 0, nil // Default behavior if no function is set
}

func (m *MockReceiveStream) CancelRead(code StreamErrorCode) {
	if m.CancelReadFunc != nil {
		m.CancelReadFunc(code)
	}
}

func (m *MockReceiveStream) SetReadDeadline(t time.Time) error {
	if m.SetReadDeadlineFunc != nil {
		return m.SetReadDeadlineFunc(t)
	}
	return nil // Default behavior if no function is set
}
