package moqt

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type publisher interface {
	//NewTrack(Announcement) Track
	Announce(Announcement)
	Unannounce(Announcement)
	OpenDataStream(Track, Group) (moq.SendStream, error)
}

var _ publisher = (*Publisher)(nil)

type Publisher struct {
	sess *session

	publisherManager *publisherManager
}

func (p *Publisher) Announce(ann Announcement) {
	p.publisherManager.publishAnnouncement(ann)
}

func (p *Publisher) Unannounce(ann Announcement) {
	p.publisherManager.cancelAnnouncement(ann)
}

func (p *Publisher) OpenDataStream(t Track, g Group) (moq.SendStream, error) {
	/*
	 *
	 */
	// Verify the group is a new one in the track
	_, ok := t.groups[g.groupSequence]
	if ok {
		return nil, errors.New("duplicated group")
	}

	//TODO: Verify the Track was subscribed
	// p.publisherManager

	return p.openDataStream(g)
}

func (p *Publisher) openGroupStream() (moq.SendStream, error) {
	slog.Debug("opening an Group Stream")

	stream, err := p.sess.conn.OpenUniStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_group,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func (p *Publisher) openDataStream(g Group) (moq.SendStream, error) {
	if g.groupSequence == 0 {
		return nil, errors.New("0 sequence number")
	}

	stream, err := p.openGroupStream()
	if err != nil {
		slog.Error("failed to open a group stream", slog.String("error", err.Error()))
		return nil, err
	}

	gm := message.GroupMessage{
		SubscribeID:       message.SubscribeID(g.subscribeID),
		GroupSequence:     message.GroupSequence(g.groupSequence),
		PublisherPriority: message.Priority(g.PublisherPriority),
	}

	// Send the GROUP message
	err = gm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func (p *Publisher) sendDatagram(g Group, payload []byte) error {
	if g.groupSequence == 0 {
		return errors.New("0 sequence number")
	}

	gm := message.GroupMessage{
		SubscribeID:       message.SubscribeID(g.subscribeID),
		GroupSequence:     message.GroupSequence(g.groupSequence),
		PublisherPriority: message.Priority(g.PublisherPriority),
	}

	var buf bytes.Buffer

	// Encode the GROUP message
	err := gm.Encode(&buf)
	if err != nil {
		slog.Error("failed to encode a GROUP message", slog.String("error", err.Error()))
		return err
	}

	// Encode the payload
	_, err = buf.Write(payload)
	if err != nil {
		slog.Error("failed to encode a payload", slog.String("error", err.Error()))
		return err
	}

	// Send the data with the GROUP message
	err = p.sess.conn.SendDatagram(buf.Bytes())
	if err != nil {
		slog.Error("failed to send a datagram", slog.String("error", err.Error()))
		return err
	}

	return nil
}

/*
 *
 */
func newPublishManager() *publisherManager {
	return &publisherManager{
		activeAnnouncements:     make(map[string]Announcement),
		interestReceivedStreams: make(map[string]*interestReceivedStream),
		receivedSubscription:    make(map[SubscribeID]*subscribeReceiveStream),
	}
}

type publisherManager struct {
	/*
	 * Active announcements
	 * Track Path -> Announcement
	 */
	activeAnnouncements map[string]Announcement

	/*
	 * Sent announcements
	 * Track Prefix -> interestReceivedStream
	 */
	interestReceivedStreams map[string]*interestReceivedStream
	irsMu                   sync.RWMutex

	/*
	 * Received Subscriptions
	 */
	receivedSubscription map[SubscribeID]*subscribeReceiveStream
	rsMu                 sync.RWMutex
}

func (pm *publisherManager) publishAnnouncement(announcement Announcement) {
	pm.irsMu.RLock()
	defer pm.irsMu.RUnlock()

	if _, ok := pm.activeAnnouncements[announcement.TrackPath]; ok {
		return
	}

	for prefix, sentAnnouncements := range pm.interestReceivedStreams {
		// Skip to announce
		if !strings.HasPrefix(announcement.TrackPath, prefix) {
			continue
		}

		err := sentAnnouncements.activateAnnouncement(announcement)
		if err != nil {
			slog.Error("failed to active an announcement", slog.String("error", err.Error()))
			continue
		}
	}

	pm.activeAnnouncements[announcement.TrackPath] = announcement
}

/****/
func (pm *publisherManager) cancelAnnouncement(announcement Announcement) {
	pm.irsMu.RLock()
	defer pm.irsMu.RUnlock()

	if _, ok := pm.activeAnnouncements[announcement.TrackPath]; !ok {
		return
	}

	for prefix, sentAnnouncements := range pm.interestReceivedStreams {
		// Skip to announce
		if !strings.HasPrefix(announcement.TrackPath, prefix) {
			continue
		}

		err := sentAnnouncements.endAnnouncement(announcement)
		if err != nil {
			slog.Error("failed to active an announcement", slog.String("error", err.Error()))
			continue
		}
	}

	delete(pm.activeAnnouncements, announcement.TrackPath)
}

func (pm *publisherManager) addInterestReceivedStream(irs *interestReceivedStream) error {
	pm.irsMu.Lock()
	defer pm.irsMu.Unlock()

	_, ok := pm.interestReceivedStreams[irs.interest.TrackPrefix]
	if ok {
		return ErrDuplicatedInterest
	}

	pm.interestReceivedStreams[irs.interest.TrackPrefix] = irs

	for _, announcement := range pm.activeAnnouncements {
		err := irs.activateAnnouncement(announcement)
		if err != nil {
			slog.Error("failed to activate an announcement")
			return err
		}
	}

	return nil
}

func (pm *publisherManager) removeInterestReceiveStream(irs *interestReceivedStream) {
	pm.irsMu.Lock()
	defer pm.irsMu.Unlock()

	delete(pm.interestReceivedStreams, irs.interest.TrackPrefix)
}

func (pm *publisherManager) addSubscribeSendStream(srs *subscribeReceiveStream) error {
	pm.rsMu.Lock()
	defer pm.rsMu.Unlock()

	_, ok := pm.receivedSubscription[srs.subscription.subscribeID]
	if ok {
		return ErrDuplicatedSubscribeID
	}

	pm.receivedSubscription[srs.subscription.subscribeID] = srs

	return nil
}

func (pm *publisherManager) removeSubscribeSendStream(id SubscribeID) {
	pm.rsMu.Lock()
	defer pm.rsMu.Unlock()

	delete(pm.receivedSubscription, id)
}

type interestReceivedStream struct {
	interest Interest
	/*
	 * Sent announcements
	 * Track Path -> Announcement
	 */
	announcements map[string]Announcement
	stream        moq.Stream
	mu            sync.RWMutex
}

func newInterestReceiveStream(interest Interest, stream moq.Stream) *interestReceivedStream {
	return &interestReceivedStream{
		interest:      interest,
		announcements: make(map[string]Announcement),
		stream:        stream,
	}
}

func (sas *interestReceivedStream) activateAnnouncement(announcement Announcement) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	// Verify if the announcement has the track prefix
	if !strings.HasPrefix(announcement.TrackPath, sas.interest.TrackPrefix) {
		return ErrInternalError
	}

	// Verify if the Track Path has been already announced
	_, ok := sas.announcements[announcement.TrackPath]
	if ok {
		return ErrDuplicatedTrackPath
	}

	// Get a suffix part of the Track Path
	suffix := strings.TrimPrefix(announcement.TrackPath, sas.interest.TrackPrefix+"/")

	//
	if announcement.AuthorizationInfo != "" {
		announcement.Parameters.Add(AUTHORIZATION_INFO, announcement.AuthorizationInfo)
	}

	// Send
	am := message.AnnounceMessage{
		AnnounceStatus:  message.ACTIVE,
		TrackPathSuffix: suffix,
		Parameters:      message.Parameters(announcement.Parameters),
	}
	err := am.Encode(sas.stream)
	if err != nil {
		return err
	}

	// Register
	sas.announcements[announcement.TrackPath] = announcement

	return nil
}

func (sas *interestReceivedStream) endAnnouncement(announcement Announcement) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	// Verify if the announcement has the track prefix
	if !strings.HasPrefix(announcement.TrackPath, sas.interest.TrackPrefix) {
		return ErrInternalError
	}

	// Verify if the Track Path has been already announced
	_, ok := sas.announcements[announcement.TrackPath]
	if !ok {
		return ErrTrackDoesNotExist
	}

	// Get a suffix part of the Track Path
	suffix := strings.TrimPrefix(announcement.TrackPath, sas.interest.TrackPrefix+"/")

	//
	if announcement.AuthorizationInfo != "" {
		announcement.Parameters.Add(AUTHORIZATION_INFO, announcement.AuthorizationInfo)
	}

	// Send
	am := message.AnnounceMessage{
		AnnounceStatus:  message.ENDED,
		TrackPathSuffix: suffix,
		Parameters:      message.Parameters(announcement.Parameters),
	}
	err := am.Encode(sas.stream)
	if err != nil {
		return err
	}

	// Remove
	delete(sas.announcements, announcement.TrackPath)

	return nil
}
