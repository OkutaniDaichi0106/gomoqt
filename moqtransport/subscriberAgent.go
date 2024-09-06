package moqtransport

import "log"

type subscriberAgenter interface {
	advertise([]AnnounceMessage) error
	acceptSubscription() error
}

type SubscriberAgent struct {
	clientAgent

	/***/
	origin *PublisherAgent
}

func (SubscriberAgent) Role() Role {
	return SUB
}

/*
 * Advertise announcements
 */
func Advertise(agent subscriberAgenter, announcements []AnnounceMessage) error {
	return agent.advertise(announcements)
}

func (a *SubscriberAgent) advertise(announcements []AnnounceMessage) error {
	var err error
	// Send all ANNOUNCE messages
	for _, am := range announcements {
		_, err = a.controlStream.Write(am.serialize())
		if err != nil {
			return err
		}
	}

	return nil
}

func DeliverObjects(agent *SubscriberAgent) {
	agent.origin.destinations.mu.Lock()
	defer agent.origin.destinations.mu.Unlock()

	// Register the session with the Subscriber to the Publisher's Agent
	agent.origin.destinations.sessions = append(agent.origin.destinations.sessions, agent.session)
	log.Println(agent.session, " is in ", agent.origin.destinations.sessions)
}

/*
 * Exchange SUBSCRIBE messages
 */
func AcceptSubscription(agent subscriberAgenter) error {
	return agent.acceptSubscription()
}

func (a *SubscriberAgent) acceptSubscription() error {
	// Receive a SUBSCRIBE message
	id, err := deserializeHeader(a.controlReader)
	if err != nil {
		return err
	}
	if id != SUBSCRIBE {
		return ErrUnexpectedMessage //TODO: handle as protocol violation
	}
	s := SubscribeMessage{}
	err = s.deserializeBody(a.controlReader)
	if err != nil {
		return err
	}

	// Find the Publisher's Agent from the Track Namespace
	pAgent, err := SERVER.getPublisherAgent(s.TrackNamespace)
	if err != nil {
		se := SubscribeError{
			subscribeID: s.subscribeID,
			Code:        SUBSCRIBE_INTERNAL_ERROR,
			Reason:      SUBSCRIBE_ERROR_REASON[SUBSCRIBE_INTERNAL_ERROR],
			TrackAlias:  s.TrackAlias,
		}
		_, err2 := a.controlStream.Write(se.serialize()) // TODO: handle the error
		log.Println(err2)

		return err
	}

	pAgent.controlCh <- &s

	// Receive SUBSCRIBE_OK message or SUBSCRIBE_ERROR message from Publisher's Agent
	// and send it to Subscriber
	data := <-a.controlCh
	switch MessageID(data[0]) {
	case SUBSCRIBE_OK:
		_, err = a.controlStream.Write(data)
		if err != nil {
			return err
		}
		// Add the agent to the Index
		SERVER.subscribers.add(s.TrackNamespace, a)
		// TODO: when delete agents from the index
	case SUBSCRIBE_ERROR:
		_, err = a.controlStream.Write(data)
		if err != nil {
			return err
		}
	default:
		return ErrUnexpectedMessage // TODO: protocol violation
	}

	a.origin = pAgent

	return nil
}
