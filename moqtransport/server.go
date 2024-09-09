package moqtransport

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/quic-go/webtransport-go"
)

/*
 * Index for searching a Publisher's Agent by Namespace
 */
var publishers publishersIndex

type publishersIndex struct {
	mu    sync.Mutex
	index map[string]*SessionWithPublisher
}

func (pi *publishersIndex) add(session *SessionWithPublisher) {
	publishers.mu.Lock()
	defer publishers.mu.Unlock()
	pi.index[session.latestAnnounceMessage.TrackNamespace] = session
}
func (pi *publishersIndex) delete(trackNamespace string) {
	publishers.mu.Lock()
	defer publishers.mu.Unlock()
	delete(pi.index, trackNamespace)
}

/*
 * Announcements received from publishers
 */
var announcements announcementIndex

type announcementIndex struct {
	mu    sync.Mutex
	index map[string]AnnounceMessage
}

func (ai *announcementIndex) add(am AnnounceMessage) {
	announcements.mu.Lock()
	defer announcements.mu.Unlock()
	ai.index[am.TrackNamespace] = am
}
func (ai *announcementIndex) delete(trackNamespace string) {
	announcements.mu.Lock()
	defer announcements.mu.Unlock()
	delete(ai.index, trackNamespace)
}

/*
 * Server Agent
 * You should use this in a goroutine such as http.HandlerFunc
 *
 * Server will perform the following operation
 * - Waiting connections by Client
 * - Accepting bidirectional stream to send control messages
 * - Receiving SETUP_CLIENT message from the client
 * - Sending SETUP_SERVER message to the client
 * - Terminating sessions
 */

type Server struct {
	WebTransportServer *webtransport.Server
}

func (s *Server) ConnectAndSetup(w http.ResponseWriter, r *http.Request) (*Session, error) {
	// Establish HTTP/3 Connection
	wtSession, err := s.WebTransportServer.Upgrade(w, r)
	if err != nil {
		log.Printf("upgrading failed: %s", err)
		w.WriteHeader(500)
		return nil, err
	}

	moqtSession := Session{
		wtSession: wtSession,
	}

	//moqtSession.setup(s.SupportedVersions)

	return &moqtSession, nil
}

func (s *Server) init() error {
	//TODO
	return nil
}

func (s *Server) ListenAndServeTLS(cert, key string) error {
	err := s.init()
	if err != nil {
		return err
	}
	return s.WebTransportServer.ListenAndServeTLS(cert, key)
}

func Announcements() []AnnounceMessage {
	allAnnouncements := make([]AnnounceMessage, len(announcements.index))

	for _, am := range announcements.index {
		allAnnouncements = append(allAnnouncements, am)
	}

	return allAnnouncements
}

var ErrUnsuitableRole = errors.New("the role cannot perform the operation ")
var ErrUnexpectedMessage = errors.New("received message is not a expected message")
var ErrInvalidRole = errors.New("given role is invalid")
var ErrDuplicatedNamespace = errors.New("given namespace is already registered")
var ErrNoAgent = errors.New("no agent")
