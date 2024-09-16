package moqtransport

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/quic-go/quic-go/quicvarint"
)

type Publisher struct {
	/*
	 * Client
	 */
	Client

	/*
	 * Track Namespace the publisher uses
	 */
	TrackNamespace TrackNamespace

	/***/
	TrackNames []string

	MaxSubscribeID subscribeID

	announceParameters Parameters
}

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
	params, err := p.setup()
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (p *Publisher) setup() (Parameters, error) {
	var err error

	// Open first stream to send control messages
	p.controlStream, err = p.session.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}

	// Get control reader
	p.controlReader = quicvarint.NewReader(p.controlStream)

	// Send SETUP_CLIENT message
	err = p.sendClientSetup()
	if err != nil {
		return nil, err
	}

	// Receive SETUP_SERVER message
	return p.receiveServerSetup()
}

func (p Publisher) sendClientSetup() error {
	// Initialize SETUP_CLIENT message
	csm := ClientSetupMessage{
		Versions:   p.Versions,
		Parameters: make(Parameters),
	}

	// Add role parameter
	csm.AddParameter(ROLE, PUB)

	// Add max subscribe id parameter
	csm.AddParameter(MAX_SUBSCRIBE_ID, p.MaxSubscribeID)

	_, err := p.controlStream.Write(csm.serialize())

	return err
}

/*
 *
 *
 */
func (p *Publisher) Announce(trackNamespace ...string) error {
	if p.announceParameters == nil {
		p.announceParameters = make(Parameters)
	}
	// Send ANNOUNCE message
	am := AnnounceMessage{
		TrackNamespace: trackNamespace,
		Parameters:     p.announceParameters,
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
		givenFullTrackNamespace := strings.Join(trackNamespace, "")
		receivedFullTrackNamespace := strings.Join(ao.TrackNamespace, "")
		if givenFullTrackNamespace != receivedFullTrackNamespace {
			return errors.New("unexpected Track Namespace")
		}

		// Register the Track Namespace
		p.TrackNamespace = ao.TrackNamespace

	case ANNOUNCE_ERROR:
		var ae AnnounceErrorMessage // TODO: Handle Error Code
		err = ae.deserializeBody(p.controlReader)
		if err != nil {
			return err
		}

		// Check if the Track Namespace is rejected by the server
		givenFullTrackNamespace := strings.Join(trackNamespace, "")
		receivedFullTrackNamespace := strings.Join(ae.TrackNamespace, "")
		if givenFullTrackNamespace != receivedFullTrackNamespace {
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
