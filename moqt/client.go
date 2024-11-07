package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/url"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

type Client struct {
	URL string

	SupportedVersions []Version

	TLSConfig *tls.Config

	QUICConfig *quic.Config

	SetupParameters Parameters

	HandleSetupResponce func(SetupResponce) error

	DataHandler

	RequestHandler

	RelayManager *RelayManager
}

type DataHandler interface {
	HandleData(Group, ReceiveStream)
}

func (c Client) Run(ctx context.Context) error {
	/*
	 * Initialize the Client
	 */
	if c.RelayManager == nil {
		c.RelayManager = defaultRelayManager
	}

	if c.SupportedVersions == nil {
		c.SupportedVersions = []Version{Default}
	}

	/*
	 * Dial
	 */
	// Verify the URI
	parsedURL, err := url.ParseRequestURI(c.URL)
	if err != nil {
		slog.Error("failed to parse the url", slog.String("error", err.Error()))
		return err
	}

	// Handle the scheme
	var conn Connection
	var req SetupRequest
	switch parsedURL.Scheme {
	case "https":
		// Dial with webtransport
		var d webtransport.Dialer
		_, sess, err := d.Dial(ctx, c.URL, http.Header{}) // TODO: configure the header

		if err != nil {
			slog.Error("failed to dial with webtransport", slog.String("error", err.Error()))
			return err
		}

		conn = NewMOWTConnection(sess)

		req = SetupRequest{
			SupportedVersions: c.SupportedVersions,
			Parameters:        c.SetupParameters,
		}
	case "moqt":
		// Dial with raw quic
		ips, err := net.LookupIP(parsedURL.Hostname())
		if err != nil {
			slog.Error("failed to look up IP address", slog.String("error", err.Error()))
			return err
		}

		// Try all IPs
		for i, ip := range ips {
			addr := ip.String() + parsedURL.Port()
			qconn, err := quic.DialAddrEarly(ctx, addr, c.TLSConfig, c.QUICConfig)
			if err != nil {
				slog.Error("failed to dial with quic", slog.String("error", err.Error()))
				if i+1 >= len(ips) {
					err = errors.New("no more IPs")
					slog.Error("failed to dial to the host", slog.String("error", err.Error()), slog.String("host", parsedURL.Hostname()))
					return err
				}
				continue
			}

			conn = NewMORQConnection(qconn)

			break
		}

		req = SetupRequest{
			Path:              parsedURL.Path,
			SupportedVersions: c.SupportedVersions,
			Parameters:        c.SetupParameters,
		}
	default:
		err := errors.New("invalid scheme")
		slog.Error("url scheme must be https or moqt", slog.String("scheme", parsedURL.Scheme))
		return err
	}

	/*
	 * Get a Session
	 */
	// Open a bidirectional Stream for the Session Stream
	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return err
	}
	// Initialize a Session
	sess := Session{
		Connection:            conn,
		SessionStream:         stream,
		subscribeWriters:      make(map[SubscribeID]*SubscribeWriter),
		receivedSubscriptions: make(map[SubscribeID]Subscription),
		terrCh:                make(chan TerminateError),
	}

	/*
	 * Set up
	 */
	// Send the Session Stream Type
	_, err = sess.SessionStream.Write([]byte{byte(SESSION)})
	if err != nil {
		slog.Error("failed to send a Session Stream Type", slog.String("error", err.Error()))
		return err
	}

	err = sendSetupRequest(sess.SessionStream, req)
	if err != nil {
		slog.Error("failed to request to set up", slog.String("error", err.Error()))
		return err
	}

	rsp, err := getSetupResponce(quicvarint.NewReader(sess.SessionStream))
	if err != nil {
		slog.Error("failed to receive a SESSION_SERVER message", slog.String("error", err.Error()))
		return err
	}

	// Verify the selceted version is contained in the
	if !ContainVersion(rsp.SelectedVersion, req.SupportedVersions) {
		err = errors.New("unexpected version was seleted")
		slog.Error("failed to negotiate versions", slog.String("error", err.Error()), slog.Any("selected version", rsp.SelectedVersion))
		return err
	}

	// Handle the responce
	if c.HandleSetupResponce != nil {
		err := c.HandleSetupResponce(rsp)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
	}

	/***/
	go c.listenBiStreams(&sess)
	go c.listenUniStreams(&sess)

	go func() {
		terr := <-sess.terrCh
		if terr != nil {
			sess.Terminate(terr)
		}
	}()

	return nil
}

func sendSetupRequest(stream Stream, req SetupRequest) error {
	scm := message.SessionClientMessage{
		SupportedVersions: make([]protocol.Version, 0),
	}

	for _, v := range req.SupportedVersions {
		scm.SupportedVersions = append(scm.SupportedVersions, protocol.Version(v))
	}

	scm.Parameters.Add(PATH, req.Path)

	_, err := stream.Write(scm.SerializePayload())
	if err != nil {
		slog.Error("failed to send a SESSION_CLIENT message", err.Error())
		return err
	}

	return nil
}

func getSetupResponce(r quicvarint.Reader) (SetupResponce, error) {
	/***/
	var ssm message.SessionServerMessage
	err := ssm.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read a SESSION_SERVER message", slog.String("error", err.Error()))
		return SetupResponce{}, err
	}

	return SetupResponce{
		SelectedVersion: Version(ssm.SelectedVersion),
		Parameters:      Parameters(ssm.Parameters),
	}, nil
}

func (c Client) listenBiStreams(sess *Session) {
	for {
		stream, err := sess.Connection.AcceptStream(context.Background())
		if err != nil {
			slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
			return
		}

		go func(stream Stream) {
			qvr := quicvarint.NewReader(stream)

			num, err := qvr.ReadByte()
			if err != nil {
				slog.Error("failed to read a Stream Type ID", slog.String("error", err.Error()))
			}

			switch StreamType(num) {
			case ANNOUNCE:
				slog.Debug("Announce Stream is opened")

				interest, err := getInterest(qvr)
				if err != nil {
					slog.Error("failed to get an Interest", slog.String("error", err.Error()))
					return
				}

				w := AnnounceWriter{
					doneCh: make(chan struct{}),
					stream: stream,
				}

				// Announce
				announcements, ok := c.RelayManager.GetAnnouncements(interest.TrackPrefix)
				if !ok || announcements == nil {
					announcements = make([]Announcement, 0)
				}

				c.HandleInterest(interest, announcements, w)
				<-w.doneCh
			case SUBSCRIBE:
				slog.Debug("Subscribe Stream was opened")

				subscription, err := getSubscription(qvr)
				if err != nil {
					slog.Debug("failed to get a subscription", slog.String("error", err.Error()))

					// Close the Stream gracefully
					slog.Debug("closing a Subscribe Stream", slog.String("error", err.Error()))
					err = stream.Close()
					if err != nil {
						slog.Debug("failed to close the stream", slog.String("error", err.Error()))
						return
					}

					return
				}

				//
				sw := SubscribeResponceWriter{
					stream: stream,
					doneCh: make(chan struct{}),
				}

				// Get any Infomation of the track
				info, ok := c.RelayManager.GetInfo(subscription.TrackNamespace, subscription.TrackName)
				if ok {
					// Handle with out
					c.HandleSubscribe(subscription, &info, sw)
				} else {
					c.HandleSubscribe(subscription, nil, sw)
				}

				<-sw.doneCh

				/*
				 * Accept the new subscription
				 */
				sess.acceptSubscription(subscription)

				/*
				 * Catch any Subscribe Update or any error from the subscriber
				 */
				for {
					update, err := getSubscribeUpdate(subscription, qvr)
					if err != nil {
						slog.Debug("catched an error from the subscriber", slog.String("error", err.Error()))
						break
					}

					slog.Debug("received a subscribe update request", slog.Any("subscription", update))

					sw := SubscribeResponceWriter{
						stream: stream,
						doneCh: make(chan struct{}),
					}

					// Get any Infomation of the track
					info, ok := c.RelayManager.GetInfo(subscription.TrackNamespace, subscription.TrackName)
					if ok {
						c.HandleSubscribe(update, &info, sw)
					} else {
						c.HandleSubscribe(update, nil, sw)
					}

					<-sw.doneCh

					slog.Info("updated a subscription", slog.Any("from", subscription), slog.Any("to", update))

					/*
					 * Update the subscription
					 */
					sess.updateSubscription(update)
					subscription = update
				}

				sess.stopSubscription(subscription.subscribeID)

				// Close the Stream gracefully
				err = stream.Close()
				if err != nil {
					slog.Debug("failed to close the stream", slog.String("error", err.Error()))
					return
				}
			case FETCH:
				slog.Debug("Fetch Stream was opened")

				fetchRequest, err := getFetchRequest(qvr)
				if err != nil {
					slog.Error("failed to get a fetch-request", slog.String("error", err.Error()))
					return
				}

				w := FetchResponceWriter{
					doneCh: make(chan struct{}),
					stream: stream,
				}

				c.HandleFetch(fetchRequest, w)

				<-w.doneCh
			case INFO:
				slog.Info("Info Stream is opened")

				infoRequest, err := getInfoRequest(qvr)
				if err != nil {
					slog.Error("failed to get a info-request", slog.String("error", err.Error()))
					return
				}

				w := InfoWriter{
					doneCh: make(chan struct{}),
					stream: stream,
				}

				info, ok := c.RelayManager.GetInfo(infoRequest.TrackNamespace, infoRequest.TrackName)
				if ok {
					c.HandleInfoRequest(infoRequest, &info, w)
				} else {
					c.HandleInfoRequest(infoRequest, nil, w)
				}

				<-w.doneCh
			default:
				err := ErrInvalidStreamType
				slog.Error(err.Error(), slog.Uint64("ID", uint64(num)))

				// Cancel reading and writing
				stream.CancelRead(err.StreamErrorCode())
				stream.CancelWrite(err.StreamErrorCode())

				return
			}
		}(stream)
	}
}

func (c Client) listenUniStreams(sess *Session) {
	for {
		stream, err := sess.Connection.AcceptUniStream(context.Background())
		if err != nil {
			slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
			return
		}

		go func(stream ReceiveStream) {
			/*
			 * Get a group
			 */
			group, err := getGroup(quicvarint.NewReader(stream))
			if err != nil {
				slog.Error("failed to get a group", slog.String("error", err.Error()))
				return
			}

			/*
			 * Find a subscription corresponding to the Subscribe ID in the Group
			 * Verify if subscribed or not
			 */
			sess.rsMu.RLock()
			defer sess.rsMu.RUnlock()

			_, ok := sess.subscribeWriters[group.SubscribeID]
			if !ok {
				slog.Error("received data of unsubscribed track", slog.Any("group", group))
				return
			}

			c.HandleData(group, stream)
		}(stream)
	}
}
