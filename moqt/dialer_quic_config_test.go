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

func TestCheckQUICDialConfig(t *testing.T) {
	tests := map[string]struct {
		given   *quic.Config
		wantErr bool
	}{
		"nil":                   {given: nil},
		"zero":                  {given: &quic.Config{}},
		"a positive keep-alive": {given: &quic.Config{KeepAlivePeriod: time.Second}},
		"the smallest positive": {given: &quic.Config{KeepAlivePeriod: time.Nanosecond}},
		"the smallest negative": {given: &quic.Config{KeepAlivePeriod: -time.Nanosecond}, wantErr: true},
		"a negative keep-alive": {given: &quic.Config{KeepAlivePeriod: -time.Second}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := checkQUICDialConfig(tt.given)

			if tt.wantErr {
				assert.ErrorContains(t, err, "KeepAlivePeriod")
				return
			}
			assert.NoError(t, err)
		})
	}
}

// The default QUIC and WebTransport dials refuse a negative keep-alive before
// they dial; the WebSocket dial takes it, as "no pings".
func TestDialer_Dial_NegativeKeepAlive(t *testing.T) {
	tests := map[string]struct {
		target  string
		refused bool
	}{
		"native QUIC":  {target: "moqt://127.0.0.1:1/session", refused: true},
		"WebTransport": {target: "https://127.0.0.1:1/session", refused: true},
		"WebSocket":    {target: "wss://127.0.0.1:1/session"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			given := &quic.Config{KeepAlivePeriod: -1}
			dialer := &Dialer{
				TLSConfig:  &tls.Config{InsecureSkipVerify: true},
				Config:     &Config{SetupTimeout: 200 * time.Millisecond},
				QUICConfig: given,
			}

			_, err := dialer.Dial(context.Background(), tt.target, nil)

			// Nothing listens on the port, so the WebSocket dial fails too,
			// but for that reason and not for the keep-alive.
			require.Error(t, err)
			if tt.refused {
				assert.ErrorContains(t, err, "KeepAlivePeriod")
			} else {
				assert.NotContains(t, err.Error(), "KeepAlivePeriod")
			}
			assert.Equal(t, &quic.Config{KeepAlivePeriod: -1}, given, "the dialer's configuration is left alone")
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
