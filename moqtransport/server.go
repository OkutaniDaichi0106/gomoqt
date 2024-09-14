package moqtransport

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/quic-go/webtransport-go"
)

/*
 * Index for searching a Publisher's Agent by Namespace
 */
var publishers publishersIndex

type publishersIndex struct {
	mu    sync.Mutex
	index map[string]*PublisherSession
}

func (pi *publishersIndex) init() {
	pi.index = make(map[string]*PublisherSession)
}

func (pi *publishersIndex) add(session *PublisherSession) {
	if pi.index == nil {
		pi.init()
	}

	publishers.mu.Lock()
	defer publishers.mu.Unlock()
	fullTrackNamespace := strings.Join(session.latestAnnounceMessage.TrackNamespace, "")
	pi.index[fullTrackNamespace] = session
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
	if ai.index == nil {
		ai.index = make(map[string]AnnounceMessage)
	}
	announcements.mu.Lock()
	defer announcements.mu.Unlock()
	fullTrackNamespace := strings.Join(am.TrackNamespace, "")
	ai.index[fullTrackNamespace] = am
}

func (ai *announcementIndex) delete(trackNamespace string) {
	announcements.mu.Lock()
	defer announcements.mu.Unlock()

	// Delete
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
	Versions           []Version
}

func (s *Server) Upgrade(w http.ResponseWriter, r *http.Request) (*ClientSession, error) {
	// Establish HTTP/3 Connection
	wtSession, err := s.WebTransportServer.Upgrade(w, r)
	if err != nil {
		log.Printf("upgrading failed: %s", err)
		w.WriteHeader(500)
		return nil, err
	}

	moqtSession := ClientSession{
		wtSession:         wtSession,
		supportedVersions: s.Versions,
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
