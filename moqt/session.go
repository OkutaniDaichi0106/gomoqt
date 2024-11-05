package moqt

import "log/slog"

type Session struct {
	Connection    Connection
	SessionStream SessionStream
	//version Version
	/*
	 * Announcement
	 */
	announcement Announcement
	/*
	 * Subscription
	 */
	subscription map[SubscribeID]Subscription
}

func (sess Session) Terminate(terr TerminateError) {
	err := sess.Connection.CloseWithError(SessionErrorCode(terr.TerminateErrorCode()), terr.Error())
	if err != nil {
		slog.Error("failed to close the Session", slog.String("error", err.Error()))
		return
	}

	slog.Info("closed the Session", slog.String("reason", terr.Error()))
}
