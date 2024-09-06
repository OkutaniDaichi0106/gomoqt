package main

import (
	"crypto/tls"
	"go-moq/gomoq"
	"log"
	"time"
)

const (
	URL = "https://localhost:8443/"
)

func main() {
	// Set client
	client := gomoq.Client{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Versions:      []gomoq.Version{gomoq.Draft05},
		ClientHandler: ClientHandle{},
	}

	// Set subscriber
	publisher := gomoq.Publisher{
		Client:           client,
		PublisherHandler: PublisherHandle{},
		TrackNamespace:   "localhost/daichi/",
	}

	err := publisher.ConnectAndSetup(URL + "setup")
	if err != nil {
		log.Fatal(err)
	}

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

func (PublisherHandle) AnnounceParameters() gomoq.Parameters {
	return gomoq.Parameters{}
}

type ClientHandle struct{}

func (ClientHandle) ClientSetupParameters() gomoq.Parameters {
	return gomoq.Parameters{}
}
