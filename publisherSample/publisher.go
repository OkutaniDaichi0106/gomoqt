package main

import (
	"log"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
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
	log.Println("Announced!!")

	subscription, err := sess.WaitSubscribe()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Subscribed!!")

	err = sess.AllowSubscribe(subscription, 0)
	if err != nil {
		log.Println(err)
		return
	}

	header := moqtransport.NewStreamHeaderTrack(*subscription, 0)

	stream, err := publisher.OpenStreamTrack(header)
	if err != nil {
		log.Println(err)
		return
	}

	data := []byte("hello")

	stream.Write(data)
	stream.Write(data)
	stream.Write(data)
}
