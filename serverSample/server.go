package main

import (
	"context"
	"go-moq/gomoq"
	"log"
	"net/http"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

func main() {
	ws := webtransport.Server{
		H3: http3.Server{
			Addr: "0.0.0.0:8443",
			//TLSConfig: tlsConfig, //TODO: set appropriate tls config
		},
		//CheckOrigin: func(r *http.Request) bool {},
	}

	ms := gomoq.Server{
		WebTransportServer: &ws,
		SupportedVersions:  []gomoq.Version{gomoq.Draft05},
	}

	gomoq.OnPublisher(&ms, func(agent gomoq.PublisherAgent) {
		err := gomoq.AcceptAnnounce(agent)
		if err != nil {
			return
		}

		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		err = gomoq.AcceptObjects(agent, ctx)
		if err != nil {
			log.Fatal(err)
			return
		}
	})

	gomoq.OnSubscriber(&ms, func(agent gomoq.SubscriberAgent) {
		// Send ANNOUNCE messages to Subscribers and let them know available Track Namespace
		err := gomoq.Advertise(agent, ms.Announcements())
		if err != nil {
			return
		}

		err = gomoq.AcceptSubscription(agent)
		if err != nil {
			return
		}

		gomoq.DeliverObjects(agent)
	})
	http.HandleFunc("/setup", func(w http.ResponseWriter, r *http.Request) {
		// Establish WebTransport connection after receive EXTEND CONNECT message
		sess, err := ws.Upgrade(w, r)
		if err != nil {
			log.Printf("upgrading failed: %s", err)
			w.WriteHeader(500)
			return
		}

		agent, err := gomoq.Setup(sess)
		if err != nil {
			log.Fatal(err)
			return
		}
		//		agent := ms.NewAgent(sess)

		gomoq.Activate(agent)

		// Exchange SETUP messages

		// When the Client is a Subscriber

		// When the Client is a Publisher

		//Handle the session. Here goes the application logic
		//sess.CloseWithError(1234, "stop connection!!!")
	})

	ms.ListenAndServeTLS("localhost.pem", "localhost-key.pem")
}
