package moqtransport

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go/quicvarint"
)

type Publisher struct {
	node    node
	session *PublishingSession

	MaxSubscribeID uint64
}

func (p *Publisher) ConnectAndSetup(url string) (*PublishingSession, error) {
	sess, err := p.node.EstablishPubSession(url, p.MaxSubscribeID)
	if err != nil {
		return nil, err
	}

	p.session = sess

	return sess, nil
}

func (p Publisher) SendDatagram(header moqtmessage.StreamHeaderDatagram, payload []byte) error {
	// Serialize the header
	b := header.Serialize()

	// Serialize the payload
	b = quicvarint.Append(b, uint64(len(payload)))
	b = append(b, payload...)

	// Send the payload as a datagram
	return p.session.trSess.SendDatagram(b)
}

func NewStreamHeaderDatagram(subscription Subscription, priority moqtmessage.PublisherPriority) moqtmessage.StreamHeaderDatagram {
	if subscription.forwardingPreference == nil {
		subscription.forwardingPreference = moqtmessage.DATAGRAM
	}

	if subscription.forwardingPreference != moqtmessage.DATAGRAM {
		panic("dont change the object forwarding preference")
	}

	return moqtmessage.StreamHeaderDatagram{
		SubscribeID:       subscription.subscribeID,
		TrackAlias:        subscription.trackAlias,
		PublisherPriority: priority,
	}
}

func NewStreamHeaderTrack(subscription Subscription, priority moqtmessage.PublisherPriority) moqtmessage.StreamHeaderTrack {
	if subscription.forwardingPreference == nil {
		subscription.forwardingPreference = moqtmessage.TRACK
	}

	if subscription.forwardingPreference != moqtmessage.TRACK {
		panic("dont change the object forwarding preference")
	}

	return moqtmessage.StreamHeaderTrack{
		SubscribeID:       subscription.subscribeID,
		TrackAlias:        subscription.trackAlias,
		PublisherPriority: priority,
	}
}

func NewStreamHeaderPeep(subscription Subscription, priority moqtmessage.PublisherPriority) moqtmessage.StreamHeaderPeep {
	if subscription.forwardingPreference == nil {
		subscription.forwardingPreference = moqtmessage.PEEP
	}

	if subscription.forwardingPreference != moqtmessage.PEEP {
		panic("dont change the object forwarding preference")
	}

	return moqtmessage.StreamHeaderPeep{
		SubscribeID:       subscription.subscribeID,
		TrackAlias:        subscription.trackAlias,
		PublisherPriority: priority,
		GroupID:           0,
		PeepID:            0,
	}
}

func (p Publisher) OpenStreamTrack(header moqtmessage.StreamHeaderTrack) (*SendDataStreamTrack, error) {
	writer, err := p.session.trSess.OpenUniStream()
	if err != nil {
		return nil, err
	}

	// Send a Stream Header
	_, err = writer.Write(header.Serialize())
	if err != nil {
		return nil, err
	}

	return &SendDataStreamTrack{
		closed:   false,
		header:   header,
		writer:   writer,
		groupID:  0,
		objectID: 0,
	}, nil
}

func (p Publisher) OpenStreamPeep(header moqtmessage.StreamHeaderPeep) (*SendDataStreamPeep, error) {
	writer, err := p.session.trSess.OpenUniStream()
	if err != nil {
		return nil, err
	}

	// Send a Stream Header
	_, err = writer.Write(header.Serialize())
	if err != nil {
		return nil, err
	}

	return &SendDataStreamPeep{
		closed:   false,
		writer:   writer,
		header:   header,
		objectID: 0,
	}, nil
}
