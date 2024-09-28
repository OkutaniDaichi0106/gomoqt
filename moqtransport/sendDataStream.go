package moqtransport

import (
	"errors"
	"go-moq/moqtransport/moqtmessage"
)

type SendDataStreamDatagram struct {
	closed   bool
	trSess   TransportSession
	header   moqtmessage.StreamHeaderDatagram
	groupID  moqtmessage.GroupID
	objectID moqtmessage.ObjectID
}

func (stream *SendDataStreamDatagram) Write(payload []byte) (int, error) {
	if stream.closed {
		return 0, ErrClosedStream
	}

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
	err := stream.trSess.SendDatagram(datagram.Serialize())
	if err != nil {
		return 0, err
	}

	// Increment the Object ID by 1
	stream.objectID++

	return len(datagram.Serialize()), err
}

func (stream *SendDataStreamDatagram) CloseWithStatus(code moqtmessage.ObjectStatusCode) error {
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

func (stream *SendDataStreamDatagram) NextGroup() (*SendDataStreamDatagram, error) {
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

type SendDataStreamTrack struct {
	closed   bool
	header   moqtmessage.StreamHeaderTrack
	writer   SendByteStream
	groupID  moqtmessage.GroupID
	objectID moqtmessage.ObjectID
}

func (stream *SendDataStreamTrack) Write(payload []byte) (int, error) {
	if stream.closed {
		return 0, ErrClosedStream
	}

	chunk := moqtmessage.GroupChunk{
		GroupID: stream.groupID,
		ObjectChunk: moqtmessage.ObjectChunk{
			ObjectID: stream.objectID,
			Payload:  payload,
		},
	}

	n, err := stream.writer.Write(chunk.Serialize())

	// Increment the Object ID by 1
	stream.objectID++

	return n, err
}

func (stream *SendDataStreamTrack) CloseWithStatus(code moqtmessage.ObjectStatusCode) error {
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

func (stream *SendDataStreamTrack) NextGroup() (*SendDataStreamTrack, error) {
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

	stream.objectID = 0

	return stream, nil
}

type SendDataStreamPeep struct {
	closed   bool
	writer   SendByteStream
	header   moqtmessage.StreamHeaderPeep
	objectID moqtmessage.ObjectID
}

func (stream *SendDataStreamPeep) Write(payload []byte) (int, error) {
	if stream.closed {
		return 0, ErrClosedStream
	}

	chunk := moqtmessage.ObjectChunk{
		ObjectID: stream.objectID,
		Payload:  payload,
	}
	n, err := stream.writer.Write(chunk.Serialize())

	// Increment the Object ID by 1
	stream.objectID++

	return n, err
}

func (stream *SendDataStreamPeep) CloseWithStatus(code moqtmessage.ObjectStatusCode) error {
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

var ErrClosedStream = errors.New("the stream was closed")
