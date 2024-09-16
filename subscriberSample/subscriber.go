package main

import (
	"context"
	"crypto/tls"
	"go-moq/moqtransport"
	"go-moq/moqtransport/moqtversion"
	"io"
	"log"
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
	subscriber := moqtransport.Subscriber{
		Client: client,
	}

	_, err := subscriber.ConnectAndSetup(URL + "setup")
	if err != nil {
		log.Fatal(err)
	}

	err = subscriber.Subscribe("localhost/daichi/", "audio", nil)
	if err != nil {
		log.Fatal(err)
	}

	//ctx, _ := context.WithCancel(context.Background()) // TODO: use cancel function
	stream, _ := subscriber.AcceptObjects(context.TODO())
	log.Println(stream.Header())

	buf := make([]byte, 1<<8)

	for {
		n, err := stream.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		log.Println(buf[:n])
	}

}
