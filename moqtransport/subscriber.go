package moqtransport

import (
	"container/heap"
	"context"
	"errors"
	"io"
	"log"

	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

type Subscriber struct {
	/*
	 * Client
	 * Subscriber is an extention of Client
	 */
	Client

	/***/
	SubscriberHandler

	maxSubscribeID subscribeID

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

func (s *Subscriber) ConnectAndSetup(url string) (Parameters, error) {
	// Check if the Client specify the Versions
	if len(s.Versions) < 1 {
		panic("no versions is specified")
	}

	// Connect
	err := s.connect(url)
	if err != nil {
		return nil, err
	}

	// Setup
	params, err := s.setup(SUB)
	if err != nil {
		return nil, err
	}

	// Check if the ROLE parameter is valid
	s.maxSubscribeID, err = params.MaxSubscribeID()
	if err != nil {
		return nil, err
	}
	// Delete the Parameter after getting it
	delete(params, MAX_SUBSCRIBE_ID)

	return params, nil
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
	err := s.sendSubscribe(trackNamespace, trackName, config)
	if err != nil {
		return err
	}

	err = s.receiveSubscribeResponce()
	if err != nil {
		return err
	}

	return nil
}

func (s Subscriber) sendSubscribe(trackNamespace, trackName string, config SubscribeConfig) error {
	// Check if the Filter is valid
	if !config.SubscriptionFilter.isOK() {
		return ErrInvalidFilter
	}
	// Check if the track is already subscribed
	// and add track alias
	trackAlias, ok := s.trackAliases[trackNamespace+trackName]

	// Get new Track Alias, if the Track does not already exist
	if !ok {
		trackAlias = TrackAlias(len(s.trackAliases))
		s.trackAliases[trackNamespace+trackName] = trackAlias
	}
	sm := SubscribeMessage{
		subscribeID:        subscribeID(len(s.subscriptions)),
		TrackAlias:         trackAlias,
		TrackNamespace:     trackNamespace,
		TrackName:          trackName,
		SubscriberPriority: config.SubscriberPriority,
		GroupOrder:         config.GroupOrder,
		SubscriptionFilter: config.SubscriptionFilter,
		Parameters:         s.SubscribeParameters(), //TODO
	}

	// Send SUBSCRIBE message
	_, err := s.controlStream.Write(sm.serialize())
	if err != nil {
		return err
	}

	// Register the SUBSCRIBE message
	s.subscriptions = append(s.subscriptions, sm)

	return nil
}

func (s Subscriber) receiveSubscribeResponce() error {
	// Receive SUBSCRIBE_OK message or SUBSCRIBE_ERROR message
	id, err := deserializeHeader(s.controlReader)
	if err != nil {
		return err
	}
	switch id {
	case SUBSCRIBE_OK:
		so := SubscribeOkMessage{}
		so.deserializeBody(s.controlReader)
		return nil
	case SUBSCRIBE_ERROR:
		se := SubscribeError{}
		se.deserializeBody(s.controlReader)
		log.Println(se)

		return errors.New(se.Reason)
	default:
		return ErrUnexpectedMessage
	}
}

var ErrInvalidFilter = errors.New("invalid filter type")

func (s *Subscriber) AcceptObjects(ctx context.Context) (<-chan DataStream, <-chan error) {
	dataCh := make(chan DataStream, 1<<4)
	errCh := make(chan error, 1) // TODO: Consider the buffer size
	defer close(errCh)

	// Receive data on a stream in a goroutine
	go func() {
		for {
			// Accept a new unidirectional stream
			stream, err := s.session.AcceptUniStream(ctx)
			// Until the new stream is opened, proccess stops here
			if err != nil {
				errCh <- err
			}

			// Get data on the stream in a goroutine
			go proccessStream(stream, dataCh, errCh)

			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case <-errCh:
				return
			default:
				continue
			}
		}
	}()

	return dataCh, errCh
}

func proccessStream(stream webtransport.ReceiveStream, dataCh chan<- DataStream, errCh chan<- error) {
	reader := quicvarint.NewReader(stream)
	// Read the first object
	id, err := deserializeHeader(reader)
	if err != nil {
		errCh <- err
		return
	}

	var dataStream DataStream

	switch id {
	case STREAM_HEADER_TRACK:
		sht := StreamHeaderTrack{}
		err = sht.deserializeBody(reader)
		if err != nil {
			errCh <- err
			return
		}

		// Create new data stream
		dataStream = newDataStream(&sht) //TODO: pass

		// Register the stream
		dataCh <- dataStream

		// Read and write chunks to the data stream
		chunk := GroupChunk{}
		for {
			err = chunk.deserializeBody(reader)
			if err != nil {
				if err == io.EOF {
					return
				}
				errCh <- err
				return
			}
			// Check if the chunk is the end of the stream
			// if chunk.StatusCode == END_OF_TRACK {
			// 	return
			// }

			// Skip queueing, if there is no payload
			if len(chunk.Payload) == 0 {
				continue
			}

			// Queue the data
			heap.Push(dataStream, chunk)
		}

	case STREAM_HEADER_PEEP:
		shp := StreamHeaderPeep{}
		err = shp.deserializeBody(reader)
		if err != nil {
			errCh <- err
			return
		}

		// Create new data stream
		dataStream = newDataStream(&shp)

		// Register the stream
		dataCh <- dataStream

		// Read and write chunks to the data stream
		chunk := ObjectChunk{}
		for {
			err = chunk.deserializeBody(reader)
			if err != nil {
				if err == io.EOF {
					return
				}
				errCh <- err
				return
			}
			// Check if the chunk is the end of the stream
			// if chunk.StatusCode == END_OF_PEEP { // TODO: wait till all data have been sent
			// 	return
			// }

			// Skip queueing, if there is no payload
			if len(chunk.Payload) == 0 {
				continue
			}

			// Queue the data
			heap.Push(dataStream, chunk)
		}
	default:
		errCh <- ErrUnexpectedMessage
		return
	}

}

func (s *Subscriber) Unsubscribe(id subscribeID) error {
	// Check if the updated subscription is narrower than the existing subscription
	if int(id) > len(s.subscriptions) {
		//This means the specifyed subscription does not exist in the subscriptions
		return errors.New("invalid Subscribe ID")
	}
	um := UnsubscribeMessage{
		subscribeID: id,
	}

	// Send UNSUBSCRIBE message
	_, err := s.controlStream.Write(um.serialize())
	if err != nil {
		return err
	}

	return nil

}

// TODO:
func (s *Subscriber) SubscribeUpdate(id subscribeID, config SubscribeConfig) error {
	// Check if the updated subscription is narrower than the existing subscription
	if id > subscribeID(len(s.subscriptions)) {
		//This means the specifyed subscription does not exist in the subscriptions
		return errors.New("invalid Subscribe ID")
	}

	if !config.SubscriptionFilter.isOK() {
		return errors.New("invalid Subscription Filter")
	}

	existingSubscription := s.subscriptions[int(id)]

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
func (s Subscriber) sendSubscribeUpdateMessage(id subscribeID, config SubscribeConfig) error {
	sum := SubscribeUpdateMessage{
		subscribeID:        id,
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
