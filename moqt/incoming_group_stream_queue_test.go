package moqt

import (
	"context"
	"testing"
	"time"
)

func TestIncomingGroupStreamQueue_EnqueueAndAccept(t *testing.T) {
	config := &SubscribeConfig{MinGroupSequence: 0, MaxGroupSequence: 100, GroupOrder: GroupOrderDefault}
	queue := newIncomingGroupStreamQueue(SubscribeID(1), BroadcastPath("test"), config)
	stream := &receiveGroupStream{}

	// Enqueue a stream
	err := queue.Enqueue(stream)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

func TestIncomingGroupStreamQueue_AcceptTimeout(t *testing.T) {
	config := &SubscribeConfig{MinGroupSequence: 0, MaxGroupSequence: 100, GroupOrder: GroupOrderDefault}
	queue := newIncomingGroupStreamQueue(SubscribeID(1), BroadcastPath("test"), config)

	// Accept with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := queue.Accept(ctx)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
