package main

import (
	"crypto/tls"
	"go-moq/moqtransport"
	"go-moq/moqtransport/moqtversion"
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
		Versions: []moqtversion.Version{moqtversion.LATEST},
	}

	// Set subscriber
	publisher := moqtransport.Publisher{
		Client:         client,
		TrackNamespace: []string{"localhost/daichi/"},
		MaxSubscribeID: 1 << 4,
	}

	_, err := publisher.ConnectAndSetup(URL + "setup")
	if err != nil {
		log.Fatal(err)
	}

	err = publisher.Announce("localhost/daichi/")
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
