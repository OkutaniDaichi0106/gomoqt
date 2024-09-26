package moqtransport

import (
	"errors"
	"go-moq/moqtransport/moqtmessage"
	"io"
	"log"
	"sync"

	"github.com/quic-go/quic-go/quicvarint"
)

type ReceiveDataStream interface {
	ForwardingPreference() moqtmessage.ObjectForwardingPreference
	io.Reader
	NextGroup() (ReceiveDataStream, error)
	NextStream() (ReceiveDataStream, error)
}

var _ ReceiveDataStream = (*receiveDataStreamDatagram)(nil)

type receiveDataStreamDatagram struct {
	closed   bool
	mu       sync.Mutex
	header   moqtmessage.StreamHeaderDatagram
	groupID  moqtmessage.GroupID
	objectID moqtmessage.ObjectID

	readerCh <-chan struct {
		moqtmessage.StreamHeaderDatagram
		quicvarint.Reader
	}

	dataMap map[moqtmessage.GroupID]map[moqtmessage.ObjectID]chan []byte
}

func (stream *receiveDataStreamDatagram) init() error {
	if stream.closed {
		return io.EOF
	}

	go func() {
		var chunk moqtmessage.GroupChunk
		for {
			// Receive a group chunk
			reader := <-stream.readerCh

			err := chunk.DeserializeBody(reader.Reader)
			if err != nil {
				log.Println(err)
				return //err
			}

			_, ok := stream.dataMap[chunk.GroupID]
			if !ok {
				stream.dataMap[chunk.GroupID] = make(map[moqtmessage.ObjectID]chan []byte)
			}

			_, ok = stream.dataMap[chunk.GroupID][chunk.ObjectID]
			if !ok {
				stream.dataMap[chunk.GroupID][stream.objectID] = make(chan []byte, 1)
			}

			stream.dataMap[chunk.GroupID][chunk.ObjectID] <- chunk.Payload
		}
	}()

	return nil
}

func (stream *receiveDataStreamDatagram) ForwardingPreference() moqtmessage.ObjectForwardingPreference {
	return moqtmessage.DATAGRAM
}

func (stream *receiveDataStreamDatagram) Read(buf []byte) (int, error) {
	if stream.closed {
		return 0, io.EOF
	}

	stream.mu.Lock()
	defer stream.mu.Unlock()

	// Increment the Object ID by 1
	stream.objectID++

	_, ok := stream.dataMap[stream.groupID]

	if !ok {
		stream.dataMap[stream.groupID] = make(map[moqtmessage.ObjectID]chan []byte)
	}

	_, ok = stream.dataMap[stream.groupID][stream.objectID]

	if !ok {
		stream.dataMap[stream.groupID][stream.objectID] = make(chan []byte)
	}

	payloadCh := stream.dataMap[stream.groupID][stream.objectID]

	payload, ok := <-payloadCh
	if !ok {
		return 0, errors.New("no data")
	}

	n := copy(buf, payload)

	return n, nil
}

func (stream *receiveDataStreamDatagram) NextGroup() (ReceiveDataStream, error) {
	if stream.closed {
		return nil, ErrClosedStream
	}

	// Delete
	delete(stream.dataMap, stream.groupID)

	// Increment the Group ID by 1
	stream.groupID++
	return stream, nil
}

func (stream *receiveDataStreamDatagram) NextStream() (ReceiveDataStream, error) {
	if stream.closed {
		return nil, ErrClosedStream
	}

	newStream := receiveDataStreamDatagram{
		closed:   false,
		header:   stream.header,
		readerCh: stream.readerCh,
		groupID:  0,
		objectID: 0,
	}

	newStream.init()

	return &newStream, nil
}

var _ ReceiveDataStream = (*receiveDataStreamTrack)(nil)

type receiveDataStreamTrack struct {
	mu       sync.Mutex
	closed   bool
	header   moqtmessage.StreamHeaderTrack
	groupID  moqtmessage.GroupID
	objectID moqtmessage.ObjectID

	readerCh chan struct {
		moqtmessage.StreamHeaderTrack
		quicvarint.Reader
	}

	dataMap map[moqtmessage.GroupID]map[moqtmessage.ObjectID]chan []byte
}

func (stream *receiveDataStreamTrack) init() error {
	if stream.closed {
		return io.EOF
	}

	// Receive a receive reader
	reader := <-stream.readerCh

	stream.header.PublisherPriority = reader.PublisherPriority

	go func(qvReader quicvarint.Reader) {

		var chunk moqtmessage.GroupChunk

		for {
			// Get a Group Chunk
			err := chunk.DeserializeBody(qvReader)
			if err != nil {
				log.Println(err)
				return
			}

			stream.mu.Lock()

			_, ok := stream.dataMap[chunk.GroupID]
			if !ok {
				stream.dataMap[chunk.GroupID] = make(map[moqtmessage.ObjectID]chan []byte)
			}

			_, ok = stream.dataMap[chunk.GroupID][chunk.ObjectID]
			if !ok {
				stream.dataMap[chunk.GroupID][stream.objectID] = make(chan []byte, 1)
			}

			stream.dataMap[chunk.GroupID][chunk.ObjectID] <- chunk.Payload

			stream.mu.Unlock()
		}

	}(reader.Reader)

	return nil
}

func (stream *receiveDataStreamTrack) ForwardingPreference() moqtmessage.ObjectForwardingPreference {
	return moqtmessage.TRACK
}

func (stream *receiveDataStreamTrack) Read(buf []byte) (int, error) {
	if stream.closed {
		return 0, io.EOF
	}

	stream.mu.Lock()
	defer stream.mu.Unlock()

	// Increment the Object ID by 1
	stream.objectID++

	_, ok := stream.dataMap[stream.groupID]

	if !ok {
		stream.dataMap[stream.groupID] = make(map[moqtmessage.ObjectID]chan []byte)
	}

	_, ok = stream.dataMap[stream.groupID][stream.objectID]

	if !ok {
		stream.dataMap[stream.groupID][stream.objectID] = make(chan []byte)
	}

	payloadCh := stream.dataMap[stream.groupID][stream.objectID]

	payload, ok := <-payloadCh
	if !ok {
		return 0, errors.New("no data")
	}

	n := copy(buf, payload)

	return n, nil
}

func (stream *receiveDataStreamTrack) NextGroup() (ReceiveDataStream, error) {
	if stream.closed {
		return nil, ErrClosedStream
	}

	// Delete
	delete(stream.dataMap, stream.groupID)

	// Increment the Group ID by 1
	stream.groupID++
	return stream, nil
}

func (stream *receiveDataStreamTrack) NextStream() (ReceiveDataStream, error) {
	if stream.closed {
		return nil, ErrClosedStream
	}

	stream.init()

	return stream, nil
}

var _ ReceiveDataStream = (*receiveDataStreamPeep)(nil)

type receiveDataStreamPeep struct {
	mu       sync.Mutex
	closed   bool
	header   moqtmessage.StreamHeaderPeep
	objectID moqtmessage.ObjectID

	readerCh chan struct {
		moqtmessage.StreamHeaderPeep
		quicvarint.Reader
	}

	readerMap map[moqtmessage.GroupID]map[moqtmessage.PeepID]chan struct {
		moqtmessage.PublisherPriority
		quicvarint.Reader
	}

	dataMap map[moqtmessage.GroupID]map[moqtmessage.PeepID]map[moqtmessage.ObjectID]chan []byte
}

func (stream *receiveDataStreamPeep) init() error {
	if stream.closed {
		return io.EOF
	}

	_, ok := stream.readerMap[stream.header.GroupID]
	if !ok {
		stream.readerMap[stream.header.GroupID] = make(map[moqtmessage.PeepID]chan struct {
			moqtmessage.PublisherPriority
			quicvarint.Reader
		})
	}

	_, ok = stream.readerMap[stream.header.GroupID][stream.header.PeepID]
	if !ok {
		stream.readerMap[stream.header.GroupID][stream.header.PeepID] = make(chan struct {
			moqtmessage.PublisherPriority
			quicvarint.Reader
		}, 1)
	}

	reader := <-stream.readerMap[stream.header.GroupID][stream.header.PeepID]

	stream.header.PublisherPriority = reader.PublisherPriority

	go func(reader quicvarint.Reader) {
		qvReader := quicvarint.NewReader(reader)

		var chunk moqtmessage.ObjectChunk

		for {
			// Get a Group Chunk
			err := chunk.DeserializeBody(qvReader)
			if err != nil {
				log.Println(err)
				return
			}

			stream.mu.Lock()

			_, ok = stream.dataMap[stream.header.GroupID]
			if !ok {
				stream.dataMap[stream.header.GroupID] = make(map[moqtmessage.PeepID]map[moqtmessage.ObjectID]chan []byte)
			}

			_, ok = stream.dataMap[stream.header.GroupID][stream.header.PeepID]
			if !ok {
				stream.dataMap[stream.header.GroupID][stream.header.PeepID] = make(map[moqtmessage.ObjectID]chan []byte)
			}

			_, ok = stream.dataMap[stream.header.GroupID][stream.header.PeepID][stream.objectID]
			if !ok {
				stream.dataMap[stream.header.GroupID][stream.header.PeepID][stream.objectID] = make(chan []byte, 1)
			}

			stream.dataMap[stream.header.GroupID][stream.header.PeepID][chunk.ObjectID] <- chunk.Payload

			stream.mu.Unlock()
		}

	}(reader.Reader)

	return nil
}
func (stream *receiveDataStreamPeep) ForwardingPreference() moqtmessage.ObjectForwardingPreference {
	return moqtmessage.PEEP
}

func (stream *receiveDataStreamPeep) Read(buf []byte) (int, error) {
	if stream.closed {
		return 0, io.EOF
	}

	stream.mu.Lock()
	defer stream.mu.Unlock()

	// Increment the Object ID by 1
	stream.objectID++

	_, ok := stream.dataMap[stream.header.GroupID]

	if !ok {
		stream.dataMap[stream.header.GroupID] = make(map[moqtmessage.PeepID]map[moqtmessage.ObjectID]chan []byte)
	}

	_, ok = stream.dataMap[stream.header.GroupID][stream.header.PeepID]

	if !ok {
		stream.dataMap[stream.header.GroupID][stream.header.PeepID] = make(map[moqtmessage.ObjectID]chan []byte)
	}

	_, ok = stream.dataMap[stream.header.GroupID][stream.header.PeepID][stream.objectID]

	if !ok {
		stream.dataMap[stream.header.GroupID][stream.header.PeepID][stream.objectID] = make(chan []byte, 1)
	}

	payloadCh := stream.dataMap[stream.header.GroupID][stream.header.PeepID][stream.objectID]

	payload, ok := <-payloadCh
	if !ok {
		return 0, errors.New("no data")
	}

	n := copy(buf, payload)

	return n, nil
}

func (stream *receiveDataStreamPeep) NextGroup() (ReceiveDataStream, error) {
	if stream.closed {
		return nil, ErrClosedStream
	}

	//Delete
	delete(stream.readerMap, stream.header.GroupID)

	// Delete
	delete(stream.dataMap, stream.header.GroupID)

	// Increment the Group ID by 1
	stream.header.GroupID++

	// Set the Peep ID to -1
	stream.header.PeepID = 1<<64 - 1

	stream.closed = true

	return stream, nil
}

func (stream *receiveDataStreamPeep) NextStream() (ReceiveDataStream, error) {
	for {
		// Receive a receive reader
		reader := <-stream.readerCh
		_, ok := stream.readerMap[reader.GroupID]
		if !ok {
			stream.readerMap[reader.GroupID] = make(map[moqtmessage.PeepID]chan struct {
				moqtmessage.PublisherPriority
				quicvarint.Reader
			})
		}

		_, ok = stream.readerMap[reader.GroupID][reader.PeepID]
		if !ok {
			stream.readerMap[reader.GroupID][reader.PeepID] = make(chan struct {
				moqtmessage.PublisherPriority
				quicvarint.Reader
			}, 1)
		}

		stream.readerMap[reader.GroupID][reader.PeepID] <- struct {
			moqtmessage.PublisherPriority
			quicvarint.Reader
		}{
			PublisherPriority: reader.PublisherPriority,
			Reader:            reader.Reader,
		}

		_, ok = stream.dataMap[reader.GroupID]
		if !ok {
			stream.dataMap[reader.GroupID] = make(map[moqtmessage.PeepID]map[moqtmessage.ObjectID]chan []byte)
		}

		_, ok = stream.dataMap[reader.GroupID][reader.PeepID]
		if !ok {
			stream.dataMap[reader.GroupID][reader.PeepID] = make(map[moqtmessage.ObjectID]chan []byte)
		}

		_, ok = stream.readerMap[stream.header.GroupID][stream.header.PeepID]
		if ok {
			break
		}
	}

	stream.init()

	return stream, nil
}
