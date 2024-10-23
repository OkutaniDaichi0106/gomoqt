package main

import (
	"crypto/tls"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/OkutaniDaichi0106/gomoqt/moqtransfork"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

func main() {
	wts := webtransport.Server{
		H3: http3.Server{
			Addr: "0.0.0.0:8443",
		},
	}

	moqs := moqt.Server{
		TLSConfig: &tls.Config{}, // Use your tls.Config here
		QUICConfig: &quic.Config{ // Use your quic.Config here
			Allow0RTT:       true,
			EnableDatagrams: true,
		},
		SupportedVersions: []moqt.Version{0xffffff01},
	}

	moqs.HandleFunc("/webtransport", moqtransfork.Handler)

	// http.HandleFunc("/webtransport", func(w http.ResponseWriter, r *http.Request) {
	// 	wtSess, err := wts.Upgrade(w, r)
	// 	if err != nil {
	// 		log.Printf("upgrading failed: %s", err)
	// 		w.WriteHeader(500)
	// 		return
	// 	}

	// 	conn := mowebtransport.NewConnection(wtSess)

	// 	handler := handler{}

	// 	moqs.Run(conn)

	// })

	moqs.ListenAndServeWT(&wts)
}

type handler struct {
	moqt.SetupHandler
	moqt.AnnounceHandler
	moqt.SubscribeHandler
}
