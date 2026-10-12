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

func TestWebTransportQUICConfig(t *testing.T) {
	tests := map[string]struct {
		given *quic.Config
		want  *quic.Config
	}{
		"nil is an empty configuration with both on": {
			given: nil,
			want:  &quic.Config{EnableDatagrams: true, EnableStreamResetPartialDelivery: true},
		},
		"the caller's settings are kept": {
			given: &quic.Config{MaxIdleTimeout: time.Minute, InitialPacketSize: 1200},
			want: &quic.Config{
				MaxIdleTimeout: time.Minute, InitialPacketSize: 1200,
				EnableDatagrams: true, EnableStreamResetPartialDelivery: true,
			},
		},
		"already on": {
			given: &quic.Config{EnableDatagrams: true, EnableStreamResetPartialDelivery: true},
			want:  &quic.Config{EnableDatagrams: true, EnableStreamResetPartialDelivery: true},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var before *quic.Config
			if tt.given != nil {
				before = tt.given.Clone()
			}

			got := webTransportQUICConfig(tt.given)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, before, tt.given, "the caller's configuration is left alone")
			assert.NotSame(t, tt.given, got, "a copy, even when nothing had to change")
		})
	}
}

func TestQUICDialConfig(t *testing.T) {
	tests := map[string]struct {
		given    *quic.Config
		want     *quic.Config
		wantSame bool // the very configuration given comes back
	}{
		"nil":                     {given: nil, want: nil, wantSame: true},
		"a positive keep-alive":   {given: &quic.Config{KeepAlivePeriod: time.Second}, want: &quic.Config{KeepAlivePeriod: time.Second}, wantSame: true},
		"zero":                    {given: &quic.Config{}, want: &quic.Config{}, wantSame: true},
		"negative becomes zero":   {given: &quic.Config{KeepAlivePeriod: -1, MaxIdleTimeout: time.Minute}, want: &quic.Config{MaxIdleTimeout: time.Minute}},
		"the smallest negative":   {given: &quic.Config{KeepAlivePeriod: -time.Nanosecond}, want: &quic.Config{}},
		"negative, others intact": {given: &quic.Config{KeepAlivePeriod: -time.Hour, InitialPacketSize: 1200}, want: &quic.Config{InitialPacketSize: 1200}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var before *quic.Config
			if tt.given != nil {
				before = tt.given.Clone()
			}

			got := quicDialConfig(tt.given)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, before, tt.given, "the caller's configuration is left alone")
			if tt.wantSame {
				assert.Same(t, tt.given, got)
			}
		})
	}
}

// TestDial_WebTransport_QUICConfig dials a real WebTransport server and checks
// that the dialer's QUIC configuration is the one the connection is made
// with: its Tracer, which quic-go asks for once per connection, is called
// once for the dial that opens the session.
func TestDial_WebTransport_QUICConfig(t *testing.T) {
	tests := map[string]struct {
		given *quic.Config // nil dials with no configuration at all
	}{
		"with what WebTransport needs": {
			given: &quic.Config{MaxIdleTimeout: time.Minute, EnableDatagrams: true, EnableStreamResetPartialDelivery: true},
		},
		// A configuration that sets neither still dials: both are turned on
		// for the connection, as a server's are.
		"without what WebTransport needs": {
			given: &quic.Config{MaxIdleTimeout: time.Minute},
		},
		// What the default dial turns a negative keep-alive into must still
		// be a configuration the library takes.
		"a negative keep-alive": {
			given: &quic.Config{KeepAlivePeriod: -1},
		},
		"no configuration": {
			given: nil,
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
			var before *quic.Config
			if tt.given != nil {
				before = tt.given.Clone()
				tt.given.Tracer = func(_ context.Context, isClient bool, _ quic.ConnectionID) qlogwriter.Trace {
					if isClient {
						asClient.Add(1)
					}
					return nil
				}
			}
			dialer := &Dialer{TLSConfig: &tls.Config{InsecureSkipVerify: true}, QUICConfig: tt.given}

			// Dialed from the test's own goroutine until the server is up, so
			// no attempt outlives the test, and counted per attempt, so the
			// count is that of the dial that opened the session.
			var sess *Session
			deadline := time.Now().Add(5 * time.Second)
			for {
				asClient.Store(0)
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				var err error
				sess, err = dialer.Dial(ctx, "https://"+addr+"/", nil)
				cancel()
				if err == nil {
					break
				}
				require.True(t, time.Now().Before(deadline), "the session never opened: %v", err)
				time.Sleep(50 * time.Millisecond)
			}
			t.Cleanup(func() { _ = sess.CloseWithError(NoError, "") })

			if tt.given == nil {
				return
			}
			assert.Equal(t, int32(1), asClient.Load(), "one connection, made with the dialer's QUIC configuration")
			tt.given.Tracer = nil // a func, which Equal cannot compare
			assert.Equal(t, before, tt.given, "the dialer's configuration is left alone")
		})
	}
}
