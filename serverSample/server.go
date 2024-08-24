package main

import (
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

	moqServer := gomoq.Server{
		WebtransportServer: &ws,
		SupportedVersions:  []gomoq.Version{gomoq.Draft05},
		TrackNames:         []string{"audio"},
	}

	http.HandleFunc("/setup", func(w http.ResponseWriter, r *http.Request) {
		// Establish WebTransport connection after receive EXTEND CONNECT message
		sess, err := ws.Upgrade(w, r)
		if err != nil {
			log.Printf("upgrading failed: %s", err)
			w.WriteHeader(500)
			return
		}

		// Set the session to a agent
		moqAgent := gomoq.Agent{
			Server:  moqServer,
			Session: sess,
		}

		// When the Agent is connected with a Publisher, perform the configured operation
		moqAgent.PublisherHandle(func() error {
			// Negotiate SETUP messages
			err = moqAgent.Setup()
			if err != nil {
				return err
			}

			// Advertise announcements
			err = moqAgent.Advertise()
			if err != nil {
				return err
			}

			// Accept SUBSCRIBE message
			err = moqAgent.AcceptSubscribe()
			if err != nil {
				return err
			}
			return nil
		})

		// When the Agent is connected with a Subscriber, perform the configured operation
		moqAgent.SubscriberHandle(func() error {
			err = moqAgent.Setup()
			if err != nil {
				return err
			}

			err = moqAgent.AcceptAnnounce()
			if err != nil {
				return err
			}
			return nil
		})

		// Exchange SETUP messages

		// When the Client is a Subscriber

		// When the Client is a Publisher

		//Handle the session. Here goes the application logic
		//sess.CloseWithError(1234, "stop connection!!!")
	})

	ws.ListenAndServeTLS("localhost.pem", "localhost-key.pem")
}
