package moqtransport

import (
	"errors"
	"sync"

	"github.com/quic-go/webtransport-go"
)

var SERVER *Server

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

	/*
	 * Supported versions by the server
	 */
	SupportedVersions []Version

	/*
	 * Agents runs on the server
	 */
	agents []AgenterWithRole

	setupParameters Parameters

	/*
	 * Index for searching a Publisher's Agent by Namespace
	 */
	publishers publisherMap

	/*
	 * Index for searching Subscriber's Agents by Namespace
	 */
	subscribers subscriberMap

	/*
	 * Announcements received from publishers
	 */
	announcements announcementMap

	/*
	 *
	 */
	subscriptionCondition func(SubscribeMessage) bool

	/****/
	handle ServerHandler
	//onPublisher  func(publisherAgenter)
	//onSubscriber func(subscriberAgenter)
}

func (s *Server) init() error {
	SERVER = s
	return nil
}

func (s *Server) ListenAndServeTLS(cert, key string) error {
	err := s.init()
	if err != nil {
		return err
	}
	return s.WebTransportServer.ListenAndServeTLS(cert, key)
}

type ServerHandler interface {
	OnPublisher()
}

func (s *Server) OnPublisher(op func(publisherAgenter)) {
	s.handle.OnPublisher = op
}

func (s *Server) OnSubscriber(op func(subscriberAgenter)) {
	s.onSubscriber = op
}

func Handle(handler ServerHandler) {
	SERVER.onPublisher = handler.OnPublisher
}

func (s *Server) Announcements() []AnnounceMessage {
	// Copy the current value for safety
	annnounceMap := s.announcements.index
	announcements := make([]AnnounceMessage, 0, len(annnounceMap))
	for _, announcement := range annnounceMap {
		announcements = append(announcements, announcement)
	}
	return announcements
}

func (s *Server) SetupParameters(params Parameters) {
	s.setupParameters = params
}

func (s *Server) getOriginAgent(trackNamespace string) (publisherAgenter, error) {
	agent, ok := s.publishers.index[trackNamespace]
	if !ok || agent == nil {
		return nil, ErrNoAgent
	}
	return agent, nil
}

type announcementMap struct {
	mu    sync.Mutex
	index map[string]AnnounceMessage
}

func (aMap *announcementMap) add(am AnnounceMessage) error {
	// Initialize if the map is nil
	if aMap.index == nil {
		aMap.index = make(map[string]AnnounceMessage, 1<<10)
	}

	aMap.mu.Lock()
	defer aMap.mu.Unlock()

	_, ok := aMap.index[am.TrackNamespace]
	if ok {
		return errors.New("duplicate announcement") //TODO: Is duplicating announcements considered an error?
	}
	aMap.index[am.TrackNamespace] = am

	return nil
}

func (aMap *announcementMap) delete(am AnnounceMessage) error {
	aMap.mu.Lock()
	defer aMap.mu.Unlock()

	_, ok := aMap.index[am.TrackNamespace]
	if !ok {
		return errors.New("no such announcement")
	}

	delete(aMap.index, am.TrackNamespace)

	return nil
}

type publisherMap struct {
	index map[string]publisherAgenter
	mu    sync.Mutex
}

// func (pMap *publisherMap) getValue(trackNamespace string) (*Agent, error) {
// 	agent, ok := pMap.index[trackNamespace]
// 	if !ok || agent == nil {
// 		return nil, ErrNoAgent
// 	}
// 	return agent, nil
// }

func (pMap *publisherMap) add(trackNamespace string, agent *Agent) error {
	// Initialize if the map is nil
	if pMap.index == nil {
		pMap.index = make(map[string]*Agent, 1<<10)
	}

	pMap.mu.Lock()
	defer pMap.mu.Unlock()

	_, ok := pMap.index[trackNamespace]
	if ok {
		return errors.New("duplicate announcement") //TODO: Is duplicating announcements considered an error?
	}
	pMap.index[trackNamespace] = agent

	return nil
}

func (pMap *publisherMap) delete(trackNamespace string) error {
	pMap.mu.Lock()
	defer pMap.mu.Unlock()

	_, ok := pMap.index[trackNamespace]
	if !ok {
		return errors.New("no such announcement")
	}

	delete(pMap.index, trackNamespace)

	return nil
}

type subscriberMap struct {
	index map[string][]*Agent
	mu    sync.Mutex
}

func (sMap *subscriberMap) add(trackNamespace string, agent *Agent) error {
	// Initialize if the map is nil
	if sMap.index == nil {
		sMap.index = make(map[string][]*Agent, 1<<10)
	}

	sMap.mu.Lock()
	defer sMap.mu.Unlock()

	_, ok := sMap.index[trackNamespace]
	if ok {
		return errors.New("duplicate announcement") //TODO: Is duplicating announcements considered an error?
	}
	sMap.index[trackNamespace] = append(sMap.index[trackNamespace], agent)

	return nil
}

func (sMap *subscriberMap) delete(trackNamespace string) error {
	sMap.mu.Lock()
	defer sMap.mu.Unlock()

	_, ok := sMap.index[trackNamespace]
	if !ok {
		return errors.New("no such announcement")
	}

	delete(sMap.index, trackNamespace)

	return nil
}

var ErrUnsuitableRole = errors.New("the role cannot perform the operation ")
var ErrUnexpectedMessage = errors.New("received message is not a expected message")
var ErrInvalidRole = errors.New("given role is invalid")
var ErrDuplicatedNamespace = errors.New("given namespace is already registered")
var ErrNoAgent = errors.New("no agent")
