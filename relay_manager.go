package moqt

import (
	"errors"
	"log/slog"
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

func (rm RelayManager) RegisterOrigin(origin *ServerSession, ann Announcement) {
	slog.Info("Registering an origin session")

	tnsNode := rm.newTrackNamespace(strings.Split(ann.TrackNamespace, "/"))

	if tnsNode.announcement != nil {
		slog.Info("updated an announcement", slog.Any("from", tnsNode.announcement), slog.Any("to", ann))
	}
	tnsNode.announcement = &ann

	if tnsNode.origin != nil {
		slog.Info("updated an origin session")
	}
	tnsNode.origin = origin
}

func (rm RelayManager) RegisterFollower(trackPrefix string, aw AnnounceWriter) {
	slog.Info("Registering a follower")

	tns := strings.Split(trackPrefix, "/")

	tnsNode := rm.newTrackNamespace(tns)

	if tnsNode.followers == nil {
		tnsNode.followers = make([]*AnnounceWriter, 1) // TODO: Tune the size
	}

	tnsNode.followers = append(tnsNode.followers, &aw)
}

func (rm RelayManager) RemoveAnnouncement(ann Announcement) {
	slog.Info("Remove an announcement")
	tns := strings.Split(ann.TrackNamespace, "/")

	err := rm.removeTrackNamespace(tns)
	if err != nil {
		slog.Error("failed to remove a Track Namespace", slog.String("error", err.Error()))
		return
	}
}

func (rm RelayManager) PublishAnnouncement(ann Announcement) {
	slog.Info("Publishing an announcement")

	tns := strings.Split(ann.TrackNamespace, "/")

	for i := range tns {
		tnsNode, ok := rm.findTrackNamespace(tns[:i])
		if !ok {
			break
		}

		for _, aw := range tnsNode.followers {
			aw.Announce(ann)
		}
	}
}

// func (rm RelayManager) RegisterTrack(subscription Subscription) {
// 	//TODO
// }

func (rm RelayManager) GetInfo(trackNamespace, trackName string) (Info, bool) {
	tns := strings.Split(trackNamespace, "/")
	tnsNode, ok := rm.findTrackNamespace(tns)
	if !ok {
		return Info{}, false
	}

	tnNode, ok := tnsNode.findTrackName(trackName)
	if !ok {
		return Info{}, false
	}

	return tnNode.info, true
}

func (rm RelayManager) RecordInfo(trackNamespace string, trackName string, info Info) error {
	slog.Info("Recording a track information")
	tns := strings.Split(trackNamespace, "/")

	tnsNode, ok := rm.findTrackNamespace(tns)
	if !ok {
		return errors.New("track namespace not found")
	}

	tnNode, ok := tnsNode.findTrackName(trackName)
	if !ok {
		return errors.New("track name not found")
	}

	tnNode.info = info

	return nil
}

func (rm RelayManager) newTrackNamespace(trackNamespace []string) *trackNamespaceNode {
	return rm.trackNamespaceTree.insert(trackNamespace)
}

func (rm RelayManager) findTrackNamespace(trackNamespace []string) (*trackNamespaceNode, bool) {
	return rm.trackNamespaceTree.trace(trackNamespace)
}

func (rm RelayManager) removeTrackNamespace(trackNamespace []string) error {
	return rm.trackNamespaceTree.remove(trackNamespace)
}

func (rm RelayManager) findDestinations(trackNamespace []string, trackName string, order GroupOrder) ([]*session, bool) {
	// Find the Track Namespace
	tnsNode, ok := rm.findTrackNamespace(trackNamespace)
	if !ok {
		return nil, false
	}

	// Find the Track Name
	tnNode, ok := tnsNode.findTrackName(trackName)
	if !ok {
		return nil, false
	}

	// Verify the Group Order of the track
	if tnNode.groupOrder != order {
		return nil, false
	}

	return tnNode.destinations, true
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
	origin *ServerSession

	/*
	 * Announcement
	 */
	announcement *Announcement

	/*
	 * Announce Streams of followers to the Track Namespace
	 */
	followers []*AnnounceWriter
}

type trackNameNode struct {
	mu sync.RWMutex

	/*
	 * The string value of the Track Name
	 */
	value string

	/*
	 * The Group's order
	 */
	groupOrder GroupOrder

	/*
	 * The destination session
	 */
	destinations []*session

	/*
	 * Information of the Track
	 */
	info Info
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

func (node *trackNamespaceNode) findTrackName(trackName string) (*trackNameNode, bool) {
	node.mu.RLock()
	defer node.mu.RUnlock()

	tnNode, ok := node.tracks[trackName]
	if !ok {
		return nil, false
	}

	return tnNode, true
}

/*
 * Create a new Track Name node when a subscriber makes a new subscription
 *
 */
func (tnsNode *trackNamespaceNode) newTrackName(trackName string) *trackNameNode {
	tnsNode.mu.Lock()
	defer tnsNode.mu.Unlock()

	tnsNode.tracks[trackName] = &trackNameNode{
		value: trackName,
	}

	return tnsNode.tracks[trackName]
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
