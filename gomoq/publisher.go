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

	Namespace string
}

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
	return p.sendAnnounceMessage(trackNamespace, Parameters{})
}

// func (p Publisher) AnnounceWithParams(trackNamespace string, params Parameters) error {
// 	return p.sendAnnounceMessage(trackNamespace, params)
// }

func (p Publisher) sendAnnounceMessage(trackNamespace string, params Parameters) error {
	a := AnnounceMessage{
		TrackNamespace: trackNamespace,
		Parameters:     params,
	}
	_, err := p.controlStream.Write(a.serialize())
	if err != nil {
		return err
	}

	return nil
}

func (p Publisher) AllowSubscribe() error {
	//
	return p.sendSubscribeOk()
}

func (p Publisher) sendSubscribeOk() error {
	return nil
}

func (p Publisher) RejectSubscribe() error {
	return p.sendSubscribeError()
}

func (p Publisher) sendSubscribeError() error {
	return nil
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
	p.controlStream.Write(um.serialize())
	return nil
}

/*
 *
 *
 */
func (p Publisher) SubscribeDone() error {
	return nil
}

/*
 * Response to a TRACK_STATUS_REQUEST
 */
func (p Publisher) sendTrackStatus() error {
	ts := TrackStatusMessage{
		TrackNamespace: p.Namespace,
		TrackName:      "",
		Code:           0,
		LastGroupID:    GroupID(0),
		LastObjectID:   ObjectID(0),
	}
	p.controlStream.Write(ts.serialize())
	return nil
}
