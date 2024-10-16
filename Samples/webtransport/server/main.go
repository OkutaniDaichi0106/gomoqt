package main

import (
	"crypto/tls"
	"log"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
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

	moqs := moqtransport.Server{
		TLSConfig: &tls.Config{}, // Use your tls.Config here
		QUICConfig: &quic.Config{ // Use your quic.Config here
			Allow0RTT:       true,
			EnableDatagrams: true,
		},
		SupportedVersions: []moqtmessage.Version{moqtmessage.FoalkDraft01},
	}

	http.HandleFunc("/webtransport", func(w http.ResponseWriter, r *http.Request) {
		wtSess, err := wts.Upgrade(w, r)
		if err != nil {
			log.Printf("upgrading failed: %s", err)
			w.WriteHeader(500)
			return
		}

		mowtSess, err := moqs.SetupMOWT(wtSess)
		if err != nil {
			return
		}

		// Handle the MOQT session
		HandleSession(mowtSess)
	})

	moqs.ListenAndServeWT(&wts)
}

func HandleSession(mowtSess *moqtransport.Session) {}
