package main

import (
	"crypto/tls"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"

	"github.com/quic-go/quic-go"
)

func main() {
	moqs := moqtransport.Server{
		TLSConfig:         &tls.Config{},
		QUICConfig:        &quic.Config{},
		SupportedVersions: []moqtransport.Version{moqtransport.FoalkDraft01},
	}

	moqs.ListenAndServeQUIC("0.0.0.0:8444", QUICHandler{}, nil, nil)
}

type QUICHandler struct{}

func (QUICHandler) HandlePath(path string) func(*moqtransport.Session) {
	switch path {
	case "/rawquic":
		return func(s *moqtransport.Session) {
			HandleSession(s)
		}
	default:
		return nil
	}
}

func HandleSession(*moqtransport.Session) {}
