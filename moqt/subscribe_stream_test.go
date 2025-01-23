package moqt

import (
	"testing"

	mock_transport "github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestReceiveSubscribeStreamQueue_Enqueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	queue := newReceiveSubscribeStreamQueue()
	stream := &receiveSubscribeStream{
		subscribeID: SubscribeID(0),
		config:      SubscribeConfig{},
		stream:      mock_transport.NewMockStream(ctrl),
	}

	queue.Enqueue(stream)

	assert.Equal(t, 1, queue.Len(), "Queue length should be 1 after enqueue")
}

func TestReceiveSubscribeStreamQueue_Dequeue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	queue := newReceiveSubscribeStreamQueue()
	stream := &receiveSubscribeStream{
		subscribeID: SubscribeID(0),
		config:      SubscribeConfig{},
		stream:      mock_transport.NewMockStream(ctrl),
	}

	queue.Enqueue(stream)
	<-queue.Chan()
	dequeuedStream := queue.Dequeue()

	assert.Equal(t, stream, dequeuedStream, "Dequeued stream should be the same as the enqueued stream")
	assert.Equal(t, 0, queue.Len(), "Queue length should be 0 after dequeue")
}
