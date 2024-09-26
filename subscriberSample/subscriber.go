package main

import (
	"go-moq/moqtransport"
	"log"
)

const (
	URL = "https://localhost:8443/"
)

func main() {
	// Set subscriber
	subscriber := moqtransport.Subscriber{}

	sess, err := subscriber.ConnectAndSetup(URL)
	if err != nil {
		log.Fatal(err)
	}

	announcement, err := sess.WaitAnnounce()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = sess.AllowAnnounce(*announcement)
	if err != nil {
		log.Fatal(err)
		return
	}

	stream, err := sess.Subscribe(*announcement, "audio", moqtransport.SubscribeConfig{})
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 1<<8)

	n, err := stream.Read(buf)

	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("data: ", buf[:n])
}
