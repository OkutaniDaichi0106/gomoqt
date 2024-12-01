package moqt

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type AnnounceStream struct {
	stream moq.Stream
}

func (a AnnounceStream) ReadAnnouncement() (Announcement, error) {
	return readAnnouncement(a.stream)
}

func (a AnnounceStream) Close(err error) {
	if err == nil {
		err = a.stream.Close()
		if err != nil {
			slog.Error("failed to close the stream", slog.String("error", err.Error()))
			return
		}
	}

	annerr, ok := err.(AnnounceError)
	if !ok {
		annerr = ErrInternalError
	}

	slog.Info("trying to close an Announce Stream", slog.String("reason", annerr.Error()))

	a.stream.CancelWrite(moq.StreamErrorCode(annerr.AnnounceErrorCode()))
	a.stream.CancelRead(moq.StreamErrorCode(annerr.AnnounceErrorCode()))
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
	Parameters        Parameters
}

type AnnounceWriter struct {
	doneCh chan struct{}
	stream moq.Stream
}

func (w AnnounceWriter) Announce(announcement Announcement) {
	// Add AUTHORIZATION_INFO parameter
	if announcement.AuthorizationInfo != "" {
		announcement.Parameters.Add(AUTHORIZATION_INFO, announcement.AuthorizationInfo)
	}

	// Initialize an ANNOUNCE message
	am := message.AnnounceMessage{
		TrackNamespace: strings.Split(announcement.TrackNamespace, "/"),
		Parameters:     message.Parameters(announcement.Parameters),
	}

	// Encode the ANNOUNCE message
	err := am.Encode(w.stream)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE message.", slog.String("error", err.Error()))
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

	var code moq.StreamErrorCode

	var strerr moq.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		annerr, ok := err.(AnnounceError)
		if ok {
			code = moq.StreamErrorCode(annerr.AnnounceErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	w.stream.CancelRead(code)
	w.stream.CancelWrite(code)

	w.doneCh <- struct{}{}

	close(w.doneCh)

	slog.Info("closed an Announce Stream")
}
