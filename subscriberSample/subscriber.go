package main

import (
	"log"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
)

const (
	URL = "https://localhost:8443/"
)

func main() {
	// Set subscriber
	subscriber := moqtransport.Subscriber{}

	sess, err := subscriber.ConnectAndSetup(URL)
	if err != nil {
		log.Println(err)
		return
	}

	announcement, err := sess.WaitAnnounce()
	if err != nil {
		log.Println(err)
		return
	}

	err = sess.AllowAnnounce(*announcement)
	if err != nil {
		log.Println(err)
		return
	}

	stream, err := sess.Subscribe(*announcement, "audio", moqtransport.SubscribeConfig{})
	if err != nil {
		log.Println(err)
		return
	}

	buf := make([]byte, 1<<8)

	n, err := stream.Read(buf)

	if err != nil {
		log.Println(err)
		return
	}

	log.Println("data: ", buf[:n])
}
