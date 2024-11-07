package moqt

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceStream struct {
	reader quicvarint.Reader
	stream Stream
}

func (a AnnounceStream) ReadAnnouncement() (Announcement, error) {
	if a.reader == nil {
		a.reader = quicvarint.NewReader(a.stream)
	}

	var am message.AnnounceMessage
	err := am.DeserializePayload(a.reader)
	if err != nil {
		slog.Debug("failed to get an announcement", slog.String("error", err.Error()))
		return Announcement{}, err
	}

	announcement := Announcement{
		TrackNamespace: strings.Join(am.TrackNamespace, "/"),
		Parameters:     am.Parameters,
	}

	if authInfo, ok := getAuthorizationInfo(am.Parameters); ok {
		announcement.AuthorizationInfo = authInfo
	}

	return announcement, nil
}

func (a AnnounceStream) Close(err error) {
	if err == nil {
		err = a.stream.Close()
		if err != nil {
			slog.Debug("failed to close the stream", slog.String("error", err.Error()))
			return
		}
	}

	annerr, ok := err.(AnnounceError)
	if !ok {
		annerr = ErrInternalError
	}

	slog.Debug("trying to close an Announce Stream", slog.String("reason", annerr.Error()))

	a.stream.CancelWrite(StreamErrorCode(annerr.AnnounceErrorCode()))
	a.stream.CancelRead(StreamErrorCode(annerr.AnnounceErrorCode()))
}

type Interest struct {
	TrackPrefix string
	Parameters  Parameters
}

type InterestHandler interface {
	HandleInterest(Interest, []Announcement, AnnounceWriter)
}

type Announcement struct {
	TrackNamespace    string
	AuthorizationInfo string
	Parameters        message.Parameters
}

type AnnounceWriter struct {
	doneCh chan struct{}
	stream Stream
}

func (w AnnounceWriter) Announce(announcement Announcement) {
	am := message.AnnounceMessage{
		TrackNamespace: strings.Split(announcement.TrackNamespace, "/"),
		Parameters:     announcement.Parameters,
	}

	if announcement.AuthorizationInfo != "" {
		am.Parameters.Add(AUTHORIZATION_INFO, announcement.AuthorizationInfo)
	}

	_, err := w.stream.Write(am.SerializePayload())
	if err != nil {
		slog.Debug("failed to send an ANNOUNCE message.", slog.String("error", err.Error()))
		return
	}

	slog.Info("announced", slog.Any("announcement", announcement))
}

func (w AnnounceWriter) Close(err error) {
	if err == nil {
		err = w.stream.Close()
		slog.Error("failed to close an Announce Stream", slog.String("error", err.Error()))
		return
	}

	var code StreamErrorCode

	var strerr StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		annerr, ok := err.(AnnounceError)
		if ok {
			code = StreamErrorCode(annerr.AnnounceErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	w.stream.CancelRead(code)
	w.stream.CancelWrite(code)

	w.doneCh <- struct{}{}

	close(w.doneCh)

	slog.Debug("close an Announce Stream")
}
