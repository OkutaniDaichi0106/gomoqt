package moqt

import (
	"bytes"
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type Track struct {
	Announcement
}

type Publisher interface {
	NewTrack(Announcement) Track
	OpenDataStream(Track, Group) (moq.SendStream, error)
}

var _ Publisher = (*publisher)(nil)

type publisher struct {
	sess    *session
	manager *publisherManager
}

func (p *publisher) NewTrack(ann Announcement) Track {
	return Track{
		Announcement: ann,
	}
}

func (p *publisher) OpenDataStream(t Track, g Group) (moq.SendStream, error) {
	//TODO: Verify the Track was subscribed
	return p.openDataStream(g)
}

func (p *publisher) openGroupStream() (moq.SendStream, error) {
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

func (p *publisher) openDataStream(g Group) (moq.SendStream, error) {
	if g.groupSequence == 0 {
		return nil, errors.New("0 sequence number")
	}

	stream, err := p.openGroupStream()
	if err != nil {
		slog.Error("failed to open an unidirectional Stream", slog.String("error", err.Error()))
		return nil, err
	}

	gm := message.GroupMessage{
		SubscribeID:       message.SubscribeID(g.subscribeID),
		GroupSequence:     message.GroupSequence(g.groupSequence),
		PublisherPriority: message.PublisherPriority(g.PublisherPriority),
	}

	// Send the GROUP message
	err = gm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func (p *publisher) sendDatagram(g Group, payload []byte) error {
	if g.groupSequence == 0 {
		return errors.New("0 sequence number")
	}

	gm := message.GroupMessage{
		SubscribeID:       message.SubscribeID(g.subscribeID),
		GroupSequence:     message.GroupSequence(g.groupSequence),
		PublisherPriority: message.PublisherPriority(g.PublisherPriority),
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
func newPublisherManager() *publisherManager {
	return &publisherManager{
		tracks:             make(map[string]Track),
		subscribeReceivers: make(map[SubscribeID]*SubscribeReceiver),
	}
}

type publisherManager struct {
	/*
	 * Announced Tracks
	 */
	tracks map[string]Track

	/*
	 * Received Subscriptions
	 */
	subscribeReceivers map[SubscribeID]*SubscribeReceiver
	mu                 sync.RWMutex
}

func (pm *publisherManager) addSubscribeReceiver(sr *SubscribeReceiver) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	_, ok := pm.subscribeReceivers[sr.subscription.subscribeID]
	if ok {
		return ErrDuplicatedSubscribeID
	}

	pm.subscribeReceivers[sr.subscription.subscribeID] = sr

	return nil
}

func (pm *publisherManager) removeSubscriberSender(id SubscribeID) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delete(pm.subscribeReceivers, id)
}
