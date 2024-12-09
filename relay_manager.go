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

func (rm RelayManager) addRelayer(relayer *Relayer) error {
	trackNameNode, ok := rm.newTrackPath(relayer.TrackPath)
	if ok {
		return errors.New("duplicated relayer")
	}

	trackNameNode.mu.Lock()
	defer trackNameNode.mu.Unlock()

	trackNameNode.relayer = relayer
}

// func (rm RelayManager) GetAnnouncements(trackPathPrefix string) []Announcement {
// 	tp := strings.Split(trackPathPrefix, "/")
// 	tnsNode, ok := rm.findTrackNamespace(tp[:len(tp)-1])
// 	if !ok {
// 		return nil

// 	}
// 	// Get any Announcements under the Track Namespace
// 	return tnsNode.getAnnouncements()
// }

// // TODO:
// func (rm RelayManager) registerOrigin(origin *ServerSession, ann Announcement) {
// 	slog.Info("Registering an origin session")

// 	tnsNode := rm.newTrackNamespace(strings.Split(ann.TrackPath, "/"))

// 	if tnsNode.announcement != nil {
// 		slog.Info("updated an announcement", slog.Any("from", tnsNode.announcement), slog.Any("to", ann))
// 	}
// 	tnsNode.announcement = &ann

// 	if tnsNode.origin != nil {
// 		slog.Info("updated an origin session")
// 	}
// 	tnsNode.origin = origin
// }

// func (rm RelayManager) registerFollower(trackPrefix string, annCh chan Announcement) {
// 	slog.Info("Registering a follower")

// 	tns := strings.Split(trackPrefix, "/")

// 	tnsNode := rm.newTrackNamespace(tns)

// 	if tnsNode.followers == nil {
// 		tnsNode.followers = make([]chan Announcement, 1) // TODO: Tune the size
// 	}

// 	tnsNode.followers = append(tnsNode.followers, annCh)
// }

// func (rm RelayManager) RemoveAnnouncement(ann Announcement) {
// 	slog.Info("Remove an announcement")
// 	tns := strings.Split(ann.TrackPath, "/")

// 	err := rm.removeTrackNamespace(tns)
// 	if err != nil {
// 		slog.Error("failed to remove a Track Namespace", slog.String("error", err.Error()))
// 		return
// 	}
// }

// func (rm RelayManager) PublishAnnouncement(ann Announcement) {
// 	slog.Info("Publishing an announcement")

// 	tns := strings.Split(ann.TrackPath, "/")

// 	for i := range tns {
// 		tnsNode, ok := rm.findTrackNamespace(tns[:i])
// 		if !ok {
// 			break
// 		}

// 		for _, annCh := range tnsNode.followers {
// 			annCh <- ann
// 		}
// 	}
// }

// func (rm RelayManager) GetInfo(trackPath string) (Info, bool) {
// 	tp := strings.Split(trackPath, "/")
// 	tnsNode, ok := rm.findTrackNamespace(tp[:len(tp)-1])
// 	if !ok {
// 		return Info{}, false
// 	}

// 	tnNode, ok := tnsNode.findTrackName(tp[len(tp)-1])
// 	if !ok {
// 		return Info{}, false
// 	}

// 	return tnNode.info, true
// }

// TODO
// func (rm RelayManager) recordInfo(trackPath string, info Info) error {
// 	slog.Info("Recording a track information")
// 	tp := strings.Split(trackPath, "/")

// 	tnsNode, ok := rm.findTrackNamespace(tp[:len(tp)-1])
// 	if !ok {
// 		return errors.New("track namespace not found")
// 	}

// 	tnNode, ok := tnsNode.findTrackName(tp[len(tp)-1])
// 	if !ok {
// 		return errors.New("track name not found")
// 	}

// 	return nil
// }

func (rm RelayManager) newTrackPath(trackPath string) (*trackNameNode, bool)

func (rm RelayManager) newTrackPrefix(trackPrefix string) (*trackPrefixNode, bool) {
	return rm.trackPathTree.insert(strings.Split(trackPrefix, "/"))
}

func (rm RelayManager) findTrackNamespace(trackNamespace []string) (*trackPrefixNode, bool) {
	return rm.trackPathTree.trace(trackNamespace)
}

func (rm RelayManager) removeTrackPrefix(trackNamespace []string) error {
	return rm.trackPathTree.remove(trackNamespace)
}

// func (rm RelayManager) findDestinations(trackNamespace []string, trackName string, order GroupOrder) ([]*session, bool) {
// 	// Find the Track Namespace
// 	tnsNode, ok := rm.findTrackNamespace(trackNamespace)
// 	if !ok {
// 		return nil, false
// 	}

// 	// Find the Track Name
// 	tnNode, ok := tnsNode.findTrackName(trackName)
// 	if !ok {
// 		return nil, false
// 	}

// 	return tnNode.subscribers, true
// }

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
	tracks map[string]*trackNameNode

	// /*
	//  * Announce Streams of followers to the Track Namespace
	//  */
	// followers []chan Announcement
}

type trackNameNode struct {
	mu sync.RWMutex

	trackNamePart string

	relayer Relayer
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
func (tnsNode *trackPrefixNode) newTrackName(trackName string) *trackNameNode {
	tnsNode.mu.Lock()
	defer tnsNode.mu.Unlock()

	tnsNode.tracks[trackName] = &trackNameNode{
		value: trackName,
	}

	return tnsNode.tracks[trackName]
}

func (node *trackPrefixNode) getAnnouncements() []Announcement {
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
