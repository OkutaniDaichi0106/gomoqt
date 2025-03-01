package main

import (
	"crypto/tls"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/quic-go/quic-go"
)

func main() {
	server := moqt.Server{
		Addr: "localhost:4444",
		TLSConfig: &tls.Config{
			NextProtos:         []string{"h3", "moq-00"},
			Certificates:       []tls.Certificate{generateCert()},
			InsecureSkipVerify: true, // TODO: Not recommended for production
		},
		QUICConfig: &quic.Config{
			Allow0RTT: true,
		},
	}

	err := server.ListenAndServe()
	if err != nil {
		slog.Error("failed to listen and serve", slog.String("error", err.Error()))
	}
}

func generateCert() tls.Certificate {
	cert, err := tls.LoadX509KeyPair("../../cert/localhost.pem", "../../cert/localhost-key.pem")
	if err != nil {
		panic(err)
	}

	return cert
}
