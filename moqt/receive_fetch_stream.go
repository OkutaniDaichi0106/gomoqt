package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type ReceiveFetchStream interface {
	// Get a SendDataStream
	SendGroupStream

	// Get a fetch request
	FetchRequest() FetchRequest

	// Close the stream
	Close() error

	// Close the stream with an error
	CloseWithError(err error) error
}

type receiveFetchStream struct {
	internalStream *internal.ReceiveFetchStream
}

func (rfs *receiveFetchStream) FetchRequest() FetchRequest {
	return FetchRequest{
		SubscribeID:   SubscribeID(rfs.internalStream.FetchMessage.SubscribeID),
		TrackPath:     rfs.internalStream.FetchMessage.TrackPath,
		TrackPriority: TrackPriority(rfs.internalStream.FetchMessage.TrackPriority),
		GroupSequence: GroupSequence(rfs.internalStream.FetchMessage.GroupSequence),
		FrameSequence: FrameSequence(rfs.internalStream.FetchMessage.FrameSequence),
	}
}

func (rfs *receiveFetchStream) SubscribeID() SubscribeID {
	return SubscribeID(rfs.internalStream.FetchMessage.SubscribeID)
}

func (rfs *receiveFetchStream) GroupSequence() GroupSequence {
	return GroupSequence(rfs.internalStream.FetchMessage.GroupSequence)
}

func (rfs *receiveFetchStream) WriteFrame(frame []byte) error {
	return rfs.internalStream.WriteFrame(frame)
}

func (rfs *receiveFetchStream) SetWriteDeadline(t time.Time) error {
	return rfs.internalStream.SetWriteDeadline(t)
}

func (rfs *receiveFetchStream) CancelWrite(code GroupErrorCode) {
	rfs.internalStream.CancelWrite(message.GroupErrorCode(code))
}

func (rfs *receiveFetchStream) Close() error {
	return rfs.internalStream.Close()
}

func (rfs *receiveFetchStream) CloseWithError(err error) error {
	return rfs.internalStream.CloseWithError(err)
}
