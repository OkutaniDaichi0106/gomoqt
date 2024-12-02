package moqt

// import (
// 	"log/slog"
// 	"sync"
// )

// type TrackManager struct {
// 	mu            sync.RWMutex
// 	announcements map[string]*announcementNode
// }

// func (tm *TrackManager) NewTrack(a Announcement, s Subscription) {
// 	tm.mu.Lock()
// 	defer tm.mu.Unlock()

// 	if tm.announcements == nil {
// 		tm.announcements = make(map[string]*announcementNode)
// 	}

// 	aNode, ok := tm.announcements[a.TrackNamespace]
// 	if !ok {
// 		tm.announcements[a.TrackNamespace] = &announcementNode{
// 			announcement:  a,
// 			subscriptions: make(map[string]*subscriptionNode),
// 		}
// 	}

// 	_, ok = aNode.subscriptions[s.TrackName]
// 	if ok {
// 		slog.Error("duplicated track")
// 		return
// 	}
// 	aNode.subscriptions[s.TrackName] = &subscriptionNode{
// 		subscription: s,
// 		info:         Info{},
// 	}
// }

// func (tm *TrackManager) UpdateInfo(s Subscription, i Info) error {
// 	aNode, ok := tm.announcements[s.TrackNamespace]
// 	if !ok {
// 		return ErrTrackDoesNotExist
// 	}

// 	sNode, ok := aNode.subscriptions[s.TrackName]
// 	if !ok {
// 		return ErrTrackDoesNotExist
// 	}

// 	sNode.info = i

// 	return nil
// }

// func (tm *TrackManager) GetAnnouncement(trackNamespace string) (Announcement, bool) {
// 	tm.mu.RLock()
// 	defer tm.mu.RUnlock()

// 	aNode, ok := tm.announcements[trackNamespace]
// 	if !ok {
// 		return Announcement{}, false
// 	}

// 	return aNode.announcement, true
// }

// func (tm *TrackManager) GetSubscription(trackNamespace, trackName string) (Subscription, bool) {
// 	tm.mu.RLock()
// 	defer tm.mu.RUnlock()

// 	aNode, ok := tm.announcements[trackNamespace]
// 	if !ok {
// 		return Subscription{}, false
// 	}

// 	sNode, ok := aNode.subscriptions[trackName]
// 	if !ok {
// 		return Subscription{}, false
// 	}

// 	return sNode.subscription, true
// }

// func (tm *TrackManager) GetInfo(trackNamespace, trackName string) (Info, bool) {
// 	tm.mu.RLock()
// 	defer tm.mu.RUnlock()

// 	aNode, ok := tm.announcements[trackNamespace]
// 	if !ok {
// 		return Info{}, false
// 	}

// 	sNode, ok := aNode.subscriptions[trackName]
// 	if !ok {
// 		return Info{}, false
// 	}

// 	return sNode.info, true
// }

// type announcementNode struct {
// 	mu sync.RWMutex

// 	/*
// 	 *
// 	 */
// 	announcement Announcement

// 	/*
// 	 *
// 	 */
// 	subscriptions map[string]*subscriptionNode
// }

// func (aNode *announcementNode) addSubscription(sub Subscription) {
// 	aNode.mu.Lock()
// 	defer aNode.mu.Unlock()
// }

// type subscriptionNode struct {
// 	/***/
// 	subscription Subscription

// 	/***/
// 	info Info
// }
