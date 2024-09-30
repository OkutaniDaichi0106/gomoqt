package moqtransport

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go/quicvarint"
)

type ReceiveDataStream interface {
	GetSubscribeID() moqtmessage.SubscribeID
	GetTrackAlias() moqtmessage.TrackAlias
	GetPublisherPriority() moqtmessage.PublisherPriority
	ReadChunk([]byte) (int, ChunkData, error)
}

type ChunkData struct {
	groupID  moqtmessage.GroupID
	peepID   *moqtmessage.PeepID
	objectID moqtmessage.ObjectID
}

func (cd ChunkData) NextGroup() ChunkData {
	newChunkData := ChunkData{
		groupID:  cd.groupID + 1,
		objectID: 0,
	}

	if cd.peepID != nil {
		newPeepID := moqtmessage.PeepID(0)
		newChunkData.peepID = &newPeepID
	}

	return newChunkData
}

func (cd ChunkData) NextPeep() ChunkData {
	if cd.peepID == nil {
		panic("prior peep id not found")
	}

	newPeepID := moqtmessage.PeepID(*cd.peepID + 1)

	newChunkData := ChunkData{
		groupID:  cd.groupID,
		peepID:   &newPeepID,
		objectID: 0,
	}

	return newChunkData
}

func (cd ChunkData) NextObject() ChunkData {
	newChunkData := ChunkData{
		groupID:  cd.groupID,
		peepID:   cd.peepID,
		objectID: cd.objectID,
	}

	return newChunkData
}

func (cd ChunkData) GetGroupID() moqtmessage.GroupID {
	return cd.groupID
}

func (cd ChunkData) GetPeepID() *moqtmessage.PeepID {
	return cd.peepID
}

func (cd ChunkData) GetObjectID() moqtmessage.ObjectID {
	return cd.objectID
}

var _ ReceiveDataStream = (*receiveDataStreamDatagram)(nil)

type receiveDataStreamDatagram struct {
	header moqtmessage.StreamHeaderDatagram
	reader quicvarint.Reader
}

func (stream receiveDataStreamDatagram) GetSubscribeID() moqtmessage.SubscribeID {
	return stream.header.GetSubscribeID()
}

func (stream receiveDataStreamDatagram) GetTrackAlias() moqtmessage.TrackAlias {
	return stream.header.GetTrackAlias()
}

func (stream receiveDataStreamDatagram) GetPublisherPriority() moqtmessage.PublisherPriority {
	return stream.header.GetPublisherPriority()
}

func (stream receiveDataStreamDatagram) ReadChunk(buf []byte) (int, ChunkData, error) {
	var chunk moqtmessage.GroupChunk

	err := chunk.DeserializeBody(stream.reader)
	if err != nil {
		return 0, ChunkData{}, err
	}

	n := copy(buf, chunk.Payload)

	return n, ChunkData{
		groupID:  chunk.GroupID,
		peepID:   nil,
		objectID: chunk.ObjectID,
	}, nil
}

var _ ReceiveDataStream = (*receiveDataStreamTrack)(nil)

type receiveDataStreamTrack struct {
	header moqtmessage.StreamHeaderTrack
	reader quicvarint.Reader
}

func (stream receiveDataStreamTrack) GetSubscribeID() moqtmessage.SubscribeID {
	return stream.header.SubscribeID
}

func (stream receiveDataStreamTrack) GetTrackAlias() moqtmessage.TrackAlias {
	return stream.header.TrackAlias
}

func (stream receiveDataStreamTrack) GetPublisherPriority() moqtmessage.PublisherPriority {
	return stream.header.PublisherPriority
}

func (stream receiveDataStreamTrack) ReadChunk(buf []byte) (int, ChunkData, error) {
	var chunk moqtmessage.GroupChunk

	err := chunk.DeserializeBody(stream.reader)
	if err != nil {
		return 0, ChunkData{}, err
	}

	n := copy(buf, chunk.Payload)

	return n, ChunkData{
		groupID:  chunk.GroupID,
		peepID:   nil,
		objectID: chunk.ObjectID,
	}, nil
}

var _ ReceiveDataStream = (*receiveDataStreamPeep)(nil)

type receiveDataStreamPeep struct {
	header moqtmessage.StreamHeaderPeep
	reader quicvarint.Reader
}

func (stream receiveDataStreamPeep) GetSubscribeID() moqtmessage.SubscribeID {
	return stream.header.SubscribeID
}

func (stream receiveDataStreamPeep) GetTrackAlias() moqtmessage.TrackAlias {
	return stream.header.TrackAlias
}

func (stream receiveDataStreamPeep) GetPublisherPriority() moqtmessage.PublisherPriority {
	return stream.header.PublisherPriority
}

func (stream receiveDataStreamPeep) ReadChunk(buf []byte) (int, ChunkData, error) {
	var chunk moqtmessage.ObjectChunk

	err := chunk.DeserializeBody(stream.reader)
	if err != nil {
		return 0, ChunkData{}, err
	}

	n := copy(buf, chunk.Payload)

	return n, ChunkData{
		groupID:  stream.header.GroupID,
		peepID:   &stream.header.PeepID,
		objectID: chunk.ObjectID,
	}, nil
}
