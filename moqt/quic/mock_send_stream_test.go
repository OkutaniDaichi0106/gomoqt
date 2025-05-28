package quic

import "time"

var _ SendStream = (*MockSendStream)(nil)

type MockSendStream struct {
	StreamIDValue        StreamID
	WriteFunc            func(p []byte) (n int, err error)
	CancelWriteFunc      func(StreamErrorCode) error
	SetWriteDeadlineFunc func(t time.Time) error
	CloseFunc            func() error
}

func (m *MockSendStream) StreamID() StreamID {
	return m.StreamIDValue
}
func (m *MockSendStream) Write(p []byte) (n int, err error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(p)
	}
	return 0, nil // Default behavior if no function is set
}
func (m *MockSendStream) CancelWrite(code StreamErrorCode) {
	if m.CancelWriteFunc != nil {
		m.CancelWriteFunc(code)
	}
}
func (m *MockSendStream) SetWriteDeadline(t time.Time) error {
	if m.SetWriteDeadlineFunc != nil {
		return m.SetWriteDeadlineFunc(t)
	}
	return nil // Default behavior if no function is set
}
func (m *MockSendStream) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil // Default behavior if no function is set
}
