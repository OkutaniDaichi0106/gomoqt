package rawquic

import (
	"context"
	"crypto/tls"
	"log"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/quic-go/quic-go"
)

var defaultServeMux map[string]func(c quic.Connection)

func HandleFunc(pattern string, op func(c quic.Connection)) {
	defaultServeMux[pattern] = op
}

func RawQUICServer(moqs moqtransport.Server) Server {
	return Server{
		Addr:       moqs.Addr,
		TLSConfig:  moqs.TLSConfig,
		QUICConfig: moqs.QUICConfig,
	}
}

type Server struct {
	Addr       string
	TLSConfig  *tls.Config
	QUICConfig *quic.Config
}

func (s Server) ListenAndServe() error {
	ln, err := quic.ListenAddrEarly(s.Addr, s.TLSConfig, s.QUICConfig)
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := ln.Accept(context.Background()) // TODO:
			if err != nil {
				log.Println(err)
				return
			}

			go func(conn quic.Connection) {

			}(conn)
		}
	}()

	return nil
}
