package main

import (
	"go-moq/moqtransport"
	"go-moq/moqtransport/moqtmessage"
	"log"
)

const (
	URL = "https://localhost:8443/"
)

func main() {
	// Set subscriber
	publisher := moqtransport.Publisher{
		MaxSubscribeID: 1 << 4,
	}

	sess, err := publisher.ConnectAndSetup(URL)
	if err != nil {
		log.Println(err)
		return
	}

	err = sess.Announce(moqtmessage.NewTrackNamespace("localhost", "daichi"), moqtransport.AnnounceConfig{})
	if err != nil {
		log.Println(err)
		return
	}

	subscription, err := sess.WaitSubscribe()
	if err != nil {
		log.Println(err)
		return
	}

	err = sess.AllowSubscribe(subscription, 0)
	if err != nil {
		log.Println(err)
		return
	}

	stream, err := publisher.NewTrack(*subscription, moqtmessage.TRACK, 0)
	if err != nil {
		log.Println(err)
		return
	}

	data := []byte("hello")

	stream.Write(data)
}
