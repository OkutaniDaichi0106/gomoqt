package moqt

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	authRoom1Path BroadcastPath = "/room/1/alice"
	authRoom2Path BroadcastPath = "/room/2/bob"
)

// roomTicketHandler grants each room's ticket the room's paths, refuses
// "bad", and grants the connection's own credential everything when the
// request URI carries "?admin" and nothing otherwise.
var roomTicketHandler = AuthHandlerFunc(func(w *AuthWriter, r *AuthRequest) {
	var g Grant
	switch string(r.Token) {
	case "":
		if u, err := url.Parse(r.RequestURI); err == nil && u.Query().Has("admin") {
			g.Subscribe = []PathPattern{MustParsePathPattern("/")}
		}
	case "room1":
		g.Subscribe = []PathPattern{MustParsePathPattern("/room/1/")}
	case "room2":
		g.Subscribe = []PathPattern{MustParsePathPattern("/room/2/")}
	case "publish":
		g.Publish = []PathPattern{MustParsePathPattern("/room/1/")}
	default:
		_ = w.Refuse(UnauthorizedSessionErrorCode, "unknown ticket")
		return
	}
	if err := w.Grant(g); err != nil {
		return
	}
	<-r.Context().Done()
})

// authGroupHandler opens a group every few milliseconds until the
// subscription ends.
func authGroupHandler(tw *TrackWriter) {
	ctx := tw.Context()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		gw, err := tw.OpenGroup(ctx)
		if err != nil {
			return
		}
		frame := NewFrame(1)
		_, _ = frame.Write([]byte{1})
		_ = gw.WriteFrame(frame)
		_ = gw.Close()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// authProbePath is the path startAuthServer dials to learn the server is up.
const authProbePath = "/probe"

// startAuthServer serves authRoom1Path and authRoom2Path behind handler. The
// handler on the server side receives every session it accepts.
func startAuthServer(t *testing.T, ctx context.Context, handler AuthHandler, sessions chan<- *Session) string {
	t.Helper()

	addr := freePort(t)
	mux := NewTrackMux(0)
	mux.PublishFunc(ctx, authRoom1Path, authGroupHandler)
	mux.PublishFunc(ctx, authRoom2Path, authGroupHandler)

	srv := &Server{
		Addr: addr,
		TLSConfig: &tls.Config{
			NextProtos:         []string{NextProtoMOQ},
			Certificates:       []tls.Certificate{generateTestCert(t)},
			InsecureSkipVerify: true,
		},
		QUICConfig:  &quic.Config{EnableDatagrams: true},
		TrackMux:    mux,
		AuthHandler: handler,
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		Handler: HandleFunc(func(sess *Session) {
			// The readiness probe below is not a session the test dialed.
			if sessions != nil && sess.RequestURI() != authProbePath {
				sessions <- sess
			}
			<-sess.Context().Done()
		}),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, ErrServerClosed) {
			t.Logf("server: %v", err)
		}
	}()
	t.Cleanup(func() { closeServer(t, srv) })

	require.Eventually(t, func() bool {
		sess, err := (&Dialer{TLSConfig: &tls.Config{InsecureSkipVerify: true}}).Dial(ctx, "moqt://"+addr+authProbePath, nil)
		if err != nil {
			return false
		}
		_ = sess.CloseWithError(NoError, "probe")
		return true
	}, 10*time.Second, 50*time.Millisecond)

	return addr
}

func dialAuth(t *testing.T, ctx context.Context, url string, mux *TrackMux) *Session {
	t.Helper()
	sess, err := (&Dialer{TLSConfig: &tls.Config{InsecureSkipVerify: true}}).Dial(ctx, url, mux)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sess.CloseWithError(NoError, "done") })
	return sess
}

// subscribeCode subscribes to path and returns the subscribe error code, or
// -1 when the subscription is accepted.
func subscribeCode(t *testing.T, ctx context.Context, sess *Session, path BroadcastPath) int {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tr, err := sess.Subscribe(ctx, path, "audio", nil)
	if err == nil {
		_ = tr.Close()
		return -1
	}
	var subErr *SubscribeError
	require.ErrorAs(t, err, &subErr)
	return int(subErr.SubscribeErrorCode())
}

func TestSession_Auth(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	addr := startAuthServer(t, ctx, roomTicketHandler, nil)
	sess := dialAuth(t, ctx, "moqt://"+addr+"/app", nil)

	t.Run("a ticket returns its grant", func(t *testing.T) {
		grant, err := sess.Auth(ctx, []byte("room1"))
		require.NoError(t, err)
		defer func() { _ = grant.Close() }() // not actionable: the test is over

		g := grant.Grant()
		require.Len(t, g.Subscribe, 1)
		assert.Equal(t, "/room/1/", g.Subscribe[0].String())
		assert.Empty(t, g.Publish)
		assert.Zero(t, g.Expires)
	})

	t.Run("an unknown ticket is refused", func(t *testing.T) {
		_, err := sess.Auth(ctx, []byte("bad"))

		var authErr *AuthError
		require.ErrorAs(t, err, &authErr)
		assert.Equal(t, UnauthorizedSessionErrorCode, authErr.Code)
		assert.Equal(t, "unknown ticket", authErr.Reason)
	})

	t.Run("an empty token returns the connection's grant", func(t *testing.T) {
		grant, err := sess.Auth(ctx, nil)
		require.NoError(t, err)
		defer func() { _ = grant.Close() }() // not actionable: the test is over

		assert.Empty(t, grant.Grant().Subscribe)
	})
}

func TestSession_Auth_NotSupported(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	addr := startAuthServer(t, ctx, nil, nil)
	sess := dialAuth(t, ctx, "moqt://"+addr, nil)

	_, err := sess.Auth(ctx, []byte("room1"))

	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	assert.Equal(t, NotSupportedErrorCode, authErr.Code)
}

func TestSession_Auth_NoHandlerEnforcesNothing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	addr := startAuthServer(t, ctx, nil, nil)
	sess := dialAuth(t, ctx, "moqt://"+addr, nil)

	assert.Equal(t, -1, subscribeCode(t, ctx, sess, authRoom1Path))
}

func TestSession_Subscribe_Scope(t *testing.T) {
	tests := map[string]struct {
		url     string
		tickets []string
		path    BroadcastPath
		want    int
	}{
		"nothing granted":             {url: "/app", path: authRoom1Path, want: int(SubscribeErrorCodeUnauthorized)},
		"the room's ticket":           {url: "/app", tickets: []string{"room1"}, path: authRoom1Path, want: -1},
		"another room's ticket":       {url: "/app", tickets: []string{"room2"}, path: authRoom1Path, want: int(SubscribeErrorCodeUnauthorized)},
		"one of two tickets":          {url: "/app", tickets: []string{"room2", "room1"}, path: authRoom1Path, want: -1},
		"the connection's credential": {url: "/app?admin", path: authRoom2Path, want: -1},
		"a publish-only ticket":       {url: "/app", tickets: []string{"publish"}, path: authRoom1Path, want: int(SubscribeErrorCodeUnauthorized)},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			addr := startAuthServer(t, ctx, roomTicketHandler, nil)
			sess := dialAuth(t, ctx, "moqt://"+addr+tt.url, nil)
			for _, ticket := range tt.tickets {
				grant, err := sess.Auth(ctx, []byte(ticket))
				require.NoError(t, err)
				defer func() { _ = grant.Close() }() // not actionable: the test is over
			}

			got := subscribeCode(t, ctx, sess, tt.path)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSession_Auth_WithdrawCancelsOnlyItsSubscriptions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	addr := startAuthServer(t, ctx, roomTicketHandler, nil)
	sess := dialAuth(t, ctx, "moqt://"+addr+"/app", nil)
	room1, err := sess.Auth(ctx, []byte("room1"))
	require.NoError(t, err)
	defer func() { _ = room1.Close() }() // not actionable: the test is over
	room2, err := sess.Auth(ctx, []byte("room2"))
	require.NoError(t, err)
	tr1, err := sess.Subscribe(ctx, authRoom1Path, "audio", nil)
	require.NoError(t, err)
	defer func() { _ = tr1.Close() }() // not actionable: the test is over
	tr2, err := sess.Subscribe(ctx, authRoom2Path, "audio", nil)
	require.NoError(t, err)
	defer func() { _ = tr2.Close() }() // not actionable: the test is over

	require.NoError(t, room2.Close())

	// Room 2's subscription ends with UNAUTHORIZED.
	var subErr *SubscribeError
	require.Eventually(t, func() bool {
		gr, err := tr2.AcceptGroup(ctx)
		if err == nil {
			gr.CancelRead(InternalGroupErrorCode)
			return false
		}
		return errors.As(err, &subErr)
	}, 5*time.Second, time.Millisecond)
	assert.Equal(t, SubscribeErrorCodeUnauthorized, subErr.SubscribeErrorCode())
	// Room 1's subscription and the session carry on.
	gctx, gcancel := context.WithTimeout(ctx, 2*time.Second)
	defer gcancel()
	gr, err := tr1.AcceptGroup(gctx)
	require.NoError(t, err)
	gr.CancelRead(InternalGroupErrorCode)
	assert.NoError(t, sess.Context().Err())
	// A new subscription to room 2 is refused.
	assert.Equal(t, int(SubscribeErrorCodeUnauthorized), subscribeCode(t, ctx, sess, authRoom2Path))
}

func TestSession_Auth_RefusalAfterGrantRevokes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	revoke := make(chan struct{})
	handler := AuthHandlerFunc(func(w *AuthWriter, r *AuthRequest) {
		if len(r.Token) == 0 {
			_ = w.Grant(Grant{})
			<-r.Context().Done()
			return
		}
		_ = w.Grant(Grant{Subscribe: []PathPattern{MustParsePathPattern("/room/1/")}})
		select {
		case <-revoke:
			_ = w.Refuse(UnauthorizedSessionErrorCode, "key retired")
		case <-r.Context().Done():
		}
	})
	addr := startAuthServer(t, ctx, handler, nil)
	sess := dialAuth(t, ctx, "moqt://"+addr+"/app", nil)
	grant, err := sess.Auth(ctx, []byte("room1"))
	require.NoError(t, err)
	tr, err := sess.Subscribe(ctx, authRoom1Path, "audio", nil)
	require.NoError(t, err)
	defer func() { _ = tr.Close() }() // not actionable: the test is over

	close(revoke)

	select {
	case <-grant.Done():
	case <-ctx.Done():
		t.Fatal("grant did not end")
	}
	var authErr *AuthError
	require.ErrorAs(t, grant.Err(), &authErr)
	assert.Equal(t, "key retired", authErr.Reason)
	assert.Equal(t, int(SubscribeErrorCodeUnauthorized), subscribeCode(t, ctx, sess, authRoom1Path))
	assert.NoError(t, sess.Context().Err())
}

func TestSession_AcceptAnnounce_PeerSeesOnlyItsScope(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	addr := startAuthServer(t, ctx, roomTicketHandler, nil)
	sess := dialAuth(t, ctx, "moqt://"+addr+"/app", nil)
	grant, err := sess.Auth(ctx, []byte("room1"))
	require.NoError(t, err)
	ar, err := sess.AcceptAnnounce("/room/")
	require.NoError(t, err)
	defer func() { _ = ar.Close() }() // not actionable: the test is over

	actx, acancel := context.WithTimeout(ctx, 5*time.Second)
	defer acancel()
	ann, err := ar.ReceiveAnnouncement(actx)
	require.NoError(t, err)
	assert.Equal(t, authRoom1Path, ann.BroadcastPath())

	// Room 2 is never announced.
	short, scancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer scancel()
	_, err = ar.ReceiveAnnouncement(short)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	// Withdrawing the ticket ends room 1's announcement.
	require.NoError(t, grant.Close())
	select {
	case <-ann.Done():
	case <-ctx.Done():
		t.Fatal("announcement did not end")
	}
}

func TestSession_Announcements_PeerPublishesOnlyItsScope(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	sessions := make(chan *Session, 1)
	addr := startAuthServer(t, ctx, roomTicketHandler, sessions)
	clientMux := NewTrackMux(0)
	clientMux.PublishFunc(ctx, "/room/1/carol", authGroupHandler)
	clientMux.PublishFunc(ctx, "/room/2/dave", authGroupHandler)
	sess := dialAuth(t, ctx, "moqt://"+addr+"/app", clientMux)
	var server *Session
	select {
	case server = <-sessions:
	case <-ctx.Done():
		t.Fatal("no session")
	}
	ar, err := server.AcceptAnnounce("/room/")
	require.NoError(t, err)
	defer func() { _ = ar.Close() }() // not actionable: the test is over

	// Nothing is published before a grant allows it.
	short, scancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer scancel()
	_, err = ar.ReceiveAnnouncement(short)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	grant, err := sess.Auth(ctx, []byte("publish"))
	require.NoError(t, err)
	defer func() { _ = grant.Close() }() // not actionable: the test is over

	actx, acancel := context.WithTimeout(ctx, 5*time.Second)
	defer acancel()
	ann, err := ar.ReceiveAnnouncement(actx)
	require.NoError(t, err)
	assert.Equal(t, BroadcastPath("/room/1/carol"), ann.BroadcastPath())
}
