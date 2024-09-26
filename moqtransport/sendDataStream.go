package moqtransport

import (
	"errors"
	"go-moq/moqtransport/moqtmessage"
	"io"
)

type SendDataStream interface {
	ForwardingPreference() moqtmessage.ObjectForwardingPreference
	io.Writer
	CloseWithStatus(moqtmessage.ObjectStatusCode) error
	NextGroup() (SendDataStream, error)
	NextStream(moqtmessage.PublisherPriority) (SendDataStream, error)
}

var _ SendDataStream = (*sendDataStreamDatagram)(nil)

type sendDataStreamDatagram struct {
	closed bool
	trSess TransportSession
	moqtmessage.SubscribeID
	moqtmessage.TrackAlias
	moqtmessage.PublisherPriority
	groupID  moqtmessage.GroupID
	objectID moqtmessage.ObjectID
}

func (stream *sendDataStreamDatagram) ForwardingPreference() moqtmessage.ObjectForwardingPreference {
	return moqtmessage.DATAGRAM
}

func (stream *sendDataStreamDatagram) Write(payload []byte) (int, error) {
	if stream.closed {
		return 0, ErrClosedStream
	}

	// Increment the Object ID by 1
	stream.objectID++

	datagram := moqtmessage.ObjectDatagram{
		SubscribeID:       stream.SubscribeID,
		TrackAlias:        stream.TrackAlias,
		PublisherPriority: stream.PublisherPriority,
		GroupChunk: moqtmessage.GroupChunk{
			GroupID: stream.groupID,
			ObjectChunk: moqtmessage.ObjectChunk{
				ObjectID: stream.objectID,
				Payload:  payload,
			},
		},
	}

	return len(datagram.Serialize()), stream.trSess.SendDatagram(datagram.Serialize())
}

func (stream *sendDataStreamDatagram) CloseWithStatus(code moqtmessage.ObjectStatusCode) error {
	if stream.closed {
		return ErrClosedStream
	}

	finalDatagram := moqtmessage.ObjectDatagram{
		SubscribeID:       stream.SubscribeID,
		TrackAlias:        stream.TrackAlias,
		PublisherPriority: stream.PublisherPriority,
		GroupChunk: moqtmessage.GroupChunk{
			GroupID: stream.groupID,
			ObjectChunk: moqtmessage.ObjectChunk{
				ObjectID:   stream.objectID,
				Payload:    []byte{},
				StatusCode: code,
			},
		},
	}

	err := stream.trSess.SendDatagram(finalDatagram.Serialize())
	if err != nil {
		return err
	}

	stream.closed = true

	return nil
}

func (stream *sendDataStreamDatagram) NextGroup() (SendDataStream, error) {
	if stream.closed {
		return nil, ErrClosedStream
	}

	finalDatagram := moqtmessage.ObjectDatagram{
		SubscribeID:       stream.SubscribeID,
		TrackAlias:        stream.TrackAlias,
		PublisherPriority: stream.PublisherPriority,
		GroupChunk: moqtmessage.GroupChunk{
			GroupID: stream.groupID,
			ObjectChunk: moqtmessage.ObjectChunk{
				ObjectID:   stream.objectID,
				Payload:    []byte{},
				StatusCode: moqtmessage.END_OF_GROUP,
			},
		},
	}

	err := stream.trSess.SendDatagram(finalDatagram.Serialize())
	if err != nil {
		return nil, err
	}

	// Increment the Group ID by 1
	stream.groupID++

	return stream, nil
}

func (stream *sendDataStreamDatagram) NextStream(priority moqtmessage.PublisherPriority) (SendDataStream, error) {
	if stream.closed {
		return nil, ErrClosedStream
	}

	// Send a message to indicate the end of the track
	finalDatagram := moqtmessage.ObjectDatagram{
		SubscribeID:       stream.SubscribeID,
		TrackAlias:        stream.TrackAlias,
		PublisherPriority: stream.PublisherPriority,
		GroupChunk: moqtmessage.GroupChunk{
			GroupID: stream.groupID,
			ObjectChunk: moqtmessage.ObjectChunk{
				ObjectID:   stream.objectID,
				Payload:    []byte{},
				StatusCode: moqtmessage.END_OF_TRACK,
			},
		},
	}

	err := stream.trSess.SendDatagram(finalDatagram.Serialize())
	if err != nil {
		return nil, err
	}

	newStream := sendDataStreamDatagram{
		closed:            false,
		trSess:            stream.trSess,
		SubscribeID:       stream.SubscribeID,
		TrackAlias:        stream.TrackAlias,
		PublisherPriority: priority,
		groupID:           0,
		objectID:          0,
	}

	return &newStream, nil
}

var _ SendDataStream = (*sendDataStreamTrack)(nil)

type sendDataStreamTrack struct {
	closed       bool
	writerClosed bool
	trSess       TransportSession
	writer       SendByteStream
	header       moqtmessage.StreamHeaderTrack
	groupID      moqtmessage.GroupID
	objectID     moqtmessage.ObjectID
}

func (stream *sendDataStreamTrack) ForwardingPreference() moqtmessage.ObjectForwardingPreference {
	return moqtmessage.TRACK
}

func (stream *sendDataStreamTrack) Write(payload []byte) (int, error) {
	if stream.closed {
		return 0, ErrClosedStream
	}

	if stream.writerClosed {
		return 0, ErrClosedStream
	}

	// Increment the Object ID by 1
	stream.objectID++

	chunk := moqtmessage.GroupChunk{
		GroupID: stream.groupID,
		ObjectChunk: moqtmessage.ObjectChunk{
			ObjectID: stream.objectID,
			Payload:  payload,
		},
	}

	return stream.writer.Write(chunk.Serialize())
}

func (stream *sendDataStreamTrack) CloseWithStatus(code moqtmessage.ObjectStatusCode) error {
	if stream.closed {
		return ErrClosedStream
	}

	finalChunk := moqtmessage.GroupChunk{
		GroupID: stream.groupID,
		ObjectChunk: moqtmessage.ObjectChunk{
			ObjectID:   stream.objectID,
			Payload:    []byte{},
			StatusCode: code,
		},
	}

	_, err := stream.writer.Write(finalChunk.Serialize())
	if err != nil {
		return err
	}

	stream.closed = true

	return stream.writer.Close()
}

func (stream *sendDataStreamTrack) NextGroup() (SendDataStream, error) {
	if stream.closed {
		return nil, ErrClosedStream
	}

	// Send a message to indicate the end of the group
	finalChunk := moqtmessage.GroupChunk{
		GroupID: stream.groupID,
		ObjectChunk: moqtmessage.ObjectChunk{
			ObjectID:   stream.objectID,
			Payload:    []byte{},
			StatusCode: moqtmessage.END_OF_GROUP,
		},
	}

	_, err := stream.writer.Write(finalChunk.Serialize())
	if err != nil {
		return nil, err
	}

	// Increment the Group ID by 1
	stream.groupID++

	stream.objectID = 1<<64 - 1

	return stream, nil
}

func (stream *sendDataStreamTrack) NextStream(priority moqtmessage.PublisherPriority) (SendDataStream, error) {
	if stream.closed {
		return nil, ErrClosedStream
	}

	if !stream.writerClosed {
		// Close the writer
		finalChunk := moqtmessage.GroupChunk{
			GroupID: stream.groupID,
			ObjectChunk: moqtmessage.ObjectChunk{
				ObjectID:   stream.objectID,
				Payload:    []byte{},
				StatusCode: moqtmessage.END_OF_TRACK,
			},
		}

		_, err := stream.writer.Write(finalChunk.Serialize())
		if err != nil {
			return nil, err
		}

		err = stream.writer.Close()
		if err != nil {
			return nil, err
		}

		stream.writerClosed = true
	}

	// Open a new writer
	writer, err := stream.trSess.OpenUniStream()
	if err != nil {
		return nil, err
	}

	// Send a stream header
	newHeader := moqtmessage.StreamHeaderTrack{
		SubscribeID:       stream.header.GetSubscribeID(),
		TrackAlias:        stream.header.TrackAlias,
		PublisherPriority: priority,
	}

	_, err = stream.writer.Write(newHeader.Serialize())
	if err != nil {
		return nil, errors.New(err.Error() + ", and failed to send header")
	}

	return &sendDataStreamTrack{
		closed:   false,
		trSess:   stream.trSess,
		writer:   writer,
		header:   newHeader,
		groupID:  0,
		objectID: 1<<64 - 1,
	}, nil
}

var _ SendDataStream = (*sendDataStreamPeep)(nil)

type sendDataStreamPeep struct {
	closed   bool
	trSess   TransportSession
	writer   SendByteStream
	header   moqtmessage.StreamHeaderPeep
	objectID moqtmessage.ObjectID

	writerClosed bool
}

func (stream *sendDataStreamPeep) ForwardingPreference() moqtmessage.ObjectForwardingPreference {
	return moqtmessage.PEEP
}

func (stream *sendDataStreamPeep) Write(payload []byte) (int, error) {
	if stream.closed || stream.writerClosed {
		return 0, ErrClosedStream
	}

	// Increment the Object ID by 1
	stream.objectID++

	chunk := moqtmessage.ObjectChunk{
		ObjectID: stream.objectID,
		Payload:  payload,
	}

	return stream.writer.Write(chunk.Serialize())
}

func (stream *sendDataStreamPeep) CloseWithStatus(code moqtmessage.ObjectStatusCode) error {
	if stream.closed {
		return ErrClosedStream
	}

	finalChunk := moqtmessage.ObjectChunk{
		ObjectID:   stream.objectID,
		Payload:    []byte{},
		StatusCode: code,
	}

	_, err := stream.writer.Write(finalChunk.Serialize())
	if err != nil {
		return err
	}

	stream.closed = true

	return stream.writer.Close()
}

func (stream *sendDataStreamPeep) NextGroup() (SendDataStream, error) {
	if stream.closed || stream.writerClosed {
		return nil, ErrClosedStream
	}

	if !stream.writerClosed {
		// Send final chunk
		finalChunk := moqtmessage.ObjectChunk{
			ObjectID:   stream.objectID,
			Payload:    []byte{},
			StatusCode: moqtmessage.END_OF_GROUP,
		}

		_, err := stream.writer.Write(finalChunk.Serialize())
		if err != nil {
			return nil, err
		}

		err = stream.writer.Close()
		if err != nil {
			return nil, err
		}

		stream.writerClosed = true
	}

	// Increment the Group ID by 1
	stream.header.GroupID++

	// Set the Peep ID to -1
	stream.header.PeepID = 1<<64 - 1

	return stream, nil
}
func (stream *sendDataStreamPeep) NextStream(priority moqtmessage.PublisherPriority) (SendDataStream, error) {
	if stream.closed {
		return nil, ErrClosedStream
	}

	if !stream.writerClosed {
		// Send final chunk
		finalChunk := moqtmessage.ObjectChunk{
			ObjectID:   stream.objectID,
			Payload:    []byte{},
			StatusCode: moqtmessage.END_OF_PEEP,
		}

		_, err := stream.writer.Write(finalChunk.Serialize())
		if err != nil {
			return nil, err
		}

		err = stream.writer.Close()
		if err != nil {
			return nil, err
		}

		stream.writerClosed = true
	}

	// Open a new writer
	writer, err := stream.trSess.OpenUniStream()
	if err != nil {
		return nil, err
	}

	// Increment the Peep ID by 1
	newPeepID := stream.header.PeepID + 1

	// Send a stream header
	newHeader := moqtmessage.StreamHeaderPeep{
		SubscribeID:       stream.header.SubscribeID,
		TrackAlias:        stream.header.TrackAlias,
		GroupID:           stream.header.GroupID,
		PeepID:            newPeepID,
		PublisherPriority: stream.header.PublisherPriority,
	}

	_, err = writer.Write(newHeader.Serialize())
	if err != nil {
		return nil, err
	}

	return &sendDataStreamPeep{
		closed:       false,
		writerClosed: false,
		trSess:       stream.trSess,
		writer:       writer,
		header:       newHeader,
		objectID:     1<<64 - 1,
	}, nil
}

var ErrClosedStream = errors.New("the stream was closed")
