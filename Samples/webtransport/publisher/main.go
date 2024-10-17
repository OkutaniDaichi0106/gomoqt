package main

import (
	"context"
	"log"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/quic-go/webtransport-go"
)

const (
	URL = "https://localhost:8443"
)

func main() {
	// Initialize a Subscriber
	publisher := moqtransport.Publisher{
		SupportedVersions: []moqtransport.Version{moqtransport.FoalkDraft01},
		MaxSubscribeID:    1 << 4,
	}

	/*
	 * Dial to a server
	 */
	wtd := webtransport.Dialer{}
	var headers http.Header
	_, wtconn, err := wtd.Dial(context.Background(), URL+"/webtransport", headers)
	if err != nil {
		log.Println(err)
		return
	}
	//
	mowtSess, err := publisher.SetupMOWT(wtconn)
	if err != nil {
		log.Println(err)
		return
	}

	HandleSession(mowtSess)
}

func HandleSession(mowtSess *moqtransport.Session) {}
