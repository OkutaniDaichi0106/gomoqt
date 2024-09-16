package moqtransport

import (
	"context"
	"errors"
	"go-moq/moqtransport/moqterror"
	"go-moq/moqtransport/moqtmessage"
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
	TrackNamespace moqtmessage.TrackNamespace

	/***/
	TrackNames []string

	MaxSubscribeID moqtmessage.SubscribeID

	announceParameters moqtmessage.Parameters
}

func (p *Publisher) ConnectAndSetup(url string) (moqtmessage.Parameters, error) {
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

func (p *Publisher) setup() (moqtmessage.Parameters, error) {
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
	csm := moqtmessage.ClientSetupMessage{
		Versions:   p.Versions,
		Parameters: make(moqtmessage.Parameters),
	}

	// Add role parameter
	csm.AddParameter(moqtmessage.ROLE, moqtmessage.PUB)

	// Add max subscribe id parameter
	csm.AddParameter(moqtmessage.MAX_SUBSCRIBE_ID, p.MaxSubscribeID)

	_, err := p.controlStream.Write(csm.Serialize())

	return err
}

/*
 *
 *
 */
func (p *Publisher) Announce(trackNamespace ...string) error {
	if p.announceParameters == nil {
		p.announceParameters = make(moqtmessage.Parameters)
	}
	// Send ANNOUNCE message
	am := moqtmessage.AnnounceMessage{
		TrackNamespace: trackNamespace,
		Parameters:     p.announceParameters,
	}

	_, err := p.controlStream.Write(am.Serialize())
	if err != nil {
		return err
	}

	//Receive ANNOUNCE_OK message or ANNOUNCE_ERROR message
	id, err := moqtmessage.DeserializeMessageID(p.controlReader)
	if err != nil {
		return err
	}

	switch id {
	case moqtmessage.ANNOUNCE_OK:
		var ao moqtmessage.AnnounceOkMessage
		err = ao.DeserializeBody(p.controlReader)
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

	case moqtmessage.ANNOUNCE_ERROR:
		var ae moqterror.AnnounceErrorMessage // TODO: Handle Error Code
		err = ae.DeserializeBody(p.controlReader)
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

func (p Publisher) SendObjectDatagram(od moqtmessage.ObjectDatagram) error { //TODO:
	return p.session.SendDatagram(od.Serialize())
}

func (p Publisher) SendSingleObject(priority moqtmessage.PublisherPriority, payload []byte) <-chan error {
	dataCh := make(chan []byte, 1)
	defer close(dataCh)

	header := moqtmessage.StreamHeaderTrack{
		//subscribeID: ,
		//TrackAlias: ,
		PublisherPriority: priority,
	}

	dataCh <- payload

	return p.sendMultipleObject(&header, dataCh)
}

func (p Publisher) SendMultipleObject(priority moqtmessage.PublisherPriority, payload <-chan []byte) <-chan error {
	header := moqtmessage.StreamHeaderTrack{
		//subscribeID: ,
		//TrackAlias: ,
		PublisherPriority: priority,
	}
	return p.sendMultipleObject(&header, payload) // TODO:
}

func (p *Publisher) sendMultipleObject(header moqtmessage.StreamHeader, payloadCh <-chan []byte) <-chan error {

	errCh := make(chan error, 1)
	stream, err := p.session.OpenUniStream()
	if err != nil {
		errCh <- err
	}

	go func() {
		// Send the header
		_, err := stream.Write(header.Serialize())
		if err != nil {
			log.Println(err)
			errCh <- err
			return
		}

		// Get chunk stream to get chunks
		chunkStream := moqtmessage.NewChunkStream(header)
		var chunk moqtmessage.Chunk
		for payload := range payloadCh {
			chunk = chunkStream.CreateChunk(payload)
			_, err = stream.Write(chunk.Serialize())
			if err != nil {
				log.Println(err)
				errCh <- err
				return
			}
		}

		// Send final chunk as end of the stream
		chunk = chunkStream.CreateFinalChunk()
		_, err = stream.Write(chunk.Serialize())
		if err != nil {
			log.Println(err)
			errCh <- err
			return
		}
	}()

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
	um := moqtmessage.UnannounceMessage{
		TrackNamespace: trackNamespace,
	}
	_, err := p.controlStream.Write(um.Serialize())
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
	sd := moqtmessage.SubscribeDoneMessage{
		//SubscribeID: ,
		//StatusCode:,
		//Reason:,
		//ContentExists:,
		//FinalGroupID:,
		//FinalObjectID:,
	}
	_, err := p.controlStream.Write(sd.Serialize())
	if err != nil {
		return err
	}

	return nil
}

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
