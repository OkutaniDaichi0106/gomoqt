package moqt

import (
	"errors"
	"strings"
	"sync"
)

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

func (rm *RelayManager) AddRelayer(trackPath string, relayer *Relayer) error {
	trackParts := splitTrackPath(trackPath)

	// Insert the track path to the tree
	nameNode, err := rm.trackPathTree.insertTrackName(trackParts[:len(trackParts)-1], trackParts[len(trackParts)-1])
	if nameNode == nil {
		return err
	}

	nameNode.mu.Lock()
	defer nameNode.mu.Unlock()

	if nameNode.relayer != nil {
		return ErrDuplicatedTrack
	}

	nameNode.relayer = relayer

	return nil
}

func (rm *RelayManager) RemoveRelayer(trackPath string, relayer *Relayer) {
	trackParts := splitTrackPath(trackPath)

	// Trace the track path
	prefixNode, ok := rm.trackPathTree.traceTrackPrefix(trackParts)
	if !ok {
		return
	}

	// Find the track name node
	nameNode, ok := prefixNode.findTrackName(trackParts[len(trackParts)-1])
	if !ok {
		return
	}

	// Remove the track name node
	nameNode.mu.Lock()
	defer nameNode.mu.Unlock()

	nameNode.relayer = nil

	// Remove the track path if there is no track name node
	if len(prefixNode.trackNames) == 0 {
		rm.trackPathTree.removeTrackPrefix(trackParts)
	}
}

func (rm *RelayManager) GetRelayer(trackPath string) *Relayer {
	trackParts := splitTrackPath(trackPath)

	// Trace the track path
	prefixNode, ok := rm.trackPathTree.traceTrackPrefix(trackParts)
	if !ok {
		return nil
	}

	// Find the track name node
	nameNode, ok := prefixNode.findTrackName(trackParts[len(trackParts)-1])
	if !ok {
		return nil
	}

	nameNode.mu.RLock()
	defer nameNode.mu.RUnlock()

	return nameNode.relayer
}

func (rm *RelayManager) AddInterest(trackPrefix string, interest *ReceivedInterest) error {
	trackPrefixParts := splitTrackPath(trackPrefix)

	// Trace the track path
	prefixNode, ok := rm.trackPathTree.traceTrackPrefix(trackPrefixParts)
	if !ok || prefixNode == nil {
		return ErrTrackDoesNotExist
	}

	prefixNode.riMu.Lock()
	defer prefixNode.riMu.Unlock()

	prefixNode.interests = append(prefixNode.interests, interest)

	return nil
}

func (rm *RelayManager) RemoveInterest(trackPrefix string, interest *ReceivedInterest) {
	trackPrefixParts := splitTrackPath(trackPrefix)

	// Trace the track path
	prefixNode, ok := rm.trackPathTree.traceTrackPrefix(trackPrefixParts)
	if !ok || prefixNode == nil {
		return
	}

	prefixNode.riMu.Lock()
	defer prefixNode.riMu.Unlock()

	for i, ri := range prefixNode.interests {
		if ri == interest {
			prefixNode.interests = append(prefixNode.interests[:i], prefixNode.interests[i+1:]...)
			break
		}
	}
}

type trackPathTree struct {
	rootNode *trackPrefixNode
}

func (tree trackPathTree) insertTrackName(trackPrefixParts []string, trackName string) (*trackNameNode, error) {
	// Trace the track prefix
	prefixNode := tree.insertTrackPrefix(trackPrefixParts)

	// Find the track name node
	_, ok := prefixNode.findTrackName(trackName)
	if ok {
		return nil, ErrDuplicatedTrack
	}

	// Insert the track name node
	return prefixNode.insertTrackName(trackName)
}

func (tree trackPathTree) insertTrackPrefix(trackPrefixParts []string) *trackPrefixNode {
	// Set the current node to the root node
	currentNode := tree.rootNode
	//
	var exists bool

	// track the tree
	for _, trackPart := range trackPrefixParts {
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

	return currentNode
}

func (tree trackPathTree) removeTrackPrefix(trackPrefixParts []string) error {
	_, err := tree.rootNode.removeDescendants(trackPrefixParts, 0)
	return err
}

func (tree trackPathTree) traceTrackPrefix(trackPrefixParts []string) (*trackPrefixNode, bool) {
	return tree.rootNode.traceDescendant(trackPrefixParts)
}

type trackPrefixNode struct {

	/*
	 * A Part of the Track Prefix
	 */
	trackPrefixPart string

	/*
	 * Children of the node
	 */
	children map[string]*trackPrefixNode
	cdMu     sync.RWMutex

	/*
	 * Track Name Nodes
	 */
	trackNames map[string]*trackNameNode
	tnMu       sync.RWMutex

	//
	interests []*ReceivedInterest
	riMu      sync.RWMutex
}

type trackNameNode struct {
	trackNamePart string

	relayer *Relayer

	mu sync.RWMutex
}

func (node *trackPrefixNode) removeDescendants(tns []string, depth int) (bool, error) {
	if node == nil {
		return false, errors.New("track namespace not found at " + tns[depth])
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

	node.cdMu.RLock()
	defer node.cdMu.RUnlock()
	child, exists := node.children[value]

	if !exists {
		return false, errors.New("track namespace not found at " + value)
	}

	ok, err := child.removeDescendants(tns, depth+1)
	if err != nil {
		return false, err
	}

	if ok {
		delete(node.children, value)

		return (len(node.children) == 0), nil
	}

	return false, nil
}

func (node *trackPrefixNode) traceDescendant(values []string) (*trackPrefixNode, bool) {
	node.cdMu.RLock()
	defer node.cdMu.RUnlock()

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
	node.tnMu.RLock()
	defer node.tnMu.RUnlock()

	tnNode, ok := node.trackNames[trackName]
	if !ok {
		return nil, false
	}

	return tnNode, true
}

func (node *trackPrefixNode) insertTrackName(trackName string) (*trackNameNode, error) {
	node.tnMu.Lock()
	defer node.tnMu.Unlock()

	if _, ok := node.trackNames[trackName]; ok {
		return nil, ErrDuplicatedTrack
	}

	trackNameNode := &trackNameNode{
		trackNamePart: trackName,
	}

	node.trackNames[trackName] = trackNameNode

	return trackNameNode, nil
}

func splitTrackPath(trackPath string) []string {
	return strings.Split(trackPath, "/")
}
