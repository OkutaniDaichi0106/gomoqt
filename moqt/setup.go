package moqt

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

type SessionStream Stream

type SetupRequest struct {
	Path              string
	SupportedVersions []Version
	Parameters        Parameters
}

/*
 *
 */
type SetupRequestWriter interface {
	//New(Stream) SetupRequestWriter
	Setup([]Version) error
	WithExtension(Parameters) SetupRequestWriter
}

var _ SetupRequestWriter = (*defaultSetupRequestWriter)(nil)

type defaultSetupRequestWriter struct {
	once   *sync.Once
	stream SessionStream
	params Parameters
}

// func (w defaultSetupRequestWriter) New(stream Stream) SetupRequestWriter {
// 	return defaultSetupRequestWriter{
// 		once:   new(sync.Once),
// 		stream: stream,
// 		params: make(Parameters),
// 	}
// }

func (w defaultSetupRequestWriter) Setup(versions []Version) error {
	/***/
	var scm message.SessionClientMessage
	for _, v := range versions {
		scm.SupportedVersions = append(scm.SupportedVersions, protocol.Version(v))
	}

	if w.params != nil {
		scm.Parameters = message.Parameters(w.params)
	}

	_, err := w.stream.Write(scm.SerializePayload())
	if err != nil {
		slog.Error("failed to send a SESSION_CLIENT message", slog.String("error", err.Error()))
		return err
	}

	/***/
	var ssm message.SessionServerMessage
	err = ssm.DeserializePayload(quicvarint.NewReader(w.stream))
	if err != nil {
		slog.Error("failed to receive a SESSION_SERVER message", slog.String("error", err.Error()))
		return err
	}

	if !ContainVersion(Version(ssm.SelectedVersion), versions) {
		err = errors.New("unexpected version was seleted")
		slog.Error("failed to negotiate versions", slog.String("error", err.Error()), slog.Uint64("selected version", uint64(ssm.SelectedVersion)))
		return err
	}

	// TODO: Handle the parameters

	return nil
}

func (w defaultSetupRequestWriter) WithExtension(params Parameters) SetupRequestWriter {
	return defaultSetupRequestWriter{
		stream: w.stream,
		params: params,
	}
}

/*
 *
 */
type SetupResponceWriter interface {
	Accept(Version)
	Reject(TerminateError)
	WithExtension(Parameters) SetupResponceWriter
}

var _ SetupResponceWriter = (*defaultSetupResponceWriter)(nil)

type defaultSetupResponceWriter struct {
	errCh  chan error
	once   *sync.Once
	stream SessionStream
	params Parameters
}

func (w defaultSetupResponceWriter) Accept(version Version) {
	ssm := message.SessionServerMessage{
		SelectedVersion: protocol.Version(version),
		Parameters:      message.Parameters(w.params),
	}

	_, err := w.stream.Write(ssm.SerializePayload())
	if err != nil {
		slog.Error("failed to send a SESSION_SERVER message", slog.String("error", err.Error()))
		w.Reject(ErrInternalError)
	}

	w.errCh <- nil
}

func (w defaultSetupResponceWriter) Reject(err TerminateError) {
	slog.Error(err.Error(), slog.Any("Code", err.TerminateErrorCode()))

	/*
	 * Send the Error
	 */
	w.stream.CancelRead(StreamErrorCode(err.TerminateErrorCode()))
	w.stream.CancelWrite(StreamErrorCode(err.TerminateErrorCode()))

	w.errCh <- err
}

func (w defaultSetupResponceWriter) WithExtension(params Parameters) SetupResponceWriter {
	return defaultSetupResponceWriter{
		once:   w.once,
		stream: w.stream,
		params: params,
	}
}

type SetupHandler interface {
	HandleSetup(SetupRequest, SetupResponceWriter)
}
