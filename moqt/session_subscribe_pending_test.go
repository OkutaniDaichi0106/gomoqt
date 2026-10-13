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
// The subscription stays pending however long that takes, its stream left
// alone, and ends with what the publisher does next: an answer, or a reset.
func TestSession_Subscribe_PendingUntilThePublisherActs(t *testing.T) {
	refusal := &transport.StreamError{
		ErrorCode: transport.StreamErrorCode(SubscribeErrorCodeNotFound),
		Remote:    true,
	}
	tests := map[string]struct {
		then    streamResult
		wantErr bool
	}{
		"answered": {then: streamResult{Data: subscribeOkBytes(t, 7)}},
		"refused":  {then: streamResult{Err: refusal}, wantErr: true},
	}
	// t.Run is not supported inside a synctest bubble, so each case gets a
	// bubble of its own.
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				acts := make(chan struct{})
				stream := &FakeQUICStream{
					Reads:    []streamResult{{Block: true}, tt.then, {Block: true}},
					ReadGate: acts,
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
					t.Fatalf("Subscribe returned while the publisher had not acted: %v", got.err)
				default:
				}
				assert.Empty(t, stream.CancelReadCodes(), "the pending subscribe stream is left alone")

				close(acts)

				got := <-done
				if tt.wantErr {
					var subErr *SubscribeError
					require.ErrorAs(t, got.err, &subErr)
					assert.Equal(t, SubscribeErrorCodeNotFound, subErr.SubscribeErrorCode())
					assert.Nil(t, got.reader)
					return
				}
				require.NoError(t, got.err)
				require.NotNil(t, got.reader)
			})
		})
	}
}
