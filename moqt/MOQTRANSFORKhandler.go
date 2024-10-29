package moqt

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

type ServerHandler interface {
	SessionHandler
	SetupHandler
	PublisherHandler
	SubscriberHandler
}

/*
 * MOQTransfork
 */

var _ ServerHandler = (*defaultServerHandler)(nil)

func NewHandler() SessionHandler {
	return defaultServerHandler{}
}

type defaultServerHandler struct {
	SetupHandler
	PublisherHandler
	SubscriberHandler
}

func (handler defaultServerHandler) HandleSession(sess Session) {
	// Read the first byte and get Stream Type
	buf := make([]byte, 1)
	_, err := sess.stream.Read(buf)
	if err != nil {
		slog.Error("failed to read a Stream Type", slog.String("error", err.Error()))
		return
	}
	// Verify if the Stream Type is the SESSION
	if protocol.StreamType(buf[0]) != protocol.SESSION {
		slog.Error("unexpected Stream Type ID", slog.Uint64("ID", uint64(buf[0]))) // TODO
		return
	}

	// Get a set-up request
	req, err := getSetupRequest(quicvarint.NewReader(sess.stream))
	if err != nil {
		slog.Error("failed to get a set-up request", slog.String("error", err.Error()))
		return
	}

	// Get a responce writer
	w := defaultSetupResponceWriter{
		once:   new(sync.Once),
		stream: sess.stream,
		errCh:  make(chan error),
	}

	// Handle the request
	handler.SetupHandler.HandleSetup(req, w)

	// Catch any error
	err = <-w.errCh
	if err != nil {
		slog.Error("failed to set up", slog.String("error", err.Error()))
		return
	}

}

func getSetupRequest(r quicvarint.Reader) (SetupRequest, error) {
	// Receive SESSION_CLIENT message
	var scm message.SessionClientMessage
	err := scm.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read a SESSION_CLIENT message", slog.String("error", err.Error())) // TODO
		return SetupRequest{}, err
	}

	// Get a path
	path, ok := getPath(scm.Parameters)
	if !ok {
		err := errors.New("path not found")
		slog.Error("path not found")
		return SetupRequest{}, err
	}

	req := SetupRequest{
		Path:       path,
		Parameters: Parameters(scm.Parameters),
	}

	for _, v := range scm.SupportedVersions {
		req.SupportedVersions = append(req.SupportedVersions, Version(v))
	}

	return req, nil
}
