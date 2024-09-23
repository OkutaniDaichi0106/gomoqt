package rawquic

// import (
// 	"context"
// 	"crypto/tls"
// 	"net"

// 	"github.com/quic-go/quic-go"
// )

// type Server struct {
// 	QUICConfig quic.Config
// 	Port       int

// 	//quic.Connection
// }

// func (s Server) listenAndServe(tlsConfig *tls.Config, quicConfig *quic.Config) error {
// 	var err error

// 	// Create UDP connection
// 	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: s.Port})
// 	if err != nil {
// 		return err
// 	}

// 	// Create listener with UDP connection, TLS configuration and QUIC configuration
// 	transport := quic.Transport{Conn: udpConn}

// 	ln, err := transport.ListenEarly(tlsConfig, quicConfig)
// 	if err != nil {
// 		return err
// 	}

// 	for {
// 		conn, err := ln.Accept(context.TODO()) // TODO:
// 		if err != nil {
// 			return err
// 		}

// 		go func(conn quic.Connection) {
// 			stream, err := conn.AcceptStream(context.Background())
// 			if err != nil {
// 				return
// 			}

// 		}(conn)
// 	}
// }

// // QUIC Listener including quic.Listener and quic.EarlyListener
// // type QUICListener interface {
// // 	Accept(ctx context.Context) (quic.Connection, error)
// // 	Addr() net.Addr
// // 	Close() error
// // }
