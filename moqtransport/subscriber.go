package moqtransport

import (
	"context"
	"errors"
	"io"
	"log"
	"strings"

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

	maxSubscribeID subscribeID

	/*
	 * Map of the Track Alias
	 * The key is the Track Full Name
	 */
	trackAliases map[string]TrackAlias

	/*
	 * The number of the subscriptions
	 * The key is the Subscribe ID
	 */
	subscriptions map[subscribeID]SubscribeMessage

	subscribeParameters       Parameters
	subscribeUpdateParameters Parameters
}

func (s *Subscriber) SubscribeParameters(params Parameters) {
	s.subscribeParameters = params
}
func (s *Subscriber) SubscribeUpdateParameters(params Parameters) {
	s.subscribeUpdateParameters = params
}

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
	params, err := s.setup()
	if err != nil {
		return nil, err
	}

	s.maxSubscribeID, err = params.MaxSubscribeID()
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (s *Subscriber) setup() (Parameters, error) {
	var err error

	// Open first stream to send control messages
	s.controlStream, err = s.session.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}

	// Send SETUP_CLIENT message
	err = s.sendClientSetup()
	if err != nil {
		return nil, err
	}

	// Initialize control reader
	s.controlReader = quicvarint.NewReader(s.controlStream)

	// Receive SETUP_SERVER message
	return s.receiveServerSetup()
}

func (s Subscriber) sendClientSetup() error {
	// Initialize SETUP_CLIENT csm
	csm := ClientSetupMessage{
		Versions:   s.Versions,
		Parameters: make(Parameters),
	}

	// Add role parameter
	csm.AddParameter(ROLE, SUB)

	_, err := s.controlStream.Write(csm.serialize())

	return err
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

func (s *Subscriber) Subscribe(trackNamespace, trackName string, config *SubscribeConfig) error {
	if config == nil {
		config = &SubscribeConfig{}
	}
	if config.GroupOrder == 0 {
		config.GroupOrder = ASCENDING
	}
	if config.FilterCode == 0 {
		config.FilterCode = LATEST_GROUP
	}

	err := s.sendSubscribe(trackNamespace, trackName, *config)
	if err != nil {
		return err
	}

	err = s.receiveSubscribeResponce()
	if err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) sendSubscribe(trackNamespace, trackName string, config SubscribeConfig) error {
	// Check if the Filter is valid
	err := config.SubscriptionFilter.isOK()
	if err != nil {
		return err
	}

	if s.subscribeParameters == nil {
		s.subscribeParameters = make(Parameters)
	}
	if s.trackAliases == nil {
		s.trackAliases = make(map[string]TrackAlias)
	}

	// Check if the track is already subscribed
	// and add track alias
	trackAlias, ok := s.trackAliases[trackNamespace+trackName]

	// Get new Track Alias, if the Track did not already exist
	if !ok {
		trackAlias = TrackAlias(len(s.trackAliases))
		s.trackAliases[trackNamespace+trackName] = trackAlias
	}

	subscribeID := subscribeID(len(s.subscriptions))

	sm := SubscribeMessage{
		subscribeID:        subscribeID,
		TrackAlias:         trackAlias,
		TrackNamespace:     trackNamespace,
		TrackName:          trackName,
		SubscriberPriority: config.SubscriberPriority,
		GroupOrder:         config.GroupOrder,
		SubscriptionFilter: config.SubscriptionFilter,
		Parameters:         s.subscribeParameters, //TODO
	}

	// Send SUBSCRIBE message
	_, err = s.controlStream.Write(sm.serialize())
	if err != nil {
		return err
	}

	// Register the SUBSCRIBE message
	s.subscriptions[subscribeID] = sm

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

		return errors.New(se.Reason)
	default:
		return ErrUnexpectedMessage
	}
}

func (s *Subscriber) AcceptObjects(ctx context.Context) (ObjectStream, error) {
	// Accept a new unidirectional stream
	wtStream, err := s.session.AcceptUniStream(ctx)
	// Until the new stream is opened, proccess stops here
	if err != nil {
		return nil, err
	}

	stream, err := newObjectStream(wtStream)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return stream, nil
}

func newObjectStream(stream webtransport.ReceiveStream) (ObjectStream, error) {
	reader := quicvarint.NewReader(stream)
	// Read the first object
	id, err := deserializeHeader(reader)
	if err != nil {
		return nil, err
	}
	//var dataStream DataStream

	switch id {
	case STREAM_HEADER_TRACK:
		sht := StreamHeaderTrack{}
		err = sht.deserializeBody(reader)
		if err != nil {
			return nil, err
		}

		// Create new data stream
		stream := &trackStream{
			header: sht,
			chunks: make([]GroupChunk, 0, 1<<3),
			closed: false,
		}

		go func() {
			// Read and write chunks to the data stream
			var chunk GroupChunk
			for {
				err = chunk.deserializeBody(reader)
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Println(err)
					return
				}
				// Check if the chunk is the end of the stream
				if chunk.StatusCode == END_OF_TRACK {
					break
				}

				// Add the data
				stream.write(chunk)
			}
		}()

		return stream, nil
	case STREAM_HEADER_PEEP:
		shp := StreamHeaderPeep{}
		err = shp.deserializeBody(reader)
		if err != nil {
			return nil, err
		}

		// Create new data stream
		stream := &peepStream{
			header: shp,
			chunks: make([]ObjectChunk, 0, 1<<3),
			closed: false,
		}

		go func() {
			// Read and write chunks to the data stream
			var chunk ObjectChunk
			for {
				err = chunk.deserializeBody(reader)
				if err != nil {
					if err == io.EOF {
						stream.Close()
						return
					}
					log.Println(err)
					return
				}
				// Check if the chunk is the end of the stream
				if chunk.StatusCode == END_OF_PEEP {
					stream.Close()
					return
				}

				// Add the data
				stream.write(chunk)
			}
		}()
		return stream, nil
	default:
		return nil, ErrUnexpectedMessage
	}

}

func (s *Subscriber) Unsubscribe(id subscribeID) error {
	_, ok := s.subscriptions[id]
	if !ok {
		return errors.New("subscription not found")
	}

	// Delete the subscription
	delete(s.subscriptions, id)

	// Send UNSUBSCRIBE message
	um := UnsubscribeMessage{
		subscribeID: id,
	}

	_, err := s.controlStream.Write(um.serialize())

	return err
}

/*
 * Update the specifyed subscription
 */
func (s *Subscriber) SubscribeUpdate(id subscribeID, config SubscribeUpdateConfig) error {
	// Retrieve the old subscription
	old, ok := s.subscriptions[id]
	if !ok {
		return errors.New("subscription not found")
	}

	// Validate filter configuration
	filter := SubscriptionFilter{
		FilterCode:  old.FilterCode,
		FilterRange: config.FilterRange,
	}

	err := filter.isOK()
	if err != nil {
		return err
	}

	var ErrInvalidUpdate = errors.New("invalid update")

	// Check if the updated subscription is narrower than the existing subscription
	switch old.GroupOrder {
	case ASCENDING:
		switch old.FilterCode {
		case ABSOLUTE_START:
			// Check if the update is valid
			if config.startGroup < old.startGroup {
				return ErrInvalidUpdate
			}
			if old.startGroup == config.startGroup && config.startObject < old.startObject {
				return ErrInvalidUpdate
			}

		case ABSOLUTE_RANGE:
			// Check if the update is valid

			// Check if the new Start Group ID is larger than old Start Group ID
			if config.startGroup < old.startGroup {
				return ErrInvalidUpdate
			}

			// Check if the new Start Object ID is larger than old Start Object ID
			if old.startGroup == config.startGroup && config.startObject < old.startObject {
				return ErrInvalidUpdate
			}

			// Check if the new End Group ID is smaller than old End Group ID
			if old.endGroup < config.endGroup {
				return ErrInvalidUpdate
			}

			// Check if the End Object ID is smaller than old End Object ID
			if old.startGroup == config.startGroup && old.endObject < config.endObject {
				return ErrInvalidUpdate
			}
		}
	case DESCENDING:
		switch old.FilterCode {
		case ABSOLUTE_START:
			// Check if the update is valid
			if old.startGroup < config.startGroup {
				return ErrInvalidUpdate
			}
			if old.startGroup == config.startGroup && old.startObject < config.startObject {
				return ErrInvalidUpdate
			}

		case ABSOLUTE_RANGE:
			// Check if the new Start Group ID is larger than old Start Group ID
			if old.startGroup < config.startGroup {
				return ErrInvalidUpdate
			}

			// Check if the new Start Object ID is larger than old Start Object ID
			if old.startGroup == config.startGroup && old.startObject < config.startObject {
				return ErrInvalidUpdate
			}

			// Check if the new End Group ID is smaller than old End Group ID
			if config.endGroup < old.endGroup {
				return ErrInvalidUpdate
			}

			// Check if the End Object ID is smaller than old End Object ID
			if old.startGroup == config.startGroup && config.endObject < old.endObject {
				return ErrInvalidUpdate
			}
		}
	}

	new := old

	new.SubscriberPriority = config.SubscriberPriority
	new.FilterRange = config.FilterRange
	new.Parameters = config.Parameters

	s.subscriptions[id] = new

	return s.sendSubscribeUpdateMessage(id, config)
}

/*
 * Send SUBSCRIBE_UPDATE message
 */
func (s Subscriber) sendSubscribeUpdateMessage(id subscribeID, config SubscribeUpdateConfig) error {
	sum := SubscribeUpdateMessage{
		subscribeID:        id,
		FilterRange:        config.FilterRange,
		SubscriberPriority: config.SubscriberPriority,
		Parameters:         config.Parameters,
	}
	// Send SUBSCRIBE_UPDATE message
	_, err := s.controlStream.Write(sum.serialize())

	return err
}

type SubscribeUpdateConfig struct {
	SubscriberPriority
	FilterRange
	Parameters Parameters
}

func (s Subscriber) CancelAnnounce(trackNamespace ...string) error {
	// Delete the Track Namespace from the map of Track Alias
	fullTrackNamespace := strings.Join(trackNamespace, "")
	hasTrackAlias := false
	for trackFullName := range s.trackAliases {
		if strings.HasPrefix(trackFullName, fullTrackNamespace) {
			hasTrackAlias = true
			delete(s.trackAliases, trackFullName)
		}
	}

	if !hasTrackAlias {
		return errors.New("track not found")
	}

	acm := AnnounceCancelMessage{
		TrackNamespace: trackNamespace,
	}

	_, err := s.controlStream.Write(acm.serialize())
	return err
}

func (s Subscriber) GetTrackStatus(trackNamespace, trackName string) error {

	return s.sendTrackStatusRequest(trackNamespace, trackName)
}
func (s Subscriber) sendTrackStatusRequest(trackNamespace, trackName string) error {
	tsr := TrackStatusRequest{
		TrackNamespace: trackNamespace,
		TrackName:      trackName,
	}

	_, err := s.controlStream.Write(tsr.serialize())

	return err
}
