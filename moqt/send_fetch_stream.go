package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

var _ ReceiveGroupStream = (SendFetchStream)(nil)

type SendFetchStream interface {
	// Get a fetch request
	FetchRequest() FetchRequest

	// Update the fetch
	UpdateFetch(FetchUpdate) error

	SubscribeID() SubscribeID

	GroupSequence() GroupSequence

	ReadFrame() ([]byte, error)

	SetReadDeadline(time.Time) error
	// Cancel the fetch
	CancelRead(StreamErrorCode)

	// Close the stream
	Close() error // TODO: delete if not used

	// Close the stream with an error
	CloseWithError(err error) error // TODO: delete if not used or rename to CancelFetch
}

var _ SendFetchStream = (*sendFetchStream)(nil)

type sendFetchStream struct {
	internalStream *internal.SendFetchStream
}

func (sfs *sendFetchStream) FetchRequest() FetchRequest {
	return FetchRequest{
		SubscribeID:   SubscribeID(sfs.internalStream.FetchMessage.SubscribeID),
		TrackPath:     sfs.internalStream.FetchMessage.TrackPath,
		TrackPriority: TrackPriority(sfs.internalStream.FetchMessage.TrackPriority),
		GroupSequence: GroupSequence(sfs.internalStream.FetchMessage.GroupSequence),
		FrameSequence: FrameSequence(sfs.internalStream.FetchMessage.FrameSequence),
	}
}

func (sfs *sendFetchStream) UpdateFetch(update FetchUpdate) error {
	return sfs.internalStream.UpdateFetch(&message.FetchUpdateMessage{
		TrackPriority: message.TrackPriority(update.TrackPriority),
	})
}

func (sfs *sendFetchStream) SubscribeID() SubscribeID {
	return SubscribeID(sfs.internalStream.FetchMessage.SubscribeID)
}

func (sfs *sendFetchStream) GroupSequence() GroupSequence {
	return GroupSequence(sfs.internalStream.FetchMessage.GroupSequence)
}

func (sfs *sendFetchStream) ReadFrame() ([]byte, error) {
	return sfs.internalStream.ReadFrame()
}

func (sfs *sendFetchStream) SetReadDeadline(t time.Time) error {
	return sfs.internalStream.SetReadDeadline(t)
}

func (sfs *sendFetchStream) CancelRead(code StreamErrorCode) {
	sfs.internalStream.CancelRead(internal.StreamErrorCode(code))
}

func (sfs *sendFetchStream) Close() error {
	return sfs.internalStream.Close()
}

func (sfs *sendFetchStream) CloseWithError(err error) error {
	return sfs.internalStream.CloseWithError(err)
}
