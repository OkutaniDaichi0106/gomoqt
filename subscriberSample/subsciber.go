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
	subscriber := gomoq.Subscriber{
		Client: client,
	}

	err := subscriber.Connect("https://localhost:8443/setup")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Connected!!")

	subscriptionConfig := gomoq.SubscribeConfig{
		GroupOrder: gomoq.NOT_SPECIFY,
		SubscriptionFilter: gomoq.SubscriptionFilter{
			FilterCode: gomoq.LATEST_OBJECT,
		},
	}
	err = subscriber.Subscribe("localhost/daichi/", "audio", subscriptionConfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Subscribed!!")
}
