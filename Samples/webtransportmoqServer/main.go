package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go"
)

func main() {

	tlsConfig, err := generateTLSConfig("", "")
	if err != nil {
		return
	}

	ms := moqtransport.Server{
		Addr:      "0.0.0.0",
		Port:      8443,
		TLSConfig: tlsConfig,
		QUICConfig: &quic.Config{
			Allow0RTT: true,
		},
		SupportedVersions: []moqtmessage.Version{moqtmessage.FoalkDraft01},
		WTConfig: struct {
			ReorderingTimeout time.Duration
			CheckOrigin       func(r *http.Request) bool
			EnableDatagrams   bool
		}{
			EnableDatagrams: true,
		},
	}

	wts := ms.WTServer()

	http.HandleFunc("/webtransport", func(w http.ResponseWriter, r *http.Request) {
		sess, err := wts.Upgrade(w, r)
		if err != nil {
			log.Printf("upgrading failed: %s", err)
			w.WriteHeader(500)
			return
		}

		conn := moqtransport.NewMOWTConnection(*r.URL, sess)

	})

	wts.ListenAndServe()

}

func generateTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	var err error
	certs := make([]tls.Certificate, 1)
	certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: certs,
	}, nil
}
