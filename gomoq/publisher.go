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

func (p Publisher) sendSingleObject(payload []byte) <-chan error {
	header := StreamHeaderTrack{}
	payloadCh := make(chan []byte, 1)
	defer close(payloadCh)
	payloadCh <- payload
	return p.sendMultipleObject(&header, payloadCh)
}

func (p Publisher) sendMultipleObject(header StreamHeader, payloadCh <-chan []byte) <-chan error {

	errCh := make(chan error, 1)
	stream, err := p.session.OpenUniStream()
	if err != nil {
		errCh <- err
	}

	switch header.(type) {
	case *StreamHeaderTrack:
		go func() {
			/*
			 * STREAM_HEADER_TRACK {
			 * }
			 * Group Chunks {
			 *   Group Chunk {
			 *     Group ID (0)
			 *     Ojbect Chunk
			 *   }
			 *   Group Chunk {
			 *     Group ID (1)
			 *     Ojbect Chunk
			 *   }...
			 * }
			 */
			defer close(errCh)
			headerTrack, ok := header.(*StreamHeaderTrack)
			if !ok {
				errCh <- err
			}
			// Send the header
			stream.Write(headerTrack.serialize())

			// Send chunks
			var chunk GroupChunk
			var groupID GroupID = 0
			var objectID ObjectID = 0
			for payload := range payloadCh {
				//Send a Group chunk
				chunk = GroupChunk{
					GroupID: groupID,
					ObjectChunk: ObjectChunk{
						ObjectID: objectID,
						Payload:  payload,
					},
				}
				_, err = stream.Write(chunk.serialize())
				if err != nil {
					errCh <- err
				}

				// Increment the Group ID by 1
				groupID++
			}

			// Send final chunk
			chunk = GroupChunk{
				GroupID: groupID,
				ObjectChunk: ObjectChunk{
					ObjectID:   objectID,
					Payload:    []byte{},
					StatusCode: END_OF_TRACK,
				},
			}

			_, err = stream.Write(chunk.serialize())
			if err != nil {
				errCh <- err
			}
		}()
	case *StreamHeaderPeep:
		go func() {
			/*
			 * STREAM_HEADER_PEEP {}
			 * Ojbect Chunks {
			 *   Ojbect Chunk {
			 *     Object ID (0)
			 *     Payload
			 *   }
			 *   Ojbect Chunk {
			 *     Object ID (1)
			 *     Payload
			 *   }...
			 * }
			 */
			defer close(errCh)
			headerPeep, ok := header.(*StreamHeaderPeep)
			if !ok {
				errCh <- err
			}
			// Send the header
			stream.Write(headerPeep.serialize())

			// Send chunks
			var chunk ObjectChunk
			var objectID ObjectID = 0
			for payload := range payloadCh {
				//Send a Object chunk
				chunk = ObjectChunk{
					ObjectID: objectID,
					Payload:  payload,
				}
				_, err = stream.Write(chunk.serialize())
				if err != nil {
					errCh <- err
				}

				// Increment the Object ID by 1
				objectID++
			}

			// Send final chunk
			chunk = ObjectChunk{
				ObjectID:   objectID,
				Payload:    []byte{},
				StatusCode: END_OF_PEEP,
			}
			_, err = stream.Write(chunk.serialize())
			if err != nil {
				errCh <- err
			}
		}()
	default:
		panic("invalid header")
	}

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
