package moqt

import (
	"errors"
	"sync"
)

var defaultRelayManager = NewRelayManager()

func NewRelayManager() *RelayManager {
	return &RelayManager{
		trackPathTree: trackPathTree{
			rootNode: &trackPrefixNode{
				trackPrefixPart: "",
				children:        make(map[string]*trackPrefixNode),
			},
		},
	}
}

type RelayManager struct {
	trackPathTree trackPathTree
}

type trackPathTree struct {
	rootNode *trackPrefixNode
}

func (tree trackPathTree) insert(trackParts []string) (*trackPrefixNode, bool) {
	// Set the current node to the root node
	currentNode := tree.rootNode
	//
	var exists bool

	// track the tree
	for _, trackPart := range trackParts {
		// Verify the node has a child with the node value
		var child *trackPrefixNode
		child, exists = currentNode.children[trackPart]
		if exists && child != nil {
			// Move to the next child node
			currentNode = child
		} else {
			// Create new node and move to the new node
			newNode := &trackPrefixNode{
				trackPrefixPart: trackPart,
				children:        make(map[string]*trackPrefixNode),
			}
			currentNode.children[trackPart] = newNode

			currentNode = newNode
		}
	}

	return currentNode, exists
}

func (tree trackPathTree) remove(tns []string) error {
	_, err := tree.rootNode.removeDescendants(tns, 0)
	return err
}

func (tree trackPathTree) trace(tns []string) (*trackPrefixNode, bool) {
	return tree.rootNode.trace(tns...)
}

type trackPrefixNode struct {
	mu sync.RWMutex

	/*
	 * A Part of the Track Prefix
	 */
	trackPrefixPart string

	/*
	 * Children of the node
	 */
	children map[string]*trackPrefixNode

	/*
	 * Track Name Nodes
	 */
	trackNames map[string]*trackNameNode

	//
	interests []*ReceivedInterest
}

type trackNameNode struct {
	relayer Relayer

	mu sync.RWMutex

	trackNamePart string
}

func (node *trackPrefixNode) removeDescendants(tns []string, depth int) (bool, error) {
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

func (node *trackPrefixNode) trace(values ...string) (*trackPrefixNode, bool) {
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

func (node *trackPrefixNode) findTrackName(trackName string) (*trackNameNode, bool) {
	node.mu.RLock()
	defer node.mu.RUnlock()

	tnNode, ok := node.trackNames[trackName]
	if !ok {
		return nil, false
	}

	return tnNode, true
}

/*
 * Create a new Track Name node when a subscriber makes a new subscription
 *
 */
func (tnsNode *trackPrefixNode) newTrackName(trackName string) *trackNameNode {
	tnsNode.mu.Lock()
	defer tnsNode.mu.Unlock()

	tnsNode.trackNames[trackName] = &trackNameNode{
		trackNamePart: trackName,
	}

	return tnsNode.trackNames[trackName]
}
