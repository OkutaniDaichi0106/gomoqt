package moqt

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/moqtrace"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper function to create a properly configured mockQUICConnection for testing
func createMockQUICConnection() *MockQUICConnection {
	conn := &MockQUICConnection{}
	conn.On("LocalAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8081})
	conn.On("Context").Return(context.Background())
	conn.On("ConnectionState").Return(quic.ConnectionState{})
	conn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("OpenStream").Return(nil, errors.New("not implemented"))
	conn.On("OpenUniStream").Return(nil, errors.New("not implemented"))
	conn.On("OpenStreamSync", mock.Anything).Return(nil, errors.New("not implemented"))
	conn.On("OpenUniStreamSync", mock.Anything).Return(nil, errors.New("not implemented"))
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	return conn
}

// Helper function to create a mockQUICConnection that returns errors on open operations
func createMockQUICConnectionWithOpenError(err error) *MockQUICConnection {
	conn := createMockQUICConnection()
	conn.ExpectedCalls = nil // Clear previous expectations
	conn.On("LocalAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8081})
	conn.On("Context").Return(context.Background())
	conn.On("ConnectionState").Return(quic.ConnectionState{})
	conn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("OpenStream").Return(nil, err)
	conn.On("OpenUniStream").Return(nil, err)
	conn.On("OpenStreamSync", mock.Anything).Return(nil, err)
	conn.On("OpenUniStreamSync", mock.Anything).Return(nil, err)
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	return conn
}

// Helper function to create a mockQUICConnection that returns errors on accept operations
func createMockQUICConnectionWithAcceptError(err error) *MockQUICConnection {
	conn := createMockQUICConnection()
	conn.ExpectedCalls = nil // Clear previous expectations
	conn.On("LocalAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8081})
	conn.On("Context").Return(context.Background())
	conn.On("ConnectionState").Return(quic.ConnectionState{})
	conn.On("AcceptStream", mock.Anything).Return(nil, err)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, err)
	conn.On("OpenStream").Return(nil, errors.New("not implemented"))
	conn.On("OpenUniStream").Return(nil, errors.New("not implemented"))
	conn.On("OpenStreamSync", mock.Anything).Return(nil, errors.New("not implemented"))
	conn.On("OpenUniStreamSync", mock.Anything).Return(nil, errors.New("not implemented"))
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	return conn
}

func TestNewSession(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())

	// Create a proper MockQUICStream
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))

	// Create a proper mock connection
	conn := createMockQUICConnection()

	mux := NewTrackMux()

	session := newSession(sessCtx, stream, conn, mux)

	if session == nil {
		t.Fatal("newSession returned nil")
	}

	if session.ctx != sessCtx {
		t.Error("session context not set correctly")
	}

	if session.mux != mux {
		t.Error("mux not set correctly")
	}

	if session.sessionStream != stream {
		t.Error("session stream not set correctly")
	}

	if session.receiveGroupStreamQueues == nil {
		t.Error("receive group stream queues should not be nil")
	}

	if session.sendGroupStreamQueues == nil {
		t.Error("send group stream queues should not be nil")
	}

	// Give time for goroutines to start
	time.Sleep(20 * time.Millisecond)

	// Cleanup
	session.Terminate(nil)
}

func TestNewSessionWithNilMux(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnection()

	session := newSession(sessCtx, stream, conn, nil)

	if session.mux != DefaultMux {
		t.Error("should use DefaultMux when nil mux is provided")
	}

	// Cleanup
	session.Terminate(nil)
}

func TestSessionTerminate(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnection()
	mux := NewTrackMux()

	session := newSession(sessCtx, stream, conn, mux)

	// Give time for goroutines to start
	time.Sleep(20 * time.Millisecond)

	testErr := ErrProtocolViolation
	session.Terminate(testErr)

	// Verify context is cancelled
	if sessCtx.Err() == nil {
		t.Error("session context should be cancelled after terminate")
	}

	// Verify CloseWithError was called
	conn.AssertCalled(t, "CloseWithError", mock.Anything, mock.Anything)
}

func TestSessionTerminateWithNilReason(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnection()

	session := newSession(sessCtx, stream, conn, nil)

	// Give time for goroutines to start
	time.Sleep(20 * time.Millisecond)

	session.Terminate(nil)

	// Should use NoErrTerminate when nil reason is provided
	conn.AssertCalled(t, "CloseWithError", mock.Anything, mock.Anything)
}

func TestSessionOpenAnnounceStream(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnection()

	session := newSession(sessCtx, stream, conn, nil)

	prefix := "/test/prefix"
	announceReader, err := session.OpenAnnounceStream(prefix)
	if err != nil {
		t.Errorf("OpenAnnounceStream() error = %v", err)
	}

	if announceReader == nil {
		t.Error("OpenAnnounceStream() returned nil reader")
	}

	// Cleanup
	announceReader.Close()
	session.Terminate(nil)
}

func TestSessionOpenAnnounceStreamInvalidPrefix(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnection()

	session := newSession(sessCtx, stream, conn, nil)

	// Should panic for prefix not starting with '/'
	defer func() {
		if r := recover(); r == nil {
			t.Error("OpenAnnounceStream() should panic for invalid prefix")
		}
	}()

	session.OpenAnnounceStream("invalid-prefix")
}

func TestSessionOpenAnnounceStreamOpenError(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnectionWithOpenError(errors.New("open stream error"))

	session := newSession(sessCtx, stream, conn, nil)

	prefix := "/test/prefix"
	announceReader, err := session.OpenAnnounceStream(prefix)
	if err == nil {
		t.Error("OpenAnnounceStream() should return error when stream opening fails")
	}

	if announceReader != nil {
		t.Error("OpenAnnounceStream() should return nil reader on error")
	}

	// Cleanup
	session.Terminate(nil)
}

func TestSession_OpenTrackStream(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	// Set up expectations needed for sessionStream.listenUpdates
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	mockStream.On("Close").Return(nil)
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnection()

	session := newSession(sessCtx, stream, conn, nil)

	path := BroadcastPath("/test/track")
	name := TrackName("video")
	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}

	subscriber, err := session.OpenTrackStream(path, name, config)
	if err != nil {
		t.Errorf("OpenTrackStream() error = %v", err)
	}

	assert.NotNil(t, subscriber, "OpenTrackStream() should return a valid subscriber")

	assert.Equal(t, subscriber.BroadcastPath, path, "subscriber path = %v, want %v", subscriber.BroadcastPath, path)
	assert.Equal(t, subscriber.TrackName, name, "subscriber name = %v, want %v", subscriber.TrackName, name)
	gotConfig := subscriber.SubscribeStream.SubscribeConfig()
	assert.Equal(t, gotConfig, config, "subscriber config = %v, want %v", gotConfig, config)

	// Cleanup
	session.Terminate(nil)
}

func TestSessionOpenTrackStreamOpenError(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnectionWithOpenError(errors.New("open stream error"))

	session := newSession(sessCtx, stream, conn, nil)

	path := BroadcastPath("/test/track")
	name := TrackName("video")

	subscriber, err := session.OpenTrackStream(path, name, nil)
	if err == nil {
		t.Error("OpenTrackStream() should return error when stream opening fails")
	}

	if subscriber != nil {
		t.Error("OpenTrackStream() should return nil subscriber on error")
	}

	// Cleanup
	session.Terminate(nil)
}

func TestSessionContext(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnection()

	session := newSession(sessCtx, stream, conn, nil)

	ctx := session.Context()
	if ctx != sessCtx.Context {
		t.Error("Session.Context() should return session context")
	}

	// Cleanup
	session.Terminate(nil)
}

func TestSession_NextSubscribeID(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	// Set up expectations needed for sessionStream.listenUpdates
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnection()

	session := newSession(sessCtx, stream, conn, nil)

	id1 := session.nextSubscribeID()
	id2 := session.nextSubscribeID()

	if id1 == id2 {
		t.Error("nextSubscribeID() should return unique IDs")
	}

	if uint64(id2) != uint64(id1)+1 {
		t.Errorf("nextSubscribeID() should increment, got %v after %v", id2, id1)
	}

	// Cleanup
	session.Terminate(nil)
}

func TestSession_UpdateSession(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	sessstr := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnection()

	session := newSession(sessCtx, sessstr, conn, nil)

	expectedBitrate := uint64(2000000)
	err := session.updateSession(expectedBitrate)
	if err != nil {
		t.Errorf("updateSession() error = %v", err)
	}

	// Verify bitrate is updated
	assert.Equal(t, expectedBitrate, session.sessionStream.localBitrate, "session bitrate should be updated")

	// Cleanup
	session.Terminate(nil)
}

func TestSessionHandleBiStreamsAcceptError(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnectionWithAcceptError(errors.New("accept error"))

	session := newSession(sessCtx, stream, conn, nil)

	// Give time for handleBiStreams to encounter the error
	time.Sleep(50 * time.Millisecond)

	// Cleanup - session should handle the error gracefully
	session.Terminate(nil)
}

func TestSessionHandleUniStreamsAcceptError(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnectionWithAcceptError(errors.New("accept error"))

	session := newSession(sessCtx, stream, conn, nil)

	// Give time for handleUniStreams to encounter the error
	time.Sleep(50 * time.Millisecond)

	// Cleanup - session should handle the error gracefully
	session.Terminate(nil)
}

func TestSessionConcurrentAccess(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnection()

	session := newSession(sessCtx, stream, conn, nil)

	// Test concurrent access to various methods
	go func() {
		for i := 0; i < 5; i++ {
			session.nextSubscribeID()
			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		for i := 0; i < 5; i++ {
			session.updateSession(uint64(i * 1000))
			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		time.Sleep(50 * time.Millisecond)
		session.Terminate(nil)
	}()

	time.Sleep(100 * time.Millisecond)

	// Test should complete without race conditions
}

func TestSessionContextCancellation(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnection()

	session := newSession(sessCtx, stream, conn, nil)

	// Cancel the session context
	sessCtx.cancel(ErrClosedSession)

	// Give time for goroutines to react to cancellation
	time.Sleep(50 * time.Millisecond)

	// Session should be terminated
	if sessCtx.Err() == nil {
		t.Error("session context should be cancelled")
	}

	// Cleanup
	session.Terminate(nil)
}

func TestSessionWithRealMux(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	stream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	conn := createMockQUICConnection()
	mux := NewTrackMux()

	// Register a test handler
	ctx := context.Background()
	mux.Handle(ctx, BroadcastPath("/test/track"), TrackHandlerFunc(func(p *Publisher) {}))

	session := newSession(sessCtx, stream, conn, mux)

	if session.mux != mux {
		t.Error("mux not set correctly")
	}

	// Cleanup
	session.Terminate(nil)
}
