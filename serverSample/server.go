package main

import (
	"go-moq/moqtransport"
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

	ms := moqtransport.Server{
		WebTransportServer: &ws,
	}

	http.HandleFunc("/setup", func(w http.ResponseWriter, r *http.Request) {
		versions := []moqtransport.Version{moqtransport.LATEST}

		// Establish WebTransport connection after receive EXTEND CONNECT message
		sess, err := ms.ConnectAndSetup(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		// Receive SETUP_CLIENT message
		params, err := sess.ReceiveClientSetup(versions)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(params)

		// Send  SETUP_SERVER message
		err = sess.SendServerSetup(moqtransport.Parameters{})
		if err != nil {
			log.Fatal(err)
			return
		}

		sess.OnPublisher(func(swp *moqtransport.SessionWithPublisher) {
			params, err := swp.ReceiveAnnounce()
			if err != nil {
				log.Println(err)
				err = swp.SendAnnounceError()
				log.Println(err)
				return
			}
			log.Println(params)

			err = swp.SendAnnounceOk()

		})

		sess.OnSubscriber(func(sws *moqtransport.SessionWithSubscriber) {
			err := sws.Advertise(moqtransport.Announcements())
			if err != nil {
				log.Println(err)
			}

			params, err := sws.ReceiveSubscribe()
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(params)

			err = sws.SendSubscribeResponce()
			if err != nil {
				log.Println(err)
				return
			}

			sws.DeliverObjects()
		})

		//		agent := ms.NewAgent(sess)

		//moqtransport.Activate(agent)

		// Exchange SETUP messages

		// When the Client is a Subscriber

		// When the Client is a Publisher

		//Handle the session. Here goes the application logic
		//sess.CloseWithError(1234, "stop connection!!!")
	})

	ms.ListenAndServeTLS("localhost.pem", "localhost-key.pem")
}
