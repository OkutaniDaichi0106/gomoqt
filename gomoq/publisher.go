package gomoq

import (
	"errors"
)

type Publisher struct {
	/*
	 * Client
	 * If this is not initialized, use default Client
	 */
	Client

	/***/
	PublisherHandler

	/*
	 * Track Namespace the publisher uses
	 */
	TrackNamespace string

	/***/
	TrackNames []string

	/***/
	subscriptions []SubscribeMessage
}

type PublisherHandler interface {
	AnnounceParameters() Parameters
}

var _ PublisherHandler = Publisher{}

func (p *Publisher) Connect(url string) error {
	// Check if the Client specify the Versions
	if len(p.Versions) < 1 {
		return errors.New("no versions is specifyed")
	}

	// Connect to the server
	return p.connect(url, PUB)
}

/*
 *
 *
 */
func (p Publisher) Announce(trackNamespace string) error {
	return p.sendAnnounceMessage(trackNamespace)
}

func (p Publisher) sendAnnounceMessage(trackNamespace string) error {
	a := AnnounceMessage{
		TrackNamespace: trackNamespace,
		Parameters:     p.AnnounceParameters(),
	}
	_, err := p.controlStream.Write(a.serialize())
	if err != nil {
		return err
	}

	return nil
}

func (p Publisher) sendObjectDatagram() error { //TODO:
	var od ObjectDatagram
	return p.session.SendDatagram(od.serialize())
}

func (p Publisher) sendObjects() chan error {
	errCh := make(chan error, 1)
	return errCh
}

/*
 *
 *
 */
func (p Publisher) Unannounce(trackNamespace string) error {
	return p.sendUnannounceMessage(trackNamespace)
}

func (p Publisher) sendUnannounceMessage(trackNamespace string) error {
	um := UnannounceMessage{
		TrackNamespace: trackNamespace,
	}
	_, err := p.controlStream.Write(um.serialize())
	if err != nil {
		return err
	}
	return nil
}

/*
 *
 *
 */
func (p Publisher) SubscribeDone() error { //TODO:
	sd := SubscribeDoneMessage{
		//SubscribeID: ,
		//StatusCode:,
		//Reason:,
		//ContentExists:,
		//FinalGroupID:,
		//FinalObjectID:,
	}
	_, err := p.controlStream.Write(sd.serialize())
	if err != nil {
		return err
	}

	return nil
}

/*
 * Response to a TRACK_STATUS_REQUEST
 */
func (p Publisher) sendTrackStatus() error {
	ts := TrackStatusMessage{
		TrackNamespace: p.TrackNamespace,
		TrackName:      "",
		Code:           0,
		LastGroupID:    GroupID(0),
		LastObjectID:   ObjectID(0),
	}
	p.controlStream.Write(ts.serialize())
	return nil
}

type TrackStream struct{}
type PeepStream struct{}
