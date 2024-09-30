package main

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtversion"

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
		Versions: []moqtversion.Version{moqtversion.LATEST},
		WTConfig: struct {
			ReorderingTimeout time.Duration
			CheckOrigin       func(r *http.Request) bool
			EnableDatagrams   bool
		}{
			EnableDatagrams: true,
		},
	}

	moqtransport.HandlePublishingFunc("/", func(ps *moqtransport.PublishingSession) {
		ps.WaitSubscribe()
	})

	moqtransport.HandlePublishingFunc("/", func(ps *moqtransport.PublishingSession) {

	})

	// c := make(chan os.Signal)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// go func() {
	// 	<-c
	// 	// Exit server
	// 	cancel()
	// 	ms.Close()
	// }()

	ms.ListenAndServe()

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
