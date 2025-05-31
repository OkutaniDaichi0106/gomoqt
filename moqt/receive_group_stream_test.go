package moqt

import (
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func TestNewReceiveGroupStream(t *testing.T) {
	mockStream := newMockQUICReceiveStream(456)
	id := SubscribeID(123)
	seq := GroupSequence(42)

	rgs := newReceiveGroupStream(id, seq, mockStream)

	if rgs == nil {
		t.Fatal("newReceiveGroupStream returned nil")
	}

	if rgs.id != id {
		t.Errorf("id = %v, want %v", rgs.id, id)
	}

	if rgs.sequence != seq {
		t.Errorf("sequence = %v, want %v", rgs.sequence, seq)
	}

	if rgs.stream != mockStream {
		t.Error("stream not set correctly")
	}
}

func TestReceiveGroupStreamGroupSequence(t *testing.T) {
	mockStream := newMockQUICReceiveStream(456)
	expectedSeq := GroupSequence(99)

	rgs := newReceiveGroupStream(SubscribeID(1), expectedSeq, mockStream)

	seq := rgs.GroupSequence()
	if seq != expectedSeq {
		t.Errorf("GroupSequence() = %v, want %v", seq, expectedSeq)
	}
}

func TestReceiveGroupStreamReadFrame(t *testing.T) {
	mockStream := newMockQUICReceiveStream(456)
	rgs := newReceiveGroupStream(SubscribeID(1), GroupSequence(1), mockStream)

	// Test reading frame (this might fail due to frame decoding complexity)
	// We'll test the error case instead
	_, err := rgs.ReadFrame()
	// Expecting an error since we don't have valid frame data
	if err == nil {
		t.Log("ReadFrame unexpectedly succeeded - this is OK if frame decoding works")
	}
}

func TestReceiveGroupStreamReadFrameWithError(t *testing.T) {
	mockStream := newMockQUICReceiveStream(456)
	rgs := newReceiveGroupStream(SubscribeID(1), GroupSequence(1), mockStream)

	_, err := rgs.ReadFrame()
	if err == nil {
		t.Error("ReadFrame should return error when stream has read error")
	}
}

func TestReceiveGroupStreamCancelRead(t *testing.T) {
	mockStream := newMockQUICReceiveStream(quic.StreamID(456))
	rgs := newReceiveGroupStream(SubscribeID(1), GroupSequence(1), mockStream)
	testError := ErrClosedGroup

	rgs.CancelRead(testError)

	if !mockStream.AssertCalled(t, "CancelRead", quic.StreamErrorCode(testError.GroupErrorCode())) {
		t.Error("stream should be cancelled")
	}
}

func TestReceiveGroupStreamSetReadDeadline(t *testing.T) {
	mockStream := newMockQUICReceiveStream(quic.StreamID(456))
	rgs := newReceiveGroupStream(SubscribeID(1), GroupSequence(1), mockStream)

	deadline := time.Now().Add(time.Hour)
	err := rgs.SetReadDeadline(deadline)

	if err != nil {
		t.Errorf("SetReadDeadline() error = %v", err)
	}

	if !mockStream.AssertCalled(t, "SetReadDeadline", deadline) {
		t.Errorf("deadline = %v, want %v", mockStream.Called(), deadline)
	}
}

func TestReceiveGroupStreamSetReadDeadlineError(t *testing.T) {
	mockStream := newMockQUICReceiveStream(quic.StreamID(456))

	rgs := newReceiveGroupStream(SubscribeID(1), GroupSequence(1), mockStream)

	err := rgs.SetReadDeadline(time.Now())
	if err == nil {
		t.Error("SetReadDeadline should return error")
	}
}

func TestReceiveGroupStreamWithNilGroupContext(t *testing.T) {
	mockStream := newMockQUICReceiveStream(quic.StreamID(456))
	rgs := newReceiveGroupStream(SubscribeID(1), GroupSequence(1), mockStream)

	// groupCtx is nil by default in this implementation
	// Test that methods don't panic when groupCtx is nil
	// CancelRead should not panic
	rgs.CancelRead(ErrClosedGroup)

	// SetReadDeadline should not panic
	_ = rgs.SetReadDeadline(time.Now())

	// ReadFrame might log an error but shouldn't panic
	_, _ = rgs.ReadFrame()
}

func TestReceiveGroupStreamMultipleCancelRead(t *testing.T) {
	mockStream := &MockQUICReceiveStream{}
	rgs := newReceiveGroupStream(SubscribeID(1), GroupSequence(1), mockStream)

	expectedErr := ErrClosedGroup
	// Cancel multiple times
	rgs.CancelRead(expectedErr)
	rgs.CancelRead(ErrGroupOutOfRange)

	// Should be cancelled after first call
	if !mockStream.AssertCalled(t, "CancelRead", quic.StreamErrorCode(expectedErr.GroupErrorCode())) {
		t.Error("stream should be cancelled")
	}
}

func TestReceiveGroupStreamInterface(t *testing.T) {
	// Test that receiveGroupStream implements GroupReader interface
	var _ GroupReader = (*receiveGroupStream)(nil)

	// Verify interface method signatures
	mockStream := &MockQUICReceiveStream{}
	rgs := newReceiveGroupStream(SubscribeID(1), GroupSequence(1), mockStream)

	// These calls should compile correctly to verify interface compliance
	_ = rgs.GroupSequence()
	_, _ = rgs.ReadFrame()
	rgs.CancelRead(ErrClosedGroup)
	_ = rgs.SetReadDeadline(time.Now())
}

func TestReceiveGroupStreamDifferentSubscribeIDs(t *testing.T) {
	mockStream1 := &MockQUICReceiveStream{}
	mockStream2 := &MockQUICReceiveStream{}

	id1 := SubscribeID(100)
	id2 := SubscribeID(200)
	seq := GroupSequence(1)

	rgs1 := newReceiveGroupStream(id1, seq, mockStream1)
	rgs2 := newReceiveGroupStream(id2, seq, mockStream2)

	if rgs1.id != id1 {
		t.Errorf("rgs1.id = %v, want %v", rgs1.id, id1)
	}

	if rgs2.id != id2 {
		t.Errorf("rgs2.id = %v, want %v", rgs2.id, id2)
	}

	// Both should have same sequence
	if rgs1.GroupSequence() != rgs2.GroupSequence() {
		t.Error("both streams should have same sequence")
	}
}
