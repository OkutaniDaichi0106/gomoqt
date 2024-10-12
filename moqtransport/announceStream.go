package moqtransport

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go/quicvarint"
)

type SendAnnounceStream struct {
	stream   Stream
	qvReader quicvarint.Reader
}

func (sender SendAnnounceStream) ReceiveSubscribeNamespace() (moqtmessage.TrackNamespacePrefix, moqtmessage.Parameters, error) {
	/*
	 * Receive a SUBSCRIBE_NAMESPACE message
	 */
	// Read an control message
	id, preader, err := moqtmessage.ReadControlMessage(sender.qvReader)
	if err != nil {
		return nil, nil, err
	}
	// Verify the message is a SUBSCRIBE_NAMESPACE message
	if id != moqtmessage.SUBSCRIBE_NAMESPACE {
		return nil, nil, ErrProtocolViolation
	}
	// Deserialize the payload
	var snm moqtmessage.SubscribeNamespaceMessage
	err = snm.DeserializePayload(preader)
	if err != nil {
		return nil, nil, err
	}

	return snm.TrackNamespacePrefix, snm.Parameters, nil
}

func (sender SendAnnounceStream) Announce(trackNamespace moqtmessage.TrackNamespace, config AnnounceConfig) error {
	/*
	 * Send an ANNOUNCE message
	 */
	// Initialize an ANNOUNCE message
	am := moqtmessage.AnnounceMessage{
		TrackNamespace: trackNamespace,
		Parameters:     make(moqtmessage.Parameters),
	}
	// Add some Parameters
	am.Parameters.AddParameter(moqtmessage.AUTHORIZATION_INFO, config.AuthorizationInfo)
	am.Parameters.AddParameter(moqtmessage.MAX_CACHE_DURATION, config.MaxCacheDuration)

	// Send the message
	_, err := sender.stream.Write(am.Serialize())
	if err != nil {
		return err
	}

	// Catch any Stream Error
	_, err = sender.qvReader.Read([]byte{})
	if err != nil {
		return err
	}

	return nil
}

type ReceiveAnnounceStream struct {
	stream   Stream
	qvReader quicvarint.Reader
}

func (receiver ReceiveAnnounceStream) SubscribeNamespace(trackNamespacePrefix moqtmessage.TrackNamespacePrefix) error {
	/*
	 * Send a SUBSCRIBE_NAMESPACE message
	 */
	// Initialize a SUBSCRIBE_NAMESPACE message
	snm := moqtmessage.SubscribeNamespaceMessage{
		TrackNamespacePrefix: trackNamespacePrefix,
		Parameters:           make(moqtmessage.Parameters),
	}
	// TODO: Handle the parameters

	// Send the message
	_, err := receiver.stream.Write(snm.Serialize())
	if err != nil {
		return err
	}

	return nil
}

func (receiver ReceiveAnnounceStream) ReceiveAnnounce() (*Announcement, error) {
	/*
	 * Receive a ANNOUNCE message
	 */
	// Read a control message
	id, preader, err := moqtmessage.ReadControlMessage(receiver.qvReader)
	if err != nil {
		return nil, err
	}
	// Verify the message is a ANNOUNCE message
	if id != moqtmessage.ANNOUNCE {
		return nil, ErrProtocolViolation
	}
	// Deserialize the message
	var am moqtmessage.AnnounceMessage
	err = am.DeserializePayload(preader)
	if err != nil {
		return nil, err
	}

	// Initialize an Announcement
	announcement := Announcement{
		trackNamespace: am.TrackNamespace,
	}
	authInfo, ok := am.Parameters.AuthorizationInfo()
	if ok {
		announcement.AuthorizationInfo = authInfo
		am.Parameters.Remove(moqtmessage.AUTHORIZATION_INFO)
	}
	announcement.Parameters = am.Parameters

	return &announcement, nil
}
