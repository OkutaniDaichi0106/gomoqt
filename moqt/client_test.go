package moqt

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/url"
	"testing"
	"time"

	"crypto/tls"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_InitOnce(t *testing.T) {
	c := &Client{}
	c.init()
	c.init() // Should not panic or re-init
	assert.NotNil(t, c.activeSess)
	assert.NotNil(t, c.doneChan)
}

func TestClient_TimeoutDefault(t *testing.T) {
	c := &Client{}
	assert.Equal(t, 5*time.Second, c.timeout())
}

func TestClient_TimeoutCustom(t *testing.T) {
	c := &Client{Config: &Config{SetupTimeout: 123 * time.Second}}
	assert.Equal(t, 123*time.Second, c.timeout())
}

func TestClient_AddRemoveSession(t *testing.T) {
	c := &Client{}
	c.init()
	sess := &Session{}
	c.addSession(sess)
	assert.Contains(t, c.activeSess, sess)
	c.removeSession(sess)
	assert.NotContains(t, c.activeSess, sess)
}

func TestClient_ShuttingDown(t *testing.T) {
	c := &Client{}
	assert.False(t, c.shuttingDown())
	c.inShutdown.Store(true)
	assert.True(t, c.shuttingDown())
}

func TestClient_Close(t *testing.T) {
	c := &Client{}
	c.init()
	sess := &Session{}
	c.addSession(sess)
	// Should terminate session and wait for doneChan
	ch := make(chan struct{})
	go func() {
		c.Close()
		close(ch)
	}()
	// Remove session to unblock
	c.removeSession(sess)
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("Close did not return in time")
	}
}

func TestClient_ShutdownContextCancel(t *testing.T) {
	c := &Client{}
	c.init()
	waitCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockConn := &MockQUICConnection{
		AcceptStreamFunc: func(ctx context.Context) (quic.Stream, error) {
			select {
			case <-waitCtx.Done():
				return nil, waitCtx.Err()
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
		AcceptUniStreamFunc: func(ctx context.Context) (quic.ReceiveStream, error) {
			select {
			case <-waitCtx.Done():
				return nil, waitCtx.Err()
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}, OpenStreamFunc: func() (quic.Stream, error) {
			<-waitCtx.Done() // Simulate waiting for connection context
			return nil, nil  // Mock OpenStream to return nil
		}, OpenUniStreamFunc: func() (quic.SendStream, error) {
			<-waitCtx.Done() // Simulate waiting for connection context
			return nil, nil  // Mock OpenUniStream to return nil
		},
	}
	mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	mockConn.On("Context").Return(context.Background())
	mockConn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}).Maybe()

	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (n int, err error) {
			<-waitCtx.Done() // Simulate waiting for session context
			return 0, nil    // Mock Read to return no data
		},
	}
	mockStream.On("CancelRead", mock.Anything)
	mockStream.On("CancelWrite", mock.Anything)

	sess := newSession(mockConn, internal.DefaultServerVersion, "/path", NewParameters(), NewParameters(), mockStream, nil, slog.Default())
	c.addSession(sess)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Simulate context cancellation
	err := c.Shutdown(ctx)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestClient_ShutdownNoSessions(t *testing.T) {
	c := &Client{}
	c.init()
	err := c.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestClient_Dial(t *testing.T) {
	tests := map[string]struct {
		urlStr   string
		shut     bool
		wantErr  bool
		wantType error
	}{
		"invalid scheme": {urlStr: "ftp://host", wantErr: true, wantType: ErrInvalidScheme},
		"parse error":    {urlStr: "://", wantErr: true},
		"shutting down":  {urlStr: "https://host", shut: true, wantErr: true, wantType: ErrClientClosed},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{}
			if tt.shut {
				c.inShutdown.Store(true)
			}
			_, err := c.Dial(context.Background(), tt.urlStr, nil)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantType != nil {
					assert.ErrorIs(t, err, tt.wantType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_DialWebTransport(t *testing.T) {
	tests := map[string]struct {
		uri     string
		shut    bool
		wtErr   error
		wantErr bool
	}{
		"webtransport error": {uri: "https://host", wtErr: errors.New("fail"), wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{}
			if tt.shut {
				c.inShutdown.Store(true)
			}
			uri, _ := url.Parse(tt.uri)
			old := c.DialWebTransportFunc
			c.DialWebTransportFunc = func(ctx context.Context, url string, h http.Header) (*http.Response, quic.Connection, error) {
				return nil, nil, tt.wtErr
			}
			defer func() { c.DialWebTransportFunc = old }()
			_, err := c.DialWebTransport(context.Background(), uri.Hostname()+":"+uri.Port(), uri.Path, nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_DialQUIC(t *testing.T) {
	tests := map[string]struct {
		uri     string
		shut    bool
		dialErr error
		wantErr bool
	}{
		"dial error": {uri: "moqt://host", dialErr: errors.New("fail"), wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{}
			if tt.shut {
				c.inShutdown.Store(true)
			}
			uri, _ := url.Parse(tt.uri)
			old := c.DialQUICConn
			c.DialQUICConn = func(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *quic.Config) (quic.Connection, error) {
				return nil, tt.dialErr
			}
			defer func() { c.DialQUICConn = old }()
			_, err := c.DialQUIC(context.Background(), uri.Hostname()+":"+uri.Port(), uri.Path, nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_openSession(t *testing.T) {
	tests := map[string]struct {
		mockConn  func() *MockQUICConnection
		extension func() *Parameters
		wantErr   bool
	}{
		"OpenStream error": {
			mockConn: func() *MockQUICConnection {
				conn := &MockQUICConnection{}
				conn.On("OpenStream").Return(nil, errors.New("openstream fail"))
				conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}).Maybe()
				return conn
			}, wantErr: true,
		},
		"STREAM_TYPE encode error": {
			mockConn: func() *MockQUICConnection {
				conn := &MockQUICConnection{}
				stream := &MockQUICStream{}
				stream.On("StreamID").Return(quic.StreamID(1))
				stream.On("Write", mock.Anything).Return(0, errors.New("stm encode fail"))
				conn.On("OpenStream").Return(stream, nil)
				conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}).Maybe()
				return conn
			},
			extension: func() *Parameters { return &Parameters{} },
			wantErr:   true,
		},
		"SESSION_CLIENT encode error": {
			mockConn: func() *MockQUICConnection {
				conn := &MockQUICConnection{}
				stream := &MockQUICStream{}
				stream.On("StreamID").Return(quic.StreamID(1))
				stream.On("Write", mock.Anything).Return(1, nil).Once()
				stream.On("Write", mock.Anything).Return(0, errors.New("scm encode fail")).Once()
				conn.On("OpenStream").Return(stream, nil)
				conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}).Maybe()
				return conn
			},
			extension: func() *Parameters { return &Parameters{} },
			wantErr:   true,
		},
		"SESSION_SERVER decode error": {
			mockConn: func() *MockQUICConnection {
				conn := &MockQUICConnection{}
				stream := &MockQUICStream{}
				stream.On("StreamID").Return(quic.StreamID(1))
				stream.On("Write", mock.Anything).Return(1, nil)
				stream.On("Write", mock.Anything).Return(1, nil)
				stream.On("Read", mock.Anything).Return(0, errors.New("ssm decode fail"))
				conn.On("OpenStream").Return(stream, nil)
				conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}).Maybe()
				return conn
			},
			extension: func() *Parameters { return &Parameters{} },
			wantErr:   true,
		},
		"path param error": {
			mockConn: func() *MockQUICConnection {
				conn := &MockQUICConnection{}
				stream := &MockQUICStream{}
				stream.On("StreamID").Return(quic.StreamID(1))
				stream.On("Write", mock.Anything).Return(1, nil)
				stream.On("Write", mock.Anything).Return(1, nil)
				stream.On("Read", mock.Anything).Return(1, nil)
				stream.On("CancelRead", mock.Anything).Return()
				conn.On("OpenStream").Return(stream, nil)
				conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}).Maybe()
				return conn
			},
			extension: func() *Parameters { return &Parameters{} },
			wantErr:   true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{}
			conn := tt.mockConn()
			_, err := c.openSession(conn, "/path", tt.extension, nil, slog.Default())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			conn.AssertExpectations(t)
		})
	}
}
