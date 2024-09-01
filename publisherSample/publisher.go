package main

import (
	"crypto/tls"
	"go-moq/gomoq"
	"log"
)

func main() {
	// Set client
	client := gomoq.Client{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Versions: []gomoq.Version{gomoq.Draft05},
	}

	// Set subscriber
	publisher := gomoq.Publisher{
		Client:         client,
		TrackNamespace: "localhost/daichi/",
	}

	err := publisher.Connect("https://localhost:8443/setup")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Negotiated!!")

	err = publisher.Announce("audio")
	log.Print(err)
}
