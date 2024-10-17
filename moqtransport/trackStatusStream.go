package moqtransport

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/moqtmessage"
	"github.com/quic-go/quic-go/quicvarint"
)

/***/
type ReceiveTrackStatusStream struct {
	stream        Stream
	qvReader      quicvarint.Reader
	trackAliasMap *trackAliasMap
}

func (stream *ReceiveTrackStatusStream) RequestTrackStatus(trackNamespace moqtmessage.TrackNamespace, trackName string) (*TrackStatus, error) {
	/*
	 * Send a TRACK_STATUS_REQUEST message
	 */
	// Get the Track Alias
	trackAlias := stream.trackAliasMap.getAlias(trackNamespace, trackName)
	// Initialize a TRACK_STATUS_REQUEST message
	rts := moqtmessage.TrackStatusRequestMessage{
		TrackNamespace: trackNamespace,
		TrackName:      trackName,
		TrackAlias:     trackAlias,
	}

	_, err := stream.stream.Write(rts.Serialize())
	if err != nil {
		return nil, err
	}

	/*
	 * Receive a TRACK_STATUS message
	 */
	ts, err := receiveTrackStatus(stream.qvReader, trackAlias)
	if err != nil {
		return nil, err
	}

	return ts, nil
}

/***/
type SendTrackStatusStream struct {
	stream   Stream
	qvReader quicvarint.Reader
}

func (stream SendTrackStatusStream) ReceiveTrackStatusRequest() (*moqtmessage.TrackStatusRequestMessage, error) {
	// Receive a TRACK_STATUS_REQUEST message
	id, preader, err := moqtmessage.ReadControlMessage(stream.qvReader)
	if err != nil {
		return nil, err
	}
	if id != moqtmessage.TRACK_STATUS_REQUEST {
		return nil, ErrUnexpectedMessage
	}
	var tsrm moqtmessage.TrackStatusRequestMessage
	err = tsrm.DeserializePayload(preader)
	if err != nil {
		return nil, err
	}

	return &tsrm, nil
}

func (stream SendTrackStatusStream) SendTrackStatus(request moqtmessage.TrackStatusRequestMessage, ts TrackStatus) error {
	tsm := moqtmessage.TrackStatusMessage{
		TrackAlias:    request.TrackAlias,
		Code:          ts.Code,
		LatestGroupID: ts.LatestGroupID,
		GroupOrder:    ts.GroupOrder,
		GroupExpires:  ts.GroupExpires,
	}

	_, err := stream.stream.Write(tsm.Serialize())
	if err != nil {
		return err
	}

	return nil
}
