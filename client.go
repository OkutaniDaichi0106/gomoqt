package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
	"github.com/OkutaniDaichi0106/gomoqt/internal/protocol"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

type Client struct {
	TLSConfig *tls.Config

	QUICConfig *quic.Config

	supportedVersions []Version

	// CacheManager CacheManager
}

func (c Client) Dial(urlstr string, ctx context.Context) (sess clientSession, rsp SetupResponce, err error) {
	//
	req, err := NewSetupRequest(urlstr, nil)
	if err != nil {
		slog.Error("failed to get a set-up request", slog.String("error", err.Error()))
		return
	}

	return c.DialWithRequest(req, ctx)
}

func (c Client) DialWithRequest(req SetupRequest, ctx context.Context) (sess clientSession, rsp SetupResponce, err error) {

	/*
	 * Connect on QUIC or WebTransport
	 */
	var conn moq.Connection
	switch req.parsedURL.Scheme {
	case "https":
		// Dial on webtransport
		var wtsess *webtransport.Session
		var d webtransport.Dialer
		_, wtsess, err = d.Dial(ctx, req.urlstr, http.Header{}) // TODO: configure the header
		if err != nil {
			slog.Error("failed to dial with webtransport", slog.String("error", err.Error()))
			return
		}

		conn = moq.NewMOWTConnection(wtsess)
	case "moqt":
		/*
		 * Dial on raw quic
		 */
		var ips []net.IP
		ips, err = net.LookupIP(req.parsedURL.Hostname())
		if err != nil {
			slog.Error("failed to look up IP address", slog.String("error", err.Error()))
			return
		}

		// Try all IPs
		for i, ip := range ips {
			// Get Address
			addr := ip.String()
			if strings.Contains(addr, ":") && !strings.HasPrefix(addr, "[") {
				addr = "[" + addr + "]"
			}
			addr += ":" + req.parsedURL.Port()

			// Dial
			var qconn quic.Connection
			qconn, err = quic.DialAddrEarly(ctx, addr, c.TLSConfig, c.QUICConfig)
			if err != nil {
				slog.Error("failed to dial with quic", slog.String("error", err.Error()))
				if i+1 >= len(ips) {
					err = errors.New("no more IPs")
					slog.Error("failed to dial to the host",
						slog.String("error", err.Error()),
						slog.String("host", req.parsedURL.Hostname()),
					)
					return
				}
				continue
			}

			conn = moq.NewMORQConnection(qconn)

			break
		}

	default:
		err = errors.New("invalid scheme")
		slog.Error("url scheme must be https or moqt", slog.String("scheme", req.parsedURL.Scheme))
		return
	}

	/*
	 * Open a Session Stream
	 */
	stream, err := openSessionStream(conn)
	if err != nil {
		slog.Error("failed to open a Session Stream")
		return
	}

	sess = &ClientSession{
		session: session{
			conn:              conn,
			stream:            stream,
			publisherManager:  newPublisherManager(),
			subscriberManager: newSubscriberManager(),
		},
	}

	//

	// Handle the request
	switch req.parsedURL.Path {
	case "https":
	case "moqt":
		req.Parameters.Add(AUTHORIZATION_INFO, req.parsedURL.Path)
	default:
		err = errors.New("unsupported request scheme")
		return
	}

	/*
	 * Set up
	 */
	// Send a set-up request
	err = sendSetupRequest(stream, req)
	if err != nil {
		slog.Error("failed to request to set up", slog.String("error", err.Error()))
		return
	}

	// Receive a set-up responce
	rsp, err = readSetupResponce(stream)
	if err != nil {
		slog.Error("failed to receive a SESSION_SERVER message", slog.String("error", err.Error()))
		return
	}

	// Verify the selceted version is contained in the supported versions
	if !ContainVersion(rsp.SelectedVersion, req.supportedVersions) {
		err = errors.New("unexpected version was seleted")
		slog.Error("failed to negotiate versions", slog.String("error", err.Error()), slog.Any("selected version", rsp.SelectedVersion))
		return
	}

	return sess, rsp, nil
}

func openSessionStream(conn moq.Connection) (SessionStream, error) {
	slog.Debug("opening a session stream")

	/***/
	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_session,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func sendSetupRequest(w io.Writer, req SetupRequest) error {
	scm := message.SessionClientMessage{
		SupportedVersions: make([]protocol.Version, 0),
		Parameters:        message.Parameters(req.Parameters),
	}

	for _, v := range req.supportedVersions {
		scm.SupportedVersions = append(scm.SupportedVersions, protocol.Version(v))
	}

	err := scm.Encode(w)
	if err != nil {
		slog.Error("failed to send a SESSION_CLIENT message", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (c Client) listenBiStreams(sess ClientSession, ctx context.Context) {
	for {
		stream, err := sess.conn.AcceptStream(ctx)
		if err != nil {
			slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
			return
		}

		slog.Debug("some control stream was opened")

		go func(stream moq.Stream) {
			/*
			 * Read a Stream Type
			 */
			var stm message.StreamTypeMessage
			err := stm.Decode(stream)
			if err != nil {
				slog.Error("failed to get a Stream Type ID", slog.String("error", err.Error()))
				return
			}

			switch stm.StreamType {
			case stream_type_announce:
				slog.Debug("announce stream was opened")

				interest, err := readInterest(stream)
				if err != nil {
					slog.Error("failed to get an Interest", slog.String("error", err.Error()))
					return
				}

				aw := AnnounceWriter{
					stream: stream,
				}
				// Announce
				for _, announcement := range c.announcements {
					// Verify if the Announcement's Track Namespace has the Track Prefix
					if strings.HasPrefix(announcement.TrackPath, interest.TrackPrefix) {
						// Announce the Track Namespace
						aw.Announce(announcement)
					}
				}
			case stream_type_subscribe:
				slog.Debug("subscribe stream was opened")

				//
				sr := SubscribeReceiver{
					stream: stream,
				}

				subscription, err := readSubscription(stream)
				if err != nil {
					slog.Error("failed to get a subscription", slog.String("error", err.Error()))
					//
					sr.CancelRead(err)
					return
				}

				// Verify if the Track Path in the subscription is valid
				_, ok := c.announcements[subscription.TrackPath]
				if !ok {
					sr.CancelRead(ErrTrackDoesNotExist)
					return
				}

				// Set the subscription to the receiver
				sr.subscription = subscription

				// Get any Infomation of the track
				info, _ := sess.getCurrentInfo(subscription.TrackPath)

				// Send the current track information
				sr.Inform(info)

				/*
				 * Accept the new subscription
				 */
				sess.acceptNewSubscription(&sr)

				/*
				 * Catch any Subscribe Update or any error from the subscriber
				 */
				for {
					update, err := sr.ReceiveUpdate()
					if err != nil {
						slog.Info("catched an error from the subscriber", slog.String("error", err.Error()))
						break
					}

					slog.Info("received a subscribe update request", slog.Any("subscription", update))

					sw := SubscribeReceiver{
						stream: stream,
					}

					// Get any Infomation of the track
					info, ok := sess.getCurrentInfo(subscription.TrackPath)
					if ok {
						sw.Inform(info)
					} else {
						sw.CancelRead(ErrTrackDoesNotExist)
						return
					}

					slog.Info("updated a subscription", slog.Any("from", subscription), slog.Any("to", update))

					/*
					 * Update the subscription
					 */
					sr.updateSubscription(update)
				}

				sess.deleteSubscription(subscription)

				// Close the Stream gracefully
				sr.Close()
				return
			case stream_type_fetch:
				slog.Debug("fetch stream was opened")

				frw := FetchResponceWriter{
					stream: stream,
				}

				req, err := readFetchRequest(stream)
				if err != nil {
					slog.Error("failed to get a fetch-request", slog.String("error", err.Error()))
					frw.Reject(err)
					return
				}

				// Get a data reader
				r, err := c.CacheManager.GetFrame(req.TrackPath, req.GroupSequence, req.FrameSequence)
				if err != nil {
					slog.Error("failed to get a frame", slog.String("error", err.Error()))
					frw.Reject(err)
					return
				}

				// Send the data if valid subscription exists
				for _, sr := range sess.subscribeReceivers {
					subscription := sr.subscription

					if subscription.TrackPath == req.TrackPath {
						// Send the group data
						w, err := frw.SendGroup(Group{
							subscribeID:       subscription.subscribeID,
							groupSequence:     req.GroupSequence,
							PublisherPriority: PublisherPriority(req.SubscriberPriority), // TODO: Handle Publisher Priority
						})
						if err != nil {
							slog.Error("failed to send a group", slog.String("error", err.Error()))
							frw.Reject(err)
							return
						}

						// Send the data by copying it from the reader
						io.Copy(w, r)

						// Break becase data
						break
					}
				}

				// Close the Fetch Stream gracefully
				frw.Reject(nil)
				return
			case stream_type_info:
				slog.Debug("info stream was opened")

				req, err := readInfoRequest(stream)
				if err != nil {
					slog.Error("failed to get a info-request", slog.String("error", err.Error()))
					return
				}

				// Initialize an Info Writer
				iw := InfoWriter{
					stream: stream,
				}

				info, ok := sess.getCurrentInfo(req.TrackPath)
				if ok {
					iw.Inform(info)
				} else {
					iw.CancelInform(ErrTrackDoesNotExist)
					return
				}

				// Close the Info Stream gracefully
				iw.CancelInform(nil)
				return
			default:
				slog.Debug("unknown stream was opend")

				// Terminate the session
				sess.Terminate(ErrInvalidStreamType)

				return
			}
		}(stream)
	}
}
