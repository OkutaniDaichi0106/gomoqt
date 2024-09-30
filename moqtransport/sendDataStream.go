package moqtransport

import (
	"errors"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
)

type SendDataStreamDatagram struct {
	closed bool
	header moqtmessage.StreamHeaderDatagram
}

// func (stream *SendDataStreamDatagram) Write(payload []byte) (int, error) {
// 	if stream.closed {
// 		return 0, ErrClosedStream
// 	}

// 	datagram := moqtmessage.ObjectDatagram{
// 		SubscribeID:       stream.header.SubscribeID,
// 		TrackAlias:        stream.header.TrackAlias,
// 		PublisherPriority: stream.header.PublisherPriority,
// 		GroupChunk: moqtmessage.GroupChunk{
// 			GroupID: stream.groupID,
// 			ObjectChunk: moqtmessage.ObjectChunk{
// 				ObjectID: stream.objectID,
// 				Payload:  payload,
// 			},
// 		},
// 	}
// 	err := stream.trSess.SendDatagram(datagram.Serialize())
// 	if err != nil {
// 		return 0, err
// 	}

// 	// Increment the Object ID by 1
// 	stream.objectID++

// 	return len(datagram.Serialize()), err
// }

// func (stream *SendDataStreamDatagram) CloseWithStatus(code moqtmessage.ObjectStatusCode) error {
// 	if stream.closed {
// 		return ErrClosedStream
// 	}

// 	finalDatagram := moqtmessage.ObjectDatagram{
// 		SubscribeID:       stream.header.SubscribeID,
// 		TrackAlias:        stream.header.TrackAlias,
// 		PublisherPriority: stream.header.PublisherPriority,
// 		GroupChunk: moqtmessage.GroupChunk{
// 			GroupID: stream.groupID,
// 			ObjectChunk: moqtmessage.ObjectChunk{
// 				ObjectID:   stream.objectID,
// 				Payload:    []byte{},
// 				StatusCode: code,
// 			},
// 		},
// 	}

// 	err := stream.trSess.SendDatagram(finalDatagram.Serialize())
// 	if err != nil {
// 		return err
// 	}

// 	stream.closed = true

// 	return nil
// }

// func (stream *SendDataStreamDatagram) NextGroup() (*SendDataStreamDatagram, error) {
// 	if stream.closed {
// 		return nil, ErrClosedStream
// 	}

// 	finalDatagram := moqtmessage.ObjectDatagram{
// 		SubscribeID:       stream.header.SubscribeID,
// 		TrackAlias:        stream.header.TrackAlias,
// 		PublisherPriority: stream.header.PublisherPriority,
// 		GroupChunk: moqtmessage.GroupChunk{
// 			GroupID: stream.groupID,
// 			ObjectChunk: moqtmessage.ObjectChunk{
// 				ObjectID:   stream.objectID,
// 				Payload:    []byte{},
// 				StatusCode: moqtmessage.END_OF_GROUP,
// 			},
// 		},
// 	}

// 	err := stream.trSess.SendDatagram(finalDatagram.Serialize())
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Increment the Group ID by 1
// 	stream.groupID++

// 	return stream, nil
// }

type SendDataStreamTrack struct {
	closed   bool
	header   moqtmessage.StreamHeaderTrack
	writer   SendByteStream
	groupID  moqtmessage.GroupID
	objectID moqtmessage.ObjectID
}

func (stream *SendDataStreamTrack) Write(payload []byte) (int, *ChunkData, error) {
	if stream.closed {
		return 0, nil, ErrClosedStream
	}

	chunk := moqtmessage.GroupChunk{
		GroupID: stream.groupID,
		ObjectChunk: moqtmessage.ObjectChunk{
			ObjectID: stream.objectID,
			Payload:  payload,
		},
	}

	if len(payload) == 0 {
		chunk.StatusCode = moqtmessage.NOMAL_OBJECT
	}

	n, err := stream.writer.Write(chunk.Serialize())
	if err != nil {
		return 0, nil, err
	}

	newChunkData := ChunkData{
		groupID:  stream.groupID,
		objectID: stream.objectID,
	}

	stream.objectID++

	return n, &newChunkData, nil
}

func (stream *SendDataStreamTrack) CloseWithStatus(code moqtmessage.ObjectStatusCode) (*ChunkData, error) {
	if stream.closed {
		return nil, ErrClosedStream
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
		return nil, err
	}

	stream.closed = true

	err = stream.writer.Close()
	if err != nil {
		return nil, err
	}

	return &ChunkData{
		groupID:  stream.groupID,
		peepID:   nil,
		objectID: stream.objectID,
	}, nil
}

type SendDataStreamPeep struct {
	closed   bool
	header   moqtmessage.StreamHeaderPeep
	writer   SendByteStream
	objectID moqtmessage.ObjectID
}

func (stream *SendDataStreamPeep) Write(payload []byte) (int, *ChunkData, error) {
	if stream.closed {
		return 0, nil, ErrClosedStream
	}

	chunk := moqtmessage.ObjectChunk{
		ObjectID: stream.objectID,
		Payload:  payload,
	}
	n, err := stream.writer.Write(chunk.Serialize())
	if err != nil {
		return 0, nil, err
	}

	newChunkData := ChunkData{
		groupID:  stream.header.GroupID,
		peepID:   &stream.header.PeepID,
		objectID: stream.objectID,
	}

	stream.objectID++

	return n, &newChunkData, nil
}

func (stream *SendDataStreamPeep) CloseWithStatus(code moqtmessage.ObjectStatusCode) (*ChunkData, error) {
	if stream.closed {
		return nil, ErrClosedStream
	}

	finalChunk := moqtmessage.ObjectChunk{
		ObjectID:   stream.objectID + 1,
		Payload:    []byte{},
		StatusCode: code,
	}

	_, err := stream.writer.Write(finalChunk.Serialize())
	if err != nil {
		return nil, err
	}

	stream.closed = true
	err = stream.writer.Close()
	if err != nil {
		return nil, err
	}

	return &ChunkData{
		groupID:  stream.header.GroupID,
		peepID:   &stream.header.PeepID,
		objectID: stream.objectID + 1,
	}, nil
}

// func (stream *SendDataStreamPeep) NextGroup() (*SendDataStreamPeep, error) {
// 	writer, err := stream.trSess.OpenUniStream()
// 	if err != nil {
// 		return nil, err
// 	}

// 	stream.header.GroupID++
// 	stream.header.PeepID = 0

// 	return &SendDataStreamPeep{
// 		closed:   false,
// 		writer:   writer,
// 		header:   stream.header,
// 		objectID: 0,
// 	}, nil
// }

// func (stream *SendDataStreamPeep) NextPeep() (*SendDataStreamPeep, error) {
// 	writer, err := stream.trSess.OpenUniStream()
// 	if err != nil {
// 		return nil, err
// 	}

// 	stream.header.PeepID++

// 	return &SendDataStreamPeep{
// 		closed:   false,
// 		writer:   writer,
// 		header:   stream.header,
// 		objectID: 0,
// 	}, nil
// }

var ErrClosedStream = errors.New("the stream was closed")
