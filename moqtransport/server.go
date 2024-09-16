package moqtransport

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/webtransport-go"
)

/*
 * Server
 */
type Server struct {
	WebTransportServer *webtransport.Server
	Versions           []Version
	onPublisher        func(PublisherSession)
	onSubscriber       func(SubscriberSession)
}

func (s *Server) Upgrade(w http.ResponseWriter, r *http.Request) (Session, error) {
	// Establish HTTP/3 Connection
	wtSession, err := s.WebTransportServer.Upgrade(w, r)
	if err != nil {
		log.Printf("upgrading failed: %s", err)
		w.WriteHeader(500)
		return nil, err
	}

	cs := clientSession{
		wtSession: wtSession,
	}

	// Receive CLIENT_SETUP message
	// Get versions supported by the client
	versions, err := cs.receiveClientSetup()
	if err != nil {
		return nil, err
	}

	// Select later version
	selectedVersion, err := selectLaterVersion(versions, s.Versions)
	if err != nil {
		return nil, err
	}

	// Set the selected version to the version of the session
	cs.selectedVersion = selectedVersion

	//
	switch cs.role {
	case PUB:

		cs.roleHandler = func() {
			sess := PublisherSession{
				Session:         cs,
				controlChannel:  make(chan []byte, 1<<4),
				maxSubscriberID: cs.maxSubscribeID,
			}
			s.onPublisher(sess)
		}
	case SUB:

		cs.roleHandler = func() {
			sess := SubscriberSession{
				Session:        cs,
				controlChannel: make(chan []byte, 1<<4),
				subscriptions: make(map[TrackAlias]struct {
					maxSubscribeID   subscribeID
					trackSubscribeID subscribeID
				}),
			}
			s.onSubscriber(sess)
		}
	case PUB_SUB:
		//TODO
	default:
		return nil, ErrInvalidRole
	}

	// Send SERVER_SETUP message
	err = cs.sendServerSetup()
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

func (s *Server) OnPublisher(op func(PublisherSession)) {
	s.onPublisher = op
}
func (s *Server) OnSubscriber(op func(SubscriberSession)) {
	s.onSubscriber = op
}
func (s *Server) ListenAndServeTLS(cert, key string) error {
	return s.WebTransportServer.ListenAndServeTLS(cert, key)
}

func (s Server) GoAway(url string, duration time.Duration) {
	gm := GoAwayMessage{
		NewSessionURI: url,
	}
	for trackNamespace, pubSess := range publishers.index {
		// Send GO_AWAY message to the publisher
		go func(pubSess *PublisherSession) {

			_, err := pubSess.getControlStream().Write(gm.serialize())
			if err != nil {
				log.Println(err)
			}

			time.Sleep(duration)

			err = pubSess.getWebtransportSession().CloseWithError(GetSessionError(ErrGoAwayTimeout))
			if err != nil {
				log.Println(err)
				// Send Terminate Internal Error, if sending prior error was failed
				err = pubSess.getWebtransportSession().CloseWithError(GetSessionError(ErrTerminationFailed))
				if err != nil {
					log.Println(err)
				}
			}
		}(pubSess)

		// Send GO_AWAY message to the subscribers
		for _, subSess := range subscribers[trackNamespace].sessions {
			go func(subSess *SubscriberSession) {
				_, err := subSess.getControlStream().Write(gm.serialize())
				if err != nil {
					// TODO: Handle this error
					log.Println(err)
				}

				// Wait for the duration
				time.Sleep(duration)

			}(subSess)
		}
	}

}

/*
 * The key is the Track Namespace
 */
var subscribers subscribersIndex

type subscribersIndex map[string]*destinations

type destinations struct {
	sessions []*SubscriberSession
	mu       sync.Mutex
}

func (dest *destinations) add(sess *SubscriberSession) {
	dest.mu.Lock()
	defer dest.mu.Unlock()
	dest.sessions = append(dest.sessions, sess)
}

func (dest *destinations) delete(sess *SubscriberSession) {
	dest.mu.Lock()
	defer dest.mu.Unlock()
	dest.sessions = append(dest.sessions, sess)
}

/*
 * Index for searching a Publisher's Agent by Namespace
 */
var publishers publishersIndex

type publishersIndex struct {
	mu    sync.Mutex
	index map[string]*PublisherSession
}

func (pi *publishersIndex) add(session *PublisherSession) {
	if pi.index == nil {
		pi.index = make(map[string]*PublisherSession)
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

var ErrUnsuitableRole = errors.New("the role cannot perform the operation ")
var ErrUnexpectedMessage = errors.New("received message is not a expected message")
var ErrInvalidRole = errors.New("given role is invalid")
var ErrDuplicatedNamespace = errors.New("given namespace is already registered")
var ErrNoAgent = errors.New("no agent")
