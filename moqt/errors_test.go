package moqt

import (
	"errors"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
)

// Test for standard errors
func TestStandardErrors(t *testing.T) {
	tests := map[string]struct {
		err    error
		expect string
	}{
		"invalid scheme": {
			err:    ErrInvalidScheme,
			expect: "moqt: invalid scheme",
		},
		"invalid range": {
			err:    ErrInvalidRange,
			expect: "moqt: invalid range",
		},
		"closed session": {
			err:    ErrClosedSession,
			expect: "moqt: closed session",
		},
		"server closed": {
			err:    ErrServerClosed,
			expect: "moqt: server closed",
		},
		"client closed": {
			err:    ErrClientClosed,
			expect: "moqt: client closed",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.err.Error())
		})
	}
}

// Test for AnnounceErrorCode String method
func TestAnnounceErrorCode_String(t *testing.T) {
	tests := map[string]struct {
		code   AnnounceErrorCode
		expect string
	}{
		"internal error code": {
			code:   InternalAnnounceErrorCode,
			expect: "moqt: internal error",
		},
		"duplicated announce error code": {
			code:   DuplicatedAnnounceErrorCode,
			expect: "moqt: duplicated broadcast path",
		},
		"uninterested error code": {
			code:   UninterestedErrorCode,
			expect: "moqt: uninterested",
		},
		"banned prefix error code": {
			code:   BannedPrefixErrorCode,
			expect: "moqt: unknown announce error", // Should return default case
		},
		"unknown code": {
			code:   AnnounceErrorCode(0xFF), // Some arbitrary value not defined
			expect: "moqt: unknown announce error",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.code.String())
		})
	}
}

// Test for SubscribeErrorCode String method
func TestSubscribeErrorCode_String(t *testing.T) {
	tests := map[string]struct {
		code   SubscribeErrorCode
		expect string
	}{
		"internal error code": {
			code:   InternalSubscribeErrorCode,
			expect: "moqt: internal error",
		},
		"invalid range error code": {
			code:   InvalidRangeErrorCode,
			expect: "moqt: invalid range",
		},
		"duplicate subscribe ID error code": {
			code:   DuplicateSubscribeIDErrorCode,
			expect: "moqt: duplicated id",
		},
		"track not found error code": {
			code:   TrackNotFoundErrorCode,
			expect: "moqt: track does not exist",
		},
		"unauthorized subscribe error code": {
			code:   UnauthorizedSubscribeErrorCode,
			expect: "moqt: unauthorized",
		},
		"subscribe timeout error code": {
			code:   SubscribeTimeoutErrorCode,
			expect: "moqt: timeout",
		},
		"unknown code": {
			code:   SubscribeErrorCode(0xFF), // Some arbitrary value not defined
			expect: "moqt: unknown subscribe error",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.code.String())
		})
	}
}

// Test for SessionErrorCode String method
func TestSessionErrorCode_String(t *testing.T) {
	tests := map[string]struct {
		code   SessionErrorCode
		expect string
	}{
		"no error": {
			code:   NoError,
			expect: "moqt: no error",
		},
		"internal session error code": {
			code:   InternalSessionErrorCode,
			expect: "moqt: internal error",
		},
		"unauthorized session error code": {
			code:   UnauthorizedSessionErrorCode,
			expect: "moqt: unauthorized",
		},
		"protocol violation error code": {
			code:   ProtocolViolationErrorCode,
			expect: "moqt: protocol violation",
		},
		"parameter length mismatch error code": {
			code:   ParameterLengthMismatchErrorCode,
			expect: "moqt: parameter length mismatch",
		},
		"too many subscribe error code": {
			code:   TooManySubscribeErrorCode,
			expect: "moqt: too many subscribes",
		},
		"go away timeout error code": {
			code:   GoAwayTimeoutErrorCode,
			expect: "moqt: goaway timeout",
		},
		"unsupported version error code": {
			code:   UnsupportedVersionErrorCode,
			expect: "moqt: unsupported version",
		},
		"unsupported stream error code": {
			code:   UnsupportedStreamErrorCode,
			expect: "moqt: unknown session error", // Not explicitly handled in the String method
		},
		"unknown code": {
			code:   SessionErrorCode(0xFF), // Some arbitrary value not defined
			expect: "moqt: unknown session error",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.code.String())
		})
	}
}

// Test for GroupErrorCode String method
func TestGroupErrorCode_String(t *testing.T) {
	tests := map[string]struct {
		code   GroupErrorCode
		expect string
	}{
		"internal group error code": {
			code:   InternalGroupErrorCode,
			expect: "moqt: internal error",
		},
		"out of range error code": {
			code:   OutOfRangeErrorCode,
			expect: "moqt: out of range",
		},
		"expired group error code": {
			code:   ExpiredGroupErrorCode,
			expect: "moqt: group expires",
		},
		"subscribe canceled error code": {
			code:   SubscribeCanceledErrorCode,
			expect: "moqt: subscribe canceled",
		},
		"publish aborted error code": {
			code:   PublishAbortedErrorCode,
			expect: "moqt: publish aborted",
		},
		"closed session group error code": {
			code:   ClosedSessionGroupErrorCode,
			expect: "moqt: session closed",
		},
		"invalid subscribe ID error code": {
			code:   InvalidSubscribeIDErrorCode,
			expect: "moqt: invalid subscribe id",
		},
		"unknown code": {
			code:   GroupErrorCode(0xFF), // Some arbitrary value not defined
			expect: "moqt: unknown group error",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.code.String())
		})
	}
}

// Test for AnnounceError
func TestAnnounceError(t *testing.T) {
	tests := map[string]struct {
		err            AnnounceError
		expectedString string
		expectedCode   AnnounceErrorCode
	}{
		"internal error": {
			err: AnnounceError{
				&quic.StreamError{
					ErrorCode: quic.StreamErrorCode(InternalAnnounceErrorCode),
				},
			},
			expectedString: "moqt: internal error",
			expectedCode:   InternalAnnounceErrorCode,
		},
		"duplicated broadcast path": {
			err: AnnounceError{
				&quic.StreamError{
					ErrorCode: quic.StreamErrorCode(DuplicatedAnnounceErrorCode),
				},
			},
			expectedString: "moqt: duplicated broadcast path",
			expectedCode:   DuplicatedAnnounceErrorCode,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expectedString, tt.err.Error())
			assert.Equal(t, tt.expectedCode, tt.err.AnnounceErrorCode())
		})
	}
}

// Test for SubscribeError
func TestSubscribeError(t *testing.T) {
	tests := map[string]struct {
		err            SubscribeError
		expectedString string
		expectedCode   SubscribeErrorCode
	}{
		"internal error": {
			err: SubscribeError{
				&quic.StreamError{
					ErrorCode: quic.StreamErrorCode(InternalSubscribeErrorCode),
				},
			},
			expectedString: "moqt: internal error",
			expectedCode:   InternalSubscribeErrorCode,
		},
		"invalid range": {
			err: SubscribeError{
				&quic.StreamError{
					ErrorCode: quic.StreamErrorCode(InvalidRangeErrorCode),
				},
			},
			expectedString: "moqt: invalid range",
			expectedCode:   InvalidRangeErrorCode,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expectedString, tt.err.Error())
			assert.Equal(t, tt.expectedCode, tt.err.SubscribeErrorCode())
		})
	}
}

// Test for SessionError
func TestSessionError(t *testing.T) {
	tests := map[string]struct {
		err            SessionError
		expectedString string
		expectedCode   SessionErrorCode
	}{
		"local internal error": {
			err: SessionError{
				&quic.ApplicationError{
					ErrorCode: quic.ApplicationErrorCode(InternalSessionErrorCode),
					Remote:    false,
				},
			},
			expectedString: "moqt: internal error (local)",
			expectedCode:   InternalSessionErrorCode,
		},
		"remote unauthorized": {
			err: SessionError{
				&quic.ApplicationError{
					ErrorCode: quic.ApplicationErrorCode(UnauthorizedSessionErrorCode),
					Remote:    true,
				},
			},
			expectedString: "moqt: unauthorized (remote)",
			expectedCode:   UnauthorizedSessionErrorCode,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expectedString, tt.err.Error())
			assert.Equal(t, tt.expectedCode, tt.err.SessionErrorCode())
		})
	}
}

// Test for GroupError
func TestGroupError(t *testing.T) {
	tests := map[string]struct {
		err            GroupError
		expectedString string
		expectedCode   GroupErrorCode
	}{
		"internal error": {
			err: GroupError{
				&quic.StreamError{
					ErrorCode: quic.StreamErrorCode(InternalGroupErrorCode),
				},
			},
			expectedString: "moqt: internal error",
			expectedCode:   InternalGroupErrorCode,
		},
		"out of range": {
			err: GroupError{
				&quic.StreamError{
					ErrorCode: quic.StreamErrorCode(OutOfRangeErrorCode),
				},
			},
			expectedString: "moqt: out of range",
			expectedCode:   OutOfRangeErrorCode,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expectedString, tt.err.Error())
			assert.Equal(t, tt.expectedCode, tt.err.GroupErrorCode())
		})
	}
}

// Test for InternalError
func TestInternalError(t *testing.T) {
	tests := map[string]struct {
		err                      InternalError
		expectedString           string
		expectedAnnounceCode     AnnounceErrorCode
		expectedSubscribeCode    SubscribeErrorCode
		expectedSessionCode      SessionErrorCode
		expectedGroupCode        GroupErrorCode
		shouldMatchInternalError bool
	}{
		"with reason": {
			err:                      InternalError{Reason: "test error"},
			expectedString:           "moqt: internal error: test error",
			expectedAnnounceCode:     InternalAnnounceErrorCode,
			expectedSubscribeCode:    InternalSubscribeErrorCode,
			expectedSessionCode:      InternalSessionErrorCode,
			expectedGroupCode:        InternalGroupErrorCode,
			shouldMatchInternalError: true,
		},
		"empty reason": {
			err:                      InternalError{Reason: ""},
			expectedString:           "moqt: internal error: ",
			expectedAnnounceCode:     InternalAnnounceErrorCode,
			expectedSubscribeCode:    InternalSubscribeErrorCode,
			expectedSessionCode:      InternalSessionErrorCode,
			expectedGroupCode:        InternalGroupErrorCode,
			shouldMatchInternalError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expectedString, tt.err.Error())
			assert.Equal(t, tt.expectedAnnounceCode, tt.err.AnnounceErrorCode())
			assert.Equal(t, tt.expectedSubscribeCode, tt.err.SubscribeErrorCode())
			assert.Equal(t, tt.expectedSessionCode, tt.err.SessionErrorCode())
			assert.Equal(t, tt.expectedGroupCode, tt.err.GroupErrorCode())

			// Test Is method
			var internalErr InternalError
			assert.Equal(t, tt.shouldMatchInternalError, errors.Is(tt.err, internalErr))
		})
	}
}

// Test for UnauthorizedError
func TestUnauthorizedError(t *testing.T) {
	tests := map[string]struct {
		err                        UnauthorizedError
		expectedString             string
		expectedSubscribeCode      SubscribeErrorCode
		expectedSessionCode        SessionErrorCode
		shouldMatchUnauthorizedErr bool
	}{
		"default": {
			err:                        UnauthorizedError{},
			expectedString:             "moqt: unauthorized",
			expectedSubscribeCode:      UnauthorizedSubscribeErrorCode,
			expectedSessionCode:        UnauthorizedSessionErrorCode,
			shouldMatchUnauthorizedErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expectedString, tt.err.Error())
			assert.Equal(t, tt.expectedSubscribeCode, tt.err.SubscribeErrorCode())
			assert.Equal(t, tt.expectedSessionCode, tt.err.SessionErrorCode())

			// Test Is method
			var unauthorizedErr UnauthorizedError
			assert.Equal(t, tt.shouldMatchUnauthorizedErr, errors.Is(tt.err, unauthorizedErr))
		})
	}
}

// Test for error compatibility with standard errors
func TestErrorCompatibility(t *testing.T) {
	// Test errors.Is compatibility
	t.Run("errors.Is with InternalError", func(t *testing.T) {
		err := InternalError{Reason: "test"}
		var internalErr InternalError
		assert.True(t, errors.Is(err, internalErr))
	})

	t.Run("errors.Is with UnauthorizedError", func(t *testing.T) {
		err := UnauthorizedError{}
		var unauthorizedErr UnauthorizedError
		assert.True(t, errors.Is(err, unauthorizedErr))
	})

	// Test errors.As compatibility
	t.Run("errors.As with InternalError", func(t *testing.T) {
		err := InternalError{Reason: "test"}
		var target InternalError
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, "test", target.Reason)
	})

	t.Run("errors.As with UnauthorizedError", func(t *testing.T) {
		err := UnauthorizedError{}
		var target UnauthorizedError
		assert.True(t, errors.As(err, &target))
	})
}
