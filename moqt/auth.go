package moqt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/qumo-dev/gomoqt/moqt/internal/message"
	"github.com/qumo-dev/gomoqt/transport"
)

var (
	// ErrAuthNotSupported is returned by Session.Auth when the peer reset the
	// Auth Stream before answering, as a peer without the Auth Stream does.
	ErrAuthNotSupported = errors.New("moqt: peer does not support the auth stream")

	// ErrAuthEnded is returned by AuthWriter methods once the token's grant
	// has ended: refused, withdrawn by the presenter, or its handler returned.
	ErrAuthEnded = errors.New("moqt: auth ended")
)

// Grant is what an acceptor allows for one token: the broadcast paths the
// presenter may publish to it and subscribe to from it, as patterns relative
// to the session.
type Grant struct {
	// Publish matches the paths the presenter may announce and serve.
	Publish []PathPattern
	// Subscribe matches the paths the presenter may subscribe to, fetch, and
	// discover through announcements.
	Subscribe []PathPattern
	// Expires is how long until the grant lapses, or 0 for never. It tells the
	// presenter when to present a replacement token, at millisecond precision.
	Expires time.Duration
}

// CanPublish reports whether one of the grant's publish patterns matches path.
func (g Grant) CanPublish(path BroadcastPath) bool {
	return matchAny(g.Publish, path)
}

// CanSubscribe reports whether one of the grant's subscribe patterns matches
// path.
func (g Grant) CanSubscribe(path BroadcastPath) bool {
	return matchAny(g.Subscribe, path)
}

func matchAny(patterns []PathPattern, path BroadcastPath) bool {
	for _, p := range patterns {
		if p.Match(path) {
			return true
		}
	}
	return false
}

func (g Grant) message() message.AuthOkMessage {
	m := message.AuthOkMessage{
		PublishPatterns:   make([]string, len(g.Publish)),
		SubscribePatterns: make([]string, len(g.Subscribe)),
		Expires:           uint64(g.Expires.Milliseconds()),
	}
	if g.Expires > 0 && m.Expires == 0 {
		// Below a millisecond, the grant lapses at once rather than never.
		m.Expires = 1
	}
	for i, p := range g.Publish {
		m.PublishPatterns[i] = p.String()
	}
	for i, p := range g.Subscribe {
		m.SubscribePatterns[i] = p.String()
	}
	return m
}

func grantFromMessage(m message.AuthOkMessage) (Grant, error) {
	g := Grant{
		Publish:   make([]PathPattern, len(m.PublishPatterns)),
		Subscribe: make([]PathPattern, len(m.SubscribePatterns)),
		Expires:   time.Duration(m.Expires) * time.Millisecond,
	}
	var err error
	for i, s := range m.PublishPatterns {
		if g.Publish[i], err = ParsePathPattern(s); err != nil {
			return Grant{}, err
		}
	}
	for i, s := range m.SubscribePatterns {
		if g.Subscribe[i], err = ParsePathPattern(s); err != nil {
			return Grant{}, err
		}
	}
	return g, nil
}

// AuthHandler verifies the tokens a peer presents and grants what each allows.
//
// ServeAuth is called once for each Auth Stream the peer opens, and once for
// the credential the connection carried, with an empty Token, when the
// session starts. The grant lasts while ServeAuth runs: it ends when ServeAuth
// returns, so a handler that grants waits on r.Context(), which ends when the
// presenter withdraws the token or the session ends. A handler may call
// w.Grant again to replace the grant, and w.Refuse to revoke it.
//
// The session's scope is the union of every grant in force, and it enforces
// it: a request the scope does not cover is refused with UNAUTHORIZED, and
// when the scope shrinks, what it no longer covers is cancelled while the
// session stays up. A session with no AuthHandler enforces nothing.
type AuthHandler interface {
	ServeAuth(w *AuthWriter, r *AuthRequest)
}

// AuthHandlerFunc adapts a function to an AuthHandler.
type AuthHandlerFunc func(w *AuthWriter, r *AuthRequest)

func (f AuthHandlerFunc) ServeAuth(w *AuthWriter, r *AuthRequest) {
	f(w, r)
}

// AuthRequest is a token presented by the peer.
type AuthRequest struct {
	// Token is the credential to verify, opaque to moqt. It is empty for the
	// credential the connection itself carried, such as a query in the
	// request URI, a client certificate, or nothing.
	Token []byte

	// RequestURI is the session's request URI (see Session.RequestURI). Grant
	// patterns are relative to its path.
	RequestURI string

	ctx context.Context
}

// Context returns the request's context. It ends when the presenter
// withdraws the token, or when the session ends.
func (r *AuthRequest) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

// AuthWriter answers one token.
type AuthWriter struct {
	scope *authScope
	// stream is the Auth Stream, or nil for the connection's own credential,
	// whose grant is applied without a reply.
	stream transport.Stream

	mu    sync.Mutex
	ended bool
}

// Grant grants g for the token, replacing any earlier grant: it applies g to
// the session's scope, then sends AUTH_OK. A grant too large for one AUTH_OK
// is never trimmed: the token is refused with NOT_SUPPORTED instead, and the
// returned error says so.
func (w *AuthWriter) Grant(g Grant) error {
	m := g.message()
	if m.Expires > message.MaxAuthExpires {
		return fmt.Errorf("moqt: grant expires too late: %v", g.Expires)
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if w.ended {
		return ErrAuthEnded
	}

	if m.Len() > message.MaxMessageSize {
		tooLarge := fmt.Errorf("moqt: grant too large for AUTH_OK: %d bytes", m.Len())
		return errors.Join(tooLarge, w.refuse(NotSupportedErrorCode, "grant too large"))
	}

	w.scope.set(w, g)
	if w.stream == nil {
		return nil
	}
	if err := m.Encode(w.stream); err != nil {
		return fmt.Errorf("moqt: send AUTH_OK: %w", err)
	}
	return nil
}

// Refuse refuses the token, or revokes its grant, with a session error code
// such as UnauthorizedSessionErrorCode, and ends it.
func (w *AuthWriter) Refuse(code SessionErrorCode, reason string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.ended {
		return ErrAuthEnded
	}
	return w.refuse(code, reason)
}

// refuse ends the token with AUTH_ERROR. Caller MUST hold w.mu.
func (w *AuthWriter) refuse(code SessionErrorCode, reason string) error {
	w.end()
	if w.stream == nil {
		return nil
	}
	err := message.AuthErrorMessage{Code: uint64(code), Reason: reason}.Encode(w.stream)
	_ = w.stream.Close() // not actionable: the stream is done either way
	if err != nil {
		return fmt.Errorf("moqt: send AUTH_ERROR: %w", err)
	}
	return nil
}

// finish ends the token's grant, if it has not ended, once its handler
// returned or its presenter withdrew it.
func (w *AuthWriter) finish() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.ended {
		w.end()
	}
}

// end removes the grant from the scope. Caller MUST hold w.mu.
func (w *AuthWriter) end() {
	w.ended = true
	w.scope.remove(w)
}

// safeServeAuth runs handler, recovering a panic, and ends the grant when it
// returns.
func safeServeAuth(handler AuthHandler, w *AuthWriter, r *AuthRequest) (err error) {
	defer w.finish()
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("panic during auth handling: %v", p)
		}
	}()
	handler.ServeAuth(w, r)
	return nil
}

// authScope is a session's scope: the union of the grants in force on it.
type authScope struct {
	mu     sync.RWMutex
	grants map[*AuthWriter]Grant

	// ready is closed once the connection's own credential was answered, so
	// requests that arrive first are judged against its grant.
	ready     chan struct{}
	readyOnce sync.Once
	conn      *AuthWriter

	// changed is called, without mu, after the union changed.
	changed func()
}

func newAuthScope(changed func()) *authScope {
	return &authScope{
		grants:  make(map[*AuthWriter]Grant),
		ready:   make(chan struct{}),
		changed: changed,
	}
}

func (s *authScope) set(w *AuthWriter, g Grant) {
	s.mu.Lock()
	s.grants[w] = g
	s.mu.Unlock()
	s.answered(w)
	s.changed()
}

func (s *authScope) remove(w *AuthWriter) {
	s.mu.Lock()
	_, had := s.grants[w]
	delete(s.grants, w)
	s.mu.Unlock()
	s.answered(w)
	if had {
		s.changed()
	}
}

func (s *authScope) answered(w *AuthWriter) {
	if w == s.conn {
		s.readyOnce.Do(func() { close(s.ready) })
	}
}

// wait waits until the connection's own credential was answered, and reports
// false if ctx ends first.
func (s *authScope) wait(ctx context.Context) bool {
	if s == nil {
		return true
	}
	select {
	case <-s.ready:
		return true
	case <-ctx.Done():
		return false
	}
}

// canPublish reports whether a grant in force lets the peer publish path. A
// nil scope enforces nothing.
func (s *authScope) canPublish(path BroadcastPath) bool {
	if s == nil {
		return true
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, g := range s.grants {
		if g.CanPublish(path) {
			return true
		}
	}
	return false
}

// canSubscribe reports whether a grant in force lets the peer subscribe to
// path. A nil scope enforces nothing.
func (s *authScope) canSubscribe(path BroadcastPath) bool {
	if s == nil {
		return true
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, g := range s.grants {
		if g.CanSubscribe(path) {
			return true
		}
	}
	return false
}

// AuthError is a token refused, or a grant revoked, by the peer.
type AuthError struct {
	Code   SessionErrorCode
	Reason string
}

func (err *AuthError) Error() string {
	text := err.Code.String()
	if text == "" {
		text = fmt.Sprintf("moqt: auth error %#x", uint32(err.Code))
	}
	if err.Reason != "" {
		return text + ": " + err.Reason
	}
	return text
}

// AuthGrant is a token presented to the peer, and what the peer granted for
// it. The grant lasts until Close withdraws the token, the peer revokes it,
// or the session ends.
type AuthGrant struct {
	stream transport.Stream

	mu    sync.Mutex
	grant Grant
	err   error

	done chan struct{}
}

// Grant returns the latest grant the peer sent for the token.
func (ag *AuthGrant) Grant() Grant {
	ag.mu.Lock()
	defer ag.mu.Unlock()
	return ag.grant
}

// Done is closed when the grant ends.
func (ag *AuthGrant) Done() <-chan struct{} {
	return ag.done
}

// Err returns why the grant ended once Done is closed: an *AuthError when
// the peer revoked it, or nil when it ended without a reason.
func (ag *AuthGrant) Err() error {
	ag.mu.Lock()
	defer ag.mu.Unlock()
	return ag.err
}

// Close withdraws the token, which ends its grant on the peer.
func (ag *AuthGrant) Close() error {
	return ag.stream.Close()
}

// readReplies follows the peer's answers after the first: a further AUTH_OK
// replaces the grant, and AUTH_ERROR or the end of the stream ends it.
func (ag *AuthGrant) readReplies(violation func(string)) {
	defer close(ag.done)
	for {
		g, err := readAuthReply(ag.stream)
		if err != nil {
			var authErr *AuthError
			switch {
			case errors.As(err, &authErr):
				ag.mu.Lock()
				ag.err = authErr
				ag.mu.Unlock()
			case errors.Is(err, io.EOF):
			case isProtocolViolation(err):
				violation(err.Error())
			}
			cancelStreamWithError(ag.stream, transport.StreamErrorCode(SubscribeErrorCodeInternal))
			return
		}
		ag.mu.Lock()
		ag.grant = g
		ag.mu.Unlock()
	}
}

// readAuthReply reads one acceptor message: the grant of an AUTH_OK, or an
// *AuthError for AUTH_ERROR.
func readAuthReply(r io.Reader) (Grant, error) {
	t, err := message.ReadAuthReplyType(r)
	if err != nil {
		return Grant{}, err
	}
	switch t {
	case message.AuthReplyOK:
		var m message.AuthOkMessage
		if err := m.Decode(r); err != nil {
			return Grant{}, decodeViolation("AUTH_OK", err)
		}
		g, err := grantFromMessage(m)
		if err != nil {
			return Grant{}, err
		}
		return g, nil
	case message.AuthReplyError:
		var m message.AuthErrorMessage
		if err := m.Decode(r); err != nil {
			return Grant{}, decodeViolation("AUTH_ERROR", err)
		}
		return Grant{}, &AuthError{Code: SessionErrorCode(m.Code), Reason: m.Reason}
	default:
		return Grant{}, fmt.Errorf("%w: auth reply type %#x", errAuthViolation, uint64(t))
	}
}

// errAuthViolation marks an Auth Stream message that breaks the protocol,
// which closes the session.
var errAuthViolation = errors.New("moqt: auth stream protocol violation")

// decodeViolation classifies a decode error: the stream ending or being reset
// is not the peer's message, while anything else it sent wrong is a protocol
// violation.
func decodeViolation(what string, err error) error {
	if _, ok := errors.AsType[*transport.StreamError](err); ok || errors.Is(err, io.EOF) {
		return err
	}
	return fmt.Errorf("%w: %s: %w", errAuthViolation, what, err)
}

func isProtocolViolation(err error) bool {
	return errors.Is(err, errAuthViolation) || errors.Is(err, ErrInvalidPathPattern)
}

// authStreamCancelled resets the read side of an Auth Stream whose token is
// over, so a presenter that never closes its side does not hold it open.
const authStreamCancelled transport.StreamErrorCode = 0x1

// Auth presents token to the peer on a new Auth Stream and waits for its
// first answer. An empty token presents the credential the connection already
// carried. It returns an *AuthError when the peer refuses the token, and
// ErrAuthNotSupported when the peer does not support the Auth Stream.
//
// The grant lasts until AuthGrant.Close withdraws the token, the peer
// revokes it, or the session ends. ctx bounds only the wait for the answer.
func (sess *Session) Auth(ctx context.Context, token []byte) (*AuthGrant, error) {
	if ctx == nil {
		return nil, errors.New("nil context")
	}

	if sess.closed.Load() {
		return nil, ErrClosedSession
	}

	stream, err := sess.conn.OpenStreamSync(ctx)
	if err != nil {
		if appErr, ok := errors.AsType[*transport.ApplicationError](err); ok {
			return nil, &SessionError{ApplicationError: appErr}
		}
		return nil, fmt.Errorf("failed to open stream for auth: %w", err)
	}

	if err := message.StreamTypeAuth.Encode(stream); err != nil {
		cancelStreamWithError(stream, authStreamCancelled)
		return nil, fmt.Errorf("failed to encode stream type message: %w", err)
	}
	if err := (message.AuthMessage{Token: token}).Encode(stream); err != nil {
		cancelStreamWithError(stream, authStreamCancelled)
		return nil, fmt.Errorf("failed to encode AUTH message: %w", err)
	}

	abandoned := abandonOnDone(ctx, stream)
	g, err := readAuthReply(stream)
	if abandoned() {
		return nil, abandonedErr("auth", ctx)
	}
	if err != nil {
		cancelStreamWithError(stream, authStreamCancelled)
		if _, ok := errors.AsType[*AuthError](err); ok {
			return nil, err
		}
		if _, ok := errors.AsType[*transport.StreamError](err); ok {
			return nil, ErrAuthNotSupported
		}
		if errors.Is(err, io.EOF) {
			return nil, ErrAuthEnded
		}
		if isProtocolViolation(err) {
			sess.terminateProtocolViolation(err.Error())
		}
		return nil, fmt.Errorf("failed to read auth reply: %w", err)
	}

	ag := &AuthGrant{stream: stream, grant: g, done: make(chan struct{})}
	sess.wg.Go(func() {
		ag.readReplies(sess.terminateProtocolViolation)
	})
	return ag, nil
}

// handleAuthStream reads the peer's AUTH and serves it to the AuthHandler for
// as long as the token applies.
func (sess *Session) handleAuthStream(stream transport.Stream) {
	var am message.AuthMessage
	if err := am.Decode(stream); err != nil {
		if err := decodeViolation("AUTH", err); isProtocolViolation(err) {
			sess.terminateProtocolViolation(err.Error())
		}
		cancelStreamWithError(stream, authStreamCancelled)
		return
	}

	if sess.authHandler == nil {
		err := message.AuthErrorMessage{
			Code:   uint64(NotSupportedErrorCode),
			Reason: "tokens are not verified in band",
		}.Encode(stream)
		if err != nil {
			sess.logError("failed to send AUTH_ERROR", err)
		}
		stream.CancelRead(authStreamCancelled)
		return
	}

	reqCtx, cancelReq := sess.streamContext(stream)
	defer cancelReq()
	reqCtx, withdraw := context.WithCancel(reqCtx)
	defer withdraw()

	w := &AuthWriter{scope: sess.scope, stream: stream}

	// The presenter sends nothing after AUTH: the end of its side withdraws
	// the token, and anything more is a protocol violation.
	sess.wg.Go(func() {
		var b [1]byte
		n, _ := stream.Read(b[:])
		if n > 0 {
			sess.terminateProtocolViolation("message after AUTH")
		}
		w.finish()
		withdraw()
	})

	r := &AuthRequest{Token: am.Token, RequestURI: sess.path, ctx: reqCtx}
	if err := safeServeAuth(sess.authHandler, w, r); err != nil {
		sess.logError("auth handler error", err)
	}
	stream.CancelRead(authStreamCancelled)
}

// mayServe reports whether the peer's scope lets it subscribe to path, waiting
// first for the connection's own grant.
func (sess *Session) mayServe(path BroadcastPath) bool {
	if sess.scope == nil {
		return true
	}
	return sess.scope.wait(sess.ctx) && sess.scope.canSubscribe(path)
}

// rescope applies a change of the peer's scope: it cancels each subscription
// the scope no longer covers with UNAUTHORIZED, and re-filters the
// announcements on every scoped announce stream.
func (sess *Session) rescope() {
	sess.trackWriterMapLocker.RLock()
	var revoked []*TrackWriter
	for _, tw := range sess.trackWriters {
		if !sess.scope.canSubscribe(tw.BroadcastPath) {
			revoked = append(revoked, tw)
		}
	}
	sess.trackWriterMapLocker.RUnlock()
	// CloseWithError removes the writer from the map itself.
	for _, tw := range revoked {
		tw.CloseWithError(SubscribeErrorCodeUnauthorized)
	}

	sess.scopedMu.Lock()
	writers := make([]*AnnouncementWriter, 0, len(sess.announceWriters))
	for aw := range sess.announceWriters {
		writers = append(writers, aw)
	}
	readers := make([]*AnnouncementReader, 0, len(sess.announceReaders))
	for ar := range sess.announceReaders {
		readers = append(readers, ar)
	}
	sess.scopedMu.Unlock()

	for _, aw := range writers {
		aw.rescope()
	}
	for _, ar := range readers {
		ar.rescope()
	}
}
