package moqt

import (
	"errors"
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type Interest struct {
	TrackPrefix string
	Parameters  Parameters
}

type InterestHandler interface {
	HandleInterest(Interest, AnnounceSender)
}

const (
	ENDED  AnnounceStatus = AnnounceStatus(message.ENDED)
	ACTIVE AnnounceStatus = AnnounceStatus(message.ACTIVE)
	LIVE   AnnounceStatus = AnnounceStatus(message.LIVE)
)

type AnnounceStatus message.AnnounceStatus

type Announcement struct {
	// /***/
	// status AnnounceStatus
	/***/
	TrackPath string
	/***/
	AuthorizationInfo string
	Parameters        Parameters
}

func readAnnouncement(r io.Reader) (Announcement, error) {
	slog.Debug("reading an announcement")
	// Read an ANNOUNCE message
	var am message.AnnounceMessage
	err := am.Decode(r)
	if err != nil {
		slog.Error("failed to read an ANNOUNCE message", slog.String("error", err.Error()))
		return Announcement{}, err
	}

	// Initialize an Announcement
	announcement := Announcement{
		TrackPath:  am.TrackPathSuffix,
		Parameters: Parameters(am.Parameters),
	}

	//
	authInfo, ok := getAuthorizationInfo(announcement.Parameters)
	if ok {
		announcement.AuthorizationInfo = authInfo
	}

	return announcement, nil
}

func listenAnnounceReceiver(ar *AnnounceReceiver) {
	go func() {
		for {
			announcement, err := readAnnouncement(ar.stream)
			if err != nil {
				slog.Error("failed to read an announcement", slog.String("error", err.Error()))
				return
			}

			func() {
				ar.mu.Lock()
				defer ar.mu.Unlock()

				switch announcement.status {
				case ENDED:
					_, ok := ar.announcementsMap[announcement.TrackPath]
					if !ok {
						// TODO: Protocol Error
						ar.CancelInterest(ErrProtocolViolation)
						return
					}
					delete(ar.announcementsMap, announcement.TrackPath)
				case ACTIVE:
					_, ok := ar.announcementsMap[announcement.TrackPath]
					if ok {
						// TODO: Protocol Error
						ar.CancelInterest(ErrProtocolViolation)
						return
					}

					ar.announcementsMap[announcement.TrackPath] = announcement
				case LIVE:
					ar.liveCh <- struct{}{}
				}
			}()
		}
	}()
}

/*
 *
 */
type AnnounceSender struct {
	/*
	 * Received interest
	 */
	interest Interest

	/*
	 *
	 */
	stream moq.Stream
}

func (as *AnnounceSender) Announce(announcement Announcement) {

	announcement.status = ACTIVE
	//
	err := writeAnnouncement(as.stream, announcement)
	if err != nil {
		slog.Error("failed to write an announcement", slog.String("error", err.Error()))
		return
	}
}

func (as *AnnounceSender) Unannounce(announcement Announcement) {
	//
	announcement.status = ENDED
	//
	err := writeAnnouncement(as.stream, announcement)
	if err != nil {
		slog.Error("failed to write an announcement", slog.String("error", err.Error()))
		return
	}

	//
}

func (as *AnnounceSender) CancelAnnounce(err error) {
	if err == nil {
		as.Close()
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

	as.stream.CancelRead(code)
	as.stream.CancelWrite(code)

	slog.Info("closed an Announce Stream")
}

func (as *AnnounceSender) Close() {
	err := as.stream.Close()
	if err != nil {
		slog.Error("catch an erro when closing an Announce Stream", slog.String("error", err.Error()))
	}
}

func writeAnnouncement(w io.Writer, announcement Announcement) error {
	slog.Debug("writing an announcement")

	// Add AUTHORIZATION_INFO parameter
	if announcement.AuthorizationInfo != "" {
		announcement.Parameters.Add(AUTHORIZATION_INFO, announcement.AuthorizationInfo)
	}

	// Initialize an ANNOUNCE message
	am := message.AnnounceMessage{
		TrackPathSuffix: announcement.TrackPath,
		Parameters:      message.Parameters(announcement.Parameters),
	}

	// Encode the ANNOUNCE message
	err := am.Encode(w)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE message.", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Successfully announced", slog.Any("announcement", announcement))

	return nil
}
