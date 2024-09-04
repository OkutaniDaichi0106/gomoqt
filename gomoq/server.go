package gomoq

import (
	"errors"
	"sync"

	"github.com/quic-go/webtransport-go"
)

var server Server

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
	/*
	 * Supported versions by the server
	 */
	SupportedVersions []Version

	/*
	 * Agents runs on the server
	 */
	agents []*Agent

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
	onPublisher  func(*Agent)
	onSubscriber func(*Agent)
	onPubSub     func(*Agent)
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

func (s *Server) OnPublisher(op func(*Agent)) {
	s.onPublisher = func(agent *Agent) {
		if agent.role != PUB {
			panic("unsuitable role")
		}
		op(agent)
	}
}

func (s *Server) OnSubscriber(op func(*Agent)) {
	s.onSubscriber = func(agent *Agent) {
		if agent.role != SUB {
			panic("unsuitable role")
		}
		op(agent)
	}
}

func (s *Server) OnPubSub(op func(*Agent)) {
	s.onPubSub = func(agent *Agent) {
		if agent.role != PUB_SUB {
			panic("unsuitable role")
		}
		op(agent)
	}
}

func (s *Server) SetupParameters(params Parameters) {
	s.setupParameters = params
}

func (s *Server) NewAgent(sess *webtransport.Session) *Agent {
	a := Agent{
		session: sess,
	}
	s.agents = append(s.agents, &a)

	return &a
}

func (s *Server) getPublisherAgent(trackNamespace string) (*Agent, error) {
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
	index map[string]*Agent
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

// func (sMap *subscriberMap) getValue(trackNamespace string) (*Agent, error) {
// 	agent, ok := sMap.index[trackNamespace]
// 	if !ok || agent == nil {
// 		return nil, ErrNoAgent
// 	}
// 	return agent, nil
// }

func (sMap *subscriberMap) add(trackNamespace string, agent *Agent) error {
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
