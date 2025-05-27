package moqt

import (
	"context"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

// Mock ReceiveStream for testing
type mockReceiveStream struct{}

func (m *mockReceiveStream) StreamID() quic.StreamID {
	return quic.StreamID(1)
}

func (m *mockReceiveStream) Read(b []byte) (int, error) {
	return 0, nil
}

func (m *mockReceiveStream) CancelRead(code quic.StreamErrorCode) {
	// Mock implementation - does nothing
}

func (m *mockReceiveStream) SetReadDeadline(t time.Time) error {
	return nil
}

func TestIncomingGroupStreamQueue_EnqueueAndAccept(t *testing.T) {
	config := func() *SubscribeConfig {
		return &SubscribeConfig{MinGroupSequence: 0, MaxGroupSequence: 100}
	}
	queue := newIncomingGroupStreamQueue(config)

	// Create a properly initialized receiveGroupStream with a mock stream
	mockStream := &mockReceiveStream{}
	stream := newReceiveGroupStream(SubscribeID(1), GroupSequence(50), mockStream)

	// Enqueue a stream
	queue.enqueue(stream)

	// Accept the stream
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	acceptedStream, err := queue.dequeue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if acceptedStream != stream {
		t.Fatalf("expected %v, got %v", stream, acceptedStream)
	}
}

func TestIncomingGroupStreamQueue_AcceptTimeout(t *testing.T) {
	config := func() *SubscribeConfig {
		return &SubscribeConfig{MinGroupSequence: 0, MaxGroupSequence: 100}
	}
	queue := newIncomingGroupStreamQueue(config)

	// Accept with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := queue.dequeue(ctx)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
