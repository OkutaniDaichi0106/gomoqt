package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
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

	annConfig := moqtransport.AnnounceConfig{}

	moqtransport.HandlePublishingFunc("/", func(ps *moqtransport.PublishingSession) {
		err := ps.Announce(moqtmessage.NewTrackNamespace("relay"), annConfig)
		if err != nil {
			log.Println(err)
			return
		}

		subscription, err := ps.WaitSubscribe()
		if err != nil {
			log.Println(err)
			return
		}

		if subscription.GetTrackName() != "audio" {
			ps.RejectSubscribe(subscription, moqtransport.ErrTrackDoesNotExist)
			return
		}

		// TODO: Handle the received subscription

		err = ps.AllowSubscribe(subscription, 0)
		if err != nil {
			log.Println(err)
			return
		}

	})

	moqtransport.HandleSubscribingFunc("/", func(ss *moqtransport.SubscribingSession) {
		ann, err := ss.WaitAnnounce()
		if err != nil {
			log.Println(err)
			return
		}
		// TODO: Handle the announcement
		log.Println(ann)

		err = ss.AllowAnnounce(*ann)
		if err != nil {
			log.Println(err)
			return
		}

		_, err = ss.Subscribe(*ann, "audio", moqtransport.SubscribeConfig{})
		if err != nil {
			log.Println(err)
			return
		}

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
