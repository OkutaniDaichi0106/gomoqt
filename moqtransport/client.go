package moqtransport

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

/*
 * Client including Publisher and Subscriber
 *
 * Client will perform the following operation
 * - Connencting to the server
 * - Opening bidirectional stream to send control messages
 * - Sending SETUP_CLIENT message to the server
 * - Receiving SETUP_SERVER message from the server
 * - Terminating sessions
 */

type Client struct {
	/*
	 * TLS configuration
	 */
	TLSConfig *tls.Config

	/*
	 * Versions supported by the client
	 */
	Versions []Version

	/*
	 * Bidirectional stream to send control stream
	 * Set this after connection to the server
	 */
	session *webtransport.Session

	/*
	 * Bidirectional stream to send control stream
	 * Set the first bidirectional stream
	 */
	controlStream webtransport.Stream

	/***/
	controlReader quicvarint.Reader

	/*
	 * Using selectedVersion which is specifyed by the client and is selected by the server
	 */
	selectedVersion Version
}

/*
 * Client connect to the server
 * Dial to the server and establish a session
 * Open bidirectional stream to send control message
 *
 */
func (c *Client) connect(url string) error {
	// Set tls configuration
	if c.TLSConfig == nil {
		panic("no TLS configuration")
	}

	// Define new Dialer
	d := webtransport.Dialer{
		TLSClientConfig: c.TLSConfig,
	}

	// Set header //TODO: How to handle header
	var headers http.Header

	// Dial to the server with Extended CONNECT request
	_, sess, err := d.Dial(context.Background(), url, headers)
	if err != nil {
		return err
	}

	// Register the session to the client
	c.session = sess

	return nil
}

func (c *Client) receiveServerSetup() (Parameters, error) {
	// Receive SETUP_SERVER message
	id, err := deserializeHeader(c.controlReader)
	if err != nil {
		return nil, err
	}
	if id != SERVER_SETUP {
		return nil, ErrUnexpectedMessage
	}

	var ss ServerSetupMessage
	err = ss.deserializeBody(c.controlReader)
	if err != nil {
		return nil, err
	}

	// Check specified version is selected
	err = contain(ss.SelectedVersion, c.Versions)
	if err != nil {
		return nil, err
	}

	// Register the selected version
	c.selectedVersion = ss.SelectedVersion

	return ss.Parameters, nil
}

// func (c *Client) Terminate(code TerminateErrorCode, reason string) error {
// 	// Send Error message to the server before close the stream
// 	c.session.CloseWithError(webtransport.SessionErrorCode(code), reason)

// 	//c.session.CloseWithError() //TODO:
// 	return nil
// }
