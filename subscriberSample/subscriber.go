package main

import (
	"context"
	"crypto/tls"
	"go-moq/gomoq"
	"log"
)

const (
	URL = "https://localhost:8443/"
)

func main() {
	// Set client
	client := gomoq.Client{
		ClientHandler: ClientHandle{},
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Versions: []gomoq.Version{gomoq.Draft05},
	}

	// Set subscriber
	subscriber := gomoq.Subscriber{
		Client:            client,
		SubscriberHandler: SubscriberHandle{},
	}

	err := subscriber.ConnectAndSetup(URL + "setup")
	if err != nil {
		log.Fatal(err)
	}

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

	//ctx, _ := context.WithCancel(context.Background()) // TODO: use cancel function
	dataCh, errCh := subscriber.AcceptObjects(context.TODO())

	for _, chunk := range <-dataCh {
		log.Println(chunk)
	}

	log.Fatal(<-errCh)
	// TODO: handle error
	// cancel()

}

type ClientHandle struct {
}

func (ClientHandle) ClientSetupParameters() gomoq.Parameters {
	return gomoq.Parameters{}
}

type SubscriberHandle struct {
}

func (SubscriberHandle) SubscribeParameters() gomoq.Parameters {
	return gomoq.Parameters{}
}
func (SubscriberHandle) SubscribeUpdateParameters() gomoq.Parameters {
	return gomoq.Parameters{}
}
