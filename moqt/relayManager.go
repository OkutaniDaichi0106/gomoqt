package moqt

import (
	"errors"
	"sync"
)

func newRelayManager() relayManager {
	return relayManager{
		trackNamespaceTree: trackNamespaceTree{
			rootNode: &trackNamespaceNode{
				value:    "",
				children: make(map[string]*trackNamespaceNode),
			},
		},
	}
}

type relayManager struct {
	trackNamespaceTree trackNamespaceTree
}

func (tm *relayManager) newTrackNamespace(trackNamespace []string) *trackNamespaceNode {
	return tm.trackNamespaceTree.insert(trackNamespace)
}

func (tm *relayManager) findTrackNamespace(trackNamespace []string) (*trackNamespaceNode, bool) {
	return tm.trackNamespaceTree.trace(trackNamespace)
}

func (tm *relayManager) removeTrackNamespace(trackNamespace []string) error {
	return tm.trackNamespaceTree.remove(trackNamespace)
}

func (tm *relayManager) findTrack(trackNamespace []string, trackName string) (*trackNameNode, bool) {
	tnsNode, ok := tm.findTrackNamespace(trackNamespace)
	if !ok {
		return nil, false
	}

	return tnsNode.findTrackNameNode(trackName)
}

type trackNamespaceTree struct {
	rootNode *trackNamespaceNode
}

func (tree trackNamespaceTree) insert(tns []string) *trackNamespaceNode {
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

func (tree trackNamespaceTree) remove(tns []string) error {
	_, err := tree.rootNode.removeDescendants(tns, 0)
	return err
}

func (tree trackNamespaceTree) trace(tns []string) (*trackNamespaceNode, bool) {
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
	 * Track Name Nodes
	 */
	tracks map[string]*trackNameNode

	/*
	 * The origin session
	 */
	origin *Session
}

type trackNameNode struct {
	mu sync.RWMutex

	/*
	 * The string value of the Track Name
	 */
	value string

	/*
	 *
	 */
	destinations []*Session
}

func (node *trackNamespaceNode) removeDescendants(tns []string, depth int) (bool, error) {
	if node == nil {
		return false, errors.New("track namespace not found at " + tns[depth])
	}

	node.mu.Lock()
	defer node.mu.Unlock()

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
		return false, errors.New("track namespace not found at " + value)
	}

	ok, err := child.removeDescendants(tns, depth+1)
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

func (node *trackNamespaceNode) findTrackNameNode(trackName string) (*trackNameNode, bool) {
	node.mu.RLock()
	defer node.mu.RUnlock()

	tnNode, ok := node.tracks[trackName]
	if !ok {
		return nil, false
	}

	return tnNode, true
}

func (node *trackNamespaceNode) newTrackNameNode(trackName string) *trackNameNode {
	node.mu.Lock()
	defer node.mu.Unlock()

	node.tracks[trackName] = &trackNameNode{
		value: trackName,
	}

	return node.tracks[trackName]
}
