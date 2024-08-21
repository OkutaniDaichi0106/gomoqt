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

	moqs := gomoq.Server{
		WebtransportServer: &ws,
		SupportedVersions:  []gomoq.Version{gomoq.Draft05},
	}
	// moqs := gomoq.Server{
	// 	WebtransportServer: &ws,
	// 	SupportedVersions:  []gomoq.Version{gomoq.Draft05},
	// }

	http.HandleFunc("/setup", func(w http.ResponseWriter, r *http.Request) {
		// Establish WebTransport connection after receive EXTEND CONNECT message
		sess, err := ws.Upgrade(w, r)
		if err != nil {
			log.Printf("upgrading failed: %s", err)
			w.WriteHeader(500)
			return
		}

		// Set the session to a agent
		moqa := gomoq.Agent{
			Session: sess,
			Server:  moqs,
		}

		// Exchange SETUP messages
		err = moqa.Setup()

		//Handle the session. Here goes the application logic
		//sess.CloseWithError(1234, "stop connection!!!")
	})

	ws.ListenAndServeTLS("localhost.pem", "localhost-key.pem")
}
