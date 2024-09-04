package gomoq

import (
	"context"
	"errors"
	"io"
)

type Subscriber struct {
	/*
	 * Client
	 * Subscriber is an extention of Client
	 */
	Client

	/***/
	SubscriberHandler

	/*
	 * Map of the Track Alias
	 * The key is the Track Full Name
	 */
	trackAliases map[string]TrackAlias

	/*
	 * The number of the subscriptions
	 * The index is
	 */
	subscriptions []SubscribeMessage
}

type SubscriberHandler interface {
	SubscribeParameters() Parameters
	SubscribeUpdateParameters() Parameters
}

var _ SubscriberHandler = Subscriber{}

func (s *Subscriber) Connect(url string) error {
	// Check if the Client specify the Versions
	if len(s.Versions) < 1 {
		panic("no versions are specified")
	}

	return s.connect(url, SUB)
}

type SubscribeConfig struct {
	/*
	 * 0 is set by default
	 */
	SubscriberPriority

	/*
	 * NOT_SPECIFY (= 0) is set by default
	 * If not specifyed, the value is set to 0 which means NOT_SPECIFY
	 */
	GroupOrder

	/*
	 * No value is set by default
	 */
	SubscriptionFilter
}

func (s *Subscriber) Subscribe(trackNamespace, trackName string, config SubscribeConfig) error {
	// Check Subscribe Configuration
	// Check if the Track Namespace is empty
	if len(trackNamespace) == 0 {
		return errors.New("no track namespace is specifyed")
	}
	// Check if the Track Name is empty
	if len(trackName) == 0 {
		return errors.New("no track name is specifyed")
	}
	// Check if the Filter is valid
	if config.SubscriptionFilter.isOK() {
		return errors.New("invalid filter type")
	}
	// Check if the track is already subscribed
	// and add track alias
	_, ok := s.trackAliases[trackNamespace+trackName]
	if !ok {
		s.trackAliases[trackNamespace+trackName] = TrackAlias(len(s.trackAliases))
	}

	return s.sendSubscribeMessage(trackNamespace, trackName, config)
}

/*****/
func (s *Subscriber) sendSubscribeMessage(trackNamespace, trackName string, config SubscribeConfig) error {

	sm := SubscribeMessage{
		SubscribeID:        SubscribeID(len(s.subscriptions)),
		TrackAlias:         s.trackAliases[trackNamespace+trackName],
		TrackNamespace:     trackNamespace,
		TrackName:          trackName,
		SubscriberPriority: config.SubscriberPriority,
		GroupOrder:         config.GroupOrder,
		SubscriptionFilter: config.SubscriptionFilter,
		Parameters:         s.SubscribeParameters(),
	}

	// Send SUBSCRIBE message
	_, err := s.controlStream.Write(sm.serialize())
	if err != nil {
		return err
	}

	s.subscriptions = append(s.subscriptions, sm)

	return nil
}

func (s *Subscriber) AcceptObjects(ctx context.Context) (<-chan []byte, <-chan error) {
	dataCh := make(chan []byte, 1<<8) // TODO: Tune the size
	errCh := make(chan error, 1)      // TODO: Consider the buffer size

	buf := make([]byte, 1<<8)
	// Receive data on a stream in a goroutine
	go func() {
		// Close the data channel and the error channel when
		defer close(dataCh)
		defer close(errCh)

		for {
			// Catch the cancel call
			select {
			case <-ctx.Done():
				// Cancel the current process
				errCh <- ctx.Err()
				return
			default:
				// Create a unidirectional stream
				stream, err := s.session.AcceptUniStream(ctx)
				if err != nil {
					errCh <- err
				}

				// Read whole data on the stream
				data := make([]byte, 0, 1<<8)
				for {
					n, err := stream.Read(buf)
					if err != nil {
						if err == io.EOF {
							break
						}
						errCh <- err
						// Stop to read chunk when some error is detected
						// but continue to receive data if any error was detected
						break
					}
					data = append(data, buf[:n]...)
				}
				// Send data to the channel, If any data exists
				if len(data) > 0 {
					dataCh <- data
				}
			}
		}
	}()

	// Return the channels as read only channel
	return dataCh, errCh
}

// TODO: Consider a function for streaming processing
// func streamChunk(ch <-chan []byte, op func()) {

// }

func (s *Subscriber) Unsubscribe(id SubscribeID) error {
	// Check if the updated subscription is narrower than the existing subscription
	if int(id) > len(s.subscriptions) {
		//This means the specifyed subscription does not exist in the subscriptions
		return errors.New("invalid Subscribe ID")
	}

	return s.sendUnsubscribeMessage(id)
}

/******/
func (s Subscriber) sendUnsubscribeMessage(id SubscribeID) error {
	um := UnsubscribeMessage{
		SubscribeID: id,
	}

	// Send UNSUBSCRIBE message
	_, err := s.controlStream.Write(um.serialize())
	if err != nil {
		return err
	}

	return nil
}

// TODO:
func (s *Subscriber) SubscribeUpdate(id SubscribeID, config SubscribeConfig) error {
	// Check if the updated subscription is narrower than the existing subscription
	if int(id) > len(s.subscriptions) {
		//This means the specifyed subscription does not exist in the subscriptions
		return errors.New("invalid Subscribe ID")
	}

	if !config.SubscriptionFilter.isOK() {
		return errors.New("invalid Subscription Filter")
	}

	existingSubscription := &s.subscriptions[int(id)]

	// When Filter Code is ABSOLUTE_START
	if existingSubscription.FilterCode != ABSOLUTE_START {
		// Check if the update is valid
		if existingSubscription.startGroup > config.startGroup {
			return errors.New("invalid update due to Group ID")
		}
		if existingSubscription.startGroup == config.startGroup {
			if existingSubscription.startObject > config.startObject {
				return errors.New("invalid update due to Object ID")
			}
		}

		existingSubscription.SubscriptionFilter = config.SubscriptionFilter
	}

	// When Filter Code is ABSOLUTE_RANGE
	if existingSubscription.FilterCode != ABSOLUTE_RANGE {
		// Check if the update is valid
		if existingSubscription.startGroup > config.startGroup {
			return errors.New("invalid update due to Group ID")
		}
		if existingSubscription.startGroup == config.startGroup {
			if existingSubscription.startObject > config.startObject {
				return errors.New("invalid update due to Object ID")
			}
		}
		if existingSubscription.endGroup < config.endGroup {
			return errors.New("invalid update due to Group ID")
		}
		if existingSubscription.startGroup == config.startGroup {
			if existingSubscription.endObject < config.endObject {
				return errors.New("invalid update due to Object ID")
			}
		}

		existingSubscription.SubscriptionFilter = config.SubscriptionFilter
	}

	existingSubscription.Parameters = s.SubscribeParameters()

	return s.sendSubscribeUpdateMessage(id, config)
}

/****/
func (s Subscriber) sendSubscribeUpdateMessage(id SubscribeID, config SubscribeConfig) error {
	sum := SubscribeUpdateMessage{
		SubscribeID:        id,
		SubscriptionFilter: config.SubscriptionFilter,
		SubscriberPriority: config.SubscriberPriority,
		Parameters:         s.SubscribeUpdateParameters(),
	}

	// Send SUBSCRIBE_UPDATE message
	_, err := s.controlStream.Write(sum.serialize())
	if err != nil {
		return err
	}

	return nil
}

/*
 * Cancel the announce from the publisher
 * and stop the
 */
func (s Subscriber) CancelAnnounce() error {

	return s.sendAnnounceCancelMessage()
}

func (s Subscriber) sendAnnounceCancelMessage() error {
	return nil
}

func (s Subscriber) ObtainTrackStatus() error {
	return s.sendTrackStatusRequest()
}
func (s Subscriber) sendTrackStatusRequest() error {
	return nil
}
