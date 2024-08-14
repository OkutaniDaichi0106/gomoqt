package main

import (
	"go-moq/gomoq"
	"log"
	"net/http"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

type StreamingHandle struct {
}

func (StreamingHandle) Serve() {}

func Serve() {
	//Setting server
	ws := webtransport.Server{
		H3: http3.Server{
			Addr: "0.0.0.0:28443",
			//TLSConfig: tlsConfig, //TODO: set appropriate tls config
		},
		//CheckOrigin: func(r *http.Request) bool {},
	}

	gomoqServer := gomoq.Server{
		WebtransportServer: &ws,
		AvailableVersions:  []gomoq.Version{gomoq.Draft05},
	}
	http.HandleFunc("/webtransport", func(w http.ResponseWriter, r *http.Request) {
		sess, err := ws.Upgrade(w, r)
		if err != nil {
			log.Printf("upgrading failed: %s", err)
			w.WriteHeader(500)
			return
		}

		err = gomoqServer.Setup(sess)
		if err != nil {
			return
		}

		//Handle the session. Here goes the application logic
		sess.CloseWithError(1234, "stop connection!!!")
	})

	http.HandleFunc("/webtransport", func(w http.ResponseWriter, r *http.Request) {
		sess, err := ws.Upgrade(w, r)
		if err != nil {
			log.Printf("upgrading failed: %s", err)
			w.WriteHeader(500)
			return
		}

		err = gomoqServer.Setup(sess)
		if err != nil {
			return
		}

		//Handle the session. Here goes the application logic
		sess.CloseWithError(1234, "stop connection!!!")
	})

	gomoqServer.ListenAndServeTLS("localhost.pem", "localhost-key.pem")

}
