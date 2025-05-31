package moqt

import (
	"bytes"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/mock"
)

var _ quic.Stream = (*MockQUICStream)(nil)

// MockQUICStream is a mock implementation of quic.Stream using testify/mock
type MockQUICStream struct {
	mock.Mock
	WroteData *bytes.Buffer
	ReadData  *bytes.Buffer
}

func (m *MockQUICStream) StreamID() quic.StreamID {
	args := m.Called()
	return args.Get(0).(quic.StreamID)
}

func (m *MockQUICStream) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	if m.ReadData != nil {
		return m.ReadData.Read(p)
	}
	return args.Int(0), args.Error(1)
}

func (m *MockQUICStream) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	if m.WroteData != nil {
		n, err = m.WroteData.Write(p)
		if err != nil {
			return n, err
		}
	}
	return args.Int(0), args.Error(1)
}

func (m *MockQUICStream) CancelRead(code quic.StreamErrorCode) {
	m.Called(code)
}

func (m *MockQUICStream) CancelWrite(code quic.StreamErrorCode) {
	m.Called(code)
}

func (m *MockQUICStream) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockQUICStream) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockQUICStream) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockQUICStream) Close() error {
	args := m.Called()
	return args.Error(0)
}
