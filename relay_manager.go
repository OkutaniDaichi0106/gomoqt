package moqt

import (
	"errors"
	"strings"
	"sync"
)

func NewRelayManager() *RelayManager {
	return &RelayManager{
		trackNamespaceTree: trackNamespaceTree{
			rootNode: &trackNamespaceNode{
				value:    "",
				children: make(map[string]*trackNamespaceNode),
			},
		},
	}
}

type RelayManager struct {
	trackNamespaceTree trackNamespaceTree
}

// func (rm RelayManager) GetAnnouncement(trackNamespace string) (Announcement, bool) {
// 	tns := strings.Split(trackNamespace, "/")
// 	tnsNode, ok := rm.findTrackNamespace(tns)
// 	if !ok {
// 		return Announcement{}, false
// 	}

// 	return *tnsNode.announcement, true
// }

func (rm RelayManager) GetAnnouncements(trackNamespacePrefix string) ([]Announcement, bool) {
	tns := strings.Split(trackNamespacePrefix, "/")
	tnsNode, ok := rm.findTrackNamespace(tns)
	if !ok {
		return nil, false
	}

	announcements := tnsNode.getAnnouncements()
	if announcements == nil {
		return nil, false
	}

	return announcements, true
}

func (rm RelayManager) GetInfo(trackNamespace, trackName string) (Info, bool) {
	tns := strings.Split(trackNamespace, "/")
	tnsNode, ok := rm.findTrackNamespace(tns)
	if !ok {
		return Info{}, false
	}

	tnNode, ok := tnsNode.findTrackNameNode(trackName)
	if !ok {
		return Info{}, false
	}

	return tnNode.info, true
}

func (tm RelayManager) newTrackNamespace(trackNamespace []string) *trackNamespaceNode {
	return tm.trackNamespaceTree.insert(trackNamespace)
}

func (tm RelayManager) findTrackNamespace(trackNamespace []string) (*trackNamespaceNode, bool) {
	return tm.trackNamespaceTree.trace(trackNamespace)
}

func (rm RelayManager) removeTrackNamespace(trackNamespace []string) error {
	return rm.trackNamespaceTree.remove(trackNamespace)
}

func (rm RelayManager) findDestinations(trackNamespace []string, trackName string, order GroupOrder) ([]*session, bool) {
	tnsNode, ok := rm.findTrackNamespace(trackNamespace)
	if !ok {
		return nil, false
	}

	tnNode, ok := tnsNode.findTrackNameNode(trackName)
	if !ok {
		return nil, false
	}

	goNode, ok := tnNode.findGroupOrder(order)
	if !ok {
		return nil, false
	}

	return goNode.destinations, true
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
	origin *session

	/*
	 * Announcement
	 */
	announcement *Announcement
}

type trackNameNode struct {
	mu sync.RWMutex

	/*
	 * The string value of the Track Name
	 */
	value string

	orders map[GroupOrder]*groupOrderNode

	/*
	 * Information of the Track
	 */
	info Info
}

type groupOrderNode struct {
	mu sync.RWMutex

	/*
	 * The Group's order
	 */
	groupOrder GroupOrder

	/*
	 * The destination session
	 */
	destinations []*session
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

func (node *trackNamespaceNode) getAnnouncements() []Announcement {
	var announcements []Announcement
	for _, childNode := range node.children {
		if childNode == nil {
			continue
		}
		announcements = append(announcements, childNode.getAnnouncements()...)
	}

	if node.announcement != nil {
		announcements = append(announcements, *node.announcement)
	}

	return announcements
}

func (node *trackNameNode) findGroupOrder(order GroupOrder) (*groupOrderNode, bool) {
	goNode, ok := node.orders[order]
	if !ok {
		return nil, false
	}

	return goNode, true
}