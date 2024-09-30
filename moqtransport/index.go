package moqtransport

import (
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
)

var trackManager = TrackManager{
	trackNamespaceTree: *newTrackNamespaceTree(),
}

type TrackManager struct {
	trackNamespaceTree TrackNamespaceTree
}

func (tm *TrackManager) newTrackNamespace(trackNamespace moqtmessage.TrackNamespace) *trackNamespaceNode {
	return tm.trackNamespaceTree.insert(trackNamespace)
}

func (tm *TrackManager) findTrackNamespace(trackNamespace moqtmessage.TrackNamespace) (*trackNamespaceNode, bool) {
	return tm.trackNamespaceTree.trace(trackNamespace)
}

func (tm *TrackManager) addTrackName(trackNamespace moqtmessage.TrackNamespace, trackName string) *trackNameNode {
	tnsNode := tm.newTrackNamespace(trackNamespace)

	return tnsNode.newTrackNameNode(trackName)
}

func (tm *TrackManager) findTrackName(trackNamespace moqtmessage.TrackNamespace, trackName string) (*trackNameNode, bool) {
	tnsNode, ok := tm.findTrackNamespace(trackNamespace)
	if !ok {
		return nil, false
	}

	return tnsNode.findTrackName(trackName)
}

type TrackNamespaceTree struct {
	rootNode *trackNamespaceNode
}

func newTrackNamespaceTree() *TrackNamespaceTree {
	return &TrackNamespaceTree{
		rootNode: &trackNamespaceNode{
			value:    "",
			children: make(map[string]*trackNamespaceNode),
		},
	}
}

func (tree TrackNamespaceTree) insert(tns moqtmessage.TrackNamespace) *trackNamespaceNode {
	currentNode := tree.rootNode
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
	_, err := tree.rootNode.remove(tns, 0)
	return err
}

func (tree TrackNamespaceTree) trace(tns moqtmessage.TrackNamespace) (*trackNamespaceNode, bool) {
	return tree.rootNode.trace(tns...)
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
	 * The Object Forwarding Preference
	 */
	ofp moqtmessage.ObjectForwardingPreference

	/*
	 *
	 */
	contentStatus *contentStatus

	/*
	 *
	 */
	//destinationSession []*PublishingSession
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
	node.mu.RLock()
	defer node.mu.RUnlock()

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

func (node *trackNamespaceNode) findTrackName(trackName string) (*trackNameNode, bool) {
	tnNode, ok := node.tracks[trackName]
	if !ok {
		return nil, false
	}

	return tnNode, true
}

func (node *trackNamespaceNode) newTrackNameNode(trackName string) *trackNameNode {
	node.tracks[trackName] = &trackNameNode{
		value: trackName,
		contentStatus: &contentStatus{
			contentExists:   false,
			largestGroupID:  0,
			largestObjectID: 0,
		},
	}

	return node.tracks[trackName]
}

type TrackStatus struct {
	Code         moqtmessage.TrackStatusCode
	LastGroupID  moqtmessage.GroupID
	LastObjectID moqtmessage.ObjectID
}
