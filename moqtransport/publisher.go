package moqtransport

type Publisher struct {
	node           node
	MaxSubscribeID uint64
}

func (p *Publisher) ConnectAndSetup(url string) (*PublishingSession, error) {
	return p.node.EstablishPubSession(url, p.MaxSubscribeID)
}

// func (p Publisher) SendObjectDatagram(od moqtmessage.ObjectDatagram) error { //TODO:
// 	return p.session.SendDatagram(od.Serialize())
// }

// func (p Publisher) SendSingleObject(priority moqtmessage.PublisherPriority, payload []byte) <-chan error {
// 	dataCh := make(chan []byte, 1)
// 	defer close(dataCh)

// 	header := moqtmessage.StreamHeaderTrack{
// 		//subscribeID: ,
// 		//TrackAlias: ,
// 		PublisherPriority: priority,
// 	}

// 	dataCh <- payload

// 	return p.sendMultipleObject(&header, dataCh)
// }

// func (p Publisher) SendMultipleObject(priority moqtmessage.PublisherPriority, payload <-chan []byte) <-chan error {
// 	header := moqtmessage.StreamHeaderTrack{
// 		//subscribeID: ,
// 		//TrackAlias: ,
// 		PublisherPriority: priority,
// 	}
// 	return p.sendMultipleObject(&header, payload) // TODO:
// }

// func (p *Publisher) sendMultipleObject(header moqtmessage.StreamHeader, payloadCh <-chan []byte) <-chan error {

// 	errCh := make(chan error, 1)
// 	stream, err := p.session.OpenUniStream()
// 	if err != nil {
// 		errCh <- err
// 	}

// 	go func() {
// 		// Send the header
// 		_, err := stream.Write(header.Serialize())
// 		if err != nil {
// 			log.Println(err)
// 			errCh <- err
// 			return
// 		}

// 		// Get chunk stream to get chunks
// 		chunkStream := moqtmessage.NewChunkStream(header)
// 		var chunk moqtmessage.Chunk
// 		for payload := range payloadCh {
// 			chunk = chunkStream.CreateChunk(payload)
// 			_, err = stream.Write(chunk.Serialize())
// 			if err != nil {
// 				log.Println(err)
// 				errCh <- err
// 				return
// 			}
// 		}

// 		// Send final chunk as end of the stream
// 		chunk = chunkStream.CreateFinalChunk()
// 		_, err = stream.Write(chunk.Serialize())
// 		if err != nil {
// 			log.Println(err)
// 			errCh <- err
// 			return
// 		}
// 	}()

// 	return errCh
// }

/*
 *
 *
 */

// /*
//  * Response to a TRACK_STATUS_REQUEST
//  */
// func (p Publisher) sendTrackStatus() error {
// 	ts := moqtmessage.TrackStatusMessage{
// 		TrackNamespace: p.TrackNamespace,
// 		TrackName:      "",
// 		Code:           0,
// 		LastGroupID:    0, // TODO
// 		LastObjectID:   0, // TODO
// 	}
// 	p.controlStream.Write(ts.Serialize())
// 	return nil
// }
