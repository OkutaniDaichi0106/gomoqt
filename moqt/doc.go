// Package moqt implements the protocol handling for Media over QUIC.
// This follows the MOQ Lite specification (draft-lcurley-moq-lite).
//
// # Clients
//
// Dial, DialWebTransport, DialQUIC create a new MOQ client session.
//
/*
	client := &moqt.Client{}

	sess, err := client.Dial(ctx, "https://example.com:4433", mux)
	...
	session, err := client.DialWebTransport(ctx, "example.com:4433", "/path", mux)
	...
	sess, err := client.DialQUIC(ctx, "example.com:4433", "/path", mux)
*/
//
// # Servers
//
// Server listens for incoming connections and handles subscriptions.
// ListenAndServe starts the server and routes incoming connections to a setup handler. The setup handler is usually nil, which means to use DefaultRouter.
// Handle and HandleFunc register handlers to DefaultRouter for a URL paths.
//
/*
	moqt.Handle("/foo", fooHandler)

	moqt.HandleFunc("/bar", func(w moqt.SetupResponseWriter, r *moqt.SetupRequest) {
	    // handle setup request
	})

	err := moqt.ListenAndServe(":4433", tlsConfig)
	if err != nil {
	    log.Fatal(err)
	}
*/
//
// moqt.Accept accepts a new MOQ server session from a setup request within a setup handler.
//
/*
	moqt.HandleFunc("/broadcast", func(w moqt.SetupResponseWriter, r *moqt.SetupRequest) {
	    sess, err := moqt.Accept(w, r, nil)
		if err != nil {
		    w.Reject(moqt.ProtocolViolationErrorCode)
		    return
		}
		// handle session
	})
*/
//
// More control over the server can be achieved by creating a custom moqt.Server.
//
/*
	server := &moqt.Server{
	    Addr:       	":4433",
		TLSConfig:  	tlsConfig,
		QUICConfig: 	quicConfig,
		SetupHandler: 	myRouter,
		Logger:     	myLogger,
	}
	err := server.ListenAndServe()
	if err != nil {
	    log.Fatal(err)
	}
*/
//
// # Underlying Transport
//
// MOQ supports both QUIC and WebTransport.
// The package abstracts the underlying transports into the `quic` and `webtransport` subpackages.
// By default it uses github.com/quic-go/quic-go for QUIC and github.com/quic-go/webtransport for WebTransport.
// Custom transports can be provided by setting the client's `DialAddrFunc` and `DialWebTransportFunc`,
// and the server's `ListenFunc` and `NewWebtransportServerFunc`.
//
// Client example:
/*
	client := &moqt.Client{
	    DialAddrFunc: myDialQUICFunc,
	    DialWebTransportFunc: myDialWebTransportFunc,
	}
*/
//
// Server example:
//
/*
	server := &moqt.Server{
	    ListenFunc: myListenQUICFunc,
	    NewWebtransportServerFunc: myNewWebTransportServerFunc,
	}
*/
package moqt
