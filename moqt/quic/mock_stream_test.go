package quic

import "time"

var _ Stream = (*MockStream)(nil)

type MockStream struct {
	StreamIDValue        StreamID
	ReadFunc             func(p []byte) (n int, err error)
	WriteFunc            func(p []byte) (n int, err error)
	CancelReadFunc       func(StreamErrorCode) error
	CancelWriteFunc      func(StreamErrorCode) error
	SetReadDeadlineFunc  func(t time.Time) error
	SetWriteDeadlineFunc func(t time.Time) error
	CloseFunc            func() error
	SetDeadlineFunc      func(t time.Time) error
}

func (m *MockStream) StreamID() StreamID {
	return m.StreamIDValue
}
func (m *MockStream) Read(p []byte) (n int, err error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(p)
	}
	return 0, nil // Default behavior if no function is set
}
func (m *MockStream) Write(p []byte) (n int, err error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(p)
	}
	return 0, nil // Default behavior if no function is set
}

func (m *MockStream) CancelWrite(code StreamErrorCode) {
	if m.CancelWriteFunc != nil {
		m.CancelWriteFunc(code)
	}
}
func (m *MockStream) CancelRead(code StreamErrorCode) {
	if m.CancelReadFunc != nil {
		m.CancelReadFunc(code)
	}
}

func (m *MockStream) SetReadDeadline(t time.Time) error {
	if m.SetReadDeadlineFunc != nil {
		return m.SetReadDeadlineFunc(t)
	}
	return nil // Default behavior if no function is set
}
func (m *MockStream) SetWriteDeadline(t time.Time) error {
	if m.SetWriteDeadlineFunc != nil {
		return m.SetWriteDeadlineFunc(t)
	}
	return nil // Default behavior if no function is set
}

func (m *MockStream) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil // Default behavior if no function is set
}

func (m *MockStream) SetDeadline(t time.Time) error {
	if m.SetDeadlineFunc != nil {
		return m.SetDeadlineFunc(t)
	}
	return nil // Default behavior if no function is set
}
