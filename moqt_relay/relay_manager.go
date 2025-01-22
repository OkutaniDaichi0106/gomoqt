package moqtrelay

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

type RelayManager interface {
	RelayAnnouncements(moqt.Session, moqt.AnnounceConfig) error

	/*
	 * Serve subscription to the relay manager
	 */
	RelayTrack(moqt.Session, moqt.SubscribeConfig) error

	TrackManager
}

func NewRelayManager() RelayManager {
	return &relayManager{
		trackTree: newTrackTree(),
	}
}

var _ RelayManager = (*relayManager)(nil)

type relayManager struct {
	//
	trackTree *trackTree

	TrackManager
}

func (manager *relayManager) RelayAnnouncements(sess moqt.ServerSession, interest moqt.AnnounceConfig) error {
	annstr, err := sess.OpenAnnounceStream(interest)
	if err != nil {
		return err
	}

	go func() {
		for {
			// Receive announcements
			ann, err := annstr.ReceiveAnnouncements()
			if err != nil {
				slog.Error("failed to receive announcements", slog.String("error", err.Error()))
				return
			}

			// Serve announcements
			err = manager.ServeAnnouncements(ann)
			if err != nil {
				slog.Error("failed to serve announcements", slog.String("error", err.Error()))
				return
			}
		}
	}()

	return nil
}

func (manager *relayManager) RelayTrack(sess moqt.Session, sub moqt.SubscribeConfig) error {
	substr, err := sess.OpenSubscribeStream(sub)
	if err != nil {
		slog.Error("failed to open a subscribe stream", slog.String("error", err.Error()))
		return err
	}

	// Create a track buffer
	trackBuf := NewTrackBuffer(sub)

	// Serve the subscription
	err = manager.ServeTrack(sub, trackBuf)
	if err != nil {
		slog.Error("failed to serve the subscription", slog.String("error", err.Error()))
		return err
	}

	go func() {
		for {
			// Receive data
			stream, err := sess.AcceptDataStream(substr, context.Background())
			if err != nil {
				slog.Error("failed to receive data", slog.String("error", err.Error()))
				return
			}

			// Initialize a group buffer
			groupBuf := NewGroupBuffer(stream.GroupSequence(), stream.GroupPriority())

			// Add the group buffer to the track buffer
			trackBuf.AddGroup(groupBuf)

			// Receive data from the stream
			go func(stream moqt.ReceiveGroupStream) {
				/*
				 * Receive data from the stream
				 */
				for {
					// Receive the next frame
					frame, err := stream.NextFrame()
					// Store the frame to the group buffer
					if len(frame) > 0 {
						groupBuf.Write(frame)
					}

					if err != nil {
						if err == io.EOF {
							groupBuf.Close()
						} else {
							slog.Error("failed to receive a frame", slog.String("error", err.Error()))
						}
						return
					}
				}
			}(stream)
		}
	}()

	return nil
}

func newTrackTree() *trackTree {
	return &trackTree{
		rootNode: &trackPrefixNode{
			trackPrefix: "",
			children:    make(map[string]*trackPrefixNode),
		},
	}
}

type trackTree struct {
	rootNode *trackPrefixNode
}

func (tree trackTree) insertTrackPrefix(trackPrefixParts []string) *trackPrefixNode {
	return tree.rootNode.addDescendants(trackPrefixParts)
}

func (tree trackTree) removeTrackPrefix(trackPrefixParts []string) error {
	_, err := tree.rootNode.removeDescendants(trackPrefixParts)
	return err
}

func (tree trackTree) traceTrackPrefix(trackPrefixParts []string) (*trackPrefixNode, bool) {
	return tree.rootNode.traceDescendant(trackPrefixParts)
}

func (tree trackTree) handleDescendants(trackPrefixParts []string, op func(*trackPrefixNode) error) error {
	return tree.rootNode.handleDescendants(trackPrefixParts, op)
}

func (trackPrefixNode *trackPrefixNode) handleDescendants(trackPrefixParts []string, op func(*trackPrefixNode) error) error {
	err := op(trackPrefixNode)
	if err != nil {
		return err
	}

	if len(trackPrefixParts) == 0 {
		return nil
	}

	node := trackPrefixNode.addDescendants(trackPrefixParts[0:1])

	return node.handleDescendants(trackPrefixParts[1:], op)
}

type trackPrefixNode struct {
	trackPrefix string

	/*
	 * Children of the node
	 */
	children map[string]*trackPrefixNode
	cdMu     sync.RWMutex

	/*
	 * Announcer
	 */
	announcementBuffer announcementBuffer

	/*
	 * Track node
	 */
	track *trackNameNode
}

func (node *trackPrefixNode) addDescendants(trackPrefixParts []string) *trackPrefixNode {
	if node == nil {
		return nil
	}

	if len(trackPrefixParts) == 0 {
		return node
	}

	trackPrefixPart := trackPrefixParts[0]

	node.cdMu.RLock()
	defer node.cdMu.RUnlock()

	child, exists := node.children[trackPrefixPart]
	if !exists {
		child = &trackPrefixNode{
			trackPrefix:        node.trackPrefix + "/" + trackPrefixPart,
			children:           make(map[string]*trackPrefixNode),
			announcementBuffer: newAnnouncementBuffer(),
		}
		node.children[trackPrefixPart] = child
	}

	return child.addDescendants(trackPrefixParts[1:])
}

func (node *trackPrefixNode) removeDescendants(trackPrefixParts []string) (bool, error) {
	if node == nil {
		return false, errors.New("track namespace not found at " + trackPrefixParts[0])
	}

	if len(trackPrefixParts) == 0 {
		if len(node.children) == 0 {
			return true, nil
		}

		return false, nil
	}

	value := trackPrefixParts[0]

	node.cdMu.RLock()
	defer node.cdMu.RUnlock()
	child, exists := node.children[value]

	if !exists {
		return false, errors.New("track namespace not found at " + value)
	}

	ok, err := child.removeDescendants(trackPrefixParts[1:])
	if err != nil {
		return false, err
	}

	if ok {
		delete(node.children, value)

		return (len(node.children) == 0), nil
	}

	return false, nil
}

func (node *trackPrefixNode) traceDescendant(trackPrefixParts []string) (*trackPrefixNode, bool) {
	if len(trackPrefixParts) == 0 {
		return node, true
	}

	node.cdMu.RLock()
	defer node.cdMu.RUnlock()

	child, exists := node.children[trackPrefixParts[0]]
	if !exists || child == nil {
		return nil, false
	}

	return child.traceDescendant(trackPrefixParts[1:])
}

func (node *trackPrefixNode) initTrack() {
	node.cdMu.Lock()
	defer node.cdMu.Unlock()

	if node.track == nil {
		node.track = &trackNameNode{
			trackPath: node.trackPrefix,
		}
	}
}

/*
 * Announcer
 */
func newAnnouncementBuffer() announcementBuffer {
	return announcementBuffer{
		annCond:       sync.NewCond(&sync.Mutex{}),
		announcements: make([]moqt.Announcement, 0),
	}
}

type announcementBuffer struct {
	annCond *sync.Cond

	announcements []moqt.Announcement

	waiting uint64

	completed uint64
}

func (announcer announcementBuffer) WaitAnnouncements() []moqt.Announcement {
	announcer.annCond.L.Lock()
	defer announcer.annCond.L.Unlock()

	atomic.AddUint64(&announcer.waiting, 1)

	for len(announcer.announcements) == 0 {
		announcer.annCond.Wait()
	}

	// Clean the buffer if all the waiting clients have received the announcements
	if atomic.AddUint64(&announcer.completed, 1) == atomic.LoadUint64(&announcer.waiting) {
		announcer.Clean()
		atomic.StoreUint64(&announcer.waiting, 0)
		atomic.StoreUint64(&announcer.completed, 0)
	}

	return announcer.announcements
}

func (announcer announcementBuffer) Broadcast() {
	announcer.annCond.Broadcast()

}

func (announcer *announcementBuffer) Add(ann moqt.Announcement) {
	announcer.annCond.L.Lock()
	defer announcer.annCond.L.Unlock()

	announcer.announcements = append(announcer.announcements, ann)
}

func (announcer *announcementBuffer) Clean() {
	announcer.annCond.L.Lock()
	defer announcer.annCond.L.Unlock()

	announcer.announcements = make([]moqt.Announcement, 0)
}

// func (announcer announcementBuffer) WaitAndGet() []Announcement {
// 	announcer.Wait()
// 	return announcer.Get()
// }

type trackNameNode struct {
	trackPath string

	/*
	 * Session serving the track
	 */
	sess   moqt.ServerSession
	sessMu sync.Mutex

	/*
	 * Announcement received from the session
	 */
	announcement moqt.Announcement

	/*
	 * moqtransfork.Subscription sent to the session
	 */
	subscription moqt.SubscribeConfig

	/*
	 * Frame queue
	 */
	trackBuf *TrackBuffer
	mu       sync.Mutex
}

func (node *trackNameNode) SetSession(sess moqt.ServerSession) error {
	node.sessMu.Lock()
	defer node.sessMu.Unlock()

	if node.sess != nil && node.sess != sess {
		slog.Debug("the session serving the track has been changed to another session", slog.String("track path", node.trackPath))
	}

	node.sess = sess

	return nil
}
