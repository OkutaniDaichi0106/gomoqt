package moqtransport

import (
	"errors"
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

func (p *Publisher) NewTrack(subscription Subscription, forwardingPreference moqtmessage.ObjectForwardingPreference, priority moqtmessage.PublisherPriority) (SendDataStream, error) {
	// Get the Transport Session
	if p.session.trSess == nil {
		return nil, errors.New("no connection")
	}

	switch forwardingPreference {
	case moqtmessage.DATAGRAM:
		return &sendDataStreamDatagram{
			closed:            false,
			trSess:            p.session.trSess,
			SubscribeID:       subscription.subscribeID,
			TrackAlias:        subscription.trackAlias,
			PublisherPriority: priority,
			groupID:           0,
			objectID:          0,
		}, nil

	case moqtmessage.TRACK:
		header := moqtmessage.StreamHeaderTrack{
			SubscribeID:       subscription.subscribeID,
			TrackAlias:        subscription.trackAlias,
			PublisherPriority: priority,
		}

		writer, err := p.session.trSess.OpenUniStream()
		if err != nil {
			return nil, err
		}
		return &sendDataStreamTrack{
			closed:       false,
			writerClosed: false,
			trSess:       p.session.trSess,
			writer:       writer,
			header:       header,
			groupID:      0,
			objectID:     0,
		}, nil
	case moqtmessage.PEEP:
		header := moqtmessage.StreamHeaderPeep{
			SubscribeID:       subscription.subscribeID,
			TrackAlias:        subscription.trackAlias,
			PublisherPriority: priority,
			GroupID:           0,
			PeepID:            0,
		}

		writer, err := p.session.trSess.OpenUniStream()
		if err != nil {
			return nil, err
		}

		return &sendDataStreamPeep{
			closed:   false,
			trSess:   p.session.trSess,
			writer:   writer,
			header:   header,
			objectID: 0,
		}, nil
	default:
		panic("invalid forwarding preference")
	}
}

// func (p Publisher) SendObjectDatagram(od moqtmessage.ObjectDatagram) error { //TODO:
// 	return p.session.SendDatagram(od.Serialize())
// }

// func (p Publisher) SendSingleObject(priority moqtmessage.PublisherPriority, payload []byte) <-chan error {
// 	dataCh := make(chan []byte, 1)
// 	defer close(dataCh)

// 	header := moqtmessage.StreamHeaderTrack{
// 		//subscribeID: ,
// 		//TrackAlias: ,
// 		PublisherPriority: priority,
// 	}

// 	dataCh <- payload

// 	return p.sendMultipleObject(&header, dataCh)
// }

// func (p Publisher) SendMultipleObject(priority moqtmessage.PublisherPriority, payload <-chan []byte) <-chan error {
// 	header := moqtmessage.StreamHeaderTrack{
// 		//subscribeID: ,
// 		//TrackAlias: ,
// 		PublisherPriority: priority,
// 	}
// 	return p.sendMultipleObject(&header, payload) // TODO:
// }

/*
 *
 *
 */

// /*
//  * Response to a TRACK_STATUS_REQUEST
//  */
// func (p Publisher) sendTrackStatus() error {
// 	ts := moqtmessage.TrackStatusMessage{
// 		TrackNamespace: p.TrackNamespace,
// 		TrackName:      "",
// 		Code:           0,
// 		LastGroupID:    0, // TODO
// 		LastObjectID:   0, // TODO
// 	}
// 	p.controlStream.Write(ts.Serialize())
// 	return nil
// }
