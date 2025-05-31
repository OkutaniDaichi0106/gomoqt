package moqt

import (
	"testing"
	"time"
)

func TestGroupWriterInterface(t *testing.T) {
	// Test that sendGroupStream implements GroupWriter interface
	var _ GroupWriter = (*sendGroupStream)(nil)
}

func TestGroupWriterMethods(t *testing.T) {
	// This test verifies that the interface is properly defined
	// The actual implementation is tested in send_group_stream_test.go

	// Test that the interface has all required methods
	var gw GroupWriter

	// These should compile without errors
	_ = func() GroupSequence { return gw.GroupSequence() }
	_ = func() error { return gw.WriteFrame(&Frame{}) }
	_ = func() error { return gw.CloseWithError(ErrClosedGroup) }
	_ = func() error { return gw.SetWriteDeadline(time.Now()) }
	_ = func() error { return gw.Close() }
}
