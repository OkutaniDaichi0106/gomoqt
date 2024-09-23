package moqtransport

// import (
// 	"errors"
// 	"fmt"
// 	"go-moq/moqtransport/moqtmessage"
// )

// func (s *session) exchangeAnnounce() error {

// 	// Send an ANNOUNCE message
// 	err := s.sendAnnounce()
// 	if err != nil {
// 		return err
// 	}

// 	// Receive an ANNOUNCE message
// 	err = s.receiveAnnounce()
// 	if err != nil {
// 		return err
// 	}

// 	// Send a responce to the received ANNOUNCE message
// 	err = s.sendAnnounceOk()
// 	if err != nil {
// 		return err
// 	}

// 	// Register the session as publisher session
// 	publishers.add(s) //TODO

// 	// Receive a responce to the sent ANNOUNCE message
// 	err = s.receiveAnnounceResponce()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (s *session) receiveAnnounceResponce() error {
// 	//Receive ANNOUNCE_OK message or ANNOUNCE_ERROR message
// 	id, err := moqtmessage.DeserializeMessageID(s.controlReader)
// 	if err != nil {
// 		return err
// 	}

// 	switch id {
// 	case moqtmessage.ANNOUNCE_OK:
// 		var ao moqtmessage.AnnounceOkMessage
// 		err = ao.DeserializeBody(s.controlReader)
// 		if err != nil {
// 			return err
// 		}

// 		// Check if the Track Namespace in the responce is valid
// 		for i, v := range s.trackNamespace {
// 			if v != ao.TrackNamespace[i] {
// 				return errors.New("unexpected Track Namespace")
// 			}
// 		}

// 	case moqtmessage.ANNOUNCE_ERROR:
// 		var ae moqtmessage.AnnounceErrorMessage // TODO: Handle Error Code
// 		err = ae.DeserializeBody(s.controlReader)
// 		if err != nil {
// 			return err
// 		}

// 		// Check the Track Namespace in the responce
// 		for i, v := range s.trackNamespace {
// 			if v != ae.TrackNamespace[i] {
// 				return errors.New("unexpected Track Namespace")
// 			}
// 		}

// 		return errors.New(fmt.Sprint(ae.Code, ae.Reason))

// 	default:
// 		return ErrUnexpectedMessage
// 	}
// 	return nil
// }
// func (s session) receiveAnnounce() error {
// 	// Receive an ANNOUNCE message
// 	id, err := moqtmessage.DeserializeMessageID(s.controlReader)
// 	if err != nil {
// 		return err
// 	}
// 	if id != moqtmessage.ANNOUNCE {
// 		return ErrUnexpectedMessage
// 	}

// 	//TODO: handle error
// 	var am moqtmessage.AnnounceMessage
// 	err = am.DeserializeBody(s.controlReader)
// 	if err != nil {
// 		return err
// 	}

// 	// Register the ANNOUNCE message
// 	announcements.add(am)

// 	// Register the Track Namespace
// 	s.trackNamespace = am.TrackNamespace

// 	// Get MAX_CACHE_DURATION parameter
// 	s.maxCacheDuration, err = am.Parameters.MaxCacheDuration()
// 	// Ignore the ErrMaxCacheDurationNotFound
// 	if err != nil && err != moqtmessage.ErrMaxCacheDurationNotFound {
// 		return err
// 	}

// 	return nil
// }

// func (s session) sendAnnounceOk() error {
// 	// Send ANNOUNCE_OK message
// 	ao := moqtmessage.AnnounceOkMessage{
// 		TrackNamespace: s.trackNamespace,
// 	}
// 	_, err := s.controlStream.Write(ao.Serialize()) // Handle the error when wrinting message

// 	return err
// }

// func (s session) sendAnnounceError(annErr AnnounceError) error {

// }
