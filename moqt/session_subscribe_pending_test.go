package moqt

import (
	"bytes"
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/qumo-dev/gomoqt/moqt/internal/message"
	"github.com/qumo-dev/gomoqt/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// subscribeOkBytes is a SUBSCRIBE_OK resolving the start at group.
func subscribeOkBytes(tb testing.TB, group uint64) []byte {
	tb.Helper()
	var buf bytes.Buffer
	_, _ = buf.Write([]byte{byte(message.MessageTypeSubscribeOk)})
	require.NoError(tb, message.SubscribeOkMessage{Group: group}.Encode(&buf))
	return buf.Bytes()
}

type subscribeResult struct {
	reader *TrackReader
	err    error
}

// A publisher of a live track may withhold SUBSCRIBE_OK until it has a group.
// The subscription stays pending for as long as the caller waits, and is
// answered when the publisher speaks.
func TestSession_Subscribe_PendingUntilAnswered(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		answer := make(chan struct{})
		stream := &FakeQUICStream{
			Reads:    []streamResult{{Block: true}, {Data: subscribeOkBytes(t, 7)}, {Block: true}},
			ReadGate: answer,
		}
		conn := &FakeStreamConn{OpenStreams: []biStreamResult{{Stream: stream}}}
		session := newTestSession(conn)
		defer session.CloseWithError(NoError, "") //nolint:errcheck // test teardown

		done := make(chan subscribeResult, 1)
		go func() {
			reader, err := session.Subscribe(context.Background(), "/live/quiet", "events", nil)
			done <- subscribeResult{reader, err}
		}()

		time.Sleep(10 * time.Minute)
		synctest.Wait()
		select {
		case got := <-done:
			t.Fatalf("Subscribe returned while the publisher had not answered: %v", got.err)
		default:
		}
		assert.Empty(t, stream.CancelReadCodes(), "the pending subscribe stream is left alone")

		close(answer)

		got := <-done
		require.NoError(t, got.err)
		require.NotNil(t, got.reader)
	})
}

// The caller's context is the only limit on the wait: when it ends, the
// subscribe stream is cancelled and Subscribe reports why.
func TestSession_Subscribe_PendingEndsWithContext(t *testing.T) {
	tests := map[string]struct {
		newContext func() (context.Context, context.CancelFunc)
		wantErr    error
	}{
		"deadline": {
			newContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 2*time.Second)
			},
			wantErr: context.DeadlineExceeded,
		},
		"cancel": {
			newContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				time.AfterFunc(2*time.Second, cancel)
				return ctx, cancel
			},
			wantErr: context.Canceled,
		},
	}
	// t.Run is not supported inside a synctest bubble, so each case gets a
	// bubble of its own.
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				stream := &FakeQUICStream{Reads: []streamResult{{Block: true}}}
				conn := &FakeStreamConn{OpenStreams: []biStreamResult{{Stream: stream}}}
				session := newTestSession(conn)
				defer session.CloseWithError(NoError, "") //nolint:errcheck // test teardown
				ctx, cancel := tt.newContext()
				defer cancel()

				reader, err := session.Subscribe(ctx, "/live/quiet", "events", nil)

				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, reader)
				codes := stream.CancelReadCodes()
				require.NotEmpty(t, codes, "the subscribe stream is cancelled")
				assert.Equal(t, transport.StreamErrorCode(SubscribeErrorCodeTimeout), codes[0])
			})
		})
	}
}
