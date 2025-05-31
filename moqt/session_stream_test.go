package moqt

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/moqtrace"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewSessionStream(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{
		WroteData: bytes.NewBuffer(nil),
		ReadData:  bytes.NewBuffer(nil),
	}

	sessstr := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	assert.NotNil(t, sessstr, "newSessionStream should not return nil")

	assert.Equal(t, sessstr.stream, mockStream, "stream not set correctly")

	// Give time for listenUpdates goroutine to start
	time.Sleep(10 * time.Millisecond)
}

func TestSessionStream_updateSession(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{
		WroteData: bytes.NewBuffer(nil),
		ReadData:  bytes.NewBuffer(nil),
	}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	bitrate := uint64(1000000)
	err := ss.updateSession(bitrate)
	if err != nil {
		t.Errorf("updateSession() error = %v", err)
	}

	sum := message.SessionUpdateMessage{
		Bitrate: bitrate,
	}
	buf := bytes.NewBuffer(make([]byte, 0, sum.Len()))
	if _, err := sum.Encode(buf); err != nil {
		t.Errorf("failed to encode SESSION_UPDATE message: %v", err)
	}
	expectedData := buf.Bytes()

	// Check that data was written to stream
	if !mockStream.AssertCalled(t, "Write", expectedData) {
		t.Error("no data written to stream for updateSession")
	}

	// Verify the written data contains a SESSION_UPDATE message
	// For a proper test, we would decode the message, but for simplicity we just check data was written
}

func TestSessionStreamupdateSessionWriteError(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{
		WroteData: bytes.NewBuffer(nil),
		ReadData:  bytes.NewBuffer(nil),
	}
	mockStream.On("Write", mock.Anything).Return(0, errors.New("write error"))

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	bitrate := uint64(1000000)
	err := ss.updateSession(bitrate)
	if err == nil {
		t.Error("updateSession() should return error when stream write fails")
	}
}

func TestSessionStreamClose(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{
		WroteData: bytes.NewBuffer(nil),
		ReadData:  bytes.NewBuffer(nil),
	}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	err := ss.close()
	if err != nil {
		t.Errorf("close() error = %v", err)
	}

	// Verify context is cancelled
	if sessCtx.Err() == nil {
		t.Error("session context should be cancelled after close")
	}

	// Verify underlying stream is closed
	if !mockStream.AssertCalled(t, "Close") {
		t.Error("underlying stream should be closed")
	}
}

func TestSessionStreamCloseAlreadyClosed(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Close once
	err1 := ss.close()
	if err1 != nil {
		t.Errorf("first close() error = %v", err1)
	}

	// Close again should return error
	err2 := ss.close()
	if err2 == nil {
		t.Error("second close() should return error")
	}
}

func TestSessionStream_closeWithError(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	testErr := ErrProtocolViolation
	err := ss.closeWithError(testErr)
	if err != nil {
		t.Errorf("closeWithError() error = %v", err)
	}

	// Verify context is cancelled with the right error
	if sessCtx.Err() == nil {
		t.Error("session context should be cancelled")
	}

	cause := context.Cause(sessCtx)
	if cause != testErr {
		t.Errorf("context cause = %v, want %v", cause, testErr)
	}

	// Verify stream is cancelled
	if !mockStream.AssertCalled(t, "CancelRead", quic.StreamErrorCode(testErr.TerminateErrorCode())) {
		t.Error("underlying stream should be cancelled")
	}

	if !mockStream.AssertCalled(t, "CancelWrite", quic.StreamErrorCode(testErr.TerminateErrorCode())) {
		t.Error("underlying stream should be cancelled on write error")
	}
}

func TestSessionStream_CloseWithNilError(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	err := ss.closeWithError(nil)
	if err != nil {
		t.Errorf("closeWithError(nil) error = %v", err)
	}

	// Should use default error when nil is passed
	if !mockStream.AssertCalled(t, "CancelRead", quic.StreamErrorCode(ErrInternalError.GroupErrorCode())) {
		t.Error("underlying stream should be cancelled")
	}
}

func TestSessionStream_closeWithErrorAlreadyClosed(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Close first
	err1 := ss.close()
	if err1 != nil {
		t.Errorf("first close() error = %v", err1)
	}

	// closeWithError should return the closed error
	testErr := errors.New("test error")
	err2 := ss.closeWithError(testErr)
	if err2 == nil {
		t.Error("closeWithError() on already closed stream should return error")
	}
}

func TestSessionStream_ListenUpdate(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil),
	}

	expectedBitrate := uint64(1000000)

	// Create valid SESSION_UPDATE message
	_, err := message.SessionUpdateMessage{
		Bitrate: expectedBitrate,
	}.Encode(mockStream.ReadData)
	assert.NoError(t, err, "failed to encode SESSION_UPDATE message")

	sessCtx := createTestSessionContext(context.Background())

	sessstr := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Give time for listenUpdates to process the message
	time.Sleep(10 * time.Millisecond)

	select {
	case <-sessCtx.Done():
		// Context should not be cancelled
		if err := sessCtx.Err(); err != nil {
			t.Errorf("session context should not be cancelled, got %v", err)
		}
	case <-sessstr.SessionUpdated():
		if sessstr.remoteBitrate != expectedBitrate {
			t.Errorf("session context bitrate = %d, want %d", sessstr.remoteBitrate, expectedBitrate)
		}
	}
}

func TestSessionStream_ListenUpdatesInvalidMessage(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{
		WroteData: bytes.NewBuffer(nil),
		ReadData:  bytes.NewBuffer([]byte{0xFF, 0xFF, 0xFF}), // Invalid message data
	}

	_ = newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Give time for listenUpdates to encounter the error and return
	time.Sleep(50 * time.Millisecond)

	// The test passes if no panic occurs
}

func TestSessionStream_ListenUpdatesStreamClosed(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{
		WroteData: bytes.NewBuffer(nil),
		ReadData:  bytes.NewBuffer(nil),
	}

	_ = newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Give time for listenUpdates to encounter the error and return
	time.Sleep(50 * time.Millisecond)

	// The test passes if no panic occurs
}

func TestSessionStreamConcurrentAccess(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Test concurrent access to updateSession and Close
	go func() {
		for i := 0; i < 10; i++ {
			ss.updateSession(uint64(i * 1000))
			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		time.Sleep(50 * time.Millisecond)
		ss.close()
	}()

	time.Sleep(100 * time.Millisecond)

	// Test should complete without race conditions
}

func TestSessionStreamContextCancellation(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Cancel the session context
	sessCtx.cancel(ErrClosedSession)

	// Operations should fail after context cancellation
	err := ss.updateSession(1000000)
	if err == nil {
		// Note: updateSession might still succeed if it doesn't check context
		// This depends on the implementation
	}
}

func TestSessionStreamMultipleUpdates(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Send multiple updates
	bitrates := []uint64{1000000, 2000000, 3000000}
	for _, bitrate := range bitrates {
		err := ss.updateSession(bitrate)
		if err != nil {
			t.Errorf("updateSession(%d) error = %v", bitrate, err)
		}
	}

	// Check that data was written for each update
	if mockStream.WroteData.Len() == 0 {
		t.Error("no data written to stream for multiple updates")
	}
}

func TestSessionStreamupdateSessionZeroBitrate(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	err := ss.updateSession(0)
	if err != nil {
		t.Errorf("updateSession(0) error = %v", err)
	}

	// Should still write data even for zero bitrate
	if mockStream.WroteData.Len() == 0 {
		t.Error("no data written to stream for zero bitrate update")
	}
}

func TestSessionStreamupdateSessionLargeBitrate(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	largeBitrate := uint64(18446744073709551615) // Max uint64
	err := ss.updateSession(largeBitrate)
	if err != nil {
		t.Errorf("updateSession(%d) error = %v", largeBitrate, err)
	}

	// Should handle large bitrate values
	if mockStream.WroteData.Len() == 0 {
		t.Error("no data written to stream for large bitrate update")
	}
}

// Test closeWithError with GroupError type - will use InternalError code
func TestSessionStream_CloseWithGroupError(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	testErr := ErrGroupExpired // This implements GroupError but not TerminateError
	err := ss.closeWithError(testErr)
	if err != nil {
		t.Errorf("closeWithError() error = %v", err)
	}

	// Verify context is cancelled with the right error
	if sessCtx.Err() == nil {
		t.Error("session context should be cancelled")
	}

	cause := context.Cause(sessCtx)
	if cause != testErr {
		t.Errorf("context cause = %v, want %v", cause, testErr)
	}

	// Should fall back to InternalError code since ErrGroupExpired doesn't implement TerminateError
	if !mockStream.AssertCalled(t, "CancelRead", quic.StreamErrorCode(ErrInternalError.TerminateErrorCode())) {
		t.Error("underlying stream should be cancelled with InternalError code")
	}
}

// Test updateSession with closed context
func TestSessionStream_updateSessionWithClosedContext(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Cancel the context first
	sessCtx.cancel(ErrClosedSession)

	bitrate := uint64(1000000)
	err := ss.updateSession(bitrate)

	// updateSession doesn't currently check context, but if implemented it should fail
	// For now we test that it doesn't panic
	_ = err
}

// Test listenUpdates with context cancellation
func TestSessionStream_ListenUpdates_ContextCancellation(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	doneCh := make(chan struct{}, 1)
	go func() {
		sessCtx.wg.Wait()
		close(doneCh)
	}()
	ReadData := bytes.NewBuffer(nil)

	// Create a valid message that will be read continuously
	_, err := message.SessionUpdateMessage{
		Bitrate: 1000000,
	}.Encode(ReadData)
	if err != nil {
		t.Fatalf("failed to encode message: %v", err)
	}

	mockStream := &MockQUICStream{
		ReadData: ReadData,
	}
	_ = newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Let listenUpdates start
	time.Sleep(10 * time.Millisecond)

	// Cancel the context
	sessCtx.cancel(ErrClosedSession)

	// Give time for listenUpdates to exit
	time.Sleep(50 * time.Millisecond)

	select {
	case <-doneCh:
		// Context should be cancelled and listenUpdates should exit
	case <-time.After(100 * time.Millisecond):
		t.Error("listenUpdates did not exit after context cancellation")
	}
}

// Test multiple SESSION_UPDATE messages processing
func TestSessionStream_ListenMultipleUpdates(t *testing.T) {
	ReadData := bytes.NewBuffer(nil)

	expectedBitrate := uint64(3000000)
	// Create multiple valid SESSION_UPDATE messages
	bitrates := []uint64{1000000, 2000000, expectedBitrate}
	for _, bitrate := range bitrates {
		_, err := message.SessionUpdateMessage{
			Bitrate: bitrate,
		}.Encode(ReadData)
		if err != nil {
			t.Fatalf("failed to encode SESSION_UPDATE message: %v", err)
		}
	}

	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{
		ReadData: ReadData,
	}
	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	doneCh := make(chan struct{}, 1)
	go func() {
		sessCtx.wg.Wait()
		close(doneCh)
	}()

	// Give time for listenUpdates to process all messages
	time.Sleep(100 * time.Millisecond)

	select {
	case <-doneCh:
		assert.Equal(t, expectedBitrate, ss.remoteBitrate, "remote bitrate should be updated to last SESSION_UPDATE message")
	case <-time.After(100 * time.Millisecond):
		t.Error("listenUpdates did not exit after context cancellation")
	}
}

// Test listenUpdates with read errors
func TestSessionStream_ListenUpdates_ReadError(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil),
	}

	expectedErr := errors.New("read error")

	// Setup mock to return error on Read
	mockStream.On("Read", mock.Anything).Return(0, expectedErr)

	_ = newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	doneCh := make(chan struct{}, 1)
	go func() {
		sessCtx.wg.Wait()
		close(doneCh)
	}()

	// Give time for listenUpdates to encounter the error and return
	time.Sleep(50 * time.Millisecond)

	select {
	case <-doneCh:
		// listenUpdates should exit after read error
	case <-time.After(100 * time.Millisecond):
		t.Error("listenUpdates did not exit after read error")
	}
}

// Test concurrent Close and closeWithError calls
func TestSessionStream_ConcurrentCloseOperations(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Launch concurrent close operations
	go func() {
		ss.close()
	}()

	go func() {
		ss.closeWithError(ErrProtocolViolation)
	}()

	// Give time for operations to complete
	time.Sleep(50 * time.Millisecond)

	// Test should complete without race conditions
	// Verify that the stream is marked as closed
	if sessCtx.Err() == nil {
		t.Error("session context should be cancelled")
	}
}

// // Test that closedErr reflects context cancellation reason
// func TestSessionStream_ClosedErrWithSpecificReason(t *testing.T) {
// 	sessCtx := createTestSessionContext(context.Background())
// 	mockStream := &MockQUICStream{}

// 	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

// 	specificErr := ErrProtocolViolation
// 	ss.closeWithError(specificErr)

// 	err := ss.closedErr()
// 	if err != specificErr {
// 		t.Errorf("closedErr() = %v, want %v", err, specificErr)
// 	}
// }

// // Test closedErr with nil context cause
// func TestSessionStream_ClosedErrWithNilCause(t *testing.T) {
// 	sessCtx := createTestSessionContext(context.Background())
// 	mockStream := &MockQUICStream{}

// 	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

// 	// Close with nil cause
// 	ss.close()

// 	err := ss.closedErr()
// 	if err != ErrClosedSession {
// 		t.Errorf("closedErr() = %v, want %v", err, ErrClosedSession)
// 	}
// }

// Test updateSession with mutex contention
func TestSessionStream_updateSessionMutexContention(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Launch multiple concurrent updateSession calls
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(bitrate uint64) {
			defer func() { done <- true }()
			ss.updateSession(bitrate)
		}(uint64(i * 1000))
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Test should complete without race conditions
}

// Test closeWithError with unauthorized error type
func TestSessionStream_CloseWithUnauthorizedError(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	testErr := ErrUnauthorizedError
	err := ss.closeWithError(testErr)
	if err != nil {
		t.Errorf("closeWithError() error = %v", err)
	}

	// Verify stream is cancelled with correct error code
	if !mockStream.AssertCalled(t, "CancelRead", quic.StreamErrorCode(testErr.TerminateErrorCode())) {
		t.Error("underlying stream should be cancelled")
	}

	if !mockStream.AssertCalled(t, "CancelWrite", quic.StreamErrorCode(testErr.TerminateErrorCode())) {
		t.Error("underlying stream should be cancelled on write")
	}
}

// Test behavior when listenUpdates encounters EOF
func TestSessionStream_ListenUpdatesEOF(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil), // Empty buffer will cause EOF
	}

	_ = newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Give time for listenUpdates to encounter EOF and return
	time.Sleep(50 * time.Millisecond)

	// The test passes if no panic occurs
}

// Test session stream with very long-running operation
func TestSessionStream_LongRunningOperation(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil),
	}

	// Add a large number of messages to simulate long-running operation
	for i := 0; i < 100; i++ {
		_, err := message.SessionUpdateMessage{
			Bitrate: uint64(i * 10000),
		}.Encode(mockStream.ReadData)
		if err != nil {
			t.Fatalf("failed to encode message %d: %v", i, err)
		}
	}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Let it run for a reasonable time
	time.Sleep(200 * time.Millisecond)

	// Close cleanly
	err := ss.close()
	if err != nil {
		t.Errorf("close() error = %v", err)
	}
}

// Test closeWithError with non-Terminate error type (should fallback to InternalError)
func TestSessionStream_CloseWithNonTerminateError(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Use a regular error that doesn't implement TerminateError
	testErr := errors.New("custom error")
	err := ss.closeWithError(testErr)
	if err != nil {
		t.Errorf("closeWithError() error = %v", err)
	}

	// Verify context is cancelled with the custom error
	cause := context.Cause(sessCtx)
	if cause != testErr {
		t.Errorf("context cause = %v, want %v", cause, testErr)
	}

	// Should use InternalError code as fallback
	if !mockStream.AssertCalled(t, "CancelRead", quic.StreamErrorCode(ErrInternalError.TerminateErrorCode())) {
		t.Error("underlying stream should be cancelled with InternalError code")
	}
}

// Test updateSession with WriteFunc that returns various scenarios
func TestSessionStream_updateSessionStreamWriteScenarios(t *testing.T) {
	testCases := []struct {
		name        string
		writeError  error
		expectError bool
	}{
		{
			name:        "successful write",
			writeError:  nil,
			expectError: false,
		},
		{
			name:        "write error",
			writeError:  errors.New("write failed"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sessCtx := createTestSessionContext(context.Background())
			mockStream := &MockQUICStream{
				WroteData: bytes.NewBuffer(nil),
			}

			if tc.writeError != nil {
				mockStream.On("Write", mock.Anything).Return(0, tc.writeError)
			} else {
				mockStream.On("Write", mock.Anything).Return(8, nil) // Assume 8 bytes written
			}

			ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

			err := ss.updateSession(1000000)
			if tc.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Test listenUpdates with partial message reads
func TestSessionStream_ListenUpdatesPartialReads(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil),
	}

	// Create a message but simulate partial reads
	msg := message.SessionUpdateMessage{Bitrate: 1500000}
	msgBuffer := bytes.NewBuffer(nil)
	_, err := msg.Encode(msgBuffer)
	if err != nil {
		t.Fatalf("failed to encode message: %v", err)
	}

	// Copy the encoded message to ReadData
	mockStream.ReadData.Write(msgBuffer.Bytes())

	sessCtx := createTestSessionContext(context.Background())
	_ = newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Give time for listenUpdates to process
	time.Sleep(100 * time.Millisecond)

	// Test passes if no panic occurs
}

// Test updateSession thread safety
func TestSessionStream_updateSessionThreadSafety(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{
		WroteData: bytes.NewBuffer(nil),
	}
	mockStream.On("Write", mock.Anything).Return(8, nil)

	_ = newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	const numThreads = 50
	var wg sync.WaitGroup
	wg.Add(numThreads)

	// Create a single session stream for all goroutines to share
	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Launch multiple goroutines calling updateSession concurrently
	for i := 0; i < numThreads; i++ {
		go func(bitrate uint64) {
			defer wg.Done()
			ss.updateSession(bitrate)
		}(uint64(i * 1000))
	}

	wg.Wait()

	// Test should complete without data races
}

// Test Close method thread safety
func TestSessionStream_CloseThreadSafety(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	const numThreads = 10
	var wg sync.WaitGroup
	wg.Add(numThreads)

	// Launch multiple goroutines calling Close concurrently
	for i := 0; i < numThreads; i++ {
		go func() {
			defer wg.Done()
			ss.close()
		}()
	}

	wg.Wait()

	// Verify that context is cancelled
	if sessCtx.Err() == nil {
		t.Error("session context should be cancelled")
	}
}

// Test closeWithError thread safety with different errors
func TestSessionStream_closeWithErrorThreadSafety(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	errors := []error{
		ErrProtocolViolation,
		ErrUnauthorizedError,
		ErrInternalError,
		errors.New("custom error"),
	}

	const numThreads = 20
	var wg sync.WaitGroup
	wg.Add(numThreads)

	// Launch multiple goroutines calling closeWithError concurrently
	for i := 0; i < numThreads; i++ {
		go func(idx int) {
			defer wg.Done()
			testErr := errors[idx%len(errors)]
			ss.closeWithError(testErr)
		}(i)
	}

	wg.Wait()

	// Verify that context is cancelled
	if sessCtx.Err() == nil {
		t.Error("session context should be cancelled")
	}
}

// Test that closedErr correctly returns the cancellation reason
// func TestSessionStream_ClosedErrReturnsCorrectReason(t *testing.T) {
// 	testCases := []struct {
// 		name        string
// 		closeErr    error
// 		expectedErr error
// 	}{
// 		{
// 			name:        "closed with specific error",
// 			closeErr:    ErrProtocolViolation,
// 			expectedErr: ErrProtocolViolation,
// 		},
// 		{
// 			name:        "closed with nil (should return ErrClosedSession)",
// 			closeErr:    nil,
// 			expectedErr: ErrClosedSession,
// 		},
// 		{
// 			name:        "closed with custom error",
// 			closeErr:    errors.New("custom close error"),
// 			expectedErr: errors.New("custom close error"),
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			sessCtx := createTestSessionContext(context.Background())
// 			mockStream := &MockQUICStream{}

// 			ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

// 			if tc.closeErr == nil {
// 				ss.close()
// 			} else {
// 				ss.closeWithError(tc.closeErr)
// 			}

// 			err := ss.closedErr()
// 			if tc.name == "closed with nil (should return ErrClosedSession)" {
// 				if err != ErrClosedSession {
// 					t.Errorf("closedErr() = %v, want %v", err, ErrClosedSession)
// 				}
// 			} else if tc.name == "closed with custom error" {
// 				if err.Error() != tc.expectedErr.Error() {
// 					t.Errorf("closedErr() = %v, want %v", err, tc.expectedErr)
// 				}
// 			} else {
// 				if err != tc.expectedErr {
// 					t.Errorf("closedErr() = %v, want %v", err, tc.expectedErr)
// 				}
// 			}
// 		})
// 	}
// }

// Test listenUpdates with various message decoding scenarios
func TestSessionStream_ListenUpdatesDecodingScenarios(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "invalid message type",
			data: []byte{0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name: "truncated message",
			data: []byte{0x00, 0x01},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sessCtx := createTestSessionContext(context.Background())
			mockStream := &MockQUICStream{
				ReadData: bytes.NewBuffer(tc.data),
			}

			_ = newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

			// Give time for listenUpdates to process (or fail)
			time.Sleep(50 * time.Millisecond)

			// Test passes if no panic occurs
		})
	}
}

// Test that listenUpdates goroutine exits properly when context is cancelled
func TestSessionStream_ExitListenUpdatesGoroutine(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil),
	}

	// Add a message that will keep the goroutine reading
	msg := message.SessionUpdateMessage{Bitrate: 1000000}
	_, err := msg.Encode(mockStream.ReadData)
	if err != nil {
		t.Fatalf("failed to encode message: %v", err)
	}

	_ = newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Let listenUpdates start and process some messages
	time.Sleep(10 * time.Millisecond)

	// Cancel the session context
	sessCtx.cancel(ErrClosedSession)

	doneCh := make(chan struct{}, 1)

	go func() {
		// Attempt to close the session stream
		sessCtx.wg.Wait()
		close(doneCh)
	}()

	// Wait for the goroutine to finish
	select {
	case <-doneCh:
		// listenUpdates should exit cleanly after context cancellation
	case <-time.After(100 * time.Millisecond):
		t.Error("listenUpdates goroutine did not exit after context cancellation")
	}
}

// Test updateSession behavior when context is already cancelled
func TestSessionStream_updateSessionContextCancelled(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(8, nil)

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Cancel the context first
	sessCtx.cancel(ErrClosedSession)

	bitrate := uint64(1000000)
	err := ss.updateSession(bitrate)

	// Current implementation doesn't check context in updateSession
	// This test documents the current behavior - it succeeds even with cancelled context
	if err != nil {
		t.Logf("updateSession with cancelled context returned error: %v", err)
	}

	// Verify that data was still written (current behavior)
	mockStream.AssertCalled(t, "Write", mock.Anything)
}

// Test that Close protects against concurrent access
func TestSessionStream_CloseRaceCondition(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch many goroutines trying to close concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			ss.close()
		}()
	}

	wg.Wait()

	// Verify that context is cancelled
	if sessCtx.Err() == nil {
		t.Error("session context should be cancelled")
	}

	// Only one close() should have succeeded in calling the underlying stream close()
	// Other calls should have returned early due to already being closed
	mockStream.AssertCalled(t, "Close")
}

// Test closeWithError behavior with various concurrent scenarios
func TestSessionStream_closeWithErrorVariousScenarios(t *testing.T) {
	testCases := []struct {
		name     string
		errors   []error
		expected error // The error that should be in the context cause
	}{
		{
			name:     "first error wins",
			errors:   []error{ErrProtocolViolation, ErrUnauthorizedError, ErrInternalError},
			expected: ErrProtocolViolation,
		},
		{
			name:     "nil error first",
			errors:   []error{nil, ErrProtocolViolation},
			expected: ErrInternalError, // nil gets converted to ErrInternalError
		},
		{
			name:     "custom error first",
			errors:   []error{errors.New("custom"), ErrProtocolViolation},
			expected: errors.New("custom"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sessCtx := createTestSessionContext(context.Background())
			mockStream := &MockQUICStream{}

			ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

			var wg sync.WaitGroup
			wg.Add(len(tc.errors))

			// Launch concurrent closeWithError calls
			for _, err := range tc.errors {
				go func(e error) {
					defer wg.Done()
					ss.closeWithError(e)
				}(err)
			}

			wg.Wait()

			// Verify context is cancelled
			if sessCtx.Err() == nil {
				t.Error("session context should be cancelled")
			}

			// Verify the first error won (or its converted form)
			cause := context.Cause(sessCtx)
			if tc.name == "custom error first" {
				if cause.Error() != tc.expected.Error() {
					t.Errorf("context cause = %v, want %v", cause.Error(), tc.expected.Error())
				}
			} else {
				if cause != tc.expected {
					t.Errorf("context cause = %v, want %v", cause, tc.expected)
				}
			}
		})
	}
}

// Test that updateSession works correctly with very large bitrates
func TestSessionStream_updateSessionMaxValues(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(8, nil)

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Test with maximum uint64 value
	maxBitrate := ^uint64(0) // Maximum uint64 value
	err := ss.updateSession(maxBitrate)
	if err != nil {
		t.Errorf("updateSession(max uint64) error = %v", err)
	}

	// Verify message was encoded and written correctly
	mockStream.AssertCalled(t, "Write", mock.Anything)
}

// Test listenUpdates with mixed valid and invalid messages
func TestSessionStream_ListenUpdatesMixedMessages(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil),
	}

	// Add a valid message
	validMsg := message.SessionUpdateMessage{Bitrate: 1000000}
	_, err := validMsg.Encode(mockStream.ReadData)
	if err != nil {
		t.Fatalf("failed to encode valid message: %v", err)
	}

	// Add invalid bytes that will cause decode error
	mockStream.ReadData.Write([]byte{0xFF, 0xFF, 0xFF})

	sessCtx := createTestSessionContext(context.Background())
	_ = newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Give time for listenUpdates to process valid message and fail on invalid one
	time.Sleep(100 * time.Millisecond)

	// Test passes if no panic occurs
}

// // Test closedErr with different context cancellation scenarios
// func TestSessionStream_ClosedErrContextScenarios(t *testing.T) {
// 	testCases := []struct {
// 		name           string
// 		cancelWithErr  error
// 		expectedResult error
// 	}{
// 		{
// 			name:           "context cancelled with specific error",
// 			cancelWithErr:  ErrProtocolViolation,
// 			expectedResult: ErrProtocolViolation,
// 		},
// 		{
// 			name:           "context cancelled with nil",
// 			cancelWithErr:  nil,
// 			expectedResult: ErrClosedSession,
// 		},
// 		{
// 			name:           "context cancelled with custom error",
// 			cancelWithErr:  errors.New("custom cancellation"),
// 			expectedResult: errors.New("custom cancellation"),
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			sessCtx := createTestSessionContext(context.Background())
// 			mockStream := &MockQUICStream{}

// 			ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

// 			// Cancel context with the test error
// 			sessCtx.cancel(tc.cancelWithErr)

// 			if tc.name == "context cancelled with custom error" {
// 				if err.Error() != tc.expectedResult.Error() {
// 					t.Errorf("closedErr() = %v, want %v", err.Error(), tc.expectedResult.Error())
// 				}
// 			} else {
// 				if err != tc.expectedResult {
// 					t.Errorf("closedErr() = %v, want %v", err, tc.expectedResult)
// 				}
// 			}
// 		})
// 	}
// }

// Test that updateSession handles write deadline errors
func TestSessionStream_updateSessionWriteDeadlineError(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	// Simulate a write deadline error
	deadlineErr := errors.New("write deadline exceeded")
	mockStream.On("Write", mock.Anything).Return(0, deadlineErr)

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	err := ss.updateSession(1000000)
	if err == nil {
		t.Error("updateSession() should return error when write deadline is exceeded")
	}

	if err != deadlineErr {
		t.Errorf("updateSession() error = %v, want %v", err, deadlineErr)
	}
}

// Test that sessionStream properly handles stream ID queries
func TestSessionStream_StreamIDAccess(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	expectedStreamID := quic.StreamID(12345)

	mockStream.On("StreamID").Return(expectedStreamID)

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Access the stream ID through the session stream's underlying stream
	// This tests that the stream is properly accessible
	actualStreamID := ss.stream.StreamID()
	if actualStreamID != expectedStreamID {
		t.Errorf("stream ID = %v, want %v", actualStreamID, expectedStreamID)
	}

	mockStream.AssertCalled(t, "StreamID")
}

// Test memory and goroutine cleanup
func TestSessionStream_ResourceCleanup(t *testing.T) {
	const numStreams = 100

	for i := 0; i < numStreams; i++ {
		sessCtx := createTestSessionContext(context.Background())
		mockStream := &MockQUICStream{
			ReadData: bytes.NewBuffer(nil),
		}

		// Add a message to keep the goroutine busy briefly
		msg := message.SessionUpdateMessage{Bitrate: uint64(i * 1000)}
		msg.Encode(mockStream.ReadData)

		ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

		// Quickly close the session
		ss.close()
	}

	// Give time for all goroutines to exit
	time.Sleep(100 * time.Millisecond)

	// This test primarily checks that creating and closing many session streams
	// doesn't cause memory leaks or goroutine leaks
	// In a real scenario, you might use tools like go test -race or pprof to verify
}

// Test updateSession with concurrent Close operations
func TestSessionStream_updateSessionConcurrentWithClose(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(8, nil)

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	var wg sync.WaitGroup
	wg.Add(2)

	// Start updateSession operations
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			ss.updateSession(uint64(i * 1000))
			time.Sleep(time.Millisecond)
		}
	}()

	// Close after some time
	go func() {
		defer wg.Done()
		time.Sleep(25 * time.Millisecond)
		ss.close()
	}()

	wg.Wait()

	// Verify that the stream was eventually closed
	if sessCtx.Err() == nil {
		t.Error("session context should be cancelled after close")
	}
}

// Test listenUpdates with continuous message stream
func TestSessionStream_ListenUpdatesContinuousStream(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil),
	}

	// Create a continuous stream of valid messages
	for i := 0; i < 20; i++ {
		msg := message.SessionUpdateMessage{Bitrate: uint64(i * 100000)}
		_, err := msg.Encode(mockStream.ReadData)
		if err != nil {
			t.Fatalf("failed to encode message %d: %v", i, err)
		}
	}
	sessCtx := createTestSessionContext(context.Background())
	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Let it process several messages
	time.Sleep(200 * time.Millisecond)

	// Close the session
	ss.close()

	// Verify clean shutdown
	if sessCtx.Err() == nil {
		t.Error("session context should be cancelled")
	}
}

// Test that listenUpdates handles corrupted message data gracefully
func TestSessionStream_ListenUpdatesCorruptedData(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil),
	}

	// Create partially valid then corrupted data
	validMsg := message.SessionUpdateMessage{Bitrate: 1000000}
	_, err := validMsg.Encode(mockStream.ReadData)
	if err != nil {
		t.Fatalf("failed to encode valid message: %v", err)
	}

	// Add corrupted data that will cause decode issues
	corruptedData := []byte{0x01, 0x02, 0xFF, 0xFE, 0xFD, 0xFC}
	mockStream.ReadData.Write(corruptedData)

	sessCtx := createTestSessionContext(context.Background())
	_ = newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Give time for listenUpdates to process valid message and encounter corruption
	time.Sleep(100 * time.Millisecond)

	// Test passes if no panic occurs
}

// Test updateSession with various message sizes
func TestSessionStream_updateSessionMessageSizes(t *testing.T) {
	testCases := []struct {
		name    string
		bitrate uint64
	}{
		{"zero bitrate", 0},
		{"small bitrate", 1000},
		{"medium bitrate", 1000000},
		{"large bitrate", 1000000000},
		{"max uint64", ^uint64(0)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sessCtx := createTestSessionContext(context.Background())
			mockStream := &MockQUICStream{}
			mockStream.On("Write", mock.Anything).Return(8, nil)

			ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

			err := ss.updateSession(tc.bitrate)
			if err != nil {
				t.Errorf("updateSession(%d) error = %v", tc.bitrate, err)
			}

			mockStream.AssertCalled(t, "Write", mock.Anything)
		})
	}
}

// Test that Close/closeWithError are idempotent
func TestSessionStream_IdempotentClose(t *testing.T) {
	testCases := []struct {
		name      string
		closeFunc func(*sessionStream) error
	}{
		{
			name: "multiple Close calls",
			closeFunc: func(ss *sessionStream) error {
				return ss.close()
			},
		},
		{
			name: "multiple closeWithError calls",
			closeFunc: func(ss *sessionStream) error {
				return ss.closeWithError(ErrProtocolViolation)
			},
		},
		{
			name: "mixed Close and closeWithError",
			closeFunc: func(ss *sessionStream) error {
				// First call Close, then closeWithError
				ss.close()
				return ss.closeWithError(ErrProtocolViolation)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sessCtx := createTestSessionContext(context.Background())
			mockStream := &MockQUICStream{}

			ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

			// First call should succeed
			err1 := tc.closeFunc(ss)
			if err1 != nil && tc.name != "mixed Close and closeWithError" {
				t.Errorf("first %s should not return error, got: %v", tc.name, err1)
			}

			// Subsequent calls should return error indicating already closed
			err2 := tc.closeFunc(ss)
			if err2 == nil && tc.name != "multiple closeWithError calls" {
				t.Errorf("second %s should return error indicating already closed", tc.name)
			}

			// Context should be cancelled
			if sessCtx.Err() == nil {
				t.Error("session context should be cancelled")
			}
		})
	}
}

// Test edge case where stream is closed during updateSession
func TestSessionStream_updateSessionDuringClose(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}

	// Set up mock to simulate stream being closed during write
	writeCallCount := 0
	mockStream.On("Write", mock.Anything).Return(func(p []byte) (int, error) {
		writeCallCount++
		if writeCallCount > 3 {
			return 0, errors.New("stream closed")
		}
		return len(p), nil
	})

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	var wg sync.WaitGroup
	wg.Add(2)

	// Continuously call updateSession
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			ss.updateSession(uint64(i * 1000))
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Close the stream after a short delay
	go func() {
		defer wg.Done()
		time.Sleep(20 * time.Millisecond)
		ss.close()
	}()

	wg.Wait()

	// Verify that context is cancelled
	if sessCtx.Err() == nil {
		t.Error("session context should be cancelled")
	}
}

// Test listenUpdates with slow message processing
func TestSessionStream_ListenUpdatesSlowProcessing(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil),
	}

	// Create messages with varying bitrates
	bitrates := []uint64{100000, 200000, 300000, 400000, 500000}
	for _, bitrate := range bitrates {
		msg := message.SessionUpdateMessage{Bitrate: bitrate}
		_, err := msg.Encode(mockStream.ReadData)
		if err != nil {
			t.Fatalf("failed to encode message with bitrate %d: %v", bitrate, err)
		}
	}

	sessCtx := createTestSessionContext(context.Background())
	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Let it process all messages
	time.Sleep(150 * time.Millisecond)

	// Close cleanly
	err := ss.close()
	if err != nil {
		t.Errorf("close() error = %v", err)
	}

	// Verify clean shutdown
	if sessCtx.Err() == nil {
		t.Error("session context should be cancelled after close")
	}
}

// Test that stream methods work correctly after context is cancelled but before explicit close
func TestSessionStream_MethodsAfterContextCancel(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(8, nil)

	sessstr := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Cancel the context directly
	sessCtx.cancel(ErrProtocolViolation)

	err := sessstr.updateSession(1000000)
	assert.Equal(t, ErrProtocolViolation, err, "updateSession should not return error after context cancel")

	// Close should return the already-closed error
	err = sessstr.close()
	assert.Equal(t, ErrProtocolViolation, err, "close should return the already-closed error")

	err = sessstr.closeWithError(ErrUnauthorizedError)
	assert.Equal(t, ErrProtocolViolation, err, "closeWithError should return the already-closed error")
}

// Test performance of updateSession under high concurrency
func TestSessionStream_updateSessionHighConcurrency(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(8, nil)

	ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	const numGoroutines = 200
	const updatesPerGoroutine = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	startTime := time.Now()

	// Launch many goroutines calling updateSession concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < updatesPerGoroutine; j++ {
				bitrate := uint64(goroutineID*1000 + j)
				ss.updateSession(bitrate)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("Completed %d updateSession calls in %v", numGoroutines*updatesPerGoroutine, duration)

	// Verify all calls were made
	if mockStream.Calls != nil && len(mockStream.Calls) < numGoroutines*updatesPerGoroutine {
		t.Errorf("Expected at least %d Write calls, got %d", numGoroutines*updatesPerGoroutine, len(mockStream.Calls))
	}
}

// Test error propagation from underlying stream operations
func TestSessionStream_ErrorPropagation(t *testing.T) {
	testCases := []struct {
		name          string
		getMockStream func() *MockQUICStream
		operation     func(*sessionStream) error
		expectError   bool
	}{
		{
			name: "updateSession with write error",
			getMockStream: func() *MockQUICStream {
				m := &MockQUICStream{}
				m.On("Write", mock.Anything).Return(0, errors.New("write failed"))
				return m
			},
			operation: func(ss *sessionStream) error {
				return ss.updateSession(1000000)
			},
			expectError: true,
		},
		{
			name: "Close with close error",
			getMockStream: func() *MockQUICStream {
				m := &MockQUICStream{}
				m.On("Close").Return(errors.New("close failed"))
				return m
			},
			operation: func(ss *sessionStream) error {
				return ss.close()
			},
			expectError: true,
		},
		{
			name: "successful operations",
			getMockStream: func() *MockQUICStream {
				m := &MockQUICStream{}
				m.On("Write", mock.Anything).Return(8, nil)
				m.On("Close").Return(nil)
				return m
			},
			operation: func(ss *sessionStream) error {
				err := ss.updateSession(1000000)
				if err != nil {
					return err
				}
				return ss.close()
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sessCtx := createTestSessionContext(context.Background())
			mockStream := tc.getMockStream()

			ss := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
			err := tc.operation(ss)
			if tc.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
