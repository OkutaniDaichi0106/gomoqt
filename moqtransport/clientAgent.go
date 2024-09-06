package moqtransport

import (
	"context"
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

type AgenterWithRole interface {
	Role() Role
}

type clientAgent struct {
	//Server *Server
	/*
	 * WebTransport Session
	 */
	session *webtransport.Session

	/*
	 * Bidirectional stream to send control message on
	 */
	controlStream webtransport.Stream

	/*
	 *
	 */
	controlReader quicvarint.Reader

	/*
	 *
	 */
	controlCh chan []byte

	/*
	 * MOQT version
	 */
	version Version

	/*
	 * A Map of the Track Alias using two keys
	 * Get Track Alias by specifying the Track Namespace as the first key
	 * and the Track Name as the second key.
	 */
	trackAliases map[string]map[string]TrackAlias
}

/*
 * Initialize the Client Agent
 */
func (a *clientAgent) init() error {
	// Initialize the channel to send and receive control messages
	a.controlCh = make(chan []byte, 1<<5) //TODO: Tune the size

	return nil
}

// func (a *Agent) listenControlChannel() chan error {
// 	errCh := make(chan error, 1)
// 	go func() {
// 		for data := range a.controlCh {
// 			switch MessageID(data[0]) {
// 			case SUBSCRIBE:
// 				// Check if the subscribe is acceptable
// 			case SUBSCRIBE_OK:
// 				// Send it to the Subscriber
// 				_, err := a.controlStream.Write(data)
// 				if err != nil {
// 					errCh <- err
// 					return
// 				}
// 			case SUBSCRIBE_ERROR:
// 				// Send it to the Subscriber
// 				_, err := a.controlStream.Write(data)
// 				if err != nil {
// 					errCh <- err
// 					return
// 				}
// 			case UNSUBSCRIBE:
// 				// Delete the Subscriber from the destinations
// 			default:
// 				errCh <- ErrUnexpectedMessage //TODO: handle the error as protocol violation
// 			}
// 		}
// 	}()
// 	return errCh
// }

func Activate(a AgenterWithRole) error {

}

func Setup(sess *webtransport.Session) (AgenterWithRole, error) {
	// Create bidirectional stream to send control messages
	stream, err := sess.AcceptStream(context.Background())
	if err != nil {
		return nil, err
	}

	reader := quicvarint.NewReader(stream)

	// Receive SETUP_CLIENT message
	id, err := deserializeHeader(reader)
	if err != nil {
		return nil, err
	}
	if id != CLIENT_SETUP {
		return nil, ErrProtocolViolation
	}
	var cs ClientSetupMessage
	err = cs.deserializeBody(reader)
	if err != nil {
		return nil, err
	}

	// Check if the ROLE parameter is valid
	// Register the Client's role to the Agent
	v, ok := cs.Parameters.Contain(ROLE)
	if !ok {
		return nil, errors.New("no role is specified")
	}
	num, ok := v.(uint64)
	if !ok {
		return nil, errors.New("invalid value")
	}
	role := Role(num)

	// Select a version
	version, err := selectVersion(cs.Versions, SERVER.SupportedVersions)
	if err != nil {
		return nil, err
	}

	// Initialise SETUP_SERVER message
	ssm := ServerSetupMessage{
		SelectedVersion: version,
		Parameters:      SERVER.setupParameters,
	}

	// Send SETUP_SERVER message
	_, err = stream.Write(ssm.serialize())
	if err != nil {
		return nil, err
	}

	ca := clientAgent{
		// Register the session
		session: sess,
		// Register the stream as control stream
		controlStream: stream,
		// Create quic varint reader of control message and register it
		controlReader: reader,
		// Set the client version
		version: version,
	}

	switch role {
	case PUB:
		return &PublisherAgent{
			clientAgent: ca,
		}, nil
	case SUB:
		return &SubscriberAgent{
			clientAgent: ca,
		}, nil
	case PUB_SUB:
		return &PubSubAgent{
			clientAgent: ca,
		}, nil
	default:
		return nil, ErrInvalidRole
	}
}

var ErrProtocolViolation = errors.New("protocol violation")
