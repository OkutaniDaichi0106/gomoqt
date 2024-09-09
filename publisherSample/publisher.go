package main

import (
	"crypto/tls"
	"go-moq/moqtransport"
	"log"
	"time"
)

const (
	URL = "https://localhost:8443/"
)

func main() {
	// Set client
	client := moqtransport.Client{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Versions:      []moqtransport.Version{moqtransport.Draft05},
		ClientHandler: ClientHandle{},
	}

	// Set subscriber
	publisher := moqtransport.Publisher{
		Client:           client,
		PublisherHandler: PublisherHandle{},
		TrackNamespace:   "localhost/daichi/",
	}

	params, err := publisher.ConnectAndSetup(URL + "setup")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(params)

	err = publisher.Announce(URL + "daichi0106")
	if err != nil {
		log.Fatal(err)
	}

	data := []byte("hello world")

	for i := 0; i < 10; i++ {
		errCh := publisher.SendSingleObject(0, data)
		err = <-errCh
		if err != nil {
			log.Fatal(err)
			return
		}
		time.Sleep(5 * time.Second)
	}
}

type PublisherHandle struct{}

func (PublisherHandle) AnnounceParameters() moqtransport.Parameters {
	return moqtransport.Parameters{}
}

type ClientHandle struct{}

func (ClientHandle) ClientSetupParameters() moqtransport.Parameters {
	return moqtransport.Parameters{}
}
