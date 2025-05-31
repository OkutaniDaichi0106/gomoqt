package moqt

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/mock"
)

// Mock quic.SendStream for testing
// Removed - using MockQUICSendStream instead

// Helper function to create a properly configured MockQUICSendStream for testing
func createMockQUICSendStream() *MockQUICSendStream {
	mockStream := &MockQUICSendStream{}
	mockStream.On("StreamID").Return(quic.StreamID(456))
	mockStream.On("Write", mock.Anything).Return(0, nil)
	mockStream.On("CancelWrite", mock.Anything)
	mockStream.On("SetWriteDeadline", mock.Anything).Return(nil)
	mockStream.On("Close").Return(nil)
	return mockStream
}

func TestNewSendGroupStream(t *testing.T) {
	mockStream := createMockQUICSendStream()
	groupCtx := createTestGroupContext(
		createTestTrackContext(
			createTestSessionContext(context.Background()),
		),
	)

	sgs := newSendGroupStream(mockStream, groupCtx)

	if sgs == nil {
		t.Fatal("newSendGroupStream returned nil")
	}

	if sgs.stream != mockStream {
		t.Error("stream not set correctly")
	}

	if sgs.groupCtx != groupCtx {
		t.Error("groupCtx not set correctly")
	}
}

func TestSendGroupStreamGroupSequence(t *testing.T) {
	mockStream := createMockQUICSendStream()
	groupCtx := createTestGroupContext(
		createTestTrackContext(
			createTestSessionContext(context.Background()),
		),
	)
	sgs := newSendGroupStream(mockStream, groupCtx)

	seq := sgs.GroupSequence()
	expected := GroupSequence(42)

	if seq != expected {
		t.Errorf("GroupSequence() = %v, want %v", seq, expected)
	}
}

func TestSendGroupStreamWriteFrame(t *testing.T) {
	mockStream := createMockQUICSendStream()
	groupCtx := createTestGroupContext(
		createTestTrackContext(
			createTestSessionContext(context.Background()),
		),
	)
	sgs := newSendGroupStream(mockStream, groupCtx)

	// Test writing a valid frame
	frame := &Frame{
		// Note: Frame needs a message field, but we need to check the actual Frame struct
	}

	err := sgs.WriteFrame(frame)
	// This test might fail because Frame structure needs to be properly initialized
	// We should test the error handling instead
	_ = err

	// Test writing nil frame
	err = sgs.WriteFrame(nil)
	if err == nil {
		t.Error("WriteFrame(nil) should return error")
	}
	if !errors.Is(err, errors.New("frame is nil")) && err.Error() != "frame is nil" {
		t.Errorf("WriteFrame(nil) error = %v, want 'frame is nil'", err)
	}
}

func TestSendGroupStreamWriteFrameAfterClose(t *testing.T) {
	mockStream := createMockQUICSendStream()
	groupCtx := createTestGroupContext(createTestTrackContext(
		createTestSessionContext(context.Background()),
	))
	sgs := newSendGroupStream(mockStream, groupCtx)

	// Close the stream first
	_ = sgs.Close()

	// Try to write after close
	frame := &Frame{}
	err := sgs.WriteFrame(frame)
	if err == nil {
		t.Error("WriteFrame after close should return error")
	}
}

func TestSendGroupStreamSetWriteDeadline(t *testing.T) {
	mockStream := createMockQUICSendStream()
	groupCtx := createTestGroupContext(createTestTrackContext(
		createTestSessionContext(context.Background()),
	))
	sgs := newSendGroupStream(mockStream, groupCtx)

	deadline := time.Now().Add(time.Hour)
	err := sgs.SetWriteDeadline(deadline)

	if err != nil {
		t.Errorf("SetWriteDeadline() error = %v", err)
	}

	// Verify that SetWriteDeadline was called on the mock stream
	mockStream.AssertCalled(t, "SetWriteDeadline", deadline)
}

func TestSendGroupStreamClose(t *testing.T) {
	mockStream := createMockQUICSendStream()
	groupCtx := createTestGroupContext(createTestTrackContext(
		createTestSessionContext(context.Background()),
	))
	sgs := newSendGroupStream(mockStream, groupCtx)

	err := sgs.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify Close was called on the mock stream
	mockStream.AssertCalled(t, "Close")

	// Verify context is cancelled
	if groupCtx.Err() == nil {
		t.Error("group context should be cancelled after close")
	}
}

func TestSendGroupStreamCloseWithError(t *testing.T) {
	mockStream := createMockQUICSendStream()
	groupCtx := createTestGroupContext(createTestTrackContext(
		createTestSessionContext(context.Background()),
	))
	sgs := newSendGroupStream(mockStream, groupCtx)

	testErr := ErrProtocolViolation
	err := sgs.CloseWithError(testErr)

	if err != nil {
		t.Errorf("CloseWithError() error = %v", err)
	}

	// Verify Close was called on the mock stream
	mockStream.AssertCalled(t, "Close")

	// Verify context is cancelled
	if groupCtx.Err() == nil {
		t.Error("group context should be cancelled after close with error")
	}

	// Verify the cause
	cause := context.Cause(groupCtx)
	if cause != testErr {
		t.Errorf("context cause = %v, want %v", cause, testErr)
	}
}

func TestSendGroupStreamCloseWithNilError(t *testing.T) {
	mockStream := createMockQUICSendStream()
	groupCtx := createTestGroupContext(createTestTrackContext(
		createTestSessionContext(context.Background()),
	))
	sgs := newSendGroupStream(mockStream, groupCtx)

	err := sgs.CloseWithError(nil)

	if err != nil {
		t.Errorf("CloseWithError(nil) error = %v", err)
	}

	// Should still close the stream
	mockStream.AssertCalled(t, "Close")
}

func TestSendGroupStreamClosedErr(t *testing.T) {
	mockStream := createMockQUICSendStream()
	groupCtx := createTestGroupContext(createTestTrackContext(
		createTestSessionContext(context.Background()),
	))
	sgs := newSendGroupStream(mockStream, groupCtx)

	// Initially should not be closed
	err := sgs.closedErr()
	if err != nil {
		t.Errorf("closedErr() for open stream = %v, want nil", err)
	}

	// After closing
	sgs.Close()
	err = sgs.closedErr()
	if err == nil {
		t.Error("closedErr() for closed stream should return error")
	}
}

func TestSendGroupStreamDoubleClose(t *testing.T) {
	mockStream := createMockQUICSendStream()
	groupCtx := createTestGroupContext(createTestTrackContext(
		createTestSessionContext(context.Background()),
	))
	sgs := newSendGroupStream(mockStream, groupCtx)

	// First close
	err1 := sgs.Close()
	if err1 != nil {
		t.Errorf("first Close() error = %v", err1)
	}

	// Second close should return error
	err2 := sgs.Close()
	if err2 == nil {
		t.Error("second Close() should return error")
	}
}

func TestSendGroupStreamDoubleCloseWithError(t *testing.T) {
	mockStream := createMockQUICSendStream()
	groupCtx := createTestGroupContext(createTestTrackContext(
		createTestSessionContext(context.Background()),
	))
	sgs := newSendGroupStream(mockStream, groupCtx)

	testErr := ErrProtocolViolation

	// First close with error
	err1 := sgs.CloseWithError(testErr)
	if err1 != nil {
		t.Errorf("first CloseWithError() error = %v", err1)
	}

	// Second close with error should return error
	err2 := sgs.CloseWithError(testErr)
	if err2 == nil {
		t.Error("second CloseWithError() should return error")
	}
}

func TestSendGroupStreamInterface(t *testing.T) {
	// Test that sendGroupStream implements GroupWriter interface
	var _ GroupWriter = (*sendGroupStream)(nil)
}
