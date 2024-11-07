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
		Path: "/main",
	}

	moqs.RunOnQUIC(relayer)
}

var _ moqt.PublisherHandler = (*PublisherHandler)(nil)
var _ moqt.SubscriberHandler = (*SubscriberHandler)(nil)

type PublisherHandler struct{}

func (PublisherHandler) HandleInterest(i moqt.Interest, w moqt.AnnounceWriter) {
	return
}

func (PublisherHandler) HandleSubscribe(s moqt.Subscription, info *moqt.Info, w moqt.SubscribeResponceWriter) {
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

func (PublisherHandler) HandleFetch(r moqt.FetchRequest, w moqt.FetchResponceWriter) {
	w.Reject(moqt.ErrNoGroup)
}

func (PublisherHandler) HandleInfoRequest(r moqt.InfoRequest, i *moqt.Info, w moqt.InfoWriter) {
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

type SubscriberHandler struct {
}

func (SubscriberHandler) HandleInfo(i moqt.Info)
