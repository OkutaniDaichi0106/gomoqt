package main

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/quic-go/quic-go"
)

func main() {
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

	moqs.SetCertFiles("localhost.pem", "localhost-key.pem")

	relayer := moqt.Relayer{
		Path:           "/main",
		RequestHandler: handler{},
	}

	moqs.RunOnQUIC(relayer)
}

var _ moqt.RequestHandler = (*handler)(nil)

type handler struct{}

func (handler) HandleInterest(i moqt.Interest, as []moqt.Announcement, w moqt.AnnounceWriter) {
	if as == nil {
		// Handle
	}

	for _, announcement := range as {
		w.Announce(announcement)
	}

	w.Close(nil)
}

func (handler) HandleSubscribe(s moqt.Subscription, info *moqt.Info, w moqt.SubscribeResponceWriter) {
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

func (handler) HandleFetch(r moqt.FetchRequest, w moqt.FetchResponceWriter) {
	w.Reject(moqt.ErrNoGroup)
}

func (handler) HandleInfoRequest(r moqt.InfoRequest, i *moqt.Info, w moqt.InfoWriter) {
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
