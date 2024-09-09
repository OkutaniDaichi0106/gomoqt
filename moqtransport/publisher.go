package moqtransport

import (
	"errors"
	"log"
)

type Publisher struct {
	/*
	 * Client
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
}

type PublisherHandler interface {
	AnnounceParameters() Parameters
}

// Check the Publisher inplement Publisher Handler
var _ PublisherHandler = Publisher{}

func (p *Publisher) ConnectAndSetup(url string) (Parameters, error) {
	// Check if the Client specify the Versions
	if len(p.Versions) < 1 {
		return nil, errors.New("no versions is specifyed")
	}

	// Connect
	err := p.connect(url)
	if err != nil {
		return nil, err
	}

	// Setup
	params, err := p.setup(PUB)
	if err != nil {
		return nil, err
	}
	// TODO: handle params

	return params, nil
}

/*
 *
 *
 */
func (p *Publisher) Announce(trackNamespace string) error {
	// Send ANNOUNCE message
	am := AnnounceMessage{
		TrackNamespace: trackNamespace,
		Parameters:     p.AnnounceParameters(),
	}

	_, err := p.controlStream.Write(am.serialize())
	if err != nil {
		return err
	}

	//Receive ANNOUNCE_OK message or ANNOUNCE_ERROR message
	id, err := deserializeHeader(p.controlReader)
	if err != nil {
		return err
	}
	switch id {
	case ANNOUNCE_OK:
		var ao AnnounceOkMessage
		err = ao.deserializeBody(p.controlReader)
		if err != nil {
			return err
		}
		// Check if the Track Namespace is accepted by the server
		if trackNamespace != ao.TrackNamespace {
			return errors.New("unexpected Track Namespace")
		}

		// Register the Track Namespace
		p.TrackNamespace = ao.TrackNamespace

	case ANNOUNCE_ERROR:
		var ae AnnounceError // TODO: Handle Error Code
		err = ae.deserializeBody(p.controlReader)
		if err != nil {
			return err
		}
		// Check if the Track Namespace is rejected by the server
		if trackNamespace != ae.TrackNamespace {
			return errors.New("unexpected Track Namespace")
		}

		return errors.New(ae.Reason)
	default:
		return ErrUnexpectedMessage
	}

	return nil
}

func (p Publisher) SendObjectDatagram(od ObjectDatagram) error { //TODO:
	return p.session.SendDatagram(od.serialize())
}

func (p Publisher) SendSingleObject(priority PublisherPriority, payload []byte) <-chan error {

	header := StreamHeaderTrack{
		//subscribeID: ,
		//TrackAlias: ,
		PublisherPriority: priority,
	}
	return p.sendSingleObject(header, payload)
}

func (p Publisher) sendSingleObject(header StreamHeaderTrack, payload []byte) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		// Close the error channel when the goroutine ends
		defer close(errCh)

		stream, err := p.session.OpenUniStream()
		if err != nil {
			errCh <- err
			return
		}
		log.Println("Opened!!", stream)
		defer stream.Close()

		/*
		 * STREAM_HEADER_TRACK {
		 * }
		 * Group Chunk {
		 *   Group Chunk {
		 *     Group ID (0)
		 *     Ojbect Chunk
		 *   }
		 * }
		 */

		// Send the header
		_, err = stream.Write(header.serialize())
		if err != nil {
			errCh <- err
			return
		}

		//Send a Group chunk
		chunk := GroupChunk{
			groupID: 0,
			ObjectChunk: ObjectChunk{
				objectID: 0,
				Payload:  payload,
			},
		}
		_, err = stream.Write(chunk.serialize())
		if err != nil {
			errCh <- err
			return
		}

		// Send final chunk
		finalChunk := GroupChunk{
			groupID: 1,
			ObjectChunk: ObjectChunk{
				objectID:   0,
				Payload:    []byte{},
				StatusCode: END_OF_TRACK,
			},
		}

		_, err = stream.Write(finalChunk.serialize())
		if err != nil {
			errCh <- err
			return
		}

		log.Println("FINISH SENDING!!")

		errCh <- nil
	}()

	return errCh
}

func (p *Publisher) sendMultipleObject(header StreamHeader, payloadCh <-chan []byte) <-chan error {

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
			// Send the header
			stream.Write(header.serialize())

			// Send chunks
			var chunk GroupChunk
			var groupID groupID = 0
			var objectID objectID = 0
			for payload := range payloadCh {
				//Send a Group chunk
				chunk = GroupChunk{
					groupID: groupID,
					ObjectChunk: ObjectChunk{
						objectID: objectID,
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
				groupID: groupID,
				ObjectChunk: ObjectChunk{
					objectID:   objectID,
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
			// Send the header
			stream.Write(header.serialize())

			// Send chunks
			var chunk ObjectChunk
			var objectID objectID = 0
			for payload := range payloadCh {
				//Send a Object chunk
				chunk = ObjectChunk{
					objectID: objectID,
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
				objectID:   objectID,
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
		LastGroupID:    groupID(0),
		LastObjectID:   objectID(0),
	}
	p.controlStream.Write(ts.serialize())
	return nil
}
