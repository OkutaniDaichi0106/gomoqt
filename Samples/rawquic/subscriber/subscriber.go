package main

import (
	"context"
	"crypto/tls"
	"log"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/quic-go/quic-go"
)

const (
	URL = "https://localhost:8443/"
)

func main() {
	// Set subscriber
	subscriber := moqtransport.Subscriber{}

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
	morqSess, err := subscriber.SetupMORQ(qconn, "/rawquic")
	if err != nil {
		log.Println(err)
		return
	}

	HandleSession(morqSess)

}

func HandleSession(mowtSess *moqtransport.Session) {}
