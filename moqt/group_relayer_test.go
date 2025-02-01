package moqt

import (
	"errors"
	"io"
	"sync"
	"testing"
	"time"
)

// Mock implementations for testing
type mockGroupReader struct {
	frames    [][]byte
	current   int
	readError error
	mu        sync.Mutex
}

func newMockGroupReader(frames [][]byte) *mockGroupReader {
	return &mockGroupReader{
		frames:  frames,
		current: 0,
	}
}

func (m *mockGroupReader) ReadFrame() ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.readError != nil {
		return nil, m.readError
	}

	if m.current >= len(m.frames) {
		return nil, io.EOF
	}

	frame := m.frames[m.current]
	m.current++
	return frame, nil
}

func (m *mockGroupReader) GroupSequence() GroupSequence {
	return GroupSequence(1)
}

type mockGroupWriter struct {
	frames   [][]byte
	writeErr error
	mu       sync.Mutex
}

func (m *mockGroupWriter) Close() error {
	return nil
}

func (m *mockGroupWriter) GroupSequence() GroupSequence {
	return GroupSequence(1)
}

func newMockGroupWriter() *mockGroupWriter {
	return &mockGroupWriter{
		frames: make([][]byte, 0),
	}
}

func (m *mockGroupWriter) WriteFrame(frame []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.writeErr != nil {
		return m.writeErr
	}

	frameCopy := make([]byte, len(frame))
	copy(frameCopy, frame)
	m.frames = append(m.frames, frameCopy)
	return nil
}

// Tests
func TestNewGroupRelayer(t *testing.T) {
	frames := [][]byte{
		[]byte("frame1"),
		[]byte("frame2"),
	}
	gr := newMockGroupReader(frames)
	relayer := NewGroupRelayer(gr)
	defer relayer.Release()

	if relayer.GroupSequence() != GroupSequence(1) {
		t.Errorf("Expected group sequence 1, got %d", relayer.GroupSequence())
	}

	if relayer.Closed() {
		t.Error("New relayer should not be closed")
	}
}

func TestGroupRelayer_RelayFrame(t *testing.T) {
	tc := map[string]struct {
		frames      [][]byte
		readerError error
		writerError error
		expectError bool
	}{
		"successful relay": {
			frames:      [][]byte{[]byte("frame1"), []byte("frame2")},
			expectError: false,
		},
		"reader error": {
			frames:      [][]byte{[]byte("frame1")},
			readerError: errors.New("read error"),
			expectError: true,
		},
		"writer error": {
			frames:      [][]byte{[]byte("frame1")},
			writerError: errors.New("write error"),
			expectError: true,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			gr := newMockGroupReader(tc.frames)
			gw := newMockGroupWriter()

			if tc.readerError != nil {
				gr.readError = tc.readerError
			}
			if tc.writerError != nil {
				gw.writeErr = tc.writerError
			}

			relayer := NewGroupRelayer(gr)
			defer relayer.Release()

			// Use channel to handle async relay completion
			done := make(chan error)
			go func() {
				done <- relayer.Relay(gw)
			}()

			// Wait for relay to complete or timeout
			var err error
			select {
			case err = <-done:
			case <-time.After(time.Second):
				t.Fatal("Relay timeout")
			}

			if tc.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tc.expectError {
				// Verify frames were correctly relayed
				if len(gw.frames) != len(tc.frames) {
					t.Errorf("Expected %d frames, got %d", len(tc.frames), len(gw.frames))
				}
				for i, frame := range tc.frames {
					if i < len(gw.frames) && string(gw.frames[i]) != string(frame) {
						t.Errorf("Frame %d mismatch: expected %s, got %s", i, string(frame), string(gw.frames[i]))
					}
				}
			}
		})
	}
}

func TestGroupRelayer_Release(t *testing.T) {
	frames := [][]byte{[]byte("frame1")}
	gr := newMockGroupReader(frames)
	relayer := NewGroupRelayer(gr)

	if relayer.Closed() {
		t.Error("Relayer should not be closed before Release")
	}

	relayer.Release()

	if !relayer.Closed() {
		t.Error("Relayer should be closed after Release")
	}
}

func TestGroupRelayer_NilWriter(t *testing.T) {
	gr := newMockGroupReader([][]byte{})
	relayer := NewGroupRelayer(gr)
	defer relayer.Release()

	err := relayer.Relay(nil)
	if err == nil {
		t.Error("Expected error for nil writer, got nil")
	}
}
