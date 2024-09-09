package main

import (
	"context"
	"crypto/tls"
	"go-moq/moqtransport"
	"io"
	"log"
)

const (
	URL = "https://localhost:8443/"
)

func main() {
	// Set client
	client := moqtransport.Client{
		ClientHandler: ClientHandle{},
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Versions: []moqtransport.Version{moqtransport.Draft05},
	}

	// Set subscriber
	subscriber := moqtransport.Subscriber{
		Client:            client,
		SubscriberHandler: SubscriberHandle{},
	}

	params, err := subscriber.ConnectAndSetup(URL + "setup")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(params)

	subscriptionConfig := moqtransport.SubscribeConfig{
		GroupOrder: moqtransport.NOT_SPECIFY,
		SubscriptionFilter: moqtransport.SubscriptionFilter{
			FilterCode: moqtransport.LATEST_OBJECT,
		},
	}
	err = subscriber.Subscribe("localhost/daichi/", "audio", subscriptionConfig)
	if err != nil {
		log.Fatal(err)
	}

	//ctx, _ := context.WithCancel(context.Background()) // TODO: use cancel function
	dataCh, err := subscriber.AcceptObjects(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	dataStream := <-dataCh
	buf := make([]byte, 1<<8)
	for {
		n, err := dataStream.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		log.Println(buf[:n])
	}

}

type ClientHandle struct {
}

func (ClientHandle) ClientSetupParameters() moqtransport.Parameters {
	return moqtransport.Parameters{}
}

type SubscriberHandle struct {
}

func (SubscriberHandle) SubscribeParameters() moqtransport.Parameters {
	return moqtransport.Parameters{}
}
func (SubscriberHandle) SubscribeUpdateParameters() moqtransport.Parameters {
	return moqtransport.Parameters{}
}
