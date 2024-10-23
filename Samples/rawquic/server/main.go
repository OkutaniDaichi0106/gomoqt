package main

import (
	"crypto/tls"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/quic-go/quic-go"
)

func main() {
	moqs := moqt.Server{
		TLSConfig:         &tls.Config{},
		QUICConfig:        &quic.Config{},
		SupportedVersions: []moqt.Version{0xffffff01},
	}

	moqs.ListenAndServeQUIC("0.0.0.0:8444", QUICHandler{}, nil, nil)
}

type QUICHandler struct{}

func (QUICHandler) HandlePath(path string) moqt.Handler {
	switch path {
	case "/rawquic":
		return moqt.Handler{
			SetupHandler:     moqt.SetupHandlerFunc(func(srw moqt.SetupResponceWriter) {}),
			AnnounceHandler:  moqt.AnnounceHandlerFunc(func(a moqt.Announcement, arw moqt.AnnounceResponceWriter) {}),
			SubscribeHandler: moqt.SubscribeHandlerFunc(func(s moqt.Subscription, srw moqt.SubscribeResponceWriter) {}),
		}
	default:
		return nil
	}
}

func HandleSession(*moqtransport.Session) {}
