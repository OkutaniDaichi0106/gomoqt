package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/qlogwriter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDial_WebTransport_QUICConfig dials a real WebTransport server and checks
// that the dialer's QUIC configuration is the one the connection is made
// with: its Tracer, which quic-go asks for once per connection, is called.
func TestDial_WebTransport_QUICConfig(t *testing.T) {
	tests := map[string]struct {
		// config builds the dialer's configuration around tracer.
		config func(tracer func(context.Context, bool, quic.ConnectionID) qlogwriter.Trace) *quic.Config
	}{
		"with what WebTransport needs": {
			config: func(tracer func(context.Context, bool, quic.ConnectionID) qlogwriter.Trace) *quic.Config {
				return &quic.Config{Tracer: tracer, EnableDatagrams: true, EnableStreamResetPartialDelivery: true}
			},
		},
		// A configuration that sets neither still dials: both are turned on
		// for the connection, as a server's are.
		"without what WebTransport needs": {
			config: func(tracer func(context.Context, bool, quic.ConnectionID) qlogwriter.Trace) *quic.Config {
				return &quic.Config{Tracer: tracer}
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			addr := freePort(t)
			srv := &Server{
				Addr: addr,
				TLSConfig: &tls.Config{
					NextProtos:   []string{NextProtoH3, NextProtoMOQ},
					Certificates: []tls.Certificate{generateTestCert(t)},
				},
				Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				Handler: HandleFunc(func(sess *Session) {
					<-sess.Context().Done()
				}),
			}
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, ErrServerClosed) {
					t.Logf("server: %v", err)
				}
			}()
			t.Cleanup(func() { assert.NoError(t, srv.Close()) })

			var asClient atomic.Int32
			given := tt.config(func(_ context.Context, isClient bool, _ quic.ConnectionID) qlogwriter.Trace {
				if isClient {
					asClient.Add(1)
				}
				return nil
			})
			wantDatagrams, wantPartialDelivery := given.EnableDatagrams, given.EnableStreamResetPartialDelivery
			dialer := &Dialer{TLSConfig: &tls.Config{InsecureSkipVerify: true}, QUICConfig: given}

			var sess *Session
			require.Eventually(t, func() bool {
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancel()
				var err error
				sess, err = dialer.Dial(ctx, "https://"+addr+"/", nil)
				return err == nil
			}, 5*time.Second, 50*time.Millisecond, "the session never opened")
			t.Cleanup(func() { _ = sess.CloseWithError(NoError, "") })

			assert.Positive(t, asClient.Load(), "the connection was made with the dialer's QUIC configuration")
			assert.Equal(t, wantDatagrams, given.EnableDatagrams, "the dialer's configuration is left alone")
			assert.Equal(t, wantPartialDelivery, given.EnableStreamResetPartialDelivery, "the dialer's configuration is left alone")
		})
	}
}
