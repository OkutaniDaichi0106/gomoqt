package moqtransfork

type Session struct {
	Connection Connection

	version Version
}

// func (Session) OpenAnnounceStream(stream Stream) (*ReceiveAnnounceStream, error) {
// 	// Send the Stream Type ID and notify the Stream Type is the Announce
// 	_, err := stream.Write([]byte{byte(protocol.ANNOUNCE)})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &ReceiveAnnounceStream{
// 		stream:   stream,
// 		qvReader: quicvarint.NewReader(stream),
// 	}, nil
// }

// func (Session) AcceptAnnounceStream(stream Stream, ctx context.Context) (*SendAnnounceStream, error) {
// 	/*
// 	 * Verify the Stream Type is the Announce
// 	 */
// 	// Peek and read the Stream Type
// 	peeker := bufio.NewReader(stream)
// 	b, err := peeker.Peek(1)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Verify the Stream Type ID
// 	if StreamType(b[0]) != protocol.ANNOUNCE {
// 		return nil, ErrUnexpectedStreamType
// 	}
// 	// Read and advance by 1 byte
// 	streamTypeBuf := make([]byte, 1)
// 	_, err = stream.Read(streamTypeBuf)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &SendAnnounceStream{
// 		stream:   stream,
// 		qvReader: quicvarint.NewReader(stream),
// 	}, nil
// }

// func (sess Session) OpenSubscribeStream() (*SendSubscribeStream, error) {
// 	/*
// 	 * Open a bidirectional stream
// 	 */

// 	// Send the Stream Type ID and notify the Stream Type is the Subscribe
// 	_, err := stream.Write([]byte{byte(protocol.SUBSCRIBE)})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &SendSubscribeStream{
// 		stream:           stream,
// 		qvReader:         quicvarint.NewReader(stream),
// 		subscribeCounter: &sess.subscribeCounter,
// 		trackAliasMap:    sess.trackAliasMap,
// 	}, nil
// }

// func (sess Session) AcceptSubscribeStream(stream Stream, ctx context.Context) (*ReceiveSubscribeStream, error) {
// 	/*
// 	 * Verify the Stream Type is the Subscribe
// 	 */
// 	// Read and advance by 1 byte
// 	streamTypeBuf := make([]byte, 1)
// 	_, err := stream.Read(streamTypeBuf)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &ReceiveSubscribeStream{
// 		stream:   stream,
// 		qvReader: quicvarint.NewReader(stream),
// 	}, nil
// }

// func (sess Session) PeekStreamType(stream Stream) (StreamType, error) {
// 	// Peek and read the Stream Type
// 	peeker := bufio.NewReader(stream)
// 	b, err := peeker.Peek(1)
// 	if err != nil {
// 		return 0, err
// 	}

// 	return StreamType(b[0]), nil
// }
