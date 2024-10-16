package main

import (
	"context"
	"crypto/tls"
	"log"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go"
)

const (
	URL = "https://localhost:8444/"
)

func main() {
	// Set subscriber
	publisher := moqtransport.Publisher{
		Client: moqtransport.Client{
			SupportedVersions: []moqtmessage.Version{moqtmessage.FoalkDraft01},
		},
		MaxSubscribeID: 1 << 4,
	}

	/*
	 *
	 */
	// Dial
	qconn, err := quic.DialAddrEarly(context.Background(), "0.0.0.0:8444", &tls.Config{}, &quic.Config{})
	if err != nil {
		log.Println(err)
		return
	}
	//
	morqSess, err := publisher.SetupMORQ(qconn, "/rawquic")
	if err != nil {
		log.Println(err)
		return
	}

	HandleSession(morqSess)
}

func HandleSession(morqSess *moqtransport.Session) {}
