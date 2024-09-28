package moqtransport

import (
	"go-moq/moqtransport/moqtmessage"
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

func (p Publisher) NewStreamDatagram(subscription Subscription, priority moqtmessage.PublisherPriority) (*SendDataStreamDatagram, error) {
	if subscription.forwardingPreference != nil {
		panic("object forwarding preference could not change")
	}

	// Set the Object Forwarding Preference
	subscription.forwardingPreference = moqtmessage.DATAGRAM

	header := moqtmessage.StreamHeaderDatagram{
		SubscribeID:       subscription.subscribeID,
		TrackAlias:        subscription.trackAlias,
		PublisherPriority: priority,
	}

	return &SendDataStreamDatagram{
		closed:   false,
		trSess:   p.session.trSess,
		header:   header,
		groupID:  0,
		objectID: 0,
	}, nil
}

func (p Publisher) NewStreamTrack(subscription Subscription, priority moqtmessage.PublisherPriority) (*SendDataStreamTrack, error) {
	if subscription.forwardingPreference != nil {
		panic("object forwarding preference could not change")
	}

	// Set the Object Forwarding Preference
	subscription.forwardingPreference = moqtmessage.TRACK

	writer, err := p.session.trSess.OpenUniStream()
	if err != nil {
		return nil, err
	}

	// Send a Stream Header
	header := moqtmessage.StreamHeaderTrack{
		SubscribeID:       subscription.subscribeID,
		TrackAlias:        subscription.trackAlias,
		PublisherPriority: priority,
	}
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

func (p Publisher) NextTrackStream(stream SendDataStreamTrack) (*SendDataStreamTrack, error) {
	writer, err := p.session.trSess.OpenUniStream()
	if err != nil {
		return nil, err
	}

	return &SendDataStreamTrack{
		closed:   false,
		header:   stream.header,
		writer:   writer,
		groupID:  stream.groupID + 1,
		objectID: 0,
	}, nil
}

func (p Publisher) NewStreamPeep(subscription Subscription, priority moqtmessage.PublisherPriority) (*SendDataStreamPeep, error) {
	if subscription.forwardingPreference != nil {
		panic("object forwarding preference could not change")
	}

	// Set the Object Forwarding Preference
	subscription.forwardingPreference = moqtmessage.PEEP
	writer, err := p.session.trSess.OpenUniStream()
	if err != nil {
		return nil, err
	}

	// Send a Stream Header
	header := moqtmessage.StreamHeaderPeep{
		SubscribeID:       subscription.subscribeID,
		TrackAlias:        subscription.trackAlias,
		PublisherPriority: priority,
		GroupID:           0,
		PeepID:            0,
	}
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

func (p Publisher) NextGroupStream(stream SendDataStreamPeep) (*SendDataStreamPeep, error) {
	writer, err := p.session.trSess.OpenUniStream()
	if err != nil {
		return nil, err
	}

	stream.header.GroupID++
	stream.header.PeepID = 0

	return &SendDataStreamPeep{
		closed:   false,
		writer:   writer,
		header:   stream.header,
		objectID: 0,
	}, nil
}

func (p Publisher) NextPeepStream(stream SendDataStreamPeep) (*SendDataStreamPeep, error) {
	writer, err := p.session.trSess.OpenUniStream()
	if err != nil {
		return nil, err
	}

	stream.header.PeepID++

	return &SendDataStreamPeep{
		closed:   false,
		writer:   writer,
		header:   stream.header,
		objectID: 0,
	}, nil
}
