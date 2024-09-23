package moqtransport

import (
	"errors"
	"go-moq/moqtransport/moqtmessage"
	"sync"
)

var trackManager TrackManager

type TrackManager struct {
	trackNamespaceTree TrackNamespaceTree

	// fullTrackNameFromAlias map[moqtmessage.TrackAlias]struct {
	// 	moqtmessage.TrackNamespace
	// 	TrackName string
	// }

	/*
	 * map[upstream Track Alias]downstream Track Alias
	 */
	//router map[moqtmessage.TrackAlias]moqtmessage.TrackAlias
}

type TrackNamespaceTree struct {
	root *trackNamespaceNode
}

func newTrackNamespaceTree() *TrackNamespaceTree {
	return &TrackNamespaceTree{
		root: &trackNamespaceNode{
			value:    "",
			children: make(map[string]*trackNamespaceNode),
		},
	}
}

func (tree TrackNamespaceTree) insert(tns moqtmessage.TrackNamespace) *trackNamespaceNode {
	currentNode := tree.root
	for _, nodeValue := range tns {
		// Verify the node has a child with the node value
		child, exists := currentNode.children[nodeValue]
		if exists && child != nil {
			// Move to the next child node
			currentNode = child
		} else {
			// Create new node and move to the new node
			newNode := &trackNamespaceNode{
				value:    nodeValue,
				children: make(map[string]*trackNamespaceNode),
			}
			currentNode.children[nodeValue] = newNode

			currentNode = newNode
		}
	}
	return currentNode
}

func (tree TrackNamespaceTree) remove(tns moqtmessage.TrackNamespace) error {
	_, err := tree.root.remove(tns, 0)
	return err
}

func (tree TrackNamespaceTree) trace(tns moqtmessage.TrackNamespace) (*trackNamespaceNode, bool) {
	return tree.root.trace(tns...)
}

type trackNamespaceNode struct {
	mu sync.RWMutex

	/*
	 * The string value in the tuple of the Track Namespace
	 */
	value string

	/*
	 * Children of the node
	 */
	children map[string]*trackNamespaceNode

	/*
	 * Parameters in the ANNOUNCE message
	 */
	params *moqtmessage.Parameters

	//
	tracks map[string]*trackNameNode
}

type trackNameNode struct {
	mu sync.RWMutex

	/*
	 * The string value of the Track Name
	 */
	value string

	/*
	 * Session with the publisher
	 */
	sessionWithPublisher *SubscribingSession

	/*
	 * Session with subscribers
	 */
	sessionWithSubscriber []*PublishingSession
}

func (node *trackNamespaceNode) remove(tns moqtmessage.TrackNamespace, depth int) (bool, error) {
	if node == nil {
		return false, errors.New("node not found at value" + tns[depth])
	}

	if depth > len(tns) {
		return false, errors.New("invalid depth")
	}

	if depth == len(tns) {
		if len(node.children) == 0 {
			return true, nil
		}

		return false, nil
	}

	value := tns[depth]

	child, exists := node.children[value]

	if !exists {
		return false, errors.New("child node not found at value" + value)
	}

	ok, err := child.remove(tns, depth+1)
	if err != nil {
		return false, err
	}

	if ok {
		node.mu.Lock()
		defer node.mu.Unlock()
		delete(node.children, value)

		return (len(node.children) == 0), nil
	}

	return false, nil
}

func (node *trackNamespaceNode) trace(values ...string) (*trackNamespaceNode, bool) {
	currentNode := node
	for _, nodeValue := range values {
		// Verify the node has a child with the node value
		child, exists := currentNode.children[nodeValue]
		if exists && child != nil {
			// Move to the next child node
			currentNode = child
		} else {
			return nil, false
		}
	}

	return currentNode, true
}

// func (tm *TrackManager) getTrackStatus(tns moqtmessage.TrackNamespace, tn string) (*TrackStatus, error) {
// 	node, err := tm.trackNamespaceTree.trace(tns)
// 	if err != nil {
// 		return nil, err
// 	}

// 	for trackName, track := range node.tracks {
// 		if trackName == tn {
// 			return track.status, nil
// 		}
// 	}

// 	return nil, errors.New("track not found")
// }

func (tm *TrackManager) addAnnouncement(announcement moqtmessage.AnnounceMessage) {
	node := tm.trackNamespaceTree.insert(announcement.TrackNamespace)

	node.mu.Lock()
	defer node.mu.Unlock()

	node.params = &announcement.Parameters
}

// func (tm *TrackManager) addTrack(tns moqtmessage.TrackNamespace, tn string) error {
// 	node, err := tm.trackNamespaceTree.trace(tns)
// 	if err != nil {
// 		return err
// 	}

// 	for trackName, track := range node.tracks {
// 		if trackName == tn {

// 		}
// 	}
// }

type TrackStatus struct {
	Code         moqtmessage.TrackStatusCode
	LastGroupID  moqtmessage.GroupID
	LastObjectID moqtmessage.ObjectID
}

//var pubSessions map[moqtmessage.SubscribeID]PubSession

// type PubSessionManager struct {
// 	mu sync.RWMutex

// 	pubSessions map[sessionID]PubSession
// }

// var subSessionManager SubSessionManager

// type SubSessionManager struct {
// 	mu sync.RWMutex

// 	subSessions []SubSession
// }

// func (manager *SubSessionManager) getSession(id sessionID) SubSession {
// 	manager.mu.RLock()
// 	defer manager.mu.RUnlock()

// 	return manager.subSessions[id]
// }

// func (manager *SubSessionManager) addSession(sess SubSession) error {
// 	manager.mu.Lock()
// 	defer manager.mu.Unlock()

// 	_, ok := manager.subSessions[sess.sessionID]
// 	if ok {
// 		return errors.New("duplicate session ID")
// 	}

// 	manager.subSessions[sess.sessionID] = sess

// 	return nil
// }

// func getTrackStatus(tns moqtmessage.TrackNamespace, tn string) (TrackStatus, error) {
// 	// Set default Track Status
// 	defaultStatus := TrackStatus{
// 		Code: moqtmessage.TRACK_STATUS_UNTRACEABLE_RELAY,
// 	}

// 	// Listen to the local Track Manager
// 	status, err := trackManager.getTrackStatus(tns, tn)
// 	if err != nil {
// 		return defaultStatus, err
// 	}
// 	if status != nil {
// 		return *status, nil
// 	}

// 	for _, subSession := range subSessionManager.subSessions {
// 		subSession
// 	}
// }
