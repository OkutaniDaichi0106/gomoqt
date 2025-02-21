package moqt_test

import (
	"errors"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func TestErrorAs(t *testing.T) {
	cases := map[string]struct {
		err              error
		isAnnounceError  bool
		isSubscribeError bool
		isInfoError      bool
		isGroupError     bool
		isTerminateError bool
	}{
		"ErrInternalError": {
			err:              moqt.ErrInternalError,
			isAnnounceError:  true,
			isSubscribeError: true,
			isInfoError:      true,
			isGroupError:     true,
			isTerminateError: true,
		},
		"ErrUnauthorizedError": {
			err:              moqt.ErrUnauthorizedError,
			isSubscribeError: true,
			isTerminateError: true,
		},
		"ErrClosedTrack": {
			err:              moqt.ErrClosedTrack,
			isSubscribeError: true,
		},
		"ErrDuplicatedSubscribeID": {
			err:              moqt.ErrDuplicatedSubscribeID,
			isSubscribeError: true,
		},
		"ErrEndedTrack": {
			err: moqt.ErrEndedTrack,
		},
		"ErrDuplicatedTrack": {
			err:              moqt.ErrDuplicatedTrack,
			isSubscribeError: true,
		},
		"ErrGroupExpired": {
			err:          moqt.ErrGroupExpired,
			isGroupError: true,
		},
		"ErrGroupOutOfRange": {
			err:          moqt.ErrGroupOutOfRange,
			isGroupError: true,
		},
		"ErrGroupRejected": {
			err:          moqt.ErrGroupRejected,
			isGroupError: true},
		"ErrInvalidRange": {
			err:              moqt.ErrInvalidRange,
			isSubscribeError: true,
		},
		"ErrGroupClosed": {
			err:          moqt.ErrClosedGroup,
			isGroupError: true,
		},
		"NoErrTerminate": {
			err:              moqt.NoErrTerminate,
			isTerminateError: true},
		"ErrProtocolViolation": {
			err:              moqt.ErrProtocolViolation,
			isTerminateError: true,
		},
		"ErrTrackDoesNotExist": {
			err:              moqt.ErrTrackDoesNotExist,
			isSubscribeError: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var ae moqt.AnnounceError
			if got := errors.As(tc.err, &ae); got != tc.isAnnounceError {
				t.Errorf("AnnounceError: expected %v, got %v", tc.isAnnounceError, got)
			}

			var se moqt.SubscribeError
			if got := errors.As(tc.err, &se); got != tc.isSubscribeError {
				t.Errorf("SubscribeError: expected %v, got %v", tc.isSubscribeError, got)
			}

			var ie moqt.InfoError
			if got := errors.As(tc.err, &ie); got != tc.isInfoError {
				t.Errorf("InfoError: expected %v, got %v", tc.isInfoError, got)
			}

			var ge moqt.GroupError
			if got := errors.As(tc.err, &ge); got != tc.isGroupError {
				t.Errorf("GroupError: expected %v, got %v", tc.isGroupError, got)
			}

			var te moqt.TerminateError
			if got := errors.As(tc.err, &te); got != tc.isTerminateError {
				t.Errorf("TerminateError: expected %v, got %v", tc.isTerminateError, got)
			}
		})
	}
}
