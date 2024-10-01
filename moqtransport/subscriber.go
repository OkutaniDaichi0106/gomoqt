package moqtransport

import (
	"bytes"
	"context"
	"errors"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go/quicvarint"
)

type Subscriber struct {
	node node

	session *SubscribingSession

	streamMap map[moqtmessage.SubscribeID]chan struct {
		header moqtmessage.StreamHeader
		reader quicvarint.Reader
	}
}

func (s *Subscriber) ConnectAndSetup(URL string) (*SubscribingSession, error) {
	sess, err := s.node.EstablishSubSession(URL)
	if err != nil {
		return nil, err
	}

	s.session = sess

	s.streamMap = make(map[moqtmessage.SubscribeID]chan struct {
		header moqtmessage.StreamHeader
		reader quicvarint.Reader
	})

	return sess, nil
}

func (s Subscriber) ReceiveDatagram(ctx context.Context) (ReceiveDataStream, error) {
	b, err := s.session.trSess.ReceiveDatagram(ctx)
	if err != nil {
		return nil, err
	}

	reader := quicvarint.NewReader(bytes.NewReader(b))

	var header moqtmessage.StreamHeaderDatagram
	err = header.DeserializeStreamHeaderBody(reader)
	if err != nil {
		return nil, err
	}

	return &receiveDataStreamDatagram{
		header: header,
		reader: reader,
	}, nil
}

func (s Subscriber) AcceptUniStream(ctx context.Context) (ReceiveDataStream, error) {
	stream, err := s.session.trSess.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}

	reader := quicvarint.NewReader(stream)

	id, err := moqtmessage.DeserializeStreamTypeID(reader)
	if err != nil {
		return nil, err
	}

	switch id {
	case moqtmessage.TRACK_ID:
		var header moqtmessage.StreamHeaderTrack
		err := header.DeserializeStreamHeaderBody(reader)
		if err != nil {
			return nil, err
		}

		return &receiveDataStreamTrack{
			header: header,
			reader: reader,
		}, nil
	case moqtmessage.PEEP_ID:
		var header moqtmessage.StreamHeaderPeep
		err := header.DeserializeStreamHeaderBody(reader)
		if err != nil {
			return nil, err
		}

		return &receiveDataStreamPeep{
			header: header,
			reader: reader,
		}, nil
	default:
		return nil, errors.New("invalid stream type")
	}
}
