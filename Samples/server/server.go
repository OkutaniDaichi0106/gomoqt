package main

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/quic-go/quic-go"
)

func main() {
	/*
	 * Initialize a Server
	 */
	moqs := moqt.Server{
		Addr: "0.0.0.0:8443",
		QUICConfig: &quic.Config{
			Allow0RTT:       true,
			EnableDatagrams: true,
		},
		SupportedVersions: []moqt.Version{moqt.Devlop},
		SetupHandler: moqt.SetupHandlerFunc(func(sr moqt.SetupRequest, srw moqt.SetupResponceWriter) {
			if !moqt.ContainVersion(moqt.Devlop, sr.SupportedVersions) {
				srw.Reject(moqt.ErrInternalError)
			}

			srw.Accept(moqt.Devlop)
		}),
	}

	// Set certification config
	moqs.SetCertFiles("localhost.pem", "localhost-key.pem")

	// Initialize a Relayer
	relayer := moqt.Relayer{
		Path:           "/path",
		RequestHandler: requestHandler{},
		SessionHandler: sessionHandler{},
		RelayManager:   nil,
	}

	// Run the Relayer on WebTransport
	moqs.RunOnWebTransport(relayer)

	moqs.ListenAndServe()
}

var _ moqt.SessionHandler = (*sessionHandler)(nil)

type sessionHandler struct{}

func (sessionHandler) HandleSession(sess *moqt.ServerSession) {
	/*
	 * Interest
	 */
	interest := moqt.Interest{
		TrackPrefix: "relayer",
	}
	slog.Info("interest", slog.Any("interest", interest))
	annstr, err := sess.Interest(interest)
	if err != nil {
		slog.Error("failed to interest", slog.String("error", err.Error()))
		return
	}

	/*
	 * Get Announcements
	 */
	ann, err := annstr.ReadAnnouncement()
	if err != nil {
		slog.Error("failed to read an announcement", slog.String("error", err.Error()))
		return
	}

	slog.Info("received an announcement", slog.Any("announcement", ann))

	/*
	 * Subscribe
	 */
	subscription := moqt.Subscription{
		TrackNamespace: ann.TrackNamespace,
		TrackName:      "audio",
	}
	_, info, err := sess.Subscribe(subscription)
	if err != nil {
		slog.Error("failed to subscribe", slog.String("error", err.Error()))
		return
	}

	slog.Info("successfully subscribed", slog.Any("subscription", subscription), slog.Any("info", info))

	//

}

var _ moqt.RequestHandler = (*requestHandler)(nil)

type requestHandler struct{}

func (requestHandler) HandleInterest(i moqt.Interest, as []moqt.Announcement, w moqt.AnnounceWriter) {
	if as == nil {
		// Handle
	}

	for _, announcement := range as {
		w.Announce(announcement)
	}

	w.Close(nil)
}

func (requestHandler) HandleSubscribe(s moqt.Subscription, info *moqt.Info, w moqt.SubscribeResponceWriter) {
	if info == nil {
		/*
		 * When info is nil, it means the subscribed track was not found.
		 * Reject the subscription or make a new subscrition to upstream.
		 */
		w.Reject(moqt.ErrTrackDoesNotExist)
		return
	}

	w.Accept(*info)
}

func (requestHandler) HandleFetch(r moqt.FetchRequest, w moqt.FetchResponceWriter) {
	w.Reject(moqt.ErrNoGroup)
}

func (requestHandler) HandleInfoRequest(r moqt.InfoRequest, i *moqt.Info, w moqt.InfoWriter) {
	if i == nil {
		/*
		 * When info is nil, it means the subscribed track was not found.
		 * Reject the request or make a new subscrition to upstream and request information of the track.
		 */
		w.Reject(moqt.ErrTrackDoesNotExist)
		return
	}

	w.Answer(*i)
}
