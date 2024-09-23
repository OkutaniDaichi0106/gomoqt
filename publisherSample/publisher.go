package main

import (
	"go-moq/moqtransport"
	"log"
	"time"
)

const (
	URL = "https://localhost:8443/"
)

func main() {
	// Set subscriber
	publisher := moqtransport.Publisher{
		MaxSubscribeID: 1 << 4,
	}

	sess, err := publisher.ConnectAndSetup(URL + "setup")
	if err != nil {
		log.Fatal(err)
	}

	sess.Announce()

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
