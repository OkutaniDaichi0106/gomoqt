package moqtransport

import (
	"errors"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
)

type SendDataStreamTrack struct {
	closed   bool
	header   moqtmessage.StreamHeaderTrack
	writer   SendStream
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
	writer   SendStream
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

var ErrClosedStream = errors.New("the stream was closed")
