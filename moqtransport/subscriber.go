package moqtransport

import (
	"context"
	"errors"
	"go-moq/moqtransport/moqterror"
	"go-moq/moqtransport/moqtmessage"
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

	maxSubscribeID moqtmessage.SubscribeID

	/*
	 * Map of the Track Alias
	 * The key is the Track Full Name
	 */
	trackAliases map[string]moqtmessage.TrackAlias

	/*
	 * The number of the subscriptions
	 * The key is the Subscribe ID
	 */
	subscriptions map[moqtmessage.SubscribeID]moqtmessage.SubscribeMessage

	subscribeParameters       moqtmessage.Parameters
	subscribeUpdateParameters moqtmessage.Parameters
}

func (s *Subscriber) SubscribeParameters(params moqtmessage.Parameters) {
	s.subscribeParameters = params
}
func (s *Subscriber) SubscribeUpdateParameters(params moqtmessage.Parameters) {
	s.subscribeUpdateParameters = params
}

func (s *Subscriber) ConnectAndSetup(url string) (moqtmessage.Parameters, error) {
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

func (s *Subscriber) setup() (moqtmessage.Parameters, error) {
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
	csm := moqtmessage.ClientSetupMessage{
		Versions:   s.Versions,
		Parameters: make(moqtmessage.Parameters),
	}

	// Add role parameter
	csm.AddParameter(moqtmessage.ROLE, moqtmessage.SUB)

	_, err := s.controlStream.Write(csm.Serialize())

	return err
}

type SubscribeConfig struct {
	/*
	 * 0 is set by default
	 */
	moqtmessage.SubscriberPriority

	/*
	 * NOT_SPECIFY (= 0) is set by default
	 * If not specifyed, the value is set to 0 which means NOT_SPECIFY
	 */
	moqtmessage.GroupOrder

	/*
	 * No value is set by default
	 */
	moqtmessage.SubscriptionFilter
}

func (s *Subscriber) Subscribe(trackNamespace, trackName string, config *SubscribeConfig) error {
	if config == nil {
		config = &SubscribeConfig{}
	}
	if config.GroupOrder == 0 {
		config.GroupOrder = moqtmessage.ASCENDING
	}
	if config.FilterCode == 0 {
		config.FilterCode = moqtmessage.LATEST_GROUP
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
	err := config.SubscriptionFilter.IsOK()
	if err != nil {
		return err
	}

	if s.subscribeParameters == nil {
		s.subscribeParameters = make(moqtmessage.Parameters)
	}
	if s.trackAliases == nil {
		s.trackAliases = make(map[string]moqtmessage.TrackAlias)
	}

	// Check if the track is already subscribed
	// and add track alias
	trackAlias, ok := s.trackAliases[trackNamespace+trackName]

	// Get new Track Alias, if the Track did not already exist
	if !ok {
		trackAlias = moqtmessage.TrackAlias(len(s.trackAliases))
		s.trackAliases[trackNamespace+trackName] = trackAlias
	}

	subscribeID := moqtmessage.SubscribeID(len(s.subscriptions))

	sm := moqtmessage.SubscribeMessage{
		SubscribeID:        subscribeID,
		TrackAlias:         trackAlias,
		TrackNamespace:     trackNamespace,
		TrackName:          trackName,
		SubscriberPriority: config.SubscriberPriority,
		GroupOrder:         config.GroupOrder,
		SubscriptionFilter: config.SubscriptionFilter,
		Parameters:         s.subscribeParameters, //TODO
	}

	// Send SUBSCRIBE message
	_, err = s.controlStream.Write(sm.Serialize())
	if err != nil {
		return err
	}

	// Register the SUBSCRIBE message
	s.subscriptions[subscribeID] = sm

	return nil
}

func (s Subscriber) receiveSubscribeResponce() error {
	// Receive SUBSCRIBE_OK message or SUBSCRIBE_ERROR message
	id, err := moqtmessage.DeserializeMessageID(s.controlReader)
	if err != nil {
		return err
	}

	switch id {
	case moqtmessage.SUBSCRIBE_OK:
		so := moqtmessage.SubscribeOkMessage{}
		so.DeserializeBody(s.controlReader)
		return nil
	case moqtmessage.SUBSCRIBE_ERROR:
		se := moqterror.SubscribeError{}
		se.DeserializeBody(s.controlReader)

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
	id, err := moqtmessage.DeserializeStreamType(reader)
	if err != nil {
		return nil, err
	}
	//var dataStream DataStream

	switch id {
	case moqtmessage.STREAM_HEADER_TRACK:
		sht := moqtmessage.StreamHeaderTrack{}
		err = sht.DeserializeStreamHeader(reader)
		if err != nil {
			return nil, err
		}

		// Create new data stream
		stream := &trackStream{
			header: sht,
			chunks: make([]moqtmessage.GroupChunk, 0, 1<<3),
			closed: false,
		}

		go func() {
			// Read and write chunks to the data stream
			var chunk moqtmessage.GroupChunk
			for {
				err = chunk.DeserializeBody(reader)
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Println(err)
					return
				}
				// Check if the chunk is the end of the stream
				if chunk.StatusCode == moqtmessage.END_OF_TRACK {
					break
				}

				// Add the data
				stream.write(chunk)
			}
		}()

		return stream, nil
	case moqtmessage.STREAM_HEADER_PEEP:
		shp := moqtmessage.StreamHeaderPeep{}
		err = shp.DeserializeStreamHeader(reader)
		if err != nil {
			return nil, err
		}

		// Create new data stream
		stream := &peepStream{
			header: shp,
			chunks: make([]moqtmessage.ObjectChunk, 0, 1<<3),
			closed: false,
		}

		go func() {
			// Read and write chunks to the data stream
			var chunk moqtmessage.ObjectChunk
			for {
				err = chunk.DeserializeBody(reader)
				if err != nil {
					if err == io.EOF {
						stream.Close()
						return
					}
					log.Println(err)
					return
				}
				// Check if the chunk is the end of the stream
				if chunk.StatusCode == moqtmessage.END_OF_PEEP {
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

func (s *Subscriber) Unsubscribe(id moqtmessage.SubscribeID) error {
	_, ok := s.subscriptions[id]
	if !ok {
		return errors.New("subscription not found")
	}

	// Delete the subscription
	delete(s.subscriptions, id)

	// Send UNSUBSCRIBE message
	um := moqtmessage.UnsubscribeMessage{
		SubscribeID: id,
	}

	_, err := s.controlStream.Write(um.Serialize())

	return err
}

/*
 * Update the specifyed subscription
 */
func (s *Subscriber) SubscribeUpdate(id moqtmessage.SubscribeID, config SubscribeUpdateConfig) error {
	// Retrieve the old subscription
	old, ok := s.subscriptions[id]
	if !ok {
		return errors.New("subscription not found")
	}

	// Validate filter configuration
	filter := moqtmessage.SubscriptionFilter{
		FilterCode:  old.FilterCode,
		FilterRange: config.FilterRange,
	}

	err := filter.IsOK()
	if err != nil {
		return err
	}

	var ErrInvalidUpdate = errors.New("invalid update")

	// Check if the updated subscription is narrower than the existing subscription
	switch old.GroupOrder {
	case moqtmessage.ASCENDING:
		switch old.FilterCode {
		case moqtmessage.ABSOLUTE_START:
			// Check if the update is valid
			if config.StartGroup < old.StartGroup {
				return ErrInvalidUpdate
			}
			if old.StartGroup == config.StartGroup && config.StartObject < old.StartObject {
				return ErrInvalidUpdate
			}

		case moqtmessage.ABSOLUTE_RANGE:
			// Check if the update is valid

			// Check if the new Start Group ID is larger than old Start Group ID
			if config.StartGroup < old.StartGroup {
				return ErrInvalidUpdate
			}

			// Check if the new Start Object ID is larger than old Start Object ID
			if old.StartGroup == config.StartGroup && config.StartObject < old.StartObject {
				return ErrInvalidUpdate
			}

			// Check if the new End Group ID is smaller than old End Group ID
			if old.EndGroup < config.EndGroup {
				return ErrInvalidUpdate
			}

			// Check if the End Object ID is smaller than old End Object ID
			if old.StartGroup == config.StartGroup && old.EndObject < config.EndObject {
				return ErrInvalidUpdate
			}
		}
	case moqtmessage.DESCENDING:
		switch old.FilterCode {
		case moqtmessage.ABSOLUTE_START:
			// Check if the update is valid
			if old.StartGroup < config.StartGroup {
				return ErrInvalidUpdate
			}
			if old.StartGroup == config.StartGroup && old.StartObject < config.StartObject {
				return ErrInvalidUpdate
			}

		case moqtmessage.ABSOLUTE_RANGE:
			// Check if the new Start Group ID is larger than old Start Group ID
			if old.StartGroup < config.StartGroup {
				return ErrInvalidUpdate
			}

			// Check if the new Start Object ID is larger than old Start Object ID
			if old.StartGroup == config.StartGroup && old.StartObject < config.StartObject {
				return ErrInvalidUpdate
			}

			// Check if the new End Group ID is smaller than old End Group ID
			if config.EndGroup < old.EndGroup {
				return ErrInvalidUpdate
			}

			// Check if the End Object ID is smaller than old End Object ID
			if old.StartGroup == config.StartGroup && config.EndObject < old.EndObject {
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
func (s Subscriber) sendSubscribeUpdateMessage(id moqtmessage.SubscribeID, config SubscribeUpdateConfig) error {
	sum := moqtmessage.SubscribeUpdateMessage{
		SubscribeID:        id,
		FilterRange:        config.FilterRange,
		SubscriberPriority: config.SubscriberPriority,
		Parameters:         config.Parameters,
	}
	// Send SUBSCRIBE_UPDATE message
	_, err := s.controlStream.Write(sum.Serialize())

	return err
}

type SubscribeUpdateConfig struct {
	moqtmessage.SubscriberPriority
	moqtmessage.FilterRange
	Parameters moqtmessage.Parameters
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

	acm := moqtmessage.AnnounceCancelMessage{
		TrackNamespace: trackNamespace,
	}

	_, err := s.controlStream.Write(acm.Serialize())
	return err
}

func (s Subscriber) GetTrackStatus(trackNamespace, trackName string) error {

	return s.sendTrackStatusRequest(trackNamespace, trackName)
}
func (s Subscriber) sendTrackStatusRequest(trackNamespace, trackName string) error {
	tsr := moqtmessage.TrackStatusRequest{
		TrackNamespace: trackNamespace,
		TrackName:      trackName,
	}

	_, err := s.controlStream.Write(tsr.Serialize())

	return err
}
