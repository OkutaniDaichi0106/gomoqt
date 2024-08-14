package gomoq

import "errors"

type Subscriber struct {
	/*
	 * Client
	 * Subscriber is an extention of Client
	 */
	Client

	/*
	 * Latest Track Alias which start with 0 and increase by 1
	 * 0, 1, 2, ...
	 */
	latestTrackAlias int

	/*
	 * The number of the subscriptions
	 */
	subscriptions []SubscribeMessage
}

func (s *Subscriber) Connect(url string) error {
	// Check if the Client specify the Versions
	if len(s.clientSetupMessage.Versions) < 1 {
		return errors.New("no versions is specifyed")
	}

	// Add role parameter as publisher
	s.clientSetupMessage.Parameters.addIntParameter(role, uint64(sub))

	s.subscriptions = make([]SubscribeMessage, 1<<2)

	// Connect to the server
	err := s.connect(url)
	if err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) Subscribe(config SubscribeConfig) error {
	// Check Subscribe Configuration
	// Check if the Track Namespace is empty
	if len(config.trackNamespace) == 0 {
		return errors.New("no track namespace is specifyed")
	}
	// Check if the Track Name is empty
	if len(config.trackName) == 0 {
		return errors.New("no track name is specifyed")
	}
	// Check if the Filter is valid
	if config.SubscriptionFilter != LATEST_GROUP || config.SubscriptionFilter != LATEST_OBJECT || config.SubscriptionFilter != ABSOLUTE_RANGE || config.SubscriptionFilter != ABSOLUTE_START {
		return errors.New("invalid filter type")
	}

	// When the Filter is set to ABSOLUTE, Check if the Start Group ID is smaller than the End Group ID
	if config.SubscriptionFilter == ABSOLUTE_RANGE {
		if config.StartGroupID > config.EndGroupID {
			return errors.New("Start Group ID should be smaller than End Group ID")
		}
	}

	// Check if the track is already subscribed and get Track Alias
	var newTrackAlias TrackAlias
	trackExist := false
	for _, v := range s.subscriptions {
		// If the same full name already exists, reuse the Track Alias
		if v.TrackNamespace+v.TrackName == config.trackNamespace+config.trackName {
			newTrackAlias = v.TrackAlias
			trackExist = true
			break
		}
	}
	if !trackExist {
		newTrackAlias = TrackAlias(s.latestTrackAlias + 1)
	}

	sm := SubscribeMessage{
		SubscribeID:        SubscribeID(len(s.subscriptions)),
		TrackAlias:         newTrackAlias,
		TrackNamespace:     config.trackNamespace,
		TrackName:          config.trackName,
		SubscriberPriority: config.SubscriberPriority,
		GroupOrder:         config.GroupOrder,
		FilterType:         config.SubscriptionFilter,
		StartGroupID:       config.StartGroupID,
		StartObjectID:      config.StartObjectID,
		EndGroupID:         config.EndGroupID,
		EndObjectID:        config.EndObjectID,
	}

	// Send SUBSCRIBE message
	_, err = s.controlStream.Write(sm.serialize())
	if err != nil {
		return err
	}

	s.subscriptions++

	return nil
}

func (Subscriber) Unsubscribe(trackName string) {}

type SubscribeConfig struct {
	/*
	 * Track Namespace
	 */
	trackNamespace string
	/*
	 * Track Name
	 */
	trackName string
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
	 * If not specifyed, the value is set to 0 and this throughs an error
	 */
	SubscriptionFilter

	/*
	 * StartGroupID used only for "AbsoluteStart" or "AbsoluteRange"
	 */
	StartGroupID GroupID

	/*
	 * StartObjectID used only for "AbsoluteStart" or "AbsoluteRange"
	 */
	StartObjectID ObjectID

	/*
	 * EndGroupID used only for "AbsoluteRange"
	 */
	EndGroupID GroupID

	/*
	 * EndObjectID used only for "AbsoluteRange".
	 * When it is 0, it means the entire group is required
	 */
	EndObjectID ObjectID
}

func (sc *SubscribeConfig) check() error {
	// Check if the Track Namespace is empty
	if len(sc.trackNamespace) == 0 {
		return errors.New("no track namespace is specifyed")
	}

	// Check if the Track Name is empty
	if len(sc.trackName) == 0 {
		return errors.New("no track name is specifyed")
	}

	// Check if the the Filter is a defined one
	if sc.SubscriptionFilter != LATEST_GROUP || sc.SubscriptionFilter != LATEST_OBJECT || s.SubscriptionFilter != ABSOLUTE_RANGE || s.SubscriptionFilter != ABSOLUTE_START {
		return errors.New("invalid filter type")
	}

	// When the Filter is set to ABSOLUTE, Check if the Start Group ID is smaller than End Group ID
	if sc.SubscriptionFilter == ABSOLUTE_RANGE {
		if sc.StartGroupID > sc.EndGroupID {
			return errors.New("Start Group ID should be smaller than End Group ID")
		}
	}

	return nil
}

type Subscription struct {
	SubscribeID
	TrackAlias
	TrackNamespace string
	TrackName      string
}

type Subscriptions []Subscription

func (ss Subscriptions) addTrack(trackNamespace, trackName string) (Subscriptions, error) {
	trackExist := false
	for i, s := range ss {
		if s.TrackNamespace+s.TrackName == trackNamespace+trackName {
			// The track is already subscribed
			trackExist = true
		}
	}
	if !trackExist {

	}

	return s, nil
}
