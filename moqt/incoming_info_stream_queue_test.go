package moqt

import (
	"context"
	"testing"
	"time"
)

func TestIncomingInfoStreamQueue_EnqueueAndAccept(t *testing.T) {
	queue := newIncomingInfoStreamQueue()
	stream := &sendInfoStream{}

	// Enqueue a stream
	queue.Enqueue(stream)

	// Accept the stream
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	acceptedStream, err := queue.Accept(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if acceptedStream != stream {
		t.Fatalf("expected %v, got %v", stream, acceptedStream)
	}
}

func TestIncomingInfoStreamQueue_AcceptTimeout(t *testing.T) {
	queue := newIncomingInfoStreamQueue()

	// Accept with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := queue.Accept(ctx)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
