package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"

	"github.com/quic-go/quic-go"
)

func main() {

	tlsConfig, err := generateTLSConfig("cert.pem", "cert-key.pem")
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

	ln, err := quic.ListenAddrEarly(ms.Addr, ms.TLSConfig, ms.QUICConfig)
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		for {
			conn, err := ln.Accept(context.Background()) // TODO:
			if err != nil {
				log.Println(err)
				return
			}

			go func(conn quic.Connection) {
				morqSess, path, err := ms.SetupMORQ(conn)
				if err != nil {
					return
				}
				switch path {
				case "/rawquic":

				default:
					return
				}
			}(conn)
		}
	}()

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
